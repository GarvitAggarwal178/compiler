#!/bin/bash
set -e
cd /root/compiler
mkdir -p /tmp/t5smoke

run_one() {
  dl="$1"; fx="$2"; rel="$3"
  echo "=== $dl ==="
  rm -rf "/tmp/t5smoke/$dl"; mkdir -p "/tmp/t5smoke/$dl"
  souffle -F "$fx" -D "/tmp/t5smoke/$dl" "tests/corpus/BENCHMARK_FAMILY/guarded/$dl" \
    > "/tmp/t5smoke/$dl/stdout.txt" 2> "/tmp/t5smoke/$dl/stderr.txt"
  rc=$?
  echo "rc=$rc"
  cat "/tmp/t5smoke/$dl/stderr.txt"
  wc -l "/tmp/t5smoke/$dl/$rel.csv" 2>&1 || true
}

run_one reachability_complement_guarded.dl fixtures/p2-scale-250 q_unreach
run_one ancestor_nonancestor_guarded.dl fixtures/benchmark-family/ancestor_nonancestor/n500 q_nonancestor
run_one same_generation_negation_guarded.dl fixtures/benchmark-family/same_generation_negation/d4_b4 q_notsg
run_one culprit_cycle_guarded.dl fixtures/benchmark-family/culprit_cycle/n20 out
