// M3.1 -- seed collection over negated occurrences. Package guard
// implements the transform-safety analysis: this file's assertion
// (the correctness of everything else rests on it), culprit-cycle
// detection (stratify.go), and the per-SCC TRANSFORM/FALLBACK decision
// plus fallback cone (decide.go).
package guard

import (
	"fmt"

	"dlc/src/transform/magicset"
)

// AssertNegationSeeding is M2-M3-BUILD.md §5's required gate, AMENDED by
// M4-SIPS.md §1 to check both halves of the relaxation lemma rather than
// one. It REPLACES the original AssertNegationAllBound (M4-SIPS.md §1:
// "must be replaced, not deleted") -- the original invariant is still
// checked, in full, as half of this function; it is not weakened.
//
// Two checks per negated occurrence:
//
//  1. PreAdorn (the adornment before M4-SIPS.md §2's relaxation) must be
//     all-bound. This is the original M3.1 invariant, unchanged: allowedness
//     requires every variable in a negated literal to be grounded before
//     it is evaluated, so the UN-relaxed adornment is always all-b by
//     construction. A violation here means an allowedness bug or an
//     adornment bug -- exactly as before relaxation existed.
//  2. Adorn (the possibly-relaxed adornment actually used to generate a
//     magic relation) must never be bound at a position PreAdorn was not.
//     Relaxation is only ever allowed to turn a bound position free, never
//     the reverse (M4-SIPS.md §1: "replacing any b with f is sound...
//     only computing less can [decide the negation incorrectly]"). Since
//     PreAdorn is always all-bound (check 1), this reduces to "Adorn has
//     the same length as PreAdorn", but it is checked positionally and
//     explicitly rather than relied upon, so a future bug in
//     relaxNegatedAdornment that sets an f position to b is caught here
//     rather than silently producing an under-demanded, wrong-answer
//     program -- the exact failure mode this whole file exists to guard
//     against.
func AssertNegationSeeding(result *magicset.AdornResult) error {
	for _, occ := range result.NegatedOccurrenceAdornments() {
		if !occ.PreAdorn.AllBound() {
			return fmt.Errorf("guard: negated occurrence of %q has pre-relaxation adornment %q, not all-bound -- "+
				"this means either an allowedness bug (a negated atom's variable was not actually "+
				"grounded before evaluation) or an adornment bug (SIPS scheduled the negation before "+
				"its own variables were bound); escalate, do not relax this assertion. Source clause: %v",
				occ.Pred, occ.PreAdorn, occ.Rule)
		}
		if len(occ.Adorn) != len(occ.PreAdorn) {
			return fmt.Errorf("guard: negated occurrence of %q has adornment %q of different length than "+
				"pre-relaxation adornment %q -- relaxNegatedAdornment bug. Source clause: %v",
				occ.Pred, occ.Adorn, occ.PreAdorn, occ.Rule)
		}
		for i, b := range occ.Adorn {
			if b && !occ.PreAdorn[i] {
				return fmt.Errorf("guard: negated occurrence of %q relaxed a free position to bound "+
					"(adorn=%q, pre-relaxation=%q, position %d) -- relaxation must only ever turn b into f, "+
					"never the reverse (M4-SIPS.md §1); escalate, do not relax this assertion. Source clause: %v",
					occ.Pred, occ.Adorn, occ.PreAdorn, i, occ.Rule)
			}
		}
	}
	return nil
}
