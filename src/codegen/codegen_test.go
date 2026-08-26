package codegen

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"dlc/src/parser"
	"dlc/src/sema"
)

// compileAndRun generates C for src, compiles it with `cc` (skipping the
// test if no C compiler is available -- this project's environment has
// one, but a real end-to-end test should not hard-fail somewhere that
// doesn't), runs the resulting binary against factsDir, and returns the
// sorted lines of outRel's .csv.
func compileAndRun(t *testing.T, src string, factsDir string, outRel string) []string {
	t.Helper()
	if _, err := exec.LookPath("cc"); err != nil {
		t.Skip("no C compiler available")
	}

	prog, errs := parser.Parse([]byte(src))
	if len(errs) != 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	semaDiags := sema.CheckDeclType(prog)
	semaDiags = append(semaDiags, sema.CheckAllowedness(prog)...)
	stratDiags, stratResult := sema.CheckStratification(prog)
	semaDiags = append(semaDiags, stratDiags...)
	if len(semaDiags) != 0 {
		t.Fatalf("sema diagnostics: %v", semaDiags)
	}
	schemas, _ := sema.BuildSymbolTable(prog)

	cSource := Generate(prog, schemas.Relations, stratResult)

	dir := t.TempDir()
	cPath := filepath.Join(dir, "prog.c")
	if err := os.WriteFile(cPath, []byte(cSource), 0o644); err != nil {
		t.Fatalf("writing C source: %v", err)
	}
	binPath := filepath.Join(dir, "prog")
	cmd := exec.Command("cc", "-O0", "-o", binPath, cPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("cc failed: %v\n%s\n--- generated source ---\n%s", err, out, cSource)
	}

	outDir := filepath.Join(dir, "out")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		t.Fatal(err)
	}
	runCmd := exec.Command(binPath, factsDir, outDir)
	runOut, err := runCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("generated binary failed: %v\n%s", err, runOut)
	}

	data, err := os.ReadFile(filepath.Join(outDir, outRel+".csv"))
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

