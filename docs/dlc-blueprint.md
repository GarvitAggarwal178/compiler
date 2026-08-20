# `dlc` — Demand-Driven Datalog Compiler with a Guarded Magic-Set Transformation

**Course:** Compiler Design (semester project)
**Team:** 2
**Budget:** 15 weeks from 2026-08-20 → 2026-12-03. Realistically ~10 productive weeks.
**Status:** selected 2026-08-20, pending Probe 0. Not started.
**Blueprint version:** v1.1

---

## 1. One-line identity

A compiler from typed Datalog with stratified negation to an executable relational
plan, whose central optimization — the magic-set transformation — is **guarded by a
soundness check that the transformation itself can violate**, with fallback to
untransformed evaluation on the affected components.

## 2. Thesis

The magic-set transformation is not unconditionally sound in the presence of
negation. It can convert a stratified program into an unstratified one, and it can
produce a stratified program that is still wrong because a negated subgoal is
evaluated against an incomplete relation. Production engines are commonly described as declining to transform anything touching
negation; **empirically (Soufflé 2.5, Probe 0) that description is wrong.**

> Soufflé transforms relations whose bodies contain negation, but never
> demand-restricts a *negated* relation — it isolates it (`@neglabel.<rel>`) and
> computes it in full. This discharges completeness-under-negation by brute force and
> forfeits every reduction reachable through a negated literal. The gap is: **when is
> restricting a negated relation safe, and what does refusing to do it cost?**

That conservatism (on the negated side only) has a cost, that cost is an exact
integer, and a guarded transform recovers part of it. See `docs/reports/probe0.md`
and `docs/reports/probe0_5.md` for the measurements behind this restatement.

**The claim we defend:** *our transform is applied under a stated soundness guard;
here are the N programs where the guard fires, here is what the guard is conservative
on, and here is the derived-tuple reduction on the programs where it does not fire.*

**The claim we do NOT make:** "we extended magic sets to negation." We are not
reimplementing Balbin et al. We are implementing a *detector* and a *fallback*.

---

## 3. Filter compliance (the six filters, restated as an audit)

| # | Filter | Verdict | Note |
|---|---|---|---|
| 1 | Substitution | **Pass** | No downloadable small magic-sets-with-negation implementation found. Two independent engines (Soufflé, travitch/datalog) explicitly decline the transform. One open risk: *Extended Magic for Negation* (~2019) — artifact status unresolved, see §10. |
| 2 | External oracle | **Pass** | Soufflé. Not written by us. Set-equality on output relations. |
| 3 | Deterministic metric | **Pass** | Derived-tuple counts. Exact integers. No clock, no PMU, no cache, no thermal drift. |
| 4 | Front-end mass | **Weakest link** | Mitigated deliberately, see §4. Must demo *rejection*. |
| 5 | Defensibility under AI | **Pass** | Headline is rule rewriting + graph analysis: a few hundred lines of tree/graph manipulation, hand-writable. |
| 6 | Graceful degradation | **Pass** | Optimization pass (semi-naive) present at M1. See §8. |

**Filter 4 is the soft spot and the report cannot hide it.** Mitigation is not wording,
it is code:

- Real arithmetic expression grammar with precedence and unary minus.
- Typed declarations (`number` / `symbol`) enforced against every atom occurrence.
- **The demo must show the analyzer rejecting programs**, on four independent grounds:
  arity mismatch, type mismatch, allowedness (range-restriction) violation,
  unstratifiable negation. If nothing is ever rejected, an instructor sees a parser
  and a runtime, not a compiler.

---

## 4. Source language

```
program    ::= decl* clause*

decl       ::= '.decl' ident '(' param (',' param)* ')'
             | '.input' ident
             | '.output' ident
param      ::= ident ':' type
type       ::= 'number' | 'symbol'

clause     ::= atom '.'                       // fact
             | atom ':-' body '.'             // rule
body       ::= literal (',' literal)*
literal    ::= atom
             | '!' atom                       // negation
             | constraint
atom       ::= ident '(' term (',' term)* ')'
term       ::= arith | '_'                    // wildcard
constraint ::= arith relop arith
relop      ::= '=' | '!=' | '<' | '<=' | '>' | '>='

arith      ::= arith ('+' | '-') mul | mul
mul        ::= mul ('*' | '/' | '%') unary | unary
unary      ::= '-' unary | primary
primary    ::= var | number | string | '(' arith ')'
```

