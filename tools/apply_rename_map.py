#!/usr/bin/env python3
"""Phase 3 mechanics for the repository restructure.

Reads docs/rename-map.csv and executes `git mv` for every row whose `kind`
matches one of the kinds passed on the command line. Creates parent
directories as needed. Skips rows whose kind is a copy (handled separately,
by hand, since git mv does not apply) or the live-CLAUDE.md row (path
unchanged).

Usage: python3 tools/apply_rename_map.py report session_wrapper explain_sample
"""
import csv
import subprocess
import sys
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parent.parent
MAP_PATH = REPO_ROOT / "docs" / "rename-map.csv"

SKIP_KINDS = {"build_plan_copy", "agent_contract_live", "blueprint"}


def main() -> int:
    if len(sys.argv) < 2:
        print(__doc__)
        return 1
    wanted_kinds = set(sys.argv[1:])

    rows = []
    with open(MAP_PATH, newline="", encoding="utf-8") as f:
        for row in csv.DictReader(f):
            rows.append(row)

    moved = 0
    for row in rows:
        kind = row["kind"]
        if kind not in wanted_kinds:
            continue
        if kind in SKIP_KINDS:
            print(f"SKIP (kind={kind}, handled separately): {row['old_path']}")
            continue
        old = REPO_ROOT / row["old_path"]
        new = REPO_ROOT / row["new_path"]
        if not old.exists():
            if new.exists():
                print(f"already done, skipping: {row['old_path']} -> {row['new_path']}")
                continue
            print(f"MISSING SOURCE, aborting: {old}")
            return 1
        new.parent.mkdir(parents=True, exist_ok=True)
        tracked = subprocess.run(
            ["git", "ls-files", "--error-unmatch", str(old.relative_to(REPO_ROOT))],
            cwd=REPO_ROOT, capture_output=True,
        ).returncode == 0
        if tracked:
            subprocess.run(
                ["git", "mv", str(old.relative_to(REPO_ROOT)), str(new.relative_to(REPO_ROOT))],
                cwd=REPO_ROOT,
                check=True,
            )
        else:
            old.rename(new)
            subprocess.run(
                ["git", "add", str(new.relative_to(REPO_ROOT))],
                cwd=REPO_ROOT, check=True,
            )
        moved += 1
        print(f"moved ({'tracked' if tracked else 'UNTRACKED->added'}): {row['old_path']} -> {row['new_path']}")

    print(f"\n{moved} file(s) moved for kinds: {sorted(wanted_kinds)}")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
