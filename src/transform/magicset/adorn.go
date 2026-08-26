// M2.1 -- adornment. Package magicset implements the magic-set transform:
// adornment (this file), SIPS (sips.go), and magic rule + supplementary
// predicate generation (rules.go). Paper: Beeri & Ramakrishnan, "On the
// Power of Magic" (the base transformation); this package implements the
// generalized-adornment + supplementary-predicate form described there,
// not Balbin et al.'s per-literal-segment treatment of negation --
// negated IDB atoms are adorned and given magic rules by the exact same
// uniform mechanism as positive ones (M2-M3-BUILD.md §5's central claim:
// this is sound precisely because allowedness forces every negated
// atom's adornment to be all-bound). Where the paper leaves a choice open
// -- SIPS strategy -- this package picks the simplest defensible one
// (sips.go's own doc comment) and does not implement a cost model.
package magicset

import (
	"fmt"
	"sort"

	"dlc/src/ast"
)

// Adornment is one bit per predicate argument position: true = bound on
// entry. Represented as a slice, not a bitset -- arity is bounded at 14
// in this corpus (M2-M3-BUILD.md §2), nowhere near where a slice's
// overhead would matter.
type Adornment []bool

// String renders an Adornment as "bf"/"ff"/"bb"/... -- also used as the
// map key / name-suffix, so two Adornments with the same bits always
// render identically regardless of how they were constructed.
func (a Adornment) String() string {
	s := make([]byte, len(a))
	for i, b := range a {
		if b {
			s[i] = 'b'
		} else {
			s[i] = 'f'
		}
	}
	return string(s)
}

func (a Adornment) equal(o Adornment) bool {
	if len(a) != len(o) {
		return false
	}
	for i := range a {
		if a[i] != o[i] {
			return false
		}
	}
	return true
}

// adornedKey identifies one (predicate, adornment) worklist item / output
// relation -- the unit of dedup for the worklist and the map key for
// AdornResult.Rules.
type adornedKey struct {
	pred   string
	adorn  string
	arity  int
}

func keyOf(pred string, a Adornment) adornedKey {
	return adornedKey{pred: pred, adorn: a.String(), arity: len(a)}
}

// RelName is the adorned relation's name in the OUTPUT program: plain
// identifiers, no '@' (M2-M3-BUILD.md §1's "Naming" section -- dlc's
// output must be legal in dlc's own grammar).
func (k adornedKey) RelName() string { return k.pred + "_" + k.adorn }

// MagicRelName is k's magic predicate's name.
func (k adornedKey) MagicRelName() string { return "magic_" + k.pred + "_" + k.adorn }

// occurrence records one IDB atom's adorned occurrence within a specific
// rule body, in SIPS order -- positive or negated, both treated
// identically (M2-M3-BUILD.md §5).
type occurrence struct {
	literalIndex int // position within OrderedBody
	negated      bool
	atom         *ast.Atom
	target       adornedKey // the (predicate, adornment) this occurrence demands
}

// AdornedRule is one (predicate, adornment) pair's contribution from one
// of that predicate's original defining clauses.
type AdornedRule struct {
	Key         adornedKey
	Source      *ast.Clause    // the original clause (unordered body)
	OrderedBody []ast.Literal  // SIPS order
	Occurrences []occurrence   // IDB atom occurrences within OrderedBody (both polarities)
	BoundAfter  [][]string     // BoundAfter[i] = ordered, deduped var list bound after processing OrderedBody[:i]; BoundAfter[0] = the head's bound-position variables under Key.adorn
}

// QueryInfo identifies the program's bindable query: a rule whose body is
// a single positive atom with at least one constant argument, whose head
// is a `.output` relation (M2-M3-BUILD.md §2's own definition).
type QueryInfo struct {
	ProjectionRule *ast.Clause
	QueryAtom      *ast.Atom
	Key            adornedKey
}

// AdornResult is adorn.go's whole output: the worklist's processing
// order (for the §2 gate's "iteration count"), every adorned predicate's
// contributing rules, and which predicates were never reached (their
// original rules pass through unchanged).
type AdornResult struct {
	Query      *QueryInfo
	Order      []adornedKey
	Rules      map[adornedKey][]*AdornedRule
	Untouched  map[string]bool // predicate names never adorned
	Iterations int             // worklist pops -- same as len(Order), named separately for gate clarity
}

const worklistCap = 10000

