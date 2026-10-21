# dlc — Datalog compiler with guarded demand-driven optimization

A compiler written in Go: custom lexer/parser, semantic analysis, stratified
fixed-point execution, magic-set rewriting, a transformation-safety fallback,
and standalone C generation. This is a deliberately limited Datalog subset.

**Current reviews:** [Review 2](docs/09-review2.md) and
[the conservative final pass](docs/10-final-review.md).
[Pass-one maintenance](docs/07-pass1-review.md) remains a dated record.
Maintenance fixed output preservation, generated-name collisions, evaluator/backend
disagreements, arithmetic-cycle planning and optimized C integration. Structured
inspection and the local demo use actual compiler data.
Historical experiments remain in `experiments/`; the raw evidence remains
in `measurements/`. The older build plans are in `specs/`.

## Build and check

Requirements: Go 1.25+, a POSIX environment (Linux/WSL), and a C compiler with
C11 and AddressSanitizer/UndefinedBehaviorSanitizer support for the full tests.
Python 3 is used by the measurement harness. No third-party Go modules.
Install Go from the official Go distribution or your system package manager,
and install GCC or Clang, Python 3, and `make` with your system package
manager. Check the tools with `go version`, `cc --version`, and
`python3 --version`. If your Go build cache is read-only, set
`GOCACHE` to a writable directory outside this checkout.

    make verify

This builds `bin/dlc` and `bin/conecheck`, runs vet and all tests. Source errors
and malformed data now return nonzero exit codes. Missing `.facts` files retain
this project's existing convention of empty relations.

## Local review workbench

    make demo

Open **http://127.0.0.1:8080**. Edit source and facts, inspect actual compiler
AST/graphs/plans, compare Go outputs, and compile/run native C. Soufflé is an
optional answer oracle in the UI. See [the demo route](docs/08-demo.md) and
[the Review 2 draft](docs/09-review2.md). All assets are embedded and work offline.

## Structured compiler inspection

    bin/dlc inspect fixtures/demo/reachability.dl --transformer=guarded

Versioned JSON contains the actual source AST, signed dependencies, candidate
and final guard states, query demand, generation origins and executable plans.
Human `explain` now uses ORIGINAL/SPECIALIZED/FALLBACK labels and independently
reports retained full extents; worklist pops refer to adornment discovery.

## Compile an optimized program to C

    bin/dlc codegen tests/corpus/BENCHMARK_FAMILY/reachability_complement.dl /tmp/query.c --transformer=guarded
    cc -std=c11 -O2 /tmp/query.c -o /tmp/query
    /tmp/query fixtures/p2-scale-250 /tmp/query-output

The guarded AST is checked again before C generation. The generated executable
creates its output directory. Omit the transformer flag for the original
untransformed path. `--transformer=magicset` is the experimental, unguarded
rewrite; an unstratifiable candidate is rejected by execution/C generation.
Raw `emit --transformer=magicset` remains available for inspecting such a
candidate rather than executing it. Generated C uses **naive fixed-point
evaluation**, including when the guarded rewrite is selected.

Other useful commands:

    bin/dlc lex <program.dl>
    bin/dlc parse <program.dl>
    bin/dlc check <program.dl>
    bin/dlc roundtrip <program.dl>
    bin/dlc explain <program.dl>
    bin/dlc emit <program.dl> --transformer=guarded
    bin/dlc run-seminaive <program.dl> <facts-directory> <output-directory> --transformer=guarded

## Independent oracle checks

Install Souffle 2.5, then run:

    make oracle

If its executable has another location:

    DLC_SOUFFLE=/absolute/path/to/souffle make oracle

This fails if Souffle is unavailable. Ordinary Go tests skip only the optional
Souffle test when it is absent; the other regression and backend tests still run.
`harness/verify_pass1.py` resolves paths from its own location and records all
commands, raw profiles, answer files and summaries under
a new `measurements/oracle-<UTC timestamp>/` directory by default.
Set `ORACLE_OUTPUT=/new/path` for `make oracle` to select a fresh destination;
existing evidence directories are not overwritten.

## Conservative final mode

    make oracle-compact
    bin/dlc inspect fixtures/demo/reachability.dl --transformer=guarded-compact

`guarded-compact` runs the validated guard first, then removes only exact generated
checkpoint aliases or projection-preserving single uses. It retains source/I/O
relations and deduplication barriers. Inspection reports actual removals and skip
reasons; zero removals is a supported no-op. The original `guarded` mode remains
available. The same flag works for execution, emission and C generation.

## Language and data

Declare relations with `.decl name(field:number, other:symbol)`, load TSV
facts with `.input name`, and request results with `.output name`. Rules
use `head(...) :- positive(...), !negative(...), comparison.`; ground facts
use `head(...).`. The parser supports numeric and string literals,
variables, positive atoms, stratified negation, comparisons, arithmetic
`+`, `-`, `*`, `/`, `%`, unary minus, and zero-arity relations.
`_` is allowed in positive body atoms. Empty parentheses on input/output
directives are accepted. Facts and output files are tab-separated without
headers; see `docs/06-reproduce.md` for empty and nullary row handling.

The language is a fixed subset of Soufflé. It has no aggregates, records,
components, functors, general I/O options, arithmetic inversion, or
unstratified negation. Missing input files mean empty relations in this
compiler. Recursive arithmetic can produce an unbounded relation. Division
or modulo by zero suppresses that derivation; integer arithmetic wraps at
64 bits in both local backends. See `docs/02-design.md` for the grammar.

## Evidence and boundaries

Eight selected benchmark configurations were rerun after the fixes, including
all five benchmark families and three mixed-fallback scale points. Their
Souffle tuple totals reproduce the original records exactly. The larger
historical scale points were not rerun during this pass.

On reachability-complement at n=250, the same Souffle engine derives 44,811
tuples with its own transform versus 974 from this compiler's guarded rewrite.
That is a materialization result, **not an elapsed-time speedup**. On the
positive control, this compiler still materializes more checkpoint tuples.
Differential tests provide evidence for the tested programs, not a formal
correctness proof.

## Code and documentation map

- `src/lexer`, `src/parser`, `src/sema`: source parsing and validation.
- `src/plan`, `src/inspect`: shared executable scheduling and structured reporting.
- `src/transform/magicset`, `src/transform/guard`, `src/transform/compact`:
  demand rewriting, safety fallback, and conservative checkpoint removal.
- `src/demo`, `src/cmd/dlc-demo`: bounded local backend and embedded workbench.
- `src/eval`, `src/codegen`: Go evaluators and standalone C generation.
- `tests/`, `fixtures/`, `harness/`: regression programs, input data,
  and independent verification scripts.
- `docs/02-design.md`, `docs/05-limitations.md`, `docs/06-reproduce.md`:
  current design, boundaries, and verification commands.
- `docs/08-demo.md`, `docs/09-review2.md`, `docs/10-final-review.md`:
  live demo route, Review 2 draft, and measured conservative pass.
- `docs/07-pass1-review.md`: dated pass-one maintenance record.
- `experiments/`, `specs/`, `record/`, `results/`: historical experiments,
  plans, decisions, conclusions and an older HTML presentation.
- `measurements/`: original raw evidence and separate review/final verification runs.

Magic sets are established work. The project studies a particular guarded
implementation; it does not claim to invent magic sets under negation or to
have formally verified the whole compiler.
