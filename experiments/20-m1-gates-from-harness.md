# J2 — M1 gates, finished from NIGHT-BATCH-01 T7

Date: 2026-08-22. Lane B only. `src/` not touched — the parser and
pretty-printer these gates test against remain the deliberate stubs T7 first
established (`harness/differential.py`'s `run_dlc`, extended here with two
parse/print-specific counterparts).

## 1. Parse-coverage runner

`harness/parse_coverage.py` — iterates `tests/corpus/IN_GRAMMAR.txt` (195
files), invokes `harness/dlc_interface.run_dlc_parse` (new shared stub module,
see below) per file, reports parsed/not_implemented/failed/errored with
per-file diagnostics. **Today: 195/195 `not_implemented`.** Exits nonzero if
any file reports anything else — the gate itself is the assertion, not a
number a human has to eyeball. `measurements/j2-parse-coverage-summary.json`.

This is the human's day-2 acceptance gate: it exists and is runnable now, so
the first real parser commit has an immediate, pre-built coverage number to
report against instead of one improvised after the fact.

## 2. Round-trip scaffold

`harness/round_trip_scaffold.py` — parse → pretty-print → reparse → assert
structural equality (`ast1 == ast2`), over the same 195-file pool. The
pretty-printer is Lane A and does not exist (`dlc_interface.run_dlc_pretty_print`,
stub); this file is the harness and the comparison logic, which is the actual
Lane B deliverable — the printer itself is deliberately left unimplemented.
**Today: 195/195 `not_implemented`**, short-circuiting at the first stub call
(parse) for every file — never a silent skip, never a false "match".
`measurements/j2-round-trip-summary.json`.

`harness/dlc_interface.py` is new: a small shared module holding both stubs
(`run_dlc_parse`, `run_dlc_pretty_print`) so the parse-coverage runner and the
round-trip scaffold share one contract instead of drifting. Kept separate
from `differential.py`'s `run_dlc` (full compile+run against Soufflé,
T7) — parsing and printing don't need a facts directory or execution.

## 3. Rejection-test corpus — allowedness cases revised against J1

`tests/rejection/allowedness.py`'s three original cases (T7) were written
before any empirical check against Soufflé. J1 (`docs/reports/
J1-allowedness-probe.md`, same session, run first) now cross-validates all
three directly:

- `allowedness_head_var_unbound` — generalizes J1's ungrounded-variable
  rejections.
- `allowedness_var_only_in_negation` — is J1 case (f) exactly.
- `allowedness_var_only_in_constraint` — is J1 case (c) exactly.

**One case added:** `allowedness_equation_rhs_not_bound`, from J1's least
obvious finding (case d): an already-grounded variable on the left of an
equation does not ground an unbound variable on the right —
`bar(x,y):-foo(x),x=y+1.` rejects even though `x` is grounded. **This exact
program was independently re-run against Soufflé 2.5 as part of writing this
case** (not just inferred from J1's structurally-similar probe) and confirmed:
`Error: Ungrounded variable y`.

13 cases total across 4 grounds now (was 12).

## 4. Verification — everything still fails cleanly

| Check | Result |
|---|---|
| `harness/run_rejection_tests.py` | 13/13 `not_implemented` |
| `harness/parse_coverage.py` | 195/195 `not_implemented`, exit 0 |
| `harness/round_trip_scaffold.py` | 195/195 `not_implemented`, exit 0 |
| `harness/test_golden_guard.py` (T7, unaffected) | 3/3 passed |

No case anywhere passes vacuously. A green run today means "correctly blocked
on Lane A," not "the parser works" — there is no parser yet.
