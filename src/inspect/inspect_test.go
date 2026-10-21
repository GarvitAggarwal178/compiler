package inspect_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"dlc/src/inspect"
	"dlc/src/transform/guard"
)

const source = `.decl edge(a:number,b:number)
.input edge
.decl reach(a:number,b:number)
.decl out(a:number)
.output out
reach(x,y) :- edge(x,y).
reach(x,y) :- edge(x,z), reach(z,y).
out(y) :- reach(-1,y).
`

func report(s string) *inspect.Report {
	return inspect.Build([]byte(s), guard.Transformer{}, "guarded")
}
func TestSourceAndGenerationProvenance(t *testing.T) {
	r := report(source + ".output reach\n.decl reach_bf(x:number,y:number)\nreach_bf(99,99).\n")
	if r.Status != "ok" {
		t.Fatal(r.Diagnostics)
	}
	if r.AST.Kind != "Program" || r.AST.Clauses[0].ID != "rule_0" || r.AST.Clauses[0].Span == nil {
		t.Fatal("source AST or spans missing")
	}
	if len(r.Transform.Queries) != 1 || r.Transform.Queries[0].DemandArguments[0].Expression != "-1" {
		t.Fatal(r.Transform.Queries)
	}
	view := r.Transform.Queries[0].ViewID
	if view == "reach_bf" {
		t.Fatal("user name overwritten")
	}
	foundEdge := false
	for _, e := range r.Source.Dependencies {
		if e.From == "reach" && e.To == "edge" && !e.Negative {
			foundEdge = true
		}
	}
	if !foundEdge {
		t.Fatal("actual input dependency missing")
	}
	for _, rel := range r.Source.Relations {
		if rel.Name == "edge" && rel.SCC != nil {
			t.Fatal("fabricated input SCC")
		}
	}
	retained, wrapper := false, false
	for _, d := range r.Transform.Decisions {
		if d.SourcePredicate == "reach" {
			retained = d.Specialized && d.FullExtentRetained
		}
		if d.SourcePredicate == "out" {
			wrapper = !d.Specialized
		}
	}
	if !retained || !wrapper {
		t.Fatal(r.Transform.Decisions)
	}
	checkpoints, seeds := 0, 0
	for _, r := range r.Transform.Relations {
		if r.Role == "checkpoint" {
			checkpoints++
		}
	}
	for _, o := range r.Transform.Rules {
		if o.Role == "query_seed" {
			seeds++
			if o.SourceRuleID != "rule_2" {
				t.Fatal(o)
			}
		}
	}
	if checkpoints == 0 || seeds != 1 {
		t.Fatal("missing generation metadata")
	}
	if r.Final.AST.Clauses[0].Span != nil {
		t.Fatal("generated source span fabricated")
	}
}
func TestMixedFallbackCandidateAndFinal(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", "tests", "corpus", "CONE_CORPUS", "cc_growing_sibling.dl"))
	if err != nil {
		t.Fatal(err)
	}
	r := report(string(b))
	if r.Status != "ok" || r.Source.Status != "valid" || r.Candidate.Status != "unstratifiable" || r.Final.Status != "valid" {
		t.Fatalf("statuses: %+v", r)
	}
	tc, fallback := false, false
	for _, d := range r.Transform.Decisions {
		if d.SourcePredicate == "tc" {
			tc = d.Specialized && !d.Fallback
		}
		if d.SourcePredicate == "p" {
			fallback = d.Fallback && d.FallbackReason != ""
		}
	}
	if !tc || !fallback {
		t.Fatal(r.Transform.Decisions)
	}
}
func TestInvalidNoQueryAndStableReports(t *testing.T) {
	invalid := report(".decl p(x:number) p(x) :- p(x+1).")
	if invalid.Status != "rejected" || invalid.Final != nil || len(invalid.Diagnostics) == 0 {
		t.Fatal(invalid)
	}
	bad := report(".decl")
	if bad.Status != "parse_error" || bad.Final != nil {
		t.Fatal(bad)
	}
	no := report(".decl p(x:number) p(1).")
	if no.Status != "ok" || no.Candidate != nil || no.Transform.NoQueryReason == "" || len(no.Transform.Queries) != 0 {
		t.Fatal(no)
	}
	first, _ := json.Marshal(report(source))
	for i := 0; i < 20; i++ {
		report(".decl a() a().")
		b, _ := json.Marshal(report(source))
		if !bytes.Equal(first, b) {
			t.Fatal("report varies or leaked state")
		}
	}
}
func TestArithmeticPlanReportedWithoutMutatingSource(t *testing.T) {
	s := `.decl p(x:number,y:number) .decl q(x:number,y:number) .decl out(x:number,y:number) .output out
p(2,2). q(3,1). out(x,y) :- p(x+1,y),q(y+1,x).`
	r := report(s)
	if r.Status != "ok" {
		t.Fatal(r.Diagnostics)
	}
	c := r.AST.Clauses[2]
	if c.Body[0].Terms[0].Kind != "BinaryExpr" {
		t.Fatal("source was lowered in place")
	}
	found := false
	for _, p := range r.ExecutionPlans {
		if p.Program == "source" && p.RuleID == "rule_2" {
			if !p.Lowered || len(p.Steps) != 4 {
				t.Fatal(p)
			}
			for _, s := range p.Steps {
				if s.Kind == "arithmetic_filter" {
					found = true
				}
			}
		}
	}
	if !found {
		t.Fatal("real lowered plan not reported")
	}
}

func TestFinalPlansOnlyClaimKnownSourceOrigins(t *testing.T) {
	r := report(source)
	for _, p := range r.ExecutionPlans {
		if p.Program != "final" {
			continue
		}
		for _, s := range p.Steps {
			if s.SourceSpan != nil {
				t.Fatal("generated plan claimed a source span")
			}
			if strings.HasPrefix(p.RuleID, "checkpoint_step/") && s.SourceLiteralID != "" {
				t.Fatal("checkpoint wrapper misattributed to a source literal")
			}
		}
	}
	for _, e := range r.Final.Dependencies {
		if e.SourceRuleID == "" {
			t.Fatalf("missing known rule origin: %+v", e)
		}
	}
}
