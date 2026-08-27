# M3.4 — fallback evaluation (§8)

Date: 2026-08-27. Per instruction: "verify this holds before writing
machinery... if the existing SCC-ordered evaluator handles it, say so in
DESIGN.md and write only what is genuinely missing." `src/eval/fallback.go`.

## What did not work

**No corpus program before this task actually produced a genuinely mixed
program.** Every `CULPRIT_CANDIDATES` program (M3.3's gate) either
declines nothing (`cc_edb_negated`) or declines everything (all 6
culprit-cycle-bearing programs) — the culprit SCC always entangles every
IDB relation in those shapes. To exercise fallback evaluation at all, a
new construction was needed: `tests/corpus/CULPRIT_CANDIDATES/
cc_mixed_fallback.dl`, combining a genuine culprit-cycle group (`p`/`q`/`s`,
identical shape to `culprit_cycle.dl`) with a genuine safe group
(`ancestor`/`nonancestor`, identical shape to `ancestor_nonancestor.dl`)
feeding one query through a conjunction relation (`combo`). Confirmed
directly (`TestDecideProducesGenuineMixedProgram`) that the guard declines
only `{p,q,s}` and keeps `ancestor`/`nonancestor`/`combo`/`ans` transformed
— the first and only program in this corpus to produce a real mix.

## What was verified, and what was (and wasn't) written

**Verified: the existing evaluator needs no changes.** A mixed program is
an ordinary `*ast.Program` — nothing in its shape distinguishes a declined
predicate's original clause from any other clause. `sema.
CheckStratification`, re-run on the mixed program (never the pre-transform
`StratumResult` — the one piece of wiring genuinely required), yields a
valid SCC order over the mixed program's actual dependency graph, and
`RunNaive`/`RunSemiNaive` already process one SCC at a time in that order.
A declined predicate's rule reads another relation's full extent
correctly for the same reason any two relations in any ordinary program
do.

**Written:** `cmd/dlc`'s `run`/`run-seminaive` gained an optional
`--transformer=<name>` flag (shared parsing with `emit`'s existing flag),
and now apply the named `Transformer` + re-derive stratification before
evaluating — this is real, necessary CLI wiring (the task's own framing:
"fallback evaluation wiring"), not a second evaluator. `src/eval/
fallback.go` documents the verification; it contains no new evaluation
logic, per instruction ("do not build a second evaluation path to solve a
problem the first one already solves").

## Three independent confirmations

Provenance: `measurements/m3-4-fallback/`. All three answer `2` for
`cc_mixed_fallback.dl`'s `ans` relation on the shared hand-built fixture
(`fixtures/culprit_candidates/cc_mixed_fallback/`):

| method | engine | program | `ans` |
|---|---|---|---|
| `TestFallbackEvaluationMatchesDlcOwnEvaluator` | `dlc`'s own `RunNaive` | mixed (guard-decided) | `2` |
| `dlc run --transformer=guarded` (CLI, end-to-end) | `dlc`'s own `RunNaive` | mixed (guard-decided) | `2` |
| `souffle` (real) | Soufflé | **untransformed original** | `2` |
| `souffle` (real) | Soufflé | `dlc emit --transformer=guarded`'s printed output | `2` |

The Go unit test additionally confirms `dlc`'s own evaluator agrees with
itself between the untransformed original and the mixed program directly
(not only via the CLI), and the two Soufflé cross-checks confirm the
answer is correct in an absolute sense, not merely self-consistent within
`dlc`.

## What a skeptic attacks first

- The mixed test case is hand-constructed specifically to produce a
  non-trivial mix — no naturally-occurring program in this project's
  corpus does. This proves the mechanism *can* work correctly on a mix,
  not that mixed programs are common or that a more complex mix (e.g.
  three or more independently-declined groups, or a declined predicate
  feeding into ANOTHER declined predicate through a still-transformed
  intermediary) would behave identically — only the one mix shape
  actually available was tested.
- The fixture is small (4–5 facts per relation) and hand-picked to
  produce a non-empty, easily-inspectable answer — not a scale-relevant
  test. Its job is correctness verification, not measurement (M3.5's job).
- `runRun`'s new stratification-failure guard ("transformer produced an
  unstratifiable program... should be impossible for `--transformer=
  guarded`") has never actually fired in this session, by construction —
  `guard.Transformer` is specified to always return a stratifiable result,
  and every gate run so far confirms it does, but the guard code path
  itself is unexercised defensive code, not something a test drove into
  failure and back.

## Verdict

**M3.4: DONE.** Verification-first, as instructed: confirmed the existing
evaluator needs no new machinery, wrote only the CLI wiring the task's own
framing anticipated, and constructed the one test case in this project's
corpus that actually exercises a non-trivial mix. Three independent
engines/paths agree on `cc_mixed_fallback.dl`'s answer. Proceeding to
M3.5 (the headline run).
