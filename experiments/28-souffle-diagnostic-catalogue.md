# NIGHT-BATCH-02 T9 — Soufflé diagnostic catalogue

Date: 2026-08-23. Independent task, no Lane A dependency. Runner:
`harness/night02_t9_diagnostics.py`, which runs all 13 existing
`tests/rejection/*.py` cases (4 grounds: arity, type, stratification,
allowedness) plus one new minimal program for "undeclared relation"
(`tests/programs/diagnostic_undeclared_relation.dl`) against installed
Soufflé 2.5. Full data: `measurements/night02-t9-diagnostics-summary.json`.

**All 14 programs reject, and every one's actual rejection reason is
consistent with its `expected_ground` label — no mismatches found.**

## Correction, found while confirming a T5 finding (logged first, per
CLAUDE.md §7's instruction to write "what did not work" before results)

**T5's report (`docs/reports/night02-T5-guarded.md`) originally claimed
Soufflé could exit 0 on a stratification error.** That claim came from one
interactive Bash-tool invocation through the `wsl.exe` bridge that showed
`RC=0` next to `Error:` text on stderr. Re-run 5 times this task — 2 via
`subprocess.run(capture_output=True)` (what `harness/probe0.py`'s
`run_cmd` actually uses, and what every measurement in this batch is built
on) and 3 via direct shell redirection to files — **every single re-run
returned rc=1.** Root cause: the earlier interactive multi-command Bash-
tool invocation streamed output live through the `wsl.exe` bridge, and
that streaming raced/garbled across the bridge (confirmed independently:
a later, differently-structured interactive command produced visibly
corrupted, interleaved terminal text with a stray blank `RC=` line — not a
Soufflé problem, a display problem). **File-redirected invocation (what
every actual measurement script in this project does — none of them
stream interactively) is reliable; watching an interactive multi-line
command's live output through this bridge is not, at least for
determining an exit code.** Fixed: `docs/reports/night02-T5-guarded.md`
now carries a marked correction (not a silent edit) in place of the
original claim; both harness files whose comments cited the retracted
claim (`harness/night02_t5_guarded.py`, `harness/night02_t2_hostile.py`)
were updated. No `T_guard` number or any other measurement in T5 was
itself affected — the retracted claim was a side observation about the
rejection mechanism, never an input to any computed number.

This is not a CLAUDE.md §5 "two runs of the same command produce different
output" STOP condition: that clause is about Soufflé's own determinism
guarantee, and Soufflé was never shown to behave differently across
identical runs — 5/5 reliably-captured runs agree. The disagreement was
between a reliable capture method and an unreliable one, fully diagnosed,
not between two runs of Soufflé.

## The catalogue

| error class | minimal trigger | exact message (first line) | line/col? |
|---|---|---|---|
| Ungrounded variable | `bar(x,y):-foo(x).` (y unbound) | `Error: Ungrounded variable y` | yes (file+line+caret) |
| Arity mismatch | `foo` decl'd arity 2, used with 1 arg | `Error: Mismatching arity of relation foo (expected 2, got 1)` | yes |
| Type mismatch (inference) | `symbol` value flows into a `number`-typed variable | `Error: Unable to deduce type for variable y` + `Following constraints are incompatible: ...` | yes |
| Type mismatch (literal) | `.decl foo(a:number)` fed a string fact | `Error: Symbol constant (type mismatch)` | yes |
| Undeclared relation | body atom for a relation with no `.decl` | `Error: Undefined relation undeclared_rel` | yes |
| Unstratifiable negation | `a(x):-a(x),!a(x).` | `Error: Unable to stratify relation(s) {a}` + `... has cyclic negation ...` | yes (two locations: the `.decl` and the cyclic rule) |
| Syntax error | missing terminator, unbalanced parens, etc. (T2, this batch) | `Error: syntax error, unexpected <token>[, expecting <token>]` | yes |
| Duplicate declaration | same relation `.decl`'d twice (T2, this batch) | `Error: Redefinition of relation p` + `Previous definition in file ... at line N` | yes (two locations) |

