#!/usr/bin/env python3
"""
NIGHT-BATCH-03 T6: re-run gate one and gate two after the .input/.output
optional-parens amendment. Does NOT touch the committed m1-3.3-gate{1,2}
summary files (NIGHT-BATCH-03.md section 0.2 prohibits editing committed
measurements/) -- writes fresh files under measurements/night03-t6/
instead. Re-implements parse_coverage.py/round_trip_scaffold.py's logic
inline against the (now-different) built dlc binary, rather than running
those scripts and letting them overwrite the pre-existing committed files.
"""
import json
import re
import sys
from pathlib import Path

sys.path.insert(0, "/root/compiler/harness")
import dlc_interface  # noqa: E402

REPO = Path("/root/compiler")
SOUFFLE_TESTS = Path("/root/souffle-src/tests")
IN_GRAMMAR = REPO / "tests" / "corpus" / "IN_GRAMMAR.txt"
OUT_DIR = REPO / "measurements" / "night03-t6"

NEGATION_RE = re.compile(r"!\s*[A-Za-z_][A-Za-z0-9_.]*\s*\(")


def load_file_list():
    lines = IN_GRAMMAR.read_text().splitlines()
    return [line.strip() for line in lines if line.strip() and not line.startswith("#")]


def is_negation_bearing(text):
    # Same coarse heuristic as the T4 classifier's literal scan -- good
    # enough to classify "negation appears in this file's source text",
    # not to reconstruct full clause structure.
    return bool(NEGATION_RE.search(text))


def main():
    dlc_interface.build_dlc()
    files = load_file_list()

    gate1_results = []
    gate2_results = []
    negation_bearing_files = set()

    for rel in files:
        dl_path = SOUFFLE_TESTS / rel
        if not dl_path.is_file():
            gate1_results.append({"file": rel, "status": "missing_source"})
            gate2_results.append({"file": rel, "status": "missing_source"})
            continue
        source = dl_path.read_text(errors="replace")
        if is_negation_bearing(source):
            negation_bearing_files.add(rel)

        r1 = dlc_interface.run_dlc_parse(source)
        gate1_results.append({
            "file": rel, "status": r1.status, "decl_count": r1.decl_count,
            "clause_count": r1.clause_count, "error_count": r1.error_count,
        })

        r2 = dlc_interface.run_dlc_roundtrip(source)
        gate2_results.append({"file": rel, "status": r2.status})

    counts1 = {}
    for r in gate1_results:
        counts1[r["status"]] = counts1.get(r["status"], 0) + 1
    counts2 = {}
    for r in gate2_results:
        counts2[r["status"]] = counts2.get(r["status"], 0) + 1

    parsed_files = {r["file"] for r in gate1_results if r["status"] == "parsed"}
    negation_bearing_parsed = negation_bearing_files & parsed_files

    before = json.loads((REPO / "measurements" / "m1-3.3-gate1-parse-coverage-summary.json").read_text())
    before_parsed_files = {r["file"] for r in before["results"] if r["status"] == "parsed"}
    negation_bearing_parsed_before = negation_bearing_files & before_parsed_files

    summary = {
        "total": len(files),
        "total_parsed_before": len(before_parsed_files),
        "total_parsed_after": len(parsed_files),
        "negation_bearing_total_in_corpus": len(negation_bearing_files),
        "negation_bearing_parsed_before": len(negation_bearing_parsed_before),
        "negation_bearing_parsed_after": len(negation_bearing_parsed),
        "negation_bearing_parsed_after_files": sorted(negation_bearing_parsed),
        "gate1_counts": counts1,
        "gate2_counts": counts2,
        "gate1_results": gate1_results,
        "gate2_results": gate2_results,
    }
    OUT_DIR.mkdir(parents=True, exist_ok=True)
    (OUT_DIR / "gates-after-amendment.json").write_text(json.dumps(summary, indent=2))
    print(json.dumps({k: v for k, v in summary.items() if k not in ("gate1_results", "gate2_results")}, indent=2))


if __name__ == "__main__":
    main()
