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
