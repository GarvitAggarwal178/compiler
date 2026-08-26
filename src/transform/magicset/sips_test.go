package magicset

import (
	"testing"

	"dlc/src/ast"
	"dlc/src/parser"
)

func parseOneRuleBody(t *testing.T, src string) []ast.Literal {
	t.Helper()
	prog, errs := parser.Parse([]byte(src))
	if len(errs) != 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	for _, c := range prog.Clauses {
		if len(c.Body) > 0 {
			return c.Body
		}
	}
	t.Fatal("no rule with a body found")
	return nil
}

func literalString(lit ast.Literal) string {
	var sb []byte
	switch v := lit.(type) {
	case *ast.Atom:
		sb = append(sb, v.Name...)
	case *ast.NegatedAtom:
		sb = append(sb, '!')
		sb = append(sb, v.Atom.Name...)
	case *ast.Constraint:
		sb = append(sb, "constraint:"+v.Op...)
	}
	return string(sb)
}

func TestOrderBodyNegatedLiteralMovesBack(t *testing.T) {
	// !ancestor(x,y) needs y, which only person(y) grounds -- the
	// negated literal must move after person(y) even though it appears
	// before it in source order.
	src := ".decl person(x:number)\n.decl ancestor(x:number,y:number)\n.decl nonancestor(x:number,y:number)\n" +
		"nonancestor(x,y) :- person(x), !ancestor(x,y), person(y).\n"
	body := parseOneRuleBody(t, src)
	ordered := OrderBody(body, nil)
	if len(ordered) != 3 {
		t.Fatalf("expected 3 literals, got %d", len(ordered))
	}
	got := []string{literalString(ordered[0]), literalString(ordered[1]), literalString(ordered[2])}
	want := []string{"person", "person", "!ancestor"}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("order = %v, want negated literal last (both persons before it): %v", got, want)
		}
	}
}

func TestOrderBodyConstraintPulledForward(t *testing.T) {
	// x > 1 only needs x, bound by q(x) -- it should be schedulable
	// immediately after q(x), before r(y) (whose own binding it doesn't
	// need and doesn't affect).
	src := ".decl q(x:number)\n.decl r(y:number)\n.decl p(x:number,y:number)\n" +
		"p(x,y) :- q(x), r(y), x > 1.\n"
	body := parseOneRuleBody(t, src)
	ordered := OrderBody(body, nil)
	// x > 1 is schedulable right after q(x) is processed in the same
	// pass (r(y) is unconditionally schedulable too, so source order is
	// preserved among literals that don't depend on each other) --
	// the real assertion is that the constraint does NOT get stuck
	// waiting, i.e. it appears in the result at all with the same
	// length, and specifically before nothing blocks it.
	if len(ordered) != 3 {
		t.Fatalf("expected 3 literals, got %d: %v", len(ordered), ordered)
	}
	foundQ, foundConstraint := false, false
	for _, lit := range ordered {
		if literalString(lit) == "q" {
			foundQ = true
		}
		if _, ok := lit.(*ast.Constraint); ok {
			foundConstraint = true
			if !foundQ {
				t.Fatalf("constraint x>1 scheduled before q(x) grounds x: %v", ordered)
			}
		}
	}
	if !foundConstraint {
		t.Fatalf("constraint missing from ordered body: %v", ordered)
	}
}

func TestOrderBodyRespectsInitialBound(t *testing.T) {
	// With x pre-bound (as if from the head's adornment), !ancestor(x,y)
	// still needs y from person(y) but no longer needs anything for x.
	src := ".decl person(x:number)\n.decl ancestor(x:number,y:number)\n.decl nonancestor(x:number,y:number)\n" +
		"nonancestor(x,y) :- !ancestor(x,y), person(y).\n"
	body := parseOneRuleBody(t, src)
	ordered := OrderBody(body, map[string]bool{"x": true})
	if len(ordered) != 2 {
		t.Fatalf("expected 2 literals, got %d", len(ordered))
	}
	if literalString(ordered[0]) != "person" || literalString(ordered[1]) != "!ancestor" {
		t.Fatalf("expected [person, !ancestor] with x pre-bound, got %v",
			[]string{literalString(ordered[0]), literalString(ordered[1])})
	}
}
