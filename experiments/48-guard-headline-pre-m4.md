# M3.5 — the headline run (§9)

Date: 2026-08-27. The project's central number, and the only one a hand
transform cannot produce: `dlc` decides (adornment + guard) and emits;
Soufflé evaluates (the T2 protocol). Full pipeline over all 5
`BENCHMARK_FAMILY` shapes at every `SCALE_POINTS.json` point, plus every
`CULPRIT_CANDIDATES` program. Runner: `harness/night_m3_5_headline.py`.
Full data: `measurements/m3-5-headline/summary.json`.

## What did not work

Nothing. **32/32 points ran to completion, 32/32 answer-identical** to the
untransformed baseline (25 `BENCHMARK_FAMILY` points across 5 shapes, plus
7 `CULPRIT_CANDIDATES` programs). Zero DNFs at any scale point up to
n=8,000. This is the first time this project's own tree-walking evaluator
is not the measurement bottleneck (M1 §4.3 hit 4 DNFs at 120s on the
identical shapes) — exactly T2's own stated reason for building the
protocol this way.

## 1. Three-column table, per shape

`T_none` / `T_souffle` / `T_dlc`, contribution = `T_souffle / T_dlc`.
`T_none / T_dlc` is never reported as a headline (M2-M3-BUILD.md §9.1:
"credits the guard with what magic sets already deliver").

### `same_generation_negation`

| scale | `T_none` | `T_souffle` | `T_dlc` | `T_souffle/T_dlc` |
|---|---|---|---|---|
| depth=4 | 51,301 | 30,907 | 907 | 34.1× |
| depth=5 | 706,439 | 421,563 | 3,363 | 125.4× |
| depth=6 | 23,897,431 | 14,352,999 | 19,555 | 734.0× |
| depth=7 | 212,212,055 | 127,332,304 | 58,271 | 2,185.6× |

### `transitive_closure_bound`

| scale | `T_none` | `T_souffle` | `T_dlc` | `T_souffle/T_dlc` |
|---|---|---|---|---|
| n=500 | 86,618 | 101 | 207 | **0.49×** |
| n=1,000 | 366,331 | 101 | 207 | 0.49× |
| n=2,000 | 1,519,041 | 101 | 205 | 0.49× |
| n=4,000 | 6,510,414 | 101 | 209 | 0.48× |
| n=8,000 | 25,510,941 | 101 | 208 | 0.49× |

**`dlc` is worse than Soufflé's own transform here, consistently ~2× (not
just at one point).** This is the one shape in the entire headline sweep
where `T_souffle/T_dlc < 1`. Flagged in M2's own report as a candidate
open question, now confirmed stable across all 5 scale points, not a
fluctuation. Not investigated to a root cause in this session — a
concrete, reproducible finding for the write-up to name plainly, not
smooth over.

### `ancestor_nonancestor`

| scale | `T_none` | `T_souffle` | `T_dlc` | `T_souffle/T_dlc` |
|---|---|---|---|---|
| n=500 | 250,450 | 101,250 | 105,552 | 0.96× |
| n=1,000 | 1,000,950 | 322,306 | 206,453 | 1.56× |
| n=2,000 | 4,001,950 | 1,550,117 | 408,951 | 3.79× |
| n=4,000 | 16,003,950 | 6,663,717 | 812,651 | 8.20× |
| n=8,000 | 64,007,950 | 25,391,507 | 1,620,552 | 15.67× |

Contribution grows with scale (0.96× → 15.67×) rather than holding flat —
worth noting alongside `docs/reports/m2-headline.md`'s finding that
`dlc`'s mechanical `bb` adornment (both grounding atoms precede the
negation) is structurally costlier than the hand guard's `bf` at every
size; the ratio still improves with `n` because `T_souffle` itself grows
faster than `T_dlc` on this shape.

### `culprit_cycle`

| scale | `T_none` | `T_souffle` | `T_dlc` | `T_souffle/T_dlc` |
|---|---|---|---|---|
| n=20 | 286 | 272 | 286 | 0.95× |
| n=50 | 510 | 432 | 510 | 0.85× |
| n=100 | 2,632 | 2,423 | 2,632 | 0.92× |
| n=200 | 7,496 | 7,028 | 7,496 | 0.94× |
| n=500 | 84,105 | 83,295 | 84,105 | 0.99× |

`T_dlc == T_none` exactly at every point (the guard declines the entire
program — see §2/§3 below). `T_souffle/T_dlc` is therefore always
slightly below 1: Soufflé's own automatic transform achieves a small,
real reduction here (its `@poscopy_1`-style duplication strategy,
`docs/reports/night02-T7-p5-precheck.md`) that this project's
cone-untransform strategy does not attempt to replicate (M2-M3-BUILD.md
§7 names this explicitly as the alternative strategy, not defaulted to).

### `reachability_complement`

