# Corpus ruling — post NIGHT-BATCH-01

Date: 2026-08-21. Applied to blueprint v1.4. Supersedes the corpus provisions of
Phase 0.7. This is the document of record for the corpus-viability framing
`docs/phase0.7-corpus-viability.md` was meant to carry — that file never landed in
the repo (an authoring error on the human side, not an agent failure; the agent's
refusal to fabricate the missing framing, `docs/ESCALATIONS.md` 2026-08-20, was the
correct disposition). This file replaces it going forward.

## 1. Reading of the T2/T3 result

**What was measured.** 3 of 31 pre-registered programs clear `T_none ≥ 1,000`.
Floor of 8 not met. `E_recoverable / T_souffle`: min 0.0, median 0.026, max 0.571.
`E_recoverable = 0` on 10 of 27.

**What it means.** The median of 0.026 is computed over programs whose totals are
in the dozens. It measures Soufflé's unit tests being unit tests, not how much
Soufflé forfeits. The corpus cannot answer the question in either direction.

**What survives the scale problem.** The **zero rate** — 37% (T3, n=27) and 34%
(T4, n=86), two independent estimates in agreement. That is a fact about program
*structure*: on roughly a third of negation-bearing Datalog programs, no relation
is left unrestricted by Soufflé at all, so there is nothing for a guard to recover.
This is real, it is a limit on the project's reach, and it belongs in the report.

**What does not survive.** `E_recoverable / T_souffle` as a magnitude. Do not
report it as a ratio distribution until it is recomputed on a corpus with scale.

## 2. Ruling: two corpora, two jobs

The single-corpus design conflated two requirements that have nothing to do with
each other.

### 2.1 Correctness corpus — Soufflé `tests/`, in-grammar subset

**Job:** differential set-equality against the oracle. Scale is irrelevant; a
40-tuple program tests the parser, type checker, allowedness check and evaluator
exactly as well as a 40-million-tuple one.

**Source:** `tests/corpus/IN_GRAMMAR.txt` (T5). 31.4% of 622 files ≈ 195 programs.

**Status:** already pre-registered by a mechanical predicate. No change.
`tests/corpus/PREREGISTERED.txt` is retained unmodified as the historical record of
the measurement-corpus attempt; header-annotated as superseded for measurement
purposes and why. Not deleted, not edited row-wise.

