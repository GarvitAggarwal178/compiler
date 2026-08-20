# Session log

Append-only, one entry per working session.

## 2026-08-20 — Phase 0 probe

**What changed:** Repo laid out per CLAUDE.md §3 (`docs/`, `harness/`, `tests/`,
`fixtures/`, `measurements/`, `tools/`; no `src/`, per instructions). `architecture.md`
moved to `docs/dlc-blueprint.md`. Git repo initialized. Soufflé 2.5 installed from
release `.deb` (`docs/DECISIONS.md`). Wrote `harness/probe0.py` (seeded fixture
generation for P1/P2/P3, verifies P1's reachable-from-1 set before running anything,
drives Soufflé with and without `--magic-transform=*`, profiles each run),
`harness/parse_profile.py` (exact per-relation tuple counts from Soufflé's JSON
profile log, since `souffleprof`'s text table rounds), `harness/extract_exact.py`
(runs the parser against all six logs with full provenance). Wrote the three `.dl`
programs verbatim from blueprint §12 to `tests/programs/`.

**Measurement IDs produced:** `probe0-p{1,2,3}-{off,on}-{run,profile,extract}` (18
directories), each self-contained under `measurements/<id>/` with
`cmd.txt`/`stdout.txt`/`stderr.txt`/`env.txt`/`meta.json`, plus Soufflé's own raw
output (`.csv`, `.log`) for the `-run` IDs. Full table in `docs/MEASUREMENTS.md`.

**What is now blocked:** Phase 1. Three CLAUDE.md §5 STOP conditions fired: Soufflé's
`--magic-transform=*` transformed relations containing negation on both P2 and P3
(contradicting the blueprint's stated premise and its own worked STOP example);
`T_magic(P1) > T_base(P1)`; P1's magic/no-magic ratio (≈1) misses the blueprint's
stated ~10^3 expectation by more than one order of magnitude. Full escalation with
three live explanations and one proposed (unrun) distinguishing experiment
(`--inline-exclude=q`) is in `docs/reports/probe0.md`.

**Single next action:** Human decides among the three live explanations in
`docs/reports/probe0.md` (Soufflé version drift / Soufflé already guard-equivalent /
P3-specific inlining artifact) and whether to authorize the `--inline-exclude=q`
follow-up experiment before Phase 1 starts.

## 2026-08-20 — Phase 0.5: corrected probes, blueprint amended to v1.1

**What changed:** Applied the human's ruling on `docs/reports/probe0.md` (received as
a "Phase 0.5" directive). Amended `docs/dlc-blueprint.md` to v1.1: differentiator
restated (§2 — Soufflé restricts the negating relation, never the negated one),
guard clause (b) promoted to primary metric source and (a) demoted to a correctness
side-condition (§6), the "positive cycle" precondition flagged unverified (§6),
primary headline metric redefined around negated-relation materialization (§7),
Soufflé Prior Art Register entry corrected against observed 2.5 behavior (§11), P1
replaced by P1' and P3 replaced by P5, P4 added as a decisive hand-transform
experiment (§12), failure mode #6 added (§9). Wrote `tests/programs/p1prime.dl` and
`tests/programs/p4.dl`; ran both against the *existing* P1/P2 fixtures, unregenerated.
Wrote `harness/probe0_5.py` (reuses `probe0.py`'s run helpers). `docs/MEASUREMENTS.md`
gained a `convention` (`excl-copy`/`incl-copy`) treatment per the ruling in §3.

**Measurement IDs produced:** `probe0.5-p1prime-{off,on}-{run,profile,extract}`,
`probe0.5-p1prime-diff`, `probe0.5-p4-{run,profile,extract}`, `probe0.5-p4-diff`,
`probe0.5-p2-bfs-check`. Full table in `docs/MEASUREMENTS.md`.

**Result:** P1' ratio ≈1.51×10⁴ (predicted ~1.5×10⁴) — confirms the `.output path`
mechanism exactly. P4's `q2.csv` byte-identical to P2's in both configurations;
`reach_bf`=170 vs Soufflé's unrestricted `reach`=26,404 (BFS-cross-checked; the "≈50"
prediction assumed P1's engineered fixture, P2's fixture has no such construction —
logged, not acted on). Full writeup: `docs/reports/probe0_5.md`.

**What is now blocked:** Nothing — Phase 0.5 answered its four questions and the
escalation is resolved by the human ruling, not by further investigation.

**Single next action:** Human decides whether to start Phase 1 (M1: lexer, parser,
type/allowedness checks, naive + semi-naive fixpoint, Lane A) or to run P5 first as an
M3 pre-check. `--inline-exclude=q` is now a stated P5 prerequisite
(`docs/OPEN_QUESTIONS.md`), not a standalone experiment to schedule separately.

## 2026-08-20 — Phase 0.6: P4' fix, P6 counterexample hunt, Q5 pre-registration

**What changed:** Applied the Phase 0.6 directive. Blueprint amended to v1.2
(`docs/dlc-blueprint.md`): three-column headline metric (`T_none`/`T_souffle`/
`T_guard`, contribution = `T_souffle/T_guard`, `T_none/T_guard` prohibited as
headline), clause-(a)/(b)-independence note added to §6, Q5 deadline moved to now
(§10), M1/Phase-0.6-parallel note added to §8. Fixed `p4.dl`'s query-constant-in-
rule-head bug as `tests/programs/p4prime.dl`. Ran a bounded (~1hr) counterexample
search for P6 across five constructions (`tests/programs/p6start_*.dl`,
`p6a1_*.dl`, `p6a1b_*.dl`, `p6a2_*.dl`). Pre-registered the M1-M3 corpus: wrote
`harness/corpus_predicate.py` + `harness/build_corpus.py`, sparse-checked-out
Soufflé 2.5's `tests/` tree (not vendored), applied the predicate, committed
`tests/corpus/PREREGISTERED.txt` + `detail.json` + `SOURCE.md`.

**Did not touch `src/` or write any Lane A code.** M1 (lexer, parser, type/
allowedness checks, naive + semi-naive fixpoint) is the human's work, running in
parallel per §8 v1.2 — this session's scope was Phase 0.6 (Lane B) only.

**Measurement IDs produced:** `probe0.6-p4prime-{run,profile,extract,diff}`,
`probe0.6-p6start-{base,hand}-{run,profile}`, `probe0.6-p6a1-{base,hand-naive,
hand}-{run,profile}`, `probe0.6-p6a1b-{base,hand-naive,hand}-{run,profile}`,
`probe0.6-p6a2-{base,hand}-{run,profile}`, `probe0.6-q5-eval-only`,
`probe0.6-q5-corpus`. Full table in `docs/MEASUREMENTS.md`.

**Result:** P4' confirmed sound (byte-identical to P2 and to the original buggy P4);
`@neglabel.reach` confirmed `REC_T` (genuine re-derivation), not `COPY_T`. P2's
three-column table: 40,030 / 26,465 / 231 — contribution is 114.6×
(`T_souffle/T_guard`), not the 173× a `T_none/T_guard` framing would have claimed.
P6: no counterexample found in five bounded attempts (one initially degenerate,
fixed and rerun); working hypothesis is clause (b) collapses into correct seed
collection + clause (a), M3 candidate re-scope to ~2 weeks, pending human
confirmation. Q5: 36 of 612 tests pre-registered, full-tree scope (not the
narrower 11-of-149 first framing, which is disclosed not discarded). Full writeup:
`docs/reports/probe0_6.md`.

**What is now blocked:** Nothing on the Lane B side. M1 continues independently.

**Single next action:** Human confirms (or overturns) the M3 re-scope-to-~2-weeks
recommendation from the P6 result, and decides whether/when to run the 36-program
pre-registered corpus (not yet run — that's an M1/M2/M3-relevant future step, not
Phase 0.6's).
