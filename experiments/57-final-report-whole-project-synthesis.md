# `dlc` — final report

Date: 2026-08-27, revised same day per `docs/punch list.md` P1–P5.
Assembled from committed material only — every number below cites the
report or measurement ID it came from; none is computed fresh in this
document. **This revision fixes three internal contradictions in the
first draft** (a firing-program miscount, a cone-size distribution that
summed to 15 instead of 16, and a stale "0/12" guard-contribution finding
that PUNCH-LIST P1 overturned) — each correction is marked inline where
it occurs, not silently applied.

**Read before presenting, at minimum: §4 (the soundness problem) and §6
(findings with mechanisms).** These are the two sections a hostile
questioner will attack first — §4 is the project's central technical
claim, §6 is what actually carries the result.

**Dropped from this batch, disclosed rather than silently omitted**: C
(all-free duplication cone strategy) and F (C codegen for transformed
programs) were not attempted — both are explicitly marked droppable in
`NIGHT-BATCH-04.md`'s own order-of-work table, and time went to A/B/D/E
instead, all of which are marked never-drop or no-drop. G (presentation
artifact) was not attempted for the same reason (explicitly "yes,
first cut only" droppable) — the raw material for it (`dlc explain`
output, §E; every measurement JSON cited below) is committed and ready
whenever it is picked up.

---

## 1. What the compiler is, and the source language

