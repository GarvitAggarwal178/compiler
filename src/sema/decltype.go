// Package sema implements dlc's semantic analysis: declaration/arity/type
// checking (§3.4), allowedness (§3.5), and source-program stratification
// (§3.6). Lane B -- the *transformed*-program culprit-cycle detector is a
// different thing entirely and is Lane A (docs/M1-BUILD.md §1); nothing
// in this package is that.
package sema

import (
	"fmt"

	"dlc/src/ast"
	"dlc/src/token"
)

// Category classifies a Diagnostic by which of the four rejection
// grounds it belongs to (docs/reports/night02-T9-diagnostics.md's
// catalogue) -- message text need not match Soufflé's verbatim, but the
// classification must, per §3.4's own instruction.
type Category string

const (
	UndeclaredRelation Category = "undeclared_relation"
	DuplicateDecl      Category = "duplicate_declaration"
	ArityMismatch      Category = "arity_mismatch"
	TypeMismatch       Category = "type_mismatch"
)

type Diagnostic struct {
	Span     token.Span
	Category Category
	Message  string
}

// RelationSchema is one relation's declared shape, from its *single*
// authoritative .decl (schema-defining) declaration.
type RelationSchema struct {
	Name     string
	Params   []ast.Param
	DeclSpan token.Span
}

// SymbolTable maps relation name to its schema. Built once per program
// from every DeclRelation decl; a second DeclRelation for the same name
// is a DuplicateDecl diagnostic, not a second table entry (the first
// definition wins, matching Soufflé's own "Previous definition" framing).
type SymbolTable struct {
	Relations map[string]*RelationSchema
}

// BuildSymbolTable walks every .decl/.input/.output in prog. A
// .input/.output whose name has no matching schema .decl is an
// UndeclaredRelation diagnostic -- .input/.output only mark an
// *already*-declared relation as externally loaded/exposed, they do not
// themselves introduce a schema (see DESIGN.md for why this is not the
// same thing as "duplicate declaration").
func BuildSymbolTable(prog *ast.Program) (*SymbolTable, []Diagnostic) {
	st := &SymbolTable{Relations: map[string]*RelationSchema{}}
	var diags []Diagnostic
	for _, d := range prog.Decls {
		if d.Kind != ast.DeclRelation {
			continue
		}
		if existing, ok := st.Relations[d.Name]; ok {
			diags = append(diags, Diagnostic{
				Span: d.Sp, Category: DuplicateDecl,
				Message: fmt.Sprintf("relation %q redeclared (previously declared at %s)", d.Name, existing.DeclSpan.Start),
			})
			continue
		}
		st.Relations[d.Name] = &RelationSchema{Name: d.Name, Params: d.Params, DeclSpan: d.Sp}
	}
	for _, d := range prog.Decls {
		if d.Kind == ast.DeclRelation {
			continue
		}
		if _, ok := st.Relations[d.Name]; !ok {
			diags = append(diags, Diagnostic{
				Span: d.Sp, Category: UndeclaredRelation,
				Message: fmt.Sprintf("%s for relation %q, which has no .decl", declKindText(d.Kind), d.Name),
			})
		}
	}
	return st, diags
}

func declKindText(k ast.DeclKind) string {
	if k == ast.DeclInput {
		return ".input"
	}
	return ".output"
}

// CheckDeclType runs declaration/arity/type checking over prog and
// returns every Diagnostic found. Building the symbol table's own
// diagnostics (duplicate decl, .input/.output for an undeclared
// relation) are included alongside per-clause diagnostics.
func CheckDeclType(prog *ast.Program) []Diagnostic {
	st, diags := BuildSymbolTable(prog)
	for _, c := range prog.Clauses {
		diags = append(diags, checkClause(st, c)...)
	}
	return diags
}

// clauseChecker carries the per-clause variable-type environment built
// while checking one clause. Types do not persist across clauses --
// each clause's variables are local to it.
type clauseChecker struct {
	diags        []Diagnostic
	varTypes     map[string]string
	varFirstSpan map[string]token.Span
}

// ClauseVarTypes returns the same per-clause variable-type environment
// checkClause builds internally, for a caller (codegen -- NIGHT-BATCH-03
// T8) that needs to know a variable's declared type (`symbol` vs
// `number`) and has no other way to compute it without duplicating this
// logic. Only meaningful for a clause that has already passed
// CheckDeclType with zero diagnostics; a rejected clause's returned map
// may be incomplete (checkAtomOccurrence returns early on an arity
// mismatch, per its own comment) since nothing calls this on a rejected
// program in practice.
func ClauseVarTypes(st *SymbolTable, c *ast.Clause) map[string]string {
	return runClauseChecker(st, c).varTypes
}

func checkClause(st *SymbolTable, c *ast.Clause) []Diagnostic {
	return runClauseChecker(st, c).diags
}

