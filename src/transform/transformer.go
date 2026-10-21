// Package transform defines the common rewrite interface and typed provenance.
package transform

import (
	"dlc/src/ast"
	"dlc/src/sema"
)

// Transformer receives accepted source and its source stratification. Callers
// revalidate the returned program and use its resulting schemas and strata.
// ReportingTransformer in report.go additionally exposes real decisions/origins.
type Transformer interface {
	Transform(prog *ast.Program, strata *sema.StratumResult) (*ast.Program, error)
}

// PassThrough retains the original program without demand rewriting.
type PassThrough struct{}

func (PassThrough) Transform(prog *ast.Program, _ *sema.StratumResult) (*ast.Program, error) {
	return prog, nil
}
