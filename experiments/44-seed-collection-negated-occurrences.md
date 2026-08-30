# M3.1 — seed collection over negated occurrences (§5)

Date: 2026-08-27. `src/transform/guard/seeding.go`. Prerequisite to the
culprit-cycle detector — "the part the correctness of everything else
rests on."

## What did not work

The counterexample search's first draft mis-detected which programs
actually got a real transform: comparing the printed output's raw text
against the original source byte-for-byte is **not** a valid no-op test,
since `parser.Print` always reformats (full parenthesization, stripped
comments, canonical spacing) even for an unchanged AST. Fixed by checking
for the presence of `magic_`/`sup_` relation names instead — a real
transform always introduces them, a pass-through never does.

## The assertion (§5's central claim)

`AssertNegationAllBound` walks every negated IDB atom occurrence
`magicset.Adorn` discovers and checks its adornment is all-bound.
`src/transform/guard/DESIGN.md` states the 3-sentence argument for why
this must hold given allowedness. 2 unit tests
(`src/transform/guard/seeding_test.go`): the assertion accepts a real
shape (`ancestor_nonancestor`'s `!ancestor(x,y)` occurrence), and a
second test confirms the check inspects at least one real occurrence
rather than vacuously passing on an empty list.

## The counterexample search (required)

Provenance: `measurements/m3-1-counterexample/summary.json`. Method:
`harness/night_m3_1_counterexample.py` — the 4 `tests/programs/
p6*_base.dl` constructions (Phase 0.6's original counterexample-hunt
corpus) plus all 6 `tests/corpus/CULPRIT_CANDIDATES/` programs
(NIGHT-BATCH-03 T4), `dlc`'s real transform (`dlc emit
--transformer=magicset`) vs. the untransformed original, compared via
`harness/m2_accept.py`.

| | count |
|---|---|
| attempted | 10 |
| comparable and answer-identical | 5 |
| correctly unstratifiable | 5 |
| **disagreements (the counterexample)** | **0** |

**Per-file breakdown:**

| file | got a real transform? | result |
|---|---|---|
| `p6a1_base.dl` | no (no bindable query, `FindQuery`'s literal reading) | no-op, trivially agrees |
| `p6a1b_base.dl` | no | no-op, trivially agrees |
| `p6a2_base.dl` | no | no-op, trivially agrees |
| `p6start_base.dl` | no | no-op, trivially agrees |
| `cc_edb_negated.dl` | yes | **agrees**, `T_none`=870, `T_dlc`=54 |
| `cc_arity3_twobound.dl` | yes | correctly unstratifiable |
| `cc_longer_cycle.dl` | yes | correctly unstratifiable |
| `cc_neg_early.dl` | yes | correctly unstratifiable |
| `cc_query_bothbound.dl` | yes | correctly unstratifiable |
| `cc_third_relation.dl` | yes | correctly unstratifiable |

**The 4 `p6*_base.dl` files never exercise the real mechanism** — all 4
have their bindable constant one or more projection steps away from the
`.output` relation's own rule, which `FindQuery`'s literal reading of
M2-M3-BUILD.md §2 does not chase (`src/transform/guard/DESIGN.md` has the
exact per-file reason). Confirmed directly (0 occurrences of `magic_`/
`sup_` in any of their `dlc emit` output), not assumed. This means the
search's real evidentiary weight rests on the 6 `CULPRIT_CANDIDATES`
programs, which were purpose-built (NIGHT-BATCH-03 T4) to match the direct
pattern and do exercise the transform.

**All 5 of T4's structurally-flagged culprit-cycle candidates produce
`Unable to stratify` under `dlc`'s own mechanical transform — matching the
classifier's predictions exactly, file for file.** This is independent
cross-validation from a completely different code path (a Go
implementation of the actual worklist adornment algorithm, versus T4's
regex-based structural precondition scan): the same 5 files, the same
verdict, arrived at two different ways. The 1 negative control
(`cc_edb_negated.dl`, negated predicate is EDB) correctly stratifies and
agrees, with a real (if modest) reduction.

## What a skeptic attacks first

- Zero counterexamples found is a null result on a search that only
  meaningfully tested 6 programs (the `CULPRIT_CANDIDATES` set), not 10 —
  the `p6*` files' no-op status means they contribute no evidentiary
  weight to the collapse hypothesis, despite counting toward the raw
  "10 attempted" figure. Reported honestly above, not folded into an
  inflated "10 tested" headline.
- The assertion itself (`AssertNegationAllBound`) cannot currently fail
  for any program `magicset.Adorn` can produce, by construction — it is a
  defensive check against a future regression, not something this
  session's test suite demonstrates catching a real violation (there is
  none to catch). Its value is prospective.
- 6 constructed programs is still a small search. The five stratification
  failures are the expected/designed-for outcome and are informative
  about the classifier-transform correspondence, but they are not
  themselves evidence *for* clause (b)'s completeness claim (an
  unstratified program is never evaluated at all) — the actual
  completeness evidence is the 1 comparable agreement
  (`cc_edb_negated.dl`) plus M2's own 5/6 gate from `docs/reports/
  m2-headline.md`.

## Verdict

**M3.1: DONE.** Assertion implemented and tested. Counterexample search:
0 disagreements across 10 attempted (4 no-op, 6 real), with the 6 real
cases split 1 agreed / 5 correctly-unstratifiable, the latter matching
NIGHT-BATCH-03 T4's structural predictions exactly. No counterexample to
the collapse hypothesis found. Proceeding to M3.2 (culprit-cycle
detection, clause a).
