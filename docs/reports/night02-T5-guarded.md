# NIGHT-BATCH-02 T5 — hand-guarded transforms: `T_guard` (headline)

Date: 2026-08-23. Prerequisite T4 (passed, all 4 negation-bearing shapes had
`E_recoverable > 0`). `transitive_closure_bound` is excluded: T4 measured
`E_recoverable = 0` there (no negation in the shape), so there is nothing
for a guard to recover. Guarded sources:
`tests/corpus/BENCHMARK_FAMILY/guarded/*.dl` (4 files, one per shape).
Runner: `harness/night02_t5_guarded.py`. Full data:
`measurements/night02-t5/summary.json`.

**Answer equality against the T4 untransformed baseline held at every
scale point, every shape — no abort fired.**

## What did not work

**`culprit_cycle`'s first-attempt general adornment was rejected by
Soufflé.** Following the same magic-propagation pattern used successfully
for `ancestor_nonancestor` (propagate the seed through the relation's own
recursion), the natural derivation restricts `q` and `s` via a magic
relation `m_q(z) :- m_p(x), p_bf(x,z)`. Soufflé rejects the resulting
program outright:

```
Error: Unable to stratify relation(s) {m_q,p_bf,q_bf,s_bf}
...has cyclic negation...
p_bf(x,y) :- m_p(x), p_bf(x,z), !s_bf(z), q_bf(z,y).
```

