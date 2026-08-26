#!/bin/bash
cd /root/compiler
for f in same_generation_negation transitive_closure_bound ancestor_nonancestor reachability_complement culprit_cycle; do
  echo "=== $f ==="
  ./bin/dlc check "tests/corpus/BENCHMARK_FAMILY/$f.dl"
done