| scale | `T_none` | `T_souffle` | `T_dlc` | `T_souffle/T_dlc` |
|---|---|---|---|---|
| n=250 | 62,534 | 44,811 | 55,411 | 0.81× |
| n=500 | 250,093 | 168,759 | 206,228 | 0.82× |
| n=1,000 | 1,000,193 | 611,806 | 812,428 | 0.75× |
| n=2,000 | 4,000,461 | 2,471,792 | 3,088,622 | 0.80× |
| n=4,000 | 16,000,877 | 9,812,420 | 12,513,376 | 0.78× |
| n=8,000 | 64,001,614 | 40,805,441 | 51,131,165 | 0.80× |

Same `bb`-vs-`bf` cost pattern as `ancestor_nonancestor` (`docs/reports/
m2-headline.md`), consistently below 1× here rather than improving with
scale — `p2.dl` (structurally identical, confirmed in M2) produces
identical numbers.

## 2. Guard-firing table

| program | clause fired | culprit predicates | cone (beyond culprit) | declined / total IDB | declined fraction |
|---|---|---|---|---|---|
| `same_generation_negation` | none | — | — | 0/3 | 0.0 |
| `transitive_closure_bound` | none | — | — | 0/2 | 0.0 |
| `ancestor_nonancestor` | none | — | — | 0/3 | 0.0 |
| `reachability_complement` | none | — | — | 0/3 | 0.0 |
| **`culprit_cycle`** | **(a)** | `{p, q, s}` | `{}` | 3/4 | **0.75** |
| `cc_arity3_twobound` | (a) | `{p, q, s}` | `{}` | 3/4 | 0.75 |
| `cc_edb_negated` | none | — | — | 0/3 | 0.0 |
| `cc_longer_cycle` | (a) | `{p, r1, r2, s}` | `{}` | 4/5 | 0.80 |
| `cc_neg_early` | (a) | `{p, q, s}` | `{}` | 3/4 | 0.75 |
| `cc_query_bothbound` | (a) | `{p, q, s}` | `{}` | 3/4 | 0.75 |
| `cc_third_relation` | (a) | `{p, q2, s}` | `{}` | 3/4 | 0.75 |
| `cc_mixed_fallback` | (a) | `{p, q, s}` | `{}` | 3/7 | 0.43 |

Provenance: computed directly from `guard.Decide` on each source file (the
same code path `dlc emit --transformer=guarded` runs), cross-checked
against `night_m3_5_headline.py`'s `guard_fired` byte-diff signal (agrees
on every row: fires exactly where `Decide` reports a nonempty declined
set).

