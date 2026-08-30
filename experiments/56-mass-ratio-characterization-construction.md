# PUNCH-LIST-2 item 2 — a construction where the margin grows

Date: 2026-08-27. Q13 pre-registered (`docs/OPEN_QUESTIONS.md`) before
any measurement ran. `tests/corpus/CONE_CORPUS/cc_growing_sibling.dl`
(new), `harness/fixtures_lib.py`'s `gen_growing_sibling_facts` (new),
`harness/punchlist2_item2_growing_sibling.py` (new).

## What did not work

Nothing broke, but the prediction's **direction and mechanism were right
while its magnitude was a significant underestimate** — recorded exactly
that way, not adjusted after seeing the number. Q13 predicted "2×–5× by
n=100"; measured **12.7× at n=100**. The order-of-magnitude reasoning
(declined portion roughly flat, sibling's demand-restricted cost linear
in `n`, sibling's full cost worse-than-linear) was correct in every
qualitative respect — the specific 2×–5× range was too conservative,
because the full transitive closure's actual growth rate turned out
faster than the rough estimate assumed.

## Construction

Structurally identical to `cc_sibling_emptycone.dl` (same `p`/`q`/`s`
culprit core, same `tc`/`out2` sibling) — confirmed via `bin/conecheck`:
`culprit={p,q,s}`, `cone={}`, unchanged. The only difference is the
fixture: `gen_growing_sibling_facts` pins the culprit core at a FIXED
size at every scale point (independent seed, independent of `n`) while
the sibling's graph uses `gen_core_rest_graph(core_size=n)` — every node
`1..n` is reachable from node 1 by construction, so `tc`'s
reachable-from-1 set grows linearly with `n`, the deliberate opposite of
`gen_cone_corpus_facts`'s fixed-edge-count sibling (item 1's finding).

## Gate — answer-identical, hard requirement

3/3 points answer-identical (`measurements/punch-list-2/item2-growing-sibling/summary.json`).
Cone cross-check confirms `cone={}` at every point (expected, unchanged
structure).

## Measured

| n | `T_none` | `T_souffle` | `T_guarded` | declined portion | transformed portion | `T_none/T_guarded` |
|---|---|---|---|---|---|---|
| 20 | 584 | 216 | 273 | 181 | 92 | **2.14×** |
| 50 | 2,660 | 305 | 417 | 208 | 209 | **6.38×** |
| 100 | 8,577 | 457 | 676 | 269 | 407 | **12.69×** |

**The ratio grows monotonically with `n` — 2.14× → 6.38× → 12.69× —
the opposite of the original constructions' 1.75× → 1.02× decay.** The
declined portion stays nearly flat (181 → 208 → 269, the culprit core's
own fixed size, growing only mildly because `n` still sets the node-id
range the fixed-size culprit graph draws from) — confirmed, again,
bit-for-bit identical to the untransformed baseline's own declined-
predicate total at every point (181=181, 208=208, 269=269), the same
invariant item 1 found. The transformed (demand-restricted) sibling
portion grows roughly linearly (92 → 209 → 407, tracking `n` closely);
the untransformed sibling cost grows faster than linear (403 → 2,452 →
8,308) — a full transitive closure over a random recursive tree has
Θ(n log n)-or-worse many-to-many ancestor/descendant pairs, while the
demand-restricted, single-source (`x`=1) view is exactly one row per
reachable node, Θ(n). This is the concrete algorithmic reason the margin
grows here and shrank in task B's original constructions: whether the
guard's contribution grows or shrinks with scale depends entirely on
whether the transformable fraction of the program has a query shape
whose restricted cost grows more slowly than its unrestricted cost — a
property of the SHAPE being transformed, not of the guard mechanism
itself.

## Verdict

**Item 2: DONE.** Prediction's direction confirmed strongly, magnitude
underestimated — reported as measured, not rescued. The finding upgrades
task B/PUNCH-LIST-P1's "guard's contribution shrinks with scale" from a
general property to a construction-dependent one: `cc_growing_sibling`
proves a matching construction exists where it grows instead, by more
than an order of magnitude over the same scale range.