**Exclusion:** programs using subsumption (`<=` / `btree_delete`) are outside the
v1.1 grammar and are already excluded by T5's predicate. `semantic/
subsumption_multiple_rules` therefore never enters the correctness corpus. Its
divergence is handled in §4.

### 2.2 Measurement corpus — OpenRuleBench

**Job:** the three-column table. Requires scale, bound queries, and recursion.

**Why this and not the alternatives.** OpenRuleBench supplies externally-authored
*programs* and externally-authored *scalable data generators* — both halves
external, strictly stronger than Phase 0.7's fallback (scaling inputs ourselves).
It includes same-generation with and without query bindings, the canonical
magic-sets benchmark shape, at published sizes in the thousands. DOOP was the
other candidate and is rejected: it leans heavily on Soufflé-specific features
`dlc` will never parse, so `T_guard` could never be produced on it.

**Pre-registration procedure — in this order, one commit, before any measurement
run:**

1. Obtain the suite and its generators. The original distribution and the RUBEN
   repository (built on OpenRuleBench's scripts, using its generators) are both
   candidate sources; record which was used and its commit/version.
2. **Coverage check first.** Count the programs that (a) contain at least one
   negated IDB literal and (b) have a query with a bound argument. Report this
   count *before* selecting anything. OpenRuleBench's negation coverage is
   unverified — if it turns out to be thin, the measurement corpus problem is not
   solved and that must surface immediately, not after a week of integration.
3. Write a mechanical inclusion predicate, structural only. Commit the predicate
   and the resulting file list as `tests/corpus/MEASUREMENT_PREREG.txt`.
4. Pre-register the scale points: the generator sizes to be used, committed before
   any run.
5. **Do not run.**

**Floor:** if fewer than 8 programs satisfy (a) and (b), report and stop. That is a
human decision and the answer is not obvious — it may mean the honest scope of
this project is "negation-bearing recursive Datalog with bound queries is rarer
than assumed," which is a finding, not a failure.

## 3. T6 reporting correction

`T_none` is `n²` plus small change at every point, because `reach` and `unreach`
partition `node × node`. It is definitional, carries no information about the
graph or the transform, and presenting its growth as an observation invites an
immediate attack.

**Required in `night01-T6-scaling.md` and every downstream use:**

- State that `T_none = n²` by construction for this program shape.
- Present the finding as: `T_souffle ≈ 0.62n²`, `T_guard ≈ 1.2n`, hence
  `T_souffle / T_guard = Θ(n)`; mechanism is that Soufflé materializes all of
  `reach` while the guarded form materializes `reach(1,·)`.
- Label the whole table **mechanism characterization on a self-generated fixture
  with uncontrolled reachable-set size**. Six points, one shape. Not a general law.

## 4. Bounded follow-up tasks

### 4.1 Subsumption divergence → upstream issue · cap 2 hours · Lane B

Different answer sets with and without `--magic-transform=*` across four
relations, surviving sorted set comparison. Outside `dlc`'s grammar, so not this
project's correctness problem — but a reproducible discrepancy in a widely-used
tool.

1. Run the cheapest-next-experiment already identified: isolate the `A`/`graph`
   fragment into its own `.dl` and re-run both configurations.
2. Minimize further — fewest rules and facts that still diverge.
3. Check whether it is already reported. If not, file with: minimal repro, both
   outputs, Soufflé version, exact commands.

**Hard cap 2 hours.** If it does not minimize cleanly in that time, log the partial
minimization and stop. This is a side artifact and does not touch the milestone
plan.

### 4.2 `.type` support — deferred to M4, not adopted

68% of grammar exclusions come from `.type` declarations. Tempting and wrong to
take now:

- It would grow the correctness corpus from ~195 to ~400 programs. 195 is already
  more differential tests than M1 can consume.
- It buys nothing for the measurement corpus, which is moving to OpenRuleBench.
- Filter 4 gain is marginal — `number`/`symbol` typing already rejects programs;
  adding subtypes adds roughly one rejection ground.
- Cost is 2–3 days of M1, and M1's slippage comes out of M3, where the result
  lives.

Logged in `docs/OPEN_QUESTIONS.md` as an M4 candidate. Revisit only if
OpenRuleBench integration fails and the Soufflé suite has to carry measurement
after all.

### 4.3 Non-determinism in `corpus_predicate.py`

The diagnostic `matched_output_relation` field is non-deterministic (T1).
NIGHT-BATCH-01 prohibition #2 blocked touching it during the batch; that
prohibition is lifted for this narrow purpose. Fix the field's ordering only, then
re-run the predicate and verify the output file list is byte-identical to the
committed one. If the list changes, that is an escalation — the pre-registration
would have been affected.

## 5. Blueprint → v1.4

- §7 — two corpora, per §2. Correctness metric (set equality) and measurement
  metric (three columns) are separated and never reported together.
- §9 — failure mode #7 (corpus too small-scale) **fired**, resolved by migration
  rather than scope reduction. #8 added: negation-bearing recursive programs with
  bound queries may be rare in real corpora; T3/T4's 34–37% zero rate is the first
  evidence.
- §10 — Q5 reopened for the measurement corpus only; Q3 closed (T8: no artifact
  found).
- §12 — the T6 fixture family added as a named mechanism benchmark, with the
  `T_none = n²` disclosure attached permanently.

## 6. Lane A — M1

Week 1 is gone. M1 is 3 weeks and has not started; M3 is where the result lives,
and M1's slippage comes out of M3, not out of the buffer. The parser is the only
component in this project where there is no prior competence to draw on, and it is
the only component not begun. Those two facts are related.

Lexer → precedence parser → decl/type check → allowedness → naive fixpoint →
semi-naive. Next session report leads with M1 or the schedule is fiction.
