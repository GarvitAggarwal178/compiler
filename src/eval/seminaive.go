package eval

import (
	"dlc/src/ast"
	"dlc/src/ir"
	"dlc/src/sema"
)

// RunSemiNaive evaluates prog one SCC at a time, in strata.SCCOrder
// (dependency order, not stratum-number order -- see why below), using
// a semi-naive Δ-rewrite within each SCC: a clause whose body has no
// positive atom in the SAME SCC as its own head ("seed" -- everything it
// needs is a different, already-fully-computed SCC) fires exactly once;
// a clause with at least one same-SCC positive atom ("recursive") gets
// one Δ-rewritten variant per same-SCC atom *occurrence* (§3.9: "each
// with one body atom replaced by its delta"), re-evaluated every round
// using only the previous round's newly-derived tuples for that one
// occurrence, until the SCC gains nothing new.
//
// Processing by SCC (via SCCOrder), not by stratum number, was forced by
// a real bug a differential test caught, not chosen in advance:
// example/josephus/josephus.dl has a self-recursive `Relation` and a
// second, independent `Josephus` that reads it positively -- no
// negation anywhere, so both land at stratum 0. Grouping by stratum
// number put Josephus's (correctly seed-classified) clause in the same
// batch as Relation's clauses and ran it during the *seed* round, before
// Relation's own recursion had produced anything beyond its 6 initial
// facts -- silently computing Josephus against a not-yet-converged
// Relation and disagreeing with Soufflé on exactly the tuples Relation's
// recursion still had left to derive. sema/stratify.go's SCCOrder (a
// plain topological order of the SCC condensation) is what evaluation
// order actually needs; stratum number is only for the negation-safety
// rejection check, never for driving evaluation. See DESIGN.md.
func (e *Evaluator) RunSemiNaive(prog *ast.Program, strata *sema.StratumResult) {
	clausesByHead := map[string][]*ast.Clause{}
	for _, c := range prog.Clauses {
		clausesByHead[c.Head.Name] = append(clausesByHead[c.Head.Name], c)
	}
	for _, scc := range strata.SCCOrder {
		var clauses []*ast.Clause
		for _, name := range scc {
			clauses = append(clauses, clausesByHead[name]...)
		}
		if len(clauses) == 0 {
			continue // an SCC of a relation with no clauses at all (shouldn't occur: SCCOrder only contains clause-head relations)
		}
		e.semiNaiveSCC(clauses, strata.SCCOf)
	}
}

// sameSCCAtomIndices returns the indices within body of every positive
// *ast.Atom in the same SCC as headSCC -- the occurrences a Δ-rewrite
// variant must be generated for.
func sameSCCAtomIndices(body []ast.Literal, sccOf map[string]int, headSCC int) []int {
	var idxs []int
	for i, lit := range body {
		if atom, ok := lit.(*ast.Atom); ok {
			if s, known := sccOf[atom.Name]; known && s == headSCC {
				idxs = append(idxs, i)
			}
		}
	}
	return idxs
}

// semiNaiveSCC evaluates every clause whose head is in one SCC (clauses
// must all share the same head SCC -- callers guarantee this). Every
// relation this SCC's clauses read that is NOT itself in the SCC is, by
// SCCOrder's own guarantee, already fully computed and stable.
func (e *Evaluator) semiNaiveSCC(clauses []*ast.Clause, sccOf map[string]int) {
	sccRelNames := map[string]bool{}
	for _, c := range clauses {
		sccRelNames[c.Head.Name] = true
	}
	headSCC := sccOf[clauses[0].Head.Name]

	type prepared struct {
		clause     *ast.Clause
		ordered    []ast.Literal
		recIndices []int // indices (into ordered) of same-SCC positive atoms
	}
	var seed, recursive []prepared
	for _, c := range clauses {
		ordered := safeOrder(c.Body)
		idxs := sameSCCAtomIndices(ordered, sccOf, headSCC)
		p := prepared{clause: c, ordered: ordered, recIndices: idxs}
		if len(idxs) == 0 {
			seed = append(seed, p)
		} else {
			recursive = append(recursive, p)
		}
	}

	delta := map[string]*ir.Relation{}
	for name := range sccRelNames {
		delta[name] = ir.NewRelation(name, e.Relations[name].Arity)
	}

	// Seed round: every relation in this SCC starts empty (nothing
	// outside the SCC could have contributed to it -- only same-SCC
	// recursive clauses can, and those haven't run yet), so every tuple
	// a seed clause derives here is new by definition.
	for _, p := range seed {
		headRel := e.Relations[p.clause.Head.Name]
		e.evalBody(p.ordered, 0, map[string]ir.Value{}, nil, func(bindings map[string]ir.Value) {
			tup, ok := e.buildTuple(p.clause.Head, bindings)
			if !ok {
				return
			}
			e.DerivationAttempts++
			if headRel.Insert(tup) {
				headRel.RecordSeedInsert()
				delta[p.clause.Head.Name].Insert(tup)
			}
		})
	}

	if len(recursive) == 0 {
		return // singleton, non-recursive SCC -- the seed round is the whole answer
	}

	for iteration := 0; ; iteration++ {
		anyDelta := false
		for _, d := range delta {
			if d.Len() > 0 {
				anyDelta = true
				break
			}
		}
		if !anyDelta {
			return
		}

		newDelta := map[string]*ir.Relation{}
		for name := range sccRelNames {
			newDelta[name] = ir.NewRelation(name, e.Relations[name].Arity)
		}

		for _, p := range recursive {
			headRel := e.Relations[p.clause.Head.Name]
			for _, atomIdx := range p.recIndices {
				atomName := p.ordered[atomIdx].(*ast.Atom).Name
				overrideIdx := map[int]*ir.Relation{atomIdx: delta[atomName]}
				e.evalBody(p.ordered, 0, map[string]ir.Value{}, overrideIdx, func(bindings map[string]ir.Value) {
					tup, ok := e.buildTuple(p.clause.Head, bindings)
					if !ok {
						return
					}
					e.DerivationAttempts++
					if headRel.Insert(tup) {
						headRel.RecordIterationInsert(iteration)
						newDelta[p.clause.Head.Name].Insert(tup)
					}
				})
			}
		}
		delta = newDelta
	}
}