func writeFacts(t *testing.T, dir, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, name+".facts"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestCodegenFactsAndSimpleRule(t *testing.T) {
	src := ".decl edge(a:number,b:number)\n.decl reach(a:number,b:number)\n.output reach\n" +
		"edge(1,2).\nedge(2,3).\nreach(x,y) :- edge(x,y).\n"
	got := compileAndRun(t, src, t.TempDir(), "reach")
	want := []string{"1\t2", "2\t3"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestCodegenTransitiveClosureRecursion(t *testing.T) {
	src := ".decl edge(a:number,b:number)\n.decl reach(a:number,b:number)\n.output reach\n" +
		"edge(1,2).\nedge(2,3).\nedge(3,4).\n" +
		"reach(x,y) :- edge(x,y).\nreach(x,y) :- reach(x,z), edge(z,y).\n"
	got := compileAndRun(t, src, t.TempDir(), "reach")
	if len(got) != 6 {
		t.Fatalf("expected 6 reach tuples (full transitive closure), got %d: %v", len(got), got)
	}
}

func TestCodegenStratifiedNegation(t *testing.T) {
	src := ".decl edge(a:number,b:number)\n.decl node(a:number)\n.decl reach(a:number,b:number)\n.decl unreach(a:number,b:number)\n.output unreach\n" +
		"node(1).\nnode(2).\nnode(3).\n" +
		"edge(1,2).\n" +
		"reach(x,y) :- edge(x,y).\n" +
		"unreach(x,y) :- node(x), node(y), !reach(x,y).\n"
	got := compileAndRun(t, src, t.TempDir(), "unreach")
	if len(got) != 8 {
		t.Fatalf("expected 8 unreach tuples, got %d: %v", len(got), got)
	}
}

func TestCodegenInputFacts(t *testing.T) {
	dir := t.TempDir()
	writeFacts(t, dir, "edge", "1\t2\n2\t3\n3\t4\n")
	src := ".decl edge(a:number,b:number)\n.input edge\n.decl reach(a:number,b:number)\n.output reach\n" +
		"reach(x,y) :- edge(x,y).\nreach(x,y) :- reach(x,z), edge(z,y).\n"
	got := compileAndRun(t, src, dir, "reach")
	if len(got) != 6 {
		t.Fatalf("expected 6 reach tuples, got %d: %v", len(got), got)
	}
}

func TestCodegenArithmeticAndConstraint(t *testing.T) {
	src := ".decl q(a:number)\n.decl p(a:number)\n.output p\n" +
		"q(1).\nq(2).\nq(3).\nq(4).\np(y) :- q(x), y = x * 2 + 1, x > 1.\n"
	got := compileAndRun(t, src, t.TempDir(), "p")
	want := map[string]bool{"5": true, "7": true, "9": true}
	if len(got) != 3 {
		t.Fatalf("expected 3 tuples, got %v", got)
	}
	for _, g := range got {
		if !want[g] {
			t.Fatalf("unexpected value %q in %v", g, got)
		}
	}
}

func TestCodegenSymbolsAndStringLiterals(t *testing.T) {
	src := ".decl q(a:symbol)\n.decl p(a:symbol)\n.output p\n" +
		"q(\"a\").\nq(\"b\").\np(x) :- q(x), x != \"a\".\n"
	got := compileAndRun(t, src, t.TempDir(), "p")
	if len(got) != 1 || got[0] != "b" {
		t.Fatalf("expected [\"b\"], got %v", got)
	}
}

func TestCodegenSelfJoin(t *testing.T) {
	src := ".decl edge(a:number,b:number)\n.decl p(a:number,b:number)\n.output p\n" +
		"edge(1,2).\nedge(2,3).\nedge(3,4).\nedge(4,5).\n" +
		"p(x,y) :- edge(x,y).\np(x,y) :- p(x,z), p(z,y).\n"
	got := compileAndRun(t, src, t.TempDir(), "p")
	if len(got) != 10 {
		t.Fatalf("expected 10 (full transitive closure of a 5-node chain), got %d: %v", len(got), got)
	}
}

func TestCodegenZeroArity(t *testing.T) {
	// Zero-arity output is a single degenerate case (a "row" with zero
	// columns is indistinguishable, in a naive line-based reading, from
	// an empty file) -- checked directly against the raw file rather
	// than through compileAndRun's general sorted-lines-or-nil helper,
	// which cannot tell "one zero-column row" from "no rows" apart.
	if _, err := exec.LookPath("cc"); err != nil {
		t.Skip("no C compiler available")
	}
	src := ".decl p()\n.decl q()\n.output q\np().\nq() :- p().\n"

	prog, errs := parser.Parse([]byte(src))
	if len(errs) != 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	semaDiags := sema.CheckDeclType(prog)
	semaDiags = append(semaDiags, sema.CheckAllowedness(prog)...)
	stratDiags, stratResult := sema.CheckStratification(prog)
	semaDiags = append(semaDiags, stratDiags...)
	if len(semaDiags) != 0 {
		t.Fatalf("sema diagnostics: %v", semaDiags)
	}
	schemas, _ := sema.BuildSymbolTable(prog)
	cSource := Generate(prog, schemas.Relations, stratResult)

	dir := t.TempDir()
	cPath := filepath.Join(dir, "prog.c")
	os.WriteFile(cPath, []byte(cSource), 0o644)
	binPath := filepath.Join(dir, "prog")
	if out, err := exec.Command("cc", "-O0", "-o", binPath, cPath).CombinedOutput(); err != nil {
		t.Fatalf("cc failed: %v\n%s", err, out)
	}
	outDir := filepath.Join(dir, "out")
	os.MkdirAll(outDir, 0o755)
	if out, err := exec.Command(binPath, dir, outDir).CombinedOutput(); err != nil {
		t.Fatalf("generated binary failed: %v\n%s", err, out)
	}
	data, err := os.ReadFile(filepath.Join(outDir, "q.csv"))
	if err != nil {
		t.Fatalf("reading q.csv: %v", err)
	}
	if string(data) != "\n" {
		t.Fatalf("expected q.csv to contain exactly one blank line (one zero-arity tuple), got %q", string(data))
	}
}
