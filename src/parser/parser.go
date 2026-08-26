// Package parser implements dlc's recursive-descent parser (clause
// structure) with precedence climbing (arith expressions), per
// docs/M1-BUILD.md §3.3. Lane B.
package parser

import (
	"strconv"

	"dlc/src/ast"
	"dlc/src/lexer"
	"dlc/src/token"
)

// Diagnostic is one parse error: a span and a human-readable message.
type Diagnostic struct {
	Span    token.Span
	Message string
}

// Parser walks a pre-lexed token stream. It never panics: every parse
// step either produces a node or records a Diagnostic and signals
// failure to its caller; recovery is centralized in parseProgram (see
// DESIGN.md).
type Parser struct {
	toks []token.Token
	pos  int
	errs []Diagnostic
}

// Parse tokenizes and parses src, returning the resulting Program (never
// nil, even on error -- partially built from whatever clauses/decls
// parsed cleanly) and every Diagnostic recorded along the way.
func Parse(src []byte) (*ast.Program, []Diagnostic) {
	p := &Parser{toks: lexer.Tokenize(src)}
	prog := p.parseProgram()
	return prog, p.errs
}

func (p *Parser) cur() token.Token  { return p.toks[p.pos] }
func (p *Parser) atEOF() bool       { return p.cur().Kind == token.EOF }
func (p *Parser) advance() token.Token {
	t := p.toks[p.pos]
	if p.pos < len(p.toks)-1 {
		p.pos++
	}
	return t
}

func (p *Parser) errorAt(span token.Span, msg string) {
	p.errs = append(p.errs, Diagnostic{Span: span, Message: msg})
}

// expect consumes the current token if it has kind k, returning it and
// true; otherwise records a diagnostic and returns the zero Token and
// false, WITHOUT advancing (so the caller's own recovery sees the
// offending token too).
func (p *Parser) expect(k token.Kind, what string) (token.Token, bool) {
	if p.cur().Kind == token.ERROR {
		t := p.cur()
		p.errorAt(t.Span, "lex error: "+t.Message)
		return token.Token{}, false
	}
	if p.cur().Kind != k {
		p.errorAt(p.cur().Span, "expected "+what+", got "+p.cur().Kind.String())
		return token.Token{}, false
	}
	return p.advance(), true
}

// recoverToNextDot advances past tokens until it has consumed a DOT (or
// hits EOF), per §3.3's stated recovery rule: "skip to the next '.' at
// clause level, then continue." Applied uniformly to a failed decl or a
// failed clause -- see DESIGN.md for why decls (which have no '.'
// terminator of their own) still recover this way.
func (p *Parser) recoverToNextDot() {
	for !p.atEOF() {
		if p.cur().Kind == token.DOT {
			p.advance()
			return
		}
		p.advance()
	}
}

// parseProgram parses an interleaved sequence of decls and clauses (see
// DESIGN.md: blueprint §4's `decl* clause*` is read as unordered/
// interleaved, not strict staging -- this project's own already-
// Soufflé-validated .dl programs interleave them throughout).
func (p *Parser) parseProgram() *ast.Program {
	prog := &ast.Program{}
	for !p.atEOF() {
		startPos := p.pos
		switch p.cur().Kind {
		case token.DECL, token.INPUT, token.OUTPUT:
			if d, ok := p.parseDecl(); ok {
				prog.Decls = append(prog.Decls, d)
			} else {
				p.recoverToNextDot()
			}
		case token.IDENT:
			if c, ok := p.parseClause(); ok {
				prog.Clauses = append(prog.Clauses, c)
			} else {
				p.recoverToNextDot()
			}
		case token.ERROR:
			t := p.cur()
			p.errorAt(t.Span, "lex error: "+t.Message)
			p.advance()
			p.recoverToNextDot()
		default:
			p.errorAt(p.cur().Span, "expected a declaration or a clause, got "+p.cur().Kind.String())
			p.recoverToNextDot()
		}
		if p.pos == startPos {
			// Safety valve: every branch above must make forward
			// progress. If one doesn't (a future bug), force one token
			// of progress rather than looping forever -- a hang is
			// worse than a slightly-worse error recovery.
			p.advance()
		}
	}
	return prog
}

