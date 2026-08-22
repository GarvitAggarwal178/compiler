# Measurements

Append-only. Never edit a past row; if a number is superseded, add a new row and
reference the old ID. Every row's integer traces to `measurements/<id>/cmd.txt` and
`measurements/<id>/stdout.txt`, both committed.

Per the Phase 0.5 ruling (§3), rows touching a `COPY_T` relation report both
copy-relation conventions (`excl-copy`, `incl-copy`) rather than defending one; see
the "Derived" section below the table.

| ID | Date | Command | Program | Config | Integer | Interpretation |
|---|---|---|---|---|---|---|
| probe0-p1-off-extract | 2026-08-20 | `python3 harness/parse_profile.py measurements/probe0-p1-off-run/prof_off.log` | p1.dl | magic off | `path.total=1525746` | Full pairwise reachability closure over the P1 fixture graph. |
| probe0-p1-off-extract | 2026-08-20 | (same) | p1.dl | magic off | `q.total=50` | Matches the verified reachable-from-1 set size. |
| probe0-p1-on-extract | 2026-08-20 | `python3 harness/parse_profile.py measurements/probe0-p1-on-run/prof_on.log` | p1.dl | magic on (`--magic-transform=*`) | `@interm_out.path.{ff}.total=1525746` | The unadorned/unrestricted path computation, still fully materialized because `path` is itself `.output`. |
| probe0-p1-on-extract | 2026-08-20 | (same) | p1.dl | magic on | `@interm_out.path.{bf}.total=50` | The magic-restricted ("bound-first") path computation — this is the piece that actually shrank. |
| probe0-p1-on-extract | 2026-08-20 | (same) | p1.dl | magic on | `path.total=1525746`, `q.total=50` | Final output relations — identical counts to the magic-off run. |
| probe0-p2-off-extract | 2026-08-20 | `python3 harness/parse_profile.py measurements/probe0-p2-off-run/prof_off.log` | p2.dl | magic off | `reach.total=26404`, `unreach.total=13596`, `q2.total=30` | Baseline for the "benign negation" probe. |
| probe0-p2-on-extract | 2026-08-20 | `python3 harness/parse_profile.py measurements/probe0-p2-on-run/prof_on.log` | p2.dl | magic on | `@neglabel.reach.total=26404`, `unreach.{bf}.total=30`, `q2.total=30` | `unreach` (body contains `!reach`) was magic-transformed and shrank ~453x; `reach` was relabeled but not shrunk. |
| probe0-p3-off-extract | 2026-08-20 | `python3 harness/parse_profile.py measurements/probe0-p3-off-run/prof_off.log` | p3.dl | magic off | `p.total=30`, `s.total=19`, `out.total=2` | Baseline for the culprit-cycle probe. `q` does not appear — Soufflé inlined it away. |
| probe0-p3-on-extract | 2026-08-20 | `python3 harness/parse_profile.py measurements/probe0-p3-on-run/prof_on.log` | p3.dl | magic on | `p.{bf}.total=2`, `@neglabel.s.total=19`, `@magic.p.{bf}.total=1`, `out.total=2` | `p` (body contains `!s`) was magic-transformed via a genuine magic seed. |
| probe0-p1-diff | 2026-08-20 | `diff -q measurements/probe0-p1-off-run/{path,q}.csv measurements/probe0-p1-on-run/{path,q}.csv` | p1.dl | off vs on | 0 differences | Answer relations bit-identical; transform is sound here, just not smaller. |
| probe0-p2-diff | 2026-08-20 | `diff -q measurements/probe0-p2-off-run/q2.csv measurements/probe0-p2-on-run/q2.csv` | p2.dl | off vs on | 0 differences | Answer relation bit-identical. |
| probe0-p3-diff | 2026-08-20 | `diff -q measurements/probe0-p3-off-run/out.csv measurements/probe0-p3-on-run/out.csv` | p3.dl | off vs on | 0 differences | Answer relation bit-identical. |
| probe0-p1-fixture | 2026-08-20 | `python3 harness/probe0.py` (fixture generation step) | p1.dl | n/a | `reachable_from_1=50`, seed=20260820001, attempts=1 | Verified before any Soufflé run, per instructions. |
| probe0.5-p1prime-off-extract | 2026-08-20 | `python3 harness/parse_profile.py measurements/probe0.5-p1prime-off-run/prof_off.log` | p1prime.dl | magic off | `path.total=1525746`, `q.total=50` | Same fixture as P1, unchanged; `.output path` removed. |
| probe0.5-p1prime-on-extract | 2026-08-20 | `python3 harness/parse_profile.py measurements/probe0.5-p1prime-on-run/prof_on.log` | p1prime.dl | magic on | `path.{bf}.total=50`, `@magic.path.{bf}.total=1`, `q.total=50` | No `{ff}` variant appears at all — without `.output path`, Soufflé never computes the unrestricted relation. |
| probe0.5-p1prime-diff | 2026-08-20 | `diff -q measurements/probe0.5-p1prime-off-run/q.csv measurements/probe0.5-p1prime-on-run/q.csv` and vs `probe0-p1-off-run/q.csv` | p1prime.dl | off vs on, and vs original P1 | 0 differences (both) | Answer relation bit-identical to itself across configs and to the original P1's answer. |
| probe0.5-p4-run | 2026-08-20 | `souffle -F fixtures/p2 -D measurements/probe0.5-p4-run -p prof.log tests/programs/p4.dl` | p4.dl | hand-transformed (no `--magic-transform` flag; already in transformed form) | returncode=0 | P2 fixture reused unchanged. |
| probe0.5-p4-extract | 2026-08-20 | `python3 harness/parse_profile.py measurements/probe0.5-p4-run/prof.log` | p4.dl | n/a | `reach_bf.total=170`, `m_reach.total=1`, `unreach_bf.total=30`, `q2.total=30` | Hand-restricted `reach_bf` vs Soufflé's unrestricted `reach`=26,404 (`probe0-p2-off-extract`). |
| probe0.5-p4-diff | 2026-08-20 | `diff -q measurements/probe0.5-p4-run/q2.csv measurements/probe0-p2-{off,on}-run/q2.csv` | p4.dl vs p2.dl | hand-transformed vs both P2 configs | 0 differences (both) | P4's answer bit-identical to P2's, in both configurations. |
| probe0.5-p2-bfs-check | 2026-08-20 | `python3 -c "<inline BFS over fixtures/p2/edge.facts from node 1>"` | p2 fixture | n/a | `reachable_from_1_incl_self=171` | Independent cross-check of `reach_bf`=170 (171 minus the source node itself, which `reach_bf` never reflexively includes). |
| probe0.6-p4prime-run / -extract | 2026-08-20 | `souffle -F fixtures/p2 -D measurements/probe0.6-p4prime-run -p prof.log tests/programs/p4prime.dl`; `python3 harness/parse_profile.py measurements/probe0.6-p4prime-run/prof.log` | p4prime.dl | fixed hand-transform (query constant moved out of the rule head into the seed only) | `m_reach.total=1`, `reach_bf.total=170`, `unreach_bf.total=30`, `q2.total=30` | Same numbers as the original (buggy) P4 — the head-hardcoding bug didn't change the answer, only the artifact's shape. |
| probe0.6-p4prime-diff | 2026-08-20 | `diff -q measurements/probe0.6-p4prime-run/q2.csv measurements/probe0.5-p4-run/q2.csv measurements/probe0-p2-off-run/q2.csv` | p4prime.dl vs p4.dl vs p2.dl | — | 0 differences (all pairs) | P4' answer bit-identical to the original P4 and to P2. |
| probe0.6-p6start-{base,hand}-run | 2026-08-20 | `souffle ... tests/programs/p6start_{base,hand}.dl` (fixtures/p2) | p6start | untransformed vs hand-transformed, seeds `{1,2}` | `ans` 2 = 2, `diff -q` clean | Confirms the directive's stated non-counterexample prediction empirically. |
| probe0.6-p6a1-{base,hand-naive,hand}-run | 2026-08-20 | `souffle ... tests/programs/p6a1_*.dl` (fixtures/p2) | p6a1 | ∃-quantified derived binding | `ans` 200 = 200 = 200 | Degenerate: saturates to the full node domain regardless of correct/incorrect seeding — not evidence either way (`docs/reports/probe0_6.md` §3). |
| probe0.6-p6a1b-{base,hand-naive,hand}-run | 2026-08-20 | `souffle ... tests/programs/p6a1b_*.dl` (fixtures/p2) | p6a1b | ∀-quantified derived binding (fixed version of p6a1) | `ans`: base=31, naive=200, correct=31 | Naive under-seeding diverges; fixpoint-complete seeding (`m_reach(x):-seed(x).`) matches baseline exactly. |
| probe0.6-p6a2-{base,hand}-run | 2026-08-20 | `souffle ... tests/programs/p6a2_*.dl` (fixtures/p2) | p6a2 | independent bf/fb adornments on the same relation | `ans` 9 = 9, `diff -q` clean | Non-degenerate match; two independently-seeded adorned versions of the same relation agree with baseline. |
| probe0.6-q5-eval-only | 2026-08-20 | `python3 harness/corpus_predicate.py /root/souffle-src/tests/evaluation` | souffle tests/evaluation | n/a | `total_tests=149`, `included_count=11` | Superseded below — scope was narrower than the directive asked for. |
| probe0.6-q5-corpus | 2026-08-20 | `python3 harness/build_corpus.py /root/souffle-src/tests` | souffle tests/ (whole tree) | n/a | `total_dl_bearing_dirs=612`, `included_count=36` | Pre-registered corpus (`tests/corpus/PREREGISTERED.txt`); same unmodified predicate as the row above, correct scope. |
| probe0-p1-diff, probe0-p1-diff-q | 2026-08-20 | `diff -q measurements/probe0-p1-off-run/{path,q}.csv measurements/probe0-p1-on-run/{path,q}.csv` | p1.dl | off vs on | 0 differences (both) | Backfill (T1, `night01-T1-audit.md`) — same result already claimed in `docs/reports/probe0.md`, now with real provenance. |
| probe0-p2-diff | 2026-08-20 | `diff -q measurements/probe0-p2-off-run/q2.csv measurements/probe0-p2-on-run/q2.csv` | p2.dl | off vs on | 0 differences | Backfill (T1) — see above. |
| probe0-p3-diff | 2026-08-20 | `diff -q measurements/probe0-p3-off-run/out.csv measurements/probe0-p3-on-run/out.csv` | p3.dl | off vs on | 0 differences | Backfill (T1) — see above. |
| probe0-p1-fixture-bfs-check | 2026-08-20 | `python3 harness/night01_bfs_check.py fixtures/p1/edge.facts 1` | p1 fixture | n/a | `reachable_from_1_incl_self=50` | Backfill (T1) — independent (non-Soufflé) confirmation of the `reachable_from_1=50` figure `probe0-p1-fixture` already claimed; both conventions include the source node itself, so the numbers match directly (no offset). |
| probe0.6-p4prime-diff-vs-p4, -vs-p2 | 2026-08-20 | `diff -q measurements/probe0.6-p4prime-run/q2.csv measurements/probe0.5-p4-run/q2.csv` and vs `measurements/probe0-p2-off-run/q2.csv` | p4prime.dl | vs p4.dl, vs p2.dl | 0 differences (both) | Backfill (T1) — the ID `probe0.6-p4prime-diff` was cited in `docs/reports/probe0_6.md` but never actually captured through `run_cmd`; split into two clean single-command IDs. |

