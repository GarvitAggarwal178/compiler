// Command dlc is the compiler's CLI entry point. Lane B.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"dlc/src/lexer"
	"dlc/src/token"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: dlc <subcommand> <file>")
		os.Exit(2)
	}
	subcommand, path := os.Args[1], os.Args[2]

	switch subcommand {
	case "lex":
		runLex(path)
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand %q\n", subcommand)
		os.Exit(2)
	}
}

// jsonSpan and jsonToken mirror token.Span/token.Token in a shape stable
// for harness consumption -- kept separate from the internal types so a
// later internal refactor of token.Token doesn't silently change the
// harness's contract.
type jsonPosition struct {
	Offset int `json:"offset"`
	Line   int `json:"line"`
	Col    int `json:"col"`
}

type jsonSpan struct {
	Start jsonPosition `json:"start"`
	End   jsonPosition `json:"end"`
}

type jsonToken struct {
	Kind    string   `json:"kind"`
	Text    string   `json:"text"`
	Span    jsonSpan `json:"span"`
	Message string   `json:"message,omitempty"`
}

type lexOutput struct {
	Status     string      `json:"status"`
	TokenCount int         `json:"token_count"`
	ErrorCount int         `json:"error_count"`
	Tokens     []jsonToken `json:"tokens"`
	Panic      string      `json:"panic,omitempty"`
}

func toJSONSpan(s token.Span) jsonSpan {
	return jsonSpan{
		Start: jsonPosition{Offset: s.Start.Offset, Line: s.Start.Line, Col: s.Start.Col},
		End:   jsonPosition{Offset: s.End.Offset, Line: s.End.Line, Col: s.End.Col},
	}
}

// runLex tokenizes path and prints a jsonOutput document to stdout. A
// top-level recover is defense in depth only -- the lexer itself is
// specified to never panic (src/lexer/DESIGN.md) -- so that if that
// invariant is ever violated by a future change, the harness's "zero
// panics" gate sees a clean, machine-readable {"panic": "..."} document
// and a nonzero exit code instead of a raw Go stack trace on stderr.
func runLex(path string) {
	defer func() {
		if r := recover(); r != nil {
			out := lexOutput{Status: "panic", Panic: fmt.Sprintf("%v", r)}
			enc, _ := json.Marshal(out)
			fmt.Println(string(enc))
			os.Exit(1)
		}
	}()

	src, err := os.ReadFile(path)
	if err != nil {
		out := lexOutput{Status: "read_error", Panic: err.Error()}
		enc, _ := json.Marshal(out)
		fmt.Println(string(enc))
		os.Exit(1)
	}

	toks := lexer.Tokenize(src)
	out := lexOutput{Status: "lexed", TokenCount: len(toks)}
	for _, t := range toks {
		if t.Kind == token.ERROR {
			out.ErrorCount++
		}
		out.Tokens = append(out.Tokens, jsonToken{
			Kind: t.Kind.String(), Text: t.Text, Span: toJSONSpan(t.Span), Message: t.Message,
		})
	}
	enc, err := json.Marshal(out)
	if err != nil {
		fmt.Fprintf(os.Stderr, "internal error marshaling output: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(enc))
}
