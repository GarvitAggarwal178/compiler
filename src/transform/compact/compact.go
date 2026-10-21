// Package compact removes only proven identity aliases and projection-preserving
// single-use generated checkpoints from an already validated guarded program.
package compact

import (
	"encoding/json"
	"fmt"

	"dlc/src/ast"
	"dlc/src/parser"
	"dlc/src/plan"
	"dlc/src/sema"
	"dlc/src/transform"
	"dlc/src/transform/guard"
)

type use struct {
	clause   *ast.Clause
	position int
	negative bool
}
type removal struct {
	definition *ast.Clause
	alias      string
	consumer   *ast.Clause
	position   int
	names      map[string]string
	reason     string
	uses       []use
}

func distinctVars(terms []ast.Term) ([]string, bool) {
	out := []string{}
	seen := map[string]bool{}
	for _, t := range terms {
		v, ok := t.(*ast.Var)
		if !ok || seen[v.Name] {
			return nil, false
		}
		seen[v.Name] = true
		out = append(out, v.Name)
	}
	return out, true
}
func eligible(p *ast.Program, decl *ast.Decl, roles map[string]string, schemas map[string]*sema.RelationSchema) (*removal, string) {
	name := decl.Name
	if roles[name] == "source" {
		return nil, "source relation"
	}
	if roles[name] != "checkpoint" {
		return nil, "not a generated checkpoint"
	}
	for _, d := range p.Decls {
		if d.Name == name && d.Kind != ast.DeclRelation {
			return nil, "input/output designation"
		}
	}
	defs := []*ast.Clause{}
	uses := []use{}
	for _, c := range p.Clauses {
		if c.Head.Name == name {
			defs = append(defs, c)
		}
		for i, l := range c.Body {
			switch v := l.(type) {
			case *ast.Atom:
				if v.Name == name {
					uses = append(uses, use{c, i, false})
				}
			case *ast.NegatedAtom:
				if v.Atom.Name == name {
					uses = append(uses, use{c, i, true})
				}
			}
		}
	}
	if len(defs) != 1 {
		return nil, "multiple or missing definitions"
	}
	def := defs[0]
	if len(def.Body) == 0 {
		return nil, "fact definition"
	}
	head, ok := distinctVars(def.Head.Terms)
	if !ok {
		return nil, "constant or repeated head argument"
	}
	if len(def.Body) == 1 {
		if a, ok := def.Body[0].(*ast.Atom); ok && a.Name != name && len(a.Terms) == len(head) {
			identity := true
			for i, t := range a.Terms {
				v, ok := t.(*ast.Var)
				if !ok || v.Name != head[i] {
					identity = false
					break
				}
			}
			sameSchema := schemas[a.Name] != nil && len(schemas[a.Name].Params) == len(schemas[name].Params)
			if sameSchema {
				for i, p := range schemas[a.Name].Params {
					if p.Type != schemas[name].Params[i].Type {
						sameSchema = false
						break
					}
				}
			}
			if identity && sameSchema {
				return &removal{definition: def, alias: a.Name, reason: "exact identity alias", uses: uses}, ""
			}
		}
	}
	for _, u := range uses {
		if u.negative {
			return nil, "negative use of non-alias"
		}
	}
	if len(uses) != 1 {
		return nil, "shared or unused non-alias"
	}
	u := uses[0]
	if u.clause == def {
		return nil, "recursive self consumer"
	}
	call := u.clause.Body[u.position].(*ast.Atom)
	args, ok := distinctVars(call.Terms)
	if !ok || len(args) != len(head) {
		return nil, "constant or repeated call argument"
	}
	headSet := map[string]bool{}
	for _, v := range head {
		headSet[v] = true
	}
	bodyVars := map[string]bool{}
	for _, l := range def.Body {
		literalVars(l, bodyVars)
		if a, ok := l.(*ast.Atom); ok {
			for _, t := range a.Terms {
				if _, wild := t.(*ast.Wildcard); wild {
					return nil, "positive wildcard preserves dedup barrier"
				}
			}
		}
	}
	for name := range bodyVars {
		if !headSet[name] {
			return nil, "existential projection preserves dedup barrier"
		}
	}
	names := map[string]string{}
	for i, v := range head {
		names[v] = args[i]
	}
	return &removal{definition: def, consumer: u.clause, position: u.position, names: names, reason: "projection-preserving single positive use", uses: uses}, ""
}

// Apply is pure with respect to both input AST and public metadata. Generation
// roles are mandatory: a user name beginning sup_ has no special meaning here.
func Apply(input *ast.Program, metadata *transform.Report) (*ast.Program, *transform.Report, error) {
	if metadata == nil {
		return nil, nil, fmt.Errorf("compact: explicit generation metadata required")
	}
	raw, err := json.Marshal(metadata)
	if err != nil {
		return nil, nil, err
	}
	report := &transform.Report{}
	if err = json.Unmarshal(raw, report); err != nil {
		return nil, nil, err
	}
	report.Mode = "guarded-compact"
	report.Compaction = []transform.CompactAction{}
	p := cloneProgram(input)
	roles := map[string]string{}
	for _, r := range report.Relations {
		roles[r.ID] = r.Role
	}
	origins := map[*ast.Clause]transform.RuleOrigin{}
	for i, c := range p.Clauses {
		origin := transform.RuleOrigin{ID: transform.RuleID(i), Head: c.Head.Name, Role: "original", SourceRuleID: transform.RuleID(i)}
		for _, r := range report.Rules {
			if r.Index == i {
				origin = r
				break
			}
		}
		origins[c] = origin
	}
	for {
		st, diags := sema.BuildSymbolTable(p)
		if len(diags) > 0 {
			return nil, nil, fmt.Errorf("compact input schema: %v", diags)
		}
		changed := false
		for _, d := range p.Decls {
			if d.Kind != ast.DeclRelation {
				continue
			}
			step, _ := eligible(p, d, roles, st.Relations)
			if step == nil {
				continue
			}
			defOrigin := origins[step.definition]
			consumers := []string{}
			touched := map[*ast.Clause]bool{}
			if step.alias != "" {
				for _, u := range step.uses {
					if u.clause == step.definition {
						continue
					}
					switch a := u.clause.Body[u.position].(type) {
					case *ast.Atom:
						a.Name = step.alias
					case *ast.NegatedAtom:
						a.Atom.Name = step.alias
					}
					touched[u.clause] = true
				}
			} else {
				body := append([]ast.Literal{}, step.consumer.Body[:step.position]...)
				for _, l := range step.definition.Body {
					body = append(body, cloneLiteral(l, step.names))
				}
				body = append(body, step.consumer.Body[step.position+1:]...)
				step.consumer.Body = body
				touched[step.consumer] = true
			}
			// Source clause order, not map order, determines provenance and scheduling.
			clauses := []*ast.Clause{}
			for _, c := range p.Clauses {
				if c == step.definition {
					continue
				}
				if touched[c] {
					planned, err := plan.Body(c.Body, nil)
					if err != nil {
						return nil, nil, fmt.Errorf("compact reschedule %s: %w", origins[c].ID, err)
					}
					c.Body = planned.Literals
					o := origins[c]
					o.ParentRuleIDs = appendUnique(o.ParentRuleIDs, defOrigin.ID)
					origins[c] = o
					consumers = append(consumers, o.ID)
				}
				clauses = append(clauses, c)
			}
			p.Clauses = clauses
			decls := []*ast.Decl{}
			for _, decl := range p.Decls {
				if decl.Name != d.Name {
					decls = append(decls, decl)
				}
			}
			p.Decls = decls
			delete(roles, d.Name)
			report.Compaction = append(report.Compaction, transform.CompactAction{Relation: d.Name, Action: "remove", Reason: step.reason, DefinitionID: defOrigin.ID, SourceRuleID: defOrigin.SourceRuleID, ConsumerIDs: consumers})
			changed = true
			break
		}
		if !changed {
			for _, d := range p.Decls {
				if d.Kind != ast.DeclRelation {
					continue
				}
				_, reason := eligible(p, d, roles, st.Relations)
				report.Compaction = append(report.Compaction, transform.CompactAction{Relation: d.Name, Action: "skip", Reason: reason})
			}
			break
		}
	}
	report.Relations = nil
	for _, r := range metadata.Relations {
		if _, ok := roles[r.ID]; ok {
			report.Relations = append(report.Relations, r)
		}
	}
	report.Rules = []transform.RuleOrigin{}
	for i, c := range p.Clauses {
		o := origins[c]
		o.Index = i
		o.Text = parser.Print(&ast.Program{Clauses: []*ast.Clause{c}})
		report.Rules = append(report.Rules, o)
	}
	ds := append(sema.CheckDeclType(p), sema.CheckAllowedness(p)...)
	sd, _ := sema.CheckStratification(p)
	ds = append(ds, sd...)
	if len(ds) > 0 {
		return nil, nil, fmt.Errorf("compacted program failed validation: %v", ds)
	}
	return p, report, nil
}
func appendUnique(values []string, value string) []string {
	for _, v := range values {
		if v == value {
			return values
		}
	}
	return append(values, value)
}

type Transformer struct{}

func (t Transformer) Transform(p *ast.Program, s *sema.StratumResult) (*ast.Program, error) {
	out, err := t.TransformReported(p, s)
	if err != nil {
		return nil, err
	}
	return out.Program, nil
}
func (Transformer) TransformReported(p *ast.Program, s *sema.StratumResult) (*transform.Outcome, error) {
	guarded, err := (guard.Transformer{}).TransformReported(p, s)
	if err != nil {
		return nil, err
	}
	ds := append(sema.CheckDeclType(guarded.Program), sema.CheckAllowedness(guarded.Program)...)
	sd, _ := sema.CheckStratification(guarded.Program)
	ds = append(ds, sd...)
	if len(ds) > 0 {
		return nil, fmt.Errorf("compact requires validated guarded output: %v", ds)
	}
	q, report, err := Apply(guarded.Program, guarded.Report)
	if err != nil {
		return nil, err
	}
	return &transform.Outcome{Program: q, Candidate: guarded.Candidate, CandidateReport: guarded.CandidateReport, Report: report}, nil
}
