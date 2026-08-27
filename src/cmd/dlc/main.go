// Command dlc is the compiler's CLI entry point. Lane B.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"strings"

	"dlc/src/ast"
	"dlc/src/codegen"
	"dlc/src/eval"
	"dlc/src/ir"
	"dlc/src/lexer"
	"dlc/src/parser"
	"dlc/src/sema"
	"dlc/src/token"
	"dlc/src/transform"
	"dlc/src/transform/guard"
	"dlc/src/transform/magicset"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: dlc <subcommand> <file> [args...]")
		os.Exit(2)
	}
	subcommand, path := os.Args[1], os.Args[2]

	switch subcommand {
	case "lex":
		runLex(path)
	case "parse":
		runParse(path)
	case "roundtrip":
		runRoundtrip(path)
	case "check":
		runCheck(path)
	case "run":
		if len(os.Args) < 5 || len(os.Args) > 6 {
			fmt.Fprintln(os.Stderr, "usage: dlc run <file> <factsDir> <outDir> [--transformer=name]")
			os.Exit(2)
		}
		runRun(path, os.Args[3], os.Args[4], false, transformerFlag(os.Args[5:]))
	case "run-seminaive":
		if len(os.Args) < 5 || len(os.Args) > 6 {
			fmt.Fprintln(os.Stderr, "usage: dlc run-seminaive <file> <factsDir> <outDir> [--transformer=name]")
			os.Exit(2)
		}
		runRun(path, os.Args[3], os.Args[4], true, transformerFlag(os.Args[5:]))
	case "codegen":
		if len(os.Args) != 4 {
			fmt.Fprintln(os.Stderr, "usage: dlc codegen <file> <outfile.c>")
			os.Exit(2)
		}
		runCodegen(path, os.Args[3])
	case "emit":
		runEmit(path, transformerFlag(os.Args[3:]))
	case "explain":
		runExplain(path)
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

type runOutput struct {
	Status             string               `json:"status"` // "ok" | "rejected" | "parse_error" | "eval_error"
	Diagnostics        []jsonSemaDiagnostic `json:"diagnostics,omitempty"`
	ParseErrors        []jsonDiagnostic     `json:"parse_errors,omitempty"`
	DerivationAttempts int64                `json:"derivation_attempts,omitempty"`
	Panic              string               `json:"panic,omitempty"`
}

// runRun is §3.8/§3.9's entry point and harness/differential.py's
// run_dlc() target: parse -> full sema check -> (if clean) load facts,
// evaluate, write every .output relation's .csv to outDir in the same
// tab-separated, no-header shape Soufflé itself writes, plus a
// Soufflé-profile-shaped JSON (ir.EmitProfile) to <outDir>/profile.json
// for T_naive/T_semi-naive extraction. Prints a runOutput document so a
// caller can tell "ok, csvs are on disk" from "rejected" from "crashed"
// without inspecting the filesystem first.
//
// semiNaive selects §3.9's evaluator instead of §3.8's naive one --
// same parse/check/load/write pipeline either way, so the two are only
// ever compared on identical everything else (differential.py's own
// gate-one re-check, and the T_naive-vs-T_semi-naive headline number
// both depend on that).
// transformerFlag scans args (the tail of os.Args after the fixed
// positional ones) for "--transformer=name", defaulting to "passthrough"
// if absent -- shared by run/run-seminaive/emit so all three subcommands
// accept the flag identically.
func transformerFlag(args []string) string {
	name := "passthrough"
	for _, a := range args {
		if v, ok := strings.CutPrefix(a, "--transformer="); ok {
			name = v
		}
	}
	return name
}

func runRun(path, factsDir, outDir string, semiNaive bool, transformerName string) {
	defer func() {
		if r := recover(); r != nil {
			out := runOutput{Status: "panic", Panic: fmt.Sprintf("%v", r)}
			enc, _ := json.Marshal(out)
			fmt.Println(string(enc))
			os.Exit(1)
		}
	}()

	src, err := os.ReadFile(path)
	if err != nil {
		out := runOutput{Status: "read_error", Panic: err.Error()}
		enc, _ := json.Marshal(out)
		fmt.Println(string(enc))
		os.Exit(1)
	}

	prog, perrs := parser.Parse(src)
	if len(perrs) > 0 {
		out := runOutput{Status: "parse_error"}
		for _, e := range perrs {
			out.ParseErrors = append(out.ParseErrors, jsonDiagnostic{Span: toJSONSpan(e.Span), Message: e.Message})
		}
		enc, _ := json.Marshal(out)
		fmt.Println(string(enc))
		return
	}

	var diags []sema.Diagnostic
	diags = append(diags, sema.CheckDeclType(prog)...)
	diags = append(diags, sema.CheckAllowedness(prog)...)
	stratDiags, stratResult := sema.CheckStratification(prog)
	diags = append(diags, stratDiags...)
	if len(diags) > 0 {
		out := runOutput{Status: "rejected"}
		for _, d := range diags {
			out.Diagnostics = append(out.Diagnostics, jsonSemaDiagnostic{
				Span: toJSONSpan(d.Span), Category: string(d.Category), Message: d.Message,
			})
		}
		enc, _ := json.Marshal(out)
		fmt.Println(string(enc))
		return
	}

	// Apply the named Transformer, then re-run sema.CheckStratification on
	// its OUTPUT rather than reusing stratResult -- a real transform
	// changes the precedence graph (new magic-seed relations), so the
	// pre-transform StratumResult is invalid for the transformed program.
	// This is exactly the contract transform.Transformer's own doc
	// comment documents and defers to the caller (src/transform/
	// DESIGN.md); this is that caller, wired in now that a real
	// Transformer exists to justify it (transformer.go's own DESIGN.md
	// note: "the natural place to wire it in is exactly when Lane A's
	// real Transformer lands").
	transformer, ok := transformerRegistry[transformerName]
	if !ok {
		out := runOutput{Status: "eval_error", Panic: fmt.Sprintf("unknown transformer %q", transformerName)}
		enc, _ := json.Marshal(out)
		fmt.Println(string(enc))
		os.Exit(2)
	}
	transformed, terr := transformer.Transform(prog, stratResult)
	if terr != nil {
		out := runOutput{Status: "eval_error", Panic: terr.Error()}
		enc, _ := json.Marshal(out)
		fmt.Println(string(enc))
		os.Exit(1)
	}
	transformedSchemas, _ := sema.BuildSymbolTable(transformed)
	_, transformedStrat := sema.CheckStratification(transformed)
	if transformedStrat == nil {
		out := runOutput{Status: "eval_error", Panic: "transformer produced an unstratifiable program -- this should be impossible for --transformer=guarded, which is required to only ever return a stratifiable result"}
		enc, _ := json.Marshal(out)
		fmt.Println(string(enc))
		os.Exit(1)
	}

	inputNames := map[string]bool{}
	outputNames := map[string]bool{}
	for _, d := range transformed.Decls {
		switch d.Kind {
		case ast.DeclInput:
			inputNames[d.Name] = true
		case ast.DeclOutput:
			outputNames[d.Name] = true
		}
	}

	evaluator := eval.NewEvaluator(transformedSchemas.Relations)
	if err := evaluator.LoadFacts(factsDir, transformedSchemas.Relations, inputNames); err != nil {
		out := runOutput{Status: "eval_error", Panic: err.Error()}
		enc, _ := json.Marshal(out)
		fmt.Println(string(enc))
		os.Exit(1)
	}
	if semiNaive {
		evaluator.RunSemiNaive(transformed, transformedStrat)
	} else {
		evaluator.RunNaive(transformed, transformedStrat.Stratum)
	}
	if err := evaluator.WriteOutput(outDir, transformedSchemas.Relations, outputNames); err != nil {
		out := runOutput{Status: "eval_error", Panic: err.Error()}
		enc, _ := json.Marshal(out)
		fmt.Println(string(enc))
		os.Exit(1)
	}
	writeProfile(evaluator, outDir)

	out := runOutput{Status: "ok", DerivationAttempts: evaluator.DerivationAttempts}
	enc, _ := json.Marshal(out)
	fmt.Println(string(enc))
}

// writeProfile writes <outDir>/profile.json in Soufflé's own -p JSON
// profile shape (ir.EmitProfile) -- harness/parse_profile.py and
// harness/tuple_report.py, already written against real Soufflé output,
// read this file unmodified.
func writeProfile(evaluator *eval.Evaluator, outDir string) {
	doc := ir.EmitProfile(evaluator.Relations)
	enc, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(outDir+"/profile.json", enc, 0o644)
}

type codegenOutput struct {
	Status      string               `json:"status"` // "ok" | "rejected" | "parse_error"
	Diagnostics []jsonSemaDiagnostic `json:"diagnostics,omitempty"`
	ParseErrors []jsonDiagnostic     `json:"parse_errors,omitempty"`
	Panic       string               `json:"panic,omitempty"`
}

// runCodegen is §4 item 1's entry point: parse -> full sema check ->
// (if clean) codegen.Generate -> write C source to outPath. Same
// parse/check pipeline as runCheck/runRun, so a program codegen accepts
// is exactly one already-accepted by every other subcommand -- codegen
// never sees a program the rest of the pipeline hasn't already
// validated.
func runCodegen(path, outPath string) {
	defer func() {
		if r := recover(); r != nil {
			out := codegenOutput{Status: "panic", Panic: fmt.Sprintf("%v", r)}
			enc, _ := json.Marshal(out)
			fmt.Println(string(enc))
			os.Exit(1)
		}
	}()

	src, err := os.ReadFile(path)
	if err != nil {
		out := codegenOutput{Status: "read_error", Panic: err.Error()}
		enc, _ := json.Marshal(out)
		fmt.Println(string(enc))
		os.Exit(1)
	}

	prog, perrs := parser.Parse(src)
	if len(perrs) > 0 {
		out := codegenOutput{Status: "parse_error"}
		for _, e := range perrs {
			out.ParseErrors = append(out.ParseErrors, jsonDiagnostic{Span: toJSONSpan(e.Span), Message: e.Message})
		}
		enc, _ := json.Marshal(out)
		fmt.Println(string(enc))
		return
	}

	var diags []sema.Diagnostic
	diags = append(diags, sema.CheckDeclType(prog)...)
	diags = append(diags, sema.CheckAllowedness(prog)...)
	stratDiags, stratResult := sema.CheckStratification(prog)
	diags = append(diags, stratDiags...)
	if len(diags) > 0 {
		out := codegenOutput{Status: "rejected"}
		for _, d := range diags {
			out.Diagnostics = append(out.Diagnostics, jsonSemaDiagnostic{
				Span: toJSONSpan(d.Span), Category: string(d.Category), Message: d.Message,
			})
		}
		enc, _ := json.Marshal(out)
		fmt.Println(string(enc))
		return
	}

	schemas, _ := sema.BuildSymbolTable(prog)
	cSource := codegen.Generate(prog, schemas.Relations, stratResult)
	if err := os.WriteFile(outPath, []byte(cSource), 0o644); err != nil {
		out := codegenOutput{Status: "eval_error", Panic: err.Error()}
		enc, _ := json.Marshal(out)
		fmt.Println(string(enc))
		os.Exit(1)
	}

	out := codegenOutput{Status: "ok"}
	enc, _ := json.Marshal(out)
	fmt.Println(string(enc))
}

// transformerRegistry names every Transformer the CLI can select via
// --transformer=. Adding a second implementation (e.g. the real magic-set
// transform) means adding one entry here -- nothing else in main.go
// changes, per transform/transformer.go's own design note.
var transformerRegistry = map[string]transform.Transformer{
	"passthrough": transform.PassThrough{},
	"magicset":    magicset.Transformer{},
	"guarded":     guard.Transformer{},
}

type emitOutput struct {
	Status      string               `json:"status"` // "ok" | "rejected" | "parse_error" | "transform_error"
	Diagnostics []jsonSemaDiagnostic `json:"diagnostics,omitempty"`
	ParseErrors []jsonDiagnostic     `json:"parse_errors,omitempty"`
	Printed     string               `json:"printed,omitempty"`
	Panic       string               `json:"panic,omitempty"`
}

// runEmit is T2's entry point: parse -> full sema check -> (if clean)
// apply the named Transformer -> print the result. This is the M3
// measurement path's front half -- "dlc decides and emits, Soufflé
// evaluates" -- so its own evaluator is never invoked here, only
// parser.Print on whatever *ast.Program the Transformer returns.
func runEmit(path, transformerName string) {
	defer func() {
		if r := recover(); r != nil {
			out := emitOutput{Status: "panic", Panic: fmt.Sprintf("%v", r)}
			enc, _ := json.Marshal(out)
			fmt.Println(string(enc))
			os.Exit(1)
		}
	}()

	src, err := os.ReadFile(path)
	if err != nil {
		out := emitOutput{Status: "read_error", Panic: err.Error()}
		enc, _ := json.Marshal(out)
		fmt.Println(string(enc))
		os.Exit(1)
	}

	prog, perrs := parser.Parse(src)
	if len(perrs) > 0 {
		out := emitOutput{Status: "parse_error"}
		for _, e := range perrs {
			out.ParseErrors = append(out.ParseErrors, jsonDiagnostic{Span: toJSONSpan(e.Span), Message: e.Message})
		}
		enc, _ := json.Marshal(out)
		fmt.Println(string(enc))
		return
	}

	var diags []sema.Diagnostic
	diags = append(diags, sema.CheckDeclType(prog)...)
	diags = append(diags, sema.CheckAllowedness(prog)...)
	stratDiags, stratResult := sema.CheckStratification(prog)
	diags = append(diags, stratDiags...)
	if len(diags) > 0 {
		out := emitOutput{Status: "rejected"}
		for _, d := range diags {
			out.Diagnostics = append(out.Diagnostics, jsonSemaDiagnostic{
				Span: toJSONSpan(d.Span), Category: string(d.Category), Message: d.Message,
			})
		}
		enc, _ := json.Marshal(out)
		fmt.Println(string(enc))
		return
	}

	transformer, ok := transformerRegistry[transformerName]
	if !ok {
		out := emitOutput{Status: "transform_error", Panic: fmt.Sprintf("unknown transformer %q", transformerName)}
		enc, _ := json.Marshal(out)
		fmt.Println(string(enc))
		os.Exit(2)
	}
	transformed, err := transformer.Transform(prog, stratResult)
	if err != nil {
		out := emitOutput{Status: "transform_error", Panic: err.Error()}
		enc, _ := json.Marshal(out)
		fmt.Println(string(enc))
		return
	}

	out := emitOutput{Status: "ok", Printed: parser.Print(transformed)}
	enc, err := json.Marshal(out)
	if err != nil {
		fmt.Fprintf(os.Stderr, "internal error marshaling output: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(enc))
}

// runExplain is NIGHT-BATCH-04 E's entry point: a debugging/presentation
// tool, not a measurement path (contrast runEmit/runRun, which print one
// machine-readable JSON document). Plain text to stdout, one fact per
// line, a fixed "TAG key=val key=val ..." shape per line so a later
// consumer (G's presentation script) can split on the first space and
// then on "=" without a real parser. Three modes, chosen by what the
// program's own front end decides, not by a flag:
//
//   - REJECTION: the program fails an existing sema check. One REJECT
//     line per diagnostic (all four grounds already share Category/
//     Span/Message -- this mode adds nothing new, it just re-renders
//     what runCheck already computes, as fact lines instead of JSON).
//   - TRANSFORM: an accepted program with a bindable query. One line per
//     adorned (predicate,adornment) pair discovered by the worklist, one
//     WORKLIST line with the iteration count, one MAGIC line per magic
//     relation generated, and one NEGATED line per negated occurrence
//     naming its pre- and post-relaxation adornment (M4-SIPS.md §1/§2).
//   - GUARD: guard.Decide's per-predicate TRANSFORM/FALLBACK verdict; if
//     anything declines, the culprit set, the cone, and the declined
//     fraction (M2-M3-BUILD.md §6/§7).
//
// A program with no bindable query (magicset.FindQueries returns empty) gets
// a single NOQUERY line instead of TRANSFORM/GUARD output -- there is
// nothing for either mode to report (M2-M3-BUILD.md §2: "expected on the
// positive fragment").
func runExplain(path string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("PANIC message=%q\n", fmt.Sprintf("%v", r))
			os.Exit(1)
		}
	}()

	src, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("READ_ERROR message=%q\n", err.Error())
		os.Exit(1)
	}

	prog, perrs := parser.Parse(src)
	if len(perrs) > 0 {
		for _, e := range perrs {
			fmt.Printf("REJECT ground=parse span=%s message=%q\n", spanText(e.Span), e.Message)
		}
		return
	}

	var diags []sema.Diagnostic
	diags = append(diags, sema.CheckDeclType(prog)...)
	diags = append(diags, sema.CheckAllowedness(prog)...)
	stratDiags, stratResult := sema.CheckStratification(prog)
	diags = append(diags, stratDiags...)
	if len(diags) > 0 {
		// REJECTION mode.
		for _, d := range diags {
			fmt.Printf("REJECT ground=%s span=%s message=%q\n", d.Category, spanText(d.Span), d.Message)
		}
		return
	}
	_ = stratResult

	// TRANSFORM mode.
	queries := magicset.FindQueries(prog)
	if len(queries) == 0 {
		fmt.Println("NOQUERY reason=\"no .output relation with a single, constant-bearing atom body\"")
		return
	}
	result, err := magicset.Adorn(prog, queries)
	if err != nil {
		fmt.Printf("TRANSFORM_ERROR message=%q\n", err.Error())
		return
	}
	for _, q := range queries {
		fmt.Printf("QUERY pred=%s rel=%s\n", q.QueryAtom.Name, q.Key.RelName())
	}
	fmt.Printf("WORKLIST iterations=%d\n", result.Iterations)
	for _, key := range result.Order {
		fmt.Printf("ADORN rel=%s rules=%d\n", key.RelName(), len(result.Rules[key]))
		fmt.Printf("MAGIC rel=%s for_rel=%s\n", key.MagicRelName(), key.RelName())
	}
	for pred := range result.Untouched {
		fmt.Printf("UNTOUCHED pred=%s\n", pred)
	}
	for _, occ := range result.NegatedOccurrenceAdornments() {
		relaxed := occ.Adorn.String() != occ.PreAdorn.String()
		reason := "all-positions-restricting"
		if relaxed {
			reason = "non-restricting-full-scan-position-relaxed"
		}
		fmt.Printf("NEGATED pred=%s pre_adorn=%s adorn=%s relaxed=%t reason=%s\n",
			occ.Pred, occ.PreAdorn.String(), occ.Adorn.String(), relaxed, reason)
	}

	// GUARD mode.
	decideResult, err := guard.Decide(prog)
	if err != nil {
		fmt.Printf("GUARD_ERROR message=%q\n", err.Error())
		return
	}
	if len(decideResult.DeclinedRelations) == 0 {
		fmt.Println("GUARD verdict=STRATIFIABLE")
	} else {
		fmt.Println("GUARD verdict=UNSTRATIFIABLE clause=a")
		idb := map[string]bool{}
		for _, c := range prog.Clauses {
			idb[c.Head.Name] = true
		}
		fmt.Printf("GUARD culprit=%s cone=%s declined_fraction=%.3f\n",
			joinedSortedKeys(decideResult.CulpritPredicates),
			joinedSortedKeys(decideResult.ConeRelations),
			float64(len(decideResult.DeclinedRelations))/float64(len(idb)))
	}
	idbAll := map[string]bool{}
	for _, c := range prog.Clauses {
		idbAll[c.Head.Name] = true
	}
	for _, pred := range sortedKeysOf(idbAll) {
		action := "TRANSFORM"
		if decideResult.DeclinedRelations[pred] {
			action = "FALLBACK"
		}
		fmt.Printf("DECISION pred=%s action=%s\n", pred, action)
	}
}

