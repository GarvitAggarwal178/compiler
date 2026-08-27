// M3.3 -- per-SCC decision and the fallback cone. This is the part
// that is easy to get wrong (M2-M3-BUILD.md §7): FALLBACK is
// downward-closed over the ENTIRE dependency relation, positive and
// negative edges alike, not just the SCC clause (a) actually flagged.
package guard

import (
	"sort"

	"dlc/src/ast"
	"dlc/src/sema"
	"dlc/src/transform/magicset"
)

// DecideResult is M3.3's whole output.
type DecideResult struct {
	NoBindableQuery     bool
	CulpritPredicates   map[string]bool // original predicates directly implicated by an unstratifiable SCC of the transformed program (before cone closure)
	DeclinedRelations   map[string]bool // Culprit UNION its downward dependency cone -- every original predicate that ends up FALLBACK
	ConeRelations       map[string]bool // DeclinedRelations MINUS CulpritPredicates -- what the cone alone added
	FinalProgram        *ast.Program
	FinalRelationOrigin map[string]string
}

// Decide runs M3.2's culprit-cycle check and, if the fully-transformed
// program is unstratifiable, computes the per-SCC TRANSFORM/FALLBACK
// decision and the resulting mixed program: every culprit predicate (one
// implicated in an unstratifiable SCC) plus its full downward dependency
// cone falls back to reading its original, untransformed, full-extent
// form; everything else keeps its magic-set-restricted form.
func Decide(prog *ast.Program) (*DecideResult, error) {
	cc, err := CheckCulpritCycle(prog)
	if err != nil {
		return nil, err
	}
	if cc.NoBindableQuery || cc.PreconditionSkipped {
		return &DecideResult{
			NoBindableQuery: cc.NoBindableQuery, CulpritPredicates: map[string]bool{},
			DeclinedRelations: map[string]bool{}, ConeRelations: map[string]bool{},
			FinalProgram: cc.Transformed, FinalRelationOrigin: identityOrigin(prog),
		}, nil
	}
	if cc.Stratifiable {
		return &DecideResult{
			CulpritPredicates: map[string]bool{}, DeclinedRelations: map[string]bool{}, ConeRelations: map[string]bool{},
			FinalProgram: cc.Transformed, FinalRelationOrigin: cc.RelationOrigin,
		}, nil
	}

	culprit := map[string]bool{}
	for _, scc := range cc.UnstratifiableSCCs {
		for _, rel := range scc {
			if orig, ok := cc.RelationOrigin[rel]; ok {
				culprit[orig] = true
			}
		}
	}

	cone := ConeClosure(prog, culprit)
	declined := map[string]bool{}
	for k := range culprit {
		declined[k] = true
	}
	for k := range cone {
		declined[k] = true
	}

	schemas, diags := sema.BuildSymbolTable(prog)
	if len(diags) > 0 {
		return nil, errFromDiags("guard: BuildSymbolTable failed on an already-accepted program", diags)
	}
	queries := magicset.FindQueries(prog)
	adorned, err := magicset.Adorn(prog, queries)
	if err != nil {
		return nil, err
	}
	finalProg, finalOrigin := magicset.GenerateMixed(prog, schemas.Relations, adorned, declined)

	return &DecideResult{
		CulpritPredicates: culprit, DeclinedRelations: declined, ConeRelations: cone,
		FinalProgram: finalProg, FinalRelationOrigin: finalOrigin,
	}, nil
}

func identityOrigin(prog *ast.Program) map[string]string {
	out := map[string]string{}
	for _, d := range prog.Decls {
		out[d.Name] = d.Name
	}
	return out
}

// ConeClosure returns the downward dependency closure of declined over
// prog's SOURCE precedence graph, following the FULL dependency relation
// (positive and negative edges alike) -- M2-M3-BUILD.md §7: if a declined
// relation's (now untransformed) rule reads another relation, that
// relation must also be declined, because the untransformed rule expects
// its FULL extent, not a magic-restricted one. The returned set excludes
// `declined` itself (same convention harness/cone_metric.py already
// established in NIGHT-BATCH-03 T9 -- "cone_relations" is what gets
// dragged down WITH the declined set, not an echo of the declined set).
func ConeClosure(prog *ast.Program, declined map[string]bool) map[string]bool {
	nodes := map[string]bool{}
	edges := map[string]map[string]bool{}
	for _, c := range prog.Clauses {
		nodes[c.Head.Name] = true
	}
	for _, c := range prog.Clauses {
		head := c.Head.Name
		for _, lit := range c.Body {
			var target string
			switch v := lit.(type) {
			case *ast.Atom:
				target = v.Name
			case *ast.NegatedAtom:
				target = v.Atom.Name
			default:
				continue
			}
			if !nodes[target] {
				continue
			}
			if edges[head] == nil {
				edges[head] = map[string]bool{}
			}
			edges[head][target] = true
		}
	}

	cone := map[string]bool{}
	var frontier []string
	for d := range declined {
		frontier = append(frontier, d)
	}
	sort.Strings(frontier) // deterministic traversal order (CLAUDE.md determinism rule) -- does not affect the final SET, only iteration order
	visited := map[string]bool{}
	for _, d := range frontier {
		visited[d] = true
	}
	for len(frontier) > 0 {
		n := frontier[0]
		frontier = frontier[1:]
		var deps []string
		for dep := range edges[n] {
			deps = append(deps, dep)
		}
		sort.Strings(deps)
		for _, dep := range deps {
			if !visited[dep] {
				visited[dep] = true
				cone[dep] = true
				frontier = append(frontier, dep)
			}
		}
	}
	return cone
}
