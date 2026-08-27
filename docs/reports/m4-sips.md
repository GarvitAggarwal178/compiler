# M4 — demand relaxation on negated occurrences (NIGHT-BATCH-04 A)

Date: 2026-08-27. `M4 sips.md` sections 1, 2, 3, 5. Q12 pre-registered
(`docs/OPEN_QUESTIONS.md`, commit `22c42d4`) before any code change.

## What did not work

- **`same_generation_negation` does not fully collapse to a single
  adornment, unlike the other two shapes.** The negated occurrence
  `!sg(x,y)` inside `notsg`'s rule relaxes correctly (`bb`→`bf`, verified
  directly via `NegatedOccurrenceAdornments()`, not inferred), exactly as
  section 2's table predicts. But `sg`'s own recursive rule —
  `sg(x,y):-parent(x,xp),parent(y,yp),sg(xp,yp),x!=y` — schedules
  `parent(y,yp)` (a full-extent scan under the `bf` entry, since `y` is
  free) *before* the recursive call `sg(xp,yp)` in left-to-right source
  order. By the time `sg(xp,yp)` is reached, `yp` is already bound, so
  this **positive** occurrence is correctly adorned `bb` — relaxation
  only ever applies to negated occurrences (section 1's scope), so this
  is not a bug, it is a structural fact about `sg`'s own recursion shape
  that a source-order-only SIPS cannot avoid without reordering positive
  atoms (prohibited, section 6 / `m2 m3.md` §3). `sg` ends up needing
  *two* adornments (`sg_bf`, `sg_bb`) where `ancestor`/`reach` end up
  needing *one*. Measured anyway (below): the total tuple count still
  drops (907→681 incl-sup at `depth=4`), just not via a full collapse.
- **The predicted 80–180× reduction on `p2.dl` was not fully reached
  under the project's existing default convention (incl-sup).** Measured
  reduction is 56.9× (55,411→974), same order of magnitude as predicted
  but below the 80× floor. Under excl-sup it is 1.9× (468→252) — nowhere
  near the range, because excl-sup's pre-relaxation baseline (468) was
  already small; the two conventions answer different questions and only
  one of them was what Q12 predicted against. Reported as measured, not
  adjusted to fit.

## §1/§2 gate — does the rule reproduce the table

Verified directly on the **occurrence itself** (`NegatedOccurrenceAdornments()`,
not by reading the printed program), pre- vs post-relaxation:

| shape | negated occurrence | `PreAdorn` | `Adorn` (relaxed) | table predicts |
|---|---|---|---|---|
| `ancestor_nonancestor` | `!ancestor(x,y)` | `bb` | **`bf`** | `ancestor^bf` — match |
| `reachability_complement` / `p2.dl` | `!reach(x,y)` | `bb` | **`bf`** | `reach^bf` — match |
| `same_generation_negation` | `!sg(x,y)` | `bb` | **`bf`** | `sg^bf` — match |

**All three reproduce the table exactly at the occurrence level.**
`src/transform/guard/seeding_test.go`'s `TestNegatedOccurrenceRelaxationCollapsesToBF`
pins the ancestor case; `src/transform/magicset/adorn_test.go`'s
`TestAdornNegatedOccurrenceRelaxesToBF` pins it a second way. Whether this
collapses the *program's whole adornment set* to one relation is a
separate, per-shape structural question (see "what did not work" above)
— two of three do, `same_generation_negation` does not.

Emitted adornment sets, before (`measurements/m4-sips/before/`) vs after
(`measurements/m4-sips/after/`):

| program | before | after |
|---|---|---|
| `ancestor_nonancestor` | `ancestor_bb`, `ancestor_bf` | `ancestor_bf` only |
| `reachability_complement` / `p2.dl` | `reach_bb`, `reach_bf` | `reach_bf` only |
| `same_generation_negation` | `sg_bb` only | `sg_bf`, `sg_bb` |
| `culprit_cycle` (control) | `p_bf`,`q_bf`,`s_b` | unchanged |
| `transitive_closure_bound` (control) | `tc_bf` | unchanged |

## §4 gate — answer-identical, hard requirement

`harness/m2_accept.py`, all 5 `BENCHMARK_FAMILY` shapes + `p2.dl`,
smallest scale point (`measurements/m4-sips/gate/summary.json`):

| program | comparable | answers identical |
|---|---|---|
| `p2.dl` | yes | **yes** |
| `reachability_complement` | yes | **yes** |
| `ancestor_nonancestor` | yes | **yes** |
| `same_generation_negation` | yes | **yes** |
| `transitive_closure_bound` | yes | **yes** |
| `culprit_cycle` | no (unstratifiable, expected — guard territory) | n/a |

**5/5 comparable cases answer-identical. Zero divergences.** The hard
gate passes; the relaxation is not reported further pending this.

## Tuple totals, same fixture, before vs after (both conventions)

`measurements/m4-sips/gate/summary.json` (after) cross-referenced against
a same-methodology re-run of the pre-relaxation snapshot
(`measurements/m4-sips/before_gate/`, same fixtures):

| program | `T_dlc` incl-sup before→after | `T_dlc` excl-sup before→after |
|---|---|---|
| `p2.dl` / `reachability_complement` (n=250) | 55,411 → **974** (56.9×) | 468 → **252** (1.9×) |
| `ancestor_nonancestor` (n=500) | 105,552 → **6,102** (17.3×) | 27,702 → **2,752** (10.1×) |
| `same_generation_negation` (depth=4) | 907 → **681** (1.33×) | 453 → **228** (2.0×) |

`p4prime.dl` (hand guard) on the same n=250 fixture: **285** (single
number, both conventions identical — no supplementary relations in a
hand-written program). `dlc` post-relaxation: 974 incl-sup (3.4× above
p4prime, just outside the predicted "within 3×"), 252 excl-sup (**below**
p4prime's 285) — under the excl-sup convention `dlc`'s mechanical
transform is now cheaper than the hand-optimized one on this shape.

## Q8, closed

`ancestor_nonancestor_guarded.dl` (v1) measured on the same n=500 fixture,
same methodology: `T_guard = 25,500` (`measurements/m4-sips/v1_check/`,
exactly reproducing the number `docs/OPEN_QUESTIONS.md` Q11 already cited
— methodology cross-check passes). `dlc` post-relaxation at n=500:
**6,102 incl-sup / 2,752 excl-sup — both well below 25,500**, not "near"
it. Per `NIGHT-BATCH-04.md` A's own framing: this closes Q8 in the
**v1-is-suboptimal** direction. v1's choice to propagate `m_ancestor`
across the recursion (Q11's finding) is confirmed a real, avoidable cost
— `dlc`'s correct mechanical seeding beats the hand-written guard by
4.2×–9.3× at this scale point, not a rounding-distance match.
`docs/OPEN_QUESTIONS.md` Q8 entry updated below.

## §3 gate — the full 32-point headline re-run

`harness/night_m3_5_headline_m4.py` (a copy of `night_m3_5_headline.py`
writing to `measurements/m3-5-headline-m4/` so the original M3.5 baseline
stays untouched for direct citation). **32/32 points, 0 DNFs, 32/32
answers identical** — same clean result M3.5 already established, now
re-confirmed against the relaxed transform.

Three-column table, `T_none`/`T_souffle`/`T_dlc` (incl-sup, the project's
existing default), contribution = `T_souffle/T_dlc`, **before → after**
compared directly against the already-committed `m3-headline.md` numbers:

### `reachability_complement`

| scale | `T_souffle` | `T_dlc` before | `T_dlc` after | contribution before | contribution after |
|---|---|---|---|---|---|
| n=250 | 44,811 | 55,411 | **974** | 0.81× | **46.0×** |
| n=500 | 168,759 | 206,228 | **1,912** | 0.82× | **88.3×** |
| n=1,000 | 611,806 | 812,428 | **3,812** | 0.75× | **160.5×** |
| n=2,000 | 2,471,792 | 3,088,622 | **7,543** | 0.80× | **327.7×** |
| n=4,000 | 9,812,420 | 12,513,376 | **15,128** | 0.78× | **648.6×** |
| n=8,000 | 40,805,441 | — (not run) | **30,391** | — | **1,342.7×** |

### `ancestor_nonancestor`

| scale | `T_souffle` | `T_dlc` before | `T_dlc` after | contribution before | contribution after |
|---|---|---|---|---|---|
| n=500 | 101,250 | 105,552 | **6,102** | 0.96× | **16.6×** |
| n=1,000 | 322,306 | 206,453 | **7,503** | 1.56× | **43.0×** |
| n=2,000 | 1,550,117 | 408,951 | **11,001** | 3.79× | **140.9×** |
| n=4,000 | 6,663,717 | 812,651 | **16,701** | 8.20× | **399.0×** |
| n=8,000 | 25,391,507 | 1,620,552 | **28,602** | 15.67× | **887.8×** |

Both shapes move from **worse-than-or-barely-beating** Soufflé's own
transform to **beating it by two to three orders of magnitude**, growing
with scale in both cases (the relaxed transform's cost grows roughly
linearly in the answer-set size; Soufflé's un-relaxed magic transform
apparently does not benefit from the same structural fact).

### `same_generation_negation`

| scale | `T_souffle` | `T_dlc` before | `T_dlc` after | contribution before | contribution after |
|---|---|---|---|---|---|
| depth=4 | 30,907 | 907 | **681** | 34.1× | **45.4×** |
| depth=5 | 421,563 | 3,363 | **2,523** | 125.4× | **167.1×** |
| depth=6 | 14,352,999 | 19,555 | **14,667** | 734.0× | **978.7×** |
| depth=7 | 127,332,304 | 58,271 | **43,704** | 2,185.6× | **2,913.5×** |

Already the best-performing shape before this change (its negated
occurrence's magic seed was cheap even at `bb`); still improves
25–33% at every scale point from the partial collapse (`sg_bf` avoiding
`sg_bb`'s old role as the sole, more broadly-seeded entry point).

### `culprit_cycle`, `transitive_closure_bound` (controls, unaffected)

`culprit_cycle`: `T_dlc = T_none` exactly at every scale point, byte-for-
byte identical to `m3-headline.md`'s numbers (no negated occurrence
survives the guard's fallback here — nothing for section 2 to touch).
`transitive_closure_bound`: `T_dlc` (incl-sup) unchanged at 207/207/205/
209/208 — no negation in this shape, section 2 does not apply. Both
confirm no regression from this change.

## §5 gate — supplementary counting convention

`transitive_closure_bound`'s open question closes with a number:
**`T_dlc` (excl-sup) = 101 at every one of the 5 scale points, exactly
equal to `T_souffle` = 101.** The "anomaly" (`m3-headline.md`: "`dlc` is
worse than Soufflé's own transform here, consistently ~2×") **was the
counting convention, not a defect** — confirmed exactly, not
approximately. `T_dlc` (incl-sup) stays 207–209 (unchanged; this shape
has one adornment and one supplementary chain, so incl-sup counts
~101 extra checkpoint tuples Soufflé's profile has no equivalent
relation for).

`same_generation_negation`'s previously-reported 2,185.6× contribution at
depth=7 is confirmed **understated** under incl-sup, per the document's
own prediction: excl-sup contribution at depth=7 is
`127,332,304 / (T_dlc_excl_sup=14,569)` = **8,741×**, roughly 4× larger
than the incl-sup figure already reported.

Both columns now emitted by `harness/tuple_report.py`
(`T_excl_copy_excl_sup`) and `harness/m2_accept.py`
(`T_original_excl_sup`/`T_candidate_excl_sup`); both conventions are
reported side by side in every table above, neither chosen as "the"
number, per instruction.

## What a skeptic attacks first

- `same_generation_negation`'s partial (not full) collapse is the
  cleanest attack surface: a reader could reasonably ask "why does the
  rule work on two shapes and not the third," and the honest answer is a
  source-order structural fact about `sg`'s own recursive rule, not a
  limitation of the relaxation rule itself — disclosed above, not
  smoothed over.
- The `p2.dl` 80–180× prediction landed at 56.9× (incl-sup) — inside the
  right order of magnitude, below the stated floor. Reported as a miss,
  not rescued by re-deriving a different predicted range after the fact.
- `dlc` incl-sup (974) is still 3.4× worse than the hand guard `p4prime.dl`
  (285) on `p2.dl`, even after the fix — the gap did not fully close, and
  incl-sup is this project's existing default convention everywhere else,
  so this is the number that actually governs cross-report comparability,
  not the more flattering excl-sup one.

## Verdict

**M4-SIPS A: DONE.** Both changes land, the hard answer-identical gate
passes on every comparable case (5/5), the §5 counting-convention
question closes exactly (101==101), Q8 closes (v1 confirmed suboptimal),
and the headline sweep is fully re-run (32/32, 0 DNFs) with `reachability_complement`
and `ancestor_nonancestor` moving from sub-1× or single-digit contribution
to hundreds-to-low-thousands ×, growing with scale. One genuine partial
miss (`same_generation_negation`'s non-collapse) and one genuine
under-target (`p2.dl`'s 56.9× vs 80–180× predicted) are reported plainly,
not adjusted to match the pre-registered prediction.
