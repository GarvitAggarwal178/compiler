> Historical pass-one review, 19 October 2026. Test counts and verification scope
> below refer to that pass. The 20 October repairs, runnable demo and final pass
> are documented in `docs/09-review2.md` and `docs/10-final-review.md`.

# Datalog compiler — pass-one code review and maintenance

19 October 2026. This review concerns the uploaded `dlc` Datalog compiler.
The AuxFlow title in the class list is unrelated to the supplied source.

## Outcome

The compiler now has a checked path from Datalog source through guarded
magic-set rewriting to standalone C. Requested outputs survive specialization,
compiler-generated relation names cannot overwrite user relations, and the
identified interpreter/C disagreements have regression coverage.

This was an authorized, AI-assisted maintenance pass. Existing source and
historic experiment records were the starting point; no commit or push was
made. Presentation code is reserved for pass two.

## Findings and corrections

| Priority | Finding | Change |
|---|---|---|
| P0 | Directly requesting an optimized relation produced an empty output | Retain the original extent and its full original dependency closure where required |
| P0 | Untouched consumers read relations whose rules had disappeared | Follow positive and negative references; prevent missing rows and false negation results |
| P0 | User relations collided with adorned/magic/checkpoint names | Deterministic allocation reserves every source relation name |
| P0 | Generated roles could collide with each other | Name allocation distinguishes magic(`reach`) from adorned(`magic_reach`), even when preferred names coincide |
| P0 | Relations with both `.input` and rules lost loaded rows after rewriting | Preserve their original extents and bridge rows into demand-restricted views |
| P0 | Semi-naive recursion did not start from loaded rows | Seed the first delta from existing rows without counting input loads as derivations |
| P0 | Arithmetic terms ran before their inputs were bound | Share atom-input scheduling and bind bare columns before evaluating arithmetic |
| P0 | Bound arithmetic heads needed unsupported inversion | Admit seeds without inversion, then filter completed heads against actual demand |
| P0 | Comparison constraints could mix number and symbol types | Propagate operand types through constraint chains and reject incompatible comparisons |
| P1 | C generation bypassed optimization | Add `codegen --transformer=guarded` with shared transform/revalidation and new schemas/strata |
| P1 | Arithmetic/unary source facts could become zero in C | Generate complete constant expressions |
| P1 | C division by zero crashed and signed overflow was undefined | Add defined wrapping arithmetic and guards for invalid division/modulo derivations |
| P1 | Minimum int64 literal failed to parse | Handle the signed boundary and emit `INT64_MIN` in C |
| P1 | C truncated long fields and read at most 32 columns | Dynamic lines, exact arity, checked int64 parsing and untruncated symbols |
| P1 | Empty/nullary rows differed between backends | Preserve empty symbols; distinguish false empty files from true nullary rows; output Souffle's `()` encoding |
| P1 | Malformed numeric/column data could be accepted | Reject numeric overflow, junk, wrong column counts and NUL bytes |
| P1 | Nonempty I/O options were silently discarded | Diagnose unsupported options; retain bare directives and empty parentheses |
| P1 | Unsupported wildcard and inversion cases reached execution | Reject head/negated wildcards and ungrounded arithmetic explicitly |
| P1 | Failed source operations exited successfully | Return nonzero status while retaining diagnostics; reject unknown transformer arguments |
| P1 | Acyclic programs skipped guarded optimization | Transform and validate these programs too |
| P2 | Tests depended on `/root/compiler`; build did not produce advertised tools | Relative test paths, actual binary build targets and a portable verification harness |
| P2 | Nested results copies obscured the real source | Remove accidental `results/src` and `results/results` copies from the working release |

The uploaded bundle referred to an initial failing-regression log, but did
not contain `before-tests.txt`. This checkout preserves the successful
local verification log and new oracle evidence under
`measurements/pass1-2026-10-19/`; the missing initial log is not evidence
for any result claimed here.

## Files and design decisions