Deliberately a strict subset of Soufflé's syntax so that **every test program runs
unmodified on the oracle.** No functors, no aggregates, no components, no records.
Adding any of them is scope creep with no metric attached.

---

## 5. Pass pipeline

```
source
  → lex
  → parse (Pratt/precedence-climbing for arith)
  → declaration & type check                       [REJECTS]
  → allowedness / range restriction                [REJECTS]   ← hypothesis of §6
  → precedence graph + Tarjan SCC + stratification [REJECTS]
  → ┌ GUARD: transform-safety analysis ────────────────────────┐  ← HEADLINE
    │   candidate adornment → provisional transformed program  │
    │   (a) SCC/negative-edge check on transformed graph       │
    │   (b) completeness check for negated subgoals            │
    │   → per-SCC decision: TRANSFORM | FALLBACK               │
    └──────────────────────────────────────────────────────────┘
  → magic-set transformation (adornment, magic rules,
      supplementary predicates) applied to permitted SCCs only
  → semi-naive Δ-rewrite, per stratum
  → rule → join plan (variable ordering, index selection)
  → evaluation (interpreter)  [+ optional C codegen, M4, droppable]
```

**Headline pass: the guard.** Not the magic transform itself. The transform is
textbook; the guard is the result.

---

## 6. The soundness condition (state this precisely or the project has no result)

Let `P` be the source program, `Q` the query, `M(P,Q)` the naively magic-transformed
program.

**Guard(P, Q) permits transformation of SCC `C` iff:**

**Clause (b) is primary.** Probe 0 / 0.5 located the actual gap in Soufflé's
behavior on clause (b), not (a): Soufflé already isolates negated relations soundly
(no observed stratification breakage), it simply refuses to demand-restrict them,
computing every negated relation in full every time. The headline number lives in
clause (b) — restricting a negated relation exactly when it is safe to do so. Clause
(a) is retained below as a correctness side-condition on the transform, tested by P5
(§12) in M3, not as a source of the headline metric.

**(a) Stratification preservation.** The precedence graph of `M(P,Q)` has no negative
edge within any SCC intersecting `C`.

*Mechanism to detect:* magic rules take the body literals *preceding* the target
atom under the SIPS order. If a preceding literal is `!q(...)`, then `magic_r`
acquires a negative edge to `q`. If `q` transitively depends on `r`, and `r` depends
on `magic_r` (it always does, post-transform), the cycle `magic_r →¬ q → r → magic_r`
closes with a negative edge inside it. Source was stratified; transform is not.
Literature calls these *culprit cycles*.

*Cheap necessary precondition:* the affected predicates must lie in a positive cycle
in the source. Check that first; it prunes most programs in O(V+E).

**Unverified — do not build this into the guard yet.** This precondition is taken
from a paper abstract, not confirmed against a worked example. Probe 0's P3 was
intended to exercise exactly this positive-cycle-through-negation shape and did not:
Soufflé's own inliner removed the relation (`q`) the cycle was supposed to run
through before any transform-safety question was even reachable (`docs/reports/
probe0.md`). Check this precondition directly against P5 (§12), the corrected
culprit-cycle program, before relying on it for pruning.

**(b) Completeness under negation.** For every negated literal `!q(t̄)` in `M(P,Q)`,
the magic set for `q` must be complete on the instantiations reachable at the point
of evaluation.

*Why this is separate from (a):* magic sets restricts derivation to *demanded* facts.
Under negation, "not derived" and "not demanded" are indistinguishable. `!q(t̄)` can
succeed because `magic_q` never demanded `t̄`, not because `q(t̄)` is false. **The
transformed program can be stratified and still wrong.** This is why Balbin et al. do
not construct magic rules for negative literals at all, and instead build a separate
program segment per negative literal evaluated under an extra control mechanism —
that structure exists to force completeness before negation.

