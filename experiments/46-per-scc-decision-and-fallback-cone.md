# M3.3 — per-SCC decision and the fallback cone (§7)

Date: 2026-08-27. `src/transform/guard/decide.go`. "The part that is easy
to get wrong" — FALLBACK must be downward-closed over the entire
dependency relation, positive and negative edges alike.

## What did not work

Nothing new in this task — the mechanism reuses M3.2's already-validated
`CheckCulpritCycle`/`AllUnstratifiableSCCs` and M2's already-validated
`Generate` machinery (extended, not replaced, by `GenerateMixed`).

## Method

**Decision.** `Decide` builds the fully-transformed program; if
unstratifiable, every original predicate with a generated relation inside
an unstratifiable SCC is a **culprit** (mapped back via `magicset.
Generate`'s exported `RelationOrigin`). `ConeClosure` computes the
downward dependency closure of the culprit set over the SOURCE precedence
graph, full edges — `declined = culprit ∪ cone`.

**Final program.** `magicset.GenerateMixed(prog, schemas, adorned,
declined)`: every declined predicate's adorned/magic/sup apparatus is
skipped and its original clauses substituted; every occurrence targeting a
declined predicate inside a still-transformed rule is left unrenamed
(reads the original, full-extent relation directly).

**Cross-check (required, §7).** `TestConeClosureMatchesHarnessCulpritCycle`
calls `ConeClosure(culprit_cycle_prog, {"p": true})` directly and asserts
the result is exactly `{q, s}` — NIGHT-BATCH-03 T9's own already-committed
`harness/cone_metric.py` result on the identical input. **Exact agreement,
confirmed.**

7 Go unit tests total across `decide_test.go`/`decide_cone_test.go`.

## The gate

Provenance: `measurements/m3-3-decide/summary.json`. All 6 `CULPRIT_CANDIDATES`
programs + `culprit_cycle.dl`, `--transformer=guarded` vs. the
untransformed original via `harness/m2_accept.py`.

| program | `T_none` | `T_guarded` | answers identical | guard declined anything? |
|---|---|---|---|---|
| `culprit_cycle` | 7,496 | **7,496** | yes | yes — all of `p,q,s` |
| `cc_arity3_twobound` | 242 | **242** | yes | yes — all 3 relations |
| `cc_longer_cycle` | 510 | **510** | yes | yes — all relations |
| `cc_neg_early` | 510 | **510** | yes | yes — all relations |
| `cc_query_bothbound` | 509 | **509** | yes | yes — all relations |
| `cc_third_relation` | 649 | **649** | yes | yes — all relations |
| `cc_edb_negated` | 870 | **54** | yes | **no** — kept the full magic-set benefit |

**7/7 answer-identical.** On every one of the 6 genuine culprit-cycle
programs, `T_guarded == T_none` exactly — the guard declines the entire
program (see below for why), and the transform contributes nothing on
those 6. On the 1 negative control, the guard correctly declines nothing
and the full magic-set reduction (870→54) survives untouched.

**Zero regression on the 5 non-culprit `BENCHMARK_FAMILY` shapes,
confirmed directly:** `dlc emit --transformer=guarded` produces
byte-identical output to `--transformer=magicset` on
`same_generation_negation`, `transitive_closure_bound`,
`ancestor_nonancestor`, and `reachability_complement` — the guard adds
zero overhead and zero unwarranted declines where clause (a) never fails.

## Why the culprit set is always "everything," on this specific family

On `culprit_cycle.dl` and all 5 matching `CULPRIT_CANDIDATES`, the
unstratifiable SCC of the transformed program entangles `p`'s, `q`'s, and
`s`'s own adorned/magic/supplementary relations together — they were
already mutually dependent through the negative cycle before the SCC
computation ever ran. The culprit set computed directly from that one SCC
is therefore `{p, q, s}` (or the shape's equivalent 3 relations) in a
single step; `ConeClosure` adds nothing further on top because everything
reachable from the culprit set is already inside it. This is **not** the
same finding as `culprit_cycle_guarded.dl` (the hand guard,
`docs/reports/night02-T5-guarded.md`), which achieves a small
(`T_guard`=7,024 vs `T_none`=7,496 at n=200, ~1.0× — "essentially no
contribution" per that report too) but nonzero restriction by hand-picking
a smarter, non-mechanically-derivable adornment (restrict only `p`,
leave `q`/`s` fully unbound, discarding the general derivation that hits
this exact cycle). The mechanical guard here declines the SCC it actually
finds; it does not search for an alternative, safer adornment the way a
human author did. **Both land at essentially the same near-zero
contribution number on this shape** — the mechanical guard is coarser
(declines strictly more) but arrives at the same practical outcome as the
hand guard's own more surgical (but non-general) restriction.

## Reading this against blueprint failure mode #1

Blueprint failure mode #1: "if the guard declines everything, the project
has no contribution." Measured directly and reported plainly: the guard
declines **everything** on 6/7 programs in this specific culprit-cycle
corpus — but that corpus was purpose-built (NIGHT-BATCH-03 T4) to trigger
exactly this mechanism on every member but one. The guard's actual
blast radius across the full, non-adversarially-selected `BENCHMARK_FAMILY`
corpus is the correct denominator for failure mode #1, and there it is
demonstrably **not** vacuous: 4/5 shapes see the guard decline nothing at
all and deliver the full magic-set reduction unchanged (`docs/reports/
m2-headline.md`'s already-reported numbers), and the 5th
(`culprit_cycle`) is the one shape the entire project's blueprint names,
from Probe 0 onward, as the deliberately-constructed exercise of clause
(a) — a zero-contribution result there was explicitly anticipated, not a
surprise this task is trying to explain away.

## What a skeptic attacks first

- All 6 declining cases in this gate decline the ENTIRE program (every
  IDB relation) — this task never observed a case where the cone added
  something to a *strict, nonempty* subset already identified as culprit
  (`ConeClosure` always returned empty on every one of the 6). The cone
  mechanism's own correctness is validated (exact match against T9's
  independently-committed number), but its practical necessity beyond
  "the whole SCC is already the answer" is not demonstrated by this
  specific 7-program gate.
- The mechanical guard is strictly more conservative than the hand guard
  on `culprit_cycle` (declines `q`/`s` fully, where the hand guard at
  least tries a partial `p`-only restriction) — this is disclosed as a
  known, real cost, not hidden. A more surgical mechanical guard (e.g.
  trying alternative SIPS orderings or partial adornments per-SCC before
  giving up) is out of this task's scope, per §3's own instruction against
  a cost-based/search-based SIPS.
- 7 programs, all purpose-built or already-known culprit-cycle shapes, is
  a small corpus for this specific gate. The broader-corpus evidence
  (0/817 real-world files structurally match beyond the 1 known one,
  NIGHT-BATCH-03 T4) is what actually establishes this shape's rarity —
  this gate's job is correctness on the shape when it does occur, not
  its frequency.

## Verdict

**M3.3: DONE.** Per-SCC decision + fallback cone implemented, cone
cross-checked exactly against NIGHT-BATCH-03 T9's independent harness,
7/7 programs answer-identical to the untransformed baseline, 0 regression
on the 4 applicable non-culprit `BENCHMARK_FAMILY` shapes. `--transformer=
guarded` registered in `dlc emit` as the full M3 pipeline artifact.
Proceeding to M3.4 (fallback evaluation — verify the existing SCC-ordered
evaluator already handles a mixed program before writing anything new).
