package codegen

import (
	"fmt"
	"strconv"
	"strings"

	"dlc/src/ast"
	"dlc/src/sema"
)

// emitEvaluation emits evaluate(), called once from main. Fact clauses
// (Body == nil, whether from .input loading -- handled separately in
// load_facts -- or written directly in source) are inserted once,
// unconditionally, before any stratum's fixpoint loop; every stratum
// (ascending) gets its own "while (changed) { ...every clause whose
// head is in this stratum... }" naive fixpoint, matching eval.RunNaive's
// semantics exactly (§3.8 -- no semi-naive Δ-rewrite here, see
// DESIGN.md for the scoping).
func (g *generator) emitEvaluation(prog *ast.Program, strata *sema.StratumResult) {
	byStratum := map[int][]*ast.Clause{}
	maxStratum := 0
	var facts []*ast.Clause
	for _, c := range prog.Clauses {
		if len(c.Body) == 0 {
			facts = append(facts, c)
			continue
		}
		s := strata.Stratum[c.Head.Name]
		byStratum[s] = append(byStratum[s], c)
		if s > maxStratum {
			maxStratum = s
		}
	}

	g.w("static void evaluate(void) {")
	for _, c := range facts {
		g.emitFactClause(c)
	}
	for s := 0; s <= maxStratum; s++ {
		clauses := byStratum[s]
		if len(clauses) == 0 {
			continue
		}
		g.w("\t{")
		g.w("\t\tint changed = 1;")
		g.w("\t\twhile (changed) {")
		g.w("\t\t\tchanged = 0;")
		for _, c := range clauses {
			g.emitRuleClause(c)
		}
		g.w("\t\t}")
		g.w("\t}")
	}
	g.w("}")
	g.w("")
}

// emitFactClause emits a single unconditional insert for a body-less
// clause (a source fact, `p(1,2).`) -- every term must be a constant
// (no variable can be bound in a fact's head; the parser/allowedness
// pipeline already guarantees this for any program that reached
// codegen).
func (g *generator) emitFactClause(c *ast.Clause) {
	id := cIdent(c.Head.Name)
	schema := g.schemas[c.Head.Name]
	if len(schema.Params) == 0 {
		g.w("\t%s_present = 1;", id)
		return
	}
	g.w("\t{")
	g.w("\t\tTuple_%s t;", id)
	for i, term := range c.Head.Terms {
		expr := g.constExpr(term)
		g.w("\t\tt.c[%d] = %s;", i, expr)
	}
	g.w("\t\t%s_insert(t);", id)
	g.w("\t}")
}

func (g *generator) constExpr(term ast.Term) string {
	switch v := term.(type) {
	case *ast.NumberLit:
		return strconv.FormatInt(v.Value, 10)
	case *ast.StringLit:
		return fmt.Sprintf("intern(%q)", v.Value)
	}
	return "0" // unreachable for a fact clause post-parse; defensive, not a panic
}

// emitRuleClause emits one clause's contribution to its stratum's
// fixpoint round: safeOrder the body exactly as eval.safeOrder does
// (small, deliberate duplication -- see DESIGN.md), then recursively
// nest one C construct per literal.
func (g *generator) emitRuleClause(c *ast.Clause) {
	g.curVarTypes = sema.ClauseVarTypes(&sema.SymbolTable{Relations: g.schemas}, c)
	ordered := safeOrderForCodegen(c.Body)
	var sb indentWriter
	sb.level = 3
	g.emitLiteral(&sb, ordered, 0, map[string]string{}, c)
	g.sb.WriteString(sb.buf.String())
	g.curVarTypes = nil
}

// indentWriter is a tiny helper so the recursively-nested C blocks come
// out legible instead of all at one indent level.
type indentWriter struct {
	buf   strings.Builder
	level int
}

func (w *indentWriter) line(format string, args ...interface{}) {
	w.buf.WriteString(strings.Repeat("\t", w.level))
	fmt.Fprintf(&w.buf, format, args...)
	w.buf.WriteByte('\n')
}

