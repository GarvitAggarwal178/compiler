package guard

import (
	"testing"

	"dlc/src/parser"
	"dlc/src/sema"
)

func TestDecideCulpritCycleDeclinesAllThreeRelations(t *testing.T) {
	prog, errs := parser.Parse([]byte(culpritCycleSrc))
	if len(errs) != 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	result, err := Decide(prog)
	if err != nil {
		t.Fatalf("Decide error: %v", err)
	}
	for _, pred := range []string{"p", "q", "s"} {
		if !result.DeclinedRelations[pred] {
			t.Errorf("expected %q to be declined, DeclinedRelations=%v", pred, result.DeclinedRelations)
		}
	}
}

func TestDecideCulpritCycleFinalProgramStratifiesAndAgrees(t *testing.T) {
	prog, errs := parser.Parse([]byte(culpritCycleSrc))
	if len(errs) != 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	result, err := Decide(prog)
	if err != nil {
		t.Fatalf("Decide error: %v", err)
	}
	printed := parser.Print(result.FinalProgram)
	reparsed, perrs := parser.Parse([]byte(printed))
	if len(perrs) != 0 {
		t.Fatalf("final program does not reparse: %v\n%s", perrs, printed)
	}
	var semaDiags []sema.Diagnostic
	semaDiags = append(semaDiags, sema.CheckDeclType(reparsed)...)
	semaDiags = append(semaDiags, sema.CheckAllowedness(reparsed)...)
	stratDiags, _ := sema.CheckStratification(reparsed)
	semaDiags = append(semaDiags, stratDiags...)
	if len(semaDiags) != 0 {
		t.Fatalf("final mixed program fails sema: %v\n%s", semaDiags, printed)
	}
}

func TestDecideAcceptsAlreadyStratifiableTransform(t *testing.T) {
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
	result, err := Decide(prog)
	if err != nil {
		t.Fatalf("Decide error: %v", err)
	}
	if len(result.DeclinedRelations) != 0 {
		t.Fatalf("expected no declines for an already-stratifiable transform, got %v", result.DeclinedRelations)
	}
}

func TestDecideNoOpWhenNoBindableQuery(t *testing.T) {
	src := ".decl p(x:number)\n.decl q(x:number)\n.output q\np(1).\nq(x) :- p(x).\n"
	prog, errs := parser.Parse([]byte(src))
	if len(errs) != 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	result, err := Decide(prog)
	if err != nil {
		t.Fatalf("Decide error: %v", err)
	}
	if !result.NoBindableQuery || len(result.DeclinedRelations) != 0 {
		t.Fatalf("expected a no-op result, got %+v", result)
	}
}
