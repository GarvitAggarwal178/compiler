# Pass-one local verification, 2026-10-19

These files were generated in this checkout after applying the uploaded patch.
The upload did not contain the `before-tests.txt` file cited by its review.

- `tool-versions.txt`: local Go, C, Python, and Souffle versions.
- `verify.log`: `GOCACHE=/tmp/dlc-go-build-cache make verify`, exit 0.
- `oracle.log`: `GOCACHE=/tmp/dlc-go-build-cache make oracle`, exit 0.
- `rejection.log`: `GOCACHE=/tmp/dlc-go-build-cache make rejection`, exit 0.
- `oracle/commands.json`: exact argv, return codes, stdout, and stderr for
  the verification harness. Subdirectories contain generated programs, raw
  Souffle profiles, and output CSVs. `summary.json` and `hostile.json` record
  the compared cases and explicit exclusions.
- `readme-example-c-sorted.csv` and `readme-example-oracle-sorted.csv`:
  sorted output from the README's guarded C example and the same source/facts
  run by Souffle 2.5. Their byte equality checks answer sets, not row order.

The example used:

    bin/dlc codegen tests/corpus/BENCHMARK_FAMILY/reachability_complement.dl /tmp/dlc-pass1-query.c --transformer=guarded
    cc -std=c11 -O2 /tmp/dlc-pass1-query.c -o /tmp/dlc-pass1-query
    /tmp/dlc-pass1-query fixtures/p2-scale-250 /tmp/dlc-pass1-query-output
    sort /tmp/dlc-pass1-query-output/q_unreach.csv

No elapsed-time measurement or full historical-scale rerun was performed.