`dlc` is a from-scratch Datalog compiler (Go, single binary,
`src/cmd/dlc`) built as a semester compiler-design project
(`docs/dlc-blueprint.md`, the project's source of truth for scope). The
source language is a fixed subset of Soufflé's syntax: `.decl`/`.input`/
`.output` declarations, `number`/`symbol` types, Horn-clause rules with
stratified negation, no functors, no aggregates, no records, no
components (blueprint §4's grammar, closed by design — no feature has
been added to it outside two narrow, explicitly-authorized amendments:
optional zero-arity relations and optional parenthesized
`.input`/`.output Name()` syntax, `M1-BUILD.md` §3.3 and
`docs/reports/night03-T6-parens.md`).

The differentiator this project measures is Soufflé's own documented
behavior around **negation and magic sets**: Soufflé's magic-set
optimizer transforms the *negating* relation in a stratified-negation
program but never demand-restricts the *negated* relation itself
(`docs/reports/probe0.md`, `probe0_5.md`) — `dlc` implements a magic-set
transform that treats negated occurrences by the same demand-driven
mechanism as positive ones, plus a soundness guard that detects when this
would break stratification and falls back safely per affected relation.

## 2. Pipeline, pass by pass

1. **Lexer** (`src/lexer`) — hand-written, no generator, per blueprint
   §4's own instruction ("the parser is a learning objective").
2. **Precedence parser** (`src/parser`) — produces `*ast.Program`; a
   `parser.Print`/`parser.Roundtrip` pair keeps the printer honest against
   re-parsing its own output (`docs/reports/night03-T1-printer-souffle.md`:
   17/26 accepted, 17/17 answer-identical among comparable).
3. **Declaration/arity/type check** (`sema.CheckDeclType`) — rejection
   ground `arity_mismatch`/`type_mismatch`.
4. **Allowedness (range restriction)** (`sema.CheckAllowedness`) —
   rejection ground `allowedness`.
5. **Stratification** (`sema.CheckStratification`, Tarjan SCC over the
   full positive+negative precedence graph) — rejection ground
   `unstratifiable`.
6. **Magic-set transform** (`src/transform/magicset`) — §3 below.
7. **Transform-safety guard** (`src/transform/guard`) — §4 below.
8. **Evaluation** — naive (§M1) and semi-naive (`src/eval`) fixpoint
   evaluators; a mixed (partially transformed, partially fallback)
   program uses the same evaluator unmodified
   (`docs/reports/m3-4-fallback.md`).
9. **C codegen** (`src/codegen`) — naive evaluation to standalone C for
   an *untransformed* program (M1 §4.1, 8/8 compile+run tests pass);
   never wired to the transform path (task F, dropped, see above).

All four rejection grounds are demonstrated live in
`docs/reports/explain-samples/` (`dlc explain`, task E) — one committed
sample program per ground.

## 3. The magic-set transform

`src/transform/magicset`: adornment (`adorn.go`), sideways information
passing / SIPS (`sips.go`), magic rules and supplementary predicates
(`rules.go`). Left-to-right source-order SIPS, no cost model
(`docs/m2 m3.md` §3's explicit prohibition, upheld throughout — see §7
below). Negated IDB atoms are adorned and given magic rules by the exact
same mechanism as positive ones, not skipped (the naive-but-wrong
implementation the literature warns against).

**The `V_i` projection.** For an adorned rule with SIPS-ordered body
`L₁..Lₙ`, the supplementary chain `sup_r_i(V_i) :- sup_r_{i-1}(V_{i-1}),
Lᵢ` must project `V_i` down to only the variables still needed —
`(variables bound after processing L₁..Lᵢ) ∩ (variables occurring in
Lᵢ₊₁..Lₙ, or in the head)`. Skipping this projection reimplements naive
evaluation with extra relations (blueprint failure mode #2). `dlc`'s
first draft got the index shift wrong (an off-by-one that dropped a
literal's own bound-position variables from the checkpoint feeding its
own magic rule) — re-derived from first principles before any generated
code existed around the wrong version (`docs/reports/m2-headline.md`,
full derivation in `src/transform/magicset/DESIGN.md`).

**Demand relaxation on negated occurrences** (M4-SIPS, task A). A bound
position in a negated occurrence's adornment can be relaxed to free if
its variable's only binder in the SIPS prefix is an unrestricted
full-extent scan — sound because a magic set with fewer bound positions
demands a superset of what the negation needs, and completeness under
negation requires covering the queried instantiations, not equalling them
(`docs/reports/m4-sips.md` §1, `guard/DESIGN.md`). This is the single
highest-value change measured this session — see §6.

## 4. The soundness problem, and the guard — READ BEFORE PRESENTING

**Magic sets can destroy stratification.** Restricting a negated
relation's extent can introduce a negative cycle through the
magic-seed/supplementary chain that did not exist in the source program
(`tests/corpus/BENCHMARK_FAMILY/culprit_cycle.dl`'s own header comment
derives the exact cycle: `magic_q ->~ s -> q -> magic_q`). This is
blueprint failure mode #1 and the project's central risk.

**The two-clause guard.** Clause (a): re-run stratification on the
candidate transformed program; if unstratifiable, the implicated
predicates and their full dependency cone fall back to reading their
original, untransformed extent (`src/transform/guard/stratify.go`,
`decide.go`). Clause (b): completeness under negation, given correct
seeding.

**Why clause (b) collapses into correct seeding — the three-sentence
argument** (`src/transform/guard/DESIGN.md`, quoted exactly):

> Allowedness requires every variable in a negated literal to be
> grounded before it is evaluated, so the adornment `magicset` computes
> for a negated atom is always all-bound [before M4-SIPS.md's
> relaxation]. Therefore `magic_q^{b...b}` contains ground tuples only,
> and demand-restricting `q^{b...b}` to exactly those tuples decides
> exactly the membership questions the negation asks. Completeness
> under negation holds by construction, given correct seeding.

M4-SIPS.md's relaxation (§3 above) amends this without contradicting it:
the *pre-relaxation* adornment is still always all-bound (asserted,
`guard.AssertNegationSeeding`, checked directly — `docs/reports/
m4-sips.md`'s gate), and relaxation is proven separately sound as its own
lemma (replacing a bound position with free only ever computes a
superset). Both halves of the invariant are asserted in code, not just
argued in prose — a violation is an escalation, not a silent wrong
answer.

**Blast radius.** Across the 16-program corpus measured this session
(12 pre-existing + 4 constructed for task B), 38 relations decline across
**11 firing programs** (recounted directly from the committed
`bin/conecheck` JSON for this revision — see P2 below), declined
fraction 0.43–0.80 per firing program
(`docs/reports/night04-B-cone-corpus.md`). **The guard is not vacuous**
(failure mode #1's stated risk) — it fires on every program with a
genuine culprit-cycle shape and on none without one, 7/7 agreement with
real Soufflé's own stratifier where checked (`docs/reports/
m3-2-culprit-detection.md`).

## 5. Results

Every number below cites `docs/reports/m4-sips.md`,
`docs/reports/night04-B-cone-corpus.md`, or
`docs/reports/punch-list-p1.md`; the pre-M4 numbers they supersede are
`docs/reports/m3-headline.md` (kept, not overwritten, per
`MEASUREMENTS.md`'s append-only rule). **Per PUNCH-LIST.md P2**: every
ratio in this section and §6 was re-derived directly from the committed
measurement JSON (`measurements/m4-sips/`, `measurements/night04-b-cone/`,
`measurements/punch-list/`) for this revision, not copied forward from
this document's own earlier prose — the three corrections above (firing
count, cone-size sum, the P1 rewrite) are what that re-derivation found.

**Three-column table** (`T_none`/`T_souffle`/`T_dlc`, contribution =
`T_souffle/T_dlc`, `incl-sup` convention — `dlc`'s existing default,
counts every relation including supplementary checkpoints):

| shape | scale | `T_souffle/T_dlc` (before M4) | `T_souffle/T_dlc` (after M4) |
|---|---|---|---|
| `reachability_complement` | n=250..8,000 | 0.75×–0.82× | **46.0×–1,342.7×** |
| `ancestor_nonancestor` | n=500..8,000 | 0.96×–15.67× | **16.6×–887.8×** |
| `same_generation_negation` | depth=4..7 | 34.1×–2,185.6× | **45.4×–2,913.5×** |
| `culprit_cycle` | n=20..500 | 1.0× (guard territory) | 1.0× (unchanged) |
| `transitive_closure_bound` | n=500..8,000 | 0.48×–0.49× | 0.48×–0.49× (**incl-sup**); **1.00× exactly (excl-sup)** |

**Both supplementary-counting conventions** (M4-SIPS §5): `T_dlc`
(incl-sup, the default above) vs. `T_dlc` (excl-sup, excluding `dlc`'s
own `sup_*` checkpoint relations, which Soufflé's transform has no
equivalent of). `transitive_closure_bound`'s previously-reported ~0.49×
"anomaly" is confirmed to be the counting convention, not a defect:
excl-sup brings it to exactly 1.00× (101==101 at every one of 5 scale
points).

**Guard-firing table**: fires only on culprit-cycle-shaped programs
(**11/16** in the enlarged corpus — recounted directly from
`bin/conecheck`'s committed JSON for this revision, correcting an
undercount in an earlier draft of this document that said 10/16 and,
separately, 13 in §4; both are now 11, the same set:
`culprit_cycle`, `cc_arity3_twobound`, `cc_longer_cycle`, `cc_neg_early`,
`cc_query_bothbound`, `cc_third_relation`, `cc_mixed_fallback`,
`cc_cone_only`, `cc_sibling_emptycone`, `cc_both`,
`cc_cone_proper_subset` — note `cc_sibling_emptycone` DOES fire, culprit
`{p,q,s}`, it just has an empty cone), always clause (a).

**Blast-radius distribution** (recomputed over 16 programs, task B,
corrected in this revision — the number below now sums to 16, an earlier
draft's `12×0/2×1/1×2` summed to 15): 38 declined relations total,
cone-size distribution **13×0 / 2×1 / 1×2** — see §4 above and
`docs/reports/night04-B-cone-corpus.md`'s full table (13 = the 5
non-firing programs plus the 8 firing programs whose culprit SCC already
equals its full reachable IDB set).

**Cone results, and the guard's contribution (task B, then PUNCH-LIST P1).**
The fallback cone mechanism is exactly correct (4/4 cross-check against
an independent Python implementation, `harness/cone_metric.py`, unchanged
by P1) and genuinely non-vacuous when constructed for (non-empty and even
a two-hop, proper-subset cone were built and verified). Task B initially
measured **`T_guarded < T_none` on 0/12 points**, tracing the cause to
`magicset.FindQuery` seeding only the first bindable query candidate — a
"sibling" `.output` branch was left Untouched (full extent) rather than
independently demand-restricted. PUNCH-LIST P1 fixed exactly this
(`FindQueries`, seed collection from every candidate, gated on
answer-identity across all 9 comparable original+B cases before trusting
the number) — **`T_guarded < T_none` now holds on 9/12 points**, on
every construction with a sibling branch, at every scale point measured:

| program | n=20 | n=50 | n=100 |
|---|---|---|---|
| `cc_sibling_emptycone` | 1.70× | 1.28× | 1.02× |
| `cc_both` | 1.68× | 1.30× | 1.02× |
| `cc_cone_proper_subset` | 1.75× | 1.31× | 1.02× |

(`cc_cone_only`, no sibling by design, is unchanged at 1.00× — the
control, not a counterexample; full table
`docs/reports/punch-list-p1.md`.)

**The guard's contribution as a mass ratio — one characterization, both
curves (PUNCH-LIST-2 items 1–2, consolidated 2026-08-30).** The decaying
margin above and the growing one below were originally reported as two
separate findings. They are one result with a control:

> The guard's contribution equals the ratio of transformed mass to
> declined mass. Declined relations are full-extent by definition and
> grow with the program; transformed relations grow only with their own
> demanded set. Whichever grows faster determines whether the
> contribution rises or falls with scale. Both directions were
> constructed and measured.

*Direction one — declined mass grows faster (PUNCH-LIST-2 item 1).*
Decomposing `T_guarded` into its declined portion (culprit ∪ cone,
always full-extent) and its transformed portion (the sibling branch) at
every point (`docs/reports/punch-list-2-item1.md`) finds the declined
portion is **bit-for-bit identical to the untransformed baseline's own
cost for those same relations, at every point** — full-extent fallback
is literally the same computation, not an approximation. It grows with
`n` by design (the culprit core's own fixtures scale with `n`). The
transformed portion, however, does **not** stay roughly constant as
first guessed — it *shrinks* (55 → 10 → 7 tuples), because task B's
sibling fixture (`gen_cone_corpus_facts`) never scales `sibling_edges`
with `n`, so the sibling's own graph sparsifies and its reachable-from-1
set shrinks alongside it. Both effects — a growing denominator and a
shrinking numerator's-worth-of-savings — push the ratio toward 1.0; the
mechanism is more specific than "the guard's contribution is modest,"
it is "the guard's contribution here is bounded by an unscaled fixture,"
a property of this construction, not of demand restriction in general.

*Direction two — transformed mass grows faster (PUNCH-LIST-2 item 2).*
`cc_growing_sibling.dl` pins the culprit core at a fixed size and lets
the sibling's own reachable-from-1 set grow *linearly* with `n`
(`gen_core_rest_graph(core_size=n)`) — the deliberate opposite of task
B's fixture. Pre-registered prediction (`docs/OPEN_QUESTIONS.md` Q13):
ratio grows, order-of-magnitude 2×–5× by n=100. Measured
(`docs/reports/punch-list-2-item2.md`):

**Both curves, side by side:**

| n | `cc_sibling_emptycone` / `cc_both` / `cc_cone_proper_subset` (`T_none/T_guarded`, declining) | `cc_growing_sibling` (`T_none/T_guarded`, growing) |
|---|---|---|
| 20 | 1.70× / 1.68× / 1.75× | **2.14×** |
| 50 | 1.28× / 1.30× / 1.31× | **6.38×** |
| 100 | 1.02× / 1.02× / 1.02× | **12.69×** |

The growing-sibling construction's own decomposition, for completeness:

| n | `T_none` | `T_guarded` | declined portion | transformed portion | `T_none/T_guarded` |
|---|---|---|---|---|---|
| 20 | 584 | 273 | 181 | 92 | **2.14×** |
| 50 | 2,660 | 417 | 208 | 209 | **6.38×** |
| 100 | 8,577 | 676 | 269 | 407 | **12.69×** |

**The ratio grows monotonically — 2.14× → 6.38× → 12.69× — the exact
opposite of the sibling-bearing constructions above**, direction and
mechanism confirmed, magnitude underestimated (12.69× against a
predicted 2×–5×, reported as measured, not adjusted). Algorithmic reason:
a full transitive closure over a random recursive tree has
worse-than-linear (many-to-many ancestor/descendant) growth, while the
demand-restricted single-source view is exactly one row per reachable
node — linear. **The guard's contribution is real, and whether it grows
or shrinks with scale is a property of the shape being transformed, not
of the guard mechanism** — this project measured a construction on each
side of that line, not just one.

## 6. Findings with mechanisms — this section carries the presentation

1. **The `bb`→`bf` demand-relaxation rule, and what it cost before it
   existed.** `dlc`'s mechanical adornment of a negated occurrence is
   forced to `bb` (both positions bound) whenever both grounding atoms
   precede the negation in source order — a real, measured cost, up to
   ~5,300× worse than a hand-written guard (`docs/reports/m3-headline.md`).
   The fix is not a cost model (explicitly prohibited, `docs/m2 m3.md` §3)
   — it is a **structural** fact: a bound position whose only binder is an
   unrestricted full-extent scan carries no demand information and can be
   relaxed to free, soundly, because a magic set with fewer bound
   positions only ever demands a superset. Applied, this collapses
   `reachability_complement`/`p2.dl` and `ancestor_nonancestor` to a
   single adornment each (matching the hand-written guards' own shape)
   and moves their measured contribution from **sub-1×-to-16×** to
   **46×-to-1,343×** and **17×-to-888×** respectively (both figures are
   `T_souffle/T_dlc`, the contribution ratio — see below for the
   distinct, before/after ratio). It does **not** fully collapse
   `same_generation_negation` — a genuine, disclosed partial result: that
   shape's own recursive rule structurally requires a second adornment
   regardless (`docs/reports/m4-sips.md`'s "what did not work").

   **`p2.dl`, predicted vs. measured, a full miss under both
   supplementary-counting conventions, in opposite directions
   (PUNCH-LIST.md P2 item 3, corrected by PUNCH-LIST-2 item 3):**
   `docs/OPEN_QUESTIONS.md` Q12 predicted, before implementation, `T_dlc
   ≈ 300–700`, an **80–180× reduction from the PRE-relaxation `T_dlc` of
   55,411** (`measurements/m4-sips/before_gate/`) — this is a before/after
   ratio on `dlc`'s own number, not `T_souffle/T_dlc`, and the prediction
   never named a supplementary-counting convention. Measured: raw `T_dlc`
   = **974 (incl-sup)** / **252 (excl-sup)**
   (`measurements/m4-sips/gate/summary.json`). **Checked literally
   against the stated 300–700 bound, neither value falls inside it** —
   974 is above the top, 252 is below the bottom. This is not "correct
   under one convention" — it is a miss under both, in opposite
   directions (56.9× reduction achieved against an 80–180× predicted
   range: below the floor; 252 against the range's own floor of 300:
   also below). Recorded exactly this way in `docs/OPEN_QUESTIONS.md`,
   not smoothed toward either convention. The separate `T_souffle/T_dlc`
   **contribution** ratio at this same point is `44,811/974` = **46.0×**
   (the number in the three-column table above) — a different quantity
   entirely, included here only to be explicit that "56.9×" and "46.0×"
   are not the same ratio and neither should be read as the other.

   **Demand restriction now matches the hand guard; total materialization
   is still worse, for a named reason — and the arc between the two is
   the project's result (PUNCH-LIST.md P3, reframed by PUNCH-LIST-2 item
   4a).** State both conventions in one sentence: `dlc`'s demand
   restriction now **matches** the hand-written guard (`T_dlc` excl-sup
   **252** vs. `p4prime.dl`'s **285** on the identical n=250 fixture,
   `measurements/m4-sips/p4prime_check/` — "matches", not "beats": a 12%
   difference is inside the noise a different SIPS literal ordering could
   produce, not a claimed improvement), while `dlc`'s total
   materialization is still worse (**974** incl-sup vs. 285) because the
   supplementary chain that makes the transform *general* — it works for
   any program, not just this one hand-derived shape — costs checkpoint
   relations a hand transform, written for one specific program, does not
   need.

   **The arc, stated explicitly, because it is the project's result:**
   before M4-SIPS's relaxation, `dlc`'s mechanical transform on this
   shape was **194× worse than the hand guard at n=250** (`T_dlc`
   55,411 vs. `T_guard` 285, `docs/reports/m3-headline.md`) and
   **5,317× worse at n=8,000** (51,131,165 vs. 9,615) — the gap *widened*
   with scale. One structural rule about which bindings carry demand
   information (§3 above) closed that gap to **parity** (252 vs. 285,
   excl-sup) at n=250, with the residual incl-sup gap fully attributable
   to a named, general mechanism (the supplementary chain), not to an
   unexplained residue. `same_generation_negation`'s and
   `ancestor_nonancestor`'s parallel gaps close by the identical
   mechanism, to a smaller but real degree each (the three-column table
   above; Q8, §7). Q12's own stated comparison target (231) came from a different,
   earlier (Phase 0) fixture, not the n=250 one used everywhere else in
   this session — 285 is the correct same-fixture figure and the one
   used here.
2. **Cone collapse is not what "the cone is always empty" looked like.**
   Every naturally-arising culprit-cycle program measured before this
   session had an empty fallback cone, because its culprit SCC already
   equalled its entire reachable IDB set. Constructing a program where it
   does not (task B) required discovering, by direct construction and
   measurement (not assumption), that a new relation read from a
   cycle-participating rule gets swept into an **enlarged culprit set**,
   while the same relation read from a *non*-cycle-participating rule of
   the same predicate becomes a genuine, separate **cone** member — a
   real mechanism of how magic-rule backward edges interact with an
   already-cyclic caller, verified via the raw unstratifiable-SCC dump,
   not inferred.
3. **The supplementary counting convention, scoped to one job
   (PUNCH-LIST.md P4).** `incl-sup` stays the headline convention
   everywhere in this document. Supplementary predicates materialize real
   tuples and are a real cost of `dlc`'s chosen implementation strategy —
   excluding them to make a ratio larger elsewhere would be choosing a
   convention after seeing the result, which this project does not do.
   `excl-sup` has exactly one job: isolating demand-restriction from
   implementation strategy, which is what explains
   `transitive_closure_bound` — used there, and nowhere else in this
   report. The claim, stronger because both halves have a named
   mechanism: **`dlc` is measurably worse than Soufflé's own transform on
   the positive fragment (0.49×, stable across all 5 scale points),
   because its supplementary chain materializes checkpoint relations
   Soufflé does not generate. It is better by orders of magnitude on
   stratified negation, because Soufflé does not demand-restrict negated
   relations at all.** These are two different mechanisms on two
   different fragments of the language, not one number requiring a
   convention choice to look good.
4. **0/817 real-world culprit-cycle prevalence.** A structural census
   over the full Soufflé test corpus (195 in-grammar + 622 full tree,
   `docs/reports/night03-T4-culprit-corpus.md`) found the culprit-cycle
   shape (the guard's entire reason to exist) in **zero** real-world files
   beyond the one already known — consistent with, and sharper than,
   blueprint failure mode #8's general finding that negation-bearing
   programs with something recoverable are common (roughly two-thirds,
   `docs/reports/m3-headline.md` §5) while this SPECIFIC unsafe shape is
   vanishingly rare. Stated plainly: the guard's *correctness* is
   thoroughly demonstrated; its *necessity* on any known real-world
   program is not.
5. **The guard's own contribution, measured for the first time
   (PUNCH-LIST.md P1), and why its scaling direction depends on the
   shape being transformed (PUNCH-LIST-2 items 1–2).** Task B's
   `T_guarded < T_none` finding held on 0/12 points, traced to
   `magicset.FindQuery` seeding only the first bindable query — a second
   `.output` branch was left Untouched (full extent) rather than
   independently restricted. Fixing this (seed collection from every
   query candidate, not a new algorithm) flips the result: **9/12 points
   now show `T_guarded < T_none`**, 1.02×–1.75×, on every construction
   with a sibling branch, but the margin *shrinks* as `n` grows. Traced
   to a specific, verified cause, not a general property: the declined
   portion is bit-for-bit identical to the untransformed cost of those
   same relations and grows with `n` by design, while the transformed
   sibling's cost *shrinks* because its own fixture (`sibling_edges`) was
   never scaled with `n` — an unscaled-fixture artifact, not a property
   of demand restriction. **Confirmed by building the opposite
   construction**: `cc_growing_sibling.dl` pins the culprit core fixed
   and lets the sibling's own reachable set grow linearly with `n`
   instead — the ratio then **grows** monotonically, 2.14× → 6.38× →
   12.69× across the same `n=20/50/100` range (predicted 2×–5× by
   n=100, direction and mechanism confirmed, magnitude underestimated).
   Whether the guard's contribution grows or shrinks with scale is a
   property of the transformed shape's own query cost (linear
   single-source lookup vs. worse-than-linear full closure), not of the
   guard mechanism itself (`docs/reports/punch-list-p1.md`,
   `docs/reports/punch-list-2-item1.md`,
   `docs/reports/punch-list-2-item2.md`).

## 7. What did not work

- **Q12 recorded as a partial miss, both conventions, opposite
  directions (PUNCH-LIST-2 item 3).** Predicted `T_dlc ≈ 300–700` on
  `p2.dl`; measured 974 (incl-sup, above the top) and 252 (excl-sup,
  below the bottom) — neither falls inside the stated range. Not
  recorded as correct under either convention: 252's closeness to
  `p4prime.dl`'s 285 is a real, separate finding (§6 item 1), not
  evidence the pre-registered range itself was right. Cause: the
  prediction never named which convention its range referred to,
  underspecified rather than wrong in mechanism (`docs/OPEN_QUESTIONS.md`).
- **Q11 falsified.** A pre-registered prediction that
  `ancestor_nonancestor`'s hand guard (v1) propagates a seed across a
  first-argument-invariant recursion (analogous to `p4prime.dl`'s
  `reach`) was tested exactly as written and found wrong: the actual
  committed recursion is not of that shape, and the constructed "fix"
  (v2) is answer-wrong at every scale point (`docs/reports/
  night03-T5-ancestor-seed.md`). Not adjusted after the fact.
- **Q8 — closed this session**, in the "v1 is suboptimal" direction:
  `dlc`'s post-relaxation mechanical transform beats v1's hand guard by
  4.2×–9.3× at n=500, growing to 887.8× contribution at n=8,000
  (`docs/reports/m4-sips.md`, `docs/OPEN_QUESTIONS.md`). **Mechanism
  (PUNCH-LIST.md P5), confirmed against the emitted programs directly,
  not just against the aggregate numbers**: v1's own rule is
  `nonancestor_bf(x,y):-m_ancestor(x),person(x),person(y),!ancestor_bf(x,y)`
  — `x` is restricted to `m_ancestor`'s propagated, ancestor-reachable
  50-member set, but the rule's own cost is dominated by `person(y)`
  ranging freely, giving a `50 × |person|` cost (VERIFY-01 §V4: measured
  `nonancestor_bf=22,749` at n=500, `50×455=22,750` — matches to within
  1, confirming the mechanism exactly). `dlc`'s emitted program
  (`measurements/m4-sips/after/ancestor_nonancestor.transformed.dl`)
  seeds `magic_nonancestor_bf(1).` directly from the query constant — a
  **single** value, not a 50-member set — giving `nonancestor_bf`'s own
  cost as `1 × |person|`, exactly 50× smaller at that stage. `ancestor_bf`
  itself still grows to the same ~50-member reachable set eventually (via
  its own recursive rule's magic-seed propagation, `magic_ancestor_bf(z)
  :- sup_ancestor_bf_r1_1(x,z)`), so the two transforms do comparable work
  computing `ancestor`'s reachable set — the entire measured advantage is
  `nonancestor`'s own top-level scan being gated by the query's single
  constant instead of the propagated 50-member set. The hypothesis is
  confirmed, not merely plausible.
- **The corpus predicate's repeated failures.** The mechanical
  regex-based §4-compliance predicate needed three corrective passes
  before it reproduced the parser's own verdict exactly (19/622,
  100% cross-check) — each failure (missing categories, a directive/period
  splitting bug, and finally two Soufflé error-fixture files with no
  generalizable violation category) diagnosed and fixed in turn, not
  smoothed over (`docs/reports/night03-T7-grammar-v2.md`,
  `docs/reports/night04-D-reconciliations.md`).
- **DOOP and OpenRuleBench, both abandoned.** DOOP: a 3-hour capped
  attempt failed at the input-resolution stage on two independent paths;
  not retried with a higher cap, per standing discipline
  (`docs/reports/doop-attempt.md`). OpenRuleBench: no working mirror
  found through any channel this project has access to; negation
  coverage in what could be reconstructed is thin (1 shape)
  (`docs/reports/openrulebench-preregistration.md`). Neither was ever a
  hard dependency — the pre-registered `BENCHMARK_FAMILY` corpus carried
  measurement instead.
- **`T_guarded < T_none` initially achieved on 0/12 of B's points** —
  traced to a real bug (`FindQuery`'s single-candidate limitation, not a
  fundamental property of the guard as first reported) and fixed by
  PUNCH-LIST.md P1; now 9/12. See §5/§6 item 5. Recorded here as a
  correction to this document's own earlier draft, not silently dropped.
- **`same_generation_negation` does not fully collapse** to one
  adornment under M4's relaxation — a genuine, disclosed partial result
  (§6 item 1).

## 8. Limitations, stated before anyone asks

- **Single-query limitation — fixed this session (PUNCH-LIST.md P1),
  noted for the record.** `magicset.FindQuery` originally seeded only the
  first bindable query candidate found in source order, leaving a second
  independent `.output` branch Untouched (full extent). Replaced by
  `FindQueries` (seed collection from every candidate), gated on
  answer-identity across 9 comparable cases before trusting the result.
  This was the decisive reason the fallback cone's practical value could
  not be demonstrated in task B; after the fix, it can be (§5/§6 item 5).
- **No cost-based SIPS**, by design (`docs/m2 m3.md` §3) — `dlc`'s
  mechanical transform is provably beaten by a hand-written guard on
  every shape and scale point measured (up to several thousand-fold on
  `reachability_complement` before M4, narrower but still present after).
- **No wall-clock timing anywhere in this project** (`CLAUDE.md` §0.3) —
  every number is a tuple count; the hardware (WSL2, hybrid CPU, no PMU)
  cannot support a timing claim, and none is made.
- **Structural/regex census tools, not real parsers**, used for corpus
  characterization (`night03_t7_grammar_v2.py`, `cone_metric.py`) — always
  cross-checked against `dlc`'s real parser or Go implementation where a
  correctness claim depends on it (100% and 4/4 respectively, this
  session), never trusted standalone.
- **The culprit-cycle shape the guard exists for was found in 0/817
  real-world files** (§6 item 4) — the guard's contribution is
  demonstrated on constructed programs, not observed to matter on any
  known real corpus.
- **C codegen was never wired to the transform path** (task F, dropped)
  — the pipeline story (parse → transform → evaluate) is complete only
  for the tree-walking evaluator, not for generated C.
- **No presentation artifact was built this session** (task G, dropped)
  — `dlc explain`'s output and every measurement JSON cited above are
  committed and ready for one.
- **Corpus size and checking volume — two numbers with different jobs,
  both labelled, not one under-claiming figure (PUNCH-LIST.md P5,
  extended by PUNCH-LIST-2 item 4b).**

  **External program coverage: 5.** The naive/semi-naive differential
  gate (M1 §4.3, this project's correctness/performance oracle against
  real Soufflé for the naive/semi-naive comparison specifically) ran
  against **5** external programs — the pre-registered `BENCHMARK_FAMILY`
  shapes (`same_generation_negation`, `transitive_closure_bound`,
  `ancestor_nonancestor`, `reachability_complement`, `culprit_cycle`),
  **16/16 comparable scale points matched, 0 disagreements**
  (`docs/reports/m1-progress.md` §4 item 3). This is the number that
  answers "how many distinct external programs did this specific
  oracle exercise" — small and disclosed as small, not padded by
  restating it at every scale point.

  **Total answer/verdict-identity checks against Soufflé, all corpora
  combined: 93.** A different question — "how much checking actually
  happened, project-wide" — summed from every already-committed
  Soufflé-oracle gate: the 32-point transform headline sweep
  (`measurements/m3-5-headline-m4/summary.json`, all answer-identical) +
  PUNCH-LIST P1's 9 comparable re-gate points
  (`measurements/punch-list/p1-gate/summary.json`) + the 39-file hostile
  corpus's parser-verdict agreement against Soufflé's own recorded
  accept/reject behaviour (`docs/reports/night02-T2-hostile.md`,
  `M1-BUILD.md` gate three, 35/39 agree, 2 expected semantic-gap
  disagreements, 1 documented lexer divergence — all 39 accounted for,
  not silently dropped) + the 13-case rejection corpus's
  ground-classification agreement against Soufflé
  (`docs/reports/night02-T9-diagnostics.md`, 13/13 consistent) =
  **32 + 9 + 39 + 13 = 93**.

  These two numbers measure different things and neither substitutes for
  the other: 5 is external-program *coverage* (how many distinct real
  programs), 93 is *checking volume* (how many oracle comparisons ran,
  across every gate, including many repeated scale points on a small
  number of shapes). Separately, corpus-admissibility numbers — the
  strictly blueprint-§4-compliant corpus is **19** programs
  (`docs/reports/night04-D-reconciliations.md`); **89/195** files in
  `IN_GRAMMAR.txt` parse after the `.input`/`.output Name()` parens
  amendment (`docs/reports/night03-T6-parens.md`) — answer a third,
  unrelated question (how much of real-world Soufflé code this project's
  grammar admits) and must not be read as either coverage or checking
  volume.