func (p *Parser) parseDecl() (*ast.Decl, bool) {
	start := p.cur().Span.Start
	switch p.cur().Kind {
	case token.DECL:
		p.advance()
		nameTok, ok := p.expect(token.IDENT, "a relation name")
		if !ok {
			return nil, false
		}
		if _, ok := p.expect(token.LPAREN, "'('"); !ok {
			return nil, false
		}
		var params []ast.Param
		if p.cur().Kind != token.RPAREN {
			for {
				param, ok := p.parseParam()
				if !ok {
					return nil, false
				}
				params = append(params, param)
				if p.cur().Kind != token.COMMA {
					break
				}
				p.advance()
			}
		}
		rparen, ok := p.expect(token.RPAREN, "')'")
		if !ok {
			return nil, false
		}
		return &ast.Decl{Kind: ast.DeclRelation, Name: nameTok.Text, Params: params,
			Sp: token.Span{Start: start, End: rparen.Span.End}}, true
	case token.INPUT, token.OUTPUT:
		kw := p.advance()
		nameTok, ok := p.expect(token.IDENT, "a relation name")
		if !ok {
			return nil, false
		}
		kind := ast.DeclInput
		if kw.Kind == token.OUTPUT {
			kind = ast.DeclOutput
		}
		return &ast.Decl{Kind: kind, Name: nameTok.Text, Sp: token.Span{Start: start, End: nameTok.Span.End}}, true
	}
	p.errorAt(p.cur().Span, "expected '.decl', '.input', or '.output'")
	return nil, false
}

func (p *Parser) parseParam() (ast.Param, bool) {
	nameTok, ok := p.expect(token.IDENT, "a parameter name")
	if !ok {
		return ast.Param{}, false
	}
	if _, ok := p.expect(token.COLON, "':'"); !ok {
		return ast.Param{}, false
	}
	typeTok, ok := p.expect(token.IDENT, "a type ('number' or 'symbol')")
	if !ok {
		return ast.Param{}, false
	}
	if typeTok.Text != "number" && typeTok.Text != "symbol" {
		p.errorAt(typeTok.Span, "invalid type "+strconv.Quote(typeTok.Text)+", expected 'number' or 'symbol'")
		return ast.Param{}, false
	}
	return ast.Param{Name: nameTok.Text, Type: typeTok.Text, Sp: token.Span{Start: nameTok.Span.Start, End: typeTok.Span.End}}, true
}

func (p *Parser) parseClause() (*ast.Clause, bool) {
	head, ok := p.parseAtom()
	if !ok {
		return nil, false
	}
	switch p.cur().Kind {
	case token.DOT:
		dot := p.advance()
		return &ast.Clause{Head: head, Sp: token.Span{Start: head.Sp.Start, End: dot.Span.End}}, true
	case token.COLONDASH:
		p.advance()
		body, ok := p.parseBody()
		if !ok {
			return nil, false
		}
		dot, ok := p.expect(token.DOT, "'.'")
		if !ok {
			return nil, false
		}
		return &ast.Clause{Head: head, Body: body, Sp: token.Span{Start: head.Sp.Start, End: dot.Span.End}}, true
	}
	p.errorAt(p.cur().Span, "expected '.' or ':-' after atom, got "+p.cur().Kind.String())
	return nil, false
}

func (p *Parser) parseBody() ([]ast.Literal, bool) {
	var lits []ast.Literal
	for {
		lit, ok := p.parseLiteral()
		if !ok {
			return nil, false
		}
		lits = append(lits, lit)
		if p.cur().Kind != token.COMMA {
			return lits, true
		}
		p.advance()
	}
}

// parseLiteral disambiguates atom / negated-atom / constraint by
// lookahead: '!' always starts a negated atom; an IDENT immediately
// followed by '(' starts an atom; anything else is parsed as an arith
// expression and must be followed by a relop to form a constraint. See
// DESIGN.md.
func (p *Parser) parseLiteral() (ast.Literal, bool) {
	if p.cur().Kind == token.BANG {
		bang := p.advance()
		atom, ok := p.parseAtom()
		if !ok {
			return nil, false
		}
		return &ast.NegatedAtom{Atom: atom, Sp: token.Span{Start: bang.Span.Start, End: atom.Sp.End}}, true
	}
	if p.cur().Kind == token.IDENT && p.peekKind(1) == token.LPAREN {
		return p.parseAtom()
	}
	left, ok := p.parseArith()
	if !ok {
		return nil, false
	}
	op, ok := p.parseRelop()
	if !ok {
		return nil, false
	}
	right, ok := p.parseArith()
	if !ok {
		return nil, false
	}
	return &ast.Constraint{Left: left, Op: op, Right: right, Sp: token.Span{Start: left.Span().Start, End: right.Span().End}}, true
}

func (p *Parser) peekKind(n int) token.Kind {
	if p.pos+n >= len(p.toks) {
		return token.EOF
	}
	return p.toks[p.pos+n].Kind
}

var relopKinds = map[token.Kind]string{
	token.EQ: "=", token.NEQ: "!=", token.LT: "<", token.LE: "<=", token.GT: ">", token.GE: ">=",
}

