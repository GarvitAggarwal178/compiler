package lexer

import (
	"testing"

	"dlc/src/token"
)

func kinds(toks []token.Token) []token.Kind {
	ks := make([]token.Kind, len(toks))
	for i, t := range toks {
		ks[i] = t.Kind
	}
	return ks
}

func assertKinds(t *testing.T, src string, want ...token.Kind) {
	t.Helper()
	got := kinds(Tokenize([]byte(src)))
	if len(got) != len(want) {
		t.Fatalf("Tokenize(%q) = %v, want %v", src, got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("Tokenize(%q)[%d] = %v, want %v (full: %v)", src, i, got[i], want[i], got)
		}
	}
}

func TestDotVsDirective(t *testing.T) {
	assertKinds(t, ".decl", token.DECL, token.EOF)
	assertKinds(t, ".input", token.INPUT, token.EOF)
	assertKinds(t, ".output", token.OUTPUT, token.EOF)
	assertKinds(t, "p(X).", token.IDENT, token.LPAREN, token.IDENT, token.RPAREN, token.DOT, token.EOF)
	assertKinds(t, ".foo", token.ERROR, token.EOF)
}

func TestMultiCharOperators(t *testing.T) {
	assertKinds(t, ":-", token.COLONDASH, token.EOF)
	assertKinds(t, ":", token.COLON, token.EOF)
	assertKinds(t, "!=", token.NEQ, token.EOF)
	assertKinds(t, "!", token.BANG, token.EOF)
	assertKinds(t, "<=", token.LE, token.EOF)
	assertKinds(t, "<", token.LT, token.EOF)
	assertKinds(t, ">=", token.GE, token.EOF)
	assertKinds(t, ">", token.GT, token.EOF)
}

func TestWildcardVsIdent(t *testing.T) {
	assertKinds(t, "_", token.UNDERSCORE, token.EOF)
	assertKinds(t, "_foo", token.IDENT, token.EOF)
	assertKinds(t, "_1", token.IDENT, token.EOF)
}

func TestErrorRecoveryContinuesToEOF(t *testing.T) {
	// One malformed clause must not kill the rest of the file: after an
	// ERROR token, lexing resumes normally.
	toks := Tokenize([]byte("p(X) :- @ q(X)."))
	got := kinds(toks)
	want := []token.Kind{
		token.IDENT, token.LPAREN, token.IDENT, token.RPAREN, token.COLONDASH,
		token.ERROR, token.IDENT, token.LPAREN, token.IDENT, token.RPAREN, token.DOT, token.EOF,
	}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("token %d: got %v, want %v (full %v)", i, got[i], want[i], got)
		}
	}
}

func TestUnterminatedStringIsErrorNotPanic(t *testing.T) {
	toks := Tokenize([]byte(`p("unterminated`))
	if toks[len(toks)-2].Kind != token.ERROR {
		t.Fatalf("expected an ERROR token before EOF, got %v", kinds(toks))
	}
	if toks[len(toks)-1].Kind != token.EOF {
		t.Fatalf("expected trailing EOF, got %v", kinds(toks))
	}
}

func TestUnterminatedBlockCommentIsErrorNotPanic(t *testing.T) {
	toks := Tokenize([]byte("p(X) :- q(X). /* never closed"))
	found := false
	for _, tk := range toks {
		if tk.Kind == token.ERROR {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected an ERROR token for the unterminated block comment, got %v", kinds(toks))
	}
	if toks[len(toks)-1].Kind != token.EOF {
		t.Fatalf("expected trailing EOF, got %v", kinds(toks))
	}
}

func TestNonASCIIByteIsErrorNotPanic(t *testing.T) {
	toks := Tokenize([]byte("données"))
	if toks[0].Kind != token.IDENT || toks[0].Text != "donn" {
		t.Fatalf("expected IDENT %q, got %v", "donn", toks[0])
	}
	if toks[1].Kind != token.ERROR {
		t.Fatalf("expected ERROR for the non-ASCII rune, got %v", kinds(toks))
	}
}

func TestSpansAreConsistent(t *testing.T) {
	toks := Tokenize([]byte("p(X)"))
	// p
	if toks[0].Span.Start != (token.Position{Offset: 0, Line: 1, Col: 1}) {
		t.Fatalf("bad start span for 'p': %v", toks[0].Span)
	}
	if toks[0].Span.End != (token.Position{Offset: 1, Line: 1, Col: 2}) {
		t.Fatalf("bad end span for 'p': %v", toks[0].Span)
	}
}

func TestEmptyFileYieldsJustEOF(t *testing.T) {
	assertKinds(t, "", token.EOF)
}

func TestOnlyCommentsYieldsJustEOF(t *testing.T) {
	assertKinds(t, "// hello\n/* world */", token.EOF)
}

func TestNoPanicOnFullHostileAndGrammarCorpusIsCheckedByHarness(t *testing.T) {
	// The exhaustive zero-panic check over all 195 in-grammar files and
	// all 39 hostile files is the actual §3.1 acceptance gate and lives
	// in harness/lex_coverage.py (Python, drives the built dlc binary) --
	// not duplicated here. This test just documents that fact so a
	// reader of this file knows where the real gate is.
}