func runClauseChecker(st *SymbolTable, c *ast.Clause) *clauseChecker {
	cc := &clauseChecker{varTypes: map[string]string{}, varFirstSpan: map[string]token.Span{}}
	// Body first, then head: mirrors the natural "grounded by the body,
	// used by the head" reading (matches how Soufflé's own diagnostic
	// text is phrased, docs/reports/night02-T9-diagnostics.md) -- purely
	// cosmetic, detection is order-independent.
	for _, lit := range c.Body {
		switch v := lit.(type) {
		case *ast.Atom:
			cc.checkAtomOccurrence(st, v)
		case *ast.NegatedAtom:
			cc.checkAtomOccurrence(st, v.Atom)
		case *ast.Constraint:
			cc.checkConstraintSide(v.Left)
			cc.checkConstraintSide(v.Right)
		}
	}
	cc.checkAtomOccurrence(st, c.Head)
	return cc
}

func (cc *clauseChecker) checkAtomOccurrence(st *SymbolTable, a *ast.Atom) {
	schema, ok := st.Relations[a.Name]
	if !ok {
		cc.diags = append(cc.diags, Diagnostic{
			Span: a.Sp, Category: UndeclaredRelation,
			Message: fmt.Sprintf("relation %q is not declared", a.Name),
		})
		return
	}
	if len(a.Terms) != len(schema.Params) {
		cc.diags = append(cc.diags, Diagnostic{
			Span: a.Sp, Category: ArityMismatch,
			Message: fmt.Sprintf("%q declared with arity %d, used here with %d argument(s)",
				a.Name, len(schema.Params), len(a.Terms)),
		})
		return // arity already wrong; checking types against it would misattribute a second, spurious error
	}
	for i, term := range a.Terms {
		cc.checkTermAgainstType(term, schema.Params[i].Type)
	}
}

func (cc *clauseChecker) checkTermAgainstType(term ast.Term, wantType string) {
	switch v := term.(type) {
	case *ast.Wildcard:
		// A wildcard carries no type obligation of its own.
	case *ast.Var:
		cc.requireVarType(v.Name, wantType, v.Sp)
	case *ast.NumberLit:
		if wantType != "number" {
			cc.diags = append(cc.diags, Diagnostic{
				Span: v.Sp, Category: TypeMismatch,
				Message: fmt.Sprintf("expected %s, got a number literal", wantType),
			})
		}
	case *ast.StringLit:
		if wantType != "symbol" {
			cc.diags = append(cc.diags, Diagnostic{
				Span: v.Sp, Category: TypeMismatch,
				Message: fmt.Sprintf("expected %s, got a string literal", wantType),
			})
		}
	case *ast.BinaryExpr:
		if wantType != "number" {
			cc.diags = append(cc.diags, Diagnostic{
				Span: v.Sp, Category: TypeMismatch,
				Message: fmt.Sprintf("expected %s, got an arithmetic expression (always number)", wantType),
			})
		}
		cc.forceArithNumber(v)
	case *ast.UnaryExpr:
		if wantType != "number" {
			cc.diags = append(cc.diags, Diagnostic{
				Span: v.Sp, Category: TypeMismatch,
				Message: fmt.Sprintf("expected %s, got an arithmetic expression (always number)", wantType),
			})
		}
		cc.forceArithNumber(v)
	}
}

// checkConstraintSide handles one side of `arith relop arith`. A bare
// Var/NumberLit/StringLit used directly as a whole comparison side does
// NOT by itself force a type -- its type (if any) comes from wherever
// else in the clause actually declares it. Only descending into an
// actual arithmetic operator forces "number" on that operator's operands
// (arithmetic is only defined over number in this grammar). See
// DESIGN.md for why this distinction matters (a bare `X = Y` must not be
// treated the same as `X = Y + 1`).
func (cc *clauseChecker) checkConstraintSide(a ast.Arith) {
	switch v := a.(type) {
	case *ast.BinaryExpr, *ast.UnaryExpr:
		cc.forceArithNumber(v)
	}
}

// forceArithNumber requires every Var reachable inside an arithmetic
// subtree to be "number", and flags a StringLit found there directly
// (a string literal can never be typed number).
func (cc *clauseChecker) forceArithNumber(a ast.Arith) {
	switch v := a.(type) {
	case *ast.Var:
		cc.requireVarType(v.Name, "number", v.Sp)
	case *ast.NumberLit:
		// already number, nothing to require
	case *ast.StringLit:
		cc.diags = append(cc.diags, Diagnostic{
			Span: v.Sp, Category: TypeMismatch,
			Message: "string literal used in an arithmetic expression (requires number)",
		})
	case *ast.BinaryExpr:
		cc.forceArithNumber(v.Left)
		cc.forceArithNumber(v.Right)
	case *ast.UnaryExpr:
		cc.forceArithNumber(v.X)
	}
}

func (cc *clauseChecker) requireVarType(name, wantType string, span token.Span) {
	if existing, ok := cc.varTypes[name]; ok {
		if existing != wantType {
			cc.diags = append(cc.diags, Diagnostic{
				Span: span, Category: TypeMismatch,
				Message: fmt.Sprintf("variable %q used as %s here, but as %s elsewhere in the same clause (%s)",
					name, wantType, existing, cc.varFirstSpan[name]),
			})
		}
		return
	}
	cc.varTypes[name] = wantType
	cc.varFirstSpan[name] = span
}
