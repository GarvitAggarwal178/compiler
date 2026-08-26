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
`--transformer=<name>`.
