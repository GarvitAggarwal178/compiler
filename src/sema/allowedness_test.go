package sema

import (
	"os"
	"testing"

	"dlc/src/parser"
)

// checkAllowedness parses and allowedness-checks src, failing the test
// if the parser itself found errors.
func checkAllowedness(t *testing.T, src string) []Diagnostic {
	t.Helper()
	prog, errs := parser.Parse([]byte(src))
	if len(errs) != 0 {
		t.Fatalf("parser.Parse(%q) returned unexpected parse errors: %v", src, errs)
	}
	return CheckAllowedness(prog)
}

func expectAllowed(t *testing.T, src string) {
	t.Helper()
	diags := checkAllowedness(t, src)
	if len(diags) != 0 {
		t.Fatalf("expected %q to be allowed, got diagnostics: %v", src, diags)
	}
}

func expectRejected(t *testing.T, src string) {
	t.Helper()
	diags := checkAllowedness(t, src)
	if len(diags) == 0 {
		t.Fatalf("expected %q to be rejected as not allowed, got no diagnostics", src)
	}
	for _, d := range diags {
		if d.Category != Allowedness {
			t.Fatalf("expected only Allowedness diagnostics, got %v", d)
		}
	}
}

// The 15 probe programs (a-o), verdicts per docs/reports/
// J1-allowedness-probe.md (a-g) and docs/reports/night02-T1-allowedness.md
// (h-o), run against the real Soufflé 2.5 this session. §3.5's own gate.
func TestAllowednessProbeCases(t *testing.T) {
	verdict := map[string]bool{ // true = accept (allowed)
		"a": true, "b": true, "c": false, "d": false, "e": true, "f": false, "g": true,
		"h": false, "i": true, "j": false, "k": true, "l": true, "m": false, "n": true, "o": true,
	}
	for letter, allowed := range verdict {
		letter, allowed := letter, allowed
		t.Run(letter, func(t *testing.T) {
			path := "../../tests/programs/allowedness_probe_" + letter + ".dl"
			src, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("cannot read %s: %v", path, err)
			}
			if allowed {
				expectAllowed(t, string(src))
			} else {
				expectRejected(t, string(src))
			}
		})
	}
}

// The 4 allowedness cases in tests/rejection/allowedness.py, restated
// directly here (that file is Python, not importable from a Go test) --
// same programs, same expected outcome (all 4 reject).
func TestAllowednessRejectionCorpus(t *testing.T) {
	cases := map[string]string{
		"head_var_unbound": ".decl foo(a:number)\n.input foo\n.decl bar(x:number, y:number)\n.output bar\nbar(x, y) :- foo(x).\n",
		"var_only_in_negation": ".decl foo(a:number)\n.input foo\n.decl baz(a:number, b:number)\n.input baz\n.decl bar(x:number, y:number)\n.output bar\nbar(x, y) :- foo(x), !baz(x, y).\n",
		"var_only_in_constraint": ".decl foo(a:number)\n.input foo\n.decl bar(x:number, y:number)\n.output bar\nbar(x, y) :- foo(x), y > 0.\n",
		"equation_rhs_not_bound": ".decl foo(a:number)\n.input foo\n.decl bar(x:number, y:number)\n.output bar\nbar(x, y) :- foo(x), x = y + 1.\n",
	}
	for name, src := range cases {
		name, src := name, src
		t.Run(name, func(t *testing.T) {
			expectRejected(t, src)
		})
	}
}

func TestAllowednessDiagnosticSpanIsNonZero(t *testing.T) {
	diags := checkAllowedness(t, ".decl q(a:number)\n.decl p(a:number)\np(X) :- q(X), Y > 3.\n")
	if len(diags) != 1 {
		t.Fatalf("expected exactly one diagnostic, got %v", diags)
	}
	if diags[0].Span.Start.Line == 0 {
		t.Fatalf("expected a real span, got zero-value %v", diags[0].Span)
	}
}
