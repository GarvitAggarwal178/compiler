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

func (t Transformer) Transform(prog *ast.Program, strata *sema.StratumResult) (*ast.Program, error) {
	out, err := t.TransformReported(prog, strata)
	if err != nil {
		return nil, err
	}
	return out.Program, nil
}
func (Transformer) TransformReported(prog *ast.Program, _ *sema.StratumResult) (*transform.Outcome, error) {
	schemas, diags := sema.BuildSymbolTable(prog)
	if len(diags) > 0 {
		return nil, fmt.Errorf("magicset: input schema invalid: %v", diags)
	}
	queries := FindQueries(prog)
	if len(queries) == 0 {
		return &transform.Outcome{Program: prog, Report: transform.IdentityReport(prog, "magicset")}, nil
	}
	result, err := Adorn(prog, queries)
	if err != nil {
		return nil, err
	}
	p, _, report := GenerateMixedReported(prog, schemas.Relations, result, nil)
	return &transform.Outcome{Program: p, Candidate: p, CandidateReport: report, Report: report}, nil
}
