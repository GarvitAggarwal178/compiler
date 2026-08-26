package eval

import (
	"os"
	"sort"
	"testing"

	"dlc/src/parser"
	"dlc/src/sema"
)

func setupAndRun(t *testing.T, src string) *Evaluator {
	t.Helper()
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
	ev := NewEvaluator(schemas.Relations)
	ev.RunNaive(prog, stratResult.Stratum)
	return ev
}

func relRows(t *testing.T, ev *Evaluator, name string) [][]int64 {
	t.Helper()
	rel := ev.Relations[name]
	var out [][]int64
	for _, tup := range rel.All() {
		row := make([]int64, len(tup))
		for i, v := range tup {
			row[i] = v.Num
		}
		out = append(out, row)
	}
	sort.Slice(out, func(i, j int) bool {
		for k := range out[i] {
			if out[i][k] != out[j][k] {
				return out[i][k] < out[j][k]
			}
		}
		return false
	})
	return out
}

func TestFactsAndSimpleRule(t *testing.T) {
	src := ".decl edge(a:number,b:number)\n.decl reach(a:number,b:number)\n" +
		"edge(1,2).\nedge(2,3).\nreach(x,y) :- edge(x,y).\n"
	ev := setupAndRun(t, src)
	got := relRows(t, ev, "reach")
	want := [][]int64{{1, 2}, {2, 3}}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i][0] != want[i][0] || got[i][1] != want[i][1] {
			t.Fatalf("got %v, want %v", got, want)
		}
	}
}

func TestTransitiveClosureRecursion(t *testing.T) {
	src := ".decl edge(a:number,b:number)\n.decl reach(a:number,b:number)\n" +
		"edge(1,2).\nedge(2,3).\nedge(3,4).\n" +
		"reach(x,y) :- edge(x,y).\nreach(x,y) :- reach(x,z), edge(z,y).\n"
	ev := setupAndRun(t, src)
	got := relRows(t, ev, "reach")
	// full transitive closure of a 4-node chain: (1,2)(1,3)(1,4)(2,3)(2,4)(3,4)
	if len(got) != 6 {
		t.Fatalf("expected 6 reach tuples (full transitive closure), got %d: %v", len(got), got)
	}
}

func TestStratifiedNegation(t *testing.T) {
	src := ".decl edge(a:number,b:number)\n.decl node(a:number)\n.decl reach(a:number,b:number)\n.decl unreach(a:number,b:number)\n" +
		"node(1).\nnode(2).\nnode(3).\n" +
		"edge(1,2).\n" +
		"reach(x,y) :- edge(x,y).\n" +
		"unreach(x,y) :- node(x), node(y), !reach(x,y).\n"
	ev := setupAndRun(t, src)
	got := relRows(t, ev, "unreach")
	// node x node = 9 pairs, minus (1,2) which is reachable = 8
	if len(got) != 8 {
		t.Fatalf("expected 8 unreach tuples, got %d: %v", len(got), got)
	}
	for _, row := range got {
		if row[0] == 1 && row[1] == 2 {
			t.Fatalf("(1,2) should not be in unreach (it IS reachable), got %v", got)
		}
	}
}

func TestEquationBeforeGroundingAtomSafeOrder(t *testing.T) {
	// J1 probe case (b): X = Y + 1 written BEFORE q(Y), which grounds Y.
	// A naive left-to-right walk without reordering would fail here.
	src := ".decl q(a:number)\n.decl p(a:number)\n" +
		"q(5).\np(X) :- X = Y + 1, q(Y).\n"
	ev := setupAndRun(t, src)
	got := relRows(t, ev, "p")
	if len(got) != 1 || got[0][0] != 6 {
		t.Fatalf("expected p(6), got %v", got)
	}
}

func TestArithmeticInBody(t *testing.T) {
	src := ".decl q(a:number)\n.decl p(a:number)\n" +
		"q(1).\nq(2).\nq(3).\np(Y) :- q(X), Y = X * 2 + 1.\n"
	ev := setupAndRun(t, src)
	got := relRows(t, ev, "p")
	want := map[int64]bool{3: true, 5: true, 7: true}
	if len(got) != 3 {
		t.Fatalf("expected 3 tuples, got %v", got)
	}
	for _, row := range got {
		if !want[row[0]] {
			t.Fatalf("unexpected value %d in %v", row[0], got)
		}
	}
}

