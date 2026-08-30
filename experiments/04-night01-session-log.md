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
- T5 start -- hand-guarded transforms (T_guard, the headline), cap 150 min,
  prerequisite T4 (passed). transitive_closure_bound excluded (E_recoverable=0).
- T5 end -- outcome: done, complete. 4 shapes guarded, 20 scale points, no
  DNF, no abort -- answers identical to T4 baseline everywhere. Ratios:
  same_generation_negation 68x-4,371x (explained: query root=0 has no
  parent in this fixture, verified independently against T4's own
  q_notsg.csv counts); ancestor_nonancestor 4x-62x (most modest, grows
  linearly not superlinearly); culprit_cycle ~1.0x (no real contribution --
  q/s left deliberately unrestricted); reachability_complement 157x-4,244x
  (matches T6 exactly, confirms p4prime.dl port is correct). Found and
  discarded (not fixed up) a general-adornment derivation for culprit_cycle
  that Soufflé rejected for cyclic negation -- this recreates P5's own
  reason for existing. Also found a harness gotcha: Soufflé can exit 0 on a
  stratification error (fixed the runner to also check stderr for
  Error:). Report: docs/reports/night02-T5-guarded.md.
- T6 start -- cross-shape scaling analysis, cap 60 min, prerequisite T5
  (passed). Answers whether P2's Theta(n) ratio growth (night01 T6) was
  the mechanism or the shape.
- T6 end -- outcome: done, complete. No new Soufflé runs -- pairwise
  log-log slopes over already-committed T3/T4/T5 numbers. Found: 3/4
  guarded shapes (same_generation_negation, ancestor_nonancestor,
  reachability_complement) share growth CLASS (T_none/T_souffle Theta(n^2),
  T_guard Theta(n), ratio Theta(n)) despite different recursion shapes --
  but magnitudes differ by up to 2 orders of magnitude at comparable n
  (not explained here). culprit_cycle disagrees on class entirely: flat
  ~1.0x ratio, no growth, and its 5 points don't admit a clean power-law
  fit at all (reported noisy, not smoothed). transitive_closure_bound
  (no guard) has T_souffle flat Theta(1). Report:
  docs/reports/night02-T6-scaling-crossshape.md.
- T2 start -- hostile source corpus, cap 90 min, independent (no Lane A
  dependency).
- T2 end -- outcome: done. 39 .dl files under tests/hostile/, 31 accept,
  8 reject. All 14 precedence/associativity files accept and match the
  grammar-correct reading wherever numerically distinguishable (verified
  against actual output values, not just parse acceptance). Found:
  identifiers are ASCII-only (unicode rejected at the lexer); an
  unterminated block comment silently swallows to EOF (accepts, but
  produces 0 rules -- worth a rejection-corpus case despite Soufflé not
  rejecting it); .decl with arity 0 is accepted by full Soufflé though
  out of this project's own restricted grammar (not a contradiction --
  blueprint's grammar is a deliberate subset). Fixed a
  probe0.run_cmd UnicodeDecodeError crash on non-UTF-8 stderr (affects
  all harness scripts using run_cmd, backward compatible). Caught and
  fixed a harness gap that had mis-classified 2 files as reject for a
  missing-facts-file reason rather than a language reason; one is now
  correctly accept, one remains genuinely inconclusive (facts filename
  itself too long for the filesystem). Report:
  docs/reports/night02-T2-hostile.md.
- T8 start -- grammar usage census over IN_GRAMMAR.txt, cap 60 min,
  independent.
