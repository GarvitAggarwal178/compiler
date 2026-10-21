package sema

import (
	"dlc/src/ast"
	"dlc/src/token"
	"fmt"
)

type RelationParam struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type RelationInfo struct {
	Name    string          `json:"name"`
	Params  []RelationParam `json:"params"`
	Input   bool            `json:"input"`
	Output  bool            `json:"output"`
	IDB     bool            `json:"idb"`
	SCC     *int            `json:"scc,omitempty"`
	Stratum *int            `json:"stratum,omitempty"`
}
type Dependency struct {
	RuleIndex    int         `json:"rule_index"`
	SourceRuleID string      `json:"source_rule_id,omitempty"`
	From         string      `json:"from"`
	To           string      `json:"to"`
	Negative     bool        `json:"negative"`
	RuleID       string      `json:"rule_id"`
	LiteralID    string      `json:"literal_id"`
	Span         *token.Span `json:"span,omitempty"`
}
type GraphReport struct {
	Relations    []RelationInfo `json:"relations"`
	Dependencies []Dependency   `json:"dependencies"`
	SCCs         [][]string     `json:"sccs"`
	SCCOrder     [][]string     `json:"scc_order,omitempty"`
	InvalidSCCs  [][]string     `json:"invalid_sccs"`
}

// InspectGraph exports the existing IDB SCC analysis and all actual source
// dependencies. Input-only nodes have no fabricated IDB SCC or stratum.
func InspectGraph(p *ast.Program) GraphReport {
	g := buildPrecedenceGraph(p)
	of, members := tarjanSCC(g)
	out := GraphReport{Relations: []RelationInfo{}, Dependencies: []Dependency{}, SCCs: members, InvalidSCCs: [][]string{}}
	invalid := map[int]bool{}
	for _, from := range sortedGraphNames(g) {
		for _, e := range sortedEdges(g.adjacency[from]) {
			if e.negative && of[from] == of[e.to] {
				invalid[of[from]] = true
			}
		}
	}
	for i, m := range members {
		if invalid[i] {
			out.InvalidSCCs = append(out.InvalidSCCs, m)
		}
	}
	var strata map[string]int
	if len(invalid) == 0 {
		strata = computeStrata(g, of, members)
		out.SCCOrder = sccTopoOrder(g, of, members)
	}
	inputs, outputs := map[string]bool{}, map[string]bool{}
	for _, d := range p.Decls {
		if d.Kind == ast.DeclInput {
			inputs[d.Name] = true
		}
		if d.Kind == ast.DeclOutput {
			outputs[d.Name] = true
		}
	}
	for _, d := range p.Decls {
		if d.Kind != ast.DeclRelation {
			continue
		}
		r := RelationInfo{Name: d.Name, Params: []RelationParam{}, Input: inputs[d.Name], Output: outputs[d.Name], IDB: g.nodes[d.Name]}
		if r.IDB {
			i := of[d.Name]
			r.SCC = &i
			if strata != nil {
				v := strata[d.Name]
				r.Stratum = &v
			}
		}
		for _, p := range d.Params {
			r.Params = append(r.Params, RelationParam{Name: p.Name, Type: p.Type})
		}
		out.Relations = append(out.Relations, r)
	}
	for i, c := range p.Clauses {
		for j, l := range c.Body {
			var name string
			neg := false
			switch v := l.(type) {
			case *ast.Atom:
				name = v.Name
			case *ast.NegatedAtom:
				name = v.Atom.Name
				neg = true
			default:
				continue
			}
			out.Dependencies = append(out.Dependencies, Dependency{RuleIndex: i, From: c.Head.Name, To: name, Negative: neg, RuleID: fmt.Sprintf("rule_%d", i), LiteralID: fmt.Sprintf("rule_%d/body_%d", i, j), Span: nonzeroSpan(l.Span())})
		}
	}
	return out
}

func nonzeroSpan(s token.Span) *token.Span {
	if s.Start.Line == 0 {
		return nil
	}
	return &s
}
