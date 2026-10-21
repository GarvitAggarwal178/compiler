// Command dlc is the compiler's CLI entry point. Lane B.
package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"

	"strings"

	"dlc/src/ast"
	"dlc/src/codegen"
	"dlc/src/eval"
	"dlc/src/inspect"
	"dlc/src/ir"
	"dlc/src/lexer"
	"dlc/src/parser"
	"dlc/src/sema"
	"dlc/src/token"
	"dlc/src/transform"
	"dlc/src/transform/compact"
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
	case "lex", "parse", "roundtrip", "check", "explain":
		if len(os.Args) != 3 {
			fmt.Fprintf(os.Stderr, "usage: dlc %s <file>\n", subcommand)
			os.Exit(2)
		}
	}

	switch subcommand {
	case "inspect":
		runInspect(path, transformerFlag(os.Args[3:]))
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
		if len(os.Args) < 4 || len(os.Args) > 5 {
			fmt.Fprintln(os.Stderr, "usage: dlc codegen <file> <outfile.c> [--transformer=name]")
			os.Exit(2)
		}
		runCodegen(path, os.Args[3], transformerFlag(os.Args[4:]))
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
	if out.ErrorCount > 0 {
		os.Exit(1)
	}
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
	if len(args) > 1 {
		fmt.Fprintln(os.Stderr, "expected at most one --transformer flag")
		os.Exit(2)
	}
	for _, a := range args {
		if v, ok := strings.CutPrefix(a, "--transformer="); ok {
			name = v
		} else {
			fmt.Fprintf(os.Stderr, "unknown argument %q\n", a)
			os.Exit(2)
		}
	}
	if _, ok := transformerRegistry[name]; !ok {
		fmt.Fprintf(os.Stderr, "unknown transformer %q\n", name)
		os.Exit(2)
	}
	return name
}

// applyTransformer is shared by execution, C generation and emission. Never
// reuse source schemas/strata for a rewritten AST. Only experimental raw
// emission may expose an unstratifiable candidate for inspection.
func applyTransformer(prog *ast.Program, strata *sema.StratumResult, name string, allowUnstratifiable bool) (*ast.Program, *sema.SymbolTable, *sema.StratumResult, error) {
	p, st, ss, _, err := applyTransformerDetailed(prog, strata, name, allowUnstratifiable)
	return p, st, ss, err
}
func applyTransformerDetailed(prog *ast.Program, strata *sema.StratumResult, name string, allowUnstratifiable bool) (*ast.Program, *sema.SymbolTable, *sema.StratumResult, *transform.Report, error) {
	tr, ok := transformerRegistry[name]
	if !ok {
		return nil, nil, nil, nil, fmt.Errorf("unknown transformer %q", name)
	}
	var q *ast.Program
	var report *transform.Report
	var err error
	if reporting, ok := tr.(transform.ReportingTransformer); ok {
		var out *transform.Outcome
		out, err = reporting.TransformReported(prog, strata)
		if err == nil {
			q, report = out.Program, out.Report
		}
	} else {
		q, err = tr.Transform(prog, strata)
	}
	if err != nil {
		return nil, nil, nil, nil, err
	}
	diags := append(sema.CheckDeclType(q), sema.CheckAllowedness(q)...)
	sd, qs := sema.CheckStratification(q)
	if !allowUnstratifiable {
		diags = append(diags, sd...)
	}
	if len(diags) > 0 {
		return nil, nil, nil, nil, fmt.Errorf("transformed program failed semantic checks: %v", diags)
	}
	st, _ := sema.BuildSymbolTable(q)
	return q, st, qs, report, nil
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
		os.Exit(1)
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
		os.Exit(1)
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
	transformed, transformedSchemas, transformedStrat, terr := applyTransformer(prog, stratResult, transformerName, false)
	if terr != nil {
		out := runOutput{Status: "eval_error", Panic: terr.Error()}
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
	if err := writeProfile(evaluator, outDir); err != nil {
		out := runOutput{Status: "eval_error", Panic: err.Error()}
		enc, _ := json.Marshal(out)
		fmt.Println(string(enc))
		os.Exit(1)
	}

	out := runOutput{Status: "ok", DerivationAttempts: evaluator.DerivationAttempts}
	enc, _ := json.Marshal(out)
	fmt.Println(string(enc))
}

// writeProfile writes <outDir>/profile.json in Soufflé's own -p JSON
// profile shape (ir.EmitProfile) -- harness/parse_profile.py and
// harness/tuple_report.py, already written against real Soufflé output,
// read this file unmodified.
func writeProfile(evaluator *eval.Evaluator, outDir string) error {
	doc := ir.EmitProfile(evaluator.Relations)
	enc, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(outDir+"/profile.json", enc, 0o644)
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
func runCodegen(path, outPath, transformerName string) {
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
		os.Exit(1)
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
		os.Exit(1)
	}

	compiled, schemas, compiledStrata, report, err := applyTransformerDetailed(prog, stratResult, transformerName, false)
	if err != nil {
		out := codegenOutput{Status: "transform_error", Panic: err.Error()}
		enc, _ := json.Marshal(out)
		fmt.Println(string(enc))
		os.Exit(1)
	}
	cSource := fmt.Sprintf("/* source SHA-256 %x; mode %s; naive evaluation */\n", sha256.Sum256(src), transformerName) + codegen.GenerateWithReport(compiled, schemas.Relations, compiledStrata, report)
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
	"passthrough":     transform.PassThrough{},
	"magicset":        magicset.Transformer{},
	"guarded":         guard.Transformer{},
	"guarded-compact": compact.Transformer{},
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
		os.Exit(1)
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
		os.Exit(1)
	}

	transformed, _, _, err := applyTransformer(prog, stratResult, transformerName, transformerName == "magicset")
	if err != nil {
		out := emitOutput{Status: "transform_error", Panic: err.Error()}
		enc, _ := json.Marshal(out)
		fmt.Println(string(enc))
		os.Exit(1)
	}

	out := emitOutput{Status: "ok", Printed: parser.Print(transformed)}
	enc, err := json.Marshal(out)
	if err != nil {
		fmt.Fprintf(os.Stderr, "internal error marshaling output: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(string(enc))
}

// runInspect exposes versioned compiler-owned data; runExplain renders a
// human summary. The demo consumes inspection JSON instead of parsing text.
func runInspect(path, mode string) {
	src, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "reading source: %v\n", err)
		os.Exit(1)
	}
	report := inspect.Build(src, transformerRegistry[mode], mode)
	if err := json.NewEncoder(os.Stdout).Encode(report); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if report.Status != "ok" {
		os.Exit(1)
	}
}

