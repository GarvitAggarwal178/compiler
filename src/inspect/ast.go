package inspect

import (
	"dlc/src/ast"
	"dlc/src/token"
	"fmt"
	"strconv"
)

type Position struct {
	Offset int `json:"offset"`
	Line   int `json:"line"`
	Col    int `json:"col"`
}
type Span struct {
	Start Position `json:"start"`
	End   Position `json:"end"`
}

func span(s token.Span) *Span {
	if s.Start.Line == 0 {
		return nil
	}
	return &Span{Start: Position{s.Start.Offset, s.Start.Line, s.Start.Col}, End: Position{s.End.Offset, s.End.Line, s.End.Col}}
}

type Param struct {
	Name string `json:"name"`
	Type string `json:"type"`
	Span *Span  `json:"span,omitempty"`
}

// Node is a tagged view of actual compiler AST nodes. Generated nodes omit
// source spans; their source links come from generation metadata instead.
type Node struct {
	Kind     string  `json:"kind"`
	ID       string  `json:"id"`
	Span     *Span   `json:"span,omitempty"`
	Name     string  `json:"name,omitempty"`
	DeclKind string  `json:"decl_kind,omitempty"`
	Params   []Param `json:"params,omitempty"`
	Decls    []*Node `json:"decls,omitempty"`
	Clauses  []*Node `json:"clauses,omitempty"`
	Head     *Node   `json:"head,omitempty"`
	Body     []*Node `json:"body,omitempty"`
	Atom     *Node   `json:"atom,omitempty"`
	Terms    []*Node `json:"terms,omitempty"`
	Left     *Node   `json:"left,omitempty"`
	Right    *Node   `json:"right,omitempty"`
	X        *Node   `json:"x,omitempty"`
	Op       string  `json:"op,omitempty"`
	Value    *string `json:"value,omitempty"`
	Text     string  `json:"text,omitempty"`
}

func term(t ast.Term, id string, source bool) *Node {
	n := &Node{ID: id}
	if source {
		n.Span = span(t.Span())
	}
	switch v := t.(type) {
	case *ast.Var:
		n.Kind = "Var"
		n.Name = v.Name
	case *ast.Wildcard:
		n.Kind = "Wildcard"
	case *ast.NumberLit:
		n.Kind = "NumberLit"
		s := strconv.FormatInt(v.Value, 10)
		n.Value = &s
		n.Text = v.Text
	case *ast.StringLit:
		n.Kind = "StringLit"
		s := v.Value
		n.Value = &s
	case *ast.UnaryExpr:
		n.Kind = "UnaryExpr"
		n.Op = v.Op
		n.X = term(v.X, id+"/x", source)
	case *ast.BinaryExpr:
		n.Kind = "BinaryExpr"
		n.Op = v.Op
		n.Left = term(v.Left, id+"/left", source)
		n.Right = term(v.Right, id+"/right", source)
	}
	return n
}
func atom(a *ast.Atom, id string, source bool) *Node {
	n := &Node{ID: id, Kind: "Atom", Name: a.Name, Terms: []*Node{}}
	if source {
		n.Span = span(a.Sp)
	}
	for i, t := range a.Terms {
		n.Terms = append(n.Terms, term(t, fmt.Sprintf("%s/term_%d", id, i), source))
	}
	return n
}
func literal(l ast.Literal, id string, source bool) *Node {
	switch v := l.(type) {
	case *ast.Atom:
		return atom(v, id, source)
	case *ast.NegatedAtom:
		n := &Node{ID: id, Kind: "NegatedAtom", Atom: atom(v.Atom, id+"/atom", source)}
		if source {
			n.Span = span(v.Sp)
		}
		return n
	case *ast.Constraint:
		n := &Node{ID: id, Kind: "Constraint", Op: v.Op, Left: term(v.Left, id+"/left", source), Right: term(v.Right, id+"/right", source)}
		if source {
			n.Span = span(v.Sp)
		}
		return n
	}
	panic("inspect: unsupported literal")
}
func program(p *ast.Program, source bool) *Node {
	n := &Node{Kind: "Program", ID: "program", Decls: []*Node{}, Clauses: []*Node{}}
	if source {
		n.Span = span(p.Sp)
	}
	for i, d := range p.Decls {
		k := []string{"relation", "input", "output"}[d.Kind]
		dn := &Node{Kind: "Decl", ID: fmt.Sprintf("decl_%d", i), Name: d.Name, DeclKind: k, Params: []Param{}}
		if source {
			dn.Span = span(d.Sp)
		}
		for _, p := range d.Params {
			sp := (*Span)(nil)
			if source {
				sp = span(p.Sp)
			}
			dn.Params = append(dn.Params, Param{Name: p.Name, Type: p.Type, Span: sp})
		}
		n.Decls = append(n.Decls, dn)
	}
	for i, c := range p.Clauses {
		id := fmt.Sprintf("rule_%d", i)
		cn := &Node{Kind: "Clause", ID: id, Head: atom(c.Head, id+"/head", source), Body: []*Node{}}
		if source {
			cn.Span = span(c.Sp)
		}
		for j, l := range c.Body {
			cn.Body = append(cn.Body, literal(l, fmt.Sprintf("%s/body_%d", id, j), source))
		}
		n.Clauses = append(n.Clauses, cn)
	}
	return n
}
