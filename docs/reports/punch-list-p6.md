# PUNCH-LIST P6 — presentation artifact

Date: 2026-08-27. `harness/build_presentation.py` → `docs/reports/presentation.html`.
Per `docs/m2 m3.md` §10 and NIGHT-BATCH-04 G, unchanged: one static HTML
file, no server, no framework, no build step. Well under the two-hour cap.

## What did not work

Nothing did not render. The one deliberate scope cut: no JavaScript
beyond what the browser needs for anchor-link navigation (`<nav>` at the
top) — no tabs, no client-side filtering, no charts. The punch list asks
for four things visible, not an interactive tool; a longer build here
would be over-building exactly the thing §10 itself warns against
("if it exceeds two hours, ship what renders").

## What it shows

Computes nothing new — every table cell is read directly from an
already-committed file:

1. **Rejection.** All four grounds (arity, type, allowedness,
   unstratifiable negation), each a minimal `.dl` sample plus `dlc
   explain`'s own unmodified output, read from
   `docs/reports/explain-samples/`.
2. **Three-column metric.** All 5 `BENCHMARK_FAMILY` shapes, every scale
   point, `T_none`/`T_souffle`/`T_dlc` (incl-sup, the headline
   convention) plus the contribution ratio, answer-identity, and whether
   the guard fired — read from
   `measurements/m3-5-headline-m4/summary.json` (the post-PUNCH-LIST-P1
   headline, current).
3. **Guard firing and declining.** Per-program culprit SCC, cone, and
   declined fraction for all 16 corpus programs (`bin/conecheck`'s
   committed JSON, `measurements/night04-b-cone/blast_radius/`), plus the
   Go-vs-Python cone cross-check table (4/4 exact agreement, read from
   `measurements/night04-b-cone/summary.json`).
4. **Two findings with mechanisms.** The `bb`→`bf` adornment-set
   before/after table for all 4 negation shapes (read directly from
   `measurements/m4-sips/before/` and `.../after/`'s emitted programs,
   not re-typed), and the cone-behaviour `T_guarded` vs. `T_none` table
   across all 12 task-B points, current post-P1 numbers.

## Gate

`python3 harness/build_presentation.py` runs clean, writes
`docs/reports/presentation.html` (23 KB, single file, no external
references). Balanced-tag check passed (30 opening / 30 closing
`table`/`div`/`section` tags). Spot-checked every rendered number against
`docs/reports/FINAL.md`'s own citations — all match (e.g.
`reachability_complement` n=8000: `T_none=64,001,614`, contribution
`1,342.7×`, identical in both documents).

## Verdict

**P6: DONE.** First cut, ships as specified. Re-running
`harness/build_presentation.py` regenerates the page from whatever is
currently committed — no manual editing of the HTML is expected or
needed as measurements evolve further.
