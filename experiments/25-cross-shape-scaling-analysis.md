# NIGHT-BATCH-02 T6 — cross-shape scaling analysis

Date: 2026-08-23. Prerequisite T5 (passed). Answers the question
NIGHT-BATCH-01 T6 could not: was `T_souffle/T_guard`'s Θ(n) growth (found
there for the P2/`reachability_complement` shape alone) a property of the
transform mechanism, or specific to that one shape?

**Method:** pairwise log-log slope between each pair of *consecutive*
measured scale points, `slope = ln(y2/y1) / ln(x2/x1)`, using each shape's
own already-committed T4 (`T_none`, `T_souffle`) and T5 (`T_guard`) numbers.
Not a regression, not smoothed, not extrapolated past the largest completed
point in any shape. **No number is aggregated across shapes** — every
table below is one shape only.

`x` is `n` for the three graph-based shapes (`ancestor_nonancestor`,
`transitive_closure_bound`, `reachability_complement`, `culprit_cycle`,
where the fixture's own size parameter is the natural axis) and `n_persons`
for `same_generation_negation` (whose scale points are `depth`, not `n` —
`n_persons` is the structural count T3 already recorded per point, and is
the axis the tuple counts actually scale against).

## Per-shape growth

### `same_generation_negation` (x = n_persons: 226, 840, 4,888, 14,567)

| pair | `T_none` slope | `T_souffle` slope | `T_guard` slope |
|---|---|---|---|
| 226→840 | 2.00 | 1.99 | 1.00 |
| 840→4,888 | 2.00 | 2.00 | 1.00 |
| 4,888→14,567 | 2.00 | 2.00 | 1.00 |

Exactly Θ(n²) for `T_none`/`T_souffle`, exactly Θ(n) for `T_guard`, at
every consecutive pair, no exceptions. `T_souffle/T_guard` therefore grows
Θ(n): 68.4× → 250.9× → 1,468.2× → 4,370.6×.

### `ancestor_nonancestor` (x = n: 500, 1000, 2000, 4000, 8000)

| pair | `T_none` slope | `T_souffle` slope | `T_guard` slope |
|---|---|---|---|
| 500→1,000 | 2.00 | 1.67 | 1.00 |
| 1,000→2,000 | 2.00 | 2.27 | 1.00 |
| 2,000→4,000 | 2.00 | 2.10 | 1.00 |
| 4,000→8,000 | 2.00 | 1.93 | 1.00 |

`T_none` is exactly Θ(n²) at every pair (clean); `T_guard` is exactly Θ(n)
at every pair (clean, `T_guard` values are themselves exactly `51n`:
25,500 / 51,000 / 102,000 / 204,000 / 408,000). `T_souffle`'s slope is
noisier (1.67–2.27) but averages close to 2 — consistent with Θ(n²), not
cleanly confirmed the way `T_none`/`T_guard` are. `T_souffle/T_guard`
grows: 4.0× → 6.3× → 15.2× → 32.7× → 62.2× — same Θ(n) growth *class* as
`same_generation_negation`, but roughly two orders of magnitude smaller in
absolute size across the same `n` range (see cross-shape comparison below).

### `transitive_closure_bound` (x = n: 500, 1000, 2000, 4000, 8000) — no guard

| pair | `T_none` slope | `T_souffle` slope |
|---|---|---|
| 500→1,000 | 2.08 | 0.00 |
| 1,000→2,000 | 2.05 | 0.00 |
| 2,000→4,000 | 2.10 | 0.00 |
| 4,000→8,000 | 1.97 | 0.00 |

`T_none` is Θ(n²) (slopes cluster around 2, some noise: 1.97–2.10).
`T_souffle` is exactly flat (101 at every point, slope 0 — Θ(1)) — the
bound-query magic transform alone, with no negation and no guard needed,
already collapses this shape to the query's reachable set regardless of
total graph size. No `T_guard` column: `E_recoverable = 0` here (T4), so
there is nothing to guard, and `T_souffle` is already Θ(1).

### `culprit_cycle` (x = n: 20, 50, 100, 200, 500) — the disagreement

| pair | `T_none` slope | `T_souffle` slope | `T_guard` slope |
|---|---|---|---|
| 20→50 | 0.63 | 0.50 | 0.54 |
| 50→100 | 2.37 | 2.49 | 2.49 |
| 100→200 | 1.51 | 1.54 | 1.57 |
| 200→500 | 2.64 | 2.70 | 2.70 |

**No clean growth class fits these 5 points** — the slopes bounce between
0.5 and 2.7 pair to pair, for all three of `T_none`, `T_souffle`, and
`T_guard` alike. Not smoothed into an average here, since the task
explicitly prohibits that; reported as noisy/unfit rather than forced into
a power law. What *is* clean: `T_souffle` and `T_guard` track each other
almost exactly at every point (ratio 1.0–1.1× throughout) — the guard
contributes essentially nothing on this shape, at every scale measured,
regardless of the underlying noise in the absolute growth curve. This is
the genuine cross-shape disagreement the task asked this analysis to look
for.

### `reachability_complement` (x = n: 250, 500, 1000, 2000, 4000, 8000)

| pair | `T_none` slope | `T_souffle` slope | `T_guard` slope |
|---|---|---|---|
| 250→500 | 2.00 | 1.91 | 1.06 |
| 500→1,000 | 2.00 | 1.86 | 1.01 |
| 1,000→2,000 | 2.00 | 2.01 | 1.04 |
| 2,000→4,000 | 2.00 | 1.99 | 0.99 |
| 4,000→8,000 | 2.00 | 2.06 | 0.98 |

`T_none` exactly Θ(n²) at every pair; `T_souffle` and `T_guard` both
consistent with Θ(n²) and Θ(n) respectively, with the same order of
pairwise noise `ancestor_nonancestor` shows on `T_souffle`. Matches
NIGHT-BATCH-01 T6's already-reported reading of this same shape exactly
(`docs/reports/night01-T6-scaling.md`) — this is not a new finding, it is
a re-derivation from freshly-run numbers landing on the same growth class.

## Cross-shape comparison — agreement and disagreement, not aggregation

**Three of the four guarded shapes agree on growth *class*:**
`same_generation_negation`, `ancestor_nonancestor`, and
`reachability_complement` all show `T_none`/`T_souffle` ≈ Θ(n²) and
`T_guard` ≈ Θ(n), hence `T_souffle/T_guard` ≈ Θ(n) — despite one being a
tree-recursion shape (SG) and two being chain-recursion shapes (ancestor,
reachability). **They do not agree on magnitude.** At comparable `x` (~500,
capped to each shape's own units), the ratio spans two orders of magnitude:

| shape | ratio at smallest matching-order `x` | ratio at largest point |
|---|---|---|
| `ancestor_nonancestor` (n=500) | 4.0× | 62.2× (n=8,000) |
| `reachability_complement` (n=500) | 284.1× | 4,243.9× (n=8,000) |
| `same_generation_negation` (n_persons=840, closest to 500) | 250.9× | 4,370.6× (n_persons=14,567) |

`same_generation_negation`'s and `reachability_complement`'s absolute
ratios track each other closely at comparable scale; `ancestor_nonancestor`
sits roughly 50–70× lower at every comparable point despite the identical
Θ(n)/Θ(n²) growth classes. This is not explained by this task — T6's job
is the growth-class fit, not the constant-factor explanation, and
CLAUDE.md's per-shape reporting rule means this gap is recorded, not
smoothed over or explained away here.

**`culprit_cycle` disagrees on class, not just magnitude:** its ratio does
not grow with `n` at all (flat ~1.0–1.1× across the full 20–500 range,
where the other three shapes' ratios grow by 1–2 orders of magnitude over
comparable ranges). Per the task's own framing, this is the more
interesting and more honest result than uniformity would have been: the
guard's contribution is shape-dependent, not a fixed property of "applying
the guard technique," and `culprit_cycle`'s own guard (T5) was already
disclosed as a partial restriction (only `p`, not `q`/`s`, for
stratification-safety reasons) — the flat ratio is consistent with that
disclosed limitation, not a surprise on top of it.

`transitive_closure_bound` is excluded from the ratio comparison (no
`T_guard`, nothing to guard) but its own `T_souffle` growth class (Θ(1),
flat) is itself a fourth distinct growth class among the five shapes in
this family — the positive-fragment baseline needs no guard and gets the
full benefit from Soufflé's own automatic transform alone.

## What did not work / caveats

- `culprit_cycle`'s 5 scale points do not admit a clean power-law fit by
  the pairwise-slope method used for the other shapes — reported as noisy,
  not forced into an average.
- No point in any shape came close to the 300s/8GB caps in T4 or T5, so
  "ascending order so a cap truncates the top, not the middle" was never
  actually tested by a real truncation in this batch, same caveat
  NIGHT-BATCH-01 T6 recorded for `reachability_complement` alone.
- The magnitude gap between `ancestor_nonancestor` and the other two Θ(n)-
  class shapes is noted but not explained; a skeptic should ask why before
  trusting any generalization from these three shapes to a fourth.

## Provenance

All numbers in this report are read directly from
`measurements/night02-t4/summary.json` and `measurements/night02-t5/
summary.json` (already committed, T4/T5) plus `measurements/
night02-t3-fixtures-summary.json` (`n_persons` for
`same_generation_negation`, already committed, T3) — no new Soufflé
invocations in this task. Completed well inside the 60-minute cap.