// emitLiteral recursively emits body[idx:], one C construct per literal,
// nesting the continuation inside it -- the same structure
// eval.evalBody's recursive-continuation design has, just emitting C
// source text instead of calling a Go closure. bound maps a Datalog
// variable name to the C expression currently holding its value.
func (g *generator) emitLiteral(w *indentWriter, body []ast.Literal, idx int, bound map[string]string, headClause *ast.Clause) {
	if idx == len(body) {
		g.emitDerive(w, headClause, bound)
		return
	}
	switch v := body[idx].(type) {
	case *ast.Atom:
		g.emitAtomJoin(w, v, bound, func(nb map[string]string) {
			g.emitLiteral(w, body, idx+1, nb, headClause)
		})
	case *ast.NegatedAtom:
		id := cIdent(v.Atom.Name)
		schema := g.schemas[v.Atom.Name]
		w.line("{")
		w.level++
		if len(schema.Params) == 0 {
			w.line("if (!%s_present) {", id)
		} else {
			w.line("Tuple_%s neg_t;", id)
			for i, term := range v.Atom.Terms {
				w.line("neg_t.c[%d] = %s;", i, g.groundExprTerm(term, bound))
			}
			w.line("if (!%s_contains(neg_t)) {", id)
		}
		w.level++
		g.emitLiteral(w, body, idx+1, bound, headClause)
		w.level--
		w.line("}")
		w.level--
		w.line("}")
	case *ast.Constraint:
		g.emitConstraint(w, v, bound, func(nb map[string]string) {
			g.emitLiteral(w, body, idx+1, nb, headClause)
		})
	}
}

// emitAtomJoin emits either a hash-bucket walk (first term already
// bound/constant -- "with hash indices", §4 item 1's own wording) or a
// full scan (first term free), binding each term's variable and
// checking already-bound terms/constants for equality, then calls cont
// with the (possibly extended) bound map for the loop body.
func (g *generator) emitAtomJoin(w *indentWriter, a *ast.Atom, bound map[string]string, cont func(map[string]string)) {
	id := cIdent(a.Name)
	schema := g.schemas[a.Name]
	if len(schema.Params) == 0 {
		w.line("if (%s_present) {", id)
		w.level++
		cont(bound)
		w.level--
		w.line("}")
		return
	}

	firstBoundExpr, firstIsBound := g.boundExprOf(a.Terms[0], bound)
	rowVar := fmt.Sprintf("row_%s_%d", id, w.level)
	w.line("{")
	w.level++
	if firstIsBound {
		w.line("unsigned h = (unsigned)((uint64_t)(%s) %% %s_HASH_BUCKETS);", firstBoundExpr, id)
		w.line("for (HNode_%s *node = %s_buckets[h]; node; node = node->next) {", id, id)
		w.level++
		w.line("Tuple_%s %s = %s_data[node->row];", id, rowVar, id)
		w.line("if (%s.c[0] != (%s)) continue;", rowVar, firstBoundExpr)
	} else {
		w.line("for (size_t i_%s = 0; i_%s < %s_len; i_%s++) {", id, id, id, id)
		w.level++
		w.line("Tuple_%s %s = %s_data[i_%s];", id, rowVar, id, id)
	}

	nb := cloneBound(bound)
	guard := ""
	for i, term := range a.Terms {
		colExpr := fmt.Sprintf("%s.c[%d]", rowVar, i)
		if i == 0 && firstIsBound {
			continue // already checked above
		}
		switch tv := term.(type) {
		case *ast.Wildcard:
			continue
		case *ast.Var:
			if existing, ok := nb[tv.Name]; ok {
				guard += fmt.Sprintf(" && (%s) == (%s)", existing, colExpr)
			} else {
				nb[tv.Name] = colExpr
			}
		default:
			guard += fmt.Sprintf(" && (%s) == (%s)", colExpr, g.groundExprTerm(term, bound))
		}
	}
	if guard != "" {
		w.line("if (1%s) {", guard)
		w.level++
		cont(nb)
		w.level--
		w.line("}")
	} else {
		cont(nb)
	}

	w.level--
	w.line("}")
	w.level--
	w.line("}")
}

