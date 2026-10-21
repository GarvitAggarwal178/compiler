package compact_test

import (
	"encoding/json"
	"strings"
	"testing"

	"dlc/src/ast"
	"dlc/src/eval"
	"dlc/src/parser"
	"dlc/src/sema"
	"dlc/src/transform"
	"dlc/src/transform/compact"
)

func parsed(t *testing.T, s string) *ast.Program {
	t.Helper()
	p, ds := parser.Parse([]byte(s))
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	d := append(sema.CheckDeclType(p), sema.CheckAllowedness(p)...)
	sd, _ := sema.CheckStratification(p)
	d = append(d, sd...)
	if len(d) > 0 {
		t.Fatal(d)
	}
	return p
}
func apply(t *testing.T, p *ast.Program, names ...string) (*ast.Program, *transform.Report) {
	t.Helper()
	r := transform.IdentityReport(p, "guarded")
	for i := range r.Relations {
		for _, n := range names {
			if r.Relations[i].ID == n {
				r.Relations[i].Role = "checkpoint"
			}
		}
	}
	q, r, err := compact.Apply(p, r)
	if err != nil {
		t.Fatal(err)
	}
	return q, r
}
func removed(r *transform.Report, name string) bool {
	for _, a := range r.Compaction {
		if a.Relation == name && a.Action == "remove" {
			return true
		}
	}
	return false
}
func skip(r *transform.Report, name string) string {
	for _, a := range r.Compaction {
		if a.Relation == name && a.Action == "skip" {
			return a.Reason
		}
	}
	return ""
}
func evaluated(t *testing.T, p *ast.Program, semi bool) map[string]map[string]bool {
	t.Helper()
	st, d := sema.BuildSymbolTable(p)
	if len(d) > 0 {
		t.Fatal(d)
	}
	ds, s := sema.CheckStratification(p)
	if len(ds) > 0 {
		t.Fatal(ds)
	}
	e := eval.NewEvaluator(st.Relations)
	if semi {
		e.RunSemiNaive(p, s)
	} else {
		e.RunNaive(p, s.Stratum)
	}
	out := map[string]map[string]bool{}
	for _, decl := range p.Decls {
		if decl.Kind == ast.DeclOutput {
			rows := map[string]bool{}
			for _, row := range e.Relations[decl.Name].All() {
				b, _ := json.Marshal(row)
				rows[string(b)] = true
			}
			out[decl.Name] = rows
		}
	}
	return out
}
func same(t *testing.T, p, q *ast.Program) {
	t.Helper()
	before, _ := json.Marshal(evaluated(t, p, false))
	for _, semi := range []bool{false, true} {
		after, _ := json.Marshal(evaluated(t, q, semi))
		if string(before) != string(after) {
			t.Fatalf("outputs differ: %s / %s", before, after)
		}
	}
}
func TestAliasesIncludeNegativeUsesZeroArityAndDoNotMutate(t *testing.T) {
	p := parsed(t, `.decl base(x:number) base(1). base(2).
.decl alias(x:number) alias(x):-base(x).
.decl sup_user(x:number) sup_user(x):-base(x).
.decl a(x:number) .output a a(x):-alias(x).
.decl b(x:number) .output b b(x):-base(x),!alias(x).
.decl truth() truth(). .decl zero() zero():-truth().
.decl z() .output z z():-zero().`)
	r := transform.IdentityReport(p, "guarded")
	for i := range r.Relations {
		if r.Relations[i].ID == "alias" || r.Relations[i].ID == "zero" {
			r.Relations[i].Role = "checkpoint"
		}
	}
	oldAST, _ := json.Marshal(p)
	oldReport, _ := json.Marshal(r)
	q, metadata, err := compact.Apply(p, r)
	if err != nil {
		t.Fatal(err)
	}
	if !removed(metadata, "alias") || !removed(metadata, "zero") || skip(metadata, "sup_user") != "source relation" {
		t.Fatal(metadata.Compaction)
	}
	same(t, p, q)
	for _, o := range metadata.Rules {
		if o.Head == "b" && len(o.ParentRuleIDs) == 0 {
			t.Fatal("negative replacement provenance missing")
		}
	}
	q.Clauses[0].Head.Name = "mutated"
	metadata.Decisions[0].RetainedReasons[0] = "mutated"
	afterAST, _ := json.Marshal(p)
	afterReport, _ := json.Marshal(r)
	if string(oldAST) != string(afterAST) || string(oldReport) != string(afterReport) {
		t.Fatal("input AST/metadata aliased or mutated")
	}
}
func TestProjectionPreservingUnfoldingAndVariableHygiene(t *testing.T) {
	p := parsed(t, `.decl base(x:number,y:number) base(1,2). base(2,3).
.decl scratch(x:number,y:number) scratch(x,y):-base(x,y),y=x+1.
.decl keep(x:number) keep(9).
.decl out(x:number,y:number,z:number) .output out
out(x,y,compact_v0):-scratch(y,x),keep(compact_v0).`)
	q, r := apply(t, p, "scratch")
	if !removed(r, "scratch") {
		t.Fatal(r.Compaction)
	}
	same(t, p, q)
	if !strings.Contains(parser.Print(q), "base(y, x)") {
		t.Fatal("simultaneous variable substitution failed", parser.Print(q))
	}
}
func TestIntentionalSkipReasons(t *testing.T) {
	cases := []struct{ name, decl, definition, consumer, reason string }{
		{"projection", ".decl c(x:number)", "c(x):-base(x,y).", "out(x):-c(x).", "existential projection"},
		{"wildcard", ".decl c(x:number)", "c(x):-base(x,_).", "out(x):-c(x).", "wildcard"},
		{"shared", ".decl c(x:number)", "c(x):-one(x),x>0.", "out(x):-c(x),c(x).", "shared"},
		{"negative", ".decl c(x:number)", "c(x):-one(x),x>0.", "out(x):-one(x),!c(x).", "negative"},
		{"repeated_call", ".decl c(x:number,y:number)", "c(x,y):-base(x,y),x>0.", "out(x):-c(x,x).", "repeated call"},
		{"constant_call", ".decl c(x:number,y:number)", "c(x,y):-base(x,y),x>0.", "out(x):-c(1,x).", "constant or repeated call"},
		{"multiple_defs", ".decl c(x:number)", "c(x):-one(x). c(x):-one(x),x>0.", "out(x):-c(x).", "multiple"},
		{"io", ".decl c(x:number) .output c", "c(x):-one(x).", "out(x):-c(x).", "input/output"},
		{"self_recursion", ".decl c(x:number)", "c(x):-c(x),x>0.", "out(0).", "recursive self"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			p := parsed(t, ".decl base(x:number,y:number) base(1,2). .decl one(x:number) one(1). .decl out(x:number) .output out "+c.decl+" "+c.definition+" "+c.consumer)
			q, r := apply(t, p, "c")
			if removed(r, "c") || !strings.Contains(skip(r, "c"), c.reason) {
				t.Fatal(r.Compaction)
			}
			same(t, p, q)
		})
	}
}
func TestRechecksCurrentASTInStableDeclarationOrder(t *testing.T) {
	p := parsed(t, `.decl base(x:number) base(1). .decl first(x:number) .decl second(x:number)
first(x):-second(x). second(x):-base(x). .decl out(x:number) .output out out(x):-first(x).`)
	q, r := apply(t, p, "first", "second")
	names := []string{}
	for _, a := range r.Compaction {
		if a.Action == "remove" {
			names = append(names, a.Relation)
		}
	}
	if strings.Join(names, ",") != "first,second" {
		t.Fatal(names)
	}
	same(t, p, q)
	for _, o := range r.Rules {
		if o.Head == "out" && len(o.ParentRuleIDs) != 2 {
			t.Fatal("chained origins lost", o)
		}
	}
	if _, _, err := compact.Apply(p, nil); err == nil {
		t.Fatal("compacted without provenance")
	}
}