// FindQuery identifies the program's bindable query per M2-M3-BUILD.md
// §2: a rule whose head is a `.output` relation and whose body is a
// single positive atom with at least one constant argument, projected
// from that output relation. Deterministic: if more than one candidate
// exists, the first in source order wins (this corpus never has more
// than one; ties are broken this way rather than left to map iteration
// order). Returns nil if none exists -- the caller's job is then a no-op
// pass-through, not an error (§2: "expected on the positive fragment").
func FindQuery(prog *ast.Program) *QueryInfo {
	outputs := map[string]bool{}
	for _, d := range prog.Decls {
		if d.Kind == ast.DeclOutput {
			outputs[d.Name] = true
		}
	}
	for _, c := range prog.Clauses {
		if !outputs[c.Head.Name] || len(c.Body) != 1 {
			continue
		}
		atom, ok := c.Body[0].(*ast.Atom)
		if !ok {
			continue
		}
		adorn := make(Adornment, len(atom.Terms))
		hasConst := false
		for i, t := range atom.Terms {
			if isConstant(t) {
				adorn[i] = true
				hasConst = true
			}
		}
		if !hasConst {
			continue
		}
		return &QueryInfo{
			ProjectionRule: c,
			QueryAtom:      atom,
			Key:            keyOf(atom.Name, adorn),
		}
	}
	return nil
}

func isConstant(t ast.Term) bool {
	switch t.(type) {
	case *ast.NumberLit, *ast.StringLit:
		return true
	}
	return false
}

// Adorn runs the worklist algorithm (M2-M3-BUILD.md §2) starting from
// query, over every clause in prog whose head is an IDB relation
// (relations with no defining clause -- EDB/.input-only -- are never
// adorned; magic sets only ever restricts IDB predicates). Returns an
// error only if the worklist cap is hit (adornment blowup, blueprint
// failure mode #3) -- everything else about a well-typed, allowed,
// stratifiable input program is expected to succeed.
func Adorn(prog *ast.Program, query *QueryInfo) (*AdornResult, error) {
	rulesByHead := map[string][]*ast.Clause{}
	idb := map[string]bool{}
	for _, c := range prog.Clauses {
		// Matches sema/stratify.go's buildPrecedenceGraph exactly: a
		// relation is IDB (a node in the dependency graph, a target
		// magic sets can restrict) iff it is the head of at least one
		// clause -- fact or rule. EDB/.input-only relations have no
		// clause at all and are never adorned; they are read directly.
		idb[c.Head.Name] = true
		rulesByHead[c.Head.Name] = append(rulesByHead[c.Head.Name], c)
	}

	result := &AdornResult{
		Query:     query,
		Rules:     map[adornedKey][]*AdornedRule{},
		Untouched: map[string]bool{},
	}
	processed := map[adornedKey]bool{}
	var worklist []struct {
		pred   string
		adorn  Adornment
	}
	push := func(pred string, a Adornment) {
		k := keyOf(pred, a)
		if processed[k] {
			return
		}
		processed[k] = true
		worklist = append(worklist, struct {
			pred  string
			adorn Adornment
		}{pred, a})
	}
	push(query.Key.pred, adornFromKey(query.Key))

	for len(worklist) > 0 {
		if result.Iterations >= worklistCap {
			return nil, fmt.Errorf("magicset: worklist exceeded cap of %d adorned (predicate,adornment) pairs -- "+
				"adornment blowup (blueprint failure mode #3), escalate rather than raise the cap silently", worklistCap)
		}
		item := worklist[0]
		worklist = worklist[1:]
		result.Iterations++
		key := keyOf(item.pred, item.adorn)
		result.Order = append(result.Order, key)

		defs := rulesByHead[item.pred]
		// Sort defining clauses by source position for determinism (map
		// iteration is never the source of clause order -- rulesByHead's
		// slice already preserves append/source order since prog.Clauses
		// is walked in order above, so no sort is actually needed here;
		// kept as an explicit no-op comment rather than a silent
		// assumption, since a future refactor of rulesByHead's
		// construction could break this invariant silently otherwise).
		for _, clause := range defs {
			ar := adornOneRule(clause, key, item.adorn, idb, push)
			result.Rules[key] = append(result.Rules[key], ar)
		}
	}

	for pred := range idb {
		if !predicateAdorned(result, pred) {
			result.Untouched[pred] = true
		}
	}
	return result, nil
}

func predicateAdorned(r *AdornResult, pred string) bool {
	for k := range r.Rules {
		if k.pred == pred {
			return true
		}
	}
	return false
}

func adornFromKey(k adornedKey) Adornment {
	a := make(Adornment, k.arity)
	for i, c := range k.adorn {
		a[i] = c == 'b'
	}
	return a
}

