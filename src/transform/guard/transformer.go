package guard

import (
	"dlc/src/ast"
	"dlc/src/sema"
	"dlc/src/transform"
)

// Transformer implements transform.Transformer: the FULL M3 pipeline --
// magicset's transform, guarded by this package's culprit-cycle
// detection and per-SCC TRANSFORM/FALLBACK decision. Registered in
// cmd/dlc as "--transformer=guarded" -- the artifact M3's headline run
// (§9) actually measures, as distinct from "--transformer=magicset"
// (M2's ungated transform, kept registered separately so a candidate can
// still be measured without the guard for comparison, e.g. to see what
// the guard declined).
type Transformer struct{}

var _ transform.Transformer = Transformer{}

func (Transformer) Transform(prog *ast.Program, _ *sema.StratumResult) (*ast.Program, error) {
	result, err := Decide(prog)
	if err != nil {
		return nil, err
	}
	return result.FinalProgram, nil
}
