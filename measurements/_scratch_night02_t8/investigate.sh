#!/bin/bash
cd /root/compiler
echo "=== files with 0-arity .decl ==="
while read -r rel; do
  [ -z "$rel" ] && continue
  case "$rel" in \#*) continue;; esac
  f="/root/souffle-src/tests/$rel"
  [ -f "$f" ] || continue
  if grep -qE '\.decl\s+\w+\s*\(\s*\)' "$f"; then
    echo "$rel"
    grep -E '\.decl\s+\w+\s*\(\s*\)' "$f" | head -2
  fi
done < tests/corpus/IN_GRAMMAR.txt | head -60

echo
echo "=== PrimitiveAssign 1590-rule file ==="
grep -rl 'PrimitiveAssign' /root/souffle-src/tests/*/*.dl 2>/dev/null | head -3
