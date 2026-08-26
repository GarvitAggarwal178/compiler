// M2.2 -- sideways information passing. Decides body literal order, which
// decides which adornments arise, which decides how much the transform
// saves.
package magicset

import "dlc/src/ast"

// OrderBody returns body reordered for adornment purposes: left-to-right
// in source order, with constraints pulled forward as soon as every
// variable they need is already bound, and negated literals pushed back
// until every variable they need is grounded (required -- allowedness
// already guarantees such an ordering exists; §3's own gate). initBound
// seeds the starting bound set from the head's b-position variables under
// the current adornment -- SIPS must know what the CALLER already
// guarantees is bound, not just what the body itself grounds, since that
// is the entire point of demand propagation.
//
// Chosen deliberately as the simplest defensible strategy, not a
// cost-based one (M2-M3-BUILD.md §3: "a wrong [cost-based SIPS] is
// indistinguishable from a right one without a cost model you do not
// have"). Implemented as its own small copy of the same multi-pass
// greedy-schedule algorithm eval.safeOrder and codegen.safeOrderForCodegen
// already use for an analogous reason (scheduling a body so
// negation/constraints only run once grounded) -- not shared across
// packages, matching the precedent those two already set (codegen's own
// copy is called out in its DESIGN.md as "small, deliberate duplication").
func OrderBody(body []ast.Literal, initBound map[string]bool) []ast.Literal {
	bound := map[string]bool{}
	for k := range initBound {
		bound[k] = true
	}
	var scheduled []ast.Literal
	remaining := append([]ast.Literal{}, body...)
	for len(remaining) > 0 {
		var stillRemaining []ast.Literal
		progressed := false
		for _, lit := range remaining {
			if canSchedule(lit, bound) {
				scheduled = append(scheduled, lit)
				addBoundVars(lit, bound)
				progressed = true
			} else {
				stillRemaining = append(stillRemaining, lit)
			}
		}
		remaining = stillRemaining
		if !progressed {
			// Cannot happen for a clause that passed allowedness (a valid
			// order is guaranteed to exist) -- defensive fallback, not a
			// panic, matching eval.safeOrder's own precedent.
			scheduled = append(scheduled, remaining...)
			break
		}
	}
	return scheduled
}

func canSchedule(lit ast.Literal, bound map[string]bool) bool {
	switch v := lit.(type) {
	case *ast.Atom:
		return true
	case *ast.NegatedAtom:
		return allTermsBound(v.Atom.Terms, bound)
	case *ast.Constraint:
		return canScheduleConstraint(v, bound)
	}
	return true
}

func canScheduleConstraint(c *ast.Constraint, bound map[string]bool) bool {
	if c.Op == "=" {
		if v, ok := c.Left.(*ast.Var); ok && !bound[v.Name] && arithVarsBound(c.Right, bound) {
			return true
		}
		if v, ok := c.Right.(*ast.Var); ok && !bound[v.Name] && arithVarsBound(c.Left, bound) {
			return true
		}
	}
	return arithVarsBound(c.Left, bound) && arithVarsBound(c.Right, bound)
}

func allTermsBound(terms []ast.Term, bound map[string]bool) bool {
	for _, t := range terms {
		if a, ok := t.(ast.Arith); ok {
			if !arithVarsBound(a, bound) {
				return false
			}
		}
	}
	return true
}

func arithVarsBound(a ast.Arith, bound map[string]bool) bool {
	switch v := a.(type) {
	case *ast.Var:
		return bound[v.Name]
	case *ast.NumberLit, *ast.StringLit:
		return true
	case *ast.BinaryExpr:
		return arithVarsBound(v.Left, bound) && arithVarsBound(v.Right, bound)
	case *ast.UnaryExpr:
		return arithVarsBound(v.X, bound)
	}
	return true
}

// addBoundVars extends bound with every variable lit grounds: a positive
// atom's terms unconditionally (it can only ever add bindings), or a
// constraint's grounded side (mirrors sema.CheckAllowedness's own
// groundBareSide -- see that function's doc comment for the exact
// asymmetries: only '=' contributes, the grounded side must be a bare
// variable, no arithmetic inversion). A negated atom never grounds
// anything (it is only ever a boolean test once its own variables are
// already bound).
func addBoundVars(lit ast.Literal, bound map[string]bool) {
	switch v := lit.(type) {
	case *ast.Atom:
		for _, t := range v.Terms {
			collectVars(t, bound)
		}
	case *ast.Constraint:
		if v.Op != "=" {
			return
		}
		if bv, ok := v.Left.(*ast.Var); ok && !bound[bv.Name] && arithVarsBound(v.Right, bound) {
			bound[bv.Name] = true
			return
		}
		if bv, ok := v.Right.(*ast.Var); ok && !bound[bv.Name] && arithVarsBound(v.Left, bound) {
			bound[bv.Name] = true
		}
	}
}

func collectVars(t ast.Term, bound map[string]bool) {
	switch v := t.(type) {
	case *ast.Var:
		bound[v.Name] = true
	case *ast.BinaryExpr:
		collectVars(v.Left, bound)
		collectVars(v.Right, bound)
	case *ast.UnaryExpr:
		collectVars(v.X, bound)
	}
}