| Area | Main files |
|---|---|
| Output preservation and unique names | `src/transform/magicset/names.go`, `rules.go`, `adorn.go` |
| Grounding and operand typing | `src/sema/grounding.go`, `allowedness.go`, `decltype.go` |
| Binding order and loaded recursion | `src/eval/naive.go`, `seminaive.go`, `magicset/sips.go` |
| Checked optimized C path and CLI failures | `src/cmd/dlc/main.go` |
| C arithmetic and data handling | `src/codegen/arithmetic.go`, `evaluation.go`, `prelude.go`, `loading.go`, `storage.go` |
| Regression/oracle checks | `tests/integration/`, updated package tests |
| Reproducibility | `Makefile`, `harness/verify_pass1.py`, `docs/06-reproduce.md` |

A specialized relation answers a selected demand, not necessarily every
consumer in the source. Keeping its original declaration without its rules
does not preserve meaning. When another output or an untouched rule needs the
full relation, its original dependency closure must remain available alongside
specialized views. This can reduce optimization benefit when full outputs are
requested. That is necessary work, not a regression to conceal.

For arithmetic heads, the compiler does not solve an inverse expression. It
admits demand without binding that expression's input, computes inputs normally,
then checks the completed head against the magic seed. This can do more work
than an inverse solver, but it preserves the supported semantics.

The existing benchmark fixtures do not acquire unnecessary full extents: their
selected original tuple totals reproduce exactly after the fixes.

## Current pipeline

1. Tokenize and parse source into the AST.
2. Validate declarations, arity, types, grounding and source stratification.
3. Discover supported bound query projections and propagate adornments.
4. Generate uniquely named demand relations and supplementary checkpoints.
5. Check transformed stratification and retain necessary original extents.
6. Revalidate the actual resulting AST, schemas and strata.
7. Execute with Go naive/semi-naive evaluation, or generate POSIX C11.

Generated C still uses naive fixed-point execution. The report must not imply
that the C backend now implements delta iteration.

## Validation performed

The current suite has **128 named Go test functions**, many with subcases.
Independent programs and repeated configurations are distinguished below.

| Verification | Result and boundary |
|---|---|
| `make verify` | Build, Go vet and the complete current suite pass; see `measurements/pass1-2026-10-19/verify.log` |
| New Souffle 2.5 regression comparisons | 13 programs covering shared outputs, collisions, loaded recursion and arithmetic heads |
| Independent graph answers | 12 fixed-seed graphs checked against BFS on original/guarded naive/semi-naive paths; optimized C checked on three |
| Existing benchmark reproduction | Eight configurations: five families at selected small scales and three growing-sibling points |
| Hostile-source differential checks | 30 accepted executable fixtures agree with Souffle |
| Explicit source rejection | Eight hostile fixtures are diagnosed, not counted as execution passes |
| Filesystem limit | One 4 KB-identifier fixture is separately recorded as parser-only |
| Generated C | AddressSanitizer and UndefinedBehaviorSanitizer pass on the selected arithmetic/recursive/symbol case |
| LeakSanitizer | Not validated: address checks use leak detection disabled; no claim is made about the cause of earlier leak-sanitizer trouble |

The Go oracle test checks both evaluators on original and guarded programs,
plus optimized C. Benchmark verification also executes emitted guarded Datalog
on Souffle itself. This avoids relying only on project implementations that
share code. Commands, raw profiles, answers and summaries are preserved under
`measurements/pass1-2026-10-19/`. The README's optimized C example was
also run locally; its 34 output rows matched Soufflé as a set.

## Reproduced results

Every number uses the same Souffle 2.5 engine for the three columns. Counts
include supplementary checkpoints and exclude the documented copy/input
categories. These are materialization counts, not elapsed-time speedups.
All requested answer files agree.

| Configuration | Original | Souffle magic | Guarded rewrite |
|---|---:|---:|---:|
| Reachability complement, n=250 | 62,534 | 44,811 | 974 |
| Ancestor/nonancestor, n=500 | 250,450 | 101,250 | 6,102 |
| Same generation, d4/b4 | 51,301 | 30,907 | 681 |
| Positive transitive closure, n=500 | 86,618 | 101 | 207 |
| Culprit cycle, n=20 | 286 | 272 | 286 |
| Growing sibling, n=20 | 584 | 216 | 273 |
| Growing sibling, n=50 | 2,660 | 305 | 417 |
| Growing sibling, n=100 | 8,577 | 457 | 676 |

