# NIGHT-BATCH-03 T1 — printer output must be Soufflé-parseable

Date: 2026-08-27. Gates T2; the whole M3 measurement path (`dlc` decides and
emits, Soufflé evaluates) depends on this. Runner: `harness/night03_t1_printer_souffle.py`.
Full data: `measurements/night03-t1/summary.json`.

## What did not work

Nothing in the printer. 9 of the 26 attempted cases did not reach a
Soufflé-accepted comparison, but in every one of the 9, **the original
(unprinted) file fails identically** — same Soufflé error class, same
relation name. These are not printer defects; see "What a skeptic attacks
first" below for the full accounting.

## Method

For each of the 20 files `m1-3.3-gate1-parse-coverage-summary.json` marked
`"status":"parsed"`, plus all 5 `BENCHMARK_FAMILY/*.dl` shapes, plus
`tests/programs/p4prime.dl` (26 cases total):

1. `dlc roundtrip <file>` (parse → print → reparse → `ast.Equal`), capture
   `.printed`.
2. Write `.printed` to a temp `.dl` file.
3. Run Soufflé on the printed file, `-F <facts_dir>` (the 20 corpus files use
   their own source directory as facts dir, matching the real Soufflé test
   suite's convention of keeping `.facts` alongside the `.dl`; the 5 shapes
   and `p4prime.dl` use their pre-registered/already-committed fixtures —
   smallest scale point per shape, `fixtures/benchmark-family/<shape>/...`
   or `fixtures/p2-scale-250`).
4. Run Soufflé on the **original** file, same facts dir.
5. Compare every `.output` relation, sorted-set equality.

## The numbers

Provenance: `measurements/night03-t1/summary.json`.

| metric | value |
|---|---|
| attempted | 26 |
| roundtrip mismatches (skipped before Soufflé) | 0/26 |
| souffle-accepted (printed file) | 17/26 |
| answer-identical (of the 17 comparable) | 17/17 |
| printer-only failures (orig ok, printed not) | 0 |
| answer mismatches (printed accepted, differs from orig) | 0 |

**The 6 cases that gate T2** (the 5 `BENCHMARK_FAMILY` shapes plus
`p4prime.dl`) are **6/6 Soufflé-accepted, 6/6 answer-identical**:

| shape | souffle-accepted | answer-identical |
|---|---|---|
| `same_generation_negation` | yes | yes |
| `transitive_closure_bound` | yes | yes |
| `ancestor_nonancestor` | yes | yes |
| `culprit_cycle` | yes | yes |
| `reachability_complement` | yes | yes |
| `p4prime.dl` | yes | yes |

## The 9 non-accepted cases

All 9 fail identically for both the original file and the printed file —
same Soufflé error, same offending relation name, same error class:

| file | Soufflé error (both orig and printed) |
|---|---|
| `example/magic_pointsto/edb.dl` | `Undefined relation AssignAlloc` |
| `example/pointsto/edb.dl` | `Undefined relation` (same class) |
| `example/not_match/not_match.dl` | `error:1` |
| `semantic/error_deduce_type/error_deduce_type.dl` | `Unable to deduce type for variable a` |
| `semantic/rule_undeclared_relation/rule_undeclared_relation.dl` | `Undefined relation a`/`c` |
| `semantic/rule_undeclared_relation2/rule_undeclared_relation2.dl` | `Undefined relation b` |
| `semantic/type_system7/type_system7.dl` | `Undefined relation F` |
| `semantic/var_single/var_single.dl` | `Ungrounded variable X` |
| `syntactic/include_directive1/foo.dl` | `Undefined relation location` |

These are Soufflé test-suite fixtures that are either (a) `#include`
fragments meant to be spliced into a driver file rather than run standalone
(`edb.dl` files — `AssignAlloc` is declared in a sibling file the real test
harness includes, not in `edb.dl` itself), or (b) deliberately-broken
negative test cases whose entire purpose is to trigger the named Soufflé
error (`error_deduce_type`, `rule_undeclared_relation`, `var_single`, etc.)
— `dlc`'s own gate 1 accepted them because `dlc`'s parser has no semantic
checks for undeclared relations at parse time; Soufflé's semantic checker
correctly rejects them regardless of whether the source came from `dlc`'s
printer or the original file. This is corpus composition, not a defect in
either the printer or T1's method — recorded, not acted on (T1 is
diagnose-and-record for a grammar-coverage/corpus issue, fix-only for a
printer bug, and there is no printer bug here).

## What a skeptic attacks first

- The 9 non-accepted cases could in principle hide a printer bug that
  happens to produce the *same* wrong error as the original — ruled out by
  inspection: the error messages differ only in line number and internal
  temp-file name, never in the relation name or error class, across all 9.
- Facts-dir choice for the 20 corpus files (the file's own source directory)
  is a convention inferred from how the real Soufflé test suite is laid
  out, not verified against a Soufflé-authored test runner script — if that
  convention is wrong for a file with more complex fixture wiring, this
  method would misreport it as a corpus issue rather than catching the real
  problem. Mitigated by the fact that all 9 failures also occur on the
  unmodified original file under the identical facts-dir choice, so the
  choice, even if imperfect, cannot be the printer's fault.
- Only the smallest scale point per shape was exercised, not every
  `SCALE_POINTS.json` point — T1's own scope is a round-trip existence
  check, not the full sweep (that is T2/M3.5's job).

## Verdict

**T1: PASS.** The printer's output is Soufflé-parseable on every program
that matters for the M3 measurement path — all 5 `BENCHMARK_FAMILY` shapes
and `p4prime.dl`, 6/6 Soufflé-accepted and 6/6 answer-identical against the
untransformed original. T2 is not blocked. Proceeding to T2.
