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

## 2026-08-20 — `--inline-exclude=q`: demoted from experiment to M3 prerequisite

Per the Phase 0.5 ruling, this is no longer "the cheapest experiment to run next" —
it's a stated prerequisite of the P5 culprit-cycle program (`docs/dlc-blueprint.md`
§12, v1.1): P5's `q` is self-recursive (`q(x,y):-q(x,z),base(z,y).`), which the
Probe 0 evidence suggests Soufflé's inliner would not collapse the way it collapsed
P3's non-recursive `q(x,y):-base(x,y).` — but this is not yet verified, only
argued. Verify it when P5 is run (M3), not before. Note from Soufflé's docs:
`--magic-transform-exclude` already implies `--inline-exclude` for the named
relations, so P5 may not need the flag explicitly if `--magic-transform-exclude` is
used instead.

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

## 2026-08-20 — Phase 0.5 resolution of the Q1/Q2 escalation

The human ruling on `docs/reports/probe0.md` reinterprets both entries above rather
than treating them as kills: Soufflé transforms the *negating* relation but never
demand-restricts the *negated* one (`@neglabel.<rel>` isolates and fully
materializes it). P4 (`docs/reports/probe0_5.md`) demonstrates by hand that
restricting the negated relation is both sound and large (170 vs 26,404 on the
negated relation) on this program. Blueprint bumped to v1.1
(`docs/dlc-blueprint.md`) with the differentiator restated around this. Version-drift
was checked and rejected as an explanation: Soufflé 2.5 (2025-03-25) is current and
its documentation still states the blanket-skip behaviour that the observed behaviour
contradicts.

## 2026-08-20 — metric definition: resolved as "report both conventions"

Superseding the entry above (kept, append-only): rather than pick `excl-copy` or
`incl-copy`, `docs/MEASUREMENTS.md` now reports both for any row containing a
`COPY_T` relation. P1' (v1.1) turns out to have zero `COPY_T` relations in its
magic-on run at all — removing `.output path` didn't just make the copy irrelevant,
it made Soufflé never compute the unrestricted relation the copy would have been
made from. The two conventions only diverge on P1 (v1.0, superseded) — supporting
evidence that the defect was `.output path`, not the transform in general.

## 2026-08-20 — P2's fixture was not built with a bounded-reachability core

The Phase 0.5 directive predicted `reach_bf ≈ 50` for P4 by analogy with P1's
engineered fixture (`gen_p1_graph`, core size 50). P2's fixture
(`build_p2_fixture`) is a plain random digraph with no such construction; measured
`reach_bf.total = 170` (BFS-cross-checked, `docs/MEASUREMENTS.md`
`probe0.5-p2-bfs-check`). Not acted on: no fixture change needed since the check P4
was built for (sound + large reduction on a negated relation) still holds at 170.
Worth remembering if a future probe wants a *specific* reachable-set size on a
negation program — reuse `gen_p1_graph`'s core/rest construction, not
`build_p2_fixture`'s.

## 2026-08-20 — no known program where Soufflé's negation-transform selectivity is wrong

Both negation probes here (P2, P3) have the negated relation depend on nothing the
magic seed restricts — Soufflé's silent selectivity happens to agree with a
guard-shaped notion of "safe" in both cases tested. Filter-1/thesis defense needs at
least one candidate program where that selectivity is either wrong (unsound) or
needlessly conservative (declines something safe); none has been found or
constructed yet. This is a design question for whichever corpus subdirectory Q5
resolves to, not something to invent inside Phase 0.