- T8 end -- outcome: done. 195/195 files analyzed via a token-aware static
  scan (not a full parser, disclosed). Arity median 1 max 14; body-length
  median 2 max 12; expr-depth median 1 max 3 (real programs are not
  deeply nested); 69 negations across 28 files, mostly middle/last body
  position; 688 wildcards (687 body, 1 head, unresolved to a specific
  file); 61/954 relations directly recursive (lower bound, mutual
  recursion not detected by this scan, disclosed). FOUND AND FLAGGED FOR
  THE HUMAN: 11 of the 195 in-grammar files declare zero-arity
  relations, which blueprint section 4's own grammar does not admit -- a
  real gap in night01 T5's mechanical in-grammar predicate. Not fixed
  (out of this task's scope; IN_GRAMMAR.txt is not under night-02's
  corpus-file prohibitions but census != corpus maintenance). Report:
  docs/reports/night02-T8-grammar-census.md.
- CORRECTION (found during T9, logged before T9's own start/end lines):
  T5's report and NIGHT_LOG entry claimed 'Soufflé exited 0 on a
  stratification error' for culprit_cycle_unsafe_cyclic.dl. Re-run 5 times
  under T9 with output redirected to files (not streamed interactively) --
  every re-run returned rc=1, consistently. The original rc=0 reading was
  a Bash-tool/wsl.exe bridge artifact (racing on live-streamed multi-command
  terminal output), not real Soufflé behavior. Corrected in
  docs/reports/night02-T5-guarded.md (marked, not silently edited) and in
  the two harness files whose comments cited it
  (harness/night02_t5_guarded.py, harness/night02_t2_hostile.py). No
  T_guard number or any other measurement in T5 was affected -- this was
  a side note about the rejection mechanism, not about T_guard itself.
  Full account: docs/reports/night02-T9-diagnostics.md.
- T9 start -- Soufflé diagnostic catalogue, cap 45 min, independent.
- T9 end -- outcome: done. Catalogued all 7 error classes (ungrounded
  variable, arity mismatch, type mismatch [2 message shapes: inference vs
  literal], undeclared relation, unstratifiable negation, syntax error,
  duplicate declaration), each with exact first-line message and
  line/col presence (all 7 carry location info). Cross-checked all 13
  existing tests/rejection/ cases against real Soufflé: 13/13 consistent
  with their expected_ground, no corpus revision needed. Found and
  corrected (logged above, separate commit) a false claim in T5 about
  Soufflé exiting 0 on a stratification error -- was a Bash-tool/wsl.exe
  bridge artifact, root-caused and fixed. Report:
  docs/reports/night02-T9-diagnostics.md.
- T7 start -- P5 inlining prerequisite, cap 45 min, independent, last task
  in the queue.
- T7 end -- outcome: done. Check 1 (|p|>|e|, dead-rule check, done first
  per instruction): fires at 4/5 T4 scale points, dead only at n=20
  (flagged as likely fixture-sparsity at that size, not a structural
  problem -- caught by checking the other 4 points immediately after an
  initial n=20-only check looked alarming). Check 2: q survives Souffle's
  inliner, confirmed -- --inline-exclude=q changed nothing at all
  (byte-identical output with/without), meaning q was never an inlining
  candidate to begin with (self-recursive), unlike P3's non-recursive q.
  Check 3: culprit cycle confirmed -- q and s computed at full
  untransformed size under the automatic transform (exact match to their
  untransformed totals), only p genuinely restricted -- structurally the
  same restriction T5's hand-guard makes on purpose, explaining T5's
  ~1.0x contribution finding for this shape. Report:
  docs/reports/night02-T7-p5-precheck.md.
- BATCH END -- all 9 tasks (T1,T3,T4,T5,T6,T2,T8,T9,T7) reached a defined
  outcome (done, no aborts, no DNFs anywhere in the batch). src/ never
  touched. One correction issued and logged (T5's stratification rc=0
  claim, retracted, separate commit). No ESCALATIONS.md entries were
  needed -- no CLAUDE.md section 5 STOP condition actually fired this
  batch (the rc=0 anomaly was investigated and resolved as a tooling
  artifact, not a Soufflé nondeterminism or oracle-disagreement event).
  Morning summary: docs/reports/night02-summary.md. CLAUDE.md section 5
  reverts to normal STOP-and-wait semantics as of this line, per the
  batch directive section 0.
