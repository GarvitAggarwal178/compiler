// Package lexer tokenizes dlc source text per blueprint §4's grammar,
// amended per docs/M1-BUILD.md §3.3 (zero-arity atoms parse -- a lexer
// concern only insofar as it does not special-case arity anywhere; arity
// is entirely the parser's problem). Lane B.
//
// The lexer never panics. A malformed input byte, an unterminated string,
// or an unterminated block comment produces an ERROR token with a span
// and the lexer resumes scanning immediately after it -- one bad token
// never stops the rest of the file from being tokenized. See DESIGN.md
// for the '.'-disambiguation rule and the two deliberate points of
// disagreement with Soufflé's own lexer.
package lexer

import (
	"unicode/utf8"

	"dlc/src/token"
)

var directiveKinds = map[string]token.Kind{
	"decl":   token.DECL,
	"input":  token.INPUT,
	"output": token.OUTPUT,
}

// Lexer tokenizes one source file's bytes.
type Lexer struct {
	src    []byte
	offset int
	line   int
	col    int
}

// New returns a Lexer positioned at the start of src.
func New(src []byte) *Lexer {
	return &Lexer{src: src, offset: 0, line: 1, col: 1}
}

// Tokenize scans src to completion and returns every token, ending with
// exactly one EOF token. Never panics; lexical errors are ERROR tokens
// in the returned slice, not a separate error value.
func Tokenize(src []byte) []token.Token {
	l := New(src)
	var out []token.Token
	for {
		t := l.Next()
		out = append(out, t)
		if t.Kind == token.EOF {
			return out
		}
	}
}

func (l *Lexer) pos() token.Position {
	return token.Position{Offset: l.offset, Line: l.line, Col: l.col}
}

// peekByte returns the byte at src[offset+n] and true, or 0 and false if
// that is past the end of src.
func (l *Lexer) peekByte(n int) (byte, bool) {
	if l.offset+n >= len(l.src) {
		return 0, false
	}
	return l.src[l.offset+n], true
}

// advance consumes exactly one byte, tracking line/col. '\n' starts a new
// line; '\r' is treated as ordinary whitespace that does not itself
// advance the line (so a "\r\n" pair advances the line once, on the
// '\n' -- documented in DESIGN.md, a deliberate choice, not an oversight).
func (l *Lexer) advance() byte {
	b := l.src[l.offset]
	l.offset++
	if b == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return b
}

func isASCIIDigit(b byte) bool { return b >= '0' && b <= '9' }
func isASCIILetter(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}
func isIdentStart(b byte) bool { return isASCIILetter(b) || b == '_' }
func isIdentCont(b byte) bool  { return isIdentStart(b) || isASCIIDigit(b) }

// Next returns the next token, skipping whitespace and comments first.
// Returns an EOF token (repeatedly, harmlessly) once the input is
// exhausted.
func (l *Lexer) Next() token.Token {
	for {
		skipped := l.skipWhitespace()
		commented, errTok, hadErr := l.skipComment()
		if hadErr {
			return errTok
		}
		if !skipped && !commented {
			break
		}
	}

	start := l.pos()
	if l.offset >= len(l.src) {
		return token.Token{Kind: token.EOF, Span: token.Span{Start: start, End: start}}
	}

	b := l.src[l.offset]
	switch {
	case b == '.':
		return l.lexDotOrDirective(start)
	case isIdentStart(b):
		return l.lexIdentOrKeyword(start)
	case isASCIIDigit(b):
		return l.lexNumber(start)
	case b == '"':
		return l.lexString(start)
	case b == ':':
		l.advance()
		if nb, ok := l.peekByte(0); ok && nb == '-' {
			l.advance()
			return l.finish(token.COLONDASH, start)
		}
		return l.finish(token.COLON, start)
	case b == '!':
		l.advance()
		if nb, ok := l.peekByte(0); ok && nb == '=' {
			l.advance()
			return l.finish(token.NEQ, start)
		}
		return l.finish(token.BANG, start)
	case b == '<':
		l.advance()
		if nb, ok := l.peekByte(0); ok && nb == '=' {
			l.advance()
			return l.finish(token.LE, start)
		}
		return l.finish(token.LT, start)
	case b == '>':
		l.advance()
		if nb, ok := l.peekByte(0); ok && nb == '=' {
			l.advance()
			return l.finish(token.GE, start)
		}
		return l.finish(token.GT, start)
	case b == '=':
		l.advance()
		return l.finish(token.EQ, start)
	case b == ',':
		l.advance()
		return l.finish(token.COMMA, start)
	case b == '(':
		l.advance()
		return l.finish(token.LPAREN, start)
	case b == ')':
		l.advance()
		return l.finish(token.RPAREN, start)
	case b == '+':
		l.advance()
		return l.finish(token.PLUS, start)
	case b == '-':
		l.advance()
		return l.finish(token.MINUS, start)
	case b == '*':
		l.advance()
		return l.finish(token.STAR, start)
	case b == '/':
		l.advance()
		return l.finish(token.SLASH, start)
	case b == '%':
		l.advance()
		return l.finish(token.PERCENT, start)
	default:
		return l.lexInvalidRune(start)
	}
}

