# NIGHT-BATCH-01 — T1: provenance and determinism audit

Date: 2026-08-20. No abort condition fired (no non-reproducing integer, no
golden-file-not-from-Soufflé). Outcome: **done**.

## Provenance gaps found and backfilled

Inventory of every measurement ID referenced in `docs/MEASUREMENTS.md`
(`harness/night01_t1_audit.py`) found 8 real provenance gaps out of 27 distinct IDs
(the rest were false positives from table-cell parsing — brace-notation summarizing
multiple already-complete IDs, and unrelated table cells like bare `P1`/`P2`/`P3`).

| ID | Gap | Fix |
|---|---|---|
| `probe0-p1-diff`, `probe0-p1-diff-q` | No measurement directory — original Phase 0 diff ran via ad hoc Bash, never captured | Re-ran via `probe0.run_cmd`, full provenance now present |
| `probe0-p2-diff` | Same | Same |
| `probe0-p3-diff` | Same | Same |
| `probe0-p1-fixture-bfs-check` | `docs/MEASUREMENTS.md` cited `python3 harness/probe0.py` as the source of `reachable_from_1=50`, but no dedicated directory captured the verification itself (only `fixtures/p1/meta.json`) | Added `harness/night01_bfs_check.py` (independent BFS, not Soufflé, not `dlc`), ran and captured |
| `probe0.5-p2-bfs-check` | Same pattern for the P2 BFS cross-check from Phase 0.5 | Same script, captured |
| `probe0.6-p4prime-diff-vs-p4`, `-vs-p2` | Referenced in `docs/MEASUREMENTS.md` as `probe0.6-p4prime-diff`, never actually run through `run_cmd` | Re-ran both diffs individually with full provenance |
| `probe0.5-p1prime-diff` | Directory existed (cmd.txt + stdout.txt only) — `stderr.txt`/`env.txt`/`meta.json` missing | Backfilled; also **found and corrected a hygiene issue**: this directory bundles two `diff -q` commands under one ID |
| `probe0.5-p4-diff` | Same partial-provenance pattern | Same fix. **Also found:** its `stderr.txt` previously contained hand-typed commentary ("both diffs ran clean...") instead of genuine captured stderr from the Phase 0.5 session. Corrected to the actual (empty) stderr. |

All 8 gaps are Lane B harness/provenance hygiene, backfilled by re-running the real,
already-reported commands — no numbers were invented, all match what was already
documented.

## Reproducibility: stdout

`harness/night01_t1_reproduce.py` re-ran every measurement directory's recorded
command(s) (61 directories after backfill) under the batch's global caps (300s
timeout, 8GB address-space limit) and compared fresh stdout against committed
`stdout.txt` in memory, without touching the committed files.

**Result: 61/61 reproduced.** 0 mismatches, 0 DNFs.

## Reproducibility: filesystem side effects

Re-running `souffle`/`souffleprof` commands necessarily regenerates their own
`-D`-directory output files in place. `git status`/`git diff` after the sweep is the
check for those.

- **19 `prof*.log` files showed byte-level diffs.** Inspected: every diff is confined
  to Soufflé's own profiler instrumentation fields (`start`/`end` timestamps,
  `maxRSS`) — inherently non-deterministic wall-clock/memory measurements that were
  never used as data (per-relation `num-tuples`, the only field this project reads,
  verified identical across every relation in a full-file spot check on
  `probe0.6-p4prime-run/prof.log`). Reverted (`git checkout`) to avoid leaving noise
  diffs in the tree — nothing here was ever a claim, so nothing was lost by
  reverting.
- **`tests/corpus/detail.json` showed a diff** after re-running the exact recorded
  `probe0.6-q5-corpus` command. Root cause: `corpus_predicate.py`'s
  `check_program()` iterates `output_names` as a Python `set`; when more than one
  `.output` relation independently qualifies as "has a literal", *which one* gets
  recorded in the diagnostic field `matched_output_relation` depends on Python's
  randomized string-hash seed, which varies per process. **`tests/corpus/
  PREREGISTERED.txt` and the included-count (36) are unaffected** — the actual
  predicate decision only needs "at least one" match, which is order-independent;
  only an unreported diagnostic string flaps. Reverted `detail.json` to the
  committed version. **Not fixed** — hard prohibition #2 ("do not edit ... the
  predicate that produced it") applies verbatim to `corpus_predicate.py`, so the root
  cause (seed the hash, or sort `output_names` before iterating) is logged here for
  daylight, not touched tonight.
- **All other regenerated files** (`.csv` outputs, `q.csv`, `out.csv`, `ans.csv`
  etc.) showed **zero diffs** — every derived-tuple answer relation reproduces
  byte-identically, which is the number that actually matters for every headline
  claim made so far.

## Fixtures

`fixtures/p1/edge.facts`, `fixtures/p2/edge.facts`, `fixtures/p2/node.facts`,
`fixtures/p3/base.facts`, `fixtures/p3/e.facts` regenerated from their recorded
seeds via `probe0.build_p{1,2,3}_fixture()` and SHA-256-compared against the
committed files: **identical, all five.**

## Golden files

`tests/golden/` contains **0 files.** Nothing to check — no golden file has been
generated by anything, Soufflé or otherwise, yet. This is expected (M1 hasn't
started producing golden-comparable output); noted, not a finding.

## Summary table

| Check | Result |
|---|---|
| Provenance gaps found | 8 (all backfilled) |
| Stdout reproducibility | 61/61 reproduced |
| Filesystem side-effect reproducibility (derived data) | 100% (0 diffs on any `.csv` answer relation) |
| Filesystem side-effect reproducibility (profiler timing/memory) | Non-deterministic by design, never used as data, reverted |
| Non-determinism found in Lane B tooling | 1 (`corpus_predicate.py`'s diagnostic field only; decision unaffected; cannot fix tonight, prohibition #2) |
| Fixture SHA-256 | 5/5 identical |
| Golden files | 0 exist |
| **Abort condition (§ T1)** | **Not triggered** |

Continue to T2.
