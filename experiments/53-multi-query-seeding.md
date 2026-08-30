# PUNCH-LIST P1 — multi-query seeding

Date: 2026-08-27. `src/transform/magicset/adorn.go`: `FindQuery` (single
candidate) replaced by `FindQueries` (every candidate); `Adorn` seeds the
worklist from all of them. `src/transform/magicset/rules.go`'s
`Generate`/`GenerateMixed` emit one seed + one query-projection rewrite
per query, and exclude every query's own projection rule from the
pass-through set, not just one.

## What did not work

Nothing broke. The change is exactly seed collection, as the punch list
predicted — `Adorn`'s worklist already accepted multiple initial pushes
(any two occurrences discovering the same `(predicate,adornment)` pair
already deduped through the same `push`/`processed` mechanism); the only
code that assumed "exactly one query" was `FindQuery`'s own early
`return` on the first match, and the handful of call sites built around
a single `*QueryInfo` instead of a slice. Five call sites updated
(`transformer.go`, `guard/decide.go`, `guard/stratify.go`,
`cmd/dlc/main.go`'s `runExplain`, and `rules.go`'s two loops), all
Go test files updated to match, one new test added
(`TestFindQueriesCollectsMultiple`, pins the two-output case directly).

## Gate — answer-identical, hard requirement

`harness/punchlist_p1_gate.py` (new, committed): re-emits all 6 original
programs (5 `BENCHMARK_FAMILY` shapes + `p2.dl`) plus all 4
`tests/corpus/CONE_CORPUS/` (task B) constructions with the post-P1
binary, compares against the untransformed baseline via
`harness/m2_accept.py` — same methodology A's and B's own gates already
used.

**9/9 comparable cases answer-identical.** (`culprit_cycle` remains
correctly not-comparable — unstratifiable by design, unchanged.)
`measurements/punch-list/p1-gate/summary.json`.

The 6 original programs' `T_candidate` values are **byte-identical** to
their pre-P1 numbers (974, 974, 6102, 681, 207 — `culprit_cycle` n/a) —
expected and confirmed, not assumed: none of these 6 programs has a
second independent `.output` branch, so multi-query seeding changes
nothing for them.

## Re-run of task B's measurement

`harness/night04_b_cone_gate.py` (unmodified script, re-run against the
post-P1 binary). Cone cross-check unchanged (4/4 exact agreement against
`harness/cone_metric.py`, same as before P1 — expected, since P1 does not
touch culprit/cone computation, only which OTHER relations get properly
adorned instead of left Untouched). **12/12 points still
answer-identical.**

`T_guarded` vs. `T_none`, all 4 programs, all 3 scale points
(`measurements/night04-b-cone/summary.json`):

| program | n | `T_none` | `T_souffle` | `T_dlc` (before P1) | `T_dlc` (after P1) | `T_guarded < T_none`? |
|---|---|---|---|---|---|---|
| `cc_cone_only` | 20/50/100 | 255/787/3,866 | 249/762/3,836 | 255/787/3,866 | 255/787/3,866 (unchanged) | no (no sibling — expected) |
| `cc_sibling_emptycone` | 20 | 514 | 258 | 514 | **303** | **yes, 1.70×** |
| `cc_sibling_emptycone` | 50 | 1,101 | 783 | 1,101 | **861** | **yes, 1.28×** |
| `cc_sibling_emptycone` | 100 | 4,059 | 3,944 | 4,059 | **3,995** | **yes, 1.02×** |
| `cc_both` | 20 | 520 | 275 | 520 | **309** | **yes, 1.68×** |
| `cc_both` | 50 | 1,036 | 766 | 1,036 | **796** | **yes, 1.30×** |
| `cc_both` | 100 | 3,935 | 3,838 | 3,935 | **3,871** | **yes, 1.02×** |
| `cc_cone_proper_subset` | 20 | 489 | 265 | 489 | **279** | **yes, 1.75×** |
| `cc_cone_proper_subset` | 50 | 1,011 | 793 | 1,011 | **772** | **yes, 1.31×** |
| `cc_cone_proper_subset` | 100 | 3,926 | 3,963 | 3,926 | **3,863** | **yes, 1.02×** |

**`T_guarded < T_none` now holds on 9/12 points — the guard's
contribution is measured for the first time**, on every construction with
a sibling branch (`cc_sibling_emptycone`, `cc_both`,
`cc_cone_proper_subset`), at every scale point measured. `cc_cone_only`
(no sibling, by design) correctly shows no change — this is not a
counterexample, it is the control.

**The margin shrinks with scale**, from ~1.7–1.75× at n=20 to ~1.02× at
n=100, consistently across all three sibling-bearing programs — the
declined portion's cost is roughly fixed per program while the
sibling's own full computation (previously the entire benefit) grows
with `n`, so the ABSOLUTE tuples saved stay roughly constant while the
total grows around it. Reported as a real, measured trend, not smoothed
into a single number: the guard's contribution here is real and
consistent in direction, modest and scale-shrinking in magnitude, unlike
the M4-SIPS negation relaxation's growing-with-scale contribution
(`docs/reports/m4-sips.md`).

Blast radius (culprit/cone/declined sets) is **unchanged** by P1
(confirmed by re-running `bin/conecheck` on all 4 `CONE_CORPUS`
programs): P1 only changes which NON-declined relations get properly
adorned instead of left Untouched, which is orthogonal to the
culprit-cycle/cone computation itself.

## Headline re-run

`harness/night_m3_5_headline_m4.py`, full 32 points, re-run against the
post-P1 binary. **Byte-identical to the pre-P1 run** — diffed directly
against the pre-P1 snapshot already committed at `3814f3a`
(`git show 3814f3a:measurements/m3-5-headline-m4/run.stderr.txt`),
clean, zero differences — none of the 12
`BENCHMARK_FAMILY`/`CULPRIT_CANDIDATES` programs has a second `.output`
branch, so P1 changes nothing here, exactly as predicted before running
it. (The re-run's own output overwrites `measurements/m3-5-headline-m4/`
in place, byte-for-byte identical to what it replaces — no separate
"before" directory kept; git history at `3814f3a` is the record.)

## Explain-sample regenerated

`docs/reports/explain-samples/transform_and_guard_cone_proper_subset.explain.txt`
(task E) is committed evidence of specific tool behavior and would have
gone stale — regenerated against the post-P1 binary: `tc`/`direct` now
show `ADORN rel=tc_bf`/`ADORN rel=direct_bf` and `QUERY pred=tc`/
`QUERY pred=direct` lines instead of `UNTOUCHED`, an independent
third confirmation (after the gate and the headline re-run) of the exact
same fix, via `dlc explain`'s own code path rather than the measurement
harness's. The other two samples touching `magicset`/`guard`
(`transform_and_guard_culprit_cycle`, `transform_only_ancestor_nonancestor`)
were also regenerated and are byte-identical to before (single-query
programs, unaffected — confirmed, not assumed).

## Verdict

**P1: DONE.** Seed collection, not a new algorithm, confirmed by the
gate: 9/9 answer-identical (hard requirement met), 32/32 headline points
unchanged, and task B's central open finding reversed —
`T_guarded < T_none` now holds on 9/12 points, with the margin and its
scale-dependence reported as measured. `docs/OPEN_QUESTIONS.md`'s
"`magicset.FindQuery` seeds only one query per program" entry is now
resolved; not deleted (append-only), a resolution note is appended to it.
