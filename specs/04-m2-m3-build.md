# M2-M3-BUILD — magic-set transform, guard, fallback

Date: 2026-08-27.

**§0 supersedes `M1-BUILD.md` §1 and `NIGHT-BATCH-03.md` §0.2.1.** Everything else
in both documents stands.

---

## 0. Lane A is retired

`M1-BUILD.md` §1 reserved three components for the human. That boundary existed to
guarantee line-by-line defensibility under hostile questioning. The assessment is a
presentation with no viva, and the human has ruled that this project is not resume
material, so the boundary has no remaining purpose.

**All of `src/` is now yours, including `src/transform/magicset/`,
`src/transform/guard/`, and `src/eval/fallback.go`.** Delete the marker comments and
implement.

The explainability obligation from `M1-BUILD.md` §1 **hardens**, because it is now
the only mechanism keeping the human able to present this. Every package gets a
`DESIGN.md`, and for these three the standard is higher: state the algorithm in
prose, name the paper it comes from, state what you chose where the paper offered
options, and name the cases you know are unhandled.

### Prerequisite

Run `NIGHT-BATCH-03` T1, T2, T3 **first**. T3's `harness/m2_accept.py` is the gate
for every task below. If T1 came back BLOCKED, stop and report — the measurement
path needs redesign before M3 is worth starting.

---

## 1. Definitions used throughout

Fixed once so the code and the report agree.

**Adornment.** For predicate `p` of arity `n`, an adornment `α ∈ {b,f}ⁿ`. Write
`p^α`. `b` at position `i` means argument `i` is bound on entry.

**Bound, formally.** Reuse `sema`'s allowedness grounding fixpoint — do not write a
second notion of groundedness. A term is bound at a point in a rule body iff every
variable in it is in `G` computed over: head variables at `b` positions of the head's
adornment, plus all body literals strictly preceding this point. This reuse is
deliberate: that fixpoint is already validated against Soufflé on 15 probe cases, and
a divergent second definition is the most likely source of a silent wrong answer.

**Magic predicate.** For `p^α`, `magic_p^α` has arity = number of `b` positions in
`α`, holding the projection of `p`'s arguments onto those positions.

