#!/usr/bin/env python3
"""
NIGHT-BATCH-01 T1: provenance and determinism audit.

1. Extract every measurement ID referenced in docs/MEASUREMENTS.md's table
   (the `| ID | ...` rows).
2. For each: does measurements/<id>/ exist, and does it have all five
   provenance files (cmd.txt, stdout.txt, stderr.txt, env.txt, meta.json)?
3. For each measurement directory that DOES exist and has a meta.json with
   a `cmd` list: re-run that exact command (with the recorded cwd), then
   compare git's view of the tree for changes (reproducibility check done
   separately, in bash, via `git status` after replay -- this script only
   does the inventory pass).

Lane B, read-only except for re-running commands whose own side effects
are themselves under git and therefore auditable via `git diff`.
"""
import json
import re
import sys
from pathlib import Path

REPO = Path(__file__).resolve().parents[1]
MEASUREMENTS = REPO / "measurements"
MEASUREMENTS_MD = REPO / "docs" / "MEASUREMENTS.md"

REQUIRED_FILES = ["cmd.txt", "stdout.txt", "stderr.txt", "env.txt", "meta.json"]

ID_RE = re.compile(r"^\|\s*([A-Za-z0-9_./{},.\-]+)\s*\|")


def extract_ids():
    text = MEASUREMENTS_MD.read_text()
    ids = []
    for line in text.splitlines():
        if not line.startswith("|"):
            continue
        m = ID_RE.match(line)
        if not m:
            continue
        cell = m.group(1)
        if cell in ("ID", "---", "Program"):
            continue
        # cells can list multiple related ids separated by " / " or " and "
        # or contain a shell-brace-expansion-like single id; split gently
        # on " / " and " and " but keep single ids intact.
        for piece in re.split(r"\s*/\s*|\s+and\s+", cell):
            piece = piece.strip().strip("`")
            if piece and not piece.startswith("probe0") is False:
                pass
        # simpler: just take the raw cell, split on " / "
        parts = [p.strip().strip("`") for p in cell.split(" / ")]
        ids.extend(p for p in parts if p)
    # de-duplicate, preserve order
    seen = set()
    out = []
    for i in ids:
        if i not in seen:
            seen.add(i)
            out.append(i)
    return out


def main():
    ids = extract_ids()
    rows = []
    for mid in ids:
        d = MEASUREMENTS / mid
        if not d.is_dir():
            rows.append({"id": mid, "status": "missing-provenance", "detail": "no such directory"})
            continue
        missing = [f for f in REQUIRED_FILES if not (d / f).is_file()]
        if missing:
            rows.append({"id": mid, "status": "missing-provenance", "detail": f"missing: {','.join(missing)}"})
            continue
        meta_path = d / "meta.json"
        try:
            meta = json.loads(meta_path.read_text())
        except Exception as e:
            rows.append({"id": mid, "status": "missing-provenance", "detail": f"meta.json unparseable: {e}"})
            continue
        has_cmd = isinstance(meta.get("cmd"), list) and len(meta["cmd"]) > 0
        rows.append({
            "id": mid,
            "status": "provenance-ok" if has_cmd else "missing-provenance",
            "detail": "" if has_cmd else "meta.json has no replayable cmd list",
            "cmd": meta.get("cmd"),
            "cwd": meta.get("cwd"),
        })

    print(json.dumps(rows, indent=2))


if __name__ == "__main__":
    main()
