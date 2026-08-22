#!/bin/bash
set -e
cd /root/compiler
MASTER=/root/souffle-master/build/src/souffle
mkdir -p measurements/t0-version-risk/p2-master measurements/t0-version-risk/p4prime-master
"$MASTER" -F "$(pwd)/fixtures/p2" -D "$(pwd)/measurements/t0-version-risk/p2-master" -p "$(pwd)/measurements/t0-version-risk/p2-master/prof.log" --magic-transform=* "$(pwd)/tests/programs/p2.dl"
"$MASTER" -F "$(pwd)/fixtures/p2" -D "$(pwd)/measurements/t0-version-risk/p4prime-master" -p "$(pwd)/measurements/t0-version-risk/p4prime-master/prof.log" "$(pwd)/tests/programs/p4prime.dl"
echo "=== p2 master ==="
python3 harness/tuple_report.py measurements/t0-version-risk/p2-master/prof.log
echo "=== p4prime master ==="
python3 harness/tuple_report.py measurements/t0-version-risk/p4prime-master/prof.log
