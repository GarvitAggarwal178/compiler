package eval

import (
	"testing"

	"dlc/src/parser"
	"dlc/src/sema"
)

func setupAndRunSemiNaive(t *testing.T, src string) *Evaluator {
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
	ev.RunSemiNaive(prog, stratResult)
	return ev
}

func TestSemiNaiveTransitiveClosureMatchesNaive(t *testing.T) {
	src := ".decl edge(a:number,b:number)\n.decl reach(a:number,b:number)\n" +
		"edge(1,2).\nedge(2,3).\nedge(3,4).\n" +
		"reach(x,y) :- edge(x,y).\nreach(x,y) :- reach(x,z), edge(z,y).\n"
	naive := setupAndRun(t, src)
	semi := setupAndRunSemiNaive(t, src)
	if naive.Relations["reach"].Len() != semi.Relations["reach"].Len() {
		t.Fatalf("naive found %d reach tuples, semi-naive found %d", naive.Relations["reach"].Len(), semi.Relations["reach"].Len())
	}
	if semi.Relations["reach"].Len() != 6 {
		t.Fatalf("expected 6 (full transitive closure of a 4-node chain), got %d", semi.Relations["reach"].Len())
	}
}

func TestSemiNaiveStratifiedNegationMatchesNaive(t *testing.T) {
	src := ".decl edge(a:number,b:number)\n.decl node(a:number)\n.decl reach(a:number,b:number)\n.decl unreach(a:number,b:number)\n" +
		"node(1).\nnode(2).\nnode(3).\n" +
		"edge(1,2).\n" +
		"reach(x,y) :- edge(x,y).\n" +
		"unreach(x,y) :- node(x), node(y), !reach(x,y).\n"
	semi := setupAndRunSemiNaive(t, src)
	if semi.Relations["unreach"].Len() != 8 {
		t.Fatalf("expected 8 unreach tuples, got %d", semi.Relations["unreach"].Len())
	}
}

// A self-join in one recursive rule: two occurrences of the SAME
// relation in one body, both same-stratum. Pins the position-keyed (not
// name-keyed) override design in evalBody/seminaive.go -- a name-keyed
// override would force BOTH occurrences to delta simultaneously every
// variant, silently dropping old-new/new-old combinations, and would
// under-count the transitive closure.
func TestSemiNaiveSelfJoinMatchesNaive(t *testing.T) {
	src := ".decl edge(a:number,b:number)\n.decl p(a:number,b:number)\n" +
		"edge(1,2).\nedge(2,3).\nedge(3,4).\nedge(4,5).\n" +
		"p(x,y) :- edge(x,y).\np(x,y) :- p(x,z), p(z,y).\n" // self-join recursion, not the edge-driven shape
	naive := setupAndRun(t, src)
	semi := setupAndRunSemiNaive(t, src)
	if naive.Relations["p"].Len() != semi.Relations["p"].Len() {
		t.Fatalf("naive found %d p tuples, semi-naive found %d -- self-join Δ-rewrite disagreement", naive.Relations["p"].Len(), semi.Relations["p"].Len())
	}
	// full transitive closure of a 5-node chain: 10 pairs
	if semi.Relations["p"].Len() != 10 {
		t.Fatalf("expected 10 (full transitive closure of a 5-node chain), got %d", semi.Relations["p"].Len())
	}
}

func TestSemiNaiveInstrumentationHasRealSeedVsIterationSplit(t *testing.T) {
	src := ".decl edge(a:number,b:number)\n.decl reach(a:number,b:number)\n" +
		"edge(1,2).\nedge(2,3).\nedge(3,4).\n" +
		"reach(x,y) :- edge(x,y).\nreach(x,y) :- reach(x,z), edge(z,y).\n"
	ev := setupAndRunSemiNaive(t, src)
	stats := ev.Relations["reach"].Stats
	if stats.SeedInserts != 3 {
		t.Fatalf("expected 3 seed inserts (one per edge fact), got %d", stats.SeedInserts)
	}
	if len(stats.IterationInserts) == 0 {
		t.Fatalf("expected at least one recursive iteration to have fired, got none")
	}
	if stats.Total() != 6 {
		t.Fatalf("expected Total()=6 (matches naive's full derived-tuple count), got %d", stats.Total())
	}
}