**Every guard firing in this corpus is clause (a)** — no program in
either corpus triggered a rejection on clause (b)'s own separate check
(§5's `AssertNegationAllBound`, which has never failed on any program this
project has ever run, per M3.1's report). This is a structural fact of
the corpus available, not a claim that clause (b) never matters.

## 3. The blast-radius integer

**Total declined relations across the 12-program corpus (5
`BENCHMARK_FAMILY` + 7 `CULPRIT_CANDIDATES`): 22.** Per-program
distribution: 6 programs decline 0 relations (the guard never fires);
6 programs decline a large majority of their own IDB relations (0.75,
0.75, 0.75, 0.80, 0.75, 0.43 — mean 0.72 among firing programs). The
cone (`ConeClosure`, beyond the directly-implicated culprit SCC) is
**empty in every single case measured** — on this corpus, whenever the
guard declines anything, the unstratifiable SCC of the transformed
program already directly entangles every relation the cone mechanism
would otherwise need to add. This is disclosed plainly, not smoothed: the
cone's own correctness is independently verified (M3.3, exact match
against NIGHT-BATCH-03 T9's harness), but its *practical necessity beyond
the directly-flagged SCC* is not demonstrated anywhere in this specific
corpus.

**Reading this against blueprint failure mode #1** ("if the guard
declines everything, the project has no contribution"): the guard is
**not** universally vacuous — 6/12 programs (50%, and every
`BENCHMARK_FAMILY` shape except `culprit_cycle`) see it decline nothing
and deliver the full, unmodified magic-set reduction. But on every program
where it fires, it declines a large majority of that program's own
relations (43%–80%), and in this specific corpus that always means "the
entire negation-adjacent apparatus." The blast radius is real, is
measured precisely (not assumed), and its magnitude is large whenever it
occurs.

## 4. `dlc` vs. the hand transforms

Provenance: `docs/reports/night02-T5-guarded.md` (hand-guard `T_guard`
values, matched here point-for-point at identical `T_none`/`T_souffle`
values — independent cross-confirmation that this sweep's fixtures and
Soufflé invocation match the earlier session's exactly).

| shape | scale | `T_dlc` | `T_guard` (hand) | which is smaller |
|---|---|---|---|---|
| `same_generation_negation` | d4 | 907 | 452 | hand, 2.0× |
| | d5 | 3,363 | 1,680 | hand, 2.0× |
| | d6 | 19,555 | 9,776 | hand, 2.0× |
| | d7 | 58,271 | 29,134 | hand, 2.0× |
| `ancestor_nonancestor` (v1) | n500 | 105,552 | 25,500 | hand, 4.1× |
| | n1000 | 206,453 | 51,000 | hand, 4.0× |
| | n2000 | 408,951 | 102,000 | hand, 4.0× |
| | n4000 | 812,651 | 204,000 | hand, 4.0× |
| | n8000 | 1,620,552 | 408,000 | hand, 4.0× |
| `ancestor_nonancestor` (v2) | any | — | 996–15,996 | **not comparable — v2 is answer-wrong** (NIGHT-BATCH-03 T5, falsified) |
| `reachability_complement` | n250 | 55,411 | 285 | hand, 194.3× |
| | n500 | 206,228 | 594 | hand, 347.2× |
| | n1000 | 812,428 | 1,194 | hand, 680.4× |
| | n2000 | 3,088,622 | 2,462 | hand, 1,254.5× |
| | n4000 | 12,513,376 | 4,878 | hand, 2,565.6× |
| | n8000 | 51,131,165 | 9,615 | hand, 5,317.5× |
| `culprit_cycle` | n20 | 286 | 257 | hand, 1.1× |
| | n50 | 510 | 422 | hand, 1.2× |
| | n100 | 2,632 | 2,366 | hand, 1.1× |
| | n200 | 7,496 | 7,024 | hand, 1.07× |
| | n500 | 84,105 | 83,290 | hand, 1.01× |

**The hand transform wins on every shape, at every scale point measured,
without exception.** This is consistent, not surprising: every hand guard
in this project either (a) picks a smaller-than-mechanically-derivable
adornment on the `bb`-vs-`bf` shapes (`ancestor_nonancestor`,
`reachability_complement`, and structurally `same_generation_negation`'s
own `sg`/`notsg` pattern), a choice §3's own instruction explicitly
prohibits `dlc`'s SIPS from making without a cost model, or (b) achieves a
small residual reduction on `culprit_cycle` that the mechanical guard's
cone-untransform strategy does not attempt (Soufflé's own `@poscopy_1`
duplication path, not implemented here per M2-M3-BUILD.md §7's own
"implement cone-untransform as the default" instruction). **Ancestor v2's
exclusion from this comparison is itself a result, not a gap**: v2's
smaller `T_guard` numbers were never a legitimate comparison point to
begin with, since NIGHT-BATCH-03 T5 already established they come from an
answer-wrong construction.

## 5. Applicability

Restated, not re-derived (`docs/dlc-blueprint.md` §9, failure mode #8,
promoted to a finding 2026-08-22):

| source | negation-bearing programs | scale | zero rate |
|---|---|---|---|
| Soufflé `tests/`, pre-registered 36 | 31 ran | 3/31 clear 1,000 tuples | 37% (n=27) |
| Soufflé `tests/`, exploratory whole tree | 86 | not measured | 34% (n=86) |
| OpenRuleBench / RUBEN catalog | 1 shape, 3 scale points | good, but unreachable | n/a |

Stated as what it is: **roughly two-thirds of negation-bearing corpus
programs have something recoverable** (100% − 34–37%), not a claim that
the program class this project targets is rare in general — a claim
about corpus availability and about the zero rate on one dimension only.
Independently, this session's own NIGHT-BATCH-03 T4 census adds a fourth
data point on a *different* question (culprit-cycle structural
prevalence, not general negation-bearing prevalence): 0/817 real-world
files (195 in-grammar + 622 full Soufflé tree) structurally match the
culprit-cycle pattern beyond the one already-known file — consistent
with, and a sharper version of, failure mode #8's general finding.

## What did not work (restated, consolidated)

- `transitive_closure_bound`: `dlc` measurably and consistently worse
  than Soufflé's own transform (~0.49× at every one of 5 scale points) —
  a real, unexplained finding, not investigated to root cause this
  session.
- The hand transform beats `dlc`'s mechanical one on every single
  measured shape and scale point — the mechanical guard's contribution is
  real (up to 2,185× over `T_none` on `same_generation_negation`) but
  never state-of-the-art against a human author willing to hand-derive a
  non-general adornment.
- The fallback cone has never, in this corpus, added anything beyond the
  directly-flagged culprit SCC — its correctness is independently
  verified (M3.3) but its practical necessity is not demonstrated here.

## Verdict

**M3.5: DONE.** 32/32 points, 0 DNFs, 32/32 answer-identical. Three-column
table, guard-firing table, blast-radius integer (22 declined relations,
distribution 0%–80% per firing program, cone always empty), hand-transform
comparison (hand wins everywhere, with the ancestor-v2 exclusion itself a
result), and applicability restated, all delivered per §9's required
structure. Per §11, the required paragraph on why this headline is not
1.00× the way M1's was:

> Semi-naive evaluation changes the *strategy* for a fixed program — both
> evaluators compute the same minimal Herbrand model, so distinct-tuple
> counts cannot differ. The magic-set transform changes the *program* — a
> different program with a different minimal model, which happens to
> agree on the query relation. That is why one is necessarily 1.00× and
> the other is not.

Per §12, this closes the sequence through item 8 ("the whole thesis") —
M3.6 (the presentation artifact) is explicitly gated on this report
existing, and only becomes worth starting now that a blast-radius integer
(22, distribution 0–0.80) has actually been produced.