// emitConstraint mirrors eval.evalConstraint's grounding-vs-test split
// exactly (sema/DESIGN.md, eval/DESIGN.md): only "=" with one side a
// not-yet-bound bare Var and the other side fully bound is a grounding
// (a plain C declaration, no new block needed -- it stays in scope for
// the rest of this nesting level); everything else is a boolean test,
// wrapped in an if(){...cont...}.
func (g *generator) emitConstraint(w *indentWriter, c *ast.Constraint, bound map[string]string, cont func(map[string]string)) {
	if c.Op == "=" {
		if v, ok := c.Left.(*ast.Var); ok {
			if _, already := bound[v.Name]; !already {
				nb := cloneBound(bound)
				cVar := fmt.Sprintf("v_%s", sanitize(v.Name))
				w.line("int64_t %s = %s;", cVar, g.groundExpr(c.Right, bound))
				nb[v.Name] = cVar
				cont(nb)
				return
			}
		}
		if v, ok := c.Right.(*ast.Var); ok {
			if _, already := bound[v.Name]; !already {
				nb := cloneBound(bound)
				cVar := fmt.Sprintf("v_%s", sanitize(v.Name))
				w.line("int64_t %s = %s;", cVar, g.groundExpr(c.Left, bound))
				nb[v.Name] = cVar
				cont(nb)
				return
			}
		}
	}
	left := g.groundExpr(c.Left, bound)
	right := g.groundExpr(c.Right, bound)
	// NIGHT-BATCH-03 T8: <, <=, >, >= on symbol-typed operands must
	// compare the interned strings lexicographically, not the interned
	// ids by assignment order (a symbol's ir.Value IS its intern id --
	// src/ir/relation.go -- so a bare integer comparison here silently
	// compared assignment order until this fix). = and != are unaffected:
	// id equality is exactly string equality regardless of intern order,
	// so no change is needed for those two operators.
	if isOrderingOp(c.Op) && (g.isSymbolArith(c.Left) || g.isSymbolArith(c.Right)) {
		w.line("if (strcmp(str_lookup(%s), str_lookup(%s)) %s 0) {", left, right, cOp(c.Op))
	} else {
		w.line("if ((%s) %s (%s)) {", left, cOp(c.Op), right)
	}
	w.level++
	cont(bound)
	w.level--
	w.line("}")
}

func isOrderingOp(op string) bool {
	switch op {
	case "<", "<=", ">", ">=":
		return true
	}
	return false
}

// isSymbolArith reports whether a is definitely symbol-typed: a bare
// string literal, or a variable g.curVarTypes resolves to "symbol". Any
// other arith shape (BinaryExpr/UnaryExpr/NumberLit) is always "number"
// -- this grammar's arithmetic operators are only ever defined over
// number (sema/decltype.go's forceArithNumber).
func (g *generator) isSymbolArith(a ast.Arith) bool {
	switch v := a.(type) {
	case *ast.StringLit:
		return true
	case *ast.Var:
		return g.curVarTypes[v.Name] == "symbol"
	}
	return false
}

func cOp(op string) string {
	if op == "=" {
		return "=="
	}
	return op // != < <= > >= are already valid C operators
}

// emitDerive builds the head tuple from the fully-bound environment and
// inserts it, marking the stratum's fixpoint "changed" flag on a
// genuinely new insertion.
func (g *generator) emitDerive(w *indentWriter, c *ast.Clause, bound map[string]string) {
	id := cIdent(c.Head.Name)
	schema := g.schemas[c.Head.Name]
	if len(schema.Params) == 0 {
		w.line("if (!%s_present) { %s_present = 1; changed = 1; }", id, id)
		return
	}
	w.line("{")
	w.level++
	w.line("Tuple_%s t;", id)
	for i, term := range c.Head.Terms {
		w.line("t.c[%d] = %s;", i, g.groundExprTerm(term, bound))
	}
	w.line("if (%s_insert(t)) changed = 1;", id)
	w.level--
	w.line("}")
}

// groundExprTerm handles a head Term, which may be a Wildcard (real
// Soufflé rejects this, sema doesn't check for it yet -- same disclosed
// fallback as eval.buildTuple, see DESIGN.md: emit a fixed 0 rather than
// mis-generating C).
func (g *generator) groundExprTerm(term ast.Term, bound map[string]string) string {
	if a, ok := term.(ast.Arith); ok {
		return g.groundExpr(a, bound)
	}
	return "0"
}

// groundExpr renders an Arith fully evaluable in the current bound
// environment as a C expression string.
func (g *generator) groundExpr(a ast.Arith, bound map[string]string) string {
	switch v := a.(type) {
	case *ast.Var:
		if expr, ok := bound[v.Name]; ok {
			return expr
		}
		return "0" // unreachable post-allowedness/safeOrder for a real program
	case *ast.NumberLit:
		return strconv.FormatInt(v.Value, 10)
	case *ast.StringLit:
		return fmt.Sprintf("intern(%q)", v.Value)
	case *ast.UnaryExpr:
		return fmt.Sprintf("(-(%s))", g.groundExpr(v.X, bound))
	case *ast.BinaryExpr:
		return fmt.Sprintf("((%s) %s (%s))", g.groundExpr(v.Left, bound), v.Op, g.groundExpr(v.Right, bound))
	}
	return "0"
}

