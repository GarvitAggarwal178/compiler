# Subsumption divergence — minimization and upstream status

Date: 2026-08-21. Corpus ruling §4.1, cap 2 hours (used well under cap — resolved
by finding the issue already reported and already fixed upstream). Lane B.

## Background

NIGHT-BATCH-01 T3 found `semantic/subsumption_multiple_rules` (part of
`tests/corpus/PREREGISTERED.txt`) produces different answer sets with and without
`--magic-transform=*`, surviving sorted set-equality comparison (`docs/
ESCALATIONS.md`, 2026-08-20). Not this project's correctness problem — subsumption
(`<=`/`btree_delete`) is outside `dlc`'s grammar and already excluded from the
correctness corpus (`tests/corpus/IN_GRAMMAR.txt`, T5) — but a real discrepancy in
the oracle worth tracking down.

## Step 1: isolate the diverging fragment

`tests/programs/subsumption_isolated.dl` — the `A`/`graph` shortest-path-tree
fragment only, copied verbatim out of the original bundled test file. Reproduces
cleanly: sorted `A.csv` differs by exactly one tuple, `(5,6,3)`, present under
`--magic-transform=*` and absent without it (measurements:
`measurements/subsumption-repro/isolated-{none,magic}/`). Rules out the
cross-fragment-contamination explanation from the original escalation — the
7-node graph and its rules alone are sufficient.

## Step 2: minimize further

Reduced to 4 nodes, 4 edges, 2 seed facts, one recursive rule, and (eventually)
one subsumption rule:

```
.decl A(from:number, to:number, z:number) btree_delete
.output A
.decl graph(from:number, to:number)
graph(1, 2).  graph(1, 3).  graph(2, 4).  graph(3, 4).
A(1, 2, 1).  A(1, 3, 1).
A(from, to, c+1) :- A(_, from, c), graph(from, to).
A(from1, to, c) <= A(from2, to, c) :- from1 > from2.
```

(`tests/programs/subsumption_minimal2.dl` — the first subsumption rule from the
original file, `A(from,to,c1)<=A(_,to,c2):-c1>c2.`, was tried and found
unnecessary to reproduce the bug; dropped.)

Both derivation paths (`1→2→4` and `1→3→4`) reach `(to=4, cost=2)` with different
`from` (2 and 3); the subsumption rule should keep only `from=2` (lower). Without
`--magic-transform=*`: correct, `A` has 3 rows, `(3,4,2)` is gone. With it: 4 rows,
`(3,4,2)` survives — the "loser" tuple that should be subsumption-deleted is not
deleted. (`measurements/subsumption-repro/minimal2-{none,magic}/`.)

## Step 3: already reported and already fixed upstream

Two web searches found this is a known, already-fixed bug:

- [Issue #2322](https://github.com/souffle-lang/souffle/issues/2322) — same shape:
  a subsumption rule `eozn(E1,D) <= eozn(E2,D) :- E1>E2.` fails to eliminate
  dominated facts under magic-set transformation. Labeled "bug - identified."
- [Issue #2323](https://github.com/souffle-lang/souffle/issues/2323) — same shape,
  a relation ends up with an extra row a correct subsumption pass couldn't
  produce. Closed.
- Both fixed by [PR #2567](https://github.com/souffle-lang/souffle/pull/2567):
  "Strongly ignore relations with subsumptive clause and fix copy of
  `SubsumptiveClause` during magic-set transformations." **Merged into `master`
  2025-12-07**, commit `7bb8e64`.

**Our installed Soufflé is 2.5** (`docs/DECISIONS.md`, released 2025-03-25,
confirmed via `souffle --version`) — **more than 8 months before the fix landed.**
Not yet in a numbered release as of this check.

## Disposition

**No new issue filed** — already reported, already fixed, just not in the release
this project uses. `semantic/subsumption_multiple_rules` was already excluded from
the correctness corpus (out-of-grammar) and never enters the measurement corpus
(OpenRuleBench, §2.2) either. No action needed against `tests/corpus/
PREREGISTERED.txt` — it stays as the historical record per the corpus ruling §2.1;
this finding is context for that record, not a reason to touch it.

If a future session upgrades the installed Soufflé past `7bb8e64`, re-running
`measurements/subsumption-repro/minimal2-magic/` against the new version and
confirming the sorted `A.csv` sets now match would confirm the fix landed — cheap
verification, not needed now.
