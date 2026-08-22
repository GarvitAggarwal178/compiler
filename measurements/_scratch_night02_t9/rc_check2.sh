#!/bin/bash
cd /root/compiler
DL=/root/compiler/measurements/_scratch_night02_t5/culprit_cycle_unsafe_cyclic.dl
FX=/root/compiler/fixtures/benchmark-family/culprit_cycle/n20
OUT=/tmp/rc_check2_out
rm -rf "$OUT"; mkdir -p "$OUT"

echo "--- run A: redirected to files ---"
souffle -F "$FX" -D "$OUT" "$DL" > /tmp/rc2A.out 2> /tmp/rc2A.err
echo "RC_A=$?"
ls "$OUT"

echo "--- run B: python subprocess.run capture_output=True ---"
python3 -c "
import subprocess
proc = subprocess.run(['souffle','-F','$FX','-D','$OUT','$DL'], capture_output=True, encoding='utf-8', errors='replace')
print('RC_B=', proc.returncode)
"

echo "--- run C: repeat run A three more times for determinism ---"
for i in 1 2 3; do
  souffle -F "$FX" -D "$OUT" "$DL" > /tmp/rc2C_$i.out 2> /tmp/rc2C_$i.err
  echo "RC_C$i=$?"
done
