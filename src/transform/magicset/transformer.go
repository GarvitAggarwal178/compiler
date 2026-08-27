package magicset

import (
	"fmt"

	"dlc/src/ast"
	"dlc/src/sema"
	"dlc/src/transform"
)

// Transformer implements transform.Transformer: find the bindable query,
// run the worklist adornment (adorn.go), and generate the magic-set +
// supplementary-predicate program (rules.go). A program with no bindable
// query (M2-M3-BUILD.md §2) is passed through unchanged, same contract as
// transform.PassThrough for that case.
type Transformer struct{}

var _ transform.Transformer = Transformer{}

func (Transformer) Transform(prog *ast.Program, _ *sema.StratumResult) (*ast.Program, error) {
	schemas, diags := sema.BuildSymbolTable(prog)
	if len(diags) > 0 {
		return nil, fmt.Errorf("magicset: input program failed BuildSymbolTable, which should be impossible for an "+
			"already-accepted program: %v", diags)
	}
	queries := FindQueries(prog)
	if len(queries) == 0 {
		return prog, nil
	}
	result, err := Adorn(prog, queries)
	if err != nil {
		return nil, err
	}
	out, _ := Generate(prog, schemas.Relations, result)
	return out, nil
}
