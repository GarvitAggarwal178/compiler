# T0 — Soufflé version risk

Date: 2026-08-22. Cap 1 hour (used well under cap). **Outcome: Master behaves as
2.5. Differentiator holds.** Pin 2.5.

## Procedure and result

1. **Clone.** `souffle-lang/souffle` at master, full clone. HEAD:
   `a1303be3c0166400dee3d1f36f0d96abe03e6901`, dated 2026-07-13 (`git log -1
   --format='%H %ci'`). Reported by the built binary as `2.5-42-ga1303be3` — 42
   commits ahead of the `2.5` tag.

2. **Commit history on the magic-set transform path since the `2.5` tag** (titles
   and messages only, per the Prior Art Register's "do not read the source" rule):

   ```
   git log --oneline 2.5..master -- src/ast/transform/MagicSet.cpp src/ast/transform/MagicSet.h
   7bb8e64f fix: wrong handling of subsumptive clauses in magic transformers (#2567)
   ```

   Exactly **one** commit touched the magic-set transform files since 2.5 — the
   subsumption fix already found and written up
   (`docs/reports/subsumption-repro.md`). Broadened to the whole `src/ast/
   transform/` directory: 6 commits total since 2.5, none of whose titles
   (subsumption ×2, `ResolveAliases` bug fixes, a merge, a rename, a style
   cleanup) suggest any change to negation-restriction/`@neglabel` behaviour.
   This is title-level evidence, gathered before the empirical check below, not
   instead of it.

3. **Build.** Missing build deps (`bison`, `flex`, `sqlite3` CLI) installed
   (obvious-fix, same bucket as the original 2.5 `.deb` install,
   `docs/DECISIONS.md`). Configured and built (`-DSOUFFLE_TESTS=OFF`,
   `-DSOUFFLE_DOMAIN_64BIT=OFF` for build speed — doesn't affect correctness for
   these fixtures, all values are small integers) in well under the 30-minute
   sub-cap. Binary: `/root/souffle-master/build/src/souffle`, confirmed
   `2.5-42-ga1303be3`.

4. **Re-ran P2 and P4' against master**, compared to the already-committed 2.5
   numbers:

   | | 2.5 (committed) | master (`a1303be3`) | match |
   |---|---|---|---|
   | P2 `@neglabel.reach` | 26,404 | 26,404 | yes |
   | P2 `T_souffle` (excl/incl copy) | 26,465 | 26,465 | yes |
   | P2 `E_recoverable` | 26,404 | 26,404 | yes |
   | P2 `q2.csv` | — | sorted set-equal to 2.5's | yes |
   | P4' `T_guard` | 231 | 231 | yes |
   | P4' `q2.csv` | — | sorted set-equal to 2.5's | yes |

   `@neglabel.reach` still appears under master — the negated relation `reach`
   is still fully materialized (isolated, not demand-restricted) exactly as
   under 2.5. Master's own recorded numbers:
   `measurements/t0-version-risk/{p2,p4prime}-master/`.

5. **Reporting regardless of outcome**, per instruction — this is that report,
   for the outcome that occurred (master behaves as 2.5).

## Ruling per the pre-decided outcomes

**Master behaves as 2.5.** The differentiator holds: as of master
`a1303be3` (2026-07-13, 42 commits past the `2.5` tag), Soufflé still isolates
(`@neglabel.<rel>`) rather than demand-restricts negated relations. No expiry
date is currently visible. **Pin 2.5** (already the installed version,
`docs/DECISIONS.md`) — no version change needed. Due diligence recorded; not
escalated (outcome 2's "escalate, premise has an expiry date" branch does not
apply).

## Caveat

A clean master build 42 commits past the tag and a title-only scan of the
transform directory are strong evidence, not a guarantee — a change that hasn't
landed on master yet, or one whose commit title doesn't advertise its effect on
negation handling, wouldn't be caught by this check. This is due diligence at the
cap given (1 hour), not an exhaustive audit. Re-check before any release-version
bump in this project's own toolchain.
