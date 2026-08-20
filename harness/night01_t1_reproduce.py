#!/usr/bin/env python3
"""
NIGHT-BATCH-01 T1: re-run every measurement's recorded command and
compare fresh stdout against the committed stdout.txt, byte-for-byte, in
memory (never overwriting the committed provenance files themselves).
Commands with filesystem side effects (souffle writing into its own -D
dir) are allowed to run for real; `git status`/`git diff` after this
script is the check for those -- if they reproduce identically, nothing
changes; if not, this script has already found it via the stdout/return-
code comparison, or `git diff` will show it directly.

Applies the batch's global caps: timeout 300s, 8GB address-space limit.
"""
import json
import resource
import subprocess
import sys
from pathlib import Path

REPO = Path(__file__).resolve().parents[1]
MEASUREMENTS = REPO / "measurements"

MEM_LIMIT_BYTES = 8 * 1024 * 1024 * 1024


def _limit_mem():
    resource.setrlimit(resource.RLIMIT_AS, (MEM_LIMIT_BYTES, MEM_LIMIT_BYTES))


def main():
    dirs = sorted(p for p in MEASUREMENTS.iterdir() if p.is_dir())
    results = []
    for d in dirs:
        meta_path = d / "meta.json"
        if not meta_path.is_file():
            results.append({"id": d.name, "status": "no-meta"})
            continue
        try:
            meta = json.loads(meta_path.read_text())
        except Exception as e:
            results.append({"id": d.name, "status": "meta-unparseable", "detail": str(e)})
            continue

        cmds = meta.get("cmds") or ([meta["cmd"]] if meta.get("cmd") else [])
        cwd = meta.get("cwd", str(REPO))
        if not cmds:
            results.append({"id": d.name, "status": "no-cmd"})
            continue

        committed_stdout = (d / "stdout.txt").read_text() if (d / "stdout.txt").is_file() else None

        fresh_stdouts = []
        fresh_returncodes = []
        dnf = None
        for cmd in cmds:
            try:
                proc = subprocess.run(
                    cmd, cwd=cwd, capture_output=True, text=True,
                    timeout=300, preexec_fn=_limit_mem,
                )
                fresh_stdouts.append(proc.stdout)
                fresh_returncodes.append(proc.returncode)
            except subprocess.TimeoutExpired:
                dnf = "timeout-300s"
                break
            except MemoryError:
                dnf = "memcap-8gb"
                break

        if dnf:
            results.append({"id": d.name, "status": "DNF", "cap": dnf})
            continue

        fresh_stdout_joined = "".join(fresh_stdouts)
        if committed_stdout is None:
            results.append({"id": d.name, "status": "no-committed-stdout"})
        elif len(cmds) == 1 and fresh_stdout_joined == committed_stdout:
            results.append({"id": d.name, "status": "reproduced"})
        elif len(cmds) > 1:
            # multi-command dirs (bundled diffs): compare each returncode
            # is 0 and stdout empty, matching the original bundled record
            results.append({
                "id": d.name,
                "status": "reproduced" if all(rc == 0 for rc in fresh_returncodes) and fresh_stdout_joined == "" else "mismatch",
                "returncodes": fresh_returncodes,
            })
        else:
            results.append({
                "id": d.name,
                "status": "mismatch",
                "committed_len": len(committed_stdout),
                "fresh_len": len(fresh_stdout_joined),
            })

    print(json.dumps(results, indent=2))
    n_repro = sum(1 for r in results if r["status"] == "reproduced")
    n_total = len(results)
    n_mismatch = sum(1 for r in results if r["status"] == "mismatch")
    print(f"SUMMARY: {n_repro}/{n_total} reproduced, {n_mismatch} mismatch", file=sys.stderr)


if __name__ == "__main__":
    main()
