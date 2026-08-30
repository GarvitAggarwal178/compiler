# Phase 0.5 report

Date: 2026-08-20. Response to the rulings in the Phase 0.5 directive on
`docs/reports/probe0.md`. Blueprint amended to v1.1 (`docs/dlc-blueprint.md`).
Fixtures reused unchanged from Probe 0 (`fixtures/p1`, `fixtures/p2`; SHA-256 not
re-verified since the generator was not re-run — same files on disk).

## 1. P1' — `T_base`, `T_magic`, ratio

| | Relations (exact) | Total |
|---|---|---|
| `T_base` (magic off) | `path`=1,525,746, `q`=50 | **1,525,796** |
| `T_magic` (magic on) | `path.{bf}`=50, `q`=50, `@magic.path.{bf}`=1 | **101** |

Ratio `T_base / T_magic` = **≈1.51 × 10⁴**. No `@interm_out.path.{ff}` relation
appears at all in the magic-on run — without `.output path`, Soufflé never computes
the unrestricted `path`; it only ever computes `path.{bf}`. Matches the §2.1
prediction (`T_magic ≈ 10²`, ratio ≈1.5×10⁴) within rounding.

Measurement IDs: `probe0.5-p1prime-off-{run,profile,extract}`,
`probe0.5-p1prime-on-{run,profile,extract}`.

## 2. P4 vs P2 — `q2.csv` byte-identical?

**Yes.** `diff -q` reports zero differences against both P2 configurations:
P4's `q2.csv` == `probe0-p2-off-run/q2.csv` == `probe0-p2-on-run/q2.csv`.

Measurement IDs: `probe0.5-p4-run`, diff commands run directly (not separately
logged as a measurement ID — same convention as `probe0-p2-diff` in
`docs/MEASUREMENTS.md`, added there as `probe0.5-p4-diff`).

## 3. P4 — `reach_bf` vs Soufflé's `reach`

**(170, 26,404).** `reach_bf.total` = 170 (`probe0.5-p4-extract`); Soufflé's `reach`
= 26,404, unchanged in both P2 configurations (`probe0-p2-off-extract`,
`probe0-p2-on-extract`). Ratio ≈155×, not reported as the headline per §3's
noise-floor rule — both sides are well above 10³ except `reach_bf`, so the pair is
the honest statement, not the ratio.

Note: the §2.2 prediction was "expect ≈ 50," by analogy with P1's engineered
core-size-50 fixture. P2's fixture (`build_p2_fixture`, `harness/probe0.py`) was
never built with a bounded-reachability core — it is a plain random digraph
(200 nodes, 400 edges) — so there was no reason to expect 50 specifically.
Independently cross-checked with a plain-Python BFS over `fixtures/p2/edge.facts`:
171 nodes reachable from node 1 including node 1 itself; `reach_bf` (no reflexive
base case) reports 170 — consistent, not an anomaly.

## 4. Invariant that held and was not predicted

The `{bf}`-adorned restricted computation is **bit-identical** between P1 (v1.0,
`.output path` present) and P1' (v1.1, `.output path` removed): seed=2, 4 semi-naive
iterations, delta_sum=48, total=50 — same numbers in `probe0-p1-on-extract`
(`@interm_out.path.{bf}`) and `probe0.5-p1prime-on-extract` (`path.{bf}`). Removing
the `.output` declaration changed nothing about how the restricted side was computed;
it only deleted the redundant unrestricted `{ff}` side. This was assumed, not
verified, when the P1' fix was proposed — it is now confirmed by direct comparison
rather than by argument.

Secondary invariant, carried over from Probe 0 and reconfirmed here rather than
re-argued: Soufflé's isolated negated relations reproduce their untransformed
baseline count **exactly**, not approximately (`@neglabel.reach` = 26,404 =
untransformed `reach`; `@neglabel.s` = 19 = untransformed `s`, both from
`docs/MEASUREMENTS.md`). The isolation is a rename under a new evaluation label, not
an independent re-derivation that happens to converge nearby.

---

Four answers. Stopping per instruction. Blueprint v1.1 committed alongside this
report; `docs/DECISIONS.md` and `docs/OPEN_QUESTIONS.md` updated with the rulings and
their measurement IDs.
