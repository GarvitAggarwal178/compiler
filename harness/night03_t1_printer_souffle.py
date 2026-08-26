#!/usr/bin/env python3
"""
NIGHT-BATCH-03 T1: printer output must be Soufflé-parseable.

For each of the 20 dlc-gate1-parsed files, plus the 5 BENCHMARK_FAMILY
shapes, plus tests/programs/p4prime.dl:
  1. dlc roundtrip <file> (parse -> print -> reparse -> ast.Equal),
     capture the .printed text.
  2. Write printed text to a temp .dl file.
  3. Run Soufflé on the printed file with the shape's fixture.
  4. Run Soufflé on the original file with the same fixture.
  5. Compare every .output relation by sorted set equality.

Gate: souffle-accepted/attempted, answer-identical/attempted.
"""
import json
import re
import subprocess
import sys
import tempfile
from pathlib import Path

sys.path.insert(0, "/root/compiler/harness")
import dlc_interface  # noqa: E402

REPO = Path("/root/compiler")
SOUFFLE_TESTS = Path("/root/souffle-src/tests")
FAMILY_DIR = REPO / "tests" / "corpus" / "BENCHMARK_FAMILY"
FIXTURES_ROOT = REPO / "fixtures" / "benchmark-family"
OUT_DIR = REPO / "measurements" / "night03-t1"


def output_names(text):
    return sorted(set(re.findall(r"^\s*\.output\s+(\w+)\s*(?:\(\s*\))?\s*$", text, re.MULTILINE)))


def run_souffle(dl_path: Path, facts_dir: Path, workdir: Path):
    workdir.mkdir(parents=True, exist_ok=True)
    cmd = ["souffle", "-F", str(facts_dir.resolve()), "-D", str(workdir.resolve()), str(dl_path.resolve())]
    try:
        proc = subprocess.run(cmd, cwd=str(workdir), capture_output=True, text=True, timeout=120)
    except subprocess.TimeoutExpired:
        return {"status": "timeout", "stderr": ""}
    return {"status": "ok" if proc.returncode == 0 else f"error:{proc.returncode}",
            "stderr": proc.stderr[:2000], "returncode": proc.returncode}


def collect_relations(workdir: Path, names):
    out = {}
    for name in names:
        fp = workdir / f"{name}.csv"
        out[name] = sorted(fp.read_text().splitlines()) if fp.is_file() else None
    return out


def main():
    dlc_interface.build_dlc()
    gate1 = json.loads((REPO / "measurements" / "m1-3.3-gate1-parse-coverage-summary.json").read_text())
    parsed_files = [r["file"] for r in gate1["results"] if r["status"] == "parsed"]

    cases = []
    for rel in parsed_files:
        cases.append({"id": rel, "src": SOUFFLE_TESTS / rel, "facts_dir": (SOUFFLE_TESTS / rel).parent})

    family_shapes = [
        ("same_generation_negation", FAMILY_DIR / "same_generation_negation.dl", FIXTURES_ROOT / "same_generation_negation" / "d4_b4"),
        ("transitive_closure_bound", FAMILY_DIR / "transitive_closure_bound.dl", FIXTURES_ROOT / "transitive_closure_bound" / "n500"),
        ("ancestor_nonancestor", FAMILY_DIR / "ancestor_nonancestor.dl", FIXTURES_ROOT / "ancestor_nonancestor" / "n500"),
        ("culprit_cycle", FAMILY_DIR / "culprit_cycle.dl", FIXTURES_ROOT / "culprit_cycle" / "n20"),
        ("reachability_complement", FAMILY_DIR / "reachability_complement.dl", REPO / "fixtures" / "p2-scale-250"),
    ]
    for name, src, fdir in family_shapes:
        cases.append({"id": f"BENCHMARK_FAMILY/{name}", "src": src, "facts_dir": fdir})

    cases.append({"id": "tests/programs/p4prime.dl", "src": REPO / "tests" / "programs" / "p4prime.dl",
                  "facts_dir": REPO / "fixtures" / "p2-scale-250"})

    results = []
    for case in cases:
        cid, src, facts_dir = case["id"], case["src"], case["facts_dir"]
        text = src.read_text(errors="replace")
        rt = dlc_interface.run_dlc_roundtrip(text)
        row = {"id": cid, "roundtrip_status": rt.status}
        if rt.status != "match":
            row["souffle_printed"] = "skip:roundtrip_not_match"
            row["answer_identical"] = None
            results.append(row)
            continue

        safe = re.sub(r"[^A-Za-z0-9_.]", "_", cid)
        with tempfile.TemporaryDirectory() as td:
            printed_path = Path(td) / f"{safe}.printed.dl"
            printed_path.write_text(rt.printed)
            names = output_names(rt.printed) or output_names(text)

            wd_printed = Path(td) / "out_printed"
            wd_orig = Path(td) / "out_orig"
            r_printed = run_souffle(printed_path, facts_dir, wd_printed)
            r_orig = run_souffle(src, facts_dir, wd_orig)

            row["souffle_printed"] = r_printed["status"]
            row["souffle_orig"] = r_orig["status"]
            row["output_relations"] = names
            if r_printed["status"] != "ok":
                row["souffle_printed_stderr"] = r_printed["stderr"]
            if r_orig["status"] != "ok":
                row["souffle_orig_stderr"] = r_orig["stderr"]

            if r_printed["status"] == "ok" and r_orig["status"] == "ok":
                rel_printed = collect_relations(wd_printed, names)
                rel_orig = collect_relations(wd_orig, names)
                row["per_relation_match"] = {n: (rel_printed[n] == rel_orig[n]) for n in names}
                row["answer_identical"] = all(row["per_relation_match"].values()) if names else True
                if not row["answer_identical"]:
                    row["diff"] = {
                        n: {"printed_only": sorted(set(rel_printed[n] or []) - set(rel_orig[n] or [])),
                            "orig_only": sorted(set(rel_orig[n] or []) - set(rel_printed[n] or []))}
                        for n in names if rel_printed[n] != rel_orig[n]
                    }
            else:
                row["answer_identical"] = None
        results.append(row)

    attempted = len(results)
    souffle_accepted = sum(1 for r in results if r.get("souffle_printed") == "ok")
    answer_identical = sum(1 for r in results if r.get("answer_identical") is True)
    comparable = sum(1 for r in results if r.get("answer_identical") is not None)

    summary = {
        "attempted": attempted,
        "souffle_accepted": souffle_accepted,
        "souffle_accepted_of_attempted": f"{souffle_accepted}/{attempted}",
        "answer_identical": answer_identical,
        "comparable": comparable,
        "answer_identical_of_comparable": f"{answer_identical}/{comparable}",
        "results": results,
    }
    OUT_DIR.mkdir(parents=True, exist_ok=True)
    (OUT_DIR / "summary.json").write_text(json.dumps(summary, indent=2))
    print(json.dumps({k: v for k, v in summary.items() if k != "results"}, indent=2))
    for r in results:
        if r.get("souffle_printed") != "ok" or r.get("answer_identical") is False:
            print("NOTABLE:", json.dumps(r, indent=2)[:2000])


if __name__ == "__main__":
    main()
