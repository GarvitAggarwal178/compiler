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
  `docs/reports/night01-T1-audit.md`. Commit: (pending)
