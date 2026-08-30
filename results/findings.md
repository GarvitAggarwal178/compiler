# Findings

Four findings, each with the claim, the numbers, and the experiment record
behind them. None of the numbers below are softened relative to what was
measured — see `docs/05-limitations.md` for how each of these should be
read by a skeptic, and `results/claims.md` for the strength/caveat rating
of every individual claim.

## 1. The `bb`→`bf` demand-relaxation rule, and what it cost before it existed

`dlc`'s mechanical adornment of a negated occurrence is forced to `bb`
(both positions bound) whenever both grounding atoms precede the negation
in source order — a real, measured cost. The fix is not a cost model
(explicitly out of scope, `docs/02-design.md` §5): it is a **structural**
fact — a bound position whose only binder is an unrestricted full-extent
scan carries no demand information and can be relaxed to free, soundly,
because a magic set with fewer bound positions only ever demands a
superset.

**Before this rule existed, the hand-written guard matched or beat the
mechanical transform at every point measured.** Applying the rule
collapses `reachability_complement` and `ancestor_nonancestor` to a single
adornment each (matching the hand-written guards' own shape) and moves
their measured contribution from sub-1×–16× to **46×–1,343×** and
**17×–888×** respectively (`T_souffle/T_dlc`, contribution ratio). It does
**not** fully collapse `same_generation_negation` — a genuine, disclosed
partial result: that shape's recursive rule structurally requires a
second adornment regardless (`experiments/49-demand-relaxation.md`).

**The arc, stated as the project's actual result:** before the relaxation
rule, `dlc`'s mechanical transform on `reachability_complement`/`p2.dl`
was **194× worse than the hand guard at n=250** (`T_dlc` 55,411 vs.
`T_guard` 285) and **5,317× worse at n=8,000** (51,131,165 vs. 9,615) —
the gap *widened* with scale, not narrowed. The relaxation rule closed
that to **parity** (252 vs. 285, excl-sup convention) at n=250. Demand
restriction now **matches** the hand guard — "matches," not "beats": a
12% difference (252 vs. 285) is inside the noise a different SIPS literal
ordering could produce, not a claimed improvement. Total materialization
is still worse (974 incl-sup vs. 285), for a named reason: the
supplementary chain that makes the transform *general* — it works for any
program in the grammar, not one hand-derived shape — costs checkpoint
relations a hand transform, written for one specific program, does not
need.

`p2.dl`'s own pre-registered prediction (Q12, `docs/OPEN_QUESTIONS.md`)
was a **full miss under both counting conventions, in opposite
directions**: predicted `T_dlc ≈ 300–700`; measured 974 (incl-sup, above
the top) and 252 (excl-sup, below the bottom). Neither value falls inside
the stated range. See `results/superseded.md`.

## 2. The guard's contribution as a mass ratio — one characterization, both directions

**The guard's contribution equals the ratio of transformed mass to
declined mass.** Declined relations (culprit ∪ cone) are full-extent by
definition and grow with the program; transformed relations grow only
with their own demanded set. Whichever grows faster determines whether
the contribution rises or falls with scale.

**Direction one — declined mass grows faster.** On the original
task-B cone constructions, `T_guarded < T_none` holds on 9/12 measured
points, 1.02×–1.75×, but the margin *shrinks* as `n` grows:

| program | n=20 | n=50 | n=100 |
|---|---|---|---|
| `cc_sibling_emptycone` | 1.70× | 1.28× | 1.02× |
| `cc_both` | 1.68× | 1.30× | 1.02× |
| `cc_cone_proper_subset` | 1.75× | 1.31× | 1.02× |

Decomposed: the declined portion is **bit-for-bit identical to the
untransformed baseline's own cost** for those relations at every point,
and grows with `n` by design. The transformed portion *shrinks* (55 → 10
→ 7 tuples) because the sibling fixture's own edge count was never scaled
with `n` — a fixture artifact, not a property of demand restriction
(`experiments/55-mass-ratio-characterization-decomposition.md`).

**Direction two — transformed mass grows faster, confirmed by
construction.** `cc_growing_sibling.dl` pins the culprit core at a fixed
size and lets the sibling's own reachable set grow *linearly* with `n` —
the deliberate opposite of the fixture above. Pre-registered prediction:
ratio grows, 2×–5× by n=100. Measured:

| n | `T_none` | `T_guarded` | declined | transformed | ratio |
|---|---|---|---|---|---|
| 20 | 584 | 273 | 181 | 92 | **2.14×** |
| 50 | 2,660 | 417 | 208 | 209 | **6.38×** |
| 100 | 8,577 | 676 | 269 | 407 | **12.69×** |

Direction and mechanism confirmed; magnitude underestimated (12.69×
against a predicted 2×–5×, reported as measured, not adjusted). Algorithmic
reason: a full transitive closure over a random recursive tree grows
worse-than-linear, while the single-source demand-restricted view is
exactly linear (`experiments/56-mass-ratio-characterization-
construction.md`). **This project measured a construction on each side of
that line, not just one that happened to look good.**

## 3. The counting trade-off — two mechanisms, not one number needing a favorable convention

`incl-sup` (counting every relation, including `dlc`'s own supplementary
checkpoints) is the headline convention everywhere in this document.
`excl-sup` has exactly one job — isolating demand-restriction itself from
implementation-strategy cost — used only on `transitive_closure_bound`,
where it resolves a previously-reported ~0.49× "anomaly" to exactly 1.00×
(101==101 at every one of 5 scale points).

**`dlc` is measurably worse than Soufflé's own transform on the positive
fragment: 0.49×, stable across all 5 scale points**, because its
supplementary chain materializes checkpoint relations Soufflé does not
generate. **It is better by orders of magnitude on stratified negation**,
because Soufflé does not demand-restrict negated relations at all. These
are two different mechanisms on two different fragments of the language —
not one number that needed a convention choice to look good.

## 4. 0/817 real-world culprit-cycle prevalence

A structural census over the full Soufflé test corpus (195 in-grammar +
622 full tree = 817) found the culprit-cycle shape — the guard's entire
reason to exist — in **zero** real-world files beyond the one program
constructed to exhibit it. This is consistent with, and sharper than, the
general finding that negation-bearing programs with something recoverable
are common (roughly two-thirds) while this specific unsafe shape is
vanishingly rare.

**Stated plainly: the guard's correctness is thoroughly demonstrated; its
necessity on any known real-world program is not.**
