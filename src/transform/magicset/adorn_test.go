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

func TestFindQueriesDetectsBoundConstant(t *testing.T) {
	prog, errs := parser.Parse([]byte(ancestorNonancestorSrc))
	if len(errs) != 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	qs := FindQueries(prog)
	if len(qs) != 1 {
		t.Fatalf("expected exactly 1 query, got %d", len(qs))
	}
	q := qs[0]
	if q.Key.pred != "nonancestor" || q.Key.adorn != "bf" {
		t.Fatalf("expected nonancestor/bf, got %s/%s", q.Key.pred, q.Key.adorn)
	}
}

func TestFindQueriesNoOpWhenNoBoundQuery(t *testing.T) {
	src := ".decl p(x:number)\n.decl q(x:number)\n.output q\np(1).\nq(x) :- p(x).\n"
	prog, errs := parser.Parse([]byte(src))
	if len(errs) != 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	if qs := FindQueries(prog); len(qs) != 0 {
		t.Fatalf("expected none (no atom in q's projection rule has a constant arg), got %+v", qs)
	}
}

// TestFindQueriesCollectsMultiple pins PUNCH-LIST.md P1's seed-collection
// change directly: a program with two independent .output branches, each
// individually bindable, must have BOTH found -- not just the first in
// source order (the pre-P1 behaviour this test would have caught).
func TestFindQueriesCollectsMultiple(t *testing.T) {
	src := `.decl e(x:number, y:number)
.input e
.decl p(x:number, y:number)
p(x,y) :- e(x,y).
.decl tc(x:number, y:number)
tc(x,y) :- e(x,y).
tc(x,y) :- tc(x,z), e(z,y).
.decl out(y:number)
.output out
out(y) :- p(1,y).
.decl out2(y:number)
.output out2
out2(y) :- tc(1,y).
`
	prog, errs := parser.Parse([]byte(src))
	if len(errs) != 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	qs := FindQueries(prog)
	if len(qs) != 2 {
		t.Fatalf("expected 2 queries (out and out2), got %d: %+v", len(qs), qs)
	}
	preds := map[string]bool{}
	for _, q := range qs {
		preds[q.Key.pred] = true
	}
	if !preds["p"] || !preds["tc"] {
		t.Fatalf("expected queries targeting both p and tc, got %v", preds)
	}
}

// TestAdornNegatedOccurrenceRelaxesToBF pins M4-SIPS.md §2's demand
// relaxation directly at the adornment level. !ancestor(x,y) inside
// nonancestor's rule has BOTH x and y already grounded before it fires
// (by person(x) and person(y) -- a negated atom's variables must be
// grounded before it can be evaluated at all, M2-M3-BUILD.md §5's
// original claim, still true of the PRE-relaxation adornment). But y's
// only binder is person(y), an unrestricted full-extent scan (x is
// already bound from the query's magic seed when person(x) runs) -- so
// M4-SIPS.md §2 relaxes y's position to free, and the adornment ACTUALLY
// USED to generate ancestor's magic relation is bf, not the bb this test
// originally pinned before M4-SIPS (git history: "AdornNegatedOccurrence
// IsAllBound"). The pre-relaxation value is still bb, asserted directly
// via NegatedOccurrenceAdornments in adorn_report_test.go /
// guard/seeding_test.go, not re-duplicated here.
func TestAdornNegatedOccurrenceRelaxesToBF(t *testing.T) {
	prog, errs := parser.Parse([]byte(ancestorNonancestorSrc))
	if len(errs) != 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	qs := FindQueries(prog)
	result, err := Adorn(prog, qs)
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
			if k.adorn != "bf" {
				t.Fatalf("expected ancestor adorned bf (y relaxed -- its only binder is person(y), a full-extent scan), got %s", k.adorn)
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
	qs := FindQueries(prog)
	result, err := Adorn(prog, qs)
	if err != nil {
		t.Fatalf("Adorn error: %v", err)
	}
	for _, k := range result.Order {
		if k.pred == "parent" || k.pred == "person" {
			t.Fatalf("EDB relation %q was adorned -- magic sets must never restrict an EDB/.input-only relation", k.pred)
		}
	}
}
