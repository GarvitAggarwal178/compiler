# Night log — NIGHT-BATCH-01

Append-only. One line per task start/end: task ID, outcome, commit SHA.

- T0 start — batch setup (ESCALATIONS.md, NIGHT_LOG.md created; prerequisite check)
- T0 note — `docs/phase0.7-corpus-viability.md` (referenced by T2/T3) does not exist
  in the repo. Handling documented at T2 start below.
- T1 start — provenance and determinism audit
- T1 end — outcome: done. 8 provenance gaps found and backfilled; 61/61 measurement
  dirs reproduce stdout byte-identically; all derived-tuple `.csv` outputs
  byte-identical on re-run; 1 non-critical non-determinism found in
  `corpus_predicate.py`'s diagnostic field (decision unaffected, not fixed per
  prohibition #2). No abort condition fired. Report:
  `docs/reports/night01-T1-audit.md`. Commit: a994a10
- T2 start — corpus viability, untransformed only, 36 pre-registered programs.
  `docs/phase0.7-corpus-viability.md` missing — escalated (`docs/ESCALATIONS.md`),
  proceeding on the batch text's own inline spec, not aborting (core measurement is
  fully determined without the missing file).
- T2 end — outcome: done. 31/36 ran ok, 4 rejected by Soufflé (deliberate negative
  tests), 1 crashed (component feature used standalone), 0 DNF. Found and fixed a
  real bug in the new harness (relative `-p` path resolved against the wrong cwd,
  silently overwrote one shared log across all 36 runs) before trusting any number.
  `T_none >= 1,000`: 3/31. Floor of 8: not met. Not acted on per instruction. Report:
  `docs/reports/night01-T2-corpus.md`. Commit: b36dd28
- T3 start — recoverable-envelope sweep, 31 T2-ok+seedable candidates.
- T3 end — outcome: **aborted mid-sweep** at `semantic/subsumption_multiple_rules`
  (real set divergence between configs, post-fix). First alarm (`example/orbits1`)
  was a false positive from a harness bug (byte-diff instead of sorted set-equality
  per CLAUDE.md §6) -- found, fixed, full sweep re-run before trusting any result.
  27/31 clean. 1 crash (`tic-tac-toe`, `std::bad_alloc`). 1 diverged (escalated,
  not investigated further). 2 never reached (queue truncated). `E_recoverable=0`
  on 10/27 (37%). Distribution and full table:
  `docs/reports/night01-T3-envelope.md`. Commit: f21f200
