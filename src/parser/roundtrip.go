package parser

import "dlc/src/ast"

// RoundtripResult is what §3.3 gate two checks.
type RoundtripResult struct {
	Status      string // "match" | "mismatch" | "parse_error" | "reparse_error"
	ParseErrors []Diagnostic
	Printed     string
}

// Roundtrip parses src, prints the result, reparses the printed text,
// and reports whether the two ASTs are structurally equal (ast.Equal,
// which ignores spans -- see ast/DESIGN.md for why a plain
// reflect.DeepEqual could never pass this check). A parse error on
// either pass is reported directly rather than compared.
func Roundtrip(src []byte) RoundtripResult {
	prog1, errs1 := Parse(src)
	if len(errs1) > 0 {
		return RoundtripResult{Status: "parse_error", ParseErrors: errs1}
	}
	printed := Print(prog1)
	prog2, errs2 := Parse([]byte(printed))
	if len(errs2) > 0 {
		return RoundtripResult{Status: "reparse_error", ParseErrors: errs2, Printed: printed}
	}
	if ast.Equal(prog1, prog2) {
		return RoundtripResult{Status: "match", Printed: printed}
	}
	return RoundtripResult{Status: "mismatch", Printed: printed}
}
