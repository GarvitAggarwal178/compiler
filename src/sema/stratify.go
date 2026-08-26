package sema

import (
	"fmt"
	"sort"

	"dlc/src/ast"
	"dlc/src/token"
)

// Unstratifiable is the category for the one diagnostic this file
// produces.
const Unstratifiable Category = "unstratifiable"

// StratumResult is the accepted outcome: a stratum number (0-based, in
// evaluation order) per IDB relation name; which SCC each relation
// belongs to (SCCOf); and SCCOrder, every SCC's own relation-name
// members, in a valid evaluation order (index 0 first -- every relation
// an SCC depends on, by any edge, belongs to an earlier SCC in this
// list).
//
// SCCOrder exists because stratum number is too coarse a grouping for
// §3.9's semi-naive evaluator: two SCCs can share a stratum number
// without one depending on the other at all having been computed yet
// relative to a THIRD, unrelated SCC also in that stratum (found by a
// failing test, eval/DESIGN.md -- a plain relation and something that
// only reads it, both stratum 0 with nothing forcing either order
// between them and a third independent relation, but the reader still
// needs the read-from relation fully computed first). A plain
// topological order of the SCC condensation (this field) is what
// evaluation order actually needs; stratum number alone is only ever
// needed for the negation-safety rejection check above, not for driving
// evaluation.
type StratumResult struct {
	Stratum  map[string]int
	SCCOf    map[string]int
	SCCOrder [][]string
}

// edge is one precedence-graph edge: `from` (a clause head, an IDB
// relation) depends on `to` (a relation referenced in that clause's
// body), Negative if the reference was a negated atom.
type edge struct {
	to       string
	negative bool
	span     token.Span // the negated atom's span, for a diagnostic if this edge closes an unstratifiable cycle
}

// CheckStratification builds the precedence graph over IDB relations
// (relations that are the head of at least one clause -- EDB/input-only
// relations have no rules and are not nodes here), finds its SCCs via
// Tarjan, and rejects (one Unstratifiable Diagnostic) if any SCC
// contains a negative edge between two of its own members. On success,
// returns a stratum assignment instead.
//
// This is the SOURCE program's stratifier (§3.6, Lane B). The
// transformed-program culprit-cycle detector (Lane A,
// docs/M1-BUILD.md §1) is a different pass over different input and is
// not generalized from this one -- see DESIGN.md.
func CheckStratification(prog *ast.Program) ([]Diagnostic, *StratumResult) {
	g := buildPrecedenceGraph(prog)
	sccOf, sccMembers := tarjanSCC(g)

	for from, edges := range g.adjacency {
		for _, e := range edges {
			if e.negative && sccOf[from] == sccOf[e.to] {
				members := append([]string{}, sccMembers[sccOf[from]]...)
				sort.Strings(members)
				return []Diagnostic{{
					Span: e.span, Category: Unstratifiable,
					Message: fmt.Sprintf("relation %q has a negative dependency on %q, and both are in the same "+
						"mutual-recursion cycle {%v} -- no valid stratum ordering exists", from, e.to, members),
				}}, nil
			}
		}
	}

	stratum := computeStrata(g, sccOf, sccMembers)
	sccOrder := sccTopoOrder(g, sccOf, sccMembers)
	return nil, &StratumResult{Stratum: stratum, SCCOf: sccOf, SCCOrder: sccOrder}
}

// sccTopoOrder returns every SCC's members (sorted, for determinism),
// ordered so that every SCC a given SCC depends on (by any edge, positive
// or negative) appears earlier in the returned slice -- a plain
// post-order DFS over the (acyclic, by construction) SCC condensation.
func sccTopoOrder(g *graph, sccOf map[string]int, sccMembers [][]string) [][]string {
	visited := make([]bool, len(sccMembers))
	var order []int
	var visit func(idx int)
	visit = func(idx int) {
		if visited[idx] {
			return
		}
		visited[idx] = true
		members := append([]string{}, sccMembers[idx]...)
		sort.Strings(members)
		for _, m := range members {
			edges := append([]edge{}, g.adjacency[m]...)
			sort.Slice(edges, func(i, j int) bool { return edges[i].to < edges[j].to })
			for _, e := range edges {
				if t := sccOf[e.to]; t != idx {
					visit(t)
				}
			}
		}
		order = append(order, idx)
	}
	for i := range sccMembers {
		visit(i)
	}
	out := make([][]string, len(order))
	for i, idx := range order {
		members := append([]string{}, sccMembers[idx]...)
		sort.Strings(members)
		out[i] = members
	}
	return out
}