The selected historical totals reproduce exactly, including unfavorable
controls. The positive fragment incurs checkpoint overhead; safety fallback
need not beat Souffle's optimizer. The largest historical scale points and
the full upstream corpus census were not rerun during this maintenance.

## Using the release

The uploaded `changes/pass1.patch` applied cleanly to this checkout and
contains the code, test, build, and initial documentation changes. This
checkout is authoritative after the additional documentation cleanup and
local verification recorded here. The original patch and bundle are retained
in the recovery directory outside the repository; binaries and raw evidence
are not contained in that textual patch.

From the extracted project root:

    make verify
    DLC_SOUFFLE=/absolute/path/to/souffle make oracle

Compile an optimized example:

    bin/dlc codegen tests/corpus/BENCHMARK_FAMILY/reachability_complement.dl /tmp/query.c --transformer=guarded
    cc -std=c11 -O2 /tmp/query.c -o /tmp/query
    /tmp/query fixtures/p2-scale-250 /tmp/query-output

No transformer flag retains untransformed behavior. Use `guarded` for checked
execution. Raw `magicset` emission remains experimental. Historical milestone
scripts still contain author-specific absolute paths; the current portable
entry point is `harness/verify_pass1.py`.

## Remaining limits and deliberate deferrals

- No cost-based SIPS or join ordering. The corrected scheduling strategy remains
  easier to inspect and defend than an unvalidated cost model.
- No generated semi-naive C, adaptive indexes or parallel evaluation.
- No new aggregates, records, components, functors or general Souffle support.
- No application-corpus demonstration of the rare fallback-triggering cycle.
  A historic upstream-test census does not establish production prevalence.
- No controlled elapsed-time or peak-memory benchmark. WSL2 does not make such
  measurements impossible; they simply were not performed here.
- Differential checks and invariants are not a formal proof for arbitrary
  programs. Recursive arithmetic may generate unbounded data; tested fixtures
  are finite.
- Missing facts files still mean empty relations in dlc, unlike Souffle's usual
  error. This policy is documented rather than silently changed.
- Symbols retain plain TSV serialization. POSIX C11 is the tested native target;
  native MSVC/Windows was not validated.

A cost model, expanded grammar or large backend rewrite would introduce new
evaluation obligations and alter benchmark behavior. They are not needed to
make the existing project defensible for Review 2. Presentation of the verified
compiler is the next higher-value step.

## Pass two — presentation code plan

This is a plan only; no presentation features were implemented in pass one.
Build a local workbench around the compiler executable. Its decisions and
outputs must come from actual compiler runs.

1. **Program and facts:** editable source and a small visible input dataset.
2. **Compiler pipeline:** tokens, AST summary, semantic status, and errors with
   line/column positions. Add structured inspection output where needed.
3. **Dependency graph:** signed edges, SCCs, query demand, transformed predicates
   and full-extent fallback groups.
4. **Rewrite:** original and guarded Datalog side by side, expandable checkpoint
   details, and generated C from the same run.
5. **Results:** output tables, explicit answer equality and tuple counts with
   counting conventions labeled. Wrong answers must be displayed as failures.
6. **Safety:** a culprit-cycle/mixed-fallback example and the four rejection
   categories.

Start with reachability-complement for demand reduction, then the culprit-cycle
or mixed-fallback example for safety. Include the positive control to explain
why optimization need not reduce materialization.

A small local server and HTML/JavaScript can use the existing CLI without a
paid service. Add structured inspection data instead of parsing human display
text for correctness decisions. Provide a deterministic recorded fallback for
the presentation, clearly labeled. Report screenshots must come from the real
interface after its results are checked.

The Review 2 report can use a module table, the pipeline, real screenshots,
the result table, narrow novelty claims and truthful contributions. The actual
guide name, review date and individual contributions must come from the team;
they cannot be inferred from the archive.
