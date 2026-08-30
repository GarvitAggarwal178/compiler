#!/usr/bin/env python3
"""Phase 6: fix old-path citations inside src/**/DESIGN.md files.

These are documentation comments describing Lane A design, not Lane A
logic -- only the path strings are rewritten, nothing else in the file.
Uses docs/rename-map.csv for docs/reports/*.md -> experiments/*.md, plus
a fixed table for the append-only/blueprint/build-plan moves.
"""
import csv
import re
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parent.parent
MAP_PATH = REPO_ROOT / "docs" / "rename-map.csv"

FIXED = {
    "docs/dlc-blueprint.md": "docs/02-design.md",
    "docs/DECISIONS.md": "record/DECISIONS.md",
    "docs/MEASUREMENTS.md": "record/MEASUREMENTS.md",
    "docs/OPEN_QUESTIONS.md": "record/OPEN_QUESTIONS.md",
    "docs/SESSION_LOG.md": "record/SESSION_LOG.md",
    "docs/ESCALATIONS.md": "record/ESCALATIONS.md",
    "docs/M1-BUILD.md": "specs/02-m1-build.md",
    "docs/reports/explain-samples/": "experiments/explain-samples/",
}

FILES = [
    "src/cmd/dlc/DESIGN.md",
    "src/transform/guard/DESIGN.md",
    "src/transform/magicset/DESIGN.md",
    "src/eval/DESIGN.md",
    "src/sema/DESIGN.md",
    "src/codegen/DESIGN.md",
    "src/transform/DESIGN.md",
    "src/ir/DESIGN.md",
    "src/parser/DESIGN.md",
    "src/lexer/DESIGN.md",
    "src/token/DESIGN.md",
]


def build_report_map():
    mapping = dict(FIXED)
    with open(MAP_PATH, newline="", encoding="utf-8") as f:
        for row in csv.DictReader(f):
            if row["old_path"].startswith("docs/reports/") and row["kind"] not in (
                "explain_sample",
            ):
                mapping[row["old_path"]] = row["new_path"]
    return mapping


def main():
    mapping = build_report_map()
    # Longest keys first so a directory prefix doesn't shadow a longer,
    # more specific match.
    keys = sorted(mapping, key=len, reverse=True)
    total_subs = 0
    for rel in FILES:
        p = REPO_ROOT / rel
        text = p.read_text(encoding="utf-8")
        original = text
        for old in keys:
            if old in text:
                n = text.count(old)
                text = text.replace(old, mapping[old])
                total_subs += n
                print(f"{rel}: {old} -> {mapping[old]} ({n}x)")
        if text != original:
            p.write_text(text, encoding="utf-8")
    print(f"\n{total_subs} substitution(s) total")


if __name__ == "__main__":
    main()
