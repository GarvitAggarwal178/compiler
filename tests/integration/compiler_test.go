package integration_test

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"dlc/src/ast"
	"dlc/src/codegen"
	"dlc/src/eval"
	"dlc/src/parser"
	"dlc/src/sema"
	"dlc/src/transform"
	"dlc/src/transform/compact"
	"dlc/src/transform/guard"
	"dlc/src/transform/magicset"
)

const reachProgram = `.decl edge(a:number,b:number)
.decl reach(a:number,b:number)
.decl out(a:number)
.output out
edge(1,2).
edge(2,3).
reach(x,y) :- edge(x,y).
reach(x,y) :- edge(x,z), reach(z,y).
out(y) :- reach(1,y).
`

type compilerCase struct {
	name, source string
	facts        map[string]string
}

func preservationCases() []compilerCase {
	return append([]compilerCase{
		{"direct_output", reachProgram + ".output reach\n", nil},
		{"untouched_consumer", reachProgram + ".decl allreach(a:number,b:number)\n.output allreach\nallreach(x,y) :- reach(x,y).\n", nil},
		{"untouched_negation", reachProgram + ".decl node(a:number)\nnode(1).\nnode(2).\nnode(3).\n.decl missing(a:number,b:number)\n.output missing\nmissing(x,y) :- node(x), node(y), !reach(x,y).\n", nil},
		{"adorned_name_collision", reachProgram + ".decl reach_bf(a:number,b:number)\n.output reach_bf\nreach_bf(9,9).\n", nil},
		{"magic_name_collision", reachProgram + ".decl magic_reach_bf(a:number,b:number)\n.output magic_reach_bf\nmagic_reach_bf(9,9).\n", nil},
		{"checkpoint_name_collision", reachProgram + ".decl sup_reach_bf_r0_0(a:number,b:number)\n.output sup_reach_bf_r0_0\nsup_reach_bf_r0_0(9,9).\n", nil},
		{"generated_roles_collide", reachProgram + ".decl magic_reach(a:number,b:number)\n.decl second(a:number)\n.output second\nmagic_reach(1,7).\nsecond(y) :- magic_reach(1,y).\n", nil},
		{"multiple_queries", reachProgram + ".decl other(a:number)\n.output other\nother(y) :- reach(2,y).\n", nil},
		{"input_with_rules", reachProgram + ".input reach\n", map[string]string{"reach": "8\t9\n1\t8\n"}},
		{"input_only_query", ".decl edge(a:number,b:number)\n.input edge\n.decl out(a:number)\n.output out\nout(y) :- edge(1,y).\n", map[string]string{"edge": "1\t2\n2\t3\n"}},
		{"loaded_recursive_seed", ".decl edge(a:number,b:number)\nedge(2,3).\n.decl reach(a:number,b:number)\n.input reach\n.output reach\nreach(x,y) :- reach(x,z),edge(z,y).\n.decl out(y:number)\n.output out\nout(y) :- reach(1,y).\n", map[string]string{"reach": "1\t2\n"}},
		{"arithmetic_head", ".decl edge(x:number,y:number)\nedge(1,7).\nedge(2,9).\n.decl p(x:number,y:number)\np(x+1,y) :- edge(x,y).\n.decl out(y:number)\n.output out\nout(y) :- p(2,y).\n", nil},
		{"recursive_arithmetic_head", ".decl p(x:number)\np(0).\np(x+1) :- p(x), x<5.\n.decl out()\n.output out\nout() :- p(3).\n", nil},
	}, arithmeticCases()...)
}

func checked(t *testing.T, source string) (*ast.Program, *sema.SymbolTable, *sema.StratumResult) {
	t.Helper()
	p, errs := parser.Parse([]byte(source))
	if len(errs) > 0 {
		t.Fatalf("parse: %v", errs)
	}
	return checkAST(t, p)
}

func checkAST(t *testing.T, p *ast.Program) (*ast.Program, *sema.SymbolTable, *sema.StratumResult) {
	t.Helper()
	diags := append(sema.CheckDeclType(p), sema.CheckAllowedness(p)...)
	sd, strata := sema.CheckStratification(p)
	diags = append(diags, sd...)
	if len(diags) > 0 {
		t.Fatalf("semantic checks: %v\n%s", diags, parser.Print(p))
	}
	st, _ := sema.BuildSymbolTable(p)
	return p, st, strata
}