| night01-t2-summary | 2026-08-20 | `python3 harness/night01_t2_corpus.py` | 36 pre-registered programs | untransformed | 31 ok, 4 rejected, 1 crash, 0 DNF; `T_none≥1000` on 3/31; floor of 8 not met | Full table `docs/reports/night01-T2-corpus.md`; per-program provenance `measurements/night01-t2/`. |
| night01-t3-summary | 2026-08-20 | `python3 harness/night01_t3_envelope.py` | 31 T2-ok+seedable programs | `--magic-transform=*` | 27 clean, 1 crash, 1 diverged (abort), 2 not reached; `E_recoverable/T_souffle`: min 0, median 0.026, max 0.571; `E_recoverable=0` on 10/27 | Full table + distribution `docs/reports/night01-T3-envelope.md`; abort detail `docs/ESCALATIONS.md`; per-program provenance `measurements/night01-t3/`. |
| night01-t4-summary | 2026-08-20 | `python3 harness/night01_t4_exploratory.py` | whole `tests/` tree (612) fast pass; 107 negated-IDB programs slow pass | untransformed (fast, structural only) / `--magic-transform=*` (slow) | fast: 107/612 (17.5%) negated IDB, 242/612 (39.5%) seedable, 36/612 both; slow: 86 ok, `@neglabel.` on 57/86 (66%), `E_recoverable=0` on 29/86 (34%) | **EXPLORATORY, not a headline result.** Full counts `docs/reports/night01-T4-exploratory.md`; raw `measurements/night01-t4/`. |
| night01-t5-summary | 2026-08-20 | `python3 harness/night01_t5_grammar.py` | whole `tests/` tree, 622 `.dl` files | n/a (structural, no execution) | `in_grammar=195/622` (31.4%); top exclusion factor `.type` declaration, 291/427 (68.1%) of out-of-grammar files | Full histogram `docs/reports/night01-T5-grammar.md`; file list `tests/corpus/IN_GRAMMAR.txt` (NOT the pre-registered corpus); raw `measurements/night01-t5/`. |
| night01-t6-summary | 2026-08-20 | `python3 harness/night01_t6_scaling.py` | P2 fixture, scaled `n∈{250,500,1000,2000,4000,8000}` | untransformed / `--magic-transform=*` / hand-guarded (`p4prime.dl`) | `T_souffle/T_guard`: 157.2× (n=250) → 4,243.9× (n=8000); `T_none` ×4.0/doubling, `T_guard` ×~2.0/doubling; all answers set-equal at every n | Full table `docs/reports/night01-T6-scaling.md`; raw `measurements/night01-t6/`; fixtures `fixtures/p2-scale-<n>/` (each with a recorded seed). **Reporting correction 2026-08-21** (`docs/reports/corpus-ruling-2026-08-21.md` §3): `T_none=n²` is definitional (`reach`/`unreach` partition `node×node`), not a finding; report as `T_souffle≈0.62n²`, `T_guard≈1.2n`, `T_souffle/T_guard=Θ(n)` instead. Row not edited, per append-only. |

