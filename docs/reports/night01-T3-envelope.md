# NIGHT-BATCH-01 — T3: recoverable-envelope sweep

Date: 2026-08-20. Outcome: **aborted mid-sweep** at `semantic/subsumption_multiple_rules`
per T3's explicit abort condition (answer relations diverged between
configurations). Full escalation: `docs/ESCALATIONS.md`. 27 of 31 candidates
produced clean, trustworthy data before the abort; 2 never ran (queue order
truncated by the abort); 1 crashed; 1 diverged.

## What did not work, first (per CLAUDE.md §7)

The harness's first comparison method was wrong: a raw byte-diff against the
untransformed `.csv` flagged `example/orbits1` as diverging, which on inspection was
only a row-order difference (`diff` showed the same two rows transposed). CLAUDE.md
§6 already specifies the correct method ("set equality on output relations... sort,
then compare") — the harness had simply not followed its own project's rule. Fixed
(`harness/night01_t3_envelope.py`, sort before comparing) and the full sweep re-run
from scratch before any number here was trusted. `example/orbits1` is clean under
the corrected comparison (`T_souffle`=42, `E_recoverable`=3, `docs/reports/
night01-T3-envelope.md` table below).

After the fix, `semantic/subsumption_multiple_rules` still diverges — genuinely, in
sorted-set content, across 4 relations. This program uses Soufflé's subsumption
operator (`<=`) with `btree_delete`, a feature entirely outside blueprint §4's
grammar; the specific diverging rule chain has no negation in it at all. Full
characterization and three live explanations in `docs/ESCALATIONS.md`
(2026-08-20, T3). Not investigated further tonight, per instruction.

`example/tic-tac-toe` crashed under `--magic-transform=*` (`std::bad_alloc` inside
Soufflé, `returncode=-6`) — plausibly the 8GB address-space cap manifesting as a
C++ allocation failure rather than a clean Python-level DNF (the resource limit
applies to the child process; its own allocator throws, not our wrapper). Recorded
as an error status, not silently retried at a higher cap (batch protocol).

## Method

Prerequisite: T2's 31 `ok`+`seedable` programs. For each, ran
`--magic-transform=*` (T2 already has the untransformed baseline); `E_recoverable`
= Σ tuples over every `@neglabel.`-prefixed relation (`harness/tuple_report.py`,
validated against all four Phase 0/0.5 ground-truth numbers before use, see
`docs/reports/night01-T2-corpus.md`). Global caps applied (300s, 8GB).

## Per-program table (27 clean + 2 non-clean; 2 never reached)

| Program | `T_none` | `T_souffle` | `E_recoverable` | `@neglabel.` relations | Note |
|---|---|---|---|---|---|
| evaluation/bad_regex | 1 | 2 | 0 | — | |
| evaluation/components | 35 | 114 | 3 | `SocialNet.reachable` | |
| evaluation/inline_negation1 | 3 | 3 | 0 | — | |
| evaluation/magic_2sat | 1,338 | 1,338 | 1 | `+disconnected0` | |
| evaluation/magic_components | 114 | 114 | 3 | `SocialNet.reachable` | |
| evaluation/neg1 | 6 | 10 | 0 | — | |
| evaluation/neg2 | 9 | 16 | 0 | — | |
| evaluation/neg3 | 14 | 21 | 0 | — | |
| evaluation/neg4 | 9 | 16 | 0 | — | |
| evaluation/neg5 | 15 | 22 | 0 | — | |
| evaluation/set_ops | 40 | 28 | **16** | 12 relations (a2b/a2c/b2c \*) | max ratio, 57% |
| example/2sat | 200 | 1,338 | 1 | `+disconnected0` | |
| example/array | 50 | 113 | 10 | `@interm_out.element` | |
| example/cellular_automata | 165 | 366 | 33 | `@interm_out.element` | |
| example/dfa_parse | 134 | 254 | 9 | `accept` | |
| example/earley | 28 | 71 | 10 | `NotEndProd`, `NotEndString` | |
| example/flights | 20 | 34 | 4 | `@interm_out.VAflies` | |
| example/fractional_knapsack | 14 | 43 | 8 | 4 relations | |
| example/game2 | 18 | 34 | 5 | `@interm_out.canMove`, `@interm_out.possible_winning` | |
| example/grid | 99 | 254 | 9 | `@interm_out.element` | |
| example/minesweeper | 1,100 | 2,571 | 100 | `@interm_out.element` | |
| example/orbits | 7 | 19 | 1 | `@interm_out.intermediate` | |
| example/orbits1 | 7 | 42 | 3 | `@interm_out.intermediate` | initial byte-diff false alarm, see above |
| example/topological_ordering | 344 | 2,796 | 147 | 3 relations | |
| provenance/constraints | 4 | 4 | 0 | — | |
| semantic/eqrel_tests | 344 | 1,499 | 0 | — | |
| semantic/rel_nullary | 23 | 21 | 0 | — | |
| example/tic-tac-toe | 276,168 | — | — | — | **crash**, `std::bad_alloc` |
| semantic/subsumption_multiple_rules | 31 | 100 | 0 | — | **diverged**, excluded from distribution below |
| swig/java/flights | — | — | — | — | not run (abort truncated the queue) |
| swig/python/flights | — | — | — | — | not run |

Raw provenance: `measurements/night01-t3/<path>/`, `measurements/night01-t3/summary.json`.

## `E_recoverable / T_souffle` distribution (27 clean programs only)

| min | Q1 | median | Q3 | max |
|---|---|---|---|---|
| 0.0 | 0.0 | 0.026 | 0.088 | 0.571 (`evaluation/set_ops`, 16/28) |

**`E_recoverable = 0` on 10 of 27 (37%).** These are programs where Soufflé's own
transform never isolated any relation (`neglabel_relations` empty) — the gap this
project targets does not exist on them, and they count exactly as much as the
programs where it does. Listed in full: `evaluation/bad_regex`,
`evaluation/inline_negation1`, `evaluation/neg1`–`neg5`, `provenance/constraints`,
`semantic/eqrel_tests`, `semantic/rel_nullary`.

**Read `E_recoverable` as an upper bound, never as `T_guard`.** It is exactly what
Soufflé leaves fully materialized under a `@neglabel.` isolation — not what a
guarded transform would actually recover, which requires either the compiler (does
not exist) or a hand-transform (cherry-pickable, not attempted here). No ratio in
this report is `T_souffle/T_guard`; that number does not exist yet for any corpus
program.

## What a skeptic attacks first

- 37% zero rate: the "gap" this project targets is absent on over a third of the
  already-filtered pre-registered set. Whether that's representative of Datalog
  programs with negation generally, or an artifact of this specific 36-program
  corpus, is not something 27 data points settles.
- The one genuine divergence found lives in a feature (subsumption) `dlc` will never
  implement — reassuring for scope, but it does mean this specific corpus contains
  at least one program that was never going to be safely runnable through this
  pipeline regardless of the negation question.
- Two candidates never got a magic-transform run at all (queue truncated by the
  abort) — the distribution above is n=27, not n=31.