func factsDir(t *testing.T, facts map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, text := range facts {
		if err := os.WriteFile(filepath.Join(dir, name+".facts"), []byte(text), 0600); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func answers(t *testing.T, p *ast.Program, dir string, semi bool) map[string][]string {
	t.Helper()
	_, st, strata := checkAST(t, p)
	e := eval.NewEvaluator(st.Relations)
	inputs := map[string]bool{}
	outputs := map[string]bool{}
	for _, d := range p.Decls {
		if d.Kind == ast.DeclInput {
			inputs[d.Name] = true
		}
		if d.Kind == ast.DeclOutput {
			outputs[d.Name] = true
		}
	}
	if err := e.LoadFacts(dir, st.Relations, inputs); err != nil {
		t.Fatal(err)
	}
	if semi {
		e.RunSemiNaive(p, strata)
	} else {
		e.RunNaive(p, strata.Stratum)
	}
	out := t.TempDir()
	if err := e.WriteOutput(out, st.Relations, outputs); err != nil {
		t.Fatal(err)
	}
	return readAnswers(t, out, outputs)
}

func readAnswers(t *testing.T, dir string, outputs map[string]bool) map[string][]string {
	t.Helper()
	result := map[string][]string{}
	for name := range outputs {
		b, err := os.ReadFile(filepath.Join(dir, name+".csv"))
		if err != nil {
			t.Fatal(err)
		}
		rows := []string{}
		if len(b) > 0 {
			rows = strings.Split(strings.TrimSuffix(string(b), "\n"), "\n")
		}
		sort.Strings(rows)
		result[name] = rows
	}
	return result
}

func transformed(t *testing.T, p *ast.Program, tr transform.Transformer) *ast.Program {
	t.Helper()
	_, _, s := checkAST(t, p)
	q, err := tr.Transform(p, s)
	if err != nil {
		t.Fatal(err)
	}
	checkAST(t, q)
	// This also checks the actual emit/reparse route used by the external engine.
	q, _, _ = checked(t, parser.Print(q))
	return q
}

func TestTransformPreservesEveryOutput(t *testing.T) {
	for _, tc := range preservationCases() {
		t.Run(tc.name, func(t *testing.T) {
			p, _, _ := checked(t, tc.source)
			dir := factsDir(t, tc.facts)
			want := answers(t, p, dir, false)
			if got := answers(t, p, dir, true); !reflect.DeepEqual(got, want) {
				t.Fatalf("original evaluators disagree: naive=%v semi=%v", want, got)
			}
			for name, tr := range map[string]transform.Transformer{"magicset": magicset.Transformer{}, "guarded": guard.Transformer{}, "guarded-compact": compact.Transformer{}} {
				t.Run(name, func(t *testing.T) {
					q := transformed(t, p, tr)
					for _, semi := range []bool{false, true} {
						got := answers(t, q, dir, semi)
						if !reflect.DeepEqual(got, want) {
							t.Errorf("semi=%t: got %v; want %v", semi, got, want)
						}
					}
				})
			}
		})
	}
}

func compileC(t *testing.T, p *ast.Program, flags ...string) string {
	t.Helper()
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("C compiler unavailable")
	}
	_, st, s := checkAST(t, p)
	dir := t.TempDir()
	src := filepath.Join(dir, "program.c")
	bin := filepath.Join(dir, "program")
	if err := os.WriteFile(src, []byte(codegen.Generate(p, st.Relations, s)), 0600); err != nil {
		t.Fatal(err)
	}
	args := append(append([]string{}, flags...), "-o", bin, src)
	if b, err := exec.Command(cc, args...).CombinedOutput(); err != nil {
		t.Fatalf("compile: %v\n%s", err, b)
	}
	return bin
}

func compiledAnswers(t *testing.T, p *ast.Program, dir string) map[string][]string {
	t.Helper()
	bin := compileC(t, p, "-std=c11", "-O2")
	out := t.TempDir()
	if b, err := exec.Command(bin, dir, out).CombinedOutput(); err != nil {
		t.Fatalf("C execution: %v\n%s", err, b)
	}
	outputs := map[string]bool{}
	for _, d := range p.Decls {
		if d.Kind == ast.DeclOutput {
			outputs[d.Name] = true
		}
	}
	return readAnswers(t, out, outputs)
}

func TestOptimizedCMatchesOriginal(t *testing.T) {
	for _, tc := range preservationCases() {
		t.Run(tc.name, func(t *testing.T) {
			p, _, _ := checked(t, tc.source)
			dir := factsDir(t, tc.facts)
			want := answers(t, p, dir, false)
			q := transformed(t, p, guard.Transformer{})
			got := compiledAnswers(t, q, dir)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("got %v; want %v", got, want)
			}
		})
	}
}

