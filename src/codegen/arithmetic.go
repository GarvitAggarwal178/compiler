package codegen

import (
	"fmt"
	"strings"

	"dlc/src/ast"
)

// Fixed-width octal escapes avoid C's greedy hex escapes and preserve UTF-8
// bytes without depending on the C compiler's source character set.
func cStringLiteral(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '"' || c == '\\':
			b.WriteByte('\\')
			b.WriteByte(c)
		case c >= 32 && c < 127:
			b.WriteByte(c)
		default:
			fmt.Fprintf(&b, "\\%03o", c)
		}
	}
	b.WriteByte('"')
	return b.String()
}

func (g *generator) validExpr(a ast.Arith, bound map[string]string) string {
	switch v := a.(type) {
	case *ast.UnaryExpr:
		return g.validExpr(v.X, bound)
	case *ast.BinaryExpr:
		parts := []string{g.validExpr(v.Left, bound), g.validExpr(v.Right, bound)}
		if v.Op == "/" || v.Op == "%" {
			parts = append(parts, fmt.Sprintf("((%s)!=0)", g.groundExpr(v.Right, bound)))
		}
		return "(" + strings.Join(parts, " && ") + ")"
	}
	return "1"
}

func (g *generator) validTerms(terms []ast.Term, bound map[string]string) string {
	var parts []string
	for _, term := range terms {
		if a, ok := term.(ast.Arith); ok {
			if part := g.validExpr(a, bound); part != "1" {
				parts = append(parts, part)
			}
		}
	}
	if len(parts) == 0 {
		return "1"
	}
	return strings.Join(parts, " && ")
}

func (g *generator) withValidTerms(w *indentWriter, terms []ast.Term, bound map[string]string, cont func()) {
	guard := g.validTerms(terms, bound)
	if guard != "1" {
		w.line("if (%s) {", guard)
		w.level++
	}
	cont()
	if guard != "1" {
		w.level--
		w.line("}")
	}
}
