// M3.2 -- culprit-cycle detection, clause (a) of the transform-safety
// guard: does the candidate transformed program still stratify?
package guard

import (
	"dlc/src/ast"
	"dlc/src/sema"
	"dlc/src/transform/magicset"
)

// HasPositiveCycle returns, for every relation with at least one defining
// clause, whether it lies in a positive cycle in prog's SOURCE precedence
// graph -- a self-loop, or membership in a >1-relation SCC built from
// positive edges only. This is M2-M3-BUILD.md §6's "cheap necessary
// precondition," checked in O(V+E) before any transform work: the
// affected predicate must lie in a positive cycle in the source, or a
// culprit cycle cannot arise post-transform at all. A small, dedicated
// positive-edges-only graph -- not reused from sema/stratify.go's
// buildPrecedenceGraph (that graph carries both edge polarities and its
// internals are unexported; this check needs only the positive subgraph
// and is cheap enough to build directly).
func HasPositiveCycle(prog *ast.Program) map[string]bool {
	nodes := map[string]bool{}
	posEdges := map[string]map[string]bool{}
	for _, c := range prog.Clauses {
		nodes[c.Head.Name] = true
	}
	for _, c := range prog.Clauses {
		for _, lit := range c.Body {
			if a, ok := lit.(*ast.Atom); ok && nodes[a.Name] {
				if posEdges[c.Head.Name] == nil {
					posEdges[c.Head.Name] = map[string]bool{}
				}
				posEdges[c.Head.Name][a.Name] = true
			}
		}
	}
	reachable := func(from, to string) bool {
		visited := map[string]bool{from: true}
		queue := []string{from}
		for len(queue) > 0 {
			n := queue[0]
			queue = queue[1:]
			for next := range posEdges[n] {
				if next == to {
					return true
				}
				if !visited[next] {
					visited[next] = true
					queue = append(queue, next)
				}
			}
		}
		return false
	}
	result := map[string]bool{}
	for n := range nodes {
		inCycle := false
		for next := range posEdges[n] {
			if next == n || reachable(next, n) {
				inCycle = true
				break
			}
		}
		result[n] = inCycle
	}
	return result
}

// CulpritCycleResult is clause (a)'s verdict on one program.
type CulpritCycleResult struct {
	NoBindableQuery     bool         // magicset.FindQuery found nothing -- transform was a no-op, trivially stratifiable
	PreconditionSkipped bool         // no relation in the source has a positive cycle at all -- the full adorn-and-check path was skipped, trivially stratifiable
	Transformed         *ast.Program // the candidate transformed program (== the source, unchanged, if NoBindableQuery or PreconditionSkipped)
	Stratifiable        bool
	Message             string // sema.CheckStratification's own diagnostic text, if not stratifiable -- already names the offending SCC set (sema/stratify.go's own message format)
}

// CheckCulpritCycle builds the candidate transformed program (magicset.
// Adorn + Generate) and runs sema.CheckStratification on it -- the
// pre-existing SOURCE stratifier, applied to the TRANSFORMED AST
// (M2-M3-BUILD.md §6: "closes the gap the Transformer interface
// documented"). Does not itself invoke Soufflé -- the differential
// oracle cross-check against Soufflé's own stratifier is a harness-level
// concern (harness/night_m3_2_culprit_detection.py), since it needs a
// real `souffle` invocation, not something this package shells out to.
func CheckCulpritCycle(prog *ast.Program) (*CulpritCycleResult, error) {
	schemas, diags := sema.BuildSymbolTable(prog)
	if len(diags) > 0 {
		return nil, errFromDiags("guard: BuildSymbolTable failed on an already-accepted program", diags)
	}
	query := magicset.FindQuery(prog)
	if query == nil {
		return &CulpritCycleResult{NoBindableQuery: true, Transformed: prog, Stratifiable: true}, nil
	}

	// Cheap precondition (§6): a culprit cycle requires SOME relation to
	// lie in a positive cycle in the source. If none does anywhere in
	// the program, no culprit cycle can arise post-transform, and the
	// full adorn-and-check path is skipped entirely.
	anyPositiveCycle := false
	for _, inCycle := range HasPositiveCycle(prog) {
		if inCycle {
			anyPositiveCycle = true
			break
		}
	}
	if !anyPositiveCycle {
		return &CulpritCycleResult{PreconditionSkipped: true, Transformed: prog, Stratifiable: true}, nil
	}

	adorned, err := magicset.Adorn(prog, query)
	if err != nil {
		return nil, err
	}
	transformed := magicset.Generate(prog, schemas.Relations, adorned)
	stratDiags, _ := sema.CheckStratification(transformed)
	if len(stratDiags) == 0 {
		return &CulpritCycleResult{Transformed: transformed, Stratifiable: true}, nil
	}
	return &CulpritCycleResult{Transformed: transformed, Stratifiable: false, Message: stratDiags[0].Message}, nil
}

func errFromDiags(prefix string, diags []sema.Diagnostic) error {
	msg := prefix + ":"
	for _, d := range diags {
		msg += " " + d.Message + ";"
	}
	return &diagError{msg}
}

type diagError struct{ msg string }

func (e *diagError) Error() string { return e.msg }
