package ast

// Equal reports whether two Programs are structurally identical --
// same declarations, same clauses, same terms and operators -- ignoring
// every Span. This is what §3.3 gate two's round-trip check needs:
// parse -> print -> reparse necessarily produces a *different* Span on
// every node (the printed text is not byte-identical to the original
// source), so a reflect.DeepEqual over the raw structs would always fail
// even on a correct printer. A hand-written comparison, one case per
// concrete type, is more code than reflect.DeepEqual but is the only way
// to skip exactly the Span fields and nothing else.
func Equal(a, b *Program) bool {
	if a == nil || b == nil {
		return a == b
	}
	if len(a.Decls) != len(b.Decls) || len(a.Clauses) != len(b.Clauses) {
		return false
	}
	for i := range a.Decls {
		if !declEqual(a.Decls[i], b.Decls[i]) {
			return false
		}
	}
	for i := range a.Clauses {
		if !clauseEqual(a.Clauses[i], b.Clauses[i]) {
			return false
		}
	}
	return true
}

func declEqual(a, b *Decl) bool {
	if a == nil || b == nil {
		return a == b
	}
	if a.Kind != b.Kind || a.Name != b.Name || len(a.Params) != len(b.Params) {
		return false
	}
	for i := range a.Params {
		if a.Params[i].Name != b.Params[i].Name || a.Params[i].Type != b.Params[i].Type {
			return false
		}
	}
	return true
}

func clauseEqual(a, b *Clause) bool {
	if a == nil || b == nil {
		return a == b
	}
	if !atomEqual(a.Head, b.Head) || len(a.Body) != len(b.Body) {
		return false
	}
	for i := range a.Body {
		if !literalEqual(a.Body[i], b.Body[i]) {
			return false
		}
	}
	return true
}

func atomEqual(a, b *Atom) bool {
	if a == nil || b == nil {
		return a == b
	}
	if a.Name != b.Name || len(a.Terms) != len(b.Terms) {
		return false
	}
	for i := range a.Terms {
		if !termEqual(a.Terms[i], b.Terms[i]) {
			return false
		}
	}
	return true
}

func literalEqual(a, b Literal) bool {
	switch av := a.(type) {
	case *Atom:
		bv, ok := b.(*Atom)
		return ok && atomEqual(av, bv)
	case *NegatedAtom:
		bv, ok := b.(*NegatedAtom)
		return ok && atomEqual(av.Atom, bv.Atom)
	case *Constraint:
		bv, ok := b.(*Constraint)
		return ok && av.Op == bv.Op && termEqual(av.Left, bv.Left) && termEqual(av.Right, bv.Right)
	default:
		return false
	}
}

func termEqual(a, b Term) bool {
	switch av := a.(type) {
	case *Wildcard:
		_, ok := b.(*Wildcard)
		return ok
	case *BinaryExpr:
		bv, ok := b.(*BinaryExpr)
		return ok && av.Op == bv.Op && termEqual(av.Left, bv.Left) && termEqual(av.Right, bv.Right)
	case *UnaryExpr:
		bv, ok := b.(*UnaryExpr)
		return ok && av.Op == bv.Op && termEqual(av.X, bv.X)
	case *Var:
		bv, ok := b.(*Var)
		return ok && av.Name == bv.Name
	case *NumberLit:
		bv, ok := b.(*NumberLit)
		return ok && av.Value == bv.Value
	case *StringLit:
		bv, ok := b.(*StringLit)
		return ok && av.Value == bv.Value
	default:
		return false
	}
}
