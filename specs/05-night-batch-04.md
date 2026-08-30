# NIGHT-BATCH-04

Date issued: 2026-08-27. Unattended overnight batch. Eight tasks.

Protocol: `NIGHT-BATCH-03.md` §0 in full — continue-on-blocker, one commit per
task, "what did not work" first in every report, state explicitly when a check
finds nothing. Stop conditions unchanged: answers diverging between transformed
and untransformed, a broken build one attempt cannot restore, or a gate that
would pass only by weakening it.

**This is the last implementation batch.** After it, what remains is the writeup
and the demo. Scope accordingly — do not start new investigations that are not
listed here.

---

## A — M4-SIPS

Execute `docs/M4-SIPS.md` in full: §3 (pre-register Q12) before §2, then §2
(demand relaxation), then §5 (supplementary counting convention), then §4's gate
including the full 32-point headline re-run.

**Highest value in the batch.** Predicted 80–180× on `p2.dl` and structurally the
same on `reachability_complement` and `ancestor_nonancestor`.

If §2 lands, `dlc` should reproduce `ancestor_nonancestor_guarded.dl` (v1)'s
`T_guard` numbers. **Report that comparison explicitly** — it is an independent
check on Q8, which T5 reopened. If `dlc` lands near 25,500 at n=500, the shape's
modest 4×–62× band is inherent to the shape; if it lands well below, v1 is
suboptimal after all. Either answer closes Q8.

§5 alone closes the `transitive_closure_bound` question. If A runs long, ship §5.

---

## B — The cone corpus

**The last hole in the thesis.** `m3-headline.md` reports the cone empty in every
case measured. That is not a property of cones — it is a property of a corpus in
which the culprit SCC is always every IDB relation in the program.

**Construct programs with three distinct regions:**

1. A **declined SCC** — the culprit cycle (`p`, `q`, `s` pattern, already
   understood).
2. **Cone relations** — IDB relations strictly *below* the declined SCC in the
   dependency graph. Currently every `CULPRIT_CANDIDATES` program reads EDB
   relations here, which is why the cone is empty. Replace one with an IDB
   relation defined by its own rules, ideally recursive, and the cone becomes
   nonempty immediately.
3. **Sibling relations** — a second query branch that is *not* below the declined
   SCC and therefore stays transformed. Two `.output` relations, two branches
   from a shared EDB base:

   ```
   out1(y)  :- p(1,y).      // culprit branch — declined, plus its cone
   out2(y)  :- tc(1,y).     // clean branch — transformed
   ```

   Make the clean branch large (a transitive closure over the same edge
   relation), so the transformed portion has real mass.

**Build at least four:** one with a nonempty cone and no sibling, one with a
sibling and an empty cone, one with both, and one where the cone is a proper
subset of the non-declined relations. Seeded generators in `fixtures_lib.py`,
scale points pre-registered before any measurement.

**Gate:**
- `guard.Decide` output per program: declined set, cone set, transformed set.
  Cross-check the cone against `harness/cone_metric.py` — exact agreement
  required, as in M3.3.
- `T_none` / `T_souffle` / `T_dlc(guarded)` per program per scale point.
- **`T_guarded < T_none` strictly on at least one program**, or the guard's
  contribution remains unmeasured. If no construction achieves it, say so
  plainly — that is a finding about the guard, not a failure of the task.
- Answer-identical, every program, every scale point.

**Recompute the blast-radius integer** over the enlarged corpus and report cone
size as a distribution, not just a total.

---

## C — All-free duplication as the alternative cone strategy

`m2 m3.md` §7 names this and defaults away from it. It is now worth measuring,
because it is the only mechanism that would give `culprit_cycle` a nonzero
contribution — Soufflé achieves a small real reduction there via `@poscopy_1`
and `dlc` currently achieves exactly zero (`T_dlc == T_none` at all five points).

Implement behind `--cone-strategy=duplicate` alongside the default
`--cone-strategy=untransform`. Rather than untransforming the cone, emit an
all-free adorned copy of each cone relation so the fallback consumer reads full
extent while transformed consumers keep their restriction.

**Gate:** both strategies, every culprit program including B's new ones,
answer-identical under both. Report `T_dlc` side by side. Soufflé's measured cost
for the same trick on `culprit_cycle` at n=200 is 6,899 duplicated tuples — state
whether `dlc` pays more or less.

Do not change the default.

---

## D — Reconciliations

Four small, independent items. None should take long; report a number for each.

1. **T7 audit.** `night03-T7-grammar-v2.md` reports 90% cross-check — **2 of 21
   admitted files do not parse**, unexplained. Identify both, state whether the
   predicate or the parser is wrong, fix whichever it is.
