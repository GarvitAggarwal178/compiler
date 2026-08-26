// M2.3 -- magic rules and supplementary predicates. This is where the
// transform can silently pessimise (blueprint failure mode #2) if the V_i
// projection is skipped; implemented from the start, never shipped
// without it (M2-M3-BUILD.md §4).
package magicset

import (
	"fmt"
	"sort"

	"dlc/src/ast"
	"dlc/src/sema"
)

// Generate builds the full output *ast.Program from prog and an already-
// computed AdornResult: adorned rules with supplementary chains + magic
// rules + the query's seed fact + every EDB/.input declaration + every
// untouched IDB predicate's original rules, unchanged. The output is
// itself a plain, legal dlc program -- no '@'-prefixed names, ready to be
// printed (parser.Print) and fed to Soufflé (M2-M3-BUILD.md §1's naming
// section).
func Generate(prog *ast.Program, schemas map[string]*sema.RelationSchema, result *AdornResult) *ast.Program {
	g := &genState{schemas: schemas, declaredRel: map[string]bool{}, out: &ast.Program{}}

	// Every original declaration survives unchanged -- including the
	// original (now possibly rule-less) name of an adorned predicate;
	// an unused declared relation is legal and harmless, and keeping it
	// avoids ever needing to prove no other clause still references it.
	for _, d := range prog.Decls {
		g.out.Decls = append(g.out.Decls, d)
		g.declaredRel[d.Name] = true
	}

	st := &sema.SymbolTable{Relations: schemas}

	// Untouched predicates: original clauses pass through byte-for-byte
	// (same AST nodes, not copies -- immutable from here on). The
	// query's own projection rule is excluded even though its head
	// relation (the `.output` relation) is itself untouched -- it gets
	// re-emitted, rewritten, by emitQueryProjection below; including
	// both would leave the ORIGINAL (dead, since the predicate it reads
	// no longer has any defining rules) clause alongside the real one.
	untouchedNames := sortedKeys(result.Untouched)
	for _, name := range untouchedNames {
		for _, c := range prog.Clauses {
			if c.Head.Name == name && c != result.Query.ProjectionRule {
				g.out.Clauses = append(g.out.Clauses, c)
			}
		}
	}

	// Adorned predicates, processed in worklist order (deterministic,
	// and matches the order a human reading the gate's own report sees
	// them discovered in).
	for _, key := range result.Order {
		for ruleIdx, ar := range result.Rules[key] {
			varTypes := sema.ClauseVarTypes(st, ar.Source)
			g.emitAdornedRule(key, ruleIdx, ar, varTypes)
		}
	}

	// The query's own seed fact: magic_q^α0(c̄) from the query atom's
	// constant arguments.
	g.emitSeed(result.Query)

	// The original query-projection rule, rewritten to reference the
	// adorned relation instead of the original predicate name -- same
	// convention the hand-guarded files already use (q_nonancestor(y):-
	// nonancestor_bf(1,y), not nonancestor(1,y)).
	g.emitQueryProjection(result.Query)

	return g.out
}

type genState struct {
	schemas     map[string]*sema.RelationSchema
	declaredRel map[string]bool
	out         *ast.Program
}

func (g *genState) declareRelation(name string, types []string) {
	if g.declaredRel[name] {
		return
	}
	g.declaredRel[name] = true
	var params []ast.Param
	for i, t := range types {
		params = append(params, ast.Param{Name: fmt.Sprintf("c%d", i), Type: t})
	}
	g.out.Decls = append(g.out.Decls, &ast.Decl{Kind: ast.DeclRelation, Name: name, Params: params})
}

func (g *genState) addClause(head *ast.Atom, body []ast.Literal) {
	g.out.Clauses = append(g.out.Clauses, &ast.Clause{Head: head, Body: body})
}