*Our position:* we implement a conservative static approximation of (b) and we state
exactly what it over-rejects. A conservative check with a named gap is a result. A
silent one is a bug.

**Allowedness is not hygiene.** Balbin et al.'s soundness argument is stated relative
to a definition of *allowedness* (their contribution is a less restrictive one).
Range restriction as implemented in M1 is a hypothesis of the M3 theorem. Implement
it in M1 knowing that M3 depends on it, and record which definition you used.

**Fallback semantics.** Granularity is the SCC of the transformed program. Mixed
evaluation requires that fallback (untransformed) relations be **fully materialized
before** any transformed relation consumes them. Write this interface down in M1 even
though it is not exercised until M3, or M3 will require re-plumbing the evaluator.

---

## 7. Oracles and metrics

### External oracle (Filter 2)

**Primary: Soufflé.** Same `.dl` source, same `.facts`, set-equality on every output
relation. Not written by us.

**Secondary: Jatalog** (Java, semi-naive, stratified negation, no magic sets) as an
independent cross-check on *stratification decisions* specifically — it will disagree
with us on programs where our stratifier is wrong, independently of Soufflé.

**External corpus (fixes benchmark-selection gaming):** Soufflé's own `tests/`
directory. The denominator is fixed by someone else. Report coverage as
`accepted / total` over a named subdirectory, decided before running anything.
Committing to the subdirectory in advance is the whole point; choosing it after
seeing results reintroduces the defect that killed bounds-check elimination.

### Headline metric (Filter 3)

**Primary metric, as of v1.1:**

> Tuples materialized in negated relations under Soufflé's transform, against tuples
> required under a completeness-guarded restriction. Exact integer pair per program.

