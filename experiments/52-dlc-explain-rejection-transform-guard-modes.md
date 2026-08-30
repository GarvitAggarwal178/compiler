# NIGHT-BATCH-04 E — `dlc explain`

Date: 2026-08-27. `dlc explain <file>` (`src/cmd/dlc/main.go`'s
`runExplain`), no new checking/transform logic -- entirely a
re-rendering of what `sema`/`magicset`/`guard`'s existing exported API
already computes, as plain-text fact lines instead of JSON.

## What did not work

Nothing broke; the one design decision worth recording is that
`--transformer=` was dropped from this subcommand's own flags. Every
other subcommand (`run`, `run-seminaive`, `emit`) needs the flag because
it picks between `passthrough`/`magicset`/`guarded` as ALTERNATIVE
pipelines. `explain` instead always runs `magicset.Adorn` (for TRANSFORM
mode) and `guard.Decide` (for GUARD mode) unconditionally -- there is
nothing to explain about `passthrough`, and a user asking "why did this
happen" wants the REAL mechanical adornment and the REAL guard verdict,
not a choice of which to see.

## Format

One fact per line, `TAG key=val key=val ...`, chosen so `G`'s
presentation script can split on the first space and then `=` without a
real parser (E's own instruction). Three modes, chosen by what the
program's own front end decides, not by a flag:

- **REJECTION**: one `REJECT ground=<category> span=<line:col-line:col>
  message=<quoted>` line per diagnostic, covering the parser and all
  four sema grounds (arity, type, allowedness, unstratifiable negation)
  with the SAME `Category`/`Span`/`Message` fields `runCheck`'s JSON
  output already carries -- this mode adds no new information, only a
  different rendering.
- **TRANSFORM**: `QUERY`, `WORKLIST` (iteration count), one `ADORN`+`MAGIC`
  pair per discovered `(predicate,adornment)`, one `UNTOUCHED` line per
  never-adorned predicate, one `NEGATED` line per negated occurrence
  naming its pre- and post-relaxation adornment and whether M4-SIPS.md
  §2's relaxation applied (`relaxed=true/false`, with a reason).
- **GUARD**: `GUARD verdict=STRATIFIABLE` or
  `GUARD verdict=UNSTRATIFIABLE clause=a` plus culprit/cone/declined-
  fraction, then one `DECISION pred=<p> action=TRANSFORM|FALLBACK` line
  per IDB relation.

A program with no bindable query gets a single `NOQUERY` line instead of
TRANSFORM/GUARD output (nothing for either mode to report).

## Gate — exercised on at least one program per mode

`docs/reports/explain-samples/`, all committed:

| sample | mode exercised |
|---|---|
| `rejection_arity.dl` | REJECTION, ground=arity |
| `rejection_type.dl` | REJECTION, ground=type |
| `rejection_allowedness.dl` | REJECTION, ground=allowedness |
| `rejection_stratification.dl` | REJECTION, ground=unstratifiable (front-end level, not the guard) |
| `transform_only_ancestor_nonancestor` | TRANSFORM (no guard firing) -- also independently re-confirms A's relaxation: `NEGATED pred=ancestor pre_adorn=bb adorn=bf relaxed=true` |
| `transform_and_guard_culprit_cycle` | TRANSFORM + GUARD firing (clause a, `culprit={p,q,s}`, `cone={}`) |
| `transform_and_guard_cone_proper_subset` | TRANSFORM + GUARD firing with a genuine two-relation cone (`cone={gate1,gate1b}`) -- independently reproduces B's own finding a second way: `UNTOUCHED pred=tc`/`UNTOUCHED pred=direct`/`UNTOUCHED pred=out2`/`UNTOUCHED pred=out3` confirm the sibling relations are never adorned at all, from a completely separate code path (`main.go`'s own rendering, not `night04_b_cone_gate.py`) than the one that found this in task B. |

**7/7 samples produced exactly the expected mode**, all four rejection
grounds distinctly represented, all committed. The last sample is not
just a format demo -- it is an independent cross-check of task B's
central finding (the single-query limitation), agreeing exactly.

## Verdict

**E: DONE.** `dlc explain`, three modes, no new checking logic, gate
satisfied on 7 samples spanning every mode and every rejection ground.
Feeds G directly (both are pending); nothing else in this task is
unfinished.
