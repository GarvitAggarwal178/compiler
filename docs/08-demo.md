# Local compiler workbench

From the repository root:

```sh
make demo
```

Open **http://127.0.0.1:8080**. This builds `bin/dlc`, `bin/conecheck`, and
`bin/dlc-demo`. It requires Go 1.25+, Python 3 for verification, and Linux/WSL.
GCC or Clang enables native C execution. Soufflé is optional for the workbench;
Soufflé 2.5 remains required for the independent release gate. There are no
frontend dependencies, CDN resources, accounts, or network services.

If the default Go cache is read-only:

```sh
GOCACHE=/tmp/dlc-go-build-cache make demo
```

To change the loopback port or per-subprocess timeout:

```sh
make demo DEMO_ARGS='-addr 127.0.0.1:8090 -timeout 2s'
```

The default timeout is three seconds per subprocess. Native generation, compilation,
and execution each get that bound. Two requests may run concurrently; older editor
requests are cancelled. Ctrl-C shuts down the server and its active subprocesses.
Only loopback addresses are accepted. The server keeps at most eight completed
input snapshots in temporary directories outside the repository, then removes them.
This is a trusted local teaching tool; it is not a service for untrusted users.

## Five-minute route

1. Load **Two chains**. Inspect the input graph and the recursive reach rule.
   Paths have positive length: there is no reflexive rule, so node 1 appears in
   the complement query's answer. Select a source rule, AST node, or dependency
   edge; the editor selection uses the compiler's source span.
2. Explain `bf`: bound first argument and free second argument. Inspect the actual
   magic seed, selected SIPS steps, and a generated rule linked to that source rule.
   The dependency graph is head → relation read; dashed edges marked `!` are negative.
3. Run and compare. Original and guarded use the same Go evaluator on identical
   snapshots. The separate naive/semi-naive rows show repeated candidate-head work.
   Expand every output and the per-relation counters, including checkpoints.
4. Change `missing(1,y)` to `missing(5,y)`, restore it, and add graph edge 4 → 5.
   Results clear immediately on edits. Restore the example and add `.output reach`
   to show why a full original extent may coexist with a specialized view. Add a
   second constant query to see two real demand seeds. Removing constants is a
   supported no-op specialization.
5. Load **Constructed mixed fallback**. The source is valid; the attempted rewrite
   has a negative cycle; the final guarded program is valid. Inspect the p/q/s
   fallback and surviving tc specialization. These are constructed examples.
6. Generate and run C. Open its source, find a rule provenance comment, a join,
   and `while (changed)`. The C backend evaluates naively. Its output check uses
   actual compiled native execution; the Go counters are not C instrumentation.

## Edits, errors, and bounded failures

The source editor and declared `.facts` editors are authoritative. Inspect discovers
input names from the compiler AST. The graph edits those same edge/node facts.
A custom program does not have to appear in the example manifest. Inspection,
execution, and native code all record a run ID, exact source SHA-256, input-snapshot
SHA-256, mode, and real command outcomes. Native/oracle actions use that immutable
snapshot, even if a newer run exists. An edit clears the displayed old results.

A type, arity, grounding, parse, or stratification failure exposes the actual
compiler diagnostic and source location. Execution and C actions are unavailable.
Unbounded recursive arithmetic reports timeout, kills the process group, and reaps
the launched process; it does not report an empty answer as success. A new ordinary
run remains usable after a timeout. Missing C and oracle tools are displayed
separately. Native execution works when Soufflé is unavailable.

No output declaration means no reported answer verification. Comparisons require
the exact expected CSV-file manifest and exact byte/LF row sets, not counts.
Nonempty outputs must end in LF; empty files, empty symbols, and nullary `()` are
distinct. Unicode U+2028 is data, not a row separator. Plain TSV cannot encode an
embedded tab or LF inside a symbol unambiguously. Rows that do not match the
schema are displayed as escaped bytes with this limitation rather than reshaped.

Limits are 64 KiB source, 512 KiB per facts file, 2 MiB total facts, and 2 MiB
output/C display. The bounded stdout/stderr capture also fails explicitly on
excess output. A limit failure is not a correctness result. Large historical
benchmark fixtures belong in the CLI verification harness.

## Observed browser and backend checks

The 20 October run used real headless Chromium through Playwright and the local
`make demo` server. All teacher mutations, custom edits, failures, mixed fallback,
C compilation/execution, live oracle answers, timeout cleanup, rapid edits, and
controlled missing-C/oracle states passed. Commands, submitted snapshots, responses,
and two actual screenshots are in
`measurements/review-2026-10-20/browser-review2-3/`. Earlier failed browser attempts
remain alongside it. The first exposed a narrow SVG click target; a subsequent
attempt needed to open the C panel before asserting visible text.

On the original small graph, locally observed Go semi-naive counts were original
69 / guarded 38 derived materializations and 69 / 46 candidate-head attempts.
With `.output reach`, the guarded materialization was 50 and all 12 reach pairs
were preserved. These are finite work measurements, not runtime speedups.

To repeat browser acceptance, install/use Playwright and Chromium outside this
checkout, start `make demo`, and run the interpreter that provides Playwright:

```sh
python3 tests/demo/browser_smoke.py --url http://127.0.0.1:8080 --repo . \
  --output /tmp/dlc-browser-new --stage review2
```

The script requires a new output directory and fails on an unsuccessful browser
check. It also launches controlled missing-tool servers on temporary loopback ports.
`make verify` includes backend snapshot, exact-output, native, timeout, child-process,
missing-tool, and bounded-retention tests. Controlled tests also time out a C
compiler with a child and a real compiled native infinite loop, checking cleanup. A sandbox that prohibits loopback sockets
needs permission for those tests; an HTTP test cannot pass without a listener.

The final UI also offers `guarded-compact`. Expand checkpoint compaction to see
actual removed definitions and conservative skip reasons. The **projection / dedup
barrier** example makes a retained existential projection visible; custom source
with no checkpoint removal is labelled a no-op. The full browser route was rerun
with `--stage final`, including compact C and source-linked final graphs. All
14 groups passed; evidence is in `measurements/final-2026-10-20/browser-release/`.
Use a new output directory and `--stage final` in the command above to repeat it.

There is no tuple witness animation or global execution clock. Worklist pops are
adornment discovery; naive bucket 0 and semi-naive SCC-local buckets do not form
a global trace. Browser evidence covers Chromium on Linux/WSL, not every browser
or native Windows/MSVC. LeakSanitizer was not run.
