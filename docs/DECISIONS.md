# Decisions

Append-only. One line per irreversible choice, with reason.

| Date | Decision | Reason |
|---|---|---|
| 2026-08-20 | Installed Soufflé 2.5 from the `x86_64-ubuntu-2404-souffle-2.5-Linux.deb` release asset on an Ubuntu 26.04 WSL2 host (no 26.04 build published) | Closest available build; `apt --fix-broken install` resolved `libsqlite3-dev`/`mcpp` from `universe` cleanly, `souffle --version` confirms 2.5 running |
| 2026-08-20 | `architecture.md` moved verbatim to `docs/dlc-blueprint.md` | Matches the layout CLAUDE.md §3 specifies; content untouched, this is a relocation not an edit |
| 2026-08-20 | Repo initialized as a git repository (was not one) | CLAUDE.md's provenance/commit discipline (§0.2, §8) assumes version control; no prior VCS state existed to disturb |
| 2026-08-20 | Exact per-relation tuple counts sourced from Soufflé's own JSON profile log (`root.program.relation.<name>.num-tuples`, summed with per-iteration deltas for recursive relations), not the rounded `souffleprof -c rel` text table | `souffleprof` text rounds to 3 significant figures ("1.52M"); CLAUDE.md forbids numbers without exact provenance. Cross-validated against `wc -l` on `.output` CSVs where available — agreed exactly (`probe0-p1-off-extract` vs `wc -l path.csv` = 1,525,746 both) |
| 2026-08-20 | `T_magic` sums exclude `COPY_T`-labeled relations (e.g. `path` in P1's magic-on run, a copy of `@interm_out.path.{ff}`) from the derived-tuple total | A copy is not a join-driven fixpoint derivation; flagged as needing a real definition before M2 (`docs/OPEN_QUESTIONS.md`), not treated as settled |
| 2026-08-20 | Phase 0 stopped after producing the six required answers; Phase 1 not started | Three independent CLAUDE.md §5 STOP conditions fired (blueprint-contradicting behavior on P2/P3, `T_magic > T_base` on P1, >1 order of magnitude deviation from the stated P1 ratio expectation); see `docs/reports/probe0.md` |
| 2026-08-20 | P1 declared defective; `.output path` removed | Required output forces full materialization; delta of 52 = 50 + 2 seeds confirms the mechanism exactly (superseding decision on row above re: `COPY_T` exclusion — the defect was `.output path`, not the transform) |
| 2026-08-20 | P3 declared void; replaced by P5 | Inlining removed `q`, and the recursive rule was semantically dead independently |
| 2026-08-20 | Differentiator restated | Soufflé restricts negating relations but never negated ones; the real gap is completeness-guarded restriction of negated relations |
| 2026-08-20 | Guard clause (b) promoted to primary, (a) demoted to side-condition | (b) is where the measurable cost of Soufflé's conservatism lives |
| 2026-08-20 | Metric: report both copy-relation conventions in every row | Cheaper than defending a convention; supersedes the 2026-08-20 "`T_magic` sums exclude `COPY_T`" row above, which is left standing per the append-only rule |
| 2026-08-20 | Version-drift hypothesis rejected | 2.5 is current (2025-03-25) and its docs still state the skip behaviour |
| 2026-08-20 | `docs/dlc-blueprint.md` bumped to v1.1; §2, §6, §7, §9, §11, §12 amended per the Phase 0.5 directive | Amendments were fully specified by the human ruling on `docs/reports/probe0.md`; applying them is transcription of an already-made decision, not independent resolution of the Lane A conflict CLAUDE.md's preamble reserves for a human |
| 2026-08-20 | P4/P5 written as new `.dl` test programs (`tests/programs/p4.dl`, blueprint §12 text for P5) rather than as `src/` code | Both are fixture/oracle programs executed by Soufflé, not `dlc` implementation code; Lane B per CLAUDE.md §2 ("test bodies anywhere, including tests targeting Lane A code") |
