// Package transform defines the interface Lane A's magic-set transform
// and guard (src/transform/magicset/, src/transform/guard/ -- both Lane
// A, docs/M1-BUILD.md §1) will implement, plus a Lane B pass-through
// placeholder so the full pipeline (parse -> check -> transform ->
// evaluate) runs end-to-end today, before either exists. This file
// itself is Lane B; only the two named subpackages are Lane A.
package transform

import (
	"dlc/src/ast"
	"dlc/src/sema"
)

// Transformer is what §4 item 2 asks for: "the interface Lane A will
// implement," with the signature the evaluator needs.
//
// What it expects: prog has already passed every sema check this
// project runs (decl/type, allowedness, stratification) -- it is
// guaranteed accepted, and strata is the StratumResult that acceptance
// produced. Transform decides, per blueprint's TRANSFORM/FALLBACK rule
// (evaluated per-SCC on the *transformed* precedence graph, not the
// source one CheckStratification already built -- a different pass,
// deliberately not generalized from sema/stratify.go, see that
// package's own DESIGN.md), which SCCs get magic-set adornment and
// which stay untransformed, and returns the program the evaluator
// should actually run: either prog itself unchanged (nothing cleared
// for TRANSFORM) or a rewritten program with magic predicates
// introduced for whichever SCCs did.
//
// The returned program must itself be valid under every sema check
// (parses, well-typed, allowed, stratifiable) -- Transform does not
// get a free pass to hand back something ast.Equal-incompatible with
// the rest of the pipeline's assumptions. Note for whoever implements
// this: a real transform that introduces new magic-seed relations
// changes the precedence graph, so the returned program's *own*
// stratification is not guaranteed to match the strata argument passed
// in -- callers that need strata for the transformed program must
// re-run sema.CheckStratification on the result, not reuse the
// pre-transform StratumResult.
type Transformer interface {
	Transform(prog *ast.Program, strata *sema.StratumResult) (*ast.Program, error)
}

// PassThrough is the Lane B placeholder: never transforms anything.
// Lets cmd/dlc's pipeline call a real Transformer today, computing
// exactly the answer naive/semi-naive evaluation already computes
// directly on prog -- Lane A replaces this with the real magic-set +
// guard implementation, and nothing on the evaluator or CLI side needs
// to change when that happens (docs/reports/m1-progress.md explains why
// this interface has the shape it has).
type PassThrough struct{}

func (PassThrough) Transform(prog *ast.Program, _ *sema.StratumResult) (*ast.Program, error) {
	return prog, nil
}
