package sema

import (
	"fmt"

	"dlc/src/ast"
	"dlc/src/token"
)

// Allowedness is the category for every diagnostic this file produces.
const Allowedness Category = "allowedness"

// CheckAllowedness implements exactly the fixpoint definition recorded
// in docs/DECISIONS.md, derived empirically from probe cases a-o
// (docs/reports/J1-allowedness-probe.md, docs/reports/
// night02-T1-allowedness.md):
//
//	G0 = variables occurring as an argument of a positive body atom.
//	G(i+1) = G(i) U { V : the body contains a constraint V = E or E = V
//	         where V is a bare variable and vars(E) subset-of G(i) }.
//	G = the limit of the sequence.
//	A clause is allowed iff every variable occurring anywhere in the
//	clause is in G.
//
// See DESIGN.md for the four asymmetries this implementation is
// deliberately shaped around, each pinned by a specific probe case.
func CheckAllowedness(prog *ast.Program) []Diagnostic {
	var diags []Diagnostic
	for _, c := range prog.Clauses {
		diags = append(diags, checkClauseAllowedness(c)...)
	}
	return diags
}

func checkClauseAllowedness(c *ast.Clause) []Diagnostic {
	g := map[string]bool{}

	// G0: every variable occurring as (or within) an argument of a
	// POSITIVE body atom. A NegatedAtom is a distinct Go type from Atom
	// (ast/DESIGN.md's sum-type-as-interface design), so this type
	// switch excludes negated atoms automatically -- negation never
	// grounds (probe case f, m), with no separate check needed to
	// enforce that.
	for _, lit := range c.Body {
		if atom, ok := lit.(*ast.Atom); ok {
			for _, term := range atom.Terms {
				collectVars(term, g)
			}
		}
	}

	// Fixpoint over `V = E` / `E = V` constraints. Only '=' contributes
	// (probe case c: inequalities never ground); the grounded side must
	// be a BARE variable (probe case j: `X+1 = Y` does not ground X even
	// though Y is bare and grounded -- X is not itself bare on its own
	// side); no arithmetic inversion (probe case d: `Y = X+1` with Y
	// grounded does not ground X, because X is not the bare side of its
	// own equation). Looping until no change, not a single left-to-right
	// scan, handles chains and order-independence (probe cases b, k, l, o).
	for {
		changed := false
		for _, lit := range c.Body {
			constraint, ok := lit.(*ast.Constraint)
			if !ok || constraint.Op != "=" {
				continue
			}
			if groundBareSide(constraint.Left, constraint.Right, g) {
				changed = true
			}
			if groundBareSide(constraint.Right, constraint.Left, g) {
				changed = true
			}
		}
		if !changed {
			break
		}
	}

	// The quantifier is over EVERY variable in the clause -- head
	// included, and body-only variables that appear only in a
	// constraint or only inside a negation (probe cases h, m), not only
	// head variables.
	allVars := map[string]token.Span{}
	collectAllVars(c, allVars)

	var diags []Diagnostic
	for name, span := range allVars {
		if !g[name] {
			diags = append(diags, Diagnostic{
				Span: span, Category: Allowedness,
				Message: fmt.Sprintf("variable %q is not range-restricted", name),
			})
		}
	}
	return diags
}

// groundBareSide grounds `side`'s variable into g if side is a bare
// *ast.Var not already in g, and every variable in other is already in
// g. Returns whether it made a change.
func groundBareSide(side, other ast.Arith, g map[string]bool) bool {
	v, ok := side.(*ast.Var)
	if !ok || g[v.Name] {
		return false
	}
	if !varsSubsetOf(other, g) {
		return false
	}
	g[v.Name] = true
	return true
}

// varsSubsetOf reports whether every Var reachable inside a is already
// grounded in g. A constant-only expression (no vars at all) is
// vacuously true.
func varsSubsetOf(a ast.Arith, g map[string]bool) bool {
	switch v := a.(type) {
	case *ast.Var:
		return g[v.Name]
	case *ast.NumberLit, *ast.StringLit:
		return true
	case *ast.BinaryExpr:
		return varsSubsetOf(v.Left, g) && varsSubsetOf(v.Right, g)
	case *ast.UnaryExpr:
		return varsSubsetOf(v.X, g)
	}
	return true
}

// collectVars adds every Var name reachable inside term to into,
// recording each one's Span (first occurrence wins if the same name
// appears twice, which is fine for a diagnostic pointer).
func collectVars(term ast.Term, into map[string]bool) {
	switch v := term.(type) {
	case *ast.Var:
		into[v.Name] = true
	case *ast.BinaryExpr:
		collectVars(v.Left, into)
		collectVars(v.Right, into)
	case *ast.UnaryExpr:
		collectVars(v.X, into)
	case *ast.Wildcard, *ast.NumberLit, *ast.StringLit:
		// no variable to collect
	}
}

// collectAllVars walks the entire clause (head and body) collecting
// every variable's name and one representative Span.
func collectAllVars(c *ast.Clause, into map[string]token.Span) {
	collectVarsWithSpan(c.Head, into)
	for _, lit := range c.Body {
		switch v := lit.(type) {
		case *ast.Atom:
			collectVarsWithSpan(v, into)
		case *ast.NegatedAtom:
			collectVarsWithSpan(v.Atom, into)
		case *ast.Constraint:
			collectArithVarsWithSpan(v.Left, into)
			collectArithVarsWithSpan(v.Right, into)
		}
	}
}

func collectVarsWithSpan(a *ast.Atom, into map[string]token.Span) {
	for _, term := range a.Terms {
		collectArithVarsWithSpan(term, into)
	}
}

func collectArithVarsWithSpan(term ast.Term, into map[string]token.Span) {
	switch v := term.(type) {
	case *ast.Var:
		if _, seen := into[v.Name]; !seen {
			into[v.Name] = v.Sp
		}
	case *ast.BinaryExpr:
		collectArithVarsWithSpan(v.Left, into)
		collectArithVarsWithSpan(v.Right, into)
	case *ast.UnaryExpr:
		collectArithVarsWithSpan(v.X, into)
	case *ast.Wildcard, *ast.NumberLit, *ast.StringLit:
		// no variable
	}
}
