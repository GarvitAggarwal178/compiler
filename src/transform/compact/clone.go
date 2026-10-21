package compact

import "dlc/src/ast"

// All AST nodes, parameter slices and rewritten terms are copied. A name map
// applies simultaneous substitution; it never recursively rewrites a new name.
func cloneTerm(t ast.Term, names map[string]string) ast.Term {
	switch v := t.(type) {
	case *ast.Var:
		n := v.Name
		if x, ok := names[n]; ok {
			n = x
		}
		return &ast.Var{Name: n, Sp: v.Sp}
	case *ast.NumberLit:
		c := *v
		return &c
	case *ast.StringLit:
		c := *v
		return &c
	case *ast.Wildcard:
		c := *v
		return &c
	case *ast.UnaryExpr:
		return &ast.UnaryExpr{Op: v.Op, X: cloneTerm(v.X, names).(ast.Arith), Sp: v.Sp}
	case *ast.BinaryExpr:
		return &ast.BinaryExpr{Op: v.Op, Left: cloneTerm(v.Left, names).(ast.Arith), Right: cloneTerm(v.Right, names).(ast.Arith), Sp: v.Sp}
	}
	panic("compact: unknown term")
}
func cloneAtom(a *ast.Atom, names map[string]string) *ast.Atom {
	out := &ast.Atom{Name: a.Name, Sp: a.Sp}
	for _, t := range a.Terms {
		out.Terms = append(out.Terms, cloneTerm(t, names))
	}
	return out
}
func cloneLiteral(l ast.Literal, names map[string]string) ast.Literal {
	switch v := l.(type) {
	case *ast.Atom:
		return cloneAtom(v, names)
	case *ast.NegatedAtom:
		return &ast.NegatedAtom{Atom: cloneAtom(v.Atom, names), Sp: v.Sp}
	case *ast.Constraint:
		return &ast.Constraint{Op: v.Op, Left: cloneTerm(v.Left, names).(ast.Arith), Right: cloneTerm(v.Right, names).(ast.Arith), Sp: v.Sp}
	}
	panic("compact: unknown literal")
}
func cloneClause(c *ast.Clause) *ast.Clause {
	out := &ast.Clause{Head: cloneAtom(c.Head, nil), Sp: c.Sp}
	for _, l := range c.Body {
		out.Body = append(out.Body, cloneLiteral(l, nil))
	}
	return out
}
func cloneProgram(p *ast.Program) *ast.Program {
	out := &ast.Program{Sp: p.Sp}
	for _, d := range p.Decls {
		c := *d
		c.Params = append([]ast.Param(nil), d.Params...)
		out.Decls = append(out.Decls, &c)
	}
	for _, c := range p.Clauses {
		out.Clauses = append(out.Clauses, cloneClause(c))
	}
	return out
}
func termVars(t ast.Term, out map[string]bool) {
	switch v := t.(type) {
	case *ast.Var:
		out[v.Name] = true
	case *ast.UnaryExpr:
		termVars(v.X, out)
	case *ast.BinaryExpr:
		termVars(v.Left, out)
		termVars(v.Right, out)
	}
}
func literalVars(l ast.Literal, out map[string]bool) {
	switch v := l.(type) {
	case *ast.Atom:
		for _, t := range v.Terms {
			termVars(t, out)
		}
	case *ast.NegatedAtom:
		literalVars(v.Atom, out)
	case *ast.Constraint:
		termVars(v.Left, out)
		termVars(v.Right, out)
	}
}
