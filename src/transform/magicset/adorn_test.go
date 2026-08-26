package magicset

import (
	"testing"

	"dlc/src/parser"
)

const ancestorNonancestorSrc = `.decl parent(child:number, par:number)
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

func TestFindQueryDetectsBoundConstant(t *testing.T) {
	prog, errs := parser.Parse([]byte(ancestorNonancestorSrc))
	if len(errs) != 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	q := FindQuery(prog)
	if q == nil {
		t.Fatal("expected a query, got nil")
	}
	if q.Key.pred != "nonancestor" || q.Key.adorn != "bf" {
		t.Fatalf("expected nonancestor/bf, got %s/%s", q.Key.pred, q.Key.adorn)
	}
}

func TestFindQueryNoOpWhenNoBoundQuery(t *testing.T) {
	src := ".decl p(x:number)\n.decl q(x:number)\n.output q\np(1).\nq(x) :- p(x).\n"
	prog, errs := parser.Parse([]byte(src))
	if len(errs) != 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	if q := FindQuery(prog); q != nil {
		t.Fatalf("expected nil (no atom in q's projection rule has a constant arg), got %+v", q)
	}
}

// TestAdornNegatedOccurrenceIsAllBound pins M2-M3-BUILD.md §5's central
// claim directly at the adornment level: !ancestor(x,y) inside
// nonancestor's rule has BOTH x and y already grounded (by person(x) and
// person(y), both of which SIPS is forced to schedule first -- a negated
// atom's variables must be grounded before it can be evaluated at all),
// so the computed adornment is bb, not the bf a reader might expect by
// analogy to the query's own bf binding pattern.
func TestAdornNegatedOccurrenceIsAllBound(t *testing.T) {
	prog, errs := parser.Parse([]byte(ancestorNonancestorSrc))
	if len(errs) != 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	q := FindQuery(prog)
	result, err := Adorn(prog, q)
	if err != nil {
		t.Fatalf("Adorn error: %v", err)
	}
	if len(result.Order) != 2 {
		t.Fatalf("expected exactly 2 adorned (pred,adornment) pairs, got %d: %v", len(result.Order), result.Order)
	}
	found := false
	for _, k := range result.Order {
		if k.pred == "ancestor" {
			found = true
			if k.adorn != "bb" {
				t.Fatalf("expected ancestor adorned bb (both args grounded before the negation fires), got %s", k.adorn)
			}
		}
	}
	if !found {
		t.Fatal("ancestor was never adorned at all")
	}
}

func TestAdornEDBAtomsNeverPushed(t *testing.T) {
	prog, errs := parser.Parse([]byte(ancestorNonancestorSrc))
	if len(errs) != 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	q := FindQuery(prog)
	result, err := Adorn(prog, q)
	if err != nil {
		t.Fatalf("Adorn error: %v", err)
	}
	for _, k := range result.Order {
		if k.pred == "parent" || k.pred == "person" {
			t.Fatalf("EDB relation %q was adorned -- magic sets must never restrict an EDB/.input-only relation", k.pred)
		}
	}
}
