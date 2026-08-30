# NIGHT-BATCH-02 T4 — baseline sweep: `T_none`, `T_souffle`, `E_recoverable`

Date: 2026-08-23. Soufflé 2.5 only — no `dlc`, no hand-transforms. Two
configurations per shape per scale point: untransformed, `--magic-transform=*`.
Caps: 300s timeout, 8GB memory (`RLIMIT_AS`), ascending scale-point order.
Runner: `harness/night02_t4_baseline.py`. Full data:
`measurements/night02-t4/summary.json`, per-run `cmd.txt`/`stdout.txt`/
`stderr.txt`/`prof.log` under `measurements/night02-t4/<shape>/<tag>-{none,souffle}/`.

## What did not work

Nothing. No DNF, no cap fired, no shape aborted, no answer-relation mismatch
at any scale point across all 5 shapes × their scale points (24 runs, two
configurations each = 48 Soufflé invocations). The excl-copy/incl-copy
distinction (`harness/tuple_report.py`), carried into this sweep because the
task required recording both, **never actually diverged for any shape or
scale point here** — no `COPY_T`-shaped relation appeared in any of the 48
profiles. That distinction mattered once, for P1's `.output`-forced case
(`docs/MEASUREMENTS.md`); it did not matter for any shape in this family.
Reported as a null result, not omitted.

## `reachability_complement` — reused, not regenerated, but re-run for consistent instrumentation

Per `SCALE_POINTS.json`'s own note, this shape's fixtures were not
regenerated (T3) — they are NIGHT-BATCH-01 T6's `fixtures/p2-scale-<n>/`.
This task re-ran Soufflé against those existing fixtures using
`tests/corpus/BENCHMARK_FAMILY/reachability_complement.dl` (not
`tests/programs/p2.dl`) so the row uses the same instrumentation as the
other 4 shapes. **The two `.dl` files are structurally identical** (same
`edge`/`node`/`reach`/`unreach` relations and rules; only the `.output`
relation name differs, `q2` vs `q_unreach`) — confirmed both by inspection
and by measurement: `T_none`, `T_souffle`, `E_recoverable` at every one of
the 6 points match T6's already-committed numbers
(`docs/reports/night01-T6-scaling.md`) exactly, e.g. n=250:
62,534 / 44,811 / 44,742 both times.

## Per-shape tables

`T` values are excl-copy (== incl-copy for every row this sweep, see above).
`E_recoverable` = Σ tuples over `@neglabel.`-prefixed relations.

### `same_generation_negation`

| scale point | `T_none` | `T_souffle` | `E_recoverable` | answers identical |
|---|---|---|---|---|
| depth=4,branching=4 (226 persons) | 51,301 | 30,907 | 30,456 | yes |
| depth=5,branching=4 (840 persons) | 706,439 | 421,563 | 419,884 | yes |
| depth=6,branching=4 (4,888 persons) | 23,897,431 | 14,352,999 | 14,343,224 | yes |
| depth=7,branching=4 (14,567 persons) | 212,212,055 | 127,332,304 | 127,303,171 | yes |

`@neglabel.sg` carries essentially all of `E_recoverable` at every point
(e.g. depth=4: `E_recoverable=30,456`, entirely `@neglabel.sg`) — the magic
transform isolates `sg` (the negated relation inside `notsg`) under the
`@neglabel.` marker rather than restricting it, consistent with the
project's core differentiator claim.

### `transitive_closure_bound` — positive fragment, no negation

| scale point | `T_none` | `T_souffle` | `E_recoverable` | answers identical |
|---|---|---|---|---|
| n=500 | 86,618 | 101 | 0 | yes |
| n=1,000 | 366,331 | 101 | 0 | yes |
| n=2,000 | 1,519,041 | 101 | 0 | yes |
| n=4,000 | 6,510,414 | 101 | 0 | yes |
| n=8,000 | 25,510,941 | 101 | 0 | yes |

