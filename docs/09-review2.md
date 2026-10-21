# Review 2 — current implementation draft

**Title:** A Datalog-to-C Compiler with Guarded Demand-Driven Query Optimization

**Team names/registers:** [team to supply]  
**Programme/branch:** [team to supply]  
**Guide:** [team to supply]  
**Review date:** [team to supply]

## Problem and scope

A bound Datalog query may need only part of a recursively derived relation.
Negation complicates restricting that relation: absence in an incomplete view
cannot be treated automatically as absence in the original relation. This
educational compiler implements a fixed Soufflé-compatible subset, demand
rewriting, a transformation-safety guard, full-extent preservation, and selective
fallback. Magic sets and rule unfolding are established techniques.

## Implemented prototype — rubric implementation evidence

| Responsibility | Actual code and behavior |
|---|---|
| Frontend | `src/lexer`, `src/parser`: custom tokens, precedence AST, source spans |
| Validation and planning | `src/sema`, `src/plan`: declarations/types, range restriction, signed dependencies/SCCs, executable body scheduling |
| Rewrite and safety | `src/transform/magicset`, `src/transform/guard`: demand seeds, SIPS, allocated generated names, checkpoint rules, candidate checking and fallback |
| Execution | `src/eval`: naive and semi-naive Go fixed points, stored work counters |
| Native backend | `src/codegen`: standalone POSIX C11, naive fixed-point evaluation, rule provenance comments |
| Inspection | `src/inspect`, `dlc inspect`: compiler-owned AST/graphs/plans/origins, separate source/candidate/final states |
| Demonstration | `src/demo`, `src/cmd/dlc-demo`: embedded local UI, actual subprocesses, immutable snapshots, exact answer checks, bounded failures |

The arithmetic scheduling repair is an example to defend line by line:
`p(x+1,y),q(y+1,x)` can bind columns jointly although neither original atom can
execute first. The shared planner lowers compound column tests to fresh variables
and deferred equalities only when stable scheduling stalls. It does not invert
arithmetic or admit genuinely ungrounded rules. Source AST and planned literals
remain separately inspectable. `tests/integration/arithmetic_test.go`, planner
unit tests, and the independent oracle suite cover the supplied counterexamples.

The executable workflow is:

```text
source + facts → lexer/parser → validation + body plan
              → demand rewrite + guard → revalidation
              → Go evaluation OR generated C → native compiler/executable
              → every requested output → exact row-set comparison
```

Soufflé is an independent optional runtime oracle, not the implementation of the
compiler. No TAC, SSA, parser-generator, or assembly stage is claimed.

## Verified demonstration and evidence

Run `make demo` and follow `docs/08-demo.md`. Real Chromium evidence is under
`measurements/review-2026-10-20/browser-review2-3/`; the screenshots there show
actual executions. The browser route covered edits, complete output preservation,
a new custom program, source errors and recovery, fallback, native C, the oracle,
timeouts, cancellation, and missing tools. Existing repository gates and the
supplied Review 2 black-box acceptance checks passed during phase A. Guarded
Soufflé totals for the eight selected historical configurations stayed unchanged;
raw profiles are under `measurements/review-2026-10-20/phase-a-oracle/`.

The small demo's measured materializations are 69 original and 38 guarded; its
semi-naive candidate-head attempts are 69 and 46. The full-output edit retains
12 reach pairs and materializes 50 guarded tuples. Source/fact snapshots and raw
responses support these numbers. They do not measure runtime or memory, and
finite differential agreement is not a formal proof.

## Novelty and presentation — rubric boundaries

The particular integration of demand propagation under negation, explicit
candidate safety checking, selective fallback, and preserved full extents is the
project's implementation study. Show the generated seed, a negative dependency,
a retained original, and a final C loop. Do not claim to invent magic sets under
negation or to establish production prevalence from constructed programs.

A valid source can have an invalid attempted rewrite. Fallback preserves affected
original extents while an independent branch can remain specialized. Likewise,
a specialized view and a requested full original output can coexist. Positive
controls and fallback need not improve materialization. The demo reports both
successful specialization and honest no-op cases.

The team must supply factual individual contributions; AI-assisted maintenance
must be disclosed. Names, percentages, team identity and completion percentages
cannot be inferred from Git activity. This draft does not claim total completion
or future features as implemented.

## Final-review extension and genuine limits

Phase C is now implemented as the separately named `guarded-compact` mode,
after the runnable Review 2 deliverable passed its gate. It removes only exact
generated checkpoint aliases and projection-preserving single uses. The actual
production results, invariants and adverse cases are in `docs/10-final-review.md`.
Aggressive projection-removing inlining is excluded.

The grammar remains limited; recursive arithmetic can be unbounded. Generated
C is naive. There is no cost model, parallel evaluator, runtime-speedup result,
whole-compiler proof, broad production-corpus study, or LeakSanitizer result.
Plain TSV has embedded-symbol-separator limits. Larger historical scale points
and the full upstream corpus were not rerun for this review.
