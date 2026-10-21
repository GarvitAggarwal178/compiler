# Reproducing the current compiler

Use Linux or WSL, Go 1.25+, Python 3 and GCC/Clang with C11 support. The full
backend tests use AddressSanitizer and UndefinedBehaviorSanitizer. The current
build/test path has no fixed `/root/compiler` dependency.

## 1. Build, vet, test

    make verify

The Makefile disables VCS stamping so an extracted source archive builds even
when a surrounding workspace has incomplete Git metadata. No commits are needed.
`make build` produces `bin/dlc` and `bin/conecheck`; `make demo` also builds
`bin/dlc-demo` and serves the offline workbench at http://127.0.0.1:8080; `make rejection` verifies
that all four committed semantic rejection examples fail as expected.

## 2. Independent oracle

Use Souffle 2.5 for comparisons with the historic figures. The upstream release
used by this project is:

https://github.com/souffle-lang/souffle/releases/download/2.5/x86_64-ubuntu-2404-souffle-2.5-Linux.deb

Install its runtime dependencies through your normal package manager. Confirm
`Version: 2.5 (2.5)` with `souffle --version`, then run:

    DLC_SOUFFLE=/absolute/path/to/souffle make oracle

By default the harness creates a new `measurements/oracle-<UTC timestamp>/`
run. Use `ORACLE_OUTPUT=/new/directory make oracle` to choose a fresh location.
It refuses to overwrite an existing evidence directory. Historical scripts retain
historical paths; this current harness discovers inputs/outputs from compiler
inspection and uses `harness/answers.py` for exact LF/byte sets and file manifests.

The optional Go differential test can also be run directly:

    DLC_SOUFFLE=/absolute/path/to/souffle go test -count=1 -run TestSouffleDifferential ./tests/integration

`make oracle` checks all five benchmark families at a selected small scale,
three growing-sibling points, and supported hostile-source fixtures. It compares
all requested output relations, not only the main query. Profile totals for
original, Souffle-transformed and dlc-transformed programs are all measured on
the same Souffle executable. C output and both Go evaluators are also checked
on the benchmark configurations. The regular tests cover additional collision,
shared-output, arithmetic, loaded-recursion and data-loading cases.

To verify the separate production compaction mode with eight configurations,
24 independent graphs, actual compact C and the projection control:

    make oracle-compact

`COMPACT_OUTPUT=/new/directory` chooses its fresh evidence destination. The
repository-owned `harness/verify_compact.py` records profiles, counters, hashes,
compaction actions and complete command outcomes. Soufflé totals and Go insertion/
attempt counters remain labelled separately; engine internals can differ.

## 3. Optimized C path

    bin/dlc codegen tests/corpus/BENCHMARK_FAMILY/reachability_complement.dl /tmp/query.c --transformer=guarded
    cc -std=c11 -O2 /tmp/query.c -o /tmp/query
    /tmp/query fixtures/p2-scale-250 /tmp/query-output

To inspect the source rewrite or decision:

    bin/dlc emit tests/corpus/BENCHMARK_FAMILY/culprit_cycle.dl --transformer=guarded
    bin/dlc explain tests/corpus/BENCHMARK_FAMILY/culprit_cycle.dl
    bin/dlc inspect fixtures/demo/reachability.dl --transformer=guarded

`inspect` emits schema-version-1 JSON from real AST, graph, planner, generator and
guard data. `guarded-compact` is also accepted by inspect/emit/run/run-seminaive/
codegen; it retains the guarded candidate report and revalidates the compact final
program. The C path remains naive. `explain` reports ORIGINAL/SPECIALIZED/FALLBACK,
separate full-extent retention and source/candidate/final statuses.

## 4. Data contract

Input/output files are tab-separated with no header. Numeric columns must be
valid signed int64 values, with exact column counts. Symbols are not truncated.
A blank row in a one-symbol-column input means an empty symbol. A true nullary
relation is written as `()` followed by a newline, matching Souffle 2.5. The
loaders also accept the project's older blank-row nullary input convention;
an empty file means false. Missing facts files still mean empty relations in
dlc (Souffle itself normally errors), so oracle fixtures must exist.

Nonempty I/O options, anonymous head terms, negated wildcards, arithmetic
inversion, and NUL-containing symbols are rejected explicitly. Empty I/O
parentheses remain supported. Arithmetic wraps as int64 in both dlc backends;
a failed division/modulo-by-zero expression produces no derivation. This is
an explicitly tested local arithmetic policy, not a claim about every engine.

The current output readers require LF termination for nonempty files and split
only on byte LF. Unicode line separators are not row boundaries. Plain TSV has no
unambiguous representation for a tab/LF embedded in a symbol; the UI shows
ambiguous rows as escaped bytes. Empty output files still participate in the
expected file manifest. No declared outputs means no verified-answer badge.

## 5. Browser acceptance

`docs/08-demo.md` gives the runnable teacher route and external Playwright setup.
With `make demo` running, use an interpreter with Playwright/Chromium installed
outside the checkout:

    python3 tests/demo/browser_smoke.py --url http://127.0.0.1:8080 --repo . --output /tmp/dlc-browser-new --stage final

This actually submits editor mutations, checks stale-state invalidation and
source-linked inspection, executes native C, triggers errors and bounded recursion,
checks process cleanup, and starts controlled missing-tool servers. It requires
loopback access. An inaccessible browser or listener is a failure/pending check,
not a substitute HTTP-only pass.

## 6. Historical material

The old `harness/night*.py` and milestone scripts are preserved historical
records. Many still refer to the author's `/root/compiler` directory. They
are not the portable verification entry point; use `harness/verify_pass1.py`.
The full upstream Souffle test checkout is not vendored. Historic corpus
coverage figures and largest-scale measurements therefore must not be described
as freshly reproduced by this maintenance pass.

The 4 KB-identifier fixture is parser-only during execution checks: its relation
name exceeds the filesystem's filename limit. It is recorded separately from
successful differential runs. The sanitizer test checks address and undefined
behavior with leak detection disabled. No successful LeakSanitizer run is claimed
for this integration.