This targets clause (b) directly (§6) — it is the number that isolates exactly what
Soufflé's blanket-isolate-and-fully-materialize handling of negated relations costs,
independent of whatever else the surrounding program does. Report as a pair
(Soufflé's materialized count, our guarded-restriction count), not a ratio, when
either side is below ~10³ (§3, `docs/reports/probe0_5.md`) — ratios on noise-scale
totals are not evidence.

**Secondary metric — total derived tuples** = Σ over relations of tuples inserted
into Δ across all fixpoint iterations, for a bound query. Exact integer.
Deterministic. Hardware-free.

Reported as three numbers per program:

| Configuration | Symbol |
|---|---|
| semi-naive, no magic | `T_base` |
| semi-naive + magic (positive fragment) | `T_magic` |
| semi-naive + guarded magic (negation present) | `T_guard` |

**What good looks like numerically:**

- Positive reachability, 2000 nodes, reachable-from-source ≈ 50:
  `T_base ≈ 10^6`, `T_magic ≈ 10^3`. Ratio ≥ 3 orders of magnitude, with
  **bit-identical answer relations**. Ratio without identical answers is worthless.
- Stratified-negation corpus: number of programs where the guard fires (exact
  integer), split by clause (a) vs clause (b); and for programs where it does not
  fire, `T_base / T_guard`.
- **Blast-radius integer:** number of relations Soufflé declines to transform because
  of one negated literal, vs. number our guard declines. The delta is the
  contribution, stated as an integer.
- Programs rejected by each of the four semantic checks (Filter 4 evidence).

**Wall clock does not appear in the report.** Not as a headline, not as confirmation.
The hardware cannot support the claim (no PMU under WSL2, hybrid CPU with no defined
"the cache", thermal drift correlated with sweep order). If a reviewer asks for it,
the answer is that the measurement is not executable on this machine and the metric
was chosen to not need it.

---

## 8. Build order

Milestone 1 already contains an optimization pass (Filter 6). Every milestone is
submittable.

**M1 — weeks 1–3. Front end + first optimization.**
Lexer, precedence parser, decl/type check, allowedness, naive fixpoint evaluator,
**semi-naive Δ-rewrite (the M1 optimization)**, Soufflé diff harness, tuple counter.
Deliverable: a Datalog compiler with semantic analysis and one measured optimization
(`T_naive` vs `T_semi-naive`). Fallback-interface stub written but unused.
*Do not compress below 3 weeks.* The Atlas overlap covers semi-naive and Δ-rules only
— roughly 20% of this milestone. The parser, type checker, allowedness check and diff
harness are new.

**M2 — weeks 4–6. Magic sets, positive fragment.**
Constraint normalization, SIPS, adornment fixpoint, magic rules, **supplementary
predicates**, stratification (Tarjan SCC on signed precedence graph), negation in the
evaluator. Headline `T_magic` lands here.

**M3 — weeks 7–11. The guard. This is the project.**
Transform-safety analysis: positive-cycle precondition, culprit-cycle detection
(clause a), completeness approximation for negated subgoals (clause b), per-SCC
TRANSFORM/FALLBACK decision, mixed evaluation. Corpus run. This is where the slack
goes.

**M4 — weeks 12–15. Off critical path.**
C codegen with index selection, join variable ordering with a cardinality cost model,
writeup. **Droppable.** Codegen is a rubric item; it carries no result. If M3 runs
long, M3 wins.

---

## 9. Failure modes, stated as strongly as possible

1. **The guard is vacuous.** If clause (b)'s static approximation is so conservative
   that it rejects every program with negation, the guard is behaviourally identical
   to Soufflé's blanket skip and the project has no contribution. *This is the most
   likely way this dies.* Mitigation: the blast-radius integer must be measured early
   (M3 week 1), not at the end. If it is zero, pivot the headline to the *detector*
   (culprit-cycle classification over the corpus) and drop the fallback.

2. **Magic sets pessimizes.** Omitting or mis-constructing supplementary predicates
   duplicates join work across magic rules; the literature notes magic-set processing
   can generate an order of magnitude *more* facts than naive evaluation. Headline
   metric then moves the wrong direction on half the corpus and the central claim
   inverts.

3. **Adornment blowup.** The adornment fixpoint generates exponentially many adorned
   predicates on programs with many binding patterns. Three weeks disappear into
   termination instead of into the compiler.

4. **Oracle friction.** Soufflé output formatting/ordering differences turn the diff
   harness into a two-week project. Bound it: harness is M1, timeboxed to 2 days,
   and if it exceeds that, compare sorted TSV via `sort | diff` and move on.

5. **Filter 4 collapse.** Front end ends up thin, semantic analysis never rejects
   anything in the demo, instructor reads it as a runtime with a parser bolted on.

6. **Soufflé's isolation turns out to be conditional.** Probe 0 / 0.5 established that
   Soufflé leaves the negated relation fully materialized on two programs. If
   `@neglabel` isolation is applied selectively rather than universally — restricting
   the negated relation on some shapes and not others — the gap this project fills is
   smaller than two programs suggest, and the "primary metric" in §7 could turn out to
   already be near zero on most of the corpus. Mitigation: test on ≥6 negation
   programs of differing shape before M2 closes, and report the count where the
   negated relation was left unreduced by Soufflé.

---

## 10. Open questions — resolve by the stated date

| # | Question | Deadline | Method |
|---|---|---|---|
| Q1 | Does Soufflé's magic transform still decline negation, empirically? | Probe 0, tonight | Run it |
| Q2 | What is the blast radius — one relation, or the whole SCC? | Probe 0, tonight | Run culprit-cycle program |
| Q3 | Does *Extended Magic for Negation* (~2019) ship a downloadable artifact? If yes, Filter 1 verdict must be re-run. | end of week 2 | Search + check for repo |
| Q4 | Chen (1997) labeling algorithm vs Balbin (1991): which is implementable in 5 weeks? | end of week 4 | Read Chen first |
| Q5 | Which Soufflé `tests/` subdirectory is the pre-registered corpus? | end of week 3, **before** any corpus run | Decide and commit to repo |

---

## 11. Prior art register

**Cite. Do not read the source.**

| Artifact | What it is | Filter 1 verdict |
|---|---|---|
| Soufflé | Reference engine and our oracle. Documentation states relations with negation in their body or in the body of a dependency are not transformed. **This is contradicted by observed behaviour in Soufflé 2.5** (`probe0-p2-on-extract`, `probe0-p3-on-extract`): the negation-*bearing* relation is transformed; only the negat*ed* relation is left fully materialized (`@neglabel.<rel>`). A documentation/behavior discrepancy upstream is noted here as a low-priority side artifact, not pursued as a milestone. | Cite-and-avoid. Do not open `MagicSet.cpp`. |
| `travitch/datalog` (Haskell) | `MagicSets.hs` comments that negated literals can break stratification and therefore refuses the transform; author explicitly unsure whether the restriction should cover only negated literals or everything defining them. | Cite-and-avoid. **This uncertainty is our research question — quote it in the report.** |
| Jatalog | Java, semi-naive, stratified negation, no magic sets. | Not substitutable; secondary oracle. |
| Micinski, Syracuse CIS700 Project 2 | Datalog interpreter as a course project, Soufflé-referenced test files, naive positive fragment scores 90%. | Adjacent coursework. Our project starts where it ends. Cite for honesty. |
| DLV | Disjunctive Datalog with magic-set extensions. | Different fragment. Cite. |
| RecStep / DDlog / FlowLog | Datalog on other substrates. | Out of scope. Cite if asked about alternatives. |

**Read. These are papers, not code.**

- Beeri & Ramakrishnan — magic sets, the base transformation.
- **Chen (1997), *Magic Sets and Stratified Databases*** — culprit cycles cause
  unstratification; labeling algorithm described as simpler and more efficient than
  Balbin et al., on the grounds that Balbin's analysis of the abnormal behaviours is
  incomplete. **Read this first.**
- Balbin, Port, Ramamohanarao & Meenakshi, JLP 11(3–4):295–344, 1991 — the canonical
  algorithm; introduces a less restrictive allowedness definition; evaluates negative
  literals via per-literal program segments under an extra control mechanism rather
  than uniform magic rules.
- Ross, *Modular stratification and magic sets for Datalog programs with negation*,
  JACM.
- Behrend, *Soft stratification for magic set based query evaluation*, PODS 2003.
- *Extended Magic for Negation* (~2019) — **artifact status unresolved, see Q3.**

---

## 12. Probe 0 — run tonight, ~70 minutes

Three programs. Four integers plus one format answer. Do not proceed to M1 before
this completes.

### Setup

Install Soufflé from a release `.deb`. Do not build from source.

### P1' — positive reachability (validates the M2 headline; supersedes P1, v1.0)

P1 (v1.0) declared `.output path`, forcing full materialization of `path` regardless
of the magic transform and producing a ≈1 ratio instead of the predicted ~10³
(`docs/reports/probe0.md`). `path` is an intermediate; only `q` is required output.

```
.decl edge(a:number, b:number)
.input edge
.decl path(a:number, b:number)
path(x,y) :- edge(x,y).
path(x,y) :- path(x,z), edge(z,y).
.decl q(b:number)
.output q
q(y) :- path(1,y).
```

Same fixture as P1 (v1.0), same seed, not regenerated: 2000 nodes, ~4000 edges,
reachable-set from node 1 verified at exactly 50 (`probe0-p1-fixture`).

```
souffle -F. -D. -p prof_off.log  p1prime.dl
souffle -F. -D. -p prof_on.log  --magic-transform=* p1prime.dl
souffleprof prof_off.log
souffleprof prof_on.log
```

Checking: (i) Soufflé runs on WSL2; (ii) the magic transform changes the derived-tuple
count on a bound query; (iii) **`souffleprof` reports per-relation tuple counts** —
this is load-bearing, because the headline metric's credibility depends on validating
our counter against a number the oracle reports independently; (iv) the ratio is
~10^4, not ~1. See `docs/reports/probe0_5.md` for the measured result.

### P2 — benign negation (tests the documented claim)

```
.decl edge(a:number, b:number)
.input edge
.decl node(a:number)
.input node
.decl reach(a:number, b:number)
reach(x,y) :- edge(x,y).
reach(x,y) :- reach(x,z), edge(z,y).
.decl unreach(a:number, b:number)
unreach(x,y) :- node(x), node(y), !reach(x,y).
.decl q2(b:number)
.output q2
q2(y) :- unreach(1,y).
```

Negation over a positive recursion that does not depend back on the negating
predicate — the *safe* case. Expect: Soufflé declines to transform the relations
touching negation, per its documentation. Diff `prof_on` against `prof_off`; identical
profiles mean it declined.

### P4 — decisive experiment: hand-transformed P2 (zero `dlc` code, run before M2)

Establishes the project's headline number using Soufflé as both baseline and
executor, with no `dlc` code at all. Hand-write the magic-transformed form of P2,
restricting the *negated* relation — the thing Soufflé's own transform refuses to do
(§2, §11).

```
.decl edge(a:number, b:number)
.input edge
.decl node(a:number)
.input node

.decl m_reach(a:number)                       // magic seed, binding pattern bf
m_reach(1).

.decl reach_bf(a:number, b:number)
reach_bf(x,y) :- m_reach(x), edge(x,y).
reach_bf(x,y) :- m_reach(x), reach_bf(x,z), edge(z,y).

.decl unreach_bf(a:number, b:number)
unreach_bf(1,y) :- node(1), node(y), !reach_bf(1,y).

.decl q2(b:number)
.output q2
q2(y) :- unreach_bf(1,y).
```

Soundness argument to check against the result, not assume: `reach_bf` is complete
for every instantiation the negated literal requires, because the only binding ever
demanded is `x = 1` and the seed supplies it. This is guard clause (b) discharged by
hand, for one program.

Two checks: (1) `q2.csv` byte-identical to P2's `q2.csv`, both configurations — if
not, the guard-clause-(b) argument above is wrong as stated and everything downstream
needs rethinking. (2) tuple count of `reach_bf` against Soufflé's `reach` (26,404,
unchanged in both P2 configurations) — report as an integer pair, not a ratio alone.
This establishes the reduction is achievable and sound on one program. It does **not**
establish that the guard condition can be computed automatically — M3's job.

