package parser

import (
	"strings"

	"dlc/src/ast"
)

// Print renders prog back to dlc source text. Built specifically to make
// §3.3 gate two (parse -> print -> reparse -> ast.Equal) checkable; see
// DESIGN.md for the full-parenthesization tradeoff that makes this
// printer trivially round-trip-correct at the cost of not being minimal.
func Print(prog *ast.Program) string {
	var sb strings.Builder
	for _, d := range prog.Decls {
		printDecl(&sb, d)
	}
	for _, c := range prog.Clauses {
		printClause(&sb, c)
	}
	return sb.String()
}

func printDecl(sb *strings.Builder, d *ast.Decl) {
	switch d.Kind {
	case ast.DeclRelation:
		sb.WriteString(".decl ")
		sb.WriteString(d.Name)
		sb.WriteString("(")
		for i, p := range d.Params {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(p.Name)
			sb.WriteString(":")
			sb.WriteString(p.Type)
		}
		sb.WriteString(")\n")
	case ast.DeclInput:
		sb.WriteString(".input ")
		sb.WriteString(d.Name)
		sb.WriteString("\n")
	case ast.DeclOutput:
		sb.WriteString(".output ")
		sb.WriteString(d.Name)
		sb.WriteString("\n")
	}
}

func printClause(sb *strings.Builder, c *ast.Clause) {
	printAtom(sb, c.Head)
	if len(c.Body) > 0 {
		sb.WriteString(" :- ")
		for i, lit := range c.Body {
			if i > 0 {
				sb.WriteString(", ")
			}
			printLiteral(sb, lit)
		}
	}
	sb.WriteString(".\n")
}

func printAtom(sb *strings.Builder, a *ast.Atom) {
	sb.WriteString(a.Name)
	sb.WriteString("(")
	for i, t := range a.Terms {
		if i > 0 {
			sb.WriteString(", ")
		}
		printTerm(sb, t)
	}
	sb.WriteString(")")
}

func printLiteral(sb *strings.Builder, lit ast.Literal) {
	switch v := lit.(type) {
	case *ast.Atom:
		printAtom(sb, v)
	case *ast.NegatedAtom:
		sb.WriteString("!")
		printAtom(sb, v.Atom)
	case *ast.Constraint:
		printArith(sb, v.Left, false)
		sb.WriteString(" ")
		sb.WriteString(v.Op)
		sb.WriteString(" ")
		printArith(sb, v.Right, false)
	}
}

func printTerm(sb *strings.Builder, t ast.Term) {
	switch v := t.(type) {
	case *ast.Wildcard:
		sb.WriteString("_")
	case ast.Arith:
		printArith(sb, v, false)
	}
}

// printArith prints a, wrapping it in parens when nested is true and a
// is itself a binary or unary expression. Every nested operand is
// printed with nested=true unconditionally -- not precedence-aware
// minimal parenthesization -- so the output is always unambiguous
// regardless of what operators are involved. See DESIGN.md.
func printArith(sb *strings.Builder, a ast.Arith, nested bool) {
	switch v := a.(type) {
	case *ast.BinaryExpr:
		if nested {
			sb.WriteString("(")
		}
		printArith(sb, v.Left, true)
		sb.WriteString(" ")
		sb.WriteString(v.Op)
		sb.WriteString(" ")
		printArith(sb, v.Right, true)
		if nested {
			sb.WriteString(")")
		}
	case *ast.UnaryExpr:
		if nested {
			sb.WriteString("(")
		}
		sb.WriteString(v.Op)
		printArith(sb, v.X, true)
		if nested {
			sb.WriteString(")")
		}
	case *ast.Var:
		sb.WriteString(v.Name)
	case *ast.NumberLit:
		sb.WriteString(v.Text)
	case *ast.StringLit:
		sb.WriteString("\"")
		sb.WriteString(escapeString(v.Value))
		sb.WriteString("\"")
	}
}

func escapeString(s string) string {
	var sb strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '"' || c == '\\' {
			sb.WriteByte('\\')
		}
		sb.WriteByte(c)
	}
	return sb.String()
}