func runExplain(path string) {
	src, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("READ_ERROR message=%q\n", err.Error())
		os.Exit(1)
	}
	r := inspect.Build(src, guard.Transformer{}, "guarded")
	if r.Status != "ok" {
		for _, d := range r.Diagnostics {
			location := "unavailable"
			if d.Span != nil {
				location = fmt.Sprintf("%d:%d-%d:%d", d.Span.Start.Line, d.Span.Start.Col, d.Span.End.Line, d.Span.End.Col)
			}
			fmt.Printf("REJECT ground=%s span=%s stage=%s message=%q\n", d.Category, location, d.Stage, d.Message)
		}
		os.Exit(1)
	}
	fmt.Println("SOURCE status=valid")
	report := r.Transform
	if len(report.Queries) == 0 {
		fmt.Printf("NOQUERY reason=%q\n", report.NoQueryReason)
		return
	}
	for _, q := range report.Queries {
		fmt.Printf("QUERY pred=%s rel=%s\n", q.SourcePredicate, q.ViewID)
	}
	fmt.Printf("WORKLIST pops=%d\n", report.WorklistPops)
	for _, rel := range report.Relations {
		switch rel.Role {
		case "adorned":
			fmt.Printf("ADORN rel=%s source=%s\n", rel.ID, rel.SourcePredicate)
		case "magic":
			fmt.Printf("MAGIC rel=%s source=%s\n", rel.ID, rel.SourcePredicate)
		}
	}
	for _, n := range report.Negations {
		fmt.Printf("NEGATED pred=%s pre_adorn=%s adorn=%s\n", n.Predicate, strings.Join(n.PreBindings, ""), strings.Join(n.Bindings, ""))
	}
	if r.Candidate != nil {
		fmt.Printf("CANDIDATE status=%s\n", r.Candidate.Status)
	}
	fmt.Printf("FINAL status=%s\n", r.Final.Status)
	if len(report.Culprits) > 0 {
		fmt.Printf("GUARD verdict=UNSTRATIFIABLE clause=a culprit=%s cone=%s\n", strings.Join(report.Culprits, ","), strings.Join(report.Cone, ","))
	} else {
		fmt.Println("GUARD verdict=STRATIFIABLE")
	}
	for _, d := range report.Decisions {
		action := "ORIGINAL"
		if d.Fallback {
			action = "FALLBACK"
		} else if d.Specialized {
			action = "SPECIALIZED"
		}
		fmt.Printf("DECISION pred=%s action=%s full_extent_retained=%t reasons=%q\n", d.SourcePredicate, action, d.FullExtentRetained, strings.Join(d.RetainedReasons, "; "))
	}
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
	if len(errs) > 0 {
		os.Exit(1)
	}
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
		os.Exit(1)
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
	if out.Status != "ok" {
		os.Exit(1)
	}
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
	if r.Status != "match" {
		os.Exit(1)
	}
}
