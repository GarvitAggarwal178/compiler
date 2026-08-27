package guard

import (
	"os"
	"sort"
	"strings"
	"testing"

	"dlc/src/ast"
	"dlc/src/eval"
	"dlc/src/parser"
	"dlc/src/sema"
)

const mixedFallbackSrc = `.decl base(x:number, y:number)
.input base
.decl e(x:number, y:number)
.input e
.decl blocked(x:number)
.input blocked
.decl q(x:number, y:number)
q(x,y) :- base(x,y).
q(x,y) :- q(x,z), base(z,y).
.decl s(x:number)
s(x) :- q(x,y), blocked(y).
.decl p(x:number, y:number)
p(x,y) :- e(x,y).
p(x,y) :- p(x,z), !s(z), q(z,y).
.decl parent(child:number, par:number)
.input parent
.decl person(x:number)
.input person
.decl ancestor(x:number, y:number)
ancestor(x,y) :- parent(x,y).
ancestor(x,y) :- parent(x,z), ancestor(z,y).
.decl nonancestor(x:number, y:number)
nonancestor(x,y) :- person(x), person(y), !ancestor(x,y).
.decl combo(x:number, y:number)
combo(x,y) :- p(x,y), nonancestor(x,y).
.decl ans(y:number)
.output ans
ans(y) :- combo(1,y).
`

// TestDecideProducesGenuineMixedProgram confirms the mixed-fallback shape
// actually exercises a MIXED program (some predicates transformed, some
// declined) rather than the all-or-nothing pattern every other corpus
// program produces (M2-M3-BUILD.md §8's fallback evaluation is only a
// meaningful test with a real mix to evaluate).
func TestDecideProducesGenuineMixedProgram(t *testing.T) {
	prog, errs := parser.Parse([]byte(mixedFallbackSrc))
	if len(errs) != 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	result, err := Decide(prog)
	if err != nil {
		t.Fatalf("Decide error: %v", err)
	}
	for _, pred := range []string{"p", "q", "s"} {
		if !result.DeclinedRelations[pred] {
			t.Errorf("expected %q declined, got DeclinedRelations=%v", pred, result.DeclinedRelations)
		}
	}
	for _, pred := range []string{"ancestor", "nonancestor", "combo", "ans"} {
		if result.DeclinedRelations[pred] {
			t.Errorf("expected %q NOT declined (should stay transformed), got DeclinedRelations=%v", pred, result.DeclinedRelations)
		}
	}
}

// writeFactsDir writes a small, hand-picked fixture set covering both
// groups' EDB relations -- sized for a fast unit test, not the full
// corpus-scale fixture.
func writeMixedFallbackFacts(t *testing.T, dir string) {
	t.Helper()
	files := map[string]string{
		"base":    "1\t2\n2\t3\n3\t4\n",
		"e":       "1\t2\n2\t3\n",
		"blocked": "4\n",
		"parent":  "2\t1\n3\t2\n4\t3\n",
		"person":  "1\n2\n3\n4\n",
	}
	for name, content := range files {
		if err := os.WriteFile(dir+"/"+name+".facts", []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// runAndReadAns evaluates prog with dlc's own naive evaluator against
// factsDir and returns the sorted lines of its "ans" output relation.
func runAndReadAns(t *testing.T, prog *ast.Program, factsDir string) []string {
	t.Helper()
	schemas, diags := sema.BuildSymbolTable(prog)
	if len(diags) != 0 {
		t.Fatalf("symtab diagnostics: %v", diags)
	}
	_, strat := sema.CheckStratification(prog)
	if strat == nil {
		t.Fatalf("program does not stratify")
	}
	inputNames := map[string]bool{}
	outputNames := map[string]bool{"ans": true}
	for _, d := range prog.Decls {
		if d.Kind == ast.DeclInput {
			inputNames[d.Name] = true
		}
	}
	ev := eval.NewEvaluator(schemas.Relations)
	if err := ev.LoadFacts(factsDir, schemas.Relations, inputNames); err != nil {
		t.Fatalf("LoadFacts: %v", err)
	}
	ev.RunNaive(prog, strat.Stratum)

	outDir := t.TempDir()
	if err := ev.WriteOutput(outDir, schemas.Relations, outputNames); err != nil {
		t.Fatalf("WriteOutput: %v", err)
	}
	data, err := os.ReadFile(outDir + "/ans.csv")
	if err != nil {
		return nil
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return nil
	}
	sort.Strings(lines)
	return lines
}

// TestFallbackEvaluationMatchesDlcOwnEvaluator is M2-M3-BUILD.md §8's own
// required verification: "construct a mixed program, evaluate, check
// answers... if the existing SCC-ordered evaluator handles it, say so."
// dlc's own eval.RunNaive is run on BOTH the untransformed original and
// guard.Decide's genuinely mixed final program (re-stratified, per
// Transformer's documented contract), and their "ans" answer relations
// are compared -- confirming the EXISTING evaluator, unmodified, handles
// a program with some relations transformed and others reading full,
// untransformed extent, with no new evaluation machinery required.
func TestFallbackEvaluationMatchesDlcOwnEvaluator(t *testing.T) {
	prog, errs := parser.Parse([]byte(mixedFallbackSrc))
	if len(errs) != 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	result, err := Decide(prog)
	if err != nil {
		t.Fatalf("Decide error: %v", err)
	}

	factsDir := t.TempDir()
	writeMixedFallbackFacts(t, factsDir)

	originalAns := runAndReadAns(t, prog, factsDir)
	mixedAns := runAndReadAns(t, result.FinalProgram, factsDir)

	if len(originalAns) == 0 {
		t.Fatal("test fixture produced zero ans tuples on the untransformed original -- fixture is not exercising anything, strengthen it")
	}
	if strings.Join(originalAns, ",") != strings.Join(mixedAns, ",") {
		t.Fatalf("dlc's own evaluator disagrees between original and mixed programs:\noriginal: %v\nmixed:    %v", originalAns, mixedAns)
	}
}
