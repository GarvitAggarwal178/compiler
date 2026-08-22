#!/bin/bash
cd /root/compiler
DL=/root/compiler/measurements/night02-t9-stratification-stratification_self_negative_cycle-run/prog.dl
FX=/root/compiler/measurements/_scratch_night02_t9/facts
OUT=/root/compiler/measurements/night02-t9-stratification-stratification_self_negative_cycle-run

echo "--- run A: redirected to files (like probe0.run_cmd does NOT do -- it uses capture_output) ---"
souffle -F "$FX" -D "$OUT" "$DL" > /tmp/rcA.out 2> /tmp/rcA.err
echo "RC_A=$?"

echo "--- run B: python subprocess.run with capture_output=True (mirrors probe0.run_cmd exactly) ---"
python3 -c "
import subprocess
proc = subprocess.run(['souffle','-F','$FX','-D','$OUT','$DL'], capture_output=True, encoding='utf-8', errors='replace')
print('RC_B=', proc.returncode)
"

echo "--- run C: python subprocess.run with capture_output=False, inheriting stdio ---"
python3 -c "
import subprocess
proc = subprocess.run(['souffle','-F','$FX','-D','$OUT','$DL'])
print('RC_C=', proc.returncode)
"
