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

(NIGHT-BATCH-01, 2026-08-20 to -21, is tracked separately in `docs/NIGHT_LOG.md`
and `docs/reports/night01-summary.md`, per the batch's own protocol — not
duplicated here.)

## 2026-08-21 — Corpus ruling: two corpora, T6 correction, subsumption resolved, OpenRuleBench blocked

**What changed:** Applied the corpus ruling (`docs/reports/
corpus-ruling-2026-08-21.md`, saved verbatim as the document of record —
`docs/phase0.7-corpus-viability.md` never landed, this supersedes that reference
everywhere). Blueprint bumped to v1.4: §7 splits correctness corpus (Soufflé
`tests/`, unchanged) from measurement corpus (OpenRuleBench, new); §9 gets
failure mode #7 (fired, resolved by migration) and #8 (negation-bearing
recursive+bound-query programs may be rare, T3/T4's 34-37% zero rate is first
evidence); §10 closes Q3, closes Q5 for correctness/reopens-then-blocks it for
measurement; §12 adds the P2-scale named benchmark with the `T_none=n²`
disclosure attached permanently.

`tests/corpus/PREREGISTERED.txt` header-annotated as superseded for measurement
purposes (data rows unchanged). `corpus_predicate.py`'s non-determinism fixed
(prohibition lifted narrowly per the ruling §4.3) — verified 3 consecutive runs
byte-identical, `PREREGISTERED.txt` unaffected. `docs/reports/
night01-T6-scaling.md` corrected per §3 (`T_none=n²` is definitional, not a
finding; `T_souffle≈0.62n²`, `T_guard≈1.2n`, `T_souffle/T_guard=Θ(n)`).

Subsumption divergence (§4.1) minimized to 4 nodes/1 rule
(`tests/programs/subsumption_minimal2.dl`), found already reported and already
fixed upstream (souffle-lang/souffle#2322, #2323, PR #2567, merged 2025-12-07 —
8+ months after our installed 2.5). No new issue filed. `docs/reports/
subsumption-repro.md`.

OpenRuleBench pre-registration (§2.2) attempted and **blocked**: neither the
original distribution (dead DNS) nor RUBEN (github.com/kev-ang/RUBEN, cloned and
inspected directly — no rule files, private local dataset path, dead referenced
external host) yields the actual rule-program files. `tests/corpus/
MEASUREMENT_PREREG.txt` not created — no fabrication. Catalog metadata found
before the block (RUBEN's `all_tests.json`) independently suggests the floor of
8 may not be met regardless (`negation` category is 1 program shape,
`same_generation`, at 3 scale points). `docs/reports/
openrulebench-preregistration.md`, `docs/ESCALATIONS.md` (2026-08-21).

**What is now blocked:** OpenRuleBench measurement-corpus pre-registration —
needs either a working data source this session's tools can't reach, or a human
decision to route around it (accept the thin coverage as the answer to failure
mode #8, or select a different measurement corpus with the same rejection-reasoned
rigor DOOP got).

**Single next action:** Human decides how to unblock or route around the
OpenRuleBench access problem. M1 (Lane A) still hasn't started — the ruling's own
§6 says the next session report should lead with M1 or the schedule is fiction;
this session did not touch `src/` and did not change that.
