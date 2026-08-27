# PUNCH-LIST-2 item 1 — decomposing the shrinking margin

Date: 2026-08-27. `harness/punchlist2_p1_decompose.py`
(`measurements/punch-list-2/p1-decompose/summary.json`).

## What did not work

**The hypothesis's specific mechanism claim — "the transformed sibling
branch is demand-restricted and roughly constant in n" — is refuted.**
The transformed portion is not constant: it *shrinks* with `n` (55 → 10
→ 7 tuples for `cc_sibling_emptycone`, and similarly for the other two
constructions). The direction of the overall finding (declined dominates
asymptotically, ratio → 1.0) is confirmed, but for a different, more
specific reason than "roughly constant" — see below.

## Decomposition, all 3 sibling-bearing programs, all 3 scale points

`T_guarded` split into its declined portion (relations in the culprit ∪
cone set, always full-extent) and its transformed portion (the sibling
branch, demand-restricted); the untransformed baseline (`T_none`) split
the same way for comparison:

| program | n | `T_none` | `T_guarded` | ratio | declined (guarded) | transformed (guarded) | declined-predicates (none) | sibling-predicates (none) |
|---|---|---|---|---|---|---|---|---|
| `cc_sibling_emptycone` | 20 | 514 | 303 | 1.696× | 248 | 55 | 248 | 266 |
| `cc_sibling_emptycone` | 50 | 1,101 | 861 | 1.279× | 851 | 10 | 851 | 250 |
| `cc_sibling_emptycone` | 100 | 4,059 | 3,995 | 1.016× | 3,988 | 7 | 3,988 | 71 |
| `cc_both` | 20 | 520 | 309 | 1.683× | 254 | 55 | 254 | 266 |
| `cc_both` | 50 | 1,036 | 796 | 1.302× | 787 | 9 | 787 | 249 |
| `cc_both` | 100 | 3,935 | 3,871 | 1.017× | 3,866 | 5 | 3,866 | 69 |
| `cc_cone_proper_subset` | 20 | 489 | 279 | 1.753× | 222 | 57 | 222 | 267 |
| `cc_cone_proper_subset` | 50 | 1,011 | 772 | 1.310× | 760 | 12 | 760 | 251 |
| `cc_cone_proper_subset` | 100 | 3,926 | 3,863 | 1.016× | 3,854 | 9 | 3,854 | 72 |

Every row's decomposition sums exactly (`declined + transformed ==
T_guarded`, asserted in code, not just checked by eye). **The declined
portion of `T_guarded` is bit-for-bit identical to the declined
predicates' own total in the untransformed `T_none` run, at every point**
(248=248, 851=851, 3,988=3,988, ...) — exactly what "full-extent
fallback" means: the guard is not approximating the untransformed cost of
those relations, it is *literally recomputing the same thing*.

## The actual mechanism

Two effects, not one, both pushing the ratio toward 1.0:

1. **The declined portion grows with `n`, by design** — `base`/`e`
   (feeding the culprit core `p`/`q`/`s`) scale with the scale point per
   `tests/corpus/CONE_CORPUS/SCALE_POINTS.json` (`target_base`/`target_e`
   both increase with `n`). This part of the hypothesis is confirmed
   exactly.
2. **The sibling's own fixture was never scaled with `n` at all — found
   by reading `harness/night04_b_cone_gate.py`'s call to
   `gen_cone_corpus_facts`, which passes `target_base`/`target_e`
   explicitly but not `sibling_edges`, leaving it at its default (40)
   at every scale point.** With the edge count fixed and the node-id
   range growing (`n` itself is `gen_cone_corpus_facts`'s node-count
   parameter, shared between the culprit core and the sibling), the
   sibling's random graph gets **sparser** as `n` grows — its own
   reachable-from-1 set shrinks (266 → 250 → 71 tuples, the untransformed
   `tc` cost), and the demand-restricted version shrinks even faster in
   relative terms (55/266 = 21% restricted-to-full at n=20, down to
   7/71 = 10% at n=100).

**Restated precisely, not as originally hypothesized:** the ratio tends
to 1.0 not because the transformed branch's cost is *pinned*, but because
the *fixture that feeds it was never scaled with `n` in this
construction* — an artifact of how task B's fixtures were built
(`sibling_edges` omitted from the scaled call), not a general property of
demand restriction. A sibling branch whose own input actually grows with
`n` would behave differently — this is exactly what item 2 tests.

## Verdict

**Item 1: DONE.** The general shape of the finding (declined dominates,
ratio → 1.0) is confirmed. The specific stated mechanism ("transformed
stays roughly constant") is refuted — the transformed portion shrinks,
and the reason is a fixture-scaling omission, named and cited, not
assumed. `docs/reports/FINAL.md` is updated to state this precisely, not
as a bare "margin shrinks" limitation.