**Naming.** Use `magic_<pred>_<adornment>` and `sup_<rule-id>_<i>`, plain
identifiers, no `@`. Soufflé's `@`-prefixed names are its internal convention and
you are not imitating it — `dlc`'s output must be a legal program in `dlc`'s own
grammar, and must round-trip through the printer into something Soufflé accepts
(that is T1's guarantee, and the transform must not break it).

---

## 2. M2.1 — Adornment

`src/transform/magicset/adorn.go`

**Input:** accepted `*ast.Program`, plus the query — the rule whose body is a single
atom with at least one constant, projected from the `.output` relation. Identify it
structurally; if a program has no bindable query, the transform is a no-op and the
program passes through unchanged (report this case, it is expected on the positive
fragment).

**Algorithm.** Worklist over `(predicate, adornment)` pairs.

1. Seed the worklist with `(q, α₀)` where `α₀` has `b` exactly at the query's
   constant positions.
2. Pop `(p, α)`. For each rule `r` defining `p`:
   a. Order the body literals by SIPS (§3).
   b. Walk the ordered body, maintaining the bound-variable set `G`, initialised
      from `p`'s arguments at `b` positions of `α`.
   c. For each IDB atom `q(s̄)` encountered (**including negated ones** — see §5,
      this is load-bearing), compute `β` where `β[i] = b` iff `s̄[i]` is bound in the
      current `G`. Record the adorned occurrence. Push `(q, β)` if not already
      processed.
   d. After each literal, extend `G` with the variables it grounds.
3. Terminate when the worklist empties.

**Termination:** finite predicates × finite adornments per predicate (2^arity), so
the fixpoint terminates. Arity is bounded at 14 in this corpus. But guard it anyway:
cap the worklist at 10,000 adorned pairs and escalate if hit — adornment blowup is
blueprint failure mode #3 and a cap is how you find out you have it.

**Gate:** for each of the five `BENCHMARK_FAMILY` shapes plus `p2.dl` and
`culprit_cycle.dl`, report the set of adorned predicates produced and the worklist
iteration count. No correctness gate yet — this is an intermediate artifact.

---

## 3. M2.2 — SIPS

`src/transform/magicset/sips.go`

Sideways information passing decides body literal order, which decides which
adornments arise, which decides how much the transform saves.

**Choose the simplest defensible strategy and say so in `DESIGN.md`:** left-to-right
in source order, with constraints pulled forward as soon as all their variables are
bound.

Do not implement a cost-based SIPS. It is a research topic, it is not what carries
the result, and a wrong one is indistinguishable from a right one without a cost
model you do not have.

**One required deviation from pure source order:** a negated literal `!q(t̄)` must be
placed after every positive literal that grounds one of its variables. Allowedness
already guarantees such literals exist. If you find a program where no ordering
satisfies this, the program should have been rejected by allowedness — escalate, do
not work around it.

**Gate:** unit tests on ordering, including a case where a constraint moves forward
and a case where a negated literal moves back.

---

## 4. M2.3 — Magic rules and supplementary predicates

`src/transform/magicset/rules.go`

This is where the transform can silently pessimise. Implement supplementary
predicates from the start; do not ship a version without them and plan to add them.
Blueprint failure mode #2 is exactly this, and the literature is explicit that magic
sets without supplementaries can derive an order of magnitude *more* facts than
naive evaluation.

For an adorned rule `r`: `p^α(t̄) :- L₁, …, Lₙ` with body in SIPS order:

**Supplementary chain.**

```
sup_r_0(bound(t̄, α))                    :- magic_p^α(bound(t̄, α)).
sup_r_i(V_i)                            :- sup_r_{i-1}(V_{i-1}), L_i.     for i = 1..n
p^α(t̄)                                  :- sup_r_n(V_n).
```

**`V_i` is the projection that makes this worth doing.** Define:

> `V_i` = (variables bound after processing `L₁..L_i`) ∩ (variables occurring in
> `L_{i+1}..L_n`, or in `t̄`)

That is: keep a variable only if something later still needs it. Without this
projection the supplementary predicates carry every bound variable forward and you
have reimplemented the naive form with extra relations. **This projection is the
single most important line in M2.** Test it directly.

**Magic rules.** For each IDB atom `L_i = q^β(s̄)`:

```
magic_q^β(bound(s̄, β)) :- sup_r_{i-1}(V_{i-1}).
```

**Seed.** From the query's constants:

```
magic_q^α₀(c̄).
```

**Output program** = adorned rules with supplementary chains + magic rules + seed +
all EDB facts + all rules for predicates that were never adorned, unchanged.

**Gate — this is M2's real gate.** Run `harness/m2_accept.py` on all five
`BENCHMARK_FAMILY` shapes plus `p2.dl`:

| check | requirement |
|---|---|
| answer relations | set-equal to the untransformed program, every shape, every scale point |
| tuple total | report `T_none` / `T_souffle` / `T_dlc`, three columns, never aggregated |
| vs hand transform | report `T_dlc` against the committed `guarded/` files |

Answer inequality on any shape at any scale point is a stop-the-line defect. Do not
proceed to M3 with a transform that changes answers.

**Report explicitly:** `T_dlc` vs `T_guard` for `ancestor_nonancestor`, against
**both** `ancestor_nonancestor_guarded.dl` (v1) and `_v2.dl` if NIGHT-BATCH-03 T5
produced it. The prediction pre-registered as Q11 is that v1's hand transform
propagates a seed across a first-argument-invariant recursion and is therefore ~20×
worse than it should be, and that a correct mechanical adornment will not make that
error. Whether `dlc` matches v2 is the test of that prediction. State the numbers
whichever way they fall.

---

## 5. M3.1 — Seed collection over negated occurrences

`src/transform/guard/` — do this **before** the culprit-cycle detector. It is the
part the correctness of everything else rests on.

**The rule:** negated IDB atoms generate magic rules exactly like positive ones.
Skipping them is the obvious implementation and it is wrong — `q` ends up
under-demanded, `!q(t̄)` succeeds because `t̄` was never demanded rather than because
`q(t̄)` is false, and the program returns wrong answers with no negative cycle and no
error.

**Why this is sound, and why clause (b) collapses.** Allowedness requires every
variable in a negated literal to be grounded before it is evaluated. Therefore the
adornment computed for a negated atom is **always all-`b`**. Therefore `magic_q^b…b`
contains ground tuples only, and demand-restricting `q^b…b` to exactly those tuples
decides exactly the membership questions the negation asks. Completeness under
negation holds by construction, given correct seeding.

This is the collapse hypothesis, made concrete: clause (b) is not a separate check to
implement, it is a property of getting the seeding right. Write this argument into
`src/transform/guard/DESIGN.md` — it is the project's central technical claim and
the presentation should be able to state it in three sentences.

**Gate:** assert in code that every adorned occurrence of a negated atom has an
all-`b` adornment. If one does not, that is either an allowedness bug or an
adornment bug — escalate, do not relax the assertion.

**Counterexample search, required.** Take the five constructions in
`tests/programs/p6*.dl` plus every program in `tests/corpus/CULPRIT_CANDIDATES/`
from NIGHT-BATCH-03 T4. For each, run `dlc`'s transform and compare answers against
untransformed. Report `agreed/attempted`. A disagreement here is the counterexample
five bounded attempts failed to find, and it is a genuine result — report it loudly,
do not patch around it.

---

## 6. M3.2 — Culprit-cycle detection (clause a)

`src/transform/guard/stratify.go`

**Cheap precondition first.** A culprit cycle requires the affected predicate to lie
in a positive cycle in the *source* program. Check that first over the source
precedence graph — it is O(V+E) and prunes most programs before any transform work.

**The check.**

1. Build the candidate transformed program (§2–§4).
2. Run `sema.CheckStratification` on it. This is the pre-existing source stratifier,
   applied to the transformed AST — which closes the gap the `Transformer` interface
   documented ("`strata` reflects the pre-transform precedence graph").
3. If it rejects, identify the offending SCCs from its output.

**Differential oracle.** Print the candidate transformed program and feed it to
Soufflé. Soufflé must also refuse to stratify it. This is not self-consistency
testing — Soufflé's stratification checker was written by someone else, and
`night02-T5` already recorded it producing `Unable to stratify {m_q, p_bf, q_bf,
s_bf}` on exactly this shape.

**Gate:** `dlc`-verdict vs Soufflé-verdict on every program in
`CULPRIT_CANDIDATES/` plus `culprit_cycle.dl`. Report `agreed/attempted` and every
disagreement individually. Disagreement is a finding, not a nuisance.

---

## 7. M3.3 — Per-SCC decision and the fallback cone

`src/transform/guard/decide.go`

**The decision.** Per SCC of the transformed program:
- Clause (a) fails on this SCC → `FALLBACK`.
- Otherwise → `TRANSFORM`.

**The cone — this is the part that is easy to get wrong.** `FALLBACK` is downward-
closed over the **entire dependency relation**, positive and negative edges alike.
An untransformed rule iterates over the full extent of the relations it reads and
generates no magic seeds for them; if one of those relations has been demand-
restricted by a transformed consumer elsewhere, the fallback rule reads a relation
missing tuples it needs, and answers are wrong with no negative cycle to catch it.

So: `FALLBACK(C)` ⟹ `FALLBACK(D)` for every `D` in the downward dependency closure
of `C`.

This is observed, not theoretical — `night02-T7` check 3 measured `q` and `s` at
full untransformed extent under Soufflé's own transform when only `p` was
restricted, and `night02-T5` reproduced the identical pattern in the hand guard.

**Alternative strategy, measure but do not default to it.** Soufflé's `@poscopy_1`
is the other option: rather than untransforming the cone, duplicate the relation at
all-free adornment so the fallback consumer gets full extent while the transformed
consumer keeps its restriction. Measured cost on `culprit_cycle` at n=200: 6,899
duplicated tuples. Implement cone-untransform as the default; if time permits,
implement all-free duplication behind a flag and report both numbers. Do not
implement only the second.

**Gate:** for each shape, report `declined_sccs`, `cone_relations`, `cone_size`,
`cone_fraction`, using `harness/cone_metric.py` from NIGHT-BATCH-03 T9 as the
independent cross-check. `dlc`'s own cone computation and the harness's must agree
exactly.

---

## 8. M3.4 — Fallback evaluation

`src/eval/fallback.go`

**Likely much smaller than it looks.** The transformed program contains both
transformed and untransformed relations. Since the cone is downward-closed, and
`sema.CheckStratification` on the transformed program yields an SCC topological
order, and `eval` already processes one SCC at a time in that order (the bug fixed
in M1 §3.9), untransformed relations are already materialised before transformed
consumers read them.

**So: verify this holds before writing machinery.** Construct a mixed program,
evaluate, check answers. If the existing SCC-ordered evaluator handles it, say so in
`DESIGN.md` and write only what is genuinely missing. Do not build a second
evaluation path to solve a problem the first one already solves.

**Gate:** answer set-equality against Souffly on every program with at least one
`FALLBACK` SCC — at minimum `culprit_cycle` and everything from
`CULPRIT_CANDIDATES/`.

---

## 9. M3.5 — The headline run

The project's central number, and the only one a hand transform cannot produce.

**Do:** run the full pipeline over the five `BENCHMARK_FAMILY` shapes at every
`SCALE_POINTS.json` point, plus `CULPRIT_CANDIDATES/`, using the T2 protocol
(`dlc` decides and emits, Soufflé evaluates).

**Produce `docs/reports/m3-headline.md` containing:**

1. Three-column table per shape: `T_none` / `T_souffle` / `T_dlc`. Contribution is
   `T_souffle / T_dlc`. `T_none / T_dlc` is prohibited as a headline — it credits
   the guard with what magic sets already deliver.
2. Guard-firing table: programs where the guard fired, which clause, cone size,
   cone fraction.
3. **The blast-radius integer.** Total cone size across the corpus, and the
   distribution. If the guard declines everything, that is blueprint failure mode
   #1 — the guard is behaviourally identical to Soufflé's blanket skip and there is
   no contribution. Report it plainly if so.
4. `dlc` vs the hand transforms, per shape, including both `ancestor` variants.
5. Applicability, carried forward from §4.3 of the blueprint: the 37% / 34% /
   one-shape zero rates. State it as what it is — a claim about corpus
   availability and about the zero rate on one dimension, not a claim that the
   program class is rare in general. Roughly two thirds of negation-bearing corpus
   programs have something recoverable.

---

## 10. Presentation artifact

**Only after §9 produces a blast-radius integer.** Not before — before that there is
nothing to show that a hand-written `.dl` file did not already show.

**Two days, capped.** A Python script in `harness/` reading committed measurement
JSON plus `dlc --explain` output, emitting one static HTML file. No server, no
framework, no build step, no state. If it exceeds two days it is being over-built;
ship what renders.

Four things must be visible:

1. **The analyzer rejecting programs**, all four grounds, with spans. This is the
   most compiler-looking thing in the project and the gate output already exists.
2. **The three-column metric** and the five-shape taxonomy.
3. **The guard firing and declining**, with the reason named — culprit cycle
   detected, SCC listed, cone reported.
4. **Cone collapse**, with the `culprit_cycle` numbers. This is the most
   interesting finding in the project and it is currently buried in a sub-check.

**`dlc --explain` is worth more than the HTML.** A compiler that prints why it
rejected a program, or why it declined to transform an SCC, is a demo and a
debugging tool and a presentation script at once. Build it during M3, not as demo
work.

---

## 11. One paragraph the report must contain

M1's headline is `T_naive == T_semi_naive` at exactly 1.00×, correctly reframed onto
`DerivationAttempts`. Anyone reading that will ask why M3's headline is not also
1.00×. Write the answer down now:

> Semi-naive evaluation changes the *strategy* for a fixed program — both
> evaluators compute the same minimal Herbrand model, so distinct-tuple counts
> cannot differ. The magic-set transform changes the *program* — a different
> program with a different minimal model, which happens to agree on the query
> relation. That is why one is necessarily 1.00× and the other is not.

---

## 12. Order of work, and what to drop

| # | Item | Drop if short |
|---|---|---|
| 1 | §2 adornment | never |
| 2 | §3 SIPS | never |
| 3 | §4 magic rules + supplementaries | never |
| 4 | §5 negated-occurrence seeding | never |
| 5 | §6 culprit-cycle detection | never |
| 6 | §7 cone decision | never |
| 7 | §8 fallback evaluation | never |
| 8 | §9 headline run | never |
| 9 | §10 presentation artifact | yes |
| 10 | all-free duplication strategy (§7) | yes |
| 11 | C codegen for transformed programs | yes — Q7 closed, this carries no requirement |

**Stopping points are submittable.** Through item 3: a Datalog compiler with a
working magic-set optimizer and a measured reduction. Through item 7: the guard.
Through item 8: the whole thesis.

If the batch dies mid-way, the report says which items completed, with numbers. A
partial M2 with an honest number beats a complete M2 with an unverified one.

---

## 13. Escalation

Continue-on-blocker per `NIGHT-BATCH-03` §0.1, with three exceptions that stop the
batch:

1. **Answers diverge** between transformed and untransformed on any shape. Stop.
   This is the one thing that invalidates everything downstream.
2. **`go build` breaks** and one focused attempt does not restore it. Revert, stop.
3. **A gate would pass only by weakening it.** Report and stop that task.

Never regenerate a golden from `dlc`. Never edit committed `measurements/`. No
wall-clock anywhere.