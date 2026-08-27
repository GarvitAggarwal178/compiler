package magicset

import (
	"strings"
	"testing"

	"dlc/src/parser"
	"dlc/src/sema"
)

// generateFor is the shared test helper: parse, adorn, generate, and
// return both the output program and its printed text.
func generateFor(t *testing.T, src string) (*sema.SymbolTable, string) {
	t.Helper()
	prog, errs := parser.Parse([]byte(src))
	if len(errs) != 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	st, diags := sema.BuildSymbolTable(prog)
	if len(diags) != 0 {
		t.Fatalf("symtab diagnostics: %v", diags)
	}
	qs := FindQueries(prog)
	if len(qs) == 0 {
		t.Fatal("expected a bindable query")
	}
	result, err := Adorn(prog, qs)
	if err != nil {
		t.Fatalf("Adorn error: %v", err)
	}
	out, _ := Generate(prog, st.Relations, result)
	return st, parser.Print(out)
}

// TestGenerateOutputPassesEverySemaCheck is §4's own "the returned
// program must itself be valid" requirement, checked directly rather
// than inferred from an end-to-end Souffle run.
func TestGenerateOutputPassesEverySemaCheck(t *testing.T) {
	prog, errs := parser.Parse([]byte(ancestorNonancestorSrc))
	if len(errs) != 0 {
		t.Fatalf("parse errors: %v", errs)
	}
	st, diags := sema.BuildSymbolTable(prog)
	if len(diags) != 0 {
		t.Fatalf("symtab diagnostics: %v", diags)
	}
	qs := FindQueries(prog)
	result, err := Adorn(prog, qs)
	if err != nil {
		t.Fatalf("Adorn error: %v", err)
	}
	out, origin := Generate(prog, st.Relations, result)

	// ancestor_bf/magic_ancestor_bf, not _bb: M4-SIPS.md §2's demand
	// relaxation collapses !ancestor(x,y)'s adornment (y's only binder is
	// person(y), a full-extent scan) before Generate ever sees it -- this
	// test pins RelationOrigin's mistagging-bug regression guard
	// (M2-M3-BUILD.md §7-prep commit), updated for the relaxed name.
	if origin["nonancestor_bf"] != "nonancestor" || origin["magic_ancestor_bf"] != "ancestor" || origin["ancestor_bf"] != "ancestor" {
		t.Fatalf("RelationOrigin mismatch: nonancestor_bf=%q, magic_ancestor_bf=%q, ancestor_bf=%q",
			origin["nonancestor_bf"], origin["magic_ancestor_bf"], origin["ancestor_bf"])
	}
	if origin["parent"] != "parent" || origin["person"] != "person" {
		t.Fatalf("expected EDB relations to map to themselves, got parent=%q person=%q", origin["parent"], origin["person"])
	}

	// Round-trips through the printer too -- the output must remain a
	// legal dlc program after being printed and reparsed, since that is
	// the actual path (dlc emit -> printed text -> Souffle) M3's
	// measurement protocol depends on.
	printed := parser.Print(out)
	reparsed, perrs := parser.Parse([]byte(printed))
	if len(perrs) != 0 {
		t.Fatalf("printed output does not reparse: %v\n%s", perrs, printed)
	}

	var semaDiags []sema.Diagnostic
	semaDiags = append(semaDiags, sema.CheckDeclType(reparsed)...)
	semaDiags = append(semaDiags, sema.CheckAllowedness(reparsed)...)
	stratDiags, _ := sema.CheckStratification(reparsed)
	semaDiags = append(semaDiags, stratDiags...)
	if len(semaDiags) != 0 {
		t.Fatalf("output program fails sema: %v\n%s", semaDiags, printed)
	}
}

// TestGenerateNoAtNames pins M2-M3-BUILD.md §1's naming rule: no
// '@'-prefixed identifiers anywhere in the output, unlike Souffle's own
// internal convention.
func TestGenerateNoAtNames(t *testing.T) {
	_, printed := generateFor(t, ancestorNonancestorSrc)
	if strings.Contains(printed, "@") {
		t.Fatalf("output contains '@', violating the plain-identifier naming rule:\n%s", printed)
	}
}

// TestGenerateMagicSeedFromQueryConstant checks the seed fact is emitted
// with the query's actual constant, not a placeholder.
func TestGenerateMagicSeedFromQueryConstant(t *testing.T) {
	_, printed := generateFor(t, ancestorNonancestorSrc)
	if !strings.Contains(printed, "magic_nonancestor_bf(1).") {
		t.Fatalf("expected seed fact magic_nonancestor_bf(1). in output:\n%s", printed)
	}
}

// TestGenerateQueryProjectionRewritten checks the final query-projection
// rule is redirected to the adorned relation, not the original.
func TestGenerateQueryProjectionRewritten(t *testing.T) {
	_, printed := generateFor(t, ancestorNonancestorSrc)
	if !strings.Contains(printed, "q_nonancestor(y) :- nonancestor_bf(1, y).") {
		t.Fatalf("expected the query projection rule to reference nonancestor_bf, got:\n%s", printed)
	}
	if strings.Contains(printed, "nonancestor(1, y)") {
		t.Fatalf("query projection rule still references the unadorned relation:\n%s", printed)
	}
}