func (p *Parser) parseRelop() (string, bool) {
	if text, ok := relopKinds[p.cur().Kind]; ok {
		p.advance()
		return text, true
	}
	p.errorAt(p.cur().Span, "expected a relational operator (=, !=, <, <=, >, >=), got "+p.cur().Kind.String())
	return "", false
}

func (p *Parser) parseAtom() (*ast.Atom, bool) {
	nameTok, ok := p.expect(token.IDENT, "a relation name")
	if !ok {
		return nil, false
	}
	if _, ok := p.expect(token.LPAREN, "'('"); !ok {
		return nil, false
	}
	var terms []ast.Term
	if p.cur().Kind != token.RPAREN {
		for {
			term, ok := p.parseTerm()
			if !ok {
				return nil, false
			}
			terms = append(terms, term)
			if p.cur().Kind != token.COMMA {
				break
			}
			p.advance()
		}
	}
	rparen, ok := p.expect(token.RPAREN, "')'")
	if !ok {
		return nil, false
	}
	return &ast.Atom{Name: nameTok.Text, Terms: terms, Sp: token.Span{Start: nameTok.Span.Start, End: rparen.Span.End}}, true
}

func (p *Parser) parseTerm() (ast.Term, bool) {
	if p.cur().Kind == token.UNDERSCORE {
		u := p.advance()
		return &ast.Wildcard{Sp: u.Span}, true
	}
	return p.parseArith()
}

// --- arith, precedence climbing, tightest first: unary '-'; '*' '/' '%'; '+' '-' ---

func (p *Parser) parseArith() (ast.Arith, bool) { return p.parseAdditive() }

func (p *Parser) parseAdditive() (ast.Arith, bool) {
	left, ok := p.parseMultiplicative()
	if !ok {
		return nil, false
	}
	for p.cur().Kind == token.PLUS || p.cur().Kind == token.MINUS {
		opTok := p.advance()
		right, ok := p.parseMultiplicative()
		if !ok {
			return nil, false
		}
		left = &ast.BinaryExpr{Op: opTok.Text, Left: left, Right: right, Sp: token.Span{Start: left.Span().Start, End: right.Span().End}}
	}
	return left, true
}

func (p *Parser) parseMultiplicative() (ast.Arith, bool) {
	left, ok := p.parseUnary()
	if !ok {
		return nil, false
	}
	for p.cur().Kind == token.STAR || p.cur().Kind == token.SLASH || p.cur().Kind == token.PERCENT {
		opTok := p.advance()
		right, ok := p.parseUnary()
		if !ok {
			return nil, false
		}
		left = &ast.BinaryExpr{Op: opTok.Text, Left: left, Right: right, Sp: token.Span{Start: left.Span().Start, End: right.Span().End}}
	}
	return left, true
}

func (p *Parser) parseUnary() (ast.Arith, bool) {
	if p.cur().Kind == token.MINUS {
		minus := p.advance()
		x, ok := p.parseUnary()
		if !ok {
			return nil, false
		}
		return &ast.UnaryExpr{Op: "-", X: x, Sp: token.Span{Start: minus.Span.Start, End: x.Span().End}}, true
	}
	return p.parsePrimary()
}

func (p *Parser) parsePrimary() (ast.Arith, bool) {
	switch p.cur().Kind {
	case token.IDENT:
		t := p.advance()
		return &ast.Var{Name: t.Text, Sp: t.Span}, true
	case token.NUMBER:
		t := p.advance()
		v, err := strconv.ParseInt(t.Text, 10, 64)
		if err != nil {
			p.errorAt(t.Span, "invalid number literal "+strconv.Quote(t.Text))
			return nil, false
		}
		return &ast.NumberLit{Value: v, Text: t.Text, Sp: t.Span}, true
	case token.STRING:
		t := p.advance()
		return &ast.StringLit{Value: t.Text, Sp: t.Span}, true
	case token.LPAREN:
		lparen := p.advance()
		inner, ok := p.parseArith()
		if !ok {
			return nil, false
		}
		rparen, ok := p.expect(token.RPAREN, "')'")
		if !ok {
			return nil, false
		}
		// The paren'd expression keeps its own span extended to cover
		// the parens, so a diagnostic on it points at the whole
		// "(...)" rather than just the inner expression.
		switch v := inner.(type) {
		case *ast.BinaryExpr:
			v.Sp = token.Span{Start: lparen.Span.Start, End: rparen.Span.End}
		case *ast.UnaryExpr:
			v.Sp = token.Span{Start: lparen.Span.Start, End: rparen.Span.End}
		}
		return inner, true
	}
	p.errorAt(p.cur().Span, "expected a variable, number, string, or '(', got "+p.cur().Kind.String())
	return nil, false
}
