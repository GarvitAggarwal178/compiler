# NIGHT-BATCH-03 T9 — fallback cone metric

Date: 2026-08-27. M3's blast-radius integer is the project's only number a
hand transform cannot produce; this is the independent Lane B measurement
of it. Runner: `harness/cone_metric.py` (the reusable function),
`harness/night03_t9_run.py` (the per-shape sweep).

## What did not work

Nothing. Validated against the known-observed `culprit_cycle` pattern on
the first run.

## Method

`cone_size(idb_relations, all_edges, declined_sccs)`: the downward
dependency closure of the declined set, over the **full** dependency
relation (positive and negative edges alike) — a plain BFS outward from
the declined relations, following every edge regardless of polarity, per
M2-M3-BUILD.md §7's own specification. The **decision** of which SCCs to
decline is guard/Lane-A territory (`src/transform/guard/decide.go`, not yet
written); this module only answers "given a declined set, what comes down
with it" — a pure graph query, reusable regardless of how the declined set
is chosen. SCCs computed via Tarjan over the full (not positive-only) edge
set, matching `sema/stratify.go`'s own SCC convention. Dependency graph
built via `harness/night03_t4_culprit_classifier.py`'s already-validated
structural parser (`parse_structure`/`build_graphs`) — same regex-based,
disclosed-approximate extraction, not a second parser.

## Validation

`culprit_cycle.dl`, `{p}` declined:

```json
{"declined_sccs": 1, "declined_relations": ["p"], "cone_relations": ["q", "s"], "cone_size": 2, "cone_fraction": 0.5}
```

**Matches the required validation exactly** — the cone contains `q` and
`s`, the same full-extent relations observed empirically in both the
automatic Soufflé transform (`docs/reports/night02-T7-p5-precheck.md`:
`@poscopy_1.q`=6,899=untransformed `q`; `@neglabel.s`=120=untransformed `s`)
and the hand guard (`docs/reports/night02-T5-guarded.md`: `culprit_cycle_
guarded.dl` leaves `q`/`s` fully unbound). This is not a new finding — it
is confirmation that the independent graph-query tool reproduces a pattern
this project already measured twice by other means.

## The numbers, all 5 shapes

Provenance: `measurements/night03-t9/cone-metric-summary.json`. Declined
set per shape: that shape's own negation-bearing relation (a hypothetical
single-relation decline, chosen because it is the smallest possible
declined set that still exercises the mechanism — not a claim about what
the real guard, once it exists, would actually decline).

| shape | declined | cone relations | cone size | cone fraction (of IDB relations) |
|---|---|---|---|---|
| `same_generation_negation` | `{notsg}` | `{sg}` | 1 | 1/3 |
| `transitive_closure_bound` | — | — | — | n/a — no negation in this shape (`E_recoverable=0`, `docs/reports/night02-T4-baseline.md`) |
| `ancestor_nonancestor` | `{nonancestor}` | `{ancestor}` | 1 | 1/3 |
| `reachability_complement` | `{unreach}` | `{reach}` | 1 | 1/3 |
| `culprit_cycle` | `{p}` | `{q, s}` | 2 | 1/2 |

**The pattern:** for the three partition-style shapes (`sg`/`notsg`,
`ancestor`/`nonancestor`, `reach`/`unreach`), declining the negation-bearing
relation drags down exactly one other relation — its positive counterpart —
uniformly `cone_fraction=1/3`. `culprit_cycle` is structurally different:
`p` has both a positive dependency (`q`) and a negative one (`s`), and `s`
itself further depends on `q` (already counted, no new relation added) —
the cone is larger (2, `cone_fraction=1/2`) because `p`'s own rule mixes
both edge polarities into the same declined relation's dependency set, not
because the graph is deeper.

## What a skeptic attacks first

- The declined set for each shape is a **hypothetical** single-relation
  decline, chosen as the smallest exercising set, not the output of a real
  guard decision (`src/transform/guard/decide.go` does not exist yet). The
  numbers here validate the *measurement*, not a *prediction* of what M3's
  real guard will decline.
- `cone_relations` deliberately excludes the declined relations
  themselves (it reports only what gets dragged *down*, not the declined
  set echoed back) — a caller computing "total relations affected" must
  add `declined_relations` back in; not doing so would undercount by
  exactly the size of the declined set.
- The dependency graph reuses T4's regex-based structural parser, with the
  same disclosed approximation caveats (a functor call nested in an
  argument list could in principle be miscounted as an edge) — no false
  edges were found in this run's 5 small, already-well-understood shapes,
  but this was not independently stress-tested against `sema.
  CheckStratification`'s real precedence graph on these files.

## Verdict

**T9: DONE.** Validated exactly against `culprit_cycle`'s already-observed
behaviour (2 independent prior measurements). All 5 shapes reported — 3 at
`cone_fraction=1/3`, 1 (`culprit_cycle`) at `1/2`, 1
(`transitive_closure_bound`) not applicable (no negation to decline).
