# NIGHT-BATCH-01 — T6: P2 scaling sweep, three columns

**Mechanism characterization on a self-generated fixture, not corpus evidence.**

Date: 2026-08-20. Outcome: **done, complete.** All 6 sizes ran, no DNF, no abort —
answer relations (`q2.csv`, sorted set-equality per CLAUDE.md §6) identical across
all three configurations at every `n`.

## Method

`harness/night01_t6_scaling.py`. `n ∈ {250, 500, 1000, 2000, 4000, 8000}`, density
held at the existing P2 fixture's ratio (2 edges/node), seeded deterministically
per `n` (`BASE_SEED + n`, recorded in each `fixtures/p2-scale-<n>/meta.json`).
Three configurations per `n`: untransformed (`p2.dl`), `--magic-transform=*`
(`p2.dl`), hand-guarded (`p4prime.dl`, unmodified from Phase 0.6 — it already
generalizes to any `edge`/`node` fixture, that was the point of moving the query
constant out of the rule head). Global caps applied (300s, 8GB); none fired at any
size, including `n=8000`.

## The three-column table

| `n` | `T_none` | `T_souffle` | `T_guard` | `T_souffle/T_guard` | `E_recoverable` | answers identical |
|---|---|---|---|---|---|---|
| 250 | 62,534 | 44,811 | 285 | 157.2× | 44,742 | yes |
| 500 | 250,093 | 168,759 | 594 | 284.1× | 168,572 | yes |
| 1,000 | 1,000,193 | 611,806 | 1,194 | 512.4× | 611,419 | yes |
| 2,000 | 4,000,461 | 2,471,792 | 2,462 | 1,004.0× | 2,470,869 | yes |
| 4,000 | 16,000,877 | 9,812,420 | 4,878 | 2,011.6× | 9,810,665 | yes |
| 8,000 | 64,001,614 | 40,805,441 | 9,615 | 4,243.9× | 40,802,212 | yes |

Raw provenance: `measurements/night01-t6/summary.json`, per-`n`-per-config
directories `measurements/night01-t6/n<n>-{none,souffle,guard}/`.

## Reading the integers, not asserting a theory

Per-doubling-of-`n` growth factors, computed directly from the table above:

- `T_none`: ×4.0, ×4.0, ×4.0, ×4.0, ×4.0 (every step) — consistent with quadratic
  growth in `n`, as expected for unrestricted all-pairs-style reachability.
- `T_guard`: ×2.08, ×2.01, ×2.06, ×1.98, ×1.97 — consistent with linear growth in
  `n`, as expected for a computation restricted to one source node's reachable set
  in a fixed-density graph.
- `T_souffle`: ×3.77, ×3.63, ×4.04, ×3.97, ×4.16 — tracks `T_none`'s growth rate
  closely, not `T_guard`'s. Soufflé's own transform reduces the constant, not the
  growth order, on this fixture.
- `T_souffle/T_guard`: ×1.81, ×1.80, ×1.96, ×2.00, ×2.11 — itself growing at
  roughly the same per-doubling rate as `n` (×2), consistent with the ratio scaling
  **linearly in `n`**, not staying constant.

This is what the numbers show on **this self-generated, uncontrolled-reachability
fixture** (`gen_random_graph`, no engineered reachable-set size, unlike P1's
`gen_core_rest_graph`). It is not evidence about the pre-registered or exploratory
corpora, and it is not a proof that the ratio grows linearly in general — six points
on one fixture shape is a characterization, not a theorem.

## What did not work / caveats

- All three configurations at `n=8000` completed inside the 300s/8GB caps with
  room to spare — the sweep was never actually cap-limited, so "ascending order so
  a cap truncates the top" never had to matter. Larger `n` was not attempted;
  8,000 was the pre-specified ceiling, not a discovered one.
- `gen_random_graph` (P2's original construction) gives no control over the
  reachable-set size the way `gen_core_rest_graph` (P1's) does — the reachable set
  from node 1 at each `n` was not independently measured or held constant across
  sizes. The scaling curve above reflects whatever reachable-set size falls out of
  a uniform-random graph at each density and `n`, not a deliberately controlled
  variable. A future sweep wanting to isolate "does the ratio grow with `n` at
  fixed reachable-set size" would need `gen_core_rest_graph` instead
  (`harness/fixtures_lib.py`, built tonight per T7 item 4).
- `E_recoverable ≈ T_souffle` at every `n` here (the two numbers track within
  0.1–0.2%) — on this fixture almost all of what Soufflé computes under
  `--magic-transform=*` ends up inside the isolated `@neglabel.unreach`... actually
  `@neglabel.reach` (the negated relation) rather than the restricted `unreach`
  side. Consistent with Phase 0/0.5's P2 finding at `n=200` (`reach`=26,404 of
  `T_souffle`=26,465), now confirmed to hold across two orders of magnitude of `n`.
