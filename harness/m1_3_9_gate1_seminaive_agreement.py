#!/usr/bin/env python3
"""
M1 §3.9 gate one: "same set equality as 3.8, unchanged." Re-runs the
exact same differential comparison §3.8's gate used (the 20/195
in-grammar files dlc's front end accepts, against real Soufflé, empty
facts dir) but through `dlc run-seminaive` instead of `dlc run` --
confirming semi-naive's Δ-rewrite produces the identical answer set
naive evaluation already did, not a new/different comparison.
"""
import json
import subprocess
import sys
import tempfile
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import dlc_interface  # noqa: E402
from differential import run_souffle, compare, _extract_output_names, EngineResult  # noqa: E402

REPO = Path(__file__).resolve().parents[1]
SOUFFLE_TESTS = Path("/root/souffle-src/tests")
GATE1_SUMMARY = REPO / "measurements" / "m1-3.3-gate1-parse-coverage-summary.json"


def run_dlc_seminaive(dl_path: Path, facts_dir: Path) -> EngineResult:
    output_names = _extract_output_names(dl_path)
    with tempfile.TemporaryDirectory() as out_dir_str:
        out_dir = Path(out_dir_str)
        proc = subprocess.run(
            [str(dlc_interface.DLC_BINARY), "run-seminaive", str(dl_path), str(facts_dir), str(out_dir)],
            capture_output=True, encoding="utf-8", errors="replace",
        )
        if proc.returncode != 0 and not proc.stdout.strip():
            return EngineResult(engine="dlc", status="panic", stderr=proc.stderr)
        try:
            doc = json.loads(proc.stdout.strip().splitlines()[-1]) if proc.stdout.strip() else {}
        except (json.JSONDecodeError, IndexError) as e:
            return EngineResult(engine="dlc", status="panic", stderr=f"non-JSON output: {e}")
        status = doc.get("status", "panic")
        if status != "ok":
            return EngineResult(engine="dlc", status=status, stderr=json.dumps(doc)[:1000])
        relations = {}
        for name in output_names:
            fp = out_dir / f"{name}.csv"
            relations[name] = sorted(fp.read_text().splitlines()) if fp.is_file() else []
        return EngineResult(engine="dlc", status="ok", output_relations=relations)


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
            dlc_result = run_dlc_seminaive(dl_path, empty_facts)
            souffle_result = run_souffle(dl_path, empty_facts, workdir)
            comparison = compare(dlc_result, souffle_result)
            comparison["file"] = rel
            comparison["dlc_status"] = dlc_result.status
            results.append(comparison)

    attempted = len(results)
    agreed = sum(1 for r in results if r.get("comparable") and r.get("all_match"))
    disagreements = [r for r in results if r.get("comparable") and not r.get("all_match")]

    summary = {"agreed": agreed, "attempted": attempted, "results": results}
    out_path = REPO / "measurements" / "m1-3.9-gate1-seminaive-agreement-summary.json"
    out_path.write_text(json.dumps(summary, indent=2))
    print(json.dumps({"agreed": agreed, "attempted": attempted}, indent=2))
    print(f"GATE RESULT: agreed/attempted = {agreed}/{attempted} (§3.8 baseline was 11/20)", file=sys.stderr)
    for r in disagreements:
        print(f"  DISAGREEMENT: {r['file']}", file=sys.stderr)
    if disagreements:
        raise SystemExit(1)
    if agreed != 11 or attempted != 20:
        print(f"NOTE: differs from §3.8's baseline (11/20) -- investigate before trusting", file=sys.stderr)


if __name__ == "__main__":
    main()
