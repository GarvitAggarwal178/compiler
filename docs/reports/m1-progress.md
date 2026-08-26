# M1 progress report

Date: 2026-08-26. Per `docs/M1-BUILD.md` §7. **All 9 §3 work items
complete; all 3 §4 items complete.** `src/` did not exist at session
start; `CLAUDE.md` §2 was superseded by `M1-BUILD.md` §1 for this entire
session, narrowing Lane A to exactly three components (magic-set
transform, guard, fallback evaluation wiring) — all three remain
untouched: `doc.go` markers only.

## What did not work (before the results, per `CLAUDE.md` §7)

- **§3.1:** `IN_GRAMMAR.txt` (NIGHT-BATCH-01 T5's mechanical scan) admits
  more out-of-grammar content than previously known — 42/195 files
  produce lex-error tokens (aggregates, records, `#include`, pragmas),
  none of it a lexer bug. Flagged in `docs/OPEN_QUESTIONS.md`, not fixed
  (out of this session's scope).
- **§3.3 gate one:** `parsed/195 = 20/195`, far below a naive reading of
  "195/195." The dominant cause (105/175 failures, 60%) is a single,
  deliberate grammar decision: `dlc` does not admit `.input`/`.output
  Name()` with cosmetic trailing parens, a real Soufflé idiom not
  authorized by M1-BUILD.md's one grammar amendment. Reported as a human
  decision point (`docs/OPEN_QUESTIONS.md`), not silently added as a
  second amendment.
- **§3.6:** the intended Soufflé oracle signal (`--show=initial-ram`'s
  `SUBROUTINE` list) turned out to be alphabetically sorted, not
  evaluation order — the real signal (`BEGIN MAIN ... CALL stratum_X`)
  was found only after that first attempt didn't hold up.
- **§3.9:** a first version of the semi-naive evaluator grouped clauses
  by stratum *number*, and the differential harness immediately caught a
  real disagreement against Soufflé on `example/josephus/josephus.dl` —
  fixed by adding `SCCOrder` (a topological order of the SCC condensation)
  to `sema.StratumResult` and processing one SCC at a time instead of one
  stratum-number batch. Full account: `src/eval/DESIGN.md`.
- **§4 item 1:** a first draft of the C codegen's column-0 lookup used a
  fixed `1<<20`-entry buffer that could plausibly overflow at this
  project's own larger measured scale points — replaced with an inline
  hash-bucket walk before any generated C was ever compiled.
- **This session's own tooling:** the Bash-tool/`wsl.exe` bridge
  corrupted two heredoc-based file writes early on (backtick spans inside
  double-quoted shell content got interpreted as command substitution by
  the outer shell) — caught and corrected via `Edit`, not left standing;
  all later multi-line content went through `Write`/`Edit` instead of
  inline heredocs.

## Items complete, gate numbers

| item | gate | result |
|---|---|---|
| 3.1 lexer | zero panics, 195 in-grammar + 39 hostile | **0/234 panics** |
| 3.2 AST | none stated; compiles + `ast.Equal` built | done |
| 3.3 parser gate one | parsed/195 | **20/195** (attributed shortfall, see above) |
| 3.3 parser gate two | round-trip match/195 | **20/195** (== gate one's parseable set; **20/20 of those match**) |
| 3.3 parser gate three | hostile-corpus oracle agreement | **35/39 agree**, 2 expected sema-not-yet-implemented gaps, 1 known-inconclusive (T2), **1 deliberate, documented disagreement** (unterminated block comment) |
| 3.4 sema decl/type | none stated; rejection-corpus classification | **6/6** (arity+type) |
| 3.5 allowedness | 15 probe programs + 13 rejection cases | **15/15** probes; **10/13** rejection cases (3 attributed to §3.6 not yet existing at that point) |
| 3.6 stratification | rejects unstratifiable + agrees with Soufflé's order | **3/3** reject; **1/1** usable Soufflé-order comparison agrees (only 1 of 195 files both parses and has real negation) |
| 3.7 storage/indices | none stated; instrumentation required | done, 8 tests |
| 3.8 naive eval | agreed/attempted vs Soufflé | **11/20 agreed, 0 disagreements** (9 correctly sema-rejected before evaluation) |
| 3.9 semi-naive gate one | same as 3.8, unchanged | **11/20**, re-matched exactly after the SCC-order fix |
| 3.9 semi-naive gate two (**M1's headline**) | `T_naive` vs `T_semi-naive`, 5 pre-registered shapes | **identical, 1.00x, every shape** (expected — see below); real signal is `DerivationAttempts`: 2.00x–18.93x fewer, per shape, never aggregated |
| 4.1 C codegen | none stated; built + tested | **8/8** real generate→compile→run tests pass |
| 4.2 transform interface | none stated; defined + placeholder | done, 2 tests |
| 4.3 full-family differential | agreed/attempted, every scale point | **16/16 comparable points matched, 0 disagreements** (4 DNF at each shape's own scale-up point, capped at 120s, never retried higher — see below) |

**Capstone, after 3.4–3.6 all landed:** all 13 `tests/rejection/` cases
across all 4 grounds now correctly rejected end-to-end
(`measurements/m1-sema-full-rejection-corpus-13-of-13.json`).

**Test suite:** 83 Go tests across 9 packages, 0 failing.
`go build`/`go vet`/`go test ./...` clean throughout every commit.

## M1's headline number, in full

`T_naive == T_semi_naive` exactly, all 5 `BENCHMARK_FAMILY` shapes, at
each shape's own smallest pre-registered scale point:

| shape | `T_naive` = `T_semi_naive` | `DerivationAttempts` naive → semi-naive | ratio |
|---|---|---|---|
| same_generation_negation | 51,301 | 102,602 → 51,301 | 2.00× |
| transitive_closure_bound | 86,618 | 3,541,194 → 187,059 | 18.93× |
| ancestor_nonancestor | 250,450 | 1,969,412 → 318,769 | 6.18× |
| reachability_complement | 62,534 | 1,786,767 → 110,237 | 16.21× |
| culprit_cycle | 286 | 2,680 → 443 | 6.05× |

The exact-equality result is **mathematically expected, not a null
result**: both evaluators compute the identical minimal Herbrand model
of the same program, and both count each distinct tuple once (this
project's own convention since Phase 0) — the final counts cannot differ
by construction. `DerivationAttempts` (every candidate head tuple a
clause match builds, counted before dedup) is what actually shows
semi-naive's avoided redundant work, and it does — substantially, and
differently per shape, never aggregated into one number.
(`measurements/m1-3.9-gate2-headline-summary.json`)

## Lane A interfaces defined, and what they expect

One interface, per §4 item 2's own instruction (magic-set/guard
themselves were never touched — `doc.go` markers only):

```go
// src/transform/transformer.go
type Transformer interface {
    Transform(prog *ast.Program, strata *sema.StratumResult) (*ast.Program, error)
}
```

**What it expects:** `prog` has already passed every sema check this
project runs (decl/type §3.4, allowedness §3.5, source stratification
§3.6) — Lane A's implementation receives only accepted programs. `strata`
is that acceptance's `StratumResult`. **What it must return:** a plain
`*ast.Program` the existing evaluator (`eval.RunNaive`/`RunSemiNaive`)
can run unchanged — either `prog` itself (nothing cleared for TRANSFORM)
or a rewritten program with magic predicates for whichever SCCs the
guard's TRANSFORM/FALLBACK decision cleared. **One documented gap Lane A
needs to know about:** the `strata` argument reflects the *pre*-transform
precedence graph; a program with newly-introduced magic-seed relations
needs its own fresh `sema.CheckStratification` call if strata for the
*output* program is needed — this interface does not solve that,
deliberately (solving it would mean pre-committing Lane A to a specific
re-stratification contract before the guard's own design exists, which
is exactly the kind of resolve-it-for-them overreach `CLAUDE.md`'s Lane A
boundary exists to prevent). `PassThrough` (same file) is the Lane B
placeholder implementation, not wired into the CLI yet (routing through
a no-op changes nothing observable — full reasoning in
`src/transform/DESIGN.md`).

## What a skeptic attacks first

- **§3.3's 20/195 parse rate** is the single number most likely to be
  read as "the parser doesn't work." It isn't a parser defect — 175/195
  failures are independently confirmed genuinely out-of-grammar
  (aggregates, records, `unsigned`/`float` types, the `.input Name()`
  parens idiom) — but a skeptic should ask why the gate's own headline
  framing ("report it as parsed/195") wasn't hit, not just accept the
  attribution at face value.
- **§3.6's stratification-order gate (1/1)** is a single data point. The
  in-grammar corpus's own 20/195 parse rate directly bottlenecks how many
  real negation-bearing files could even be tested — this is a real,
  disclosed statistical thinness, not a hidden one.
- **The C codegen's symbol-ordering-comparison gap** (`<`/`<=`/`>`/`>=`
  on `symbol` columns compare interned ids, not strings) is untested by
  everything this project currently runs through codegen — a skeptic
  should not assume it's fine just because nothing caught it; it hasn't
  been exercised, not verified.
- **`Transformer`'s pass-through is unverified against a real
  transform.** The interface's shape is a bet on what Lane A will need;
  it has only ever been implemented by a function that returns its input
  unchanged. The first real use will be the actual test of whether the
  signature was right.
- **§4 item 3's full-family sweep is capped at 120s per point**, a
  bound chosen for this task alone, not pre-registered — some of the
  family's larger scale points (up to 64M measured tuples at n=8000,
  NIGHT-BATCH-02 T4) may DNF against a tree-walking interpreter with no
  query planning; see that section for the actual outcome.
- **This report's own "M1 headline" framing** (1.00x is expected,
  `DerivationAttempts` is the real signal) is this session's own
  interpretation of an ambiguous instruction ("T_naive vs T_semi-naive,
  exact tuple counts"). A skeptic could read the original instruction as
  expecting the counts to *differ*, and read a perfect 1.00x as evidence
  something is wrong rather than mathematically necessary — the
  reasoning for why it isn't is given in full above and in
  `src/eval/DESIGN.md`, not just asserted.

## §4 item 3 — full-family differential, every pre-registered scale point

`dlc run` (naive) against real Soufflé, ascending per shape, 120s cap per
point (this task's own choice, not previously pre-registered), never
retried at a higher cap on a DNF — this project's established convention.

| shape | matched | DNF | comparable |
|---|---|---|---|
| same_generation_negation | 2/2 | 1 (`d6_b4`) | 2/2 |
| transitive_closure_bound | 3/3 | 1 (`n4000`) | 3/3 |
| ancestor_nonancestor | 3/3 | 1 (`n4000`) | 3/3 |
| reachability_complement | 3/3 | 1 (`n2000`) | 3/3 |
| culprit_cycle | 5/5 | 0 | 5/5 |

**16/16 comparable points matched, zero disagreements, across every
shape in the pre-registered family.** Every DNF is exactly where
expected: `dlc` is a tree-walking interpreter with no query planning or
codegen-level optimization (§3.8's own framing, "correctness before
speed"), and NIGHT-BATCH-02 T4 already measured these shapes reaching
tens of millions of derived tuples at their largest pre-registered
points — a 120s cap was never going to survive that, and didn't, at
exactly the scale point where the curve gets steep for each shape.
`culprit_cycle` (small `T_none` even at its largest point, 84,105 per
NIGHT-BATCH-02 T4) completed all 5 of its own points with room to spare.
(`measurements/m1-4.3-full-family-differential-summary.json`)
