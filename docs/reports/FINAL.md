# `dlc` — final report

Date: 2026-08-27. Assembled from committed material only — every number
below cites the report or measurement ID it came from; none is computed
fresh in this document.

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
`docs/reports/night03-T6-grammar-amendment.md`).

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
13 firing programs, declined fraction 0.43–0.80 per firing program
(`docs/reports/night04-B-cone-corpus.md`). **The guard is not vacuous**
(failure mode #1's stated risk) — it fires on every program with a
genuine culprit-cycle shape and on none without one, 7/7 agreement with
real Soufflé's own stratifier where checked (`docs/reports/
m3-2-culprit-detection.md`).

## 5. Results

Every number below cites `docs/reports/m4-sips.md` or
`docs/reports/night04-B-cone-corpus.md`, both fully re-derived this
session; the pre-M4 numbers they supersede are `docs/reports/
m3-headline.md` (kept, not overwritten, per `MEASUREMENTS.md`'s
append-only rule).

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
(10/16 in the enlarged corpus), always clause (a). Cone: empty on all 12
pre-existing programs, non-empty on 3/4 constructed ones (sizes 1, 1, 2 —
task B).

**Blast-radius distribution** (recomputed over 16 programs, task B): 38
declined relations total, cone-size distribution 12×0 / 2×1 / 1×2 — see
§4 above and `docs/reports/night04-B-cone-corpus.md`'s full table.

**Cone results** (task B): the fallback cone mechanism is exactly
correct (4/4 cross-check against an independent Python implementation,
`harness/cone_metric.py`) and genuinely non-vacuous when constructed for
(non-empty and even a two-hop, proper-subset cone were built and
verified) — but **`T_guarded < T_none` held on 0/12 measured points**,
for two identified, structural reasons: declined relations are always
full-extent by definition, and "sibling" relations (a second `.output`
branch) are never actually demand-restricted at all under the current
single-query limitation of `magicset.FindQuery` (confirmed by reading the
emitted program directly, independently reconfirmed by `dlc explain`,
task E). Reported as a finding about the guard, per instruction, not
softened.

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
   **46×-to-1,343×** and **17×-to-888×** respectively. It does **not**
   fully collapse `same_generation_negation` — a genuine, disclosed
   partial result: that shape's own recursive rule structurally requires
   a second adornment regardless (`docs/reports/m4-sips.md`'s "what did
   not work").
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
3. **The supplementary counting effect.** `dlc`'s magic-set transform
   materializes a `sup_*` checkpoint relation per body-literal boundary;
   Soufflé's own transform has no equivalent generated relation. Counting
   them (the project's existing default convention, "incl-sup") makes
   every single reported contribution ratio *understated* relative to
   Soufflé's own accounting — confirmed exactly on
   `transitive_closure_bound` (0.49× incl-sup → 1.00× excl-sup, exact) and
   directionally on `same_generation_negation` (2,185.6× incl-sup →
   ~8,741× excl-sup at depth=7).
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

## 7. What did not work

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
  (`docs/reports/m4-sips.md`, `docs/OPEN_QUESTIONS.md`).
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
- **`T_guarded < T_none` never achieved** on any of B's four constructed
  cone-bearing programs — see §5/§6, reported as a finding, not chased
  with further variants per instruction.
- **`same_generation_negation` does not fully collapse** to one
  adornment under M4's relaxation — a genuine, disclosed partial result
  (§6 item 1).

## 8. Limitations, stated before anyone asks

- **Single-query limitation.** `magicset.FindQuery` seeds only the first
  bindable query candidate found in source order; a second independent
  `.output` branch in the same program is never demand-restricted at all
  (`docs/OPEN_QUESTIONS.md`, found via task B). This is the decisive
  reason the fallback cone's practical value could not be demonstrated
  this session.
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