### P5 — corrected culprit cycle (M3, not Probe 0/0.5 — supersedes P3, v1.0)

P3 (v1.0) was void for two independent reasons (`docs/reports/probe0.md`,
`docs/reports/probe0_5.md`): Soufflé's inliner removed `q` before any transform ran,
and `s(x) :- q(x,_)` made `s` the exact projection of `q`'s first column, so
`!s(z)` and `q(z,y)` were mutually exclusive and the recursive `p` rule could never
fire (`p = 30 = |e|` confirmed this — a dead rule, not a declined transform).

```
.decl base(x:number, y:number)     .input base
.decl e(x:number, y:number)        .input e
.decl blocked(x:number)            .input blocked

.decl q(x:number, y:number)
q(x,y) :- base(x,y).
q(x,y) :- q(x,z), base(z,y).          // recursive: not inlinable

.decl s(x:number)
s(x) :- q(x,y), blocked(y).            // depends on q, strict subset

.decl p(x:number, y:number)
p(x,y) :- e(x,y).
p(x,y) :- p(x,z), !s(z), q(z,y).       // fires when z reaches nothing blocked

.decl out(y:number)   .output out
out(y) :- p(1,y).
```

Source is stratified (`q` ≺ `s` ≺ `p`). The second `p` rule is non-vacuous by
construction. Under MST with query `p(1,·)`:
`magic_q(z) :- magic_p(1), p^bf(1,z), !s(z)` gives `magic_q →¬ s → q → magic_q`, a
negative edge inside an SCC.

