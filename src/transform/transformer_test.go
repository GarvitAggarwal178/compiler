package transform

import (
	"testing"

	"dlc/src/ast"
	"dlc/src/sema"
)

func TestPassThroughReturnsInputUnchanged(t *testing.T) {
	prog := &ast.Program{Clauses: []*ast.Clause{{Head: &ast.Atom{Name: "p"}}}}
	strata := &sema.StratumResult{Stratum: map[string]int{"p": 0}}
	got, err := PassThrough{}.Transform(prog, strata)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != prog {
		t.Fatalf("expected PassThrough to return the exact same *ast.Program pointer, got a different one")
	}
}

func TestPassThroughSatisfiesTransformer(t *testing.T) {
	var _ Transformer = PassThrough{}
}
