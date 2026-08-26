package parser

import (
	"testing"

	"dlc/src/ast"
)

func mustParse(t *testing.T, src string) *ast.Program {
	t.Helper()
	prog, errs := Parse([]byte(src))
	if len(errs) != 0 {
		t.Fatalf("Parse(%q) returned unexpected errors: %v", src, errs)
	}
	return prog
}

func TestBasicDeclAndFact(t *testing.T) {
	prog := mustParse(t, ".decl p(a:number)\np(1).\n")
	if len(prog.Decls) != 1 || prog.Decls[0].Name != "p" {
		t.Fatalf("expected 1 decl named p, got %+v", prog.Decls)
	}
	if len(prog.Clauses) != 1 || len(prog.Clauses[0].Body) != 0 {
		t.Fatalf("expected 1 fact with empty body, got %+v", prog.Clauses)
	}
}

func TestInterleavedDeclsAndClauses(t *testing.T) {
	// Blueprint's "decl* clause*" is read as unordered/interleaved --
	// this project's own culprit_cycle.dl interleaves throughout and
	// was already validated against real Soufflé. Reject-on-interleave
	// would break that file.
	src := `
.decl q(x:number, y:number)
q(x,y) :- q(x,y).
.decl s(x:number)
s(x) :- q(x,x).
`
	prog := mustParse(t, src)
	if len(prog.Decls) != 2 || len(prog.Clauses) != 2 {
		t.Fatalf("expected 2 decls and 2 clauses interleaved, got decls=%d clauses=%d", len(prog.Decls), len(prog.Clauses))
	}
}

func TestZeroArityDeclAndAtom(t *testing.T) {
	prog := mustParse(t, ".decl p()\n.output p\np().\n")
	if len(prog.Decls[0].Params) != 0 {
		t.Fatalf("expected zero params, got %v", prog.Decls[0].Params)
	}
	if len(prog.Clauses[0].Head.Terms) != 0 {
		t.Fatalf("expected zero-arity head atom, got %v", prog.Clauses[0].Head.Terms)
	}
}

func TestNegationAndConstraintDisambiguation(t *testing.T) {
	prog := mustParse(t, ".decl q(a:number)\n.decl p(a:number)\np(X) :- q(X), !q(X), X > 0.\n")
	body := prog.Clauses[0].Body
	if len(body) != 3 {
		t.Fatalf("expected 3 body literals, got %d: %v", len(body), body)
	}
	if _, ok := body[0].(*ast.Atom); !ok {
		t.Fatalf("literal 0: expected *ast.Atom, got %T", body[0])
	}
	if _, ok := body[1].(*ast.NegatedAtom); !ok {
		t.Fatalf("literal 1: expected *ast.NegatedAtom, got %T", body[1])
	}
	c, ok := body[2].(*ast.Constraint)
	if !ok {
		t.Fatalf("literal 2: expected *ast.Constraint, got %T", body[2])
	}
	if c.Op != ">" {
		t.Fatalf("expected relop '>', got %q", c.Op)
	}
}

// Precedence table, mirrors tests/hostile/precedence_*.dl (already
// oracle-verified against real Soufflé, docs/reports/night02-T2-hostile.md).
func TestPrecedence(t *testing.T) {
	cases := []struct {
		name string
		expr string
		want string // canonical fully-parenthesized shape, innermost-first reading
	}{
		{"add_mul", "2 + 3 * 4", "(2+(3*4))"},        // not (2+3)*4
		{"sub_mul", "10 - 2 * 3", "(10-(2*3))"},       // not (10-2)*3
		{"mul_div_left_assoc", "20 / 4 * 2", "((20/4)*2)"}, // not 20/(4*2)
		{"left_assoc_add", "10 - 3 - 2", "((10-3)-2)"},     // not 10-(3-2)
		{"unary_sub", "-2 - 3", "((-2)-3)"},                // not -(2-3)
		{"paren_override", "(2 + 3) * 4", "((2+3)*4)"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			src := ".decl p(a:number)\np(X) :- X = " + c.expr + ".\n"
			prog := mustParse(t, src)
			constraint := prog.Clauses[0].Body[0].(*ast.Constraint)
			got := canon(constraint.Right.(ast.Arith))
			if got != c.want {
				t.Fatalf("%s: got shape %s, want %s", c.expr, got, c.want)
			}
		})
	}
}

// canon renders an Arith as a fully-parenthesized canonical string,
// independent of the pretty-printer, so this test does not depend on
// Print's own output format and would still catch a precedence bug even
// if the printer had a compensating bug.
func canon(a ast.Arith) string {
	switch v := a.(type) {
	case *ast.BinaryExpr:
		return "(" + canon(v.Left) + v.Op + canon(v.Right) + ")"
	case *ast.UnaryExpr:
		return "(" + v.Op + canon(v.X) + ")"
	case *ast.Var:
		return v.Name
	case *ast.NumberLit:
		return v.Text
	}
	return "?"
}

func TestRelopsDoNotChain(t *testing.T) {
	// X < Y < Z is not valid: relop is not part of arith, so after the
	// first constraint's right-hand arith is parsed, a second relop
	// cannot follow inside the same literal.
	_, errs := Parse([]byte(".decl q(a:number,b:number,c:number)\np(X) :- q(X,Y,Z), X < Y < Z.\n"))
	if len(errs) == 0 {
		t.Fatalf("expected a parse error for chained relops, got none")
	}
}

func TestErrorRecoveryOneMalformedClauseDoesNotKillFile(t *testing.T) {
	src := ".decl p(a:number)\np(X) :- .\np(1).\n"
	prog, errs := Parse([]byte(src))
	if len(errs) == 0 {
		t.Fatalf("expected an error for the malformed empty-body clause")
	}
	if len(prog.Clauses) != 1 || len(prog.Clauses[0].Head.Terms) != 1 {
		t.Fatalf("expected the second, valid clause (p(1).) to still parse; got clauses=%+v", prog.Clauses)
	}
}

func TestRoundtripMatchesOnRepresentativePrograms(t *testing.T) {
	srcs := []string{
		".decl p(a:number)\n.output p\np(1).\np(2).\n",
		".decl q(a:number,b:number)\n.input q\n.decl p(a:number)\n.output p\np(X) :- q(X,Y), X = Y + 1, !q(Y,X).\n",
		".decl p()\n.output p\np().\n",
		".decl p(a:number)\n.output p\np(X) :- p(X), X = 2 + 3 * 4 - -5.\n",
	}
	for _, src := range srcs {
		r := Roundtrip([]byte(src))
		if r.Status != "match" {
			t.Fatalf("Roundtrip(%q) = %+v, want status=match", src, r)
		}
	}
}