func TestSemiNaiveNoRecursionIsJustTheSeedRound(t *testing.T) {
	src := ".decl q(a:number)\n.decl p(a:number)\nq(1).\nq(2).\np(x) :- q(x), x > 0.\n"
	ev := setupAndRunSemiNaive(t, src)
	if ev.Relations["p"].Len() != 2 {
		t.Fatalf("expected 2 tuples, got %d", ev.Relations["p"].Len())
	}
	if ev.Relations["p"].Stats.SeedInserts != 2 {
		t.Fatalf("expected 2 seed inserts and no iteration rounds for a non-recursive stratum, got seed=%d iterations=%v",
			ev.Relations["p"].Stats.SeedInserts, ev.Relations["p"].Stats.IterationInserts)
	}
}

// The exact shape from example/josephus/josephus.dl that found the
// stratum-vs-SCC evaluation-order bug: a self-recursive relation and a
// second, independent relation reading it positively, both at stratum 0
// (no negation anywhere in the program), where the second relation must
// still wait for the first's SCC to fully converge. Matches real
// Soufflé's own answer, verified separately
// (harness/m1_3_9_gate1_seminaive_agreement.py).
func TestSemiNaiveWaitsForDependencySCCEvenAtSameStratum(t *testing.T) {
	src := ".decl Relation(a:symbol,b:symbol,step:number,turn:symbol)\n" +
		"Relation(a,c,step+1,c) :- Relation(a,b,step,a), Relation(b,c,step,a), a != c.\n" +
		"Relation(d,e,step+1,c) :- Relation(a,b,step,a), Relation(b,c,step,a), Relation(d,e,step,a), a != c, d != a, d != b.\n" +
		".decl Josephus(a:symbol)\n" +
		"Josephus(a) :- Relation(a,b,step,a), Relation(b,a,step,a).\n" +
		"Relation(\"a\",\"b\",0,\"a\").\nRelation(\"b\",\"c\",0,\"a\").\nRelation(\"c\",\"d\",0,\"a\").\n" +
		"Relation(\"d\",\"e\",0,\"a\").\nRelation(\"e\",\"f\",0,\"a\").\nRelation(\"f\",\"a\",0,\"a\").\n"
	naive := setupAndRun(t, src)
	semi := setupAndRunSemiNaive(t, src)
	if naive.Relations["Josephus"].Len() != semi.Relations["Josephus"].Len() {
		t.Fatalf("naive found %d Josephus tuples, semi-naive found %d -- SCC-order regression",
			naive.Relations["Josephus"].Len(), semi.Relations["Josephus"].Len())
	}
	if semi.Relations["Josephus"].Len() == 0 {
		t.Fatalf("expected at least one Josephus tuple (real Soufflé finds \"e\"), got 0 -- " +
			"almost certainly evaluated against a not-yet-converged Relation")
	}
}

func TestSemiNaiveDoesFewerDerivationAttemptsThanNaive(t *testing.T) {
	// Distinct-tuple counts (ir.RelationStats) are expected to match
	// exactly between naive and semi-naive (both compute the same
	// minimal model, DESIGN.md) -- DerivationAttempts is the number that
	// is actually supposed to differ, showing semi-naive's saved
	// redundant work, on a chain long enough for naive's repeated full
	// rescans to add up.
	src := ".decl edge(a:number,b:number)\n.decl reach(a:number,b:number)\n" +
		"edge(1,2).\nedge(2,3).\nedge(3,4).\nedge(4,5).\nedge(5,6).\nedge(6,7).\n" +
		"reach(x,y) :- edge(x,y).\nreach(x,y) :- reach(x,z), edge(z,y).\n"
	naive := setupAndRun(t, src)
	semi := setupAndRunSemiNaive(t, src)
	if naive.Relations["reach"].Len() != semi.Relations["reach"].Len() {
		t.Fatalf("distinct-tuple counts disagree: naive=%d semi=%d", naive.Relations["reach"].Len(), semi.Relations["reach"].Len())
	}
	if !(semi.DerivationAttempts < naive.DerivationAttempts) {
		t.Fatalf("expected semi-naive to attempt strictly fewer derivations than naive on a 6-edge chain, got naive=%d semi=%d",
			naive.DerivationAttempts, semi.DerivationAttempts)
	}
}
