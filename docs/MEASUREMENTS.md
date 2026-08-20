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
