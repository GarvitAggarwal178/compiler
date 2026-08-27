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

// AssertNegationAllBound is M2-M3-BUILD.md §5's required gate: every
// adorned occurrence of a negated atom must have an all-bound ("b...b")
// adornment. This is not a separate algorithm to implement -- it is a
// property that falls out of magicset's adornment being computed
// correctly, given that allowedness (already validated, sema/
// allowedness.go) requires every variable in a negated literal to be
// grounded before it is evaluated. Written as an assertion specifically
// so a violation (an allowedness bug or an adornment bug -- the two live
// possibilities per §5) is caught loudly rather than silently producing
// an under-demanded, wrong-answer program.
func AssertNegationAllBound(result *magicset.AdornResult) error {
	for _, occ := range result.NegatedOccurrenceAdornments() {
		if !occ.Adorn.AllBound() {
			return fmt.Errorf("guard: negated occurrence of %q has adornment %q, not all-bound -- "+
				"this means either an allowedness bug (a negated atom's variable was not actually "+
				"grounded before evaluation) or an adornment bug (SIPS scheduled the negation before "+
				"its own variables were bound); escalate, do not relax this assertion. Source clause: %v",
				occ.Pred, occ.Adorn, occ.Rule)
		}
	}
	return nil
}