// finish builds a token of kind with text taken from src[start.Offset:offset).
func (l *Lexer) finish(kind token.Kind, start token.Position) token.Token {
	text := string(l.src[start.Offset:l.offset])
	return token.Token{Kind: kind, Text: text, Span: token.Span{Start: start, End: l.pos()}}
}

func (l *Lexer) errAt(start token.Position, text, message string) token.Token {
	return token.Token{Kind: token.ERROR, Text: text, Message: message, Span: token.Span{Start: start, End: l.pos()}}
}

// skipWhitespace consumes ASCII space/tab/CR/LF and reports whether it
// consumed anything.
func (l *Lexer) skipWhitespace() bool {
	consumed := false
	for {
		b, ok := l.peekByte(0)
		if !ok {
			return consumed
		}
		if b == ' ' || b == '\t' || b == '\r' || b == '\n' {
			l.advance()
			consumed = true
			continue
		}
		return consumed
	}
}

// skipComment consumes one '//...' or '/*...*/' comment if present at the
// current position. Returns (consumedSomething, errorToken, hadError).
// An unterminated '/*' produces an ERROR token instead of Soufflé's own
// silent-swallow-to-EOF behaviour -- a deliberate disagreement, see
// DESIGN.md.
func (l *Lexer) skipComment() (bool, token.Token, bool) {
	b0, ok0 := l.peekByte(0)
	if !ok0 || b0 != '/' {
		return false, token.Token{}, false
	}
	b1, ok1 := l.peekByte(1)
	if !ok1 {
		return false, token.Token{}, false
	}
	start := l.pos()
	if b1 == '/' {
		l.advance()
		l.advance()
		for {
			b, ok := l.peekByte(0)
			if !ok || b == '\n' {
				break
			}
			l.advance()
		}
		return true, token.Token{}, false
	}
	if b1 == '*' {
		l.advance()
		l.advance()
		for {
			b, ok := l.peekByte(0)
			if !ok {
				return false, l.errAt(start, string(l.src[start.Offset:l.offset]),
					"unterminated block comment"), true
			}
			if b == '*' {
				if nb, ok2 := l.peekByte(1); ok2 && nb == '/' {
					l.advance()
					l.advance()
					return true, token.Token{}, false
				}
			}
			l.advance()
		}
	}
	return false, token.Token{}, false
}

func (l *Lexer) lexDotOrDirective(start token.Position) token.Token {
	l.advance() // consume '.'
	nb, ok := l.peekByte(0)
	if !ok || !isASCIILetter(nb) {
		return l.finish(token.DOT, start)
	}
	identStart := l.offset
	for {
		b, ok := l.peekByte(0)
		if !ok || !isIdentCont(b) {
			break
		}
		l.advance()
	}
	name := string(l.src[identStart:l.offset])
	if kind, isDirective := directiveKinds[name]; isDirective {
		return l.finish(kind, start)
	}
	text := string(l.src[start.Offset:l.offset])
	return l.errAt(start, text, "unknown directive '."+name+"'")
}

func (l *Lexer) lexIdentOrKeyword(start token.Position) token.Token {
	for {
		b, ok := l.peekByte(0)
		if !ok || !isIdentCont(b) {
			break
		}
		l.advance()
	}
	text := string(l.src[start.Offset:l.offset])
	if text == "_" {
		return token.Token{Kind: token.UNDERSCORE, Text: text, Span: token.Span{Start: start, End: l.pos()}}
	}
	return l.finish(token.IDENT, start)
}

func (l *Lexer) lexNumber(start token.Position) token.Token {
	for {
		b, ok := l.peekByte(0)
		if !ok || !isASCIIDigit(b) {
			break
		}
		l.advance()
	}
	return l.finish(token.NUMBER, start)
}

// lexString consumes a "..." literal. Supports \" and \\ as escapes.
// Decodes them into Text; an unterminated string (newline or EOF before
// the closing quote) produces an ERROR token.
func (l *Lexer) lexString(start token.Position) token.Token {
	l.advance() // consume opening '"'
	var decoded []byte
	for {
		b, ok := l.peekByte(0)
		if !ok || b == '\n' {
			raw := string(l.src[start.Offset:l.offset])
			return l.errAt(start, raw, "unterminated string literal")
		}
		if b == '"' {
			l.advance()
			return token.Token{Kind: token.STRING, Text: string(decoded), Span: token.Span{Start: start, End: l.pos()}}
		}
		if b == '\\' {
			if nb, ok2 := l.peekByte(1); ok2 && (nb == '"' || nb == '\\') {
				l.advance()
				l.advance()
				decoded = append(decoded, nb)
				continue
			}
		}
		decoded = append(decoded, b)
		l.advance()
	}
}

func (l *Lexer) lexInvalidRune(start token.Position) token.Token {
	r, size := utf8.DecodeRune(l.src[l.offset:])
	if r == utf8.RuneError && size <= 1 {
		size = 1
	}
	for i := 0; i < size; i++ {
		if l.offset < len(l.src) {
			l.advance()
		}
	}
	text := string(l.src[start.Offset:l.offset])
	return l.errAt(start, text, "unexpected character")
}
