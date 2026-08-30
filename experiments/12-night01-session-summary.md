# NIGHT-BATCH-01 — morning summary

Date: 2026-08-20. All 8 tasks ran to completion or a defined stop condition. Nothing
under `src/` touched (M1/Lane A untouched all night, per instruction). Full detail
in the per-task reports linked below; this is the index, not a replacement for them.

## 1. Escalations (verbatim, `docs/ESCALATIONS.md`)

> **T2/T3 — `docs/phase0.7-corpus-viability.md` does not exist.** T2 is specified
> as "Exactly as specified in `docs/phase0.7-corpus-viability.md` §2.2"; the file
> is absent from the repo. Proceeded on the night-batch directive's own inline
> field spec (T2's per-program table, and the two thresholds `T_none ≥ 1,000` /
> "floor of 8" given directly in §0.2.3) rather than aborting, since the core
> measurement doesn't depend on the missing file. Did **not** fabricate "the four
> numbers from Phase 0.7 §3" — that framing is not recoverable without the file.
> T3 proceeds on the same basis (its stated prerequisite, T2's seedable subset,
> doesn't depend on the missing doc either).

> **T3 — `semantic/subsumption_multiple_rules` diverges under `--magic-transform=*`.**
> Post-fix (see task outcomes), 4 output relations (`A`, `E`, `F`, `rel`) show
> genuine sorted-set-content differences, not just row-order. Uses Souffle's
> subsumption operator (`<=`) with `btree_delete`, entirely outside blueprint §4's
> grammar; the diverging `A`-relation rule chain has no negation in it at all.
> Three live explanations (subsumption/magic-set evaluation-order interaction;
> known Souffle limitation; cross-fragment corpus contamination — the file bundles
> many independent test cases) and one unrun cheapest-next-experiment (isolate the
> `A`/`graph` fragment into its own file) are in the full entry. T3 aborted here
> per instruction; not investigated further tonight.

## 2. Task outcomes

| Task | Outcome | Report |
|---|---|---|
| T1 | done | `docs/reports/night01-T1-audit.md` |
| T2 | done (missing-doc gap disclosed, not resolved) | `docs/reports/night01-T2-corpus.md` |
| T3 | **aborted mid-sweep** (27/31 clean before the abort) | `docs/reports/night01-T3-envelope.md` |
| T4 | done (exploratory, capped at 107/612 for the execution pass) | `docs/reports/night01-T4-exploratory.md` |
| T5 | done (full 622-file coverage, no execution needed) | `docs/reports/night01-T5-grammar.md` |
| T6 | done, complete (all 6 sizes, no DNF, no abort) | `docs/reports/night01-T6-scaling.md` |
| T7 | done | `docs/reports/night01-T7-harness.md` |
| T8 | done (no artifact found) | `docs/OPEN_QUESTIONS.md` (Q3) |

## 3. The four Phase 0.7 numbers (T2)

Not producible as specified — see escalation above. What the directive's own inline
spec gives instead: 36 pre-registered programs, 31 ran untransformed successfully
(4 rejected by Souffle as deliberate negative tests, 1 crashed). `T_none ≥ 1,000`
on **3 of 31**. Floor of 8: **not met**. Not acted on, per instruction.

## 4. `E_recoverable` distribution, pre-registered subset (T3)

27 of 31 candidates clean (1 crash, 1 diverged/aborted, 2 never reached).
`E_recoverable / T_souffle`: min 0.0, median 0.026, max 0.571
(`evaluation/set_ops`). **`E_recoverable = 0` on 10 of 27 (37%)** — the gap this
project targets is absent on over a third of this already-filtered corpus, called
out plainly rather than folded into an average. T4's broader (n=86, not
seedability-filtered) exploratory pass found 34% zero — the two rates agree.

## 5. Scaling table (T6)

| `n` | `T_none` | `T_souffle` | `T_guard` | `T_souffle/T_guard` |
|---|---|---|---|---|
| 250 | 62,534 | 44,811 | 285 | 157.2× |
| 500 | 250,093 | 168,759 | 594 | 284.1× |
| 1,000 | 1,000,193 | 611,806 | 1,194 | 512.4× |
| 2,000 | 4,000,461 | 2,471,792 | 2,462 | 1,004.0× |
| 4,000 | 16,000,877 | 9,812,420 | 4,878 | 2,011.6× |
| 8,000 | 64,001,614 | 40,805,441 | 9,615 | 4,243.9× |

All answers set-equal at every `n`. Ratio roughly doubles as `n` doubles — one
fixture shape, six points, not asserted as a general law (`docs/reports/
night01-T6-scaling.md`).

## 6. What did not work (before results, per CLAUDE.md §7)

- T1: 8 real provenance gaps found (ad hoc checks from earlier sessions never
  captured through `run_cmd`), backfilled.
- T2: first run silently overwrote one shared profile log across all 36 programs
  (relative `-p` path, wrong `cwd`) — every `T_none` came back null before the fix.
- T3: first comparison method (raw byte-diff) produced a false-positive abort
  (`example/orbits1`, row-order only) before being corrected to sorted set-equality.
- T4: the first "exploratory" candidate set accidentally duplicated T3's exact 36
  pre-registered programs (reused the wrong predicate condition); and
  `tuple_report.py` crashed on a profile log with no `relation` key at all. Both
  fixed before results were trusted.
- T7: `differential.py` had the identical relative-path bug T2 found, caught by a
  smoke test this time before the module was used anywhere real.
- A genuine, non-critical non-determinism was found and left unfixed by rule:
  `corpus_predicate.py`'s diagnostic `matched_output_relation` field (T1) —
  hard prohibition #2 blocks touching the predicate this batch.

## 7. What a skeptic attacks first

- The missing `docs/phase0.7-corpus-viability.md` means T2/T3's framing against
  "the four numbers" and named thresholds is this session's best-effort inline
  reconstruction, not a verified match to whatever that document actually says.
- T3's abort means the pre-registered corpus's `E_recoverable` distribution has
  n=27, not n=31 — two programs never got measured.
- T3's real divergence lives in a feature (`<=`/`btree_delete`) `dlc` will never
  implement, but it's still an unresolved correctness question about Souffle
  sitting inside the pre-registered corpus.
- T6's scaling curve is one fixture shape (uncontrolled reachable-set size) at six
  points — suggestive, not a proof the ratio grows linearly in general.
- T4's 66%/34% rates are per-program, not weighted by size — one large and one
  trivial program count equally.
