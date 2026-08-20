#!/usr/bin/env python3
"""
Run harness/parse_profile.py against each probe0 log and record the
exact-integer output under measurements/<id>-extract/, with the same
cmd.txt/stdout.txt/stderr.txt/meta.json provenance shape as probe0.py.

Lane B. Depends only on Soufflé's own JSON profile log format.
"""
import json
import subprocess
import sys
from pathlib import Path

REPO = Path(__file__).resolve().parents[1]
MEASUREMENTS = REPO / "measurements"

RUNS = [
    ("probe0-p1-off", MEASUREMENTS / "probe0-p1-off-run" / "prof_off.log"),
    ("probe0-p1-on", MEASUREMENTS / "probe0-p1-on-run" / "prof_on.log"),
    ("probe0-p2-off", MEASUREMENTS / "probe0-p2-off-run" / "prof_off.log"),
    ("probe0-p2-on", MEASUREMENTS / "probe0-p2-on-run" / "prof_on.log"),
    ("probe0-p3-off", MEASUREMENTS / "probe0-p3-off-run" / "prof_off.log"),
    ("probe0-p3-on", MEASUREMENTS / "probe0-p3-on-run" / "prof_on.log"),
    ("probe0.5-p1prime-off", MEASUREMENTS / "probe0.5-p1prime-off-run" / "prof_off.log"),
    ("probe0.5-p1prime-on", MEASUREMENTS / "probe0.5-p1prime-on-run" / "prof_on.log"),
    ("probe0.5-p4", MEASUREMENTS / "probe0.5-p4-run" / "prof.log"),
]


def main():
    for mid, log_path in RUNS:
        out_id = f"{mid}-extract"
        outdir = MEASUREMENTS / out_id
        outdir.mkdir(parents=True, exist_ok=True)

        cmd = ["python3", "harness/parse_profile.py", str(log_path)]
        (outdir / "cmd.txt").write_text(" ".join(cmd) + "\n")

        proc = subprocess.run(cmd, cwd=str(REPO), capture_output=True, text=True)
        (outdir / "stdout.txt").write_text(proc.stdout)
        (outdir / "stderr.txt").write_text(proc.stderr)
        (outdir / "env.txt").write_text(f"source_log: {log_path}\n")

        meta = {
            "id": out_id,
            "cmd": cmd,
            "source_log": str(log_path),
            "returncode": proc.returncode,
        }
        (outdir / "meta.json").write_text(json.dumps(meta, indent=2) + "\n")

        print(f"{out_id}: returncode={proc.returncode}", file=sys.stderr)
        if proc.returncode != 0:
            raise SystemExit(f"extraction failed: {out_id}")


if __name__ == "__main__":
    main()
