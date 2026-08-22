#!/bin/bash
cd /root/compiler
for f in precedence_add_mul precedence_sub_mul precedence_add_div precedence_add_mod precedence_sub_mod precedence_mul_div precedence_mul_mod precedence_div_mod precedence_left_assoc_add precedence_left_assoc_div precedence_unary_mul precedence_unary_sub precedence_paren_override unary_double_minus unary_paren_double_neg; do
  val=$(cat "measurements/night02-t2-hostile-$f-run/p.csv" 2>/dev/null | tr '\n' ',' )
  echo "$f: $val"
done
echo "--- unary_sub_minus (r relation, needs q.facts data to produce rows -- expect empty since q.facts is empty) ---"
cat measurements/night02-t2-hostile-unary_sub_minus-run/p.csv 2>/dev/null
echo "(end)"