**Before running:** verify on the chosen fixture that the second `p` rule actually
fires (`|p| > |e|`) and that `q`'s recursive rule survives Soufflé's inliner (it
should — `q` is self-recursive, and P3's inlined `q` was not). A dead rule or an
inlined-away pivot relation is how P3 failed; check both before trusting any count.

### Report back

1. `T_base` (P1, magic off)
2. `T_magic` (P1, magic on)
3. `T_base` (P3, magic off)
4. `T_magic` (P3, magic on)
5. Does `souffleprof` give per-relation derived-tuple counts? Paste the format.
6. P3 blast radius: how many relations skipped.

Not a plan. Six answers.

*(v1.1 note: the six answers above were produced against P1/P3 as specified in
Blueprint v1.0, before the corrections in this section. They stand as the historical
record of what Probe 0 asked and found — see `docs/reports/probe0.md`. P1' and P4/P5
above are the corrected/superseding programs; results against them are in
`docs/reports/probe0_5.md`.)*

---

## 13. Graveyard (carried forward — do not re-propose without new information)

| Idea | Killed by |
|---|---|
| Array/tensor DSL → scheduled C with Z3 schedule validation | **Filter 1.** `BuildIt-lang/buildit-array` is a maintained hands-on tutorial re-run at PLDI 2024, CGO 2025 and ISCA 2025 — an annually refreshed, student-reproducible walkthrough of the exact deliverable. Same pattern as TeXpresso. Compounded by a vacuous dependence-legality pass (all interchanges legal on pure contractions) and a ~3-week affine-indexing tax to express convolution. |
| Incremental LaTeX compilation | Filter 1 — TeXpresso, texlode. Premise also wrong: line/page breaking is a global optimization, so there is no reusable intermediate typesetting state to depend on. |
| Config / workflow / IoT / network-policy conflict analyzers | Filter 2 — output is a warning list with no ground truth. Also Filter 4: no concrete input language. |
| DL-compiler fuzzing | Filter 1 — NNSmith is pip-installable. Arguably not a compiler. |
| Superoptimizer (Souper-style) | Filter 4 — almost no front end, no semantic analysis. |
| eBPF bytecode optimization | Saturated (K2, Merlin, EPSO, ePass, Kops); duplicates existing personal project. |
| Regex → native code | Filter 1 and Filter 4 — re2c, ragel, Cox, BREeze; regex is one production. |
| Bounds-check elimination | Filter 2 — "% of checks eliminated" measured on self-chosen benchmarks. |
| SQL subset → plan with decorrelation | **Not dead. Ranked #2.** Better oracle (sqllogictest: 7.2M queries, expected results generated by PostgreSQL/MySQL/SQL Server/Oracle) and much stronger front end. Deferred over NULL-semantics risk (the count bug) and unknown grammar coverage of the corpus. **Revive if Probe 0 kills `dlc`.** First action if revived: measure SLT parse coverage of the subset grammar before anything else. |
| Imperative language + Z3 translation validation | **Not dead. Ranked #3.** Salvage of the array-DSL candidate with a real source language, where loop-interchange legality is genuinely non-vacuous. Deferred because the counterexample metric leans on self-generated mutants (partially circular) and loop encoding forces bounded claims. |

---

## 14. Decision log

| Date | Decision | Reason |
|---|---|---|
| 2026-08-20 | Array/tensor DSL killed | `buildit-array` is a maintained annual tutorial, not a one-off artifact — Filter 1, severity under-weighted in first review |
| 2026-08-20 | `dlc` selected | Only candidate where external oracle, exact-integer metric and non-vacuous headline pass are simultaneously strong |
| 2026-08-20 | Headline moved from "magic sets" to "the guard" | The transform is textbook; the soundness condition is the result |
| 2026-08-20 | Guard specified as two clauses, not one | Stratification preservation is necessary but not sufficient; completeness under negation is a separate failure invisible in the precedence graph |
| 2026-08-20 | M1 held at 3 weeks against a proposal to compress to 2 | Atlas overlap covers semi-naive/Δ only (~20% of M1); parser, type checker, allowedness and diff harness are all new |
| 2026-08-20 | C codegen moved off the critical path | Headline metric is a tuple count an interpreter produces; codegen is a rubric item, not a result carrier |

---

## 15. Process rules for this project

1. Prior-art check is the mandatory first step for any new direction. No exceptions.
2. Any scope addition is stated together with its cost against the 15-week budget in
   the same sentence. If the cost is not stated, the addition is rejected.
3. When a review returns errors plus concrete experiments, run the experiments. Asking
   for a different project instead is the failure mode this document exists to prevent.
4. Planning output is capped. If a week produces more prose than code, that is the
   signal, not a milestone.
5. Empirical check beats another search whenever a check is available in under an hour.