func spanText(s token.Span) string {
	return fmt.Sprintf("%d:%d-%d:%d", s.Start.Line, s.Start.Col, s.End.Line, s.End.Col)
}

func joinedSortedKeys(m map[string]bool) string {
	keys := sortedKeysOf(m)
	if len(keys) == 0 {
		return "{}"
	}
	return "{" + strings.Join(keys, ",") + "}"
}

func sortedKeysOf(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

type jsonDiagnostic struct {
	Span    jsonSpan `json:"span"`
	Message string   `json:"message"`
}

type parseOutput struct {
	Status      string           `json:"status"` // "parsed" | "error"
	DeclCount   int              `json:"decl_count"`
	ClauseCount int              `json:"clause_count"`
	ErrorCount  int              `json:"error_count"`
	Diagnostics []jsonDiagnostic `json:"diagnostics"`
	Panic       string           `json:"panic,omitempty"`
}

// runParse is §3.3 gates one and three's entry point: parse the file,
// report whether it parsed with zero errors, and the diagnostics if not.
func runParse(path string) {
	defer func() {
		if r := recover(); r != nil {
			out := parseOutput{Status: "panic", Panic: fmt.Sprintf("%v", r)}
			enc, _ := json.Marshal(out)
			fmt.Println(string(enc))
			os.Exit(1)
		}
	}()

	src, err := os.ReadFile(path)
	if err != nil {
		out := parseOutput{Status: "read_error", Panic: err.Error()}
		enc, _ := json.Marshal(out)
		fmt.Println(string(enc))
		os.Exit(1)
	}

	prog, errs := parser.Parse(src)
	out := parseOutput{
		Status:      "parsed",
		DeclCount:   len(prog.Decls),
		ClauseCount: len(prog.Clauses),
		ErrorCount:  len(errs),
	}
	if len(errs) > 0 {
		out.Status = "error"
	}
	for _, e := range errs {
		out.Diagnostics = append(out.Diagnostics, jsonDiagnostic{Span: toJSONSpan(e.Span), Message: e.Message})
	}
	enc, err := json.Marshal(out)
	if err != nil {
		fmt.Fprintf(os.Stderr, "internal error marshaling output: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(enc))
}

type jsonSemaDiagnostic struct {
	Span     jsonSpan `json:"span"`
	Category string   `json:"category"`
	Message  string   `json:"message"`
}

type checkOutput struct {
	Status      string               `json:"status"` // "ok" | "rejected" | "parse_error"
	Diagnostics []jsonSemaDiagnostic `json:"diagnostics,omitempty"`
	ParseErrors []jsonDiagnostic     `json:"parse_errors,omitempty"`
	Strata      map[string]int       `json:"strata,omitempty"` // relation -> stratum, only when status=="ok"
	Panic       string               `json:"panic,omitempty"`
}

// runCheck is §3.4/§3.5/§3.6's entry point: parse, then run every sema
// checker implemented so far, reporting every Diagnostic with its
// Category (docs/reports/night02-T9-diagnostics.md's four grounds).
// Message text is dlc's own; only the classification is required to
// agree with Soufflé (§3.4's own instruction).
func runCheck(path string) {
	defer func() {
		if r := recover(); r != nil {
			out := checkOutput{Status: "panic", Panic: fmt.Sprintf("%v", r)}
			enc, _ := json.Marshal(out)
			fmt.Println(string(enc))
			os.Exit(1)
		}
	}()

	src, err := os.ReadFile(path)
	if err != nil {
		out := checkOutput{Status: "read_error", Panic: err.Error()}
		enc, _ := json.Marshal(out)
		fmt.Println(string(enc))
		os.Exit(1)
	}

	prog, perrs := parser.Parse(src)
	if len(perrs) > 0 {
		out := checkOutput{Status: "parse_error"}
		for _, e := range perrs {
			out.ParseErrors = append(out.ParseErrors, jsonDiagnostic{Span: toJSONSpan(e.Span), Message: e.Message})
		}
		enc, _ := json.Marshal(out)
		fmt.Println(string(enc))
		return
	}

	var diags []sema.Diagnostic
	diags = append(diags, sema.CheckDeclType(prog)...)
	diags = append(diags, sema.CheckAllowedness(prog)...)
	stratDiags, stratResult := sema.CheckStratification(prog)
	diags = append(diags, stratDiags...)

	out := checkOutput{Status: "ok"}
	if len(diags) > 0 {
		out.Status = "rejected"
	} else if stratResult != nil {
		out.Strata = stratResult.Stratum
	}
	for _, d := range diags {
		out.Diagnostics = append(out.Diagnostics, jsonSemaDiagnostic{
			Span: toJSONSpan(d.Span), Category: string(d.Category), Message: d.Message,
		})
	}
	enc, err := json.Marshal(out)
	if err != nil {
		fmt.Fprintf(os.Stderr, "internal error marshaling output: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(enc))
}

type roundtripOutput struct {
	Status      string           `json:"status"` // "match" | "mismatch" | "parse_error" | "reparse_error"
	Diagnostics []jsonDiagnostic `json:"diagnostics,omitempty"`
	Printed     string           `json:"printed,omitempty"`
	Panic       string           `json:"panic,omitempty"`
}

// runRoundtrip is §3.3 gate two's entry point: parse -> print -> reparse
// -> ast.Equal, entirely in Go (parser.Roundtrip), reporting only the
// verdict -- see parser/DESIGN.md for why the comparison itself belongs
// in Go, not reimplemented against a JSON dump in Python.
func runRoundtrip(path string) {
	defer func() {
		if r := recover(); r != nil {
			out := roundtripOutput{Status: "panic", Panic: fmt.Sprintf("%v", r)}
			enc, _ := json.Marshal(out)
			fmt.Println(string(enc))
			os.Exit(1)
		}
	}()

	src, err := os.ReadFile(path)
	if err != nil {
		out := roundtripOutput{Status: "read_error", Panic: err.Error()}
		enc, _ := json.Marshal(out)
		fmt.Println(string(enc))
		os.Exit(1)
	}

	r := parser.Roundtrip(src)
	out := roundtripOutput{Status: r.Status, Printed: r.Printed}
	for _, e := range r.ParseErrors {
		out.Diagnostics = append(out.Diagnostics, jsonDiagnostic{Span: toJSONSpan(e.Span), Message: e.Message})
	}
	enc, err := json.Marshal(out)
	if err != nil {
		fmt.Fprintf(os.Stderr, "internal error marshaling output: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(enc))
}
