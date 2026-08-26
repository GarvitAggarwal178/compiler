// Package ast defines dlc's abstract syntax tree. Defined before any
// parse function exists (M1-BUILD.md §3.2) since sema (§3.4-3.6), eval
// (§3.8-3.9), and the pretty-printer (§3.3 gate two) are all written
// against this shape. Lane B.
package ast

import "dlc/src/token"

// Program is the root node: every declaration and every clause in one
// source file, in source order (order matters for nothing in this
// grammar's semantics, but preserving it is required for the pretty-
// printer's round-trip gate to be checkable at all).
type Program struct {
	Decls   []*Decl
	Clauses []*Clause
	Sp      token.Span
}

func (p *Program) Span() token.Span { return p.Sp }

// DeclKind distinguishes .decl / .input / .output.
type DeclKind int

const (
	DeclRelation DeclKind = iota
	DeclInput
	DeclOutput
)

// Param is one (name, type) pair in a .decl's parameter list.
type Param struct {
	Name string
	Type string // "number" | "symbol" -- sema's job to validate, not ast's
	Sp   token.Span
}

// Decl is one .decl / .input / .output directive. Params is empty for
// DeclInput/DeclOutput and for a zero-arity DeclRelation (M1-BUILD.md
// §3.3's grammar amendment: the term/param list is optional).
type Decl struct {
	Kind   DeclKind
	Name   string
	Params []Param
	Sp     token.Span
}

func (d *Decl) Span() token.Span { return d.Sp }

// Clause is a fact (Body == nil) or a rule (Body non-empty).
type Clause struct {
	Head *Atom
	Body []Literal
	Sp   token.Span
}

func (c *Clause) Span() token.Span { return c.Sp }

// Atom is a relation name applied to a term list. Terms is empty for a
// zero-arity atom.
type Atom struct {
	Name  string
	Terms []Term
	Sp    token.Span
}

func (a *Atom) Span() token.Span { return a.Sp }

// Literal is one body element: a positive atom, a negated atom, or a
// constraint. A sum type via interface, matching the grammar's own
// `literal ::= atom | '!' atom | constraint` production one-to-one.
type Literal interface {
	Span() token.Span
	literalNode()
}

func (*Atom) literalNode() {}

// NegatedAtom is `! atom`.
type NegatedAtom struct {
	Atom *Atom
	Sp   token.Span
}

func (n *NegatedAtom) Span() token.Span { return n.Sp }
func (*NegatedAtom) literalNode()       {}

// Constraint is `arith relop arith`. Op is the relop's literal text
// ("=", "!=", "<", "<=", ">", ">=") -- kept as text rather than a second
// enum since token.Kind already names these exactly and sema needs no
// more structure than a string switch provides.
type Constraint struct {
	Left  Arith
	Op    string
	Right Arith
	Sp    token.Span
}

func (c *Constraint) Span() token.Span { return c.Sp }
func (*Constraint) literalNode()       {}

// Term is one atom argument: an arith expression or a wildcard. Matches
// `term ::= arith | '_'` exactly.
type Term interface {
	Span() token.Span
	termNode()
}

// Wildcard is a bare `_` in term position.
type Wildcard struct {
	Sp token.Span
}

func (w *Wildcard) Span() token.Span { return w.Sp }
func (*Wildcard) termNode()          {}

// Arith is every node of the `arith` grammar production: binary op,
// unary op, variable, number, string. It embeds Term (every arith
// expression is also a valid term) rather than duplicating termNode on
// each concrete type.
type Arith interface {
	Term
	arithNode()
}

// BinaryExpr is `left op right` for one of + - * / %.
type BinaryExpr struct {
	Op    string
	Left  Arith
	Right Arith
	Sp    token.Span
}

func (b *BinaryExpr) Span() token.Span { return b.Sp }
func (*BinaryExpr) termNode()          {}
func (*BinaryExpr) arithNode()         {}

// UnaryExpr is unary `-x`. The grammar has only one unary operator, but
// Op is still carried as text (not a bare bool) for symmetry with
// BinaryExpr and because a diagnostic printing an UnaryExpr wants the
// operator's text anyway.
type UnaryExpr struct {
	Op string
	X  Arith
	Sp token.Span
}

func (u *UnaryExpr) Span() token.Span { return u.Sp }
func (*UnaryExpr) termNode()          {}
func (*UnaryExpr) arithNode()         {}

// Var is a variable reference (an identifier used as a term, not a
// relation name).
type Var struct {
	Name string
	Sp   token.Span
}

func (v *Var) Span() token.Span { return v.Sp }
func (*Var) termNode()          {}
func (*Var) arithNode()         {}

// NumberLit is an integer literal. Value is the parsed int64; Text is
// the original source digits, kept alongside Value so the pretty-printer
// can round-trip exactly (e.g. a source literal with leading zeros)
// without needing to re-derive source text from an int64.
type NumberLit struct {
	Value int64
	Text  string
	Sp    token.Span
}

func (n *NumberLit) Span() token.Span { return n.Sp }
func (*NumberLit) termNode()          {}
func (*NumberLit) arithNode()         {}

// StringLit is a string literal; Value holds the decoded (escapes
// resolved) contents, matching token.Token's STRING convention.
type StringLit struct {
	Value string
	Sp    token.Span
}

func (s *StringLit) Span() token.Span { return s.Sp }
func (*StringLit) termNode()          {}
func (*StringLit) arithNode()         {}
