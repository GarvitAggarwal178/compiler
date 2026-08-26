# M2 — magic-set transform: adornment, SIPS, magic rules (§2–§4)

Date: 2026-08-27. `src/transform/magicset/` (Lane A retired, `docs/m2 m3.md`
§0). Gate: `harness/m2_accept.py` (NIGHT-BATCH-03 T3, validated on 3
known-good pairs before this transform existed).

## What did not work

- **The first `Adorn` draft pushed every atom occurrence onto the
  worklist, including EDB atoms** (`parent`, `person`, `blocked`, ...).
  Caught before any rule generation existed, by re-deriving the IDB
  membership test from `sema/stratify.go`'s own definition (a relation is
  IDB iff it is the head of at least one clause) and gating occurrence
  recording on it. Confirmed fixed by `TestAdornEDBAtomsNeverPushed`.
- **The `V_i` supplementary-predicate projection, taken literally at face
  value from its own stated formula, produces an off-by-one that drops a
  literal's own bound-position variables from the checkpoint feeding its
  magic rule.** Re-derived from first principles (shifting the index by
  one recovers the correct, *inclusive* range) before writing any
  generated code around the wrong version — full derivation in
  `src/transform/magicset/DESIGN.md`. This is exactly the failure mode
  §4 warns about ("the single most important line in M2").
- **The first `Generate` draft duplicated the query-projection rule**: the
  original (dead, since its own body predicate no longer has any defining
  rules) alongside the rewritten one, because the `.output` relation is
  itself "untouched" (never adorned) and its original clauses were copied
  through unconditionally. Fixed by excluding the query's own projection
  clause from the untouched-predicate pass-through.
- **`ancestor_nonancestor.dl`'s mechanical adornment for `ancestor` is
  `bb`, not the `bf` the hand-guarded file uses** — not a bug, a measured
  consequence of the negated literal's own variables both being grounded
  by `person(x),person(y)` before the negation is even reachable under
  SIPS. Pinned by `TestAdornNegatedOccurrenceIsAllBound`; the cost
  consequence is reported below, not hidden.
