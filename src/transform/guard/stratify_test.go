package guard

import (
	"strings"
	"testing"

	"dlc/src/parser"
)

const culpritCycleSrc = `.decl base(x:number, y:number)
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
.decl out(y:number)
.output out
out(y) :- p(1,y).
`

func TestHasPositiveCycleMatchesCulpritCycleShape(t *testing.T) {
	prog, errs := parser.Parse([]byte(culpritCycleSrc))
	if len(errs) != 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	cycles := HasPositiveCycle(prog)
	if !cycles["q"] {
		t.Error("expected q to be in a positive cycle (self-recursive)")
	}
	if !cycles["p"] {
		t.Error("expected p to be in a positive cycle (self-recursive)")
	}
	if cycles["s"] {
		t.Error("expected s NOT to be in a positive cycle (non-recursive)")
	}
	if cycles["out"] {
		t.Error("expected out NOT to be in a positive cycle (non-recursive)")
	}
}

func TestCheckCulpritCycleDetectsUnstratifiableTransform(t *testing.T) {
	prog, errs := parser.Parse([]byte(culpritCycleSrc))
	if len(errs) != 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	result, err := CheckCulpritCycle(prog)
	if err != nil {
		t.Fatalf("CheckCulpritCycle error: %v", err)
	}
	if result.Stratifiable {
		t.Fatal("expected culprit_cycle's mechanical transform to be unstratifiable")
	}
	if !strings.Contains(result.Message, "cycle") {
		t.Fatalf("expected the diagnostic message to name a cycle, got: %q", result.Message)
	}
}

func TestCheckCulpritCycleAcceptsAncestorNonancestor(t *testing.T) {
	src := `.decl parent(child:number, par:number)
.input parent
.decl person(x:number)
.input person
.decl ancestor(x:number, y:number)
ancestor(x,y) :- parent(x,y).
ancestor(x,y) :- parent(x,z), ancestor(z,y).
.decl nonancestor(x:number, y:number)
nonancestor(x,y) :- person(x), person(y), !ancestor(x,y).
.decl q_nonancestor(y:number)
.output q_nonancestor
q_nonancestor(y) :- nonancestor(1,y).
`
	prog, errs := parser.Parse([]byte(src))
	if len(errs) != 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	result, err := CheckCulpritCycle(prog)
	if err != nil {
		t.Fatalf("CheckCulpritCycle error: %v", err)
	}
	if !result.Stratifiable {
		t.Fatalf("expected ancestor_nonancestor's mechanical transform to stratify, got: %s", result.Message)
	}
}

func TestCheckCulpritCyclePreconditionSkipsNonRecursiveProgram(t *testing.T) {
	// No relation here is in a positive cycle at all (p, q, ans are all
	// non-recursive) -- the cheap precondition must skip the full
	// adorn-and-check path entirely, not just happen to return
	// Stratifiable=true via the general path.
	src := ".decl p(x:number,y:number)\n.decl ans(y:number)\n.output ans\n" +
		"p(1,2).\nans(y) :- p(1,y).\n"
	prog, errs := parser.Parse([]byte(src))
	if len(errs) != 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	result, err := CheckCulpritCycle(prog)
	if err != nil {
		t.Fatalf("CheckCulpritCycle error: %v", err)
	}
	if !result.PreconditionSkipped {
		t.Fatal("expected the cheap precondition to skip a program with no positive cycle anywhere")
	}
	if !result.Stratifiable {
		t.Fatal("expected a precondition-skipped result to be trivially stratifiable")
	}
}

func TestCheckCulpritCycleNoOpWhenNoBindableQuery(t *testing.T) {
	src := ".decl p(x:number)\n.decl q(x:number)\n.output q\np(1).\nq(x) :- p(x).\n"
	prog, errs := parser.Parse([]byte(src))
	if len(errs) != 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	result, err := CheckCulpritCycle(prog)
	if err != nil {
		t.Fatalf("CheckCulpritCycle error: %v", err)
	}
	if !result.NoBindableQuery || !result.Stratifiable {
		t.Fatalf("expected a no-op, trivially-stratifiable result, got %+v", result)
	}
}
