#!/usr/bin/env python3
"""
M1 §3.8 gate: "set equality against Soufflé on every in-grammar corpus
program the front end accepts. Use the existing differential harness.
Report agreed/attempted with the symmetric difference for any
disagreement."

"The front end accepts" = the 20/195 files §3.3 gate one found dlc
parses cleanly (measurements/m1-3.3-gate1-parse-coverage-summary.json).
None of these 20 files has an accompanying .facts file in Soufflé's own
tree (checked directly) -- every one defines its EDB entirely via source
fact clauses, so both engines run against the same empty facts directory,
which is a fair, symmetric comparison (an .input relation with no facts
file is empty for both engines identically).
"""
import json
import sys
import tempfile
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import dlc_interface  # noqa: E402
from differential import run_dlc, run_souffle, compare  # noqa: E402

REPO = Path(__file__).resolve().parents[1]
SOUFFLE_TESTS = Path("/root/souffle-src/tests")
GATE1_SUMMARY = REPO / "measurements" / "m1-3.3-gate1-parse-coverage-summary.json"


def main():
    dlc_interface.build_dlc()
    gate1 = json.loads(GATE1_SUMMARY.read_text())
    parsed_files = [r["file"] for r in gate1["results"] if r["status"] == "parsed"]

    empty_facts = Path(tempfile.mkdtemp()) / "facts"
    empty_facts.mkdir()

    results = []
    for rel in parsed_files:
        dl_path = SOUFFLE_TESTS / rel
        with tempfile.TemporaryDirectory() as workdir_str:
            workdir = Path(workdir_str)
            dlc_result = run_dlc(dl_path, empty_facts)
            souffle_result = run_souffle(dl_path, empty_facts, workdir)
            comparison = compare(dlc_result, souffle_result)
            comparison["file"] = rel
            comparison["dlc_status"] = dlc_result.status
            comparison["souffle_status"] = souffle_result.status
            results.append(comparison)

    attempted = len(results)
    agreed = sum(1 for r in results if r.get("comparable") and r.get("all_match"))
    not_comparable = [r for r in results if not r.get("comparable")]
    disagreements = [r for r in results if r.get("comparable") and not r.get("all_match")]

    summary = {"agreed": agreed, "attempted": attempted, "results": results}
    out_path = REPO / "measurements" / "m1-3.8-naive-eval-summary.json"
    out_path.write_text(json.dumps(summary, indent=2))

    print(json.dumps({"agreed": agreed, "attempted": attempted,
                       "not_comparable": len(not_comparable), "disagreements": len(disagreements)}, indent=2))
    print(f"GATE RESULT: agreed/attempted = {agreed}/{attempted}", file=sys.stderr)
    for r in not_comparable:
        print(f"  NOT COMPARABLE: {r['file']} (dlc_status={r['dlc_status']}, reason={r.get('reason')})", file=sys.stderr)
    for r in disagreements:
        print(f"  DISAGREEMENT: {r['file']}", file=sys.stderr)
        for c in r["relations"]:
            if not c["match"]:
                print(f"    {c['relation']}: dlc_only={c['dlc_only'][:5]} souffle_only={c['souffle_only'][:5]}", file=sys.stderr)

    if disagreements:
        raise SystemExit(1)


if __name__ == "__main__":
    main()
