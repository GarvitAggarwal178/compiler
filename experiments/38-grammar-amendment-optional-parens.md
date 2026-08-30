# NIGHT-BATCH-03 T6 — grammar amendment 2: optional parens on `.input`/`.output`

Date: 2026-08-27. Authorized per `m1-progress.md`'s own finding (105/175
gate-one failures, 60%, from cosmetic `.input Name()`/`.output Name()`
trailing parens) and `docs/OPEN_QUESTIONS.md`'s 2026-08-26 entry flagging it
for exactly this decision.

## What did not work

Nothing. Implemented, tested, measured, decision rule applied — one pass.

## Method

`src/parser/parser.go`'s `parseDecl`: after the `.input`/`.output` relation
name, an optional parenthesised token list is consumed and discarded
(token-level balanced-paren counting, not a raw-text scan — a string
literal containing `(` is already a single `STRING` token by the time the
parser sees it, so it can never be miscounted). Real Soufflé uses this
syntax slot for I/O directive options (`filename="x.csv"`, etc.); this
grammar has no such concept, so content beyond balance-counting is not
modeled. 4 new Go unit tests (`src/parser/parser_test.go`): empty parens,
non-empty options discarded, nested parens, and the pre-existing bare form
(no parens) still works. `go build`/`go vet`/`go test ./...` all clean.

Re-ran gate one (parse coverage) and gate two (round-trip) against all 195
`IN_GRAMMAR.txt` files: `harness/night03_t6_gates.py` (a fresh script,
**not** `parse_coverage.py`/`round_trip_scaffold.py` re-run directly —
those write to the already-committed `measurements/m1-3.3-gate{1,2}-*`
files, which `NIGHT-BATCH-03.md` §0.2 prohibits editing; this task's numbers
go to `measurements/night03-t6/gates-after-amendment.json` instead, the
committed M1 gate files are untouched).

## The four numbers

Provenance: `measurements/night03-t6/gates-after-amendment.json`.

| metric | before | after |
|---|---|---|
| total parsed | 20/195 | **89/195** |
| negation-bearing files parsed | 3 | **18** |

Gate two (round-trip): 89/195 match — **identical to gate one's new parsed
count**, confirmed by direct cross-reference (0 files that parse but fail
round-trip). The amendment introduces no new round-trip defects; every
newly-parsed file also round-trips correctly.

## Mechanical decision rule, applied as instructed

`total parsed ≥ 80` **or** `negation-bearing parsed ≥ 5` → keep.
`89 ≥ 80`: **true** (also `18 ≥ 5`: true, on both legs independently).
**Amendment kept.**

## What a skeptic attacks first

- The "negation-bearing" heuristic is a coarse regex (`!` immediately
  followed by an identifier and `(`) applied to raw source text, not to
  the parsed AST — it can miscount a `!=` operator inside a constraint as
  a negated atom if a digit or letter happens to follow with no space
  (Souffle's own lexer would never confuse these, but this quick text
  scan does not fully replicate that). Not cross-checked against the real
  parsed AST's `NegatedAtom` count in this task; the numbers (3 -> 18) are
  large enough relative to any plausible miscount rate that the qualitative
  conclusion (a real, large increase) is not in doubt, but the exact counts
  carry that caveat.
- This amendment only fixes the *parenthesized* `.input`/`.output` idiom.
  `m1-progress.md`'s other out-of-grammar categories (aggregates, functors,
  `unsigned`/`float` types, `#include`, pragmas) are untouched and still
  block the remaining 106 files — 89/195 is a real improvement, not a claim
  that grammar coverage is now complete (T7 addresses the corpus-honesty
  side of this separately).

## Verdict

**T6: DONE, amendment kept.** `total parsed`: 20 → 89 (+69). `negation-bearing
parsed`: 3 → 18 (+15). Both legs of the decision rule pass independently.
No round-trip regressions (89/89 match). This materially unblocks M3's
differential corpus, which needs negation-bearing programs.