func TestBackendDataAndArithmeticAgreement(t *testing.T) {
	wideParams, wideVars, wideRow := []string{}, []string{}, []string{}
	for i := 0; i < 40; i++ {
		wideParams = append(wideParams, fmt.Sprintf("c%d:number", i))
		wideVars = append(wideVars, fmt.Sprintf("x%d", i))
		wideRow = append(wideRow, fmt.Sprint(i))
	}
	cases := []compilerCase{
		{"long_symbol", ".decl p(x:symbol)\n.input p\n.output p\n", map[string]string{"p": strings.Repeat("abcdefghij", 1000) + "\n"}},
		{"empty_symbol", ".decl p(x:symbol)\n.input p\n.output p\n", map[string]string{"p": "\n"}},
		{"wide_relation", ".decl p(" + strings.Join(wideParams, ",") + ")\n.input p\n.output p\n", map[string]string{"p": strings.Join(wideRow, "\t") + "\n"}},
		{"zero_arity_true", ".decl p()\n.input p\n.output p\n", map[string]string{"p": "\n"}},
		{"zero_arity_false", ".decl p()\n.input p\n.output p\n", map[string]string{"p": ""}},
		{"arithmetic_fact", ".decl p(x:number)\n.output p\np(1+2).\np(-3).\n", nil},
		{"arithmetic_atom_order", ".decl p(x:number,y:number)\np(3,2).\n.decl out(x:number)\n.output out\nout(x) :- p(x+1,x).\n", nil},
		{"arithmetic_scheduling", ".decl p(x:number)\np(3).\n.decl q(x:number)\nq(2).\n.decl out(x:number)\n.output out\nout(x) :- p(x+1), q(x).\n", nil},
		{"int64_boundaries", ".decl p(x:number)\n.input p\n.decl out(x:number)\n.output out\nout(x+1) :- p(x).\nout(x * 2) :- p(x).\nout(-x) :- p(x).\nout(x / (-1)) :- p(x).\nout(x % (-1)) :- p(x).\np(-9223372036854775808).\n", map[string]string{"p": "9223372036854775807\n-9223372036854775808\n"}},
		{"unicode_symbols", ".decl p(x:symbol)\n.output p\np(\"नमस्ते\").\np(\"😀\").\n", nil},
		{"zero_divisor", ".decl p(x:number)\np(0).\np(2).\n.decl out(x:number)\n.output out\nout(y) :- p(x), y=10/x.\n", nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p, _, _ := checked(t, tc.source)
			dir := factsDir(t, tc.facts)
			want := answers(t, p, dir, false)
			got := compiledAnswers(t, p, dir)
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("Go=%v C=%v", want, got)
			}
			if tc.name == "zero_arity_true" && !reflect.DeepEqual(want, map[string][]string{"p": {"()"}}) {
				t.Fatal("true zero-arity output must match Souffle's encoding")
			}
			if tc.name == "empty_symbol" && !reflect.DeepEqual(want, map[string][]string{"p": {""}}) {
				t.Fatal("empty symbols must be preserved")
			}
			if tc.name == "arithmetic_atom_order" || tc.name == "arithmetic_scheduling" {
				if !reflect.DeepEqual(want, map[string][]string{"out": {"2"}}) {
					t.Fatal("ground arithmetic after binding its input")
				}
			}
		})
	}
}

func TestInvalidFactsRejectedByBothBackends(t *testing.T) {
	p, st, _ := checked(t, ".decl p(x:number,y:number)\n.input p\n.output p\n")
	bin := compileC(t, p, "-std=c11", "-O2")
	for name, facts := range map[string]string{"missing_column": "1\n", "extra_column": "1\t2\t3\n", "overflow": "9223372036854775808\t1\n", "junk": "1abc\t2\n", "NUL": "1\t2\x00\n"} {
		t.Run(name, func(t *testing.T) {
			dir := factsDir(t, map[string]string{"p": facts})
			e := eval.NewEvaluator(st.Relations)
			if err := e.LoadFacts(dir, st.Relations, map[string]bool{"p": true}); err == nil {
				t.Fatal("Go accepted malformed facts")
			}
			if b, err := exec.Command(bin, dir, t.TempDir()).CombinedOutput(); err == nil {
				t.Fatalf("C accepted malformed facts: %s", b)
			}
		})
	}
}

func TestGeneratedCSanitizers(t *testing.T) {
	source := reachProgram + ".decl n(x:number)\nn(9223372036854775807).\nn(-9223372036854775808).\nn(0).\n.decl arithmetic(x:number)\n.output arithmetic\narithmetic(x+1) :- n(x).\narithmetic(x*2) :- n(x).\narithmetic(-x) :- n(x).\narithmetic(x/(-1)) :- n(x).\narithmetic(x%(-1)) :- n(x).\narithmetic(10/x) :- n(x).\n.decl label(x:symbol)\n.output label\nlabel(\"नमस्ते\").\n"
	p, _, _ := checked(t, source)
	p = transformed(t, p, guard.Transformer{})
	bin := compileC(t, p, "-std=c11", "-O1", "-g", "-fsanitize=address,undefined", "-fno-omit-frame-pointer")
	out := t.TempDir()
	cmd := exec.Command(bin, t.TempDir(), out)
	cmd.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=0:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1")
	if b, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("sanitizers: %v\n%s", err, b)
	}
}