**`E_recoverable = 0` at every scale point.** This shape has no negation
(disclosed in its own header comment as the family's positive-fragment
baseline) — there is nothing for `@neglabel.` to isolate, so it has no gap
to recover, and that is as much a result as the shapes that do. `T_souffle`
is flat at 101 regardless of `n` (core_size 50 reachable nodes + the seed
relation + bookkeeping, not decomposed further here) while `T_none` grows
with `n` — the bound-query magic transform collapses the unbound
whole-graph transitive closure to just the query's reachable set, exactly
Beeri & Ramakrishnan's point.

### `ancestor_nonancestor`

| scale point | `T_none` | `T_souffle` | `E_recoverable` | answers identical |
|---|---|---|---|---|
| n=500 | 250,450 | 101,250 | 100,349 | yes |
| n=1,000 | 1,000,950 | 322,306 | 320,405 | yes |
| n=2,000 | 4,001,950 | 1,550,117 | 1,546,216 | yes |
| n=4,000 | 16,003,950 | 6,663,717 | 6,655,816 | yes |
| n=8,000 | 64,007,950 | 25,391,507 | 25,375,606 | yes |

`@neglabel.ancestor` dominates `E_recoverable` (e.g. n=500:
`E_recoverable=100,349`, entirely `@neglabel.ancestor`) — same pattern as
`same_generation_negation`: the negated relation is isolated, not
restricted.

### `culprit_cycle` (P5 shape)

| scale point | `T_none` | `T_souffle` | `E_recoverable` | answers identical |
|---|---|---|---|---|
| n=20 | 286 | 272 | 18 | yes |
| n=50 | 510 | 432 | 34 | yes |
| n=100 | 2,632 | 2,423 | 60 | yes |
| n=200 | 7,496 | 7,028 | 120 | yes |
| n=500 | 84,105 | 83,295 | 340 | yes |

`E_recoverable` is small relative to `T_souffle` at every point here (e.g.
n=500: 340 of 83,295, ~0.4%) — a visibly different profile from the other
three negation-bearing shapes, where `@neglabel` carries the bulk of
`T_souffle`. `@neglabel.s` is the isolated relation (n=20:
`E_recoverable=18`, entirely `@neglabel.s`). Whether `T_souffle` itself is
even meaningful here depends on whether the second `p` rule fires and `q`
survives the inliner — that is T7's job (P5 precheck), not this one; this
row does not presuppose T7's answer.

### `reachability_complement` (reused fixtures, T6 lineage)

| scale point | `T_none` | `T_souffle` | `E_recoverable` | answers identical |
|---|---|---|---|---|
| n=250 | 62,534 | 44,811 | 44,742 | yes |
| n=500 | 250,093 | 168,759 | 168,572 | yes |
| n=1,000 | 1,000,193 | 611,806 | 611,419 | yes |
| n=2,000 | 4,000,461 | 2,471,792 | 2,470,869 | yes |
| n=4,000 | 16,000,877 | 9,812,420 | 9,810,665 | yes |
| n=8,000 | 64,001,614 | 40,805,441 | 40,802,212 | yes |

Identical to T6's already-committed numbers at every point (see note above).

## What a skeptic attacks first

- `transitive_closure_bound`'s `T_souffle=101` is suspiciously flat across
  a 16× range of `n` — worth an independent per-relation breakdown before
  trusting it further (not done here; `tuple_report.py`'s `per_relation`
  field has it, just not surfaced in this report's tables).
- `culprit_cycle`'s small `E_recoverable`-to-`T_souffle` ratio compared to
  the other three shapes could mean the guard has little to recover here,
  or could mean the shape's magic-transform behavior is qualitatively
  different (dead rule, inlined pivot) — T7 checks this directly and this
  report does not anticipate that answer.
- No `T_guard` column yet — this is the baseline only. T5 is the headline.

## Provenance

`measurements/night02-t4/summary.json`, 24 scale-point rows across 5 shapes,
48 Soufflé invocations, all `status: ok`, no DNF. Runner:
`harness/night02_t4_baseline.py`. Completed inside the 120-minute cap.