type graph struct {
	nodes     map[string]bool
	adjacency map[string][]edge
}

func buildPrecedenceGraph(prog *ast.Program) *graph {
	g := &graph{nodes: map[string]bool{}, adjacency: map[string][]edge{}}
	for _, c := range prog.Clauses {
		g.nodes[c.Head.Name] = true
	}
	for _, c := range prog.Clauses {
		head := c.Head.Name
		for _, lit := range c.Body {
			switch v := lit.(type) {
			case *ast.Atom:
				if g.nodes[v.Name] {
					g.adjacency[head] = append(g.adjacency[head], edge{to: v.Name, negative: false, span: v.Sp})
				}
			case *ast.NegatedAtom:
				if g.nodes[v.Atom.Name] {
					g.adjacency[head] = append(g.adjacency[head], edge{to: v.Atom.Name, negative: true, span: v.Sp})
				}
			}
		}
	}
	return g
}

// tarjanSCC returns, for every node, its SCC index, and the list of
// node names per SCC index. Standard Tarjan, iterative would be nicer
// for very deep recursion but the corpus this runs over (blueprint §4
// grammar, no arbitrary nesting) never approaches stack-depth trouble.
func tarjanSCC(g *graph) (map[string]int, [][]string) {
	index := 0
	indices := map[string]int{}
	lowlink := map[string]int{}
	onStack := map[string]bool{}
	var stack []string
	sccOf := map[string]int{}
	var sccMembers [][]string

	var names []string
	for n := range g.nodes {
		names = append(names, n)
	}
	sort.Strings(names) // deterministic iteration order -- required by CLAUDE.md's determinism rule

	var strongconnect func(v string)
	strongconnect = func(v string) {
		indices[v] = index
		lowlink[v] = index
		index++
		stack = append(stack, v)
		onStack[v] = true

		edges := append([]edge{}, g.adjacency[v]...)
		sort.Slice(edges, func(i, j int) bool { return edges[i].to < edges[j].to })
		for _, e := range edges {
			w := e.to
			if _, seen := indices[w]; !seen {
				strongconnect(w)
				if lowlink[w] < lowlink[v] {
					lowlink[v] = lowlink[w]
				}
			} else if onStack[w] {
				if indices[w] < lowlink[v] {
					lowlink[v] = indices[w]
				}
			}
		}

		if lowlink[v] == indices[v] {
			sccIdx := len(sccMembers)
			var members []string
			for {
				n := len(stack) - 1
				w := stack[n]
				stack = stack[:n]
				onStack[w] = false
				sccOf[w] = sccIdx
				members = append(members, w)
				if w == v {
					break
				}
			}
			sccMembers = append(sccMembers, members)
		}
	}

	for _, v := range names {
		if _, seen := indices[v]; !seen {
			strongconnect(v)
		}
	}
	return sccOf, sccMembers
}

// computeStrata assigns each SCC a stratum via memoized DFS over the
// condensation DAG (guaranteed acyclic -- Tarjan guarantees no cycles
// between distinct SCCs by construction): a relation's stratum is one
// more than the highest stratum of anything it negatively depends on,
// and at least as high as the stratum of anything it positively depends
// on. All members of one SCC share the SCC's stratum.
func computeStrata(g *graph, sccOf map[string]int, sccMembers [][]string) map[string]int {
	stratumOfSCC := make([]int, len(sccMembers))
	computed := make([]bool, len(sccMembers))

	var visit func(idx int) int
	visit = func(idx int) int {
		if computed[idx] {
			return stratumOfSCC[idx]
		}
		computed[idx] = true
		maxStratum := 0
		for _, member := range sccMembers[idx] {
			for _, e := range g.adjacency[member] {
				targetIdx := sccOf[e.to]
				if targetIdx == idx {
					continue // internal edge: doesn't affect this SCC's stratum relative to itself
				}
				s := visit(targetIdx)
				if e.negative {
					s++
				}
				if s > maxStratum {
					maxStratum = s
				}
			}
		}
		stratumOfSCC[idx] = maxStratum
		return maxStratum
	}
	for i := range sccMembers {
		visit(i)
	}

	result := map[string]int{}
	for idx, members := range sccMembers {
		for _, m := range members {
			result[m] = stratumOfSCC[idx]
		}
	}
	return result
}