// boundExprOf reports the C expression for term if it is already fully
// evaluable (a bound Var, or any constant/arith expression over only
// already-bound vars) -- used to decide whether an atom join can use the
// hash index (first term bound) or must fall back to a full scan.
func (g *generator) boundExprOf(term ast.Term, bound map[string]string) (string, bool) {
	a, ok := term.(ast.Arith)
	if !ok {
		return "", false // Wildcard: never "bound"
	}
	if !arithFullyBound(a, bound) {
		return "", false
	}
	return g.groundExpr(a, bound), true
}

func arithFullyBound(a ast.Arith, bound map[string]string) bool {
	switch v := a.(type) {
	case *ast.Var:
		_, ok := bound[v.Name]
		return ok
	case *ast.NumberLit, *ast.StringLit:
		return true
	case *ast.UnaryExpr:
		return arithFullyBound(v.X, bound)
	case *ast.BinaryExpr:
		return arithFullyBound(v.Left, bound) && arithFullyBound(v.Right, bound)
	}
	return false
}

func cloneBound(b map[string]string) map[string]string {
	nb := make(map[string]string, len(b)+1)
	for k, v := range b {
		nb[k] = v
	}
	return nb
}

func sanitize(name string) string {
	var sb strings.Builder
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			sb.WriteRune(r)
		} else {
			sb.WriteByte('_')
		}
	}
	return sb.String()
}

// safeOrderForCodegen is eval.safeOrder's algorithm, reimplemented here
// rather than shared -- see DESIGN.md for why this small duplication was
// accepted instead of exporting eval's unexported helper.
func safeOrderForCodegen(body []ast.Literal) []ast.Literal {
	bound := map[string]bool{}
	var scheduled []ast.Literal
	remaining := append([]ast.Literal{}, body...)
	for len(remaining) > 0 {
		var stillRemaining []ast.Literal
		progressed := false
		for _, lit := range remaining {
			if canScheduleForCodegen(lit, bound) {
				scheduled = append(scheduled, lit)
				addBoundVarsForCodegen(lit, bound)
				progressed = true
			} else {
				stillRemaining = append(stillRemaining, lit)
			}
		}
		remaining = stillRemaining
		if !progressed {
			scheduled = append(scheduled, remaining...)
			break
		}
	}
	return scheduled
}

func canScheduleForCodegen(lit ast.Literal, bound map[string]bool) bool {
	switch v := lit.(type) {
	case *ast.Atom:
		return true
	case *ast.NegatedAtom:
		for _, t := range v.Atom.Terms {
			if a, ok := t.(ast.Arith); ok && !arithVarsBoundForCodegen(a, bound) {
				return false
			}
		}
		return true
	case *ast.Constraint:
		if v.Op == "=" {
			if vv, ok := v.Left.(*ast.Var); ok && !bound[vv.Name] && arithVarsBoundForCodegen(v.Right, bound) {
				return true
			}
			if vv, ok := v.Right.(*ast.Var); ok && !bound[vv.Name] && arithVarsBoundForCodegen(v.Left, bound) {
				return true
			}
		}
		return arithVarsBoundForCodegen(v.Left, bound) && arithVarsBoundForCodegen(v.Right, bound)
	}
	return true
}

func arithVarsBoundForCodegen(a ast.Arith, bound map[string]bool) bool {
	switch v := a.(type) {
	case *ast.Var:
		return bound[v.Name]
	case *ast.NumberLit, *ast.StringLit:
		return true
	case *ast.BinaryExpr:
		return arithVarsBoundForCodegen(v.Left, bound) && arithVarsBoundForCodegen(v.Right, bound)
	case *ast.UnaryExpr:
		return arithVarsBoundForCodegen(v.X, bound)
	}
	return true
}

func addBoundVarsForCodegen(lit ast.Literal, bound map[string]bool) {
	switch v := lit.(type) {
	case *ast.Atom:
		for _, t := range v.Terms {
			if vv, ok := t.(*ast.Var); ok {
				bound[vv.Name] = true
			}
		}
	case *ast.Constraint:
		if v.Op == "=" {
			if vv, ok := v.Left.(*ast.Var); ok {
				bound[vv.Name] = true
			}
			if vv, ok := v.Right.(*ast.Var); ok {
				bound[vv.Name] = true
			}
		}
	}
}
