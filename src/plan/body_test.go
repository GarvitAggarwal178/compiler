package plan_test

import (
	"dlc/src/ast"
	"dlc/src/codegen"
	"dlc/src/eval"
	"dlc/src/parser"
	"dlc/src/plan"
	"dlc/src/sema"
	"dlc/src/transform/guard"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
)

const cycle = `.decl p(a:number,b:number)
.decl q(a:number,b:number)
.decl out(a:number,b:number)
.output out
p(2,2).
q(3,1).
out(x,y) :- p(x+1,y),q(y+1,x).
`

func parsed(t *testing.T, s string) *ast.Program {
	t.Helper()
	p, d := parser.Parse([]byte(s))
	if len(d) > 0 {
		t.Fatal(d)
	}
	return p
}

func TestRepairAcrossExistingBackends(t *testing.T) {
	p := parsed(t, cycle)
	original := parser.Print(p)
	for _, c := range p.Clauses {
		r, e := plan.Body(c.Body, nil)
		if e != nil {
			t.Fatal(e)
		}
		if len(c.Body) > 0 {
			if !r.Lowered {
				t.Fatal("cycle was not lowered")
			}
			if len(r.Literals) != 4 {
				t.Fatal("expected two atoms and two filters")
			}
		}
		c.Body = r.Literals
	}
	if original == parser.Print(p) {
		t.Fatal("repair missing")
	}
	if d := sema.CheckAllowedness(p); len(d) > 0 {
		t.Fatal(d)
	}
	for _, guarded := range []bool{false, true} {
		q := p
		sd, s := sema.CheckStratification(q)
		if len(sd) > 0 {
			t.Fatal(sd)
		}
		if guarded {
			var e error
			q, e = (guard.Transformer{}).Transform(p, s)
			if e != nil {
				t.Fatal(e)
			}
		}
		st, d := sema.BuildSymbolTable(q)
		if len(d) > 0 {
			t.Fatal(d)
		}
		sd, s = sema.CheckStratification(q)
		if len(sd) > 0 {
			t.Fatal(sd)
		}
		for _, semi := range []bool{false, true} {
			e := eval.NewEvaluator(st.Relations)
			if semi {
				e.RunSemiNaive(q, s)
			} else {
				e.RunNaive(q, s.Stratum)
			}
			out := t.TempDir()
			if err := e.WriteOutput(out, st.Relations, map[string]bool{"out": true}); err != nil {
				t.Fatal(err)
			}
			b, _ := os.ReadFile(filepath.Join(out, "out.csv"))
			if string(b) != "1\t2\n" {
				t.Fatalf("guarded=%v semi=%v got %q", guarded, semi, b)
			}
		}
		cc, err := exec.LookPath("cc")
		if err != nil {
			t.Fatal(err)
		}
		dir := t.TempDir()
		cf := filepath.Join(dir, "p.c")
		binary := filepath.Join(dir, "p")
		out := filepath.Join(dir, "out")
		if err := os.WriteFile(cf, []byte(codegen.Generate(q, st.Relations, s)), 0600); err != nil {
			t.Fatal(err)
		}
		if b, err := exec.Command(cc, "-std=c11", "-O2", cf, "-o", binary).CombinedOutput(); err != nil {
			t.Fatalf("compile: %v %s", err, b)
		}
		if b, err := exec.Command(binary, dir, out).CombinedOutput(); err != nil {
			t.Fatalf("run: %v %s", err, b)
		}
		b, _ := os.ReadFile(filepath.Join(out, "out.csv"))
		if string(b) != "1\t2\n" {
			t.Fatalf("C got %q", b)
		}
	}
}
func TestStableOriginalOrderAndNoMutation(t *testing.T) {
	sources := []string{
		`.decl p(x:number,y:number) .decl out(x:number) out(x) :- p(x+1,x).`,
		`.decl p(x:number) .decl q(x:number) .decl out(x:number) out(x) :- p(x+1),q(x).`,
		`.decl p(x:number) .decl q(x:number) .decl out(x:number) out(x) :- !q(x),p(x).`,
		`.decl out(x:number) out(x) :- x=y+1,y=2.`,
	}
	for _, src := range sources {
		p := parsed(t, src)
		before := parser.Print(p)
		c := p.Clauses[len(p.Clauses)-1]
		a, e := plan.Body(c.Body, nil)
		if e != nil {
			t.Fatal(e)
		}
		b, e := plan.Body(c.Body, nil)
		if e != nil {
			t.Fatal(e)
		}
		if a.Lowered {
			t.Fatalf("unnecessary lowering: %s", src)
		}
		if !reflect.DeepEqual(a, b) || parser.Print(p) != before {
			t.Fatal("unstable or mutated")
		}
	}
}
func TestFreshNamesAndRejection(t *testing.T) {
	p := parsed(t, `.decl out(x:number,y:number) out(x,y) :- p(x+1,y,dlc_expr_0),q(y+1,x).`)
	r, e := plan.Body(p.Clauses[0].Body, nil)
	if e != nil {
		t.Fatal(e)
	}
	if !r.Lowered {
		t.Fatal("not lowered")
	}
	first := r.Literals[0].(*ast.Atom).Terms[0].(*ast.Var).Name
	if first == "dlc_expr_0" {
		t.Fatal("generated variable collision")
	}
	p = parsed(t, `.decl out(x:number) out(x) :- p(x+1).`)
	if _, e := plan.Body(p.Clauses[0].Body, nil); e == nil {
		t.Fatal("unsupported inversion accepted")
	}
}
