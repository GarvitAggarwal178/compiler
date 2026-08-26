#!/usr/bin/env python3
"""
M1 §4 item 3: "Extend the differential harness to run dlc against
Soufflé across the full benchmark family at every pre-registered scale
point."

Runs `dlc run` (naive, §3.8) against every pre-registered scale point in
tests/corpus/BENCHMARK_FAMILY/SCALE_POINTS.json, ascending per shape (so
a cap truncates the top of a shape's curve, not the middle -- this
project's established convention since NIGHT-BATCH-02 T4). dlc is a
tree-walking interpreter with no query planning or codegen (§3.8's own
framing: "correctness before speed") -- some of these scale points
(reachability_complement/transitive_closure_bound/ancestor_nonancestor
at n=8000 measured up to 64M derived tuples for T_none, NIGHT-BATCH-02
T4) are not expected to finish inside any reasonable cap, and a DNF is
recorded, never retried with a higher one, exactly as this project has
always handled a cap firing.

Per-point timeout: 120s (chosen for this task, not previously
pre-registered -- reported plainly as this script's own choice, not
implied to be inherited from elsewhere).
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
FAMILY_DIR = REPO / "tests" / "corpus" / "BENCHMARK_FAMILY"
SCALE_POINTS = json.loads((FAMILY_DIR / "SCALE_POINTS.json").read_text())
TIMEOUT_S = 120

SHAPE_POINTS = {
    "same_generation_negation": [
        (f"d{pt['depth']}_b{pt['branching']}", REPO / "fixtures" / "benchmark-family" / "same_generation_negation" / f"d{pt['depth']}_b{pt['branching']}")
        for pt in SCALE_POINTS["same_generation_negation"]["points"]
    ],
    "transitive_closure_bound": [
        (f"n{pt['n']}", REPO / "fixtures" / "benchmark-family" / "transitive_closure_bound" / f"n{pt['n']}")
        for pt in SCALE_POINTS["transitive_closure_bound"]["points"]
    ],
    "ancestor_nonancestor": [
        (f"n{pt['n']}", REPO / "fixtures" / "benchmark-family" / "ancestor_nonancestor" / f"n{pt['n']}")
        for pt in SCALE_POINTS["ancestor_nonancestor"]["points"]
    ],
    "reachability_complement": [
        (f"n{n}", REPO / "fixtures" / f"p2-scale-{n}")
        for n in SCALE_POINTS["reachability_complement"]["points_already_run"]
    ],
    "culprit_cycle": [
        (f"n{pt['n']}", REPO / "fixtures" / "benchmark-family" / "culprit_cycle" / f"n{pt['n']}")
        for pt in SCALE_POINTS["culprit_cycle"]["points"]
    ],
}


def run_dlc_timed(dl_path: Path, facts_dir: Path) -> EngineResult:
    output_names = _extract_output_names(dl_path)
    with tempfile.TemporaryDirectory() as out_dir_str:
        out_dir = Path(out_dir_str)
        try:
            proc = subprocess.run(
                [str(dlc_interface.DLC_BINARY), "run", str(dl_path), str(facts_dir), str(out_dir)],
                capture_output=True, encoding="utf-8", errors="replace", timeout=TIMEOUT_S,
            )
        except subprocess.TimeoutExpired:
            return EngineResult(engine="dlc", status=f"DNF:timeout-{TIMEOUT_S}s")
        if proc.returncode != 0 and not proc.stdout.strip():
            return EngineResult(engine="dlc", status="panic", stderr=proc.stderr[:500])
        try:
            doc = json.loads(proc.stdout.strip().splitlines()[-1]) if proc.stdout.strip() else {}
        except (json.JSONDecodeError, IndexError) as e:
            return EngineResult(engine="dlc", status="panic", stderr=f"non-JSON output: {e}")
        status = doc.get("status", "panic")
        if status != "ok":
            return EngineResult(engine="dlc", status=status, stderr=json.dumps(doc)[:500])
        relations = {}
        for name in output_names:
            fp = out_dir / f"{name}.csv"
            relations[name] = sorted(fp.read_text().splitlines()) if fp.is_file() else []
        return EngineResult(engine="dlc", status="ok", output_relations=relations)


def main():
    dlc_interface.build_dlc()
    results = {}
    for shape, points in SHAPE_POINTS.items():
        dl_path = FAMILY_DIR / f"{shape}.dl"
        shape_results = []
        for tag, facts_dir in points:
            print(f"=== {shape} {tag} ===", file=sys.stderr)
            with tempfile.TemporaryDirectory() as workdir_str:
                workdir = Path(workdir_str)
                dlc_result = run_dlc_timed(dl_path, facts_dir)
                if dlc_result.status.startswith("DNF"):
                    shape_results.append({"tag": tag, "status": dlc_result.status})
                    print(f"    {dlc_result.status} -- stopping ascent for this shape here", file=sys.stderr)
                    break
                souffle_result = run_souffle(dl_path, facts_dir, workdir)
                comparison = compare(dlc_result, souffle_result)
                comparison["tag"] = tag
                comparison["dlc_status"] = dlc_result.status
                comparison["souffle_status"] = souffle_result.status
                shape_results.append(comparison)
                match = comparison.get("all_match") if comparison.get("comparable") else None
                print(f"    dlc={dlc_result.status} souffle={souffle_result.status} match={match}", file=sys.stderr)
        results[shape] = shape_results

    out_path = REPO / "measurements" / "m1-4.3-full-family-differential-summary.json"
    out_path.write_text(json.dumps(results, indent=2))

    print("\n=== summary, per shape (never aggregated) ===")
    any_disagreement = False
    for shape, shape_results in results.items():
        comparable = [r for r in shape_results if r.get("comparable")]
        matched = [r for r in comparable if r.get("all_match")]
        dnf = [r for r in shape_results if str(r.get("status", "")).startswith("DNF")]
        print(f"{shape}: {len(matched)}/{len(comparable)} matched, {len(dnf)} DNF, "
              f"{len(shape_results) - len(comparable) - len(dnf)} other")
        for r in comparable:
            if not r.get("all_match"):
                any_disagreement = True
                print(f"  DISAGREEMENT at {r['tag']}: {r['relations']}")

    if any_disagreement:
        raise SystemExit(1)


if __name__ == "__main__":
    main()
