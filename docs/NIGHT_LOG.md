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
  `docs/reports/night01-T3-envelope.md`. Commit: (pending)