// emitAdornedRule emits one AdornedRule's full contribution: the
// supplementary chain sup_0..sup_n, a magic rule per IDB occurrence
// (positive or negated, treated identically -- §5), and the final
// p^alpha(t̄):-sup_n(V_n) rule.
func (g *genState) emitAdornedRule(key adornedKey, ruleIdx int, ar *AdornedRule, varTypes map[string]string) {
	n := len(ar.OrderedBody)
	headVars := map[string]bool{}
	for _, t := range ar.Source.Head.Terms {
		collectVarNames(t, headVars)
	}

	// V[k] = BoundAfter[k] INTERSECT (vars occurring in OrderedBody[k:]
	// INCLUSIVE of position k itself, UNION vars(t̄)). Inclusive of
	// position k is what makes the magic rule for the literal AT k
	// correct: that literal's own bound-position variables must survive
	// into V[k] (the checkpoint immediately BEFORE it is processed), or
	// the magic rule referencing sup[k] would use a variable sup[k]
	// never carries.
	V := make([][]string, n+1)
	for k := 0; k <= n; k++ {
		needed := map[string]bool{}
		for name := range headVars {
			needed[name] = true
		}
		for _, lit := range ar.OrderedBody[k:] {
			for _, name := range varsInLit(lit) {
				needed[name] = true
			}
		}
		var vk []string
		for _, name := range ar.BoundAfter[k] {
			if needed[name] {
				vk = append(vk, name)
			}
		}
		V[k] = vk
	}

	supName := func(k int) string { return fmt.Sprintf("sup_%s_r%d_%d", key.RelName(), ruleIdx, k) }
	varTypesFor := func(names []string) []string {
		types := make([]string, len(names))
		for i, n := range names {
			types[i] = varTypes[n]
		}
		return types
	}

	// sup_0(V_0) :- magic_p^alpha(bound(t̄,alpha)).
	// bound(t̄,alpha)'s ARGS reuse the head's actual terms at alpha's
	// bound positions (var or constant, whichever is there) -- NOT V_0,
	// which is names-only and used for sup_0's own head/param list.
	boundHeadTerms := boundTermsAt(ar.Source.Head.Terms, adornFromKey(key))
	g.declareRelation(supName(0), varTypesFor(V[0]))
	g.addClause(mkAtomFromNames(supName(0), V[0]), []ast.Literal{mkAtomFromTerms(key.MagicRelName(), boundHeadTerms)})

	for k := 1; k <= n; k++ {
		lit := rewriteLiteralForAdornment(ar, k-1)
		g.declareRelation(supName(k), varTypesFor(V[k]))
		body := []ast.Literal{mkAtomFromNames(supName(k-1), V[k-1]), lit}
		g.addClause(mkAtomFromNames(supName(k), V[k]), body)
	}

	// Magic rules: one per IDB occurrence (both polarities), fed from
	// the sup checkpoint immediately BEFORE that occurrence's own
	// literal position.
	for _, occ := range ar.Occurrences {
		k := occ.literalIndex
		beta := adornFromKey(occ.target)
		boundArgs := boundTermsAt(occ.atom.Terms, beta)
		g.declareRelation(occ.target.MagicRelName(), typesForBoundPositions(g.schemas[occ.target.pred], beta))
		g.addClause(mkAtomFromTerms(occ.target.MagicRelName(), boundArgs), []ast.Literal{mkAtomFromNames(supName(k), V[k])})
	}

	// p^alpha(t̄) :- sup_n(V_n). Reuses the ORIGINAL head atom's terms
	// exactly (Datalog binds head<->body by variable NAME, not by
	// matching argument order across different atoms, so V_n's internal
	// ordering need not match t̄'s).
	g.declareRelation(key.RelName(), typesInOrder(g.schemas[key.pred]))
	g.addClause(&ast.Atom{Name: key.RelName(), Terms: ar.Source.Head.Terms}, []ast.Literal{mkAtomFromNames(supName(n), V[n])})
}

// emitSeed emits the query's magic seed fact: magic_q^alpha0(c̄), c̄ being
// the query atom's constant arguments (the bound positions of alpha0 are
// exactly the positions FindQuery required to be constants).
func (g *genState) emitSeed(q *QueryInfo) {
	beta := adornFromKey(q.Key)
	boundArgs := boundTermsAt(q.QueryAtom.Terms, beta)
	g.declareRelation(q.Key.MagicRelName(), typesForBoundPositions(g.schemas[q.Key.pred], beta))
	g.addClause(mkAtomFromTerms(q.Key.MagicRelName(), boundArgs), nil)
}

// emitQueryProjection re-emits the original query-projection rule
// (head unchanged -- it was never adorned, it is the `.output` relation
// itself) with its single body atom's name redirected to the adorned
// relation.
func (g *genState) emitQueryProjection(q *QueryInfo) {
	renamed := &ast.Atom{Name: q.Key.RelName(), Terms: q.QueryAtom.Terms}
	g.addClause(q.ProjectionRule.Head, []ast.Literal{renamed})
}

func rewriteLiteralForAdornment(ar *AdornedRule, idx int) ast.Literal {
	lit := ar.OrderedBody[idx]
	for _, occ := range ar.Occurrences {
		if occ.literalIndex == idx {
			renamed := &ast.Atom{Name: occ.target.RelName(), Terms: occ.atom.Terms}
			if occ.negated {
				return &ast.NegatedAtom{Atom: renamed}
			}
			return renamed
		}
	}
	return lit
}

func boundTermsAt(terms []ast.Term, beta Adornment) []ast.Term {
	var out []ast.Term
	for i, t := range terms {
		if i < len(beta) && beta[i] {
			out = append(out, t)
		}
	}
	return out
}

func typesInOrder(schema *sema.RelationSchema) []string {
	out := make([]string, len(schema.Params))
	for i, p := range schema.Params {
		out[i] = p.Type
	}
	return out
}

func typesForBoundPositions(schema *sema.RelationSchema, beta Adornment) []string {
	var out []string
	for i, p := range schema.Params {
		if i < len(beta) && beta[i] {
			out = append(out, p.Type)
		}
	}
	return out
}

func mkAtomFromNames(name string, varNames []string) *ast.Atom {
	terms := make([]ast.Term, len(varNames))
	for i, n := range varNames {
		terms[i] = &ast.Var{Name: n}
	}
	return &ast.Atom{Name: name, Terms: terms}
}

func mkAtomFromTerms(name string, terms []ast.Term) *ast.Atom {
	return &ast.Atom{Name: name, Terms: terms}
}

func collectVarNames(t ast.Term, into map[string]bool) {
	switch v := t.(type) {
	case *ast.Var:
		into[v.Name] = true
	case *ast.BinaryExpr:
		collectVarNames(v.Left, into)
		collectVarNames(v.Right, into)
	case *ast.UnaryExpr:
		collectVarNames(v.X, into)
	}
}

func sortedKeys(m map[string]bool) []string {
	var out []string
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
