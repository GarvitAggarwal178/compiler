# Souffle reference transforms

Reference material, not golden targets -- Souffle applies inlining and its
own `@`-prefixed naming that `dlc` will not replicate. A syntactic diff
against these is worthless and none is built. Value: shows what a correct
transform of a known program looks like, for human reading.

Generated via `souffle --show=transformed-ast [--magic-transform=*] <file>`,
confirmed option name (not `--show=transformed-datalog`, which does not
exist in Souffle 2.5 -- checked via `souffle --help` before use).

| file | source | plain rc | magic rc |
|---|---|---|---|
| same_generation_negation | tests/corpus/BENCHMARK_FAMILY/same_generation_negation.dl | 0 | 0 |
| transitive_closure_bound | tests/corpus/BENCHMARK_FAMILY/transitive_closure_bound.dl | 0 | 0 |
| ancestor_nonancestor | tests/corpus/BENCHMARK_FAMILY/ancestor_nonancestor.dl | 0 | 0 |
| culprit_cycle | tests/corpus/BENCHMARK_FAMILY/culprit_cycle.dl | 0 | 0 |
| reachability_complement | tests/corpus/BENCHMARK_FAMILY/reachability_complement.dl | 0 | 0 |
| p1prime | tests/programs/p1prime.dl | 0 | 0 |
| p2 | tests/programs/p2.dl | 0 | 0 |
| p4prime | tests/programs/p4prime.dl | 0 | 0 |