// Package inspect reports data owned by the frontend, planner and transformers.
package inspect

import (
	"crypto/sha256"
	"dlc/src/ast"
	"dlc/src/lexer"
	"dlc/src/parser"
	"dlc/src/plan"
	"dlc/src/sema"
	"dlc/src/transform"
	"fmt"
)

type Token struct {
	Kind    string `json:"kind"`
	Text    string `json:"text"`
	Span    *Span  `json:"span"`
	Message string `json:"message,omitempty"`
}
type Diagnostic struct {
	Stage    string `json:"stage"`
	Category string `json:"category"`
	Message  string `json:"message"`
	Span     *Span  `json:"span,omitempty"`
}
type Program struct {
	Status  string `json:"status"`
	Printed string `json:"printed"`
	sema.GraphReport
	AST     *Node             `json:"ast"`
	Origins *transform.Report `json:"origins,omitempty"`
}
type PlanStep struct {
	SourceLiteralID string   `json:"source_literal_id"`
	Kind            string   `json:"kind"`
	SourceSpan      *Span    `json:"source_span,omitempty"`
	Literal         *Node    `json:"literal"`
	BoundBefore     []string `json:"bound_before"`
	BoundAfter      []string `json:"bound_after"`
}
type ExecutionPlan struct {
	Program      string     `json:"program"`
	RuleID       string     `json:"rule_id"`
	SourceRuleID string     `json:"source_rule_id,omitempty"`
	Lowered      bool       `json:"lowered"`
	Steps        []PlanStep `json:"steps"`
}
type Report struct {
	SchemaVersion  int               `json:"schema_version"`
	Status         string            `json:"status"`
	SourceSHA256   string            `json:"source_sha256"`
	Tokens         []Token           `json:"tokens"`
	AST            *Node             `json:"ast,omitempty"`
	Diagnostics    []Diagnostic      `json:"diagnostics"`
	Source         *Program          `json:"source,omitempty"`
	Candidate      *Program          `json:"candidate,omitempty"`
	Final          *Program          `json:"final,omitempty"`
	Transform      *transform.Report `json:"transform,omitempty"`
	ExecutionPlans []ExecutionPlan   `json:"execution_plans,omitempty"`
}

func view(p *ast.Program, source bool) *Program {
	g := sema.InspectGraph(p)
	status := "valid"
	if len(g.InvalidSCCs) > 0 {
		status = "unstratifiable"
	}
	return &Program{Status: status, Printed: parser.Print(p), GraphReport: g, AST: program(p, source)}
}
func (r *Report) addDiagnostics(ds []sema.Diagnostic, stage string) {
	for _, d := range ds {
		r.Diagnostics = append(r.Diagnostics, Diagnostic{Stage: stage, Category: string(d.Category), Message: d.Message, Span: span(d.Span)})
	}
}
func Validate(p *ast.Program) ([]sema.Diagnostic, *sema.StratumResult) {
	ds := append(sema.CheckDeclType(p), sema.CheckAllowedness(p)...)
	sd, s := sema.CheckStratification(p)
	return append(ds, sd...), s
}
func Build(src []byte, tr transform.Transformer, mode string) *Report {
	r := &Report{SchemaVersion: 1, Status: "ok", SourceSHA256: fmt.Sprintf("%x", sha256.Sum256(src)), Tokens: []Token{}, Diagnostics: []Diagnostic{}}
	for _, t := range lexer.Tokenize(src) {
		r.Tokens = append(r.Tokens, Token{Kind: t.Kind.String(), Text: t.Text, Span: span(t.Span), Message: t.Message})
	}
	p, errors := parser.Parse(src)
	if len(errors) > 0 {
		r.Status = "parse_error"
		for _, e := range errors {
			r.Diagnostics = append(r.Diagnostics, Diagnostic{Stage: "parse", Category: "parse", Message: e.Message, Span: span(e.Span)})
		}
		return r
	}
	r.AST = program(p, true)
	r.Source = view(p, true)
	ds, strata := Validate(p)
	if len(ds) > 0 {
		r.Status = "rejected"
		r.Source.Status = "rejected"
		r.addDiagnostics(ds, "source")
		return r
	}
	r.ExecutionPlans = plans(p, "source", nil)
	var out *transform.Outcome
	var err error
	if reporting, ok := tr.(transform.ReportingTransformer); ok {
		out, err = reporting.TransformReported(p, strata)
	} else {
		var q *ast.Program
		q, err = tr.Transform(p, strata)
		out = &transform.Outcome{Program: q, Report: transform.IdentityReport(p, mode)}
	}
	if err != nil {
		r.Status = "transform_error"
		r.Diagnostics = append(r.Diagnostics, Diagnostic{Stage: "transform", Category: "transform", Message: err.Error()})
		return r
	}
	r.Transform = out.Report
	if out.Candidate != nil {
		r.Candidate = view(out.Candidate, false)
		setOrigins(r.Candidate, out.CandidateReport)
	}
	r.Final = view(out.Program, false)
	setOrigins(r.Final, out.Report)
	ds, _ = Validate(out.Program)
	if len(ds) > 0 {
		r.Status = "transform_error"
		r.Final.Status = "invalid"
		r.addDiagnostics(ds, "final")
		return r
	}
	r.ExecutionPlans = append(r.ExecutionPlans, plans(out.Program, "final", out.Report)...)
	return r
}
func plans(p *ast.Program, which string, origins *transform.Report) []ExecutionPlan {
	out := []ExecutionPlan{}
	for i, c := range p.Clauses {
		if len(c.Body) == 0 {
			continue
		}
		id := transform.RuleID(i)
		sourceID := id
		directSource := which == "source"
		if origins != nil {
			sourceID = ""
			for _, o := range origins.Rules {
				if o.Index == i {
					id = o.ID
					sourceID = o.SourceRuleID
					directSource = (o.Role == "original" || o.Role == "retained_original" || o.Role == "query_projection") && len(o.ParentRuleIDs) == 0
					break
				}
			}
		}
		body, err := plan.Body(c.Body, nil)
		if err != nil {
			panic(err)
		}
		e := ExecutionPlan{Program: which, RuleID: id, SourceRuleID: sourceID, Lowered: body.Lowered, Steps: []PlanStep{}}
		for j, s := range body.Steps {
			sp := (*Span)(nil)
			if which == "source" {
				sp = span(s.Span)
			}
			literalID := ""
			if directSource && sourceID != "" {
				literalID = fmt.Sprintf("%s/body_%d", sourceID, s.SourceIndex)
			}
			e.Steps = append(e.Steps, PlanStep{SourceLiteralID: literalID, Kind: s.Kind, SourceSpan: sp, Literal: literal(body.Literals[j], fmt.Sprintf("%s/step_%d", id, j), false), BoundBefore: s.BoundBefore, BoundAfter: s.BoundAfter})
		}
		out = append(out, e)
	}
	return out
}

func setOrigins(p *Program, r *transform.Report) {
	p.Origins = r
	if r == nil {
		return
	}
	for i := range p.Dependencies {
		for _, o := range r.Rules {
			if p.Dependencies[i].RuleIndex == o.Index {
				p.Dependencies[i].SourceRuleID = o.SourceRuleID
				break
			}
		}
	}
}
