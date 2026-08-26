package sema

import (
	"testing"

	"dlc/src/parser"
)

// mustParseClean parses src and fails the test if the parser itself
// found errors -- these tests are about sema, not about re-testing the
// parser.
func mustParseClean(t *testing.T, src string) []Diagnostic {
	t.Helper()
	prog, errs := parser.Parse([]byte(src))
	if len(errs) != 0 {
		t.Fatalf("parser.Parse(%q) returned unexpected parse errors: %v", src, errs)
	}
	return CheckDeclType(prog)
}

func categories(diags []Diagnostic) []Category {
	cats := make([]Category, len(diags))
	for i, d := range diags {
		cats[i] = d.Category
	}
	return cats
}

func TestCleanProgramHasNoDiagnostics(t *testing.T) {
	diags := mustParseClean(t, ".decl foo(a:number)\n.input foo\n.decl bar(x:number)\n.output bar\nbar(x) :- foo(x), x > 0.\n")
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
}

// The three rejection-test/night02-T9 arity cases, tests/rejection/arity.py.
func TestArityTooFewArgs(t *testing.T) {
	diags := mustParseClean(t, ".decl foo(a:number, b:number)\n.input foo\n.decl bar(x:number)\n.output bar\nbar(x) :- foo(x).\n")
	if len(diags) != 1 || diags[0].Category != ArityMismatch {
		t.Fatalf("expected exactly one ArityMismatch, got %v", diags)
	}
}

func TestArityTooManyArgs(t *testing.T) {
	diags := mustParseClean(t, ".decl foo(a:number)\n.input foo\n.decl bar(x:number, y:number)\n.output bar\nbar(x, y) :- foo(x, y).\n")
	if len(diags) != 1 || diags[0].Category != ArityMismatch {
		t.Fatalf("expected exactly one ArityMismatch, got %v", diags)
	}
}

func TestArityMismatchOnFact(t *testing.T) {
	diags := mustParseClean(t, ".decl foo(a:number, b:number)\n.output foo\nfoo(1, 2, 3).\n")
	if len(diags) != 1 || diags[0].Category != ArityMismatch {
		t.Fatalf("expected exactly one ArityMismatch, got %v", diags)
	}
}

// The three rejection-test/night02-T9 type cases, tests/rejection/type.py.
func TestTypeNumberVsSymbolAcrossRule(t *testing.T) {
	diags := mustParseClean(t, ".decl foo(a:number, b:symbol)\n.input foo\n.decl bar(x:number)\n.output bar\nbar(y) :- foo(x, y).\n")
	if len(diags) != 1 || diags[0].Category != TypeMismatch {
		t.Fatalf("expected exactly one TypeMismatch, got %v", diags)
	}
}

func TestTypeSymbolInArithmetic(t *testing.T) {
	diags := mustParseClean(t, ".decl foo(a:symbol)\n.input foo\n.decl bar(x:number)\n.output bar\nbar(y) :- foo(x), y = x + 1.\n")
	if len(diags) != 1 || diags[0].Category != TypeMismatch {
		t.Fatalf("expected exactly one TypeMismatch, got %v", diags)
	}
}

func TestTypeMismatchedFactLiteral(t *testing.T) {
	diags := mustParseClean(t, ".decl foo(a:number)\n.output foo\nfoo(\"not_a_number\").\n")
	if len(diags) != 1 || diags[0].Category != TypeMismatch {
		t.Fatalf("expected exactly one TypeMismatch, got %v", diags)
	}
}

func TestUndeclaredRelationInBody(t *testing.T) {
	diags := mustParseClean(t, ".decl p(a:number)\n.output p\np(X) :- undeclared_rel(X).\n")
	if len(diags) != 1 || diags[0].Category != UndeclaredRelation {
		t.Fatalf("expected exactly one UndeclaredRelation, got %v", diags)
	}
}

func TestUndeclaredRelationViaInput(t *testing.T) {
	diags := mustParseClean(t, ".input nope\n")
	if len(diags) != 1 || diags[0].Category != UndeclaredRelation {
		t.Fatalf("expected exactly one UndeclaredRelation, got %v", diags)
	}
}

func TestDuplicateDeclaration(t *testing.T) {
	diags := mustParseClean(t, ".decl p(a:number)\n.decl p(a:number)\n.output p\np(1).\n")
	if len(diags) != 1 || diags[0].Category != DuplicateDecl {
		t.Fatalf("expected exactly one DuplicateDecl, got %v", diags)
	}
}

func TestInputAfterDeclIsNotDuplicate(t *testing.T) {
	// The normal, required pattern: schema decl + input marker for the
	// same relation is NOT a duplicate declaration.
	diags := mustParseClean(t, ".decl p(a:number)\n.input p\n.decl q(a:number)\n.output q\nq(x):-p(x).\n")
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics for decl+input pairing, got %v", diags)
	}
}

func TestBareComparisonDoesNotForceNumber(t *testing.T) {
	// X = "foo" must NOT be treated as forcing X to number just because
	// it appears on one side of a relop -- see DESIGN.md.
	diags := mustParseClean(t, ".decl p(a:symbol)\n.output p\np(X) :- X = \"foo\".\n")
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics for a bare symbol comparison, got %v", diags)
	}
}

func TestTypeErrorSpansAreNonZero(t *testing.T) {
	diags := mustParseClean(t, ".decl foo(a:number)\n.output foo\nfoo(\"x\").\n")
	if len(diags) != 1 {
		t.Fatalf("expected exactly one diagnostic, got %v", diags)
	}
	if diags[0].Span.Start.Line == 0 {
		t.Fatalf("expected a real span, got zero-value %v", diags[0].Span)
	}
}