**Non-determinism found, not fixed (T1, hard prohibition #2 applies):**
`tests/corpus/detail.json`'s `matched_output_relation` diagnostic field varies
between runs of `harness/build_corpus.py` due to Python's randomized `set` iteration
order in `corpus_predicate.py`'s `check_program()`. `tests/corpus/
PREREGISTERED.txt` and the included-count (36) are **not** affected — the predicate
only needs "at least one" qualifying relation, which is order-independent. Root
cause logged in `docs/OPEN_QUESTIONS.md`; not fixed tonight because
`corpus_predicate.py` is the predicate NIGHT-BATCH-01 §0.2.2 forbids editing.

**Fixed 2026-08-21 (`corpus-ruling-t4.3-verify`):** `corpus_predicate.py` now
iterates `sorted(output_names)`, authorized narrowly by `docs/reports/
corpus-ruling-2026-08-21.md` §4.3. Command: `python3 harness/build_corpus.py
/root/souffle-src/tests`, run 3 consecutive times — `tests/corpus/detail.json`
byte-identical across all 3; `tests/corpus/PREREGISTERED.txt`'s 36 data rows
unchanged from the pre-fix version (only a superseding header comment added,
separately, §2.1 of the ruling). `docs/OPEN_QUESTIONS.md` updated to resolved.

Three-column headline metric (blueprint §7, v1.2). `T_none` = no transform,
`T_souffle` = Soufflé `--magic-transform=*`, `T_guard` = completeness-guarded hand
restriction (P4'). Contribution is `T_souffle / T_guard`, not `T_none / T_guard`:

| Program | `T_none` | `T_souffle` | `T_guard` | `T_none/T_souffle` | `T_souffle/T_guard` |
|---|---|---|---|---|---|
| P2 | 40,030 (26,404+13,596+30) | 26,465 (26,404+30+30+1) | **231** (1+170+30+30, P4') | ≈1.5× | **≈114.6×** |

Derived (not independently committed, arithmetic only, from the rows above). Per
Phase 0.5 §3, `COPY_T` relations are reported under both conventions rather than one
being defended:

| Program | Config | `T` (excl-copy) | `T` (incl-copy) | Note |
|---|---|---|---|---|
| P1 | off (`T_base`) | 1,525,796 | 1,525,796 | `path.total + q.total`; no copy in this run |
| P1 | on (`T_magic`) | 1,525,848 | 3,051,594 | excl: `@interm_out.path.{ff}` + `@interm_out.path.{bf}` + `q` + 2 seeds = 1,525,746+50+50+1+1. incl: same + the `path` copy (1,525,746) counted again |
| P1' | off (`T_base`) | 1,525,796 | 1,525,796 | identical to P1 off — same fixture, `.output path` is the only change |
| P1' | on (`T_magic`) | 101 | 101 | `path.{bf}` + `q` + 1 seed = 50+50+1; no copy exists in this run at all |
| P3 | off (`T_base`) | 51 | 51 | `p.total + s.total + out.total`; no copy in this run |
| P3 | on (`T_magic`) | 24 | 24 | `p.{bf}` + `@neglabel.s` + `out` + 1 seed = 2+19+2+1; no copy in this run |

P1's magic-on run is the only row where the two conventions diverge — it is also the
only run in this table containing a `.output`-forced `COPY_T` relation. That
divergence is itself evidence for the §1 ruling: the defect was specific to P1's
`.output path`, not general to the transform (P1', P3 show no such divergence).

| subsumption-repro-minimal2 | 2026-08-21 | `souffle -F /tmp/emptyfacts -D measurements/subsumption-repro/minimal2-{none,magic} -p prof.log [--magic-transform=*] tests/programs/subsumption_minimal2.dl` | `subsumption_minimal2.dl` | off vs on | `A.csv`: 3 rows vs 4 rows, extra tuple `(3,4,2)` under magic-transform | Minimized repro of the T3 subsumption divergence; already reported/fixed upstream (souffle-lang/souffle#2322, #2323, PR #2567). Full writeup `docs/reports/subsumption-repro.md`. |
| t0-version-risk | 2026-08-22 | master build `souffle -F fixtures/p2 -D ... -p ... [--magic-transform=*] tests/programs/{p2,p4prime}.dl` (commit `a1303be3c0166400dee3d1f36f0d96abe03e6901`) | p2.dl, p4prime.dl | master vs committed 2.5 | `@neglabel.reach=26404`, `T_souffle=26465`, `E_recoverable=26404`, `T_guard=231` — all identical to 2.5; `q2.csv` sorted set-equal in both | Master (42 commits past `2.5` tag) behaves as 2.5; differentiator holds, 2.5 pinned. Full writeup `docs/reports/T0-version-risk.md`. |
| j1-allowedness-probe-{a..g} | 2026-08-22 | `souffle -F measurements/_scratch_j1/facts -D ... tests/programs/allowedness_probe_{a..g}.dl` | 7 probe programs | n/a | accept: a,b,e,g; reject: c,d,f (`Ungrounded variable X`) | Observed Soufflé allowedness behaviour only, no definition proposed. Full table `docs/reports/J1-allowedness-probe.md`. |
