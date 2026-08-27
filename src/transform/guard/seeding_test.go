package guard

import (
	"testing"

	"dlc/src/parser"
	"dlc/src/transform/magicset"
)

func TestAssertNegationSeedingAcceptsRealShape(t *testing.T) {
	src := `.decl parent(child:number, par:number)
.input parent
.decl person(x:number)
.input person
.decl ancestor(x:number, y:number)
ancestor(x,y) :- parent(x,y).
ancestor(x,y) :- parent(x,z), ancestor(z,y).
.decl nonancestor(x:number, y:number)
nonancestor(x,y) :- person(x), person(y), !ancestor(x,y).
.decl q_nonancestor(y:number)
.output q_nonancestor
q_nonancestor(y) :- nonancestor(1,y).
`
	prog, errs := parser.Parse([]byte(src))
	if len(errs) != 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	qs := magicset.FindQueries(prog)
	result, err := magicset.Adorn(prog, qs)
	if err != nil {
		t.Fatalf("Adorn error: %v", err)
	}
	if err := AssertNegationSeeding(result); err != nil {
		t.Fatalf("expected no violation on a real shape, got: %v", err)
	}
}

func TestAssertNegationSeedingFindsAtLeastOneOccurrence(t *testing.T) {
	// Not a violation test, but confirms the plumbing actually inspects
	// real occurrences rather than vacuously passing on an empty list.
	src := `.decl parent(child:number, par:number)
.input parent
.decl person(x:number)
.input person
.decl ancestor(x:number, y:number)
ancestor(x,y) :- parent(x,y).
ancestor(x,y) :- parent(x,z), ancestor(z,y).
.decl nonancestor(x:number, y:number)
nonancestor(x,y) :- person(x), person(y), !ancestor(x,y).
.decl q_nonancestor(y:number)
.output q_nonancestor
q_nonancestor(y) :- nonancestor(1,y).
`
	prog, errs := parser.Parse([]byte(src))
	if len(errs) != 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	qs := magicset.FindQueries(prog)
	result, err := magicset.Adorn(prog, qs)
	if err != nil {
		t.Fatalf("Adorn error: %v", err)
	}
	occs := result.NegatedOccurrenceAdornments()
	if len(occs) == 0 {
		t.Fatal("expected at least one negated occurrence (!ancestor(x,y)), found none")
	}
	for _, occ := range occs {
		// PreAdorn -- the original M3.1 invariant -- must still be
		// all-bound unconditionally; this is the property allowedness
		// guarantees and relaxation is never allowed to touch.
		if !occ.PreAdorn.AllBound() {
			t.Fatalf("occurrence of %q has non-all-bound PRE-relaxation adornment %q", occ.Pred, occ.PreAdorn)
		}
	}
}

// TestNegatedOccurrenceRelaxationCollapsesToBF pins M4-SIPS.md §2's
// worked instance directly: on this exact program, !ancestor(x,y)'s
// grounding atoms are `person(x)` (x already bound from the query
// constant) and `person(y)` (a full-extent scan -- y's only binder).
// The pre-relaxation adornment must still be "bb" (all-bound, the M3.1
// invariant); the post-relaxation adornment actually used to generate a
// magic relation must be "bf" -- y carries no demand information and is
// relaxed to free.
func TestNegatedOccurrenceRelaxationCollapsesToBF(t *testing.T) {
	src := `.decl parent(child:number, par:number)
.input parent
.decl person(x:number)
.input person
.decl ancestor(x:number, y:number)
ancestor(x,y) :- parent(x,y).
ancestor(x,y) :- parent(x,z), ancestor(z,y).
.decl nonancestor(x:number, y:number)
nonancestor(x,y) :- person(x), person(y), !ancestor(x,y).
.decl q_nonancestor(y:number)
.output q_nonancestor
q_nonancestor(y) :- nonancestor(1,y).
`
	prog, errs := parser.Parse([]byte(src))
	if len(errs) != 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	qs := magicset.FindQueries(prog)
	result, err := magicset.Adorn(prog, qs)
	if err != nil {
		t.Fatalf("Adorn error: %v", err)
	}
	occs := result.NegatedOccurrenceAdornments()
	found := false
	for _, occ := range occs {
		if occ.Pred != "ancestor" {
			continue
		}
		found = true
		if occ.PreAdorn.String() != "bb" {
			t.Fatalf("expected pre-relaxation adornment bb, got %q", occ.PreAdorn.String())
		}
		if occ.Adorn.String() != "bf" {
			t.Fatalf("expected post-relaxation adornment bf, got %q", occ.Adorn.String())
		}
	}
	if !found {
		t.Fatal("expected a negated occurrence of ancestor, found none")
	}
	if err := AssertNegationSeeding(result); err != nil {
		t.Fatalf("expected no violation, got: %v", err)
	}
}