2. **T7 vs V3.** VERIFY-01 §V3 argued the §4-compliant pool is bounded above by
   19, since any compliant file necessarily passes T5's weaker filter and would
   already be inside the 195. T7 reports 21 over the full 622. Both cannot be
   right. Resolve it.
3. **M2 count.** `m2-headline.md` says "5/6 comparable cases"; the session log
   says "5/5." Reconcile and correct whichever is wrong.
4. **Q8 status.** T5 reopened it. Record its current state in
   `OPEN_QUESTIONS.md` explicitly rather than leaving it implied, and update it
   again after A's ancestor comparison lands.

---

## E — `dlc --explain`

Worth more than the demo it feeds. A compiler that prints *why* it did something
is a debugging tool, a presentation script, and a rubric artifact at once.

Three modes:

- **Rejection.** For a program the front end rejects: which check fired, the
  span, the offending variable or relation, and what the rule requires. All four
  grounds — arity, type, allowedness, unstratifiable negation.
- **Transform.** Adorned predicates discovered, worklist iterations, magic
  relations generated, and for each negated occurrence its adornment plus
  whether §2's relaxation applied and why.
- **Guard.** Per SCC: TRANSFORM or FALLBACK; if FALLBACK, which clause fired,
  the culprit SCC, the cone, and the declined fraction.

Plain text to stdout, one fact per line, machine-readable enough that G can
consume it. No formatting work.

**Gate:** exercised on at least one program per mode, output committed under
`docs/reports/explain-samples/`.

---

## F — C codegen for transformed programs

Cheap and it completes the pipeline story. The transformed program is an ordinary
`*ast.Program` and `codegen/` already works; wire `--transformer=` into the
codegen path so `dlc` can emit C for a magic-transformed or guarded program.

**Gate:** generate → compile → run on three programs, answers identical to the
interpreted path and to Soufflé. Report `passing/attempted`.

Q7 is closed and this carries no requirement — drop it if the batch is running
long.

---

## G — Presentation artifact

Unblocked: the blast-radius integer exists. Scope per `m2 m3.md` §10 — a Python
script in `harness/` reading committed measurement JSON plus `dlc --explain`
output, emitting **one static HTML file**. No server, no framework, no build
step.

Four things visible:

1. The analyzer rejecting programs, all four grounds, with spans.
2. Three-column metric, per shape, both counting conventions.
3. The guard firing and declining, with the reason named and the cone shown.
4. Cone collapse and the `bb`→`bf` relaxation, before and after — the two
   findings with mechanisms behind them.

**First cut only in this batch.** If it exceeds two hours, ship what renders and
stop. It is decoration on top of results, and the results are what matter.

---

## H — Final report draft

Last. Everything before it changes the numbers.

Assemble `docs/reports/FINAL.md` from committed material. Structure:

1. What the compiler is, and the source language.
2. Pipeline, pass by pass, with the rejection grounds.
3. The magic-set transform, with the `V_i` projection explained.
4. The soundness problem: magic sets can destroy stratification; the two-clause
   guard; why clause (b) collapses into correct seeding (the three-sentence
   argument from `guard/DESIGN.md`).
5. Results: three-column tables, both conventions, guard-firing table,
   blast-radius distribution, cone results from B.
6. **Findings with mechanisms** — this is the section that carries the
   presentation. The `bb`→`bf` demand-relaxation rule and what it cost before it
   existed; cone collapse; the supplementary counting effect; 0/817 real-world
   culprit-cycle prevalence.
7. What did not work: Q11 falsified, Q8 open or closed depending on A, the
   corpus predicate's repeated failures, DOOP and OpenRuleBench abandoned.
8. Limitations, stated before anyone asks.

**Every number cites its report.** No number appears in FINAL.md that is not
already committed elsewhere.

Mark clearly at the top which sections the human must read before presenting —
at minimum §4 and §6, which are the technical claims he will be asked about.

---

## Order and drop list

| # | Task | Drop if short |
|---|---|---|
| A | M4-SIPS | never (ship §5 alone if needed) |
| B | Cone corpus | never |
| C | All-free duplication | yes |
| D | Reconciliations | no — small and closes open questions |
| E | `dlc --explain` | no — feeds G and the presentation |
| F | Codegen for transformed programs | yes |
| G | Presentation artifact | yes, first cut only |
| H | Final report draft | yes, but say what is missing |

---

## What this batch does not do

- No new benchmark shapes beyond B's cone constructions.
- No cost-based SIPS. A's rule is structural; the prohibition stands.
- No new corpus hunt.
- No further specification documents. This is the last one.