# NIGHT-BATCH-01 — T7: M1 harness build-out

Date: 2026-08-20. Outcome: **done**, no prerequisite, Lane B only — nothing under
`src/` touched or created.

## What exists

| Item | File | Status |
|---|---|---|
| 1. Differential runner | `harness/differential.py` | Working. `run_souffle` runs for real; `compare()` does set-equality per relation (CLAUDE.md §6) and reports the symmetric difference, not a boolean. |
| 2. Golden generation + guard | `harness/golden.py` | Working. `generate_golden()` refuses (raises `GoldenGuardError`) unless `Path(cmd[0]).name == "souffle"`, no bypass flag. |
| 2b. Guard test | `harness/test_golden_guard.py` | 3/3 passing (no external test framework — none installed, none otherwise depended on). |
| 3. Rejection-test scaffolding | `tests/rejection/{arity,type,allowedness,stratification}.py` | 12 cases (3 per ground), each a full `.dl` source + `expected_ground` + `expected_diagnosis` as data. |
| 3b. Rejection-scaffolding runner | `harness/run_rejection_tests.py` | 12/12 correctly report `not_implemented` — verified no case vacuously "passes." |
| 4. Fixture generator library | `harness/fixtures_lib.py` | Both shapes (`gen_core_rest_graph` — P1's bounded-reachability construction, `gen_random_graph` — P2's plain-random construction) factored out and parameterized; validated byte-identical against the original committed fixtures before use (T6). |
| 5. Generalized tuple extraction | `harness/tuple_report.py` | Both copy conventions, `@neglabel.` sum (`E_recoverable`), correct `COPY_T` vs. genuine-rederivation discrimination (a real bug here was caught and fixed against known ground truth before T2 trusted any number). |

## What is stubbed, and what M1 must supply

**`differential.py`'s `run_dlc()`** is the one deliberate stub. Right now it
unconditionally returns `EngineResult(status="not_implemented")`. M1 replaces its
body with a real subprocess invocation of the built `dlc` binary — everything
downstream (`compare()`, the CLI in `main()`, the rejection-test runner) is already
correct for that day and does not need to change. The expected shape M1's `run_dlc`
needs to produce: `EngineResult(engine="dlc", status="ok"|"error:<msg>",
output_relations={name: sorted_lines})` for the acceptance path, matching
`run_souffle`'s shape exactly so `compare()` keeps working unmodified.

**Rejection-test cases** are Lane B data (test bodies), not a checker. Running them
today just proves the scaffolding is wired honestly — every case reports
`not_implemented`, none silently pass, none fabricate a verdict. When M1's
decl/arity check, type check, allowedness check, and stratification check exist,
`run_rejection_tests.py`'s assertion (`case["expected_ground"] == ground`) already
holds; the missing piece is asserting `result.status` reports the *right* rejection
reason once `run_dlc` is real, which needs the real diagnostic format M1 decides on
— not designed here, since that's a Lane A interface decision.

## Validation performed tonight, not deferred

- `fixtures_lib.py`'s two generators reproduce the original committed P1/P2
  fixtures byte-for-byte (checked before use in T6, not assumed).
- `tuple_report.py`'s copy-detection was validated against all four known Phase
  0/0.5 ground-truth numbers (P1 on/off, P2 on, P3 on, P1' on) before T2 trusted it
  — this caught and fixed a real false-positive bug (`docs/reports/
  night01-T2-corpus.md`).
- Every rejection-test case's representative program was independently run through
  real Soufflé tonight and confirmed rejected for the matching reason (arity, type,
  ungrounded-variable, stratification) — evidence the hand-written malformed
  programs are genuinely malformed, not typos that happen to parse wrong.
- `differential.py` had the same relative-path/wrong-cwd bug T2 found and fixed
  earlier tonight; caught by a smoke test before this report was written, not left
  latent for M1 to discover.

## What did not work

Two bugs, both caught by validating against known-good data before trusting
anything, not by inspection:

1. `tuple_report.py`'s first copy-detection regex matched any single-literal rule
   body, not just Soufflé's synthetic `@`-prefixed copy-out pattern — flagged
   `@neglabel.s`'s genuine re-derivation as a spurious "copy" and undercounted P3's
   `T_magic` by 19. Fixed by requiring the body atom to be `@`-prefixed (T2).
2. `differential.py`'s `run_souffle` resolved the `.dl` and facts-dir paths
   relative to the subprocess's `cwd` (the per-run workdir), not the caller's —
   identical bug class to the one found in T2's harness, this time caught by a
   smoke test immediately after writing the module, before it was used anywhere
   real.

## Not built tonight, out of T7's scope

- No real M1 test *content* beyond the 12 rejection cases and the differential
  runner's plumbing — T5's `tests/corpus/IN_GRAMMAR.txt` (195 files) is the actual
  differential-test pool once `run_dlc` is real; nothing here decides which of
  those 195 M1 should prioritize.
- No CI wiring — `tools/` is empty; running these scripts is manual tonight.
