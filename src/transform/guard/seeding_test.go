package guard

import (
	"testing"

	"dlc/src/parser"
	"dlc/src/transform/magicset"
)

func TestAssertNegationAllBoundAcceptsRealShape(t *testing.T) {
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
	q := magicset.FindQuery(prog)
	result, err := magicset.Adorn(prog, q)
	if err != nil {
		t.Fatalf("Adorn error: %v", err)
	}
	if err := AssertNegationAllBound(result); err != nil {
		t.Fatalf("expected no violation on a real shape, got: %v", err)
	}
}

func TestAssertNegationAllBoundFindsAtLeastOneOccurrence(t *testing.T) {
	// Not a violation test (adorn.go always produces all-bound negated
	// occurrences by construction -- there is no code path that could
	// produce anything else without allowedness having already rejected
	// the program), but confirms the plumbing actually inspects real
	// occurrences rather than vacuously passing on an empty list.
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
	q := magicset.FindQuery(prog)
	result, err := magicset.Adorn(prog, q)
	if err != nil {
		t.Fatalf("Adorn error: %v", err)
	}
	occs := result.NegatedOccurrenceAdornments()
	if len(occs) == 0 {
		t.Fatal("expected at least one negated occurrence (!ancestor(x,y)), found none")
	}
	for _, occ := range occs {
		if !occ.Adorn.AllBound() {
			t.Fatalf("occurrence of %q has non-all-bound adornment %q", occ.Pred, occ.Adorn)
		}
	}
}
