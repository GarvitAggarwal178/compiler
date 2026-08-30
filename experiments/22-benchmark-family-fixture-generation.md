# NIGHT-BATCH-02 T3 — benchmark-family fixture generation

Date: 2026-08-23. Fixture materialization only — no Soufflé invocation, no
`tests/corpus/BENCHMARK_FAMILY/` or `SCALE_POINTS.json` edits (both
pre-registered, night-02 prohibition 0.2). Generator:
`harness/run_benchmark_family.py`'s `generate_fixtures_only()` (already
existed, unauthorized-to-run half now run for the first time), driven by
`harness/night02_t3_fixtures.py`. Full data:
`measurements/night02-t3-fixtures-summary.json`.

`reachability_complement` is excluded per `SCALE_POINTS.json`'s own note: it
reuses NIGHT-BATCH-01 T6's already-generated, already-verified fixtures
(`fixtures/p2-scale-{250,500,1000,2000,4000,8000}/`, `docs/reports/
night01-T6-scaling.md`) — no new fixtures needed or generated for that shape.

## Idempotency

Ran `generate_fixtures_only()` twice in the same process, hashed every
`.facts` file (SHA-256) after each run. **38/38 files byte-identical across
both runs.** No abort condition fired.

## Fixtures generated

38 `.facts` files, 1.1 MB total, under `fixtures/benchmark-family/<shape>/<tag>/`.

### `same_generation_negation` (tree, root = node 0)

| Scale point | seed | `n_persons` | `n_edges` (parent) | reachable from 0 |
|---|---|---|---|---|
| depth=4, branching=4 | 20260822104 | 226 | 225 | 226 |
| depth=5, branching=4 | 20260822105 | 840 | 839 | 840 |
| depth=6, branching=4 | 20260822106 | 4,888 | 4,887 | 4,888 |
| depth=7, branching=4 | 20260822107 | 14,567 | 14,566 | 14,567 |

"Reachable from 0" walks `par -> child` (the reverse of the stored
`(child, par)` tuples), since node 0 is the root and has no outgoing
`(child, par)` edge of its own. It equals `n_persons` at every point, as
expected for a tree — this confirms connectivity (no orphaned subtree), not
a size-controlled partition the way `core_size` is for the two graph shapes
below. **`sg`/`notsg` cardinality at node 0 is not a fixture-structural
property** — it depends on the recursive same-generation predicate itself,
which requires running Soufflé. That is T4's job, not this one.

### `transitive_closure_bound` and `ancestor_nonancestor` (core/rest graph)

Both use `gen_core_rest_graph` with `core_size=50` (the P1/Phase-0.5
convention: verify the reachable-set size before trusting a ratio built on
it).

| n | seed (`transitive_closure_bound`) | seed (`ancestor_nonancestor`) | `n_edges` | reachable from 1 |
|---|---|---|---|---|
| 500 | 20260822700 | 20260822800 | 1,000 | 50 |
| 1000 | 20260823200 | 20260823300 | 2,000 | 50 |
| 2000 | 20260824200 | 20260824300 | 4,000 | 50 |
| 4000 | 20260826200 | 20260826300 | 8,000 | 50 |
| 8000 | 20260830200 | 20260830300 | 16,000 | 50 |

`n_edges` is identical across both shapes at a given `n` because both use
`gen_core_rest_graph` with the same `target_edges` per point (only the seed
and the relation name differ). Seed = `seed_base + n` (`transitive_closure_bound`
seed_base 20260822200, `ancestor_nonancestor` seed_base 20260822300), per
`SCALE_POINTS.json`.

**Reachable-from-1 is exactly 50 at every scale point, both shapes.** This is
`gen_core_rest_graph`'s designed invariant (`core_size=50`, matching P1's
own convention) — confirmed, not assumed, before any ratio gets built on top
of it in T4/T5.

### `culprit_cycle` (constructed, P5 shape)

| n | seed | `n_base_edges` | `n_e_edges` | `n_blocked` | reachable from 1 via `e` only |
|---|---|---|---|---|---|
| 20 | 20260822420 | 30 | 31 | 4 | 11 |
| 50 | 20260822450 | 75 | 76 | 11 | 27 |
| 100 | 20260822500 | 150 | 150 | 12 | 40 |
| 200 | 20260822600 | 300 | 300 | 33 | 143 |
| 500 | 20260822900 | 750 | 750 | 118 | 289 |

`e_edges` count is `target_e` or `target_e + 1` — `gen_culprit_cycle_facts`
force-adds one `(1, ·)` edge if none was already generated randomly, so `p`
has at least one base-case fact from node 1.

"Reachable from 1 via `e` only" is a **lower bound**, not `p`'s true
reachable set: `p(x,y) :- p(x,z), !s(z), q(z,y)` extends reachability
recursively through `q`/`base`, gated by `!s(z)`. The true reachable set from
node 1 requires running Soufflé — T4's job.

## What did not work

Nothing aborted. One bug caught before trusting the numbers: the first draft
computed "reachable from 0" for `same_generation_negation` by walking the
stored `(child, par)` tuples forward from node 0, which is the *root* and so
has no outgoing edge in that direction — this produced a nonsense `1` for
every scale point (root reaches only itself). Fixed by reversing the walk to
`par -> child` before the numbers above were trusted; re-run confirmed
`reachable_from_0 == n_persons` at every point, which is the expected
invariant for a tree. Caught by eyeballing an implausible number before
writing it into this report, not by a pre-existing test — worth remembering
for T4/T5, which reuse this pattern.

## Provenance

`measurements/night02-t3-fixtures-summary.json` (full hash table, per-shape
structural properties, idempotency verdict). Runner:
`harness/night02_t3_fixtures.py`. Well under the 60-minute cap.
