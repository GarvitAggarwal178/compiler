# package codegen

Emits a standalone C program from an already-checked `*ast.Program` --
`Generate` returns C source text; nothing here executes anything itself.
§4 item 1.

**Scoped to naive evaluation only, not semi-naive.** §4 item 1's own
wording ("nested loops over relations with hash indices") describes
naive evaluation's shape, not semi-naive's Δ-rewrite; implementing the
Δ-rewrite in generated C (per-SCC delta relations, position-keyed
overrides, `eval/seminaive.go`'s whole design) would roughly double this
package's size for a milestone item explicitly framed as "may be
rubric-mandatory... getting it early de-risks that," not as a second
full evaluator. Naive is what's here; disclosed, not silently narrowed.

**Types are erased to `int64_t` uniformly — numbers and interned symbol
ids share one C representation.** Unlike the Go interpreter's `ir.Value`
(which needs a runtime `IsString` tag because the same Go code path
handles every relation), codegen knows each column's declared type
*at generation time* from the schema, so there is nothing to tag at
runtime: a `symbol` column's `int64_t` is always an interned id, a
`number` column's is always the number, and the generated code for a
given column never needs to ask which. This is a real advantage of AOT
codegen over the interpreter, not a corner cut.

**FIXED (NIGHT-BATCH-03 T8): ordering comparisons (`<`,`<=`,`>`,`>=`) on
`symbol` columns now compare strings, not interned ids.** Previously
compared interned ids directly, and interned ids are assigned in
first-seen order, not alphabetical order, so `id(a) < id(b)` did not
generally mean `a` sorts before `b` lexicographically. Fixed by exporting
`sema.ClauseVarTypes` (a thin wrapper around the same per-clause
`clauseChecker.varTypes` this note used to say was unexported) and, in
`emitConstraint`, emitting `strcmp(str_lookup(x), str_lookup(y)) OP 0`
instead of a raw integer comparison whenever an ordering operator's
operand is symbol-typed (`isSymbolArith`: a bare string literal, or a
`Var` `ClauseVarTypes` resolves to `"symbol"` — arithmetic operators are
always number, so no other `ast.Arith` shape can be symbol-typed).
`=`/`!=` were never affected (id equality is string equality regardless
of intern order) and are unchanged. 4 new end-to-end tests
(`codegen_test.go`, `TestCodegenSymbolOrdering{LessThan,LessOrEqual,
GreaterThan,GreaterOrEqual}`) construct facts where intern order is the
*reverse* of lexicographic order (`"zebra"` interned before `"apple"`,
exactly the case this note used to warn about) so a regression produces a
different answer *set*, not just a different row order; all 4 confirmed
to fail before the fix and pass after, and independently cross-checked
against real Soufflé on the same program
(`experiments/40-codegen-symbol-ordering-gap.md`).

**No fixed-size scratch buffer anywhere, on purpose — a real bug caught
before it was ever run.** A first draft's `_lookup_c0` helper collected
matching row indices into a `1 << 20`-entry static array before handing
them to the caller; `NIGHT-BATCH-02`'s own measurements found relations
in the millions of tuples at this project's larger scale points, so a
single hash bucket could plausibly exceed that bound and silently
overflow. Replaced before any generated C was ever compiled: join
codegen (`emitAtomJoin`) walks the hash bucket's linked list directly
inline instead of through any intermediate buffer, so there is no size
bound to exceed regardless of relation size.

**`safeOrderForCodegen` duplicates `eval.safeOrder`'s algorithm rather
than importing/exporting it.** Both need the identical greedy
literal-reordering logic (a negated atom or a grounding-role constraint
schedules only once every variable it needs is already bound; a positive
atom is always immediately schedulable) for the identical reason (naive
left-to-right evaluation order is not what allowedness's fixpoint,
`sema/DESIGN.md`, guarantees). Exporting `eval`'s version was considered
and rejected: `codegen` operates on `ast.Literal`/`ast.Term` directly to
emit *text*, not to build Go-side bindings the way `eval.evalBody` does,
so the two callers' actual needs diverge past the shared reordering
step, and forcing one shared exported function to serve both would
couple two otherwise-independent packages for a ~30-line algorithm. The
duplication is small and disclosed here rather than hidden.

**Head-position `Wildcard` gets the same fixed-value fallback as
`eval.buildTuple`** (`groundExprTerm` emits `"0"`) — same reasoning,
`eval/DESIGN.md`: real Soufflé rejects this construct outright and sema
doesn't check for it yet; codegen degrades to "compiles and produces a
wrong-but-harmless tuple" rather than emitting invalid C or crashing the
generator.