`m_q` depends positively on `p_bf` (through `p_bf`'s own recursion), and
`p_bf` depends negatively on `s_bf`, which depends on `m_q` — a negative
cycle. **This is not a bug; it is P5's reason for existing** (`docs/
dlc-blueprint.md` §12, and `culprit_cycle.dl`'s own header comment) — the
general magic-set derivation recreates, on its own, the exact cycle the
shape was constructed to exercise. Per requirement 2 ("a transform violating
[stratifiability / answer equality] is discarded, not fixed up"), this
derivation was discarded. Kept as scratch evidence, not a T5 deliverable:
`measurements/_scratch_night02_t5/culprit_cycle_unsafe_cyclic.dl`. The
committed guard (`culprit_cycle_guarded.dl`) instead restricts only `p`
(positively, acyclically reachable from the query) and leaves `q`/`s` fully
unbound — see that file's header comment for the full account.

**Harness gotcha found while confirming the rejection above: Soufflé
exited 0 on this stratification error**, with `Error:` on stderr and no
output file produced (confirmed: `out.csv` did not exist in the run
directory). `harness/night02_t5_guarded.py`'s `run()` now checks
`"Error:" in stderr` in addition to the return code — return code alone
would have silently treated a rejected program as a successful, empty-
answer run. Not exercised by T4 (nothing there was rejected), so this did
not surface until T5.

## Per-shape tables

`T` values are excl-copy (no `COPY_T` relation appeared in any T5 run,
consistent with T4). Contribution = `T_souffle / T_guard`, never
`T_none / T_guard` (headline-ratio prohibition, CLAUDE.md §7). Never
aggregated across shapes — see per-shape numbers only.

### `same_generation_negation`

`sg ∪ notsg` partitions `person × person`, same pattern as P2 — `T_none`
is definitional here, not an arbitrary baseline.

| scale point | `T_none` | `T_souffle` | `T_guard` | `T_souffle/T_guard` |
|---|---|---|---|---|
| depth=4 (226 persons) | 51,301 | 30,907 | 452 | 68.4× |
| depth=5 (840 persons) | 706,439 | 421,563 | 1,680 | 250.9× |
| depth=6 (4,888 persons) | 23,897,431 | 14,352,999 | 9,776 | 1,468.2× |
| depth=7 (14,567 persons) | 212,212,055 | 127,332,304 | 29,134 | 4,370.6× |

**This ratio needs explaining, not just reporting** (CLAUDE.md §5: results
that look better than expected get more scrutiny, not less). The guarded
`sg_bf` collapses almost to nothing because of a structural fact of this
family's fixture, not because the transform is unusually powerful in
general: the query is bound at node 0, which `gen_family_tree` always makes
the tree *root* — root 0 has no `parent` edge of its own (confirmed:
`grep -c '^0<TAB>' fixtures/.../parent.facts` returns 0 matches, i.e. 0
never appears as a `child` in any generated tuple). `sg`'s recursive rule
requires `parent(x,xp)` to advance — with `x=0` that literal is never
satisfiable, so `sg_bf(0,y)` can only ever be produced by the base case
(`sg_bf(0,0)`). `notsg_bf(0,y)` is therefore true for every other person and
`q_notsg` has exactly `n_persons - 1` rows — independently verified against
T4's own `q_notsg.csv` row counts (225 at depth 4 = 226 persons − 1) before
this guard was written, not after. The guard is general (it does not
hardcode this fact — the propagation rule `m_sg(xp):-m_sg(x),parent(x,xp)`
would extend `m_sg` past `{0}` for a fixture where the query node had
ancestors); it is this specific, pre-registered fixture choice (root as the
query constant) that makes the recursive branch structurally dead. The
`same_generation_negation_guarded.dl` header comment also flags the
independent, better-known reason magic sets are usually *weak* on
same-generation queries: the free variable `y` is never bound by the
adornment (`parent(y,yp)` enumerates unconditionally) — that difficulty is
real and would show up on a fixture where the query node were not the root.

### `ancestor_nonancestor`

`ancestor ∪ nonancestor` partitions `person × person`; `T_none` is
definitional here too.

| scale point | `T_none` | `T_souffle` | `T_guard` | `T_souffle/T_guard` |
|---|---|---|---|---|
| n=500 | 250,450 | 101,250 | 25,500 | 4.0× |
| n=1,000 | 1,000,950 | 322,306 | 51,000 | 6.3× |
| n=2,000 | 4,001,950 | 1,550,117 | 102,000 | 15.2× |
| n=4,000 | 16,003,950 | 6,663,717 | 204,000 | 32.7× |
| n=8,000 | 64,007,950 | 25,391,507 | 408,000 | 62.2× |

The most modest contribution of the four guarded shapes, and the only one
whose ratio grows roughly linearly with `n` rather than accelerating —
worth carrying into T6's cross-shape comparison rather than treated as an
outlier to explain away.

### `culprit_cycle` (P5 shape) — the partial-restriction guard

No partition structure (`p`/`out` has no complement relation); `T_none` is
a plain baseline count here, not definitional.

| scale point | `T_none` | `T_souffle` | `T_guard` | `T_souffle/T_guard` |
|---|---|---|---|---|
| n=20 | 286 | 272 | 257 | 1.1× |
| n=50 | 510 | 432 | 422 | 1.0× |
| n=100 | 2,632 | 2,423 | 2,366 | 1.0× |
| n=200 | 7,496 | 7,028 | 7,024 | 1.0× |
| n=500 | 84,105 | 83,295 | 83,290 | 1.0× |

**Essentially no contribution.** Consistent with the guard's own design:
only `p` is restricted; `q` and `s` (which `p` depends on, and which the
general/unsafe derivation could not safely restrict — see above) are left
fully unbound, identical in cost to the untransformed baseline. `T_guard`
tracks `T_souffle` closely because both leave the dominant `q`/`s`
computation untouched — Soufflé's own automatic transform gets essentially
the same result here without needing a hand guard at all. Whether this
shape's rules even fire meaningfully on this fixture (dead-rule risk, the
way P3 failed) is T7's question, not answered here.

### `reachability_complement` (P2 lineage)

`reach ∪ unreach` partitions `node × node`; `T_none` is definitional, as it
was for the original P2 (`docs/reports/night01-T6-scaling.md`).

| scale point | `T_none` | `T_souffle` | `T_guard` | `T_souffle/T_guard` |
|---|---|---|---|---|
| n=250 | 62,534 | 44,811 | 285 | 157.2× |
| n=500 | 250,093 | 168,759 | 594 | 284.1× |
| n=1,000 | 1,000,193 | 611,806 | 1,194 | 512.4× |
| n=2,000 | 4,000,461 | 2,471,792 | 2,462 | 1,004.0× |
| n=4,000 | 16,000,877 | 9,812,420 | 4,878 | 2,011.6× |
| n=8,000 | 64,001,614 | 40,805,441 | 9,615 | 4,243.9× |

Every `T_guard` value here is identical to NIGHT-BATCH-01 T6's already-
committed numbers at the same `n` (`docs/reports/night01-T6-scaling.md`) —
this guard is not a new derivation, it is `p4prime.dl` (already validated
across all 6 of T6's points) ported onto this family's relation names.

## What a skeptic attacks first

- `same_generation_negation`'s 4,370× ratio at depth=7 is the largest number
  in this report by far, and it is explained by a fixture-structural
  accident (query node = tree root, which happens to have no parent edge),
  not by a generally-superior transform. A fixture where the query node
  were a leaf or mid-tree node would very likely show a much smaller ratio,
  possibly closer to `ancestor_nonancestor`'s modest 4×–62× — this was not
  tested (the scale points are pre-registered and root=0 is fixed by
  `SCALE_POINTS.json`'s own note; changing it is out of scope for this
  batch).
- `culprit_cycle`'s ~1.0× ratio could be read two ways: "the guard correctly
  found nothing safe to restrict beyond `p`" (this report's reading) or
  "the guard is under-powered and a cleverer safe derivation exists that
  this session didn't find." Only one general derivation was attempted and
  it hit the negative cycle; no second attempt was made.
- The `culprit_cycle_guarded.dl` restriction of `p` alone was validated for
  answer-equality but its *soundness argument* (why restricting `p`'s first
  argument via `m_p` cannot itself interact unsafely with the `!s(z)` inside
  `p`'s own rule) rests on `p` being invariant in its first argument across
  its own recursion — stated in the file's header comment, not proven here
  beyond the empirical answer-match at all 5 scale points.
- The Soufflé-exits-0-on-stratification-error behavior (found this task) was
  observed on exactly one program; not systematically explored (that would
  be T9's job, the diagnostic catalogue, later in this batch).

## Provenance

`measurements/night02-t5/summary.json` (20 scale-point rows across 4
shapes, all `answers_identical_vs_t4_baseline: true`, no DNF, no abort).
Guarded sources: `tests/corpus/BENCHMARK_FAMILY/guarded/*.dl`. Scratch
(discarded) attempt: `measurements/_scratch_night02_t5/
culprit_cycle_unsafe_cyclic.dl`. Runner: `harness/night02_t5_guarded.py`.
Completed inside the 150-minute cap.
