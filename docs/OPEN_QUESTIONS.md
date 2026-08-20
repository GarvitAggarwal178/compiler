# Open questions

Append-only. Things noticed and not acted on, with the date noticed.

## 2026-08-20 — Q1 (blueprint §10), empirical answer

"Does Soufflé's magic transform still decline negation, empirically?" — **No, not on
either negation probe tested.** P2: `unreach` (body contains `!reach`) was
magic-transformed, `unreach.total` 13,596 → 30, answer bit-identical
(`probe0-p2-off-extract`, `probe0-p2-on-extract`, `probe0-p2-diff`). P3: `p` (body
contains `!s`) was magic-transformed with a real magic seed
(`@magic.p.{bf}(1)`), answer bit-identical (`probe0-p3-on-profile`,
`probe0-p3-diff`). See `docs/reports/probe0.md` STOP section — this is escalated, not
resolved here.

## 2026-08-20 — Q2 (blueprint §10), empirical answer

"What is the blast radius — one relation, or the whole SCC?" — **Zero relations
declined on P3.** `p` was transformed. `s` was relabeled (`@neglabel.s`) but computed
identically to the untransformed run. `q` never materializes in either run (see next
entry). Nothing in P3 was skipped by the magic-transform pass. See
`docs/reports/probe0.md` STOP section.

## 2026-08-20 — metric definition: what counts as a "derived" tuple across a copy

P1's magic-on run produces `path` as a `COPY_T` (not `NREC_T`/`REC_T`) of
`@interm_out.path.{ff}` (`probe0-p1-on-profile`). I excluded copies from the
`T_magic` sum on the grounds that a rename isn't a fixpoint-driven join, and said so
in `docs/MEASUREMENTS.md`. This needs a real, stated definition before M2's
headline metric depends on it — "derived tuple" should be pinned to something in the
semi-naive Δ-rewrite itself (Lane A), not to whatever Soufflé happens to label
`COPY_T` in a given build.

## 2026-08-20 — Soufflé inlines pass-through IDB relations before magic-transform runs

`q(x,y):-base(x,y)` in P3 never appears as a materialized relation in either the
magic-off or magic-on profile; post-transform rule bodies reference `base` directly
(`probe0-p3-on-profile`). This means any culprit-cycle argument that runs *through* a
relation like this (blueprint §6a's `magic_q →¬ s → q → magic_q`) can be silently
defused by Soufflé's own optimizer before the magic-transform pass ever sees the
program — independent of whether the guard concept is sound. `souffle` has
`--inline-exclude=<relations>` (confirmed via `souffle --help`, not yet used). Needs
resolving before any culprit-cycle program is added to a corpus: either use
`--inline-exclude` on Soufflé's baseline runs so both engines see the same IR, or
construct corpus programs where the negated-dependency chain can't be trivially
inlined away (e.g. give `q` a second use that blocks inlining).

## 2026-08-20 — no known program where Soufflé's negation-transform selectivity is wrong

Both negation probes here (P2, P3) have the negated relation depend on nothing the
magic seed restricts — Soufflé's silent selectivity happens to agree with a
guard-shaped notion of "safe" in both cases tested. Filter-1/thesis defense needs at
least one candidate program where that selectivity is either wrong (unsound) or
needlessly conservative (declines something safe); none has been found or
constructed yet. This is a design question for whichever corpus subdirectory Q5
resolves to, not something to invent inside Phase 0.
