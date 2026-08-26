# NIGHT-BATCH-03 — summary

Date: 2026-08-27. All 11 tasks: **DONE**. No BLOCKED, no PARTIAL. Zero
escalations to `docs/ESCALATIONS.md` — no blocker was hit that couldn't be
resolved within the task's own scope.

## What did not work (first, per protocol)

- **T4's structural classifier scored zero findings everywhere on its
  first draft**, including on `culprit_cycle.dl` itself — a real bug
  (`.decl`/`.input`/`.output` directives don't end in a period, so a
  naive top-level-period split merged each directive with the following
  clause), not a corpus result. Fixed before trusting any count.
- **T7's predicate first draft admitted 43/622 files with only 19/43
  (44%) actually parsing under `dlc`** — investigated every discrepancy
  against real file content and real `dlc` diagnostics, found 3 missing
  categories and 2 bugs in existing ones, converged to 21/622 admitted,
  19/21 (90%) parsing. 2 residual gaps (both deliberately-malformed
  hostile test fixtures) disclosed rather than force-fit into a synthetic
  category.
- **T5's prediction was wrong, for a verified structural reason**: Q11's
  premise misidentified `ancestor_nonancestor.dl`'s actual recursive rule
  shape as first-argument-invariant (confusing it with `p4prime.dl`'s
  `reach`). Built exactly as pre-registered anyway; falsified at all 5
  scale points, answers diverging from the untransformed baseline every
  time.
- **T8 found a real, previously-only-suspected defect**: codegen's
  ordering comparisons on `symbol` columns did in fact compare interned
  ids, not strings — confirmed by temporarily disabling the fix and
  watching all 4 new tests fail with exactly the predicted wrong answer.

## Results, one number per task

| task | result |
|---|---|
| T1 — printer round-trip | 6/6 shapes Souffle-accepted + answer-identical (full 26-case corpus: 17/26 accepted, 17/17 identical among comparable, 0 printer-only failures) |
| T2 — M3 protocol on identity transform | 25/25 scale points, tuple totals + answers identical |
| T3 — M2 acceptance harness | 3/3 known-good pairs reproduce committed T5/T6 numbers exactly |
| T4 — culprit-cycle corpus | 0/817 real-world files match beyond the 1 known file; 6 new constructed programs (5 matching + 1 negative control), 6/6 Souffle-recorded, 0 stratification failures |
| T5 — ancestor seed prediction | Falsified, 5/5 scale points diverge from baseline |
| T6 — `.input`/`.output` parens | kept: total parsed 20→89, negation-bearing parsed 3→18 |
| T7 — corrected predicate | `\|V2\|=21`, cross-check 19/21 (90%) |
| T8 — symbol ordering | real defect found and fixed; 4/4 tests fail before, pass after; Souffle cross-check set-identical |
| T9 — cone metric | validated exactly against `culprit_cycle`'s known `{q,s}`; all 5 shapes reported |
| T10 — provenance backfill | 17 `MEASUREMENTS.md` rows + 3 `DECISIONS.md` entries |
| T11 — de-stale state doc | 1 line replaced, version untouched, no new document |

**Every task produced a number; none was reported as "looks fine" without
one — per protocol, "exercised N cases, zero defects" is stated explicitly
wherever that was the actual outcome** (T1's printer defects: 0; T2's
plumbing deviations: 0; T4's real-world corpus matches beyond the known
file: 0; T9's deviation from the validated `culprit_cycle` pattern: 0).

## What a skeptic attacks first, batch-wide

- T4's and T9's dependency-graph extraction, and T7's grammar predicate,
  all share one underlying structural text-scanning approach
  (`night03_t4_culprit_classifier.py`'s `parse_structure`/`build_graphs`,
  reused by `cone_metric.py`). A bug in that shared parser would propagate
  into three tasks' numbers at once. It was independently validated three
  separate ways (culprit_cycle.dl's exact known finding, the V2⊆V1 subset
  check, and the culprit_cycle cone match) — three independent
  confirmations of the same underlying tool, not three independent tools.
- T5's and T8's fixes both touched already-tested, previously-stable code
  paths (a hand guard file; codegen's constraint emission). Both were
  verified to actually change behavior in the predicted direction
  (T5: confirmed via direct answer-diff; T8: confirmed via
  disable-then-restore) rather than assumed correct from the diagnosis
  alone.
- T6 and T7 together mean the corpus story is now more complex, not
  simpler: `IN_GRAMMAR.txt` (195, the old predicate), `IN_GRAMMAR_V2.txt`
  (21, the corrected predicate), and `dlc`'s own live parser count (89,
  after T6's amendment) are three different numbers answering three
  different questions, and a future report that cites "in-grammar" without
  saying which of the three it means will be ambiguous. Not resolved here
  — flagged as the first thing that needs a naming convention before M2's
  own reports start citing corpus sizes.

## Commits

One per task, `[B][night03][T<n>]`: `df652e1` (T1), `e0d550d` (T2),
`a394848` (T3), `88cfe55` (T4), `caa4c0f` + `5b6a5c7` (T5, pre-registration
committed separately before measurement, per instruction), `43a5874` (T6),
`af0a23c` (T7), `df46050` (T8), `67cba96` (T9), `82e941e` (T10), `0b94c19`
(T11). Plus one housekeeping commit, `2726009`, backfilling `SESSION_LOG.md`
entries for T1-T4 (the log edits themselves landed after those tasks'
main commits).

## Candidate open questions for the human

Not filed directly (per instruction) — listed here for a decision:

1. **Corpus-size naming ambiguity** (see above): `IN_GRAMMAR.txt`
   (195) vs. `IN_GRAMMAR_V2.txt` (21) vs. `dlc`'s live parse count (89,
   post-T6) answer different questions and nothing enforces which one a
   future report means.
2. **T7's `functor_call_bare_relop_operand` regex** is the least precise
   category added this session — no false positive was found in this run,
   but it was not independently stress-tested beyond this corpus.
3. **T9's cone numbers are hypothetical** (a chosen single-relation
   decline per shape), not the output of a real guard decision — M3's
   real guard, once built, may decline a different set entirely.

## Single next action

Per instruction, continuing directly into `docs/m2 m3.md` §§2–9 (Lane A
retired; magic-set transform, guard, and fallback evaluation are now Lane
B) in the same session, no gate/human confirmation required between the
two per the batch's own continuous-run framing.