func TestConstraintFiltersResults(t *testing.T) {
	src := ".decl q(a:number)\n.decl p(a:number)\n" +
		"q(1).\nq(2).\nq(3).\nq(4).\np(X) :- q(X), X > 2.\n"
	ev := setupAndRun(t, src)
	got := relRows(t, ev, "p")
	if len(got) != 2 {
		t.Fatalf("expected 2 tuples (3,4), got %v", got)
	}
}

func TestZeroArityRelationEvaluates(t *testing.T) {
	src := ".decl p()\n.decl q()\np().\nq() :- p().\n"
	ev := setupAndRun(t, src)
	if ev.Relations["q"].Len() != 1 {
		t.Fatalf("expected q() to derive exactly the one zero-arity tuple, got Len()=%d", ev.Relations["q"].Len())
	}
}

func TestSymbolValuesAndStringLiterals(t *testing.T) {
	src := ".decl q(a:symbol)\n.decl p(a:symbol)\n" +
		"q(\"a\").\nq(\"b\").\np(X) :- q(X), X != \"a\".\n"
	ev := setupAndRun(t, src)
	rel := ev.Relations["p"]
	if rel.Len() != 1 {
		t.Fatalf("expected 1 tuple, got %d", rel.Len())
	}
	got := ev.valueString(rel.All()[0][0])
	if got != "b" {
		t.Fatalf("expected \"b\", got %q", got)
	}
}

func TestDedupAcrossMultipleRulesForSameHead(t *testing.T) {
	src := ".decl q(a:number)\n.decl p(a:number)\n" +
		"q(1).\np(X) :- q(X).\np(X) :- q(X), X > 0.\n" // both rules derive p(1); must not double-count
	ev := setupAndRun(t, src)
	rel := ev.Relations["p"]
	if rel.Len() != 1 {
		t.Fatalf("expected exactly 1 tuple after dedup, got %d", rel.Len())
	}
}

func TestInstrumentationExcludesInputLoadsButCountsSourceFacts(t *testing.T) {
	// Soufflé's own convention (harness/tuple_report.py's
	// is_input_relation: identified by a "loadtime" attribute, which
	// only an .input-loaded relation has) only excludes .input-loaded
	// EDB data from the derived-tuple total -- a fact written directly
	// in source (`edge(1,2).`) is a non-recursive rule like any other
	// and IS counted, in both Soufflé and here. LoadFacts (io.go), which
	// backs .input, is the path that must never call
	// RecordSeedInsert/RecordIterationInsert; plain fact clauses go
	// through the ordinary clause-evaluation path and are expected to
	// be counted.
	src := ".decl edge(a:number,b:number)\n.decl reach(a:number,b:number)\n" +
		"edge(1,2).\nedge(2,3).\nreach(x,y) :- edge(x,y).\n"
	ev := setupAndRun(t, src)
	if ev.Relations["edge"].Stats.Total() != 2 {
		t.Fatalf("expected edge (source facts, non-recursive rules) to have Stats.Total()=2, got %d", ev.Relations["edge"].Stats.Total())
	}
	if ev.Relations["reach"].Stats.Total() != 2 {
		t.Fatalf("expected reach (rule-derived) to have Stats.Total()=2, got %d", ev.Relations["reach"].Stats.Total())
	}
}

func TestLoadFactsNeverRecordsStats(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(dir+"/edge.facts", []byte("1\t2\n2\t3\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	src := ".decl edge(a:number,b:number)\n.input edge\n.decl reach(a:number,b:number)\n" +
		"reach(x,y) :- edge(x,y).\n"
	prog, errs := parser.Parse([]byte(src))
	if len(errs) != 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	schemas, _ := sema.BuildSymbolTable(prog)
	ev := NewEvaluator(schemas.Relations)
	if err := ev.LoadFacts(dir, schemas.Relations, map[string]bool{"edge": true}); err != nil {
		t.Fatalf("LoadFacts: %v", err)
	}
	if ev.Relations["edge"].Len() != 2 {
		t.Fatalf("expected 2 loaded tuples, got %d", ev.Relations["edge"].Len())
	}
	if ev.Relations["edge"].Stats.Total() != 0 {
		t.Fatalf("expected .input-loaded edge to have Stats.Total()=0 (never rule-derived), got %d", ev.Relations["edge"].Stats.Total())
	}
}

func TestUnaryMinus(t *testing.T) {
	src := ".decl q(a:number)\n.decl p(a:number)\nq(5).\np(Y) :- q(X), Y = -X.\n"
	ev := setupAndRun(t, src)
	got := relRows(t, ev, "p")
	if len(got) != 1 || got[0][0] != -5 {
		t.Fatalf("expected p(-5), got %v", got)
	}
}
