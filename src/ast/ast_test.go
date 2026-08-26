package ast

import (
	"testing"

	"dlc/src/token"
)

func sp(a, b int) token.Span {
	return token.Span{Start: token.Position{Offset: a, Line: 1, Col: a + 1}, End: token.Position{Offset: b, Line: 1, Col: b + 1}}
}

func TestEqualIgnoresSpan(t *testing.T) {
	mk := func(spanOffset int) *Program {
		return &Program{
			Sp: sp(spanOffset, spanOffset+1),
			Decls: []*Decl{
				{Kind: DeclRelation, Name: "p", Params: []Param{{Name: "a", Type: "number", Sp: sp(spanOffset, spanOffset+1)}}, Sp: sp(spanOffset, spanOffset+5)},
			},
			Clauses: []*Clause{
				{
					Head: &Atom{Name: "p", Terms: []Term{&Var{Name: "X", Sp: sp(spanOffset, spanOffset+1)}}, Sp: sp(spanOffset, spanOffset+4)},
					Body: []Literal{
						&Atom{Name: "q", Terms: []Term{&Var{Name: "X", Sp: sp(spanOffset, spanOffset+1)}}, Sp: sp(spanOffset, spanOffset+4)},
					},
					Sp: sp(spanOffset, spanOffset+10),
				},
			},
		}
	}
	a := mk(0)
	b := mk(100) // identical structure, every span shifted -- must still be Equal
	if !Equal(a, b) {
		t.Fatalf("expected structurally identical programs (differing only in Span) to be Equal")
	}
}

func TestEqualDetectsRealDifference(t *testing.T) {
	a := &Program{Clauses: []*Clause{{Head: &Atom{Name: "p"}}}}
	b := &Program{Clauses: []*Clause{{Head: &Atom{Name: "q"}}}}
	if Equal(a, b) {
		t.Fatalf("expected programs with different head relation names to be unequal")
	}
}

func TestEqualDetectsOperatorDifference(t *testing.T) {
	x := &Var{Name: "X"}
	one := &NumberLit{Value: 1}
	a := &Program{Clauses: []*Clause{{Head: &Atom{Name: "p"}, Body: []Literal{
		&Constraint{Left: x, Op: "<", Right: one},
	}}}}
	b := &Program{Clauses: []*Clause{{Head: &Atom{Name: "p"}, Body: []Literal{
		&Constraint{Left: x, Op: "<=", Right: one},
	}}}}
	if Equal(a, b) {
		t.Fatalf("expected '<' vs '<=' constraint to be unequal")
	}
}

func TestArithIsTerm(t *testing.T) {
	var _ Term = &BinaryExpr{}
	var _ Term = &UnaryExpr{}
	var _ Term = &Var{}
	var _ Term = &NumberLit{}
	var _ Term = &StringLit{}
	var _ Term = &Wildcard{}
	var _ Arith = &BinaryExpr{}
	var _ Literal = &Atom{}
	var _ Literal = &NegatedAtom{}
	var _ Literal = &Constraint{}
}
