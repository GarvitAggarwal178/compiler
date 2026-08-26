package sema

import (
	"testing"

	"dlc/src/parser"
)

func mustStratify(t *testing.T, src string) ([]Diagnostic, *StratumResult) {
	t.Helper()
	prog, errs := parser.Parse([]byte(src))
	if len(errs) != 0 {
		t.Fatalf("parser.Parse(%q) returned unexpected parse errors: %v", src, errs)
	}
	return CheckStratification(prog)
}

// The 3 tests/rejection/stratification.py cases, restated directly
// (that file is Python, not importable here) -- all 3 must reject.
func TestStratificationSelfNegativeCycle(t *testing.T) {
	diags, _ := mustStratify(t, ".decl a(x:number)\n.input a\na(x) :- a(x), !a(x).\n")
	if len(diags) != 1 || diags[0].Category != Unstratifiable {
		t.Fatalf("expected exactly one Unstratifiable diagnostic, got %v", diags)
	}
}

func TestStratificationMutualNegativeCycle(t *testing.T) {
	src := ".decl base(x:number)\n.input base\n.decl p(x:number)\n.decl q(x:number)\n.decl out(x:number)\n.output out\n" +
		"p(x) :- base(x), !q(x).\nq(x) :- base(x), !p(x).\nout(x) :- p(x).\n"
	diags, _ := mustStratify(t, src)
	if len(diags) != 1 || diags[0].Category != Unstratifiable {
		t.Fatalf("expected exactly one Unstratifiable diagnostic, got %v", diags)
	}
}

func TestStratificationCycleThroughPositiveChain(t *testing.T) {
	src := ".decl base(x:number)\n.input base\n.decl p(x:number)\n.decl q(x:number)\n.decl r(x:number)\n.decl out(x:number)\n.output out\n" +
		"p(x) :- base(x), !q(x).\nq(x) :- r(x).\nr(x) :- p(x).\nout(x) :- p(x).\n"
	diags, _ := mustStratify(t, src)
	if len(diags) != 1 || diags[0].Category != Unstratifiable {
		t.Fatalf("expected exactly one Unstratifiable diagnostic, got %v", diags)
	}
}

func TestStratifiableSimpleNegation(t *testing.T) {
	// ancestor_nonancestor.dl's shape: self-recursive positive IDB,
	// then one relation negatively depending on it once -- two strata.
	src := ".decl parent(a:number,b:number)\n.input parent\n.decl person(a:number)\n.input person\n" +
		".decl ancestor(a:number,b:number)\n.decl nonancestor(a:number,b:number)\n.output nonancestor\n" +
		"ancestor(x,y) :- parent(x,y).\nancestor(x,y) :- parent(x,z), ancestor(z,y).\n" +
		"nonancestor(x,y) :- person(x), person(y), !ancestor(x,y).\n"
	diags, result := mustStratify(t, src)
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	if result.Stratum["ancestor"] != 0 {
		t.Fatalf("expected ancestor at stratum 0, got %d", result.Stratum["ancestor"])
	}
	if result.Stratum["nonancestor"] <= result.Stratum["ancestor"] {
		t.Fatalf("expected nonancestor strictly after ancestor (negative dependency), got ancestor=%d nonancestor=%d",
			result.Stratum["ancestor"], result.Stratum["nonancestor"])
	}
}

func TestStratifiableCulpritCycleShape(t *testing.T) {
	// culprit_cycle.dl's SOURCE shape is stratifiable -- it is only a
	// bad magic-set transform of it that becomes unstratifiable (T5/T7,
	// docs/reports/night02-T5-guarded.md, night02-T7-p5-precheck.md).
	// This pins that the source stratifier does not confuse the two.
	src := ".decl base(x:number,y:number)\n.input base\n.decl e(x:number,y:number)\n.input e\n" +
		".decl blocked(x:number)\n.input blocked\n" +
		".decl q(x:number,y:number)\nq(x,y) :- base(x,y).\nq(x,y) :- q(x,z), base(z,y).\n" +
		".decl s(x:number)\ns(x) :- q(x,y), blocked(y).\n" +
		".decl p(x:number,y:number)\np(x,y) :- e(x,y).\np(x,y) :- p(x,z), !s(z), q(z,y).\n" +
		".decl out(y:number)\n.output out\nout(y) :- p(1,y).\n"
	diags, result := mustStratify(t, src)
	if len(diags) != 0 {
		t.Fatalf("expected the source culprit_cycle shape to be stratifiable, got %v", diags)
	}
	if result.Stratum["p"] <= result.Stratum["s"] {
		t.Fatalf("expected p strictly after s (negative dependency), got s=%d p=%d",
			result.Stratum["s"], result.Stratum["p"])
	}
	if result.Stratum["s"] < result.Stratum["q"] {
		t.Fatalf("expected s at or after q (positive dependency), got q=%d s=%d",
			result.Stratum["q"], result.Stratum["s"])
	}
}

func TestStratificationDeterministicAcrossRuns(t *testing.T) {
	// CLAUDE.md's determinism rule: same input must produce byte-identical
	// output every time -- map iteration order must never leak in.
	src := ".decl a(x:number)\n.input a\na(x) :- a(x), !a(x).\n" // self-negative-cycle, deliberately unstratifiable
	var first string
	for i := 0; i < 20; i++ {
		diags, _ := mustStratify(t, src)
		if len(diags) != 1 {
			t.Fatalf("expected exactly one diagnostic, got %v", diags)
		}
		if i == 0 {
			first = diags[0].Message
		} else if diags[0].Message != first {
			t.Fatalf("non-deterministic diagnostic message across runs: %q vs %q", first, diags[0].Message)
		}
	}
}

// The exact shape that exposed the stratum-vs-SCC bug in eval (found via
// example/josephus/josephus.dl, M1 §3.9): a self-recursive relation
// (Relation) and a second, independent relation (Josephus) that reads it
// positively -- no negation anywhere, so both end up at stratum 0, but
// Josephus's SCC must still be evaluated strictly after Relation's SCC
// has fully converged. SCCOrder is what carries that ordering; stratum
// number alone does not.
func TestSCCOrderRespectsPositiveDependencyEvenWithinOneStratum(t *testing.T) {
	src := ".decl Relation(a:symbol,b:symbol)\n" +
		"Relation(a,c) :- Relation(a,b), Relation(b,c).\n" +
		".decl Josephus(a:symbol)\n" +
		"Josephus(a) :- Relation(a,b), Relation(b,a).\n"
	_, result := mustStratify(t, src)
	if result.Stratum["Relation"] != result.Stratum["Josephus"] {
		t.Fatalf("expected both at the same stratum number (no negation anywhere), got Relation=%d Josephus=%d",
			result.Stratum["Relation"], result.Stratum["Josephus"])
	}
	relationPos, josephusPos := -1, -1
	for i, scc := range result.SCCOrder {
		for _, name := range scc {
			if name == "Relation" {
				relationPos = i
			}
			if name == "Josephus" {
				josephusPos = i
			}
		}
	}
	if relationPos == -1 || josephusPos == -1 {
		t.Fatalf("expected both relations to appear in SCCOrder, got %v", result.SCCOrder)
	}
	if !(relationPos < josephusPos) {
		t.Fatalf("expected Relation's SCC to precede Josephus's SCC in SCCOrder (Josephus depends on Relation), got order %v", result.SCCOrder)
	}
}