**Type mismatch has two distinct message shapes** depending on whether the
mismatch is discovered by inference (`Unable to deduce type for variable`,
with a `Following constraints are incompatible` detail block) or is a
direct literal-vs-declared-type conflict (`Symbol constant (type
mismatch)`, no constraint-solving detail) — a rejection-corpus case or a
future `dlc` diagnostic that only recognizes one shape would miss the
other.

**Every class carries a `file ... at line N` location and a caret-pointer
line** except syntax errors that occur at end-of-file (still gets a line
number, just no source line to point the caret at meaningfully beyond
"end of file"). No class in this catalogue is silent about location.

## Cross-check: every `tests/rejection/` case against real Soufflé

| module | case | Soufflé outcome | consistent with `expected_ground`? |
|---|---|---|---|
| arity | `arity_too_few_args` | reject, arity mismatch | yes |
| arity | `arity_too_many_args` | reject, arity mismatch | yes |
| arity | `arity_mismatch_between_fact_and_decl` | reject, arity mismatch | yes |
| type | `type_number_vs_symbol_across_rule` | reject, type inference conflict | yes |
| type | `type_symbol_in_arithmetic` | reject, type inference conflict | yes |
| type | `type_mismatched_fact_literal` | reject, symbol-constant type mismatch | yes |
| stratification | `stratification_self_negative_cycle` | reject, unstratifiable | yes |
| stratification | `stratification_mutual_negative_cycle` | reject, unstratifiable | yes |
| stratification | `stratification_cycle_through_positive_chain` | reject, unstratifiable | yes |
| allowedness | `allowedness_head_var_unbound` | reject, ungrounded variable | yes |
| allowedness | `allowedness_var_only_in_negation` | reject, ungrounded variable | yes |
| allowedness | `allowedness_var_only_in_constraint` | reject, ungrounded variable | yes |
| allowedness | `allowedness_equation_rhs_not_bound` | reject, ungrounded variable | yes |

**13/13 consistent.** No case's `expected_diagnosis` (recall: descriptive
prose about the reasoning, not a verbatim-match target, per J2's
established convention) contradicts what Soufflé actually reports. Nothing
in the rejection corpus needs revision from this cross-check.

## What a skeptic attacks first

- The type-mismatch catalogue entry's "inference" message shape includes a
  third error line in the raw output
  (`type_symbol_in_arithmetic`'s stderr has *three* separate `Error:`
  blocks, not one) — this report shows only the first line per class for
  readability; a case built only on the first line risks missing that
  Soufflé sometimes reports the same root problem multiple times from
  different angles.
- "Consistent with `expected_ground`" was judged by reading each
  diagnostic and matching it to the case's own English description, not by
  a mechanical string-match rule — this is a human(-equivalent) judgment
  call, disclosed as such, not an automated proof.
- The correction above is itself now unverifiable by a future reader
  without re-running the exact same interactive-vs-redirected comparison;
  it is reported at the confidence level "5/5 reliable-method runs agree,
  1/1 unreliable-method run disagreed," not higher.

## Provenance

`measurements/night02-t9-diagnostics-summary.json` (14 rows, all captured
via `probe0.run_cmd`, i.e. `subprocess.run(capture_output=True)`, the
reliable method). Correction evidence:
`measurements/_scratch_night02_t9/rc_check_output.txt` and
`rc_check2_output.txt` (both file-redirected, both consistent, 5/5 rc=1).
New trigger program: `tests/programs/diagnostic_undeclared_relation.dl`.
Runner: `harness/night02_t9_diagnostics.py`. Completed inside the
45-minute cap (excluding the correction side-investigation, logged
separately above per CLAUDE.md §7's "what did not work first" instruction
since it directly affects trust in this batch's other reports).
