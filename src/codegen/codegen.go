// Package codegen emits a standalone C program implementing naive
// evaluation (§3.8's semantics, not §3.9's semi-naive Δ-rewrite -- see
// DESIGN.md for the scoping) of a dlc program that has already passed
// every sema check. §4 item 1. Lane B.
package codegen

import (
	"fmt"
	"sort"
	"strings"

	"dlc/src/ast"
	"dlc/src/sema"
)

// Generate returns a complete C source file (compiles with a plain
// `cc -O2 -o prog prog.c`) implementing prog's naive evaluation.
// The generated binary's CLI mirrors dlc's own `run` subcommand:
// `./prog <factsDir> <outDir>` -- load every .input relation's
// <name>.facts from factsDir (Soufflé's own tab-separated, no-header
// convention), evaluate to a fixpoint per stratum, write every .output
// relation's <name>.csv to outDir in the same format.
func Generate(prog *ast.Program, schemas map[string]*sema.RelationSchema, strata *sema.StratumResult) string {
	g := &generator{schemas: schemas}
	for _, d := range prog.Decls {
		switch d.Kind {
		case ast.DeclInput:
			g.inputNames = append(g.inputNames, d.Name)
		case ast.DeclOutput:
			g.outputNames = append(g.outputNames, d.Name)
		}
	}
	sort.Strings(g.inputNames)
	sort.Strings(g.outputNames)

	g.emitPrelude()
	g.emitRelationStorage()
	g.emitFactLoading()
	g.emitEvaluation(prog, strata)
	g.emitOutput()
	g.emitMain()
	return g.sb.String()
}

// generator accumulates the emitted C source. Deliberately not
// concurrent, not reused across calls.
type generator struct {
	sb          strings.Builder
	schemas     map[string]*sema.RelationSchema
	inputNames  []string
	outputNames []string

	// curVarTypes is the current clause's variable -> declared-type
	// ("number"/"symbol") map, set by emitRuleClause before emitting a
	// clause's body and read by emitConstraint (NIGHT-BATCH-03 T8) to
	// decide whether a `<`/`<=`/`>`/`>=` comparison needs lexicographic
	// (symbol) or integer (number) semantics. nil outside of emitting a
	// rule clause's body -- a fact clause has no constraints to evaluate,
	// so nothing reads it there.
	curVarTypes map[string]string
}

func (g *generator) w(format string, args ...interface{}) {
	fmt.Fprintf(&g.sb, format, args...)
	g.sb.WriteByte('\n')
}

// relNames returns every declared relation name, sorted (determinism --
// CLAUDE.md's own rule, and Go map iteration order is not stable).
func (g *generator) relNames() []string {
	names := make([]string, 0, len(g.schemas))
	for name := range g.schemas {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func cIdent(name string) string {
	// Datalog identifiers are ASCII letters/digits/underscore
	// (lexer/DESIGN.md) -- always already a valid C identifier fragment;
	// prefixed to avoid collisions with C keywords/the runtime's own
	// helper names.
	return "rel_" + name
}
