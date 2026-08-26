// Package eval implements dlc's evaluators: naive (§3.8) and semi-naive
// (§3.9). fallback.go, in this same package, is Lane A (docs/M1-BUILD.md
// §1) -- everything else here is Lane B.
package eval

import (
	"fmt"

	"dlc/src/ast"
	"dlc/src/ir"
	"dlc/src/sema"
)

// Evaluator holds the runtime state one evaluation run needs: every
// relation's storage and the shared string interner. Bundled into a
// struct (rather than threading both through every helper function's
// parameter list) purely for readability -- nothing here is safe for
// concurrent use, and nothing needs to be.
//
// DerivationAttempts counts every candidate head tuple a clause body
// match successfully builds, BEFORE Relation.Insert's dedup -- unlike
// ir.RelationStats (distinct tuples only, matching this project's T
// metric since Phase 0), this number is expected to differ between
// RunNaive and RunSemiNaive on the same program: naive re-derives
// already-known tuples on every pass until its combined fixpoint stops
// changing, semi-naive's whole point is to avoid exactly that. See
// DESIGN.md for why the *tuple-count* T_naive/T_semi-naive is expected
// to be identical (both compute the same minimal model) and why this
// separate counter is what actually shows semi-naive's saved work.
type Evaluator struct {
	Relations          map[string]*ir.Relation
	Strings            *ir.StringTable
	DerivationAttempts int64
}

// NewEvaluator creates one Relation per declared relation (schema-only
// or otherwise) so every relation named anywhere in the program has
// storage before evaluation starts, whether or not it turns out to have
// any tuples.
func NewEvaluator(schemas map[string]*sema.RelationSchema) *Evaluator {
	e := &Evaluator{Relations: map[string]*ir.Relation{}, Strings: ir.NewStringTable()}
	for name, schema := range schemas {
		e.Relations[name] = ir.NewRelation(name, len(schema.Params))
	}
	return e
}

// RunNaive evaluates prog to completion: per stratum (ascending, from
// strata), iterate every clause whose head is in that stratum to a
// naive fixpoint (re-evaluate every clause against the full current
// contents of every relation, repeat until nothing new is derived) --
// no delta/semi-naive optimization, correctness first (§3.8's own
// framing). Every rule-derived tuple is recorded via
// RecordIterationInsert(0) (see ir/DESIGN.md: naive evaluation has no
// semi-naive "seed vs delta" distinction, so everything goes in one
// bucket) -- Total() still gives the right derived-tuple count.
func (e *Evaluator) RunNaive(prog *ast.Program, strata map[string]int) {
	byStratum := map[int][]*ast.Clause{}
	maxStratum := 0
	for _, c := range prog.Clauses {
		s := strata[c.Head.Name] // 0 for a relation buildPrecedenceGraph never saw as a head elsewhere -- can't happen, c.Head.Name is always a node
		byStratum[s] = append(byStratum[s], c)
		if s > maxStratum {
			maxStratum = s
		}
	}
	for s := 0; s <= maxStratum; s++ {
		e.fixpoint(byStratum[s])
	}
}

func (e *Evaluator) fixpoint(clauses []*ast.Clause) {
	for {
		changed := false
		for _, c := range clauses {
			ordered := safeOrder(c.Body)
			headRel := e.Relations[c.Head.Name]
			e.evalBody(ordered, 0, map[string]ir.Value{}, nil, func(bindings map[string]ir.Value) {
				tup, ok := e.buildTuple(c.Head, bindings)
				if !ok {
					return
				}
				e.DerivationAttempts++
				if headRel.Insert(tup) {
					headRel.RecordIterationInsert(0)
					changed = true
				}
			})
		}
		if !changed {
			return
		}
	}
}

