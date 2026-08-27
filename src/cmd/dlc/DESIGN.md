# package main (cmd/dlc)

The CLI entry point: subcommand dispatch (`dlc <subcommand> <file>`) plus
a JSON output contract per subcommand, so `harness/` scripts can drive the
real compiler the same way they already drive Soufflé (parse the output,
don't scrape human-readable text).

**Key decision: a separate `jsonToken`/`jsonSpan` shape, not `json:` tags
on `token.Token` directly.** `token.Token` is an internal type whose shape
should be free to change as the parser/sema grow (e.g. adding fields
nothing outside the compiler needs). The CLI's JSON contract is what the
harness depends on and needs to stay stable independently of internal
refactors — coupling them would mean every internal token-package change
risks silently breaking a Python script three layers away.

**Key decision: a top-level `recover()` per subcommand, even though the
lexer is specified to never panic.** This is not a substitute for that
guarantee (`src/lexer/DESIGN.md` explains why the lexer itself doesn't
need one) — it is what turns "the guarantee turned out to be wrong" into
a clean, machine-readable `{"status":"panic", ...}` document and exit code
1, instead of an uncontrolled Go stack trace on stderr that a harness
script would have to pattern-match to even detect. The §3.1 gate is "zero
panics"; this is what makes that gate detectable at all if it ever fails,
rather than merely hoped-for.

Subcommands: `lex` (implemented, §3.1). `parse`, once §3.3 lands, will
follow the same JSON-document convention.

**`emit` (NIGHT-BATCH-03 T2).** Parses, runs the full sema pipeline, applies
a named `transform.Transformer` (via `transformerRegistry`, a
`map[string]transform.Transformer` — adding a second implementation, e.g.
the real magic-set transform, is a one-line registry addition, nothing else
in `main.go` changes), and prints the result with `parser.Print`. This is
the front half of the M3 measurement protocol: `dlc` decides and emits,
Soufflé evaluates — `emit` never runs `dlc`'s own evaluator. Default
transformer is `passthrough` (`transform.PassThrough`); select another with
`--transformer=<name>`. Registry entries as of M3: `passthrough`, `magicset`
(M2's ungated transform), `guarded` (the full M3 pipeline — magicset,
guarded by `src/transform/guard`'s culprit-cycle detection and per-SCC
TRANSFORM/FALLBACK decision).

**`run`/`run-seminaive` (§3.8/§3.9, extended M3.4).** Both now also accept
an optional trailing `--transformer=<name>` (shared parsing,
`transformerFlag`, with `emit`), defaulting to `passthrough`. After the
usual parse+sema pipeline accepts the ORIGINAL program, the named
`Transformer` is applied and `sema.CheckStratification` is re-run on its
OUTPUT — never reusing the pre-transform `StratumResult`, since a real
transform changes the precedence graph (`src/transform/DESIGN.md`'s own
documented caller contract). The transformed/possibly-mixed program is
then evaluated by the same `RunNaive`/`RunSemiNaive` any other program
uses — no second evaluation path exists for a mixed (partially
transformed, partially fallback) program (`src/eval/fallback.go`,
`docs/reports/m3-4-fallback.md`).

**`explain` (NIGHT-BATCH-04 E).** A debugging/presentation tool, not a
measurement path — plain text, one fact per line (`TAG key=val ...`),
not the JSON-document contract every other subcommand follows (nothing
in `harness/` parses this format as a measurement; a later presentation
script may). No `--transformer=` flag: unlike `emit`/`run`, `explain`
always runs both `magicset.Adorn` (unconditionally, not gated by a
chosen transformer) and `guard.Decide` — a user asking "why did this
happen" wants the real mechanical adornment and the real guard verdict,
not a choice between them. Three modes chosen by what the front end
decides, not a flag: REJECTION (one `REJECT` line per diagnostic,
covering the parser and all four sema grounds, same
`Category`/`Span`/`Message` fields `runCheck`'s JSON already carries, a
different rendering not new information), TRANSFORM (adorned predicates,
worklist iterations, magic relations, and per negated occurrence its
pre-/post-relaxation adornment — M4-SIPS.md §1/§2), GUARD (per-predicate
TRANSFORM/FALLBACK, culprit set, cone, declined fraction when the guard
fires). Samples for every mode and all four rejection grounds committed
under `docs/reports/explain-samples/` (`docs/reports/night04-E-explain.md`).
