# Measurements

Append-only. Never edit a past row; if a number is superseded, add a new row and
reference the old ID. Every row's integer traces to `measurements/<id>/cmd.txt` and
`measurements/<id>/stdout.txt`, both committed.

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

Derived (not independently committed, arithmetic only, from the rows above):

- `T_base(P1)` = `path.total + q.total` = 1,525,746 + 50 = **1,525,796**
- `T_magic(P1)` = `@interm_out.path.{ff}.total + @interm_out.path.{bf}.total + q.total + 2×magic-seed` = 1,525,746 + 50 + 50 + 1 + 1 = **1,525,848** (final `path` relation excluded from the sum — it is a `COPY_T` of `@interm_out.path.{ff}`, not a fixpoint-derived tuple)
- `T_base(P3)` = `p.total + s.total + out.total` = 30 + 19 + 2 = **51**
- `T_magic(P3)` = `p.{bf}.total + @neglabel.s.total + out.total + magic-seed` = 2 + 19 + 2 + 1 = **24**
