#!/bin/bash
cd /root/compiler
while read -r rel; do
  [ -z "$rel" ] && continue
  case "$rel" in \#*) continue;; esac
  f="/root/souffle-src/tests/$rel"
  [ -f "$f" ] || continue
  # crude: a line with a head atom containing a bare "_" before ":-"
  awk '
    /:-/ { split($0, a, ":-"); head=a[1] }
    !/:-/ { head=$0 }
    head ~ /\(.*_[ ,)].*\)/ { print FILENAME": "$0 }
  ' "$f"
done < tests/corpus/IN_GRAMMAR.txt | grep -v '^\s*//' | head -20