// adornOneRule walks clause's SIPS-ordered body once, maintaining the
// bound-variable set G (M2-M3-BUILD.md §2 step 2b-2d), recording every
// IDB atom occurrence's computed adornment (both positive and negated --
// §5) and pushing newly-discovered (predicate,adornment) pairs via push.
func adornOneRule(clause *ast.Clause, key adornedKey, adorn Adornment, idb map[string]bool, push func(string, Adornment)) *AdornedRule {
	initBound := map[string]bool{}
	var boundOrder []string
	for i, b := range adorn {
		if b {
			if v, ok := clause.Head.Terms[i].(*ast.Var); ok && !initBound[v.Name] {
				initBound[v.Name] = true
				boundOrder = append(boundOrder, v.Name)
			}
		}
	}

	ordered := OrderBody(clause.Body, initBound)
	ar := &AdornedRule{Key: key, Source: clause, OrderedBody: ordered}
	ar.BoundAfter = append(ar.BoundAfter, append([]string{}, boundOrder...))

	bound := map[string]bool{}
	for k := range initBound {
		bound[k] = true
	}
	for i, lit := range ordered {
		switch v := lit.(type) {
		case *ast.Atom:
			// EDB atoms (no clause defines them) are read directly, never
			// adorned/pushed -- magic sets only ever restricts IDB
			// predicates (M2-M3-BUILD.md §2's own framing: "For each IDB
			// atom q(s̄) encountered").
			if idb[v.Name] {
				beta := adornmentOf(v, bound)
				target := keyOf(v.Name, beta)
				ar.Occurrences = append(ar.Occurrences, occurrence{literalIndex: i, negated: false, atom: v, target: target})
				push(v.Name, beta)
			}
		case *ast.NegatedAtom:
			if idb[v.Atom.Name] {
				beta := adornmentOf(v.Atom, bound)
				target := keyOf(v.Atom.Name, beta)
				ar.Occurrences = append(ar.Occurrences, occurrence{literalIndex: i, negated: true, atom: v.Atom, target: target})
				push(v.Atom.Name, beta)
			}
		}
		addBoundVars(lit, bound)
		newVars := varsInLit(lit)
		for _, nv := range newVars {
			if bound[nv] {
				already := false
				for _, b := range boundOrder {
					if b == nv {
						already = true
						break
					}
				}
				if !already {
					boundOrder = append(boundOrder, nv)
				}
			}
		}
		ar.BoundAfter = append(ar.BoundAfter, append([]string{}, boundOrder...))
	}
	return ar
}

// adornmentOf computes beta for an IDB atom occurrence: beta[i] = b iff
// the atom's i'th argument is a variable already in bound, or a constant
// (a constant is always "bound" from the callee's point of view -- it
// never needs to be demanded, it is already known). Matches M2-M3-BUILD.md
// §2 step 2c exactly.
func adornmentOf(a *ast.Atom, bound map[string]bool) Adornment {
	beta := make(Adornment, len(a.Terms))
	for i, t := range a.Terms {
		switch v := t.(type) {
		case *ast.Var:
			beta[i] = bound[v.Name]
		case *ast.NumberLit, *ast.StringLit:
			beta[i] = true
		case *ast.Wildcard:
			beta[i] = false
		default:
			// A compound arith expression as a term: bound iff every
			// variable inside it already is (same rule as a constraint's
			// grounded-side check).
			if arith, ok := t.(ast.Arith); ok {
				beta[i] = arithVarsBound(arith, bound)
			}
		}
	}
	return beta
}

func varsInLit(lit ast.Literal) []string {
	var out []string
	seen := map[string]bool{}
	add := func(name string) {
		if !seen[name] {
			seen[name] = true
			out = append(out, name)
		}
	}
	var walk func(ast.Term)
	walk = func(t ast.Term) {
		switch v := t.(type) {
		case *ast.Var:
			add(v.Name)
		case *ast.BinaryExpr:
			walk(v.Left)
			walk(v.Right)
		case *ast.UnaryExpr:
			walk(v.X)
		}
	}
	switch v := lit.(type) {
	case *ast.Atom:
		for _, t := range v.Terms {
			walk(t)
		}
	case *ast.NegatedAtom:
		for _, t := range v.Atom.Terms {
			walk(t)
		}
	case *ast.Constraint:
		walk(v.Left)
		walk(v.Right)
	}
	sort.Strings(out) // deterministic even though insertion order above is already deterministic -- belt and suspenders for a value only used to test set membership by BoundAfter's caller
	return out
}
