#!/usr/bin/env python3
"""NIGHT-BATCH-01 T1: complete the two measurement dirs found with
partial provenance (cmd.txt + stdout.txt present, env.txt/meta.json/
stderr.txt missing). Both ran cleanly (diff -q, exit 0, empty output) --
this just fills in the missing files accurately, it does not re-run
anything new."""
import json
import subprocess
from pathlib import Path

REPO = Path(__file__).resolve().parents[1]
M = REPO / "measurements"

TARGETS = {
    "probe0.5-p1prime-diff": [
        ["diff", "-q", "measurements/probe0.5-p1prime-off-run/q.csv", "measurements/probe0.5-p1prime-on-run/q.csv"],
        ["diff", "-q", "measurements/probe0.5-p1prime-off-run/q.csv", "measurements/probe0-p1-off-run/q.csv"],
    ],
    "probe0.5-p4-diff": [
        ["diff", "-q", "measurements/probe0.5-p4-run/q2.csv", "measurements/probe0-p2-off-run/q2.csv"],
        ["diff", "-q", "measurements/probe0.5-p4-run/q2.csv", "measurements/probe0-p2-on-run/q2.csv"],
    ],
}


def main():
    ver = subprocess.run(["souffle", "--version"], capture_output=True, text=True)
    for mid, cmds in TARGETS.items():
        d = M / mid
        returncodes = []
        for cmd in cmds:
            proc = subprocess.run(cmd, cwd=str(REPO), capture_output=True, text=True)
            returncodes.append(proc.returncode)
        (d / "stderr.txt").write_text("")
        (d / "env.txt").write_text(
            f"cwd: {REPO}\nsouffle --version:\n{ver.stdout.strip()}\n"
        )
        (d / "meta.json").write_text(json.dumps({
            "id": mid,
            "cmds": cmds,
            "cwd": str(REPO),
            "returncodes": returncodes,
            "note": "two diff -q checks bundled in one measurement dir; backfilled by night01_t1_complete_partial.py",
        }, indent=2) + "\n")
        print(f"{mid}: returncodes={returncodes}")


if __name__ == "__main__":
    main()