- T4 start — whole-tree exploratory sweep (EXPLORATORY, not reportable as a result).
- T4 end — outcome: done. Fast pass (no execution) covered the full tree: 612
  `.dl`-bearing dirs, 107 (17.5%) with negated IDB literals, 242 (39.5%) seedable,
  36 both (= the pre-registered corpus, confirms internal consistency). Slow pass
  (107 negated-IDB candidates, capped at 150 -- didn't need the cap) under
  `--magic-transform=*`: 86 ok, 21 errored/crashed/DNF. `@neglabel.` produced on
  57/86 (66%). `E_recoverable=0` on 29/86 (34%, agrees with T3's 37% on the smaller
  pre-registered subset). Found and fixed two harness bugs before trusting results
  (duplicated-candidate-set bug, missing-key crash in tuple_report.py). Report:
  `docs/reports/night01-T4-exploratory.md`. Commit: 83fe5f4
- T5 start — grammar coverage census, no prerequisite, full tree.
- T5 end — outcome: done. 622 `.dl` files scanned, 195 (31.4%) fully in-grammar
  against blueprint §4. `.type` declarations are the dominant exclusion factor
  (68.1% of out-of-grammar files, ~4x the next most common feature) — even the
  mildest common pattern (a plain subtype alias) is enough to exclude a program.
  Wrote `tests/corpus/IN_GRAMMAR.txt` (195 files, explicitly NOT the pre-registered
  corpus, header says so). Report: `docs/reports/night01-T5-grammar.md`. Commit:
  70f79a0
- T6 start — P2 scaling sweep, three columns, prerequisite T1 (passed).
- T6 end — outcome: done, complete. All 6 sizes (250..8000) ran clean, no DNF, no
  abort. `T_souffle/T_guard` grows from 157x at n=250 to 4,244x at n=8,000 --
  roughly doubling with n, consistent with linear growth in the ratio (T_none
  quadratic, T_guard linear, T_souffle tracks T_none). Reported as integers, not
  asserted as a general theorem (n=6 points, one fixture shape). Built
  `harness/fixtures_lib.py` (T7 item 4 deliverable), validated byte-identical
  against the original committed P1/P2 fixtures before use. Report:
  `docs/reports/night01-T6-scaling.md`. Commit: b369cbd
- T7 start — M1 harness build-out, no prerequisite, Lane B only.
- T7 end — outcome: done. Delivered all 5 items: differential runner
  (`harness/differential.py`, dlc stubbed to always report not_implemented),
  golden-generation guard + 3/3 passing tests (`harness/golden.py`,
  `harness/test_golden_guard.py`), 12 rejection-test cases across 4 grounds
  (`tests/rejection/`, each independently cross-checked against real Soufflé
  tonight and confirmed rejected for the matching reason), fixture generator
  library and generalized tuple extraction (already delivered in T6/T2, formally
  written up here). Two bugs found and fixed by validating against known-good
  data before trusting anything (a tuple_report.py false-positive, a
  differential.py relative-path bug — same class as T2's). Nothing under `src/`
  touched. Report: `docs/reports/night01-T7-harness.md`. Commit: 65b8f9b
- T8 start — Q3 literature check, web access available.
- T8 end — outcome: done. Citation confirmed (Tekle & Liu, ICLP 2019, arXiv
  1909.08246). No downloadable/installable artifact found after two targeted
  searches and one page fetch. Filter 1 verdict unchanged (cite-and-avoid). No
  Filter-1-event escalation needed (only required if an artifact was found).
  Recorded: `docs/OPEN_QUESTIONS.md` (Q3). Commit: 002180f
- BATCH END — all 8 tasks reached a defined outcome (done/aborted/exploratory-cap).
  `src/` never touched. 2 escalations logged in full. Morning summary:
  `docs/reports/night01-summary.md`. CLAUDE.md §5 reverts to normal STOP-and-wait
  semantics as of this line, per the batch directive §0.

## NIGHT-BATCH-02 -- 2026-08-23

- T1 start -- allowedness probes h-o, cap 30 min, run first.
- T1 end -- outcome: done. 8/8 ran, no cap needed. Results: accept i,k,l,n,o;
  reject h,j,m. h is the load-bearing case: p(X):-q(X),Y>3. rejects with
  'Ungrounded variable Y' even though Y is not in the head -- allowedness
  quantifies over body-only variables too, not just head variables. No
  definition proposed (human's decision). Report:
  docs/reports/night02-T1-allowedness.md.
- T3 start -- benchmark family fixture generation, cap 60 min, prerequisite none.
- T3 end -- outcome: done. 38 .facts files generated across 4 shapes (
  reachability_complement excluded, reuses T6's existing fixtures per
  SCALE_POINTS.json's own note). Idempotent: 38/38 SHA-256 identical across
  2 runs. transitive_closure_bound/ancestor_nonancestor reachable-from-1 =
  50 at every n, confirming core_size=50 as designed. Caught and fixed one
  bug before trusting output: same_generation_negation's reachable-from-0
  walk used the wrong edge direction. Report:
  docs/reports/night02-T3-fixtures.md.
- T4 start -- baseline sweep (T_none/T_souffle/E_recoverable), cap 120 min,
  prerequisite T3 (passed).
- T4 end -- outcome: done, complete. All 5 shapes x their scale points (24
  points, 48 Soufflé runs) completed clean: no DNF, no cap fired, no
  abort, answers identical at every point. transitive_closure_bound has
  E_recoverable=0 at every point (positive fragment, no negation --
  reported as a null result per instruction). reachability_complement
  re-run (not regenerated) against its own family .dl file for consistent
  instrumentation; matches T6's committed numbers exactly at all 6 points,
  confirming reachability_complement.dl and p2.dl are equivalent by
  measurement, not just inspection. excl-copy/incl-copy never diverged
  (no COPY_T anywhere this sweep) -- null result, reported not omitted.
  Report: docs/reports/night02-T4-baseline.md.
