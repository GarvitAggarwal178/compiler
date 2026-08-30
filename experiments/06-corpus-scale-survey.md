# NIGHT-BATCH-01 — T2: corpus viability, untransformed only

Date: 2026-08-20. Outcome: **done**, with an unresolved input gap disclosed rather
than papered over — see below and `docs/ESCALATIONS.md`.

## The missing document

`docs/phase0.7-corpus-viability.md`, which T2 is specified "exactly" against, does
not exist in this repository. Full escalation entry: `docs/ESCALATIONS.md`
(2026-08-20, T2/T3). This report uses only what the night-batch directive itself
specifies inline (the per-program field table in its own §T2, and the two
thresholds named in its §0.2.3: `T_none ≥ 1,000`, floor of 8) and does **not**
attempt to reconstruct "the four numbers from Phase 0.7 §3" — that framing is not
recoverable without the missing file.

## Method

`harness/night01_t2_corpus.py`, all 36 `tests/corpus/PREREGISTERED.txt` programs,
untransformed (no `--magic-transform`), each under the global caps (300s timeout,
8GB address-space limit). Reused `corpus_predicate.check_program` (read-only import,
not modified) for `seedable` and `negated_idb_count` — those are exactly conditions
2 and 1 of the same predicate that built the corpus. `T_none` via the new
`harness/tuple_report.py` (both copy conventions; see T7). Out-of-grammar feature
counts are a fresh mechanical regex scan (functors, aggregates, components,
records, choice, ADTs) — heuristic, not a parse; see caveats below.

**One bug found and fixed before results were trusted:** the first run wrote every
program's profile log to a single relative path (`prof.log` resolved against the
script's own working directory, not each test's `-D` directory), so all 36 runs
silently overwrote one shared file and every `T_none` came back null. Fixed by
using the same `cwd=<per-test workdir>` convention `probe0.py` has used since
Phase 0 (`-p` is a relative path resolved against the process's cwd, not against
`-D`). Full re-run after the fix; the table below is from the corrected run.

## Status

| Status | Count |
|---|---|
| `ok` | 31 |
| `error:returncode-1` (compiler-rejected) | 4 |
| `error:returncode--6` (crash, `SIGABRT`) | 1 |
| `DNF` (cap) | 0 |

No DNFs — every program in this pre-registered set is small enough to run well
inside 300s/8GB untransformed.

**The 5 failures, read (not investigated further — out of T2's scope):**

- `interface/lattice2` — crashes: "cannot find user-defined operator `lub`" — this
  test exercises Soufflé's component/lattice system standalone, outside the
  component instantiation context it needs.
- `semantic/plan1`, `semantic/plan2` — both use explicit `.plan` execution-order
  directives and are testing Soufflé's *own* plan-validation error paths (a
  deliberately invalid plan). Negative tests, not executable programs.
- `semantic/rel_stratification` — `a(X) :- a(X), !a(X).` is deliberately
  unstratifiable; this is a negative test for Soufflé's stratifier, not a program
  meant to run.
- `semantic/witness_check` — deliberately triggers an aggregate witness-grounding
  error; another negative test.

**Observation for the record, not acted on:** the mechanical predicate
(`corpus_predicate.py`) has no way to distinguish "a real executable program" from
"a program deliberately designed to be rejected by Soufflé" — both can carry a
negated IDB literal and a constant-bearing `.output` relation. 4 of 36 (11%) of the
pre-registered set turned out to be the latter. Not fixed (prohibition #2); logged
to `docs/OPEN_QUESTIONS.md`.

## Full per-program table

| Program | Status | `T_none` (excl=incl copy) | negated IDB literals | out-of-grammar features | fact rows (input) |
|---|---|---|---|---|---|
| evaluation/bad_regex | ok | 1 | 1 | 0 | 2 |
| evaluation/components | ok | 35 | 2 | 9 | 0 |
| evaluation/inline_negation1 | ok | 3 | 2 | 0 | 0 |
| evaluation/magic_2sat | ok | **1,338** | 1 | 0 | 14 |
| evaluation/magic_components | ok | 114 | 1 | 9 | 0 |
| evaluation/neg1 | ok | 6 | 1 | 0 | 0 |
| evaluation/neg2 | ok | 9 | 1 | 0 | 0 |
| evaluation/neg3 | ok | 14 | 1 | 0 | 0 |
| evaluation/neg4 | ok | 9 | 1 | 0 | 0 |
| evaluation/neg5 | ok | 15 | 1 | 0 | 0 |
| evaluation/set_ops | ok | 40 | 13 | 0 | 0 |
| example/2sat | ok | 200 | 1 | 0 | 14 |
| example/array | ok | 50 | 1 | 0 | 7 |
| example/cellular_automata | ok | 165 | 1 | 0 | 3 |
| example/dfa_parse | ok | 134 | 1 | 0 | 41 |
| example/earley | ok | 28 | 2 | 0 | 13 |
| example/flights | ok | 20 | 1 | 0 | 0 |
| example/fractional_knapsack | ok | 14 | 2 | 3 | 4 |
| example/game2 | ok | 18 | 2 | 0 | 0 |
| example/grid | ok | 99 | 1 | 0 | 4 |
| example/minesweeper | ok | **1,100** | 1 | 0 | 10 |
| example/orbits | ok | 7 | 2 | 0 | 0 |
| example/orbits1 | ok | 7 | 1 | 0 | 0 |
| example/tic-tac-toe | ok | **276,168** | 3 | 0 | 0 |
| example/topological_ordering | ok | 344 | 3 | 0 | 31 |
| interface/lattice2 | error:crash | — | — | — | — |
| provenance/constraints | ok | 4 | 1 | 0 | 3 |
| semantic/eqrel_tests | ok | 344 | 1 | 0 | 0 |
| semantic/plan1 | error:rejected | — | — | — | — |
| semantic/plan2 | error:rejected | — | — | — | — |
| semantic/rel_nullary | ok | 23 | 2 | 0 | 2 |
| semantic/rel_stratification | error:rejected | — | — | — | — |
| semantic/subsumption_multiple_rules | ok | 31 | 1 | 0 | 0 |
| semantic/witness_check | error:rejected | — | — | — | — |
| swig/java/flights | ok | 20 | 1 | 0 | 0 |
| swig/python/flights | ok | 20 | 1 | 0 | 0 |

Raw per-program provenance: `measurements/night01-t2/<path-with-slashes-as-__>/`.
Machine-readable: `measurements/night01-t2/summary.json`.

## The numbers the batch text asks for directly

- **`T_none ≥ 1,000`:** 3 of 31 successfully-run programs
  (`evaluation/magic_2sat`=1,338; `example/minesweeper`=1,100;
  `example/tic-tac-toe`=276,168).
- **Floor of 8: not met.** 3 < 8.
- **`seedable`:** 31 of 31 successfully-run programs — true by construction (every
  pre-registered program already satisfies this as inclusion condition 2; this
  column re-confirms it, it does not discover anything new).

**Not acted on, per NIGHT-BATCH-01 §0.2.4 and the task's own instruction ("do not
act on the answer either way").** No corpus substitution, no threshold change, no
new fixtures. That decision is `docs/phase0.7-corpus-viability.md` §2.3's, i.e. the
human's, once the missing document exists or is replaced.

T3 proceeds against the 31 `ok` programs (all seedable by construction).
