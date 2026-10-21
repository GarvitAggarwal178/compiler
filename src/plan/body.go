// Package plan schedules executable literals and lowers arithmetic column tests
// only when ordinary binding order stalls. Source AST nodes are never mutated.
package plan

import (
	"dlc/src/ast"
	"dlc/src/token"
	"fmt"
	"sort"
)

type Step struct {
	SourceIndex             int
	Kind                    string // literal or arithmetic_filter
	Span                    token.Span
	BoundBefore, BoundAfter []string
}
type Result struct {
	Literals []ast.Literal
	Steps    []Step
	Lowered  bool
}
type item struct {
	literal ast.Literal
	source  int
	kind    string
}

func Body(body []ast.Literal, initial map[string]bool) (*Result, error) {
	original := make([]item, len(body))
	for i, l := range body {
		original[i] = item{l, i, "literal"}
	}
	if out, ok := schedule(original, initial); ok {
		return out, nil
	}
	// Positive compound terms are column tests, not inverse arithmetic.
	// Turn them into fresh column variables and deferred equalities only
	// when the existing stable ordering cannot run the original body.
	used := copySet(initial)
	for _, l := range body {
		walkLiteral(l, func(v *ast.Var) { used[v.Name] = true })
	}
	serial := 0
	fresh := func(sp token.Span) *ast.Var {
		for {
			name := fmt.Sprintf("dlc_expr_%d", serial)
			serial++
			if !used[name] {
				used[name] = true
				return &ast.Var{Name: name, Sp: sp}
			}
		}
	}
	var lowered []item
	changed := false
	for i, l := range body {
		a, ok := l.(*ast.Atom)
		if !ok {
			lowered = append(lowered, item{l, i, "literal"})
			continue
		}
		replacement := &ast.Atom{Name: a.Name, Terms: append([]ast.Term{}, a.Terms...), Sp: a.Sp}
		var filters []ast.Literal
		for j, t := range a.Terms {
			switch expr := t.(type) {
			case *ast.BinaryExpr:
				v := fresh(t.Span())
				replacement.Terms[j] = v
				filters = append(filters, &ast.Constraint{Left: v, Op: "=", Right: expr, Sp: t.Span()})
				changed = true
			case *ast.UnaryExpr:
				v := fresh(t.Span())
				replacement.Terms[j] = v
				filters = append(filters, &ast.Constraint{Left: v, Op: "=", Right: expr, Sp: t.Span()})
				changed = true
			}
		}
		lowered = append(lowered, item{replacement, i, "literal"})
		for _, f := range filters {
			lowered = append(lowered, item{f, i, "arithmetic_filter"})
		}
	}
	if changed {
		if out, ok := schedule(lowered, initial); ok {
			out.Lowered = true
			return out, nil
		}
	}
	return nil, fmt.Errorf("body cannot be scheduled without unbound inputs or arithmetic inversion")
}

func schedule(items []item, initial map[string]bool) (*Result, bool) {
	bound := copySet(initial)
	remaining := append([]item{}, items...)
	out := &Result{}
	for len(remaining) > 0 {
		next := []item{}
		progress := false
		for _, it := range remaining {
			if !ready(it.literal, bound) {
				next = append(next, it)
				continue
			}
			before := sorted(bound)
			Bind(it.literal, bound)
			out.Literals = append(out.Literals, it.literal)
			out.Steps = append(out.Steps, Step{SourceIndex: it.source, Kind: it.kind, Span: it.literal.Span(), BoundBefore: before, BoundAfter: sorted(bound)})
			progress = true
		}
		if !progress {
			return out, false
		}
		remaining = next
	}
	return out, true
}
func AtomInputsBound(a *ast.Atom, bound map[string]bool) bool {
	available := copySet(bound)
	for _, t := range a.Terms {
		if v, ok := t.(*ast.Var); ok {
			available[v.Name] = true
		}
	}
	for _, t := range a.Terms {
		if expr, ok := t.(ast.Arith); ok && !ArithBound(expr, available) {
			return false
		}
	}
	return true
}
func ready(l ast.Literal, b map[string]bool) bool {
	switch v := l.(type) {
	case *ast.Atom:
		return AtomInputsBound(v, b)
	case *ast.NegatedAtom:
		for _, t := range v.Atom.Terms {
			if a, ok := t.(ast.Arith); !ok || !ArithBound(a, b) {
				return false
			}
		}
		return true
	case *ast.Constraint:
		if v.Op == "=" {
			if x, ok := v.Left.(*ast.Var); ok && !b[x.Name] && ArithBound(v.Right, b) {
				return true
			}
			if x, ok := v.Right.(*ast.Var); ok && !b[x.Name] && ArithBound(v.Left, b) {
				return true
			}
		}
		return ArithBound(v.Left, b) && ArithBound(v.Right, b)
	}
	return false
}
func Bind(l ast.Literal, b map[string]bool) {
	switch v := l.(type) {
	case *ast.Atom:
		for _, t := range v.Terms {
			if x, ok := t.(*ast.Var); ok {
				b[x.Name] = true
			}
		}
	case *ast.Constraint:
		if v.Op == "=" {
			if x, ok := v.Left.(*ast.Var); ok && !b[x.Name] && ArithBound(v.Right, b) {
				b[x.Name] = true
				return
			}
			if x, ok := v.Right.(*ast.Var); ok && !b[x.Name] && ArithBound(v.Left, b) {
				b[x.Name] = true
			}
		}
	}
}
func ArithBound(a ast.Arith, b map[string]bool) bool {
	switch v := a.(type) {
	case *ast.Var:
		return b[v.Name]
	case *ast.NumberLit, *ast.StringLit:
		return true
	case *ast.UnaryExpr:
		return ArithBound(v.X, b)
	case *ast.BinaryExpr:
		return ArithBound(v.Left, b) && ArithBound(v.Right, b)
	}
	return false
}
func walkLiteral(l ast.Literal, visit func(*ast.Var)) {
	var term func(ast.Term)
	term = func(t ast.Term) {
		switch v := t.(type) {
		case *ast.Var:
			visit(v)
		case *ast.BinaryExpr:
			term(v.Left)
			term(v.Right)
		case *ast.UnaryExpr:
			term(v.X)
		}
	}
	switch v := l.(type) {
	case *ast.Atom:
		for _, t := range v.Terms {
			term(t)
		}
	case *ast.NegatedAtom:
		for _, t := range v.Atom.Terms {
			term(t)
		}
	case *ast.Constraint:
		term(v.Left)
		term(v.Right)
	}
}
func copySet(b map[string]bool) map[string]bool {
	out := map[string]bool{}
	for k, v := range b {
		if v {
			out[k] = true
		}
	}
	return out
}
func sorted(b map[string]bool) []string {
	out := []string{}
	for k, v := range b {
		if v {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}
