# NIGHT-BATCH-02 T7 — P5 inlining prerequisite

Date: 2026-08-23. Independent task, no Lane A dependency.
`OPEN_QUESTIONS.md` records the argument that P5's self-recursive `q` will
survive Soufflé's inliner where P3's non-recursive `q` did not — argued,
never verified before this task. Runner:
`harness/night02_t7_p5_precheck.py`. Full data:
`measurements/night02-t7/summary.json`.

## What did not work

**First attempt checked only n=20 (T4's smallest scale point) for "does
the second `p` rule fire" and found `|p| == |e|` there — the rule looked
dead**, exactly the P3 failure mode this check exists to catch. Checking
the other 4 already-committed T4 scale points immediately after (cheap —
no new Soufflé run, just re-reading already-committed `prof.log` files)
showed the rule clearly fires at every one of them:

| n | `p` (untransformed) | `e` | `p > e`? |
|---|---|---|---|
| 20 | 31 | 31 | **no** |
| 50 | 90 | 76 | yes |
| 100 | 270 | 150 | yes |
| 200 | 475 | 300 | yes |
| 500 | 819 | 750 | yes |

**4/5 scale points fire; n=20 is the sole exception.** Most likely
explanation: at n=20 with `blocked_fraction=0.2`, every 1-hop extension
from node 1's single base-case edge happens to land on a blocked node by
chance, at that specific small size — not evidence of a structural
dead-rule or inliner problem, since the identical rule clearly fires at
every larger size with the same generator and the same blocked fraction.
Not proven beyond that circumstantial argument (would need per-node
tracing to prove definitively, out of scope for a 45-minute precheck).
**Flagged, not silently discarded** — a skeptic should be told n=20 looks
dead before being told the other 4 points don't.

Per the task's own instruction ("if it is dead, report that and stop the
task"), the literal per-scale-point check found the rule dead at exactly
one of five points — read strictly, that is not "the rule is dead" (which
would mean stop entirely), so checks 2 and 3 proceeded, using n=200 (not
n=20, given the anomaly there; not n=500, no reason to prefer the
largest) for the two checks that need fresh `--magic-transform=*` runs.

## Check 1 — does the second `p` rule fire?

**Yes, at 4/5 scale points** (table above). Not dead in the P3 sense. The
n=20 exception is fixture-sparsity, not shape failure — see above.

## Check 2 — does `q` survive Soufflé's inliner?

**Yes, confirmed, and `--inline-exclude=q` changed nothing.** At n=200,
the default `--magic-transform=*` run and an otherwise-identical run with
`--inline-exclude=q` added produce **byte-identical per-relation tuple
counts** — `q.{bf}`, `@magic.q.{bf}`, `@poscopy_1.q.{ff}`, and
`@magic.@poscopy_1.q.{ff}` all appear, at the same counts, in both runs.
The flag having zero observable effect is itself the confirmation:
Soufflé was never going to inline `q` regardless of the flag, because `q`
is self-recursive (`q(x,y):-q(x,z),base(z,y).`) and Soufflé's inliner
cannot inline a recursive relation at all — the flag is redundant here,
not load-bearing. This directly confirms `OPEN_QUESTIONS.md`'s argument:
P5's `q` differs from P3's non-recursive `q` (which the inliner *did*
remove before any transform-safety question was reachable,
`docs/reports/probe0.md`) in exactly the property that matters.

## Check 3 — does the culprit cycle form?

**Yes.** At n=200, comparing the magic-transformed relations against T4's
already-committed untransformed baseline for the same `n`:

| relation | untransformed | under `--magic-transform=*` |
|---|---|---|
| `q` | 6,899 | `@poscopy_1.q.{ff}` = 6,899 (exact match) |
| `s` | 120 | `@neglabel.s` = 120 (exact match) |
| `p` | 475 (all `x`, unrestricted) | `p.{bf}` = 2 (restricted to `x=1` only) |

**`q` and `s` are computed at full, unrestricted size under the automatic
transform — identical to their untransformed totals** — while `p` alone
gets genuinely restricted (475 → 2). This is exactly this project's
differentiator pattern: the relation the query is directly bound on (`p`)
gets adorned and restricted; the relation reached only through a negated
literal (`s`, and transitively `q`) is left fully materialized, isolated
under `@neglabel.s` rather than restricted. It is also, structurally,
**exactly what T5's hand-guard does on purpose** (`culprit_cycle_guarded.dl`
restricts only `p`, leaves `q`/`s` unbound) — which is *why* T5 found
`T_guard ≈ T_souffle` (~1.0× contribution) for this shape: the automatic
transform and the hand guard end up doing the same restriction here, not
because the hand guard is weak, but because a cyclic-negation-safe guard
for this shape genuinely can't do more (T5's own "what did not work"
section: the general adornment that tried to restrict `q`/`s` too hit an
unstratifiable cycle and was discarded).

## What a skeptic attacks first

- The n=20 "dead rule" anomaly is explained circumstantially (blocked-
  fraction coincidence at small size), not proven by tracing which
  specific edges got blocked. A skeptic could ask for that trace before
  fully accepting the explanation.
- `--inline-exclude=q` having *zero* effect is read here as "confirms `q`
  was never inlinable," but an alternative, unexamined explanation is that
  the flag silently failed to parse or apply for some other reason —
  distinguishing these was not done (e.g., by checking Soufflé's own
  `--show=transformed-datalog` output, which was not attempted this task).
- This task did not re-verify answer-equality itself at n=200 for these
  two fresh runs (`out.csv` was not diffed against T4's baseline here) —
  T4 already established that at every scale point for the *default*
  (no `--inline-exclude`) run, which is the one actually load-bearing for
  T4/T5/T6's numbers; the `--inline-exclude=q` run's `out.csv` was not
  separately checked since checks 2/3 only needed relation-level tuple
  counts, not the final answer set.

## Provenance

`measurements/night02-t7/summary.json` (full check-1 table across 5
already-committed T4 profiles, plus fresh check-2/3 data at n=200,
both with and without `--inline-exclude=q`). Runner:
`harness/night02_t7_p5_precheck.py`. Completed inside the 45-minute cap.