// safeOrder reorders body so every negated atom and constraint is only
// scheduled once every variable it needs is already bound by something
// scheduled before it -- a plain left-to-right walk of an allowed
// clause's body as originally written can violate this (allowedness's
// fixpoint, sema/DESIGN.md, is explicitly order-independent; naive
// evaluation is not). Positive atoms are always immediately schedulable
// (they only ever add bindings). See DESIGN.md.
func safeOrder(body []ast.Literal) []ast.Literal {
	bound := map[string]bool{}
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
			// Cannot happen for a clause that passed allowedness (a
			// valid order is guaranteed to exist) -- defensive fallback
			// rather than an infinite loop or a panic if it somehow does.
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
		return allVarsBound(v.Atom.Terms, bound)
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

func allVarsBound(terms []ast.Term, bound map[string]bool) bool {
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

func addBoundVars(lit ast.Literal, bound map[string]bool) {
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
	// NegatedAtom binds nothing new -- allowedness already required
	// every one of its variables to be bound before it can be scheduled.
}

// evalBody walks body[idx:], trying every matching tuple/value at each
// step and recursing, calling cont once with the accumulated bindings
// for every way to satisfy the whole (ordered) body -- a plain nested-
// loop join, no smarter plan than the literal order safeOrder produced.
//
// overrideIdx, when non-nil, redirects one specific POSITIVE atom
// OCCURRENCE (keyed by its index in body, not by relation name) to read
// from a delta relation (§3.9's semi-naive Δ-rewrite) instead of
// e.Relations[name] (the full accumulated relation). Keyed by position,
// not name, deliberately: a self-join (the same relation appearing twice
// in one recursive rule's body, e.g. `p(x,y):-p(x,z),p(z,y).`) needs
// separate Δ-rewrite variants that redirect exactly one *occurrence* at
// a time -- keying by name would redirect every occurrence of that
// relation at once and silently drop the new-old / old-new combinations
// semi-naive evaluation exists to still catch. Negated atoms never
// consult overrideIdx -- they only ever reference a strictly lower,
// already-stable stratum (sema/stratify.go's computeStrata guarantees
// this), so "delta vs full" is not a meaningful distinction for them.
// nil overrideIdx (naive evaluation, §3.8) behaves exactly as before
// this parameter existed.
func (e *Evaluator) evalBody(body []ast.Literal, idx int, bindings map[string]ir.Value, overrideIdx map[int]*ir.Relation, cont func(map[string]ir.Value)) {
	if idx == len(body) {
		cont(bindings)
		return
	}
	switch v := body[idx].(type) {
	case *ast.Atom:
		rel := e.Relations[v.Name]
		if overrideIdx != nil {
			if r2, ok := overrideIdx[idx]; ok {
				rel = r2
			}
		}
		if rel == nil {
			return // undeclared relation -- sema should have already rejected this program
		}
		for _, tup := range e.candidateTuples(rel, v.Terms, bindings) {
			if newBindings, ok := e.tryUnify(v.Terms, tup, bindings); ok {
				e.evalBody(body, idx+1, newBindings, overrideIdx, cont)
			}
		}
	case *ast.NegatedAtom:
		rel := e.Relations[v.Atom.Name]
		tup, ok := e.groundTuple(v.Atom.Terms, bindings)
		if ok && rel != nil && !relContains(rel, tup) {
			e.evalBody(body, idx+1, bindings, overrideIdx, cont)
		}
	case *ast.Constraint:
		if newBindings, ok := e.evalConstraint(v, bindings); ok {
			e.evalBody(body, idx+1, newBindings, overrideIdx, cont)
		}
	}
}

func (e *Evaluator) candidateTuples(rel *ir.Relation, terms []ast.Term, bindings map[string]ir.Value) []ir.Tuple {
	if len(terms) > 0 {
		if v, ok := e.boundValueOf(terms[0], bindings); ok {
			return rel.LookupByFirstColumn(v)
		}
	}
	return rel.All()
}

func (e *Evaluator) boundValueOf(term ast.Term, bindings map[string]ir.Value) (ir.Value, bool) {
	a, ok := term.(ast.Arith)
	if !ok {
		return ir.Value{}, false // Wildcard: never "bound"
	}
	return e.evalArith(a, bindings)
}

func (e *Evaluator) tryUnify(terms []ast.Term, tup ir.Tuple, bindings map[string]ir.Value) (map[string]ir.Value, bool) {
	newBindings := copyBindings(bindings)
	for i, term := range terms {
		switch v := term.(type) {
		case *ast.Wildcard:
			continue
		case *ast.Var:
			if existing, ok := newBindings[v.Name]; ok {
				if existing != tup[i] {
					return nil, false
				}
			} else {
				newBindings[v.Name] = tup[i]
			}
		default:
			val, ok := e.evalArith(term.(ast.Arith), newBindings)
			if !ok || val != tup[i] {
				return nil, false
			}
		}
	}
	return newBindings, true
}

// groundTuple evaluates every term fully; used for a negated atom, which
// allowedness guarantees is fully ground by the time safeOrder schedules
// it (no unbound variable can remain).
func (e *Evaluator) groundTuple(terms []ast.Term, bindings map[string]ir.Value) (ir.Tuple, bool) {
	tup := make(ir.Tuple, len(terms))
	for i, term := range terms {
		a, ok := term.(ast.Arith)
		if !ok {
			return nil, false // a Wildcard inside a negated atom's terms -- degenerate, not attempted
		}
		v, ok := e.evalArith(a, bindings)
		if !ok {
			return nil, false
		}
		tup[i] = v
	}
	return tup, true
}

func relContains(rel *ir.Relation, tup ir.Tuple) bool {
	if len(tup) == 0 {
		return rel.Len() > 0
	}
	for _, cand := range rel.LookupByFirstColumn(tup[0]) {
		if tupleEqual(cand, tup) {
			return true
		}
	}
	return false
}

func tupleEqual(a, b ir.Tuple) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (e *Evaluator) buildTuple(head *ast.Atom, bindings map[string]ir.Value) (ir.Tuple, bool) {
	tup := make(ir.Tuple, len(head.Terms))
	for i, term := range head.Terms {
		a, ok := term.(ast.Arith)
		if !ok {
			// Wildcard in head position: real Soufflé rejects this
			// outright (docs/reports/night02-T2-hostile.md,
			// semantic_wildcard_in_head.dl) -- sema doesn't check for it
			// yet (out of §3.4-3.6's scope as implemented), so the
			// evaluator falls back to a fixed value rather than
			// panicking on a construct that should never have reached
			// evaluation in the first place.
			tup[i] = ir.NumberValue(0)
			continue
		}
		v, ok := e.evalArith(a, bindings)
		if !ok {
			return nil, false
		}
		tup[i] = v
	}
	return tup, true
}

func (e *Evaluator) evalArith(a ast.Arith, bindings map[string]ir.Value) (ir.Value, bool) {
	switch v := a.(type) {
	case *ast.Var:
		val, ok := bindings[v.Name]
		return val, ok
	case *ast.NumberLit:
		return ir.NumberValue(v.Value), true
	case *ast.StringLit:
		return ir.StringValue(e.Strings.Intern(v.Value)), true
	case *ast.UnaryExpr:
		x, ok := e.evalArith(v.X, bindings)
		if !ok || x.IsString {
			return ir.Value{}, false
		}
		return ir.NumberValue(-x.Num), true
	case *ast.BinaryExpr:
		l, ok1 := e.evalArith(v.Left, bindings)
		r, ok2 := e.evalArith(v.Right, bindings)
		if !ok1 || !ok2 || l.IsString || r.IsString {
			return ir.Value{}, false
		}
		switch v.Op {
		case "+":
			return ir.NumberValue(l.Num + r.Num), true
		case "-":
			return ir.NumberValue(l.Num - r.Num), true
		case "*":
			return ir.NumberValue(l.Num * r.Num), true
		case "/":
			if r.Num == 0 {
				return ir.Value{}, false
			}
			return ir.NumberValue(l.Num / r.Num), true
		case "%":
			if r.Num == 0 {
				return ir.Value{}, false
			}
			return ir.NumberValue(l.Num % r.Num), true
		}
	}
	return ir.Value{}, false
}

func (e *Evaluator) evalConstraint(c *ast.Constraint, bindings map[string]ir.Value) (map[string]ir.Value, bool) {
	if c.Op == "=" {
		if v, ok := c.Left.(*ast.Var); ok {
			if _, already := bindings[v.Name]; !already {
				val, ok2 := e.evalArith(c.Right, bindings)
				if !ok2 {
					return nil, false
				}
				nb := copyBindings(bindings)
				nb[v.Name] = val
				return nb, true
			}
		}
		if v, ok := c.Right.(*ast.Var); ok {
			if _, already := bindings[v.Name]; !already {
				val, ok2 := e.evalArith(c.Left, bindings)
				if !ok2 {
					return nil, false
				}
				nb := copyBindings(bindings)
				nb[v.Name] = val
				return nb, true
			}
		}
	}
	lv, ok1 := e.evalArith(c.Left, bindings)
	rv, ok2 := e.evalArith(c.Right, bindings)
	if !ok1 || !ok2 {
		return nil, false
	}
	if e.compareValues(lv, rv, c.Op) {
		return bindings, true
	}
	return nil, false
}

func (e *Evaluator) compareValues(l, r ir.Value, op string) bool {
	if l.IsString || r.IsString {
		ls, rs := e.valueString(l), e.valueString(r)
		switch op {
		case "=":
			return ls == rs
		case "!=":
			return ls != rs
		case "<":
			return ls < rs
		case "<=":
			return ls <= rs
		case ">":
			return ls > rs
		case ">=":
			return ls >= rs
		}
		return false
	}
	switch op {
	case "=":
		return l.Num == r.Num
	case "!=":
		return l.Num != r.Num
	case "<":
		return l.Num < r.Num
	case "<=":
		return l.Num <= r.Num
	case ">":
		return l.Num > r.Num
	case ">=":
		return l.Num >= r.Num
	}
	return false
}

func (e *Evaluator) valueString(v ir.Value) string {
	if v.IsString {
		return e.Strings.Lookup(v.StrID)
	}
	return fmt.Sprintf("%d", v.Num)
}

func copyBindings(b map[string]ir.Value) map[string]ir.Value {
	nb := make(map[string]ir.Value, len(b)+1)
	for k, v := range b {
		nb[k] = v
	}
	return nb
}