- **`culprit_cycle.dl`'s mechanical transform is unstratifiable**
  (`Unable to stratify {magic_q_bf, magic_s_b, p_bf, q_bf, s_b, ...}`,
  Soufflé's own stratifier on the *printed* candidate program) — this is
  the expected, designed-for outcome (§6's whole reason for existing), not
  an M2 defect. `answers_identical` is reported as not-comparable for this
  one shape, pending §6/§7's guard.

## §2 gate — adorned predicates and worklist iteration count

No correctness gate at this stage, per instruction — reporting the
discovered set only. `TestReportAdornedPredicates`
(`src/transform/magicset/adorn_report_test.go`).

| program | iterations | adorned (predicate, adornment) pairs |
|---|---|---|
| `same_generation_negation.dl` | 2 | `notsg/bf`, `sg/bb` |
| `transitive_closure_bound.dl` | 1 | `tc/bf` |
| `ancestor_nonancestor.dl` | 2 | `nonancestor/bf`, `ancestor/bb` |
| `reachability_complement.dl` | 3 | `unreach/bf`, `reach/bb`, `reach/bf` |
| `culprit_cycle.dl` | 3 | `p/bf`, `s/b`, `q/bf` |
| `p2.dl` | 3 | `unreach/bf`, `reach/bb`, `reach/bf` |

`reachability_complement.dl`/`p2.dl` adorn `reach` **twice**, under two
different bindings: `bb` from the negation in `unreach`/`unreach`, and
`bf` from `reach`'s own recursive rule (only its first argument is carried
into the recursive call — a second, independent demand pattern the
worklist discovers on its own, not something hand-derived). `p2.dl` and
`reachability_complement.dl` produce identical adornment sets, consistent
with their already-known structural identity (`docs/reports/
night02-T4-baseline.md`).

## §3 gate — SIPS unit tests

`src/transform/magicset/sips_test.go`, 3 tests, all pass:
`TestOrderBodyNegatedLiteralMovesBack` (a negated literal scheduled after
the two positive atoms that ground its variables, even though it appears
earlier in source order), `TestOrderBodyConstraintPulledForward` (a
constraint scheduled as soon as its one variable is bound), 
`TestOrderBodyRespectsInitialBound` (a pre-bound head variable lets a
negated literal fire one atom earlier than it otherwise could).

## §4 gate — `harness/m2_accept.py` on all 5 shapes + `p2.dl`

Provenance: `measurements/m2-gate/summary.json`,
`measurements/m2-gate/t-souffle/summary.json`. Method: `dlc emit
--transformer=magicset` → `harness/m2_accept.py` vs the untransformed
original, smallest pre-registered scale point per shape (a full sweep is
§9's job).

| shape | `T_none` | `T_souffle` | `T_dlc` | answers identical | `T_dlc` vs `T_guard` |
|---|---|---|---|---|---|
| `same_generation_negation` | 51,301 | 30,907 | **907** | yes | `T_guard`=452 — `dlc` 2.0× worse than the hand guard, both still ≫ better than `T_none` |
| `transitive_closure_bound` | 86,618 | 101 | **207** | yes | no hand guard exists (no negation to recover) |
| `ancestor_nonancestor` | 250,450 | 101,250 | **105,552** | yes | `T_guard`(v1)=25,500 — `dlc` ~4.1× worse; `T_guard`(v2)=996 but **answers disagree** (v2 is the falsified, unsound construction from NIGHT-BATCH-03 T5 — `dlc`'s correct transform disagreeing with it is a *confirmation* v2 is wrong, not a `dlc` defect) |
| `reachability_complement` | 62,534 | 44,811 | **55,411** | yes | `T_guard`=285 — `dlc` ~194× worse |
| `culprit_cycle` | 286 (`docs/reports/night02-T4-baseline.md`, same fixture) | 272 | **unstratifiable** | not comparable | guard/fallback territory (§6–§8), not measured here |
| `p2.dl` | 62,534 | 44,811 | **55,411** | yes | `T_guard`(`p4prime.dl`)=285 — identical numbers to `reachability_complement`, confirming the two files' already-known structural identity |

**4/5 shapes + `p2.dl` (5/6 total comparable cases): 5/5 answer-identical
to the untransformed baseline.** `culprit_cycle` is not a correctness
failure — it is the exact designed-for signal for the guard.

### Why `ancestor_nonancestor`, `reachability_complement`, and `p2.dl` are all worse than their hand guards

All three share the identical structural pattern:
`unreach_or_nonancestor(x,y):-node_or_person(x),node_or_person(y),
!reach_or_ancestor(x,y).` — **both** `x` and `y` are grounded by the two
positive atoms before the negation is reachable, so the mechanical
adornment is `bb`. The hand guards restrict only one argument (`bf`) —
a deliberate simplification a human author chose specifically to avoid
paying for the full cross product, not something a faithful demand-driven
adornment derives mechanically. Recorded fully in
`src/transform/magicset/DESIGN.md`; not fixed here (SIPS is deliberately
not cost-based, per §3's own instruction, and there is no general rule
that says "ignore an available bound variable" without a cost model to
justify when).

### `transitive_closure_bound`: `dlc` (207) worse than Soufflé's own transform (101)

The one shape where `dlc`'s mechanical transform is measurably less
efficient than Soufflé's own automatic one, despite having **no**
negation to complicate adornment (a single `tc/bf` pair, 1 worklist
iteration). Not investigated further in this task — flagged as a
candidate open question below, since `T_excl_copy` counts every
supplementary/magic relation `dlc` materializes and Soufflé may not
generate an equivalent number of intermediate relations for the same
binding pattern.

## What a skeptic attacks first

- The `V_i` off-by-one was caught by re-deriving the formula, not by a
  test failing first — the derivation is argued in `DESIGN.md`, and its
  only *empirical* confirmation is that `sema.CheckStratification`/
  `CheckAllowedness` accept the output and Soufflé agrees on every
  comparable case. A subtly different index error that still happened to
  produce a well-typed, stratifiable, but *answer-wrong* program is not
  ruled out by this evidence alone — only by the Soufflé cross-check,
  which is exactly why it is the real gate, not the sema check.
- `transitive_closure_bound`'s Soufflé-beats-`dlc` result (101 vs 207) is
  reported, not explained beyond a hypothesis. If this pattern recurs at
  larger scale points (§9's sweep), it would be worth a real investigation
  before quoting a headline contribution number for this shape.
- Only the smallest scale point per shape was checked here (§4's own
  gate); the full-scale-point sweep, where `dlc`'s tree-walking evaluator
  is not itself in the loop (Soufflé evaluates, per the T2 protocol) but
  where derived counts could behave non-linearly, is §9's job and has not
  run yet at the time of this report.

## Verdict

**M2 (§2–§4): DONE.** Adornment, SIPS, and magic-rule/supplementary
generation are implemented, unit-tested (12 tests,
`src/transform/magicset/*_test.go`), and validated against Soufflé on 5/6
comparable cases (5/5 answer-identical); the 6th (`culprit_cycle`)
produces the exact, expected unstratifiable output the guard exists to
catch. Proceeding to M3 (`src/transform/guard/`).

## Candidate open questions

1. `transitive_closure_bound`: `dlc`'s mechanical transform materializes
   more tuples than Soufflé's own automatic one (207 vs 101) despite no
   negation involved — not investigated further this session.
2. The `bb`-vs-`bf` cost gap on the three partition-style shapes is real
   and large (up to ~194×) — worth deciding whether a future SIPS
   refinement (e.g. deliberately not using every available bound variable)
   is in scope, given §3's explicit instruction against a cost-based SIPS.
