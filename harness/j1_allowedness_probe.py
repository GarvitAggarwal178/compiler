#!/usr/bin/env python3
"""
J1: allowedness probe. Runs each tests/programs/allowedness_probe_{a..g}.dl
through the installed Soufflé 2.5, records accept/reject and the exact
diagnostic text. No definition of allowedness is proposed here -- that is
the human's decision (CLAUDE.md's non-negotiable, reinforced by this
session's own instructions). This script only records observed behaviour.

Lane B. Reuses probe0.run_cmd for standard provenance.
"""
import json
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import probe0  # noqa: E402

REPO = probe0.REPO
PROGRAMS = probe0.PROGRAMS
MEASUREMENTS = probe0.MEASUREMENTS

CASES = "abcdefg"


def main():
    facts_dir = MEASUREMENTS / "_scratch_j1" / "facts"
    facts_dir.mkdir(parents=True, exist_ok=True)
    (facts_dir / "q.facts").write_text("")

    results = []
    for letter in CASES:
        dl_path = PROGRAMS / f"allowedness_probe_{letter}.dl"
        mid = f"j1-allowedness-probe-{letter}"
        workdir = MEASUREMENTS / f"{mid}-run"
        cmd = ["souffle", "-F", str(facts_dir), "-D", str(workdir), str(dl_path)]
        workdir.mkdir(parents=True, exist_ok=True)
        proc = probe0.run_cmd(mid, cmd, workdir)
        outcome = "accept" if proc.returncode == 0 else "reject"
        results.append({
            "case": letter,
            "outcome": outcome,
            "returncode": proc.returncode,
            "stderr": proc.stderr.strip(),
            "stdout": proc.stdout.strip(),
        })
        print(f"{letter}: {outcome} (rc={proc.returncode})", file=sys.stderr)

    out_path = MEASUREMENTS / "j1-allowedness-probe-summary.json"
    out_path.write_text(json.dumps(results, indent=2))
    print(json.dumps(results, indent=2))


if __name__ == "__main__":
    main()
