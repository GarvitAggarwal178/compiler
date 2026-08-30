# Measurement corpus — canonical benchmark family

Pre-registered 2026-08-22 (`experiments/17-corpus-closed-version-risk-opened.md` §3.2, adopting
Phase 0.7 fallback (1)). **Not run.** No `T_none`/`T_souffle`/`T_guard`/
`E_recoverable` number exists for any shape here except `reachability_complement`
(which reuses NIGHT-BATCH-01 T6's already-run sweep,
`experiments/10-guard-scaling-first-measurement.md`).

Replaces the OpenRuleBench attempt
(`experiments/14-openrulebench-unobtainable.md`) as the measurement corpus, per the
2026-08-21 ruling: *the rules are external, the fact generators are ours, and
both facts are disclosed.*

## Shapes

| File | Citation | Negation? |
|---|---|---|
| `same_generation_negation.dl` | Ullman 1988 (SG structure) + OpenRuleBench negation category (motivation; original file unobtainable, reconstructed) | yes |
| `transitive_closure_bound.dl` | Beeri & Ramakrishnan, PODS 1987 + OpenRuleBench `datalog_recursion` | no (positive-fragment baseline) |
| `ancestor_nonancestor.dl` | Ullman 1988 / Abiteboul-Hull-Vianu 1995 ch.15 | yes |
| `reachability_complement.dl` | This project's own P2 shape since Phase 0; lineage disclosed, not presented as newly canonical | yes |
| `culprit_cycle.dl` | **Constructed for clause (a), not from a published source** — disclosed | yes |

Every `.dl` file's header comment carries its own citation in full; this table
is an index, not the source of truth.

## Generators

`harness/fixtures_lib.py`: `gen_core_rest_graph` (controllable reachable-set
size, used by `transitive_closure_bound` and `ancestor_nonancestor`),
`gen_random_graph` (used by `reachability_complement`, already run in T6),
`gen_family_tree` (new, `same_generation_negation`), `gen_culprit_cycle_facts`
(new, `culprit_cycle`). All seeded, all deterministic — validated against every
registered scale point (in-memory, no fixture files written) before this commit.

## Scale points

`SCALE_POINTS.json` — committed before any run, per step 3 of the procedure.
Every point's seed is `seed_base + <size parameter>`, recorded per-shape in the
JSON, not chosen ad hoc at run time.

## Driver

`harness/run_benchmark_family.py` — fixture-generation half only
(`generate_fixtures_only()`), not invoked by this commit. The Soufflé-invocation
half is deliberately unimplemented until a human authorizes running this family
(module docstring explains why: not making it trivially easy to "just try it").

## What running this would answer

The question T2/T3 (NIGHT-BATCH-01, Soufflé `tests/`) and the OpenRuleBench
attempt both failed to answer at scale: `T_souffle/T_guard` and
`E_recoverable/T_souffle` on programs large enough for the ratio to mean
anything, with external provenance for the rule shapes and disclosed provenance
for the data. `reachability_complement`'s Θ(n) result (T6) is the model for what
the other shapes could show — or fail to show, which would also be reportable.
