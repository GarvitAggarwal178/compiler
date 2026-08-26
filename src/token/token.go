// Package token defines the lexical token vocabulary shared by the lexer,
// parser, and every diagnostic downstream of them. Lane B.
package token

import "fmt"

// Kind identifies what a Token represents.
type Kind int

const (
	EOF Kind = iota
	ERROR

	IDENT
	NUMBER
	STRING

	COLONDASH  // ':-'
	DOT        // '.'
	COMMA      // ','
	BANG       // '!'
	LPAREN     // '('
	RPAREN     // ')'
	COLON      // ':'
	UNDERSCORE // '_' (wildcard)

	EQ  // '='
	NEQ // '!='
	LT  // '<'
	LE  // '<='
	GT  // '>'
	GE  // '>='

	PLUS    // '+'
	MINUS   // '-'
	STAR    // '*'
	SLASH   // '/'
	PERCENT // '%'

	DECL   // '.decl'
	INPUT  // '.input'
	OUTPUT // '.output'
)

var kindNames = map[Kind]string{
	EOF:        "EOF",
	ERROR:      "ERROR",
	IDENT:      "IDENT",
	NUMBER:     "NUMBER",
	STRING:     "STRING",
	COLONDASH:  ":-",
	DOT:        ".",
	COMMA:      ",",
	BANG:       "!",
	LPAREN:     "(",
	RPAREN:     ")",
	COLON:      ":",
	UNDERSCORE: "_",
	EQ:         "=",
	NEQ:        "!=",
	LT:         "<",
	LE:         "<=",
	GT:         ">",
	GE:         ">=",
	PLUS:       "+",
	MINUS:      "-",
	STAR:       "*",
	SLASH:      "/",
	PERCENT:    "%",
	DECL:       ".decl",
	INPUT:      ".input",
	OUTPUT:     ".output",
}

func (k Kind) String() string {
	if s, ok := kindNames[k]; ok {
		return s
	}
	return fmt.Sprintf("Kind(%d)", int(k))
}

// Position is a single point in a source file. Offset is a 0-based byte
// offset; Line and Col are 1-based. Col counts bytes, not runes -- the
// lexer rejects non-ASCII bytes outright (see lexer/DESIGN.md), so byte
// and rune columns coincide for every position a Token can legally start
// or end at.
type Position struct {
	Offset int
	Line   int
	Col    int
}

func (p Position) String() string {
	return fmt.Sprintf("%d:%d", p.Line, p.Col)
}

// Span is a half-open [Start, End) source range.
type Span struct {
	Start Position
	End   Position
}

func (s Span) String() string {
	return fmt.Sprintf("%s-%s", s.Start, s.End)
}

// Token is one lexical token. Every token, including ERROR ones, carries
// a Span -- there is no diagnostic downstream of the lexer that can
// afford not to have one.
//
// Text holds the token's literal source text for IDENT/NUMBER, and the
// *decoded* contents (escapes resolved) for STRING. Message is set only
// on ERROR tokens and explains what went wrong; Text on an ERROR token
// holds the raw offending source text.
type Token struct {
	Kind    Kind
	Text    string
	Span    Span
	Message string
}

func (t Token) String() string {
	if t.Kind == ERROR {
		return fmt.Sprintf("%s(%q, %s) at %s", t.Kind, t.Text, t.Message, t.Span)
	}
	return fmt.Sprintf("%s(%q) at %s", t.Kind, t.Text, t.Span)
}
