#!/usr/bin/env python3
"""
NIGHT-BATCH-03 T2: M3 measurement protocol, end-to-end, on the identity
transform.

    source.dl
      -> dlc emit --transformer=<name> -> transformed.dl
      -> souffle -F<fixtures> -D<out> -p prof.log transformed.dl
      -> harness/tuple_report.py -> exact per-relation tuple counts
      -> compare answer relations against souffle(source.dl), sorted set equality

With --transformer=passthrough (the only wired implementation), T(emit) must
equal T(source) exactly on every shape, every scale point, since PassThrough
never changes the program.
"""
import json
import re
import resource
import subprocess
import sys
from pathlib import Path

sys.path.insert(0, "/root/compiler/harness")
import dlc_interface  # noqa: E402
from tuple_report import analyze as tuple_analyze  # noqa: E402

REPO = Path("/root/compiler")
FAMILY_DIR = REPO / "tests" / "corpus" / "BENCHMARK_FAMILY"
SCALE_POINTS = json.loads((FAMILY_DIR / "SCALE_POINTS.json").read_text())
FIXTURES_ROOT = REPO / "fixtures" / "benchmark-family"
MEASUREMENTS = REPO / "measurements" / "night03-t2"

MEM_LIMIT_BYTES = 8 * 1024 * 1024 * 1024
TIMEOUT_S = 300

SHAPES = {
    "same_generation_negation": {
        "dl": "same_generation_negation.dl", "answer": "q_notsg",
        "points": [(f"d{pt['depth']}_b{pt['branching']}",
                    FIXTURES_ROOT / "same_generation_negation" / f"d{pt['depth']}_b{pt['branching']}")
                   for pt in SCALE_POINTS["same_generation_negation"]["points"]],
    },
    "transitive_closure_bound": {
        "dl": "transitive_closure_bound.dl", "answer": "q_tc",
        "points": [(f"n{pt['n']}", FIXTURES_ROOT / "transitive_closure_bound" / f"n{pt['n']}")
                   for pt in SCALE_POINTS["transitive_closure_bound"]["points"]],
    },
    "ancestor_nonancestor": {
        "dl": "ancestor_nonancestor.dl", "answer": "q_nonancestor",
        "points": [(f"n{pt['n']}", FIXTURES_ROOT / "ancestor_nonancestor" / f"n{pt['n']}")
                   for pt in SCALE_POINTS["ancestor_nonancestor"]["points"]],
    },
    "culprit_cycle": {
        "dl": "culprit_cycle.dl", "answer": "out",
        "points": [(f"n{pt['n']}", FIXTURES_ROOT / "culprit_cycle" / f"n{pt['n']}")
                   for pt in SCALE_POINTS["culprit_cycle"]["points"]],
    },
    "reachability_complement": {
        "dl": "reachability_complement.dl", "answer": "q_unreach",
        "points": [(f"n{n}", REPO / "fixtures" / f"p2-scale-{n}")
                   for n in SCALE_POINTS["reachability_complement"]["points_already_run"]],
    },
}


def _limit_mem():
    resource.setrlimit(resource.RLIMIT_AS, (MEM_LIMIT_BYTES, MEM_LIMIT_BYTES))


def run_souffle(dl_path, facts_dir, workdir, log_name="prof.log"):
    workdir.mkdir(parents=True, exist_ok=True)
    cmd = ["souffle", "-F", str(facts_dir), "-D", str(workdir), "-p", log_name, str(dl_path)]
    (workdir / "cmd.txt").write_text(" ".join(cmd) + "\n")
    try:
        proc = subprocess.run(cmd, cwd=str(workdir), capture_output=True, text=True,
                               timeout=TIMEOUT_S, preexec_fn=_limit_mem)
    except subprocess.TimeoutExpired:
        return {"status": f"DNF:timeout-{TIMEOUT_S}s"}
    except MemoryError:
        return {"status": "DNF:memcap-8gb"}
    (workdir / "stdout.txt").write_text(proc.stdout)
    (workdir / "stderr.txt").write_text(proc.stderr)
    if proc.returncode != 0:
        return {"status": f"error:returncode-{proc.returncode}", "stderr": proc.stderr[:800]}
    return {"status": "ok"}


def dlc_emit(dl_path: Path, transformer: str, workdir: Path):
    proc = subprocess.run(
        [str(dlc_interface.DLC_BINARY), "emit", str(dl_path), f"--transformer={transformer}"],
        capture_output=True, encoding="utf-8", errors="replace",
    )
    workdir.mkdir(parents=True, exist_ok=True)
    (workdir / "emit_stdout.json").write_text(proc.stdout)
    try:
        doc = json.loads(proc.stdout.strip().splitlines()[-1]) if proc.stdout.strip() else {}
    except (json.JSONDecodeError, IndexError):
        return None, f"non-JSON emit output; stderr: {proc.stderr[:500]}"
    if doc.get("status") != "ok":
        return None, f"emit status={doc.get('status')}: {json.dumps(doc)[:500]}"
    return doc["printed"], None


def sorted_lines(path: Path):
    return sorted(path.read_text().splitlines()) if path.is_file() else []


def main():
    transformer = "passthrough"
    if len(sys.argv) > 1:
        transformer = sys.argv[1]
    dlc_interface.build_dlc()

    all_rows = []
    for shape, spec in SHAPES.items():
        print(f"===== shape={shape} =====", file=sys.stderr)
        dl_path = FAMILY_DIR / spec["dl"]
        answer_rel = spec["answer"]
        for tag, facts_dir in spec["points"]:
            print(f"  --- {tag} ---", file=sys.stderr)
            row = {"shape": shape, "tag": tag, "facts_dir": str(facts_dir)}
            base = MEASUREMENTS / shape / tag

            printed, err = dlc_emit(dl_path, transformer, base / "emit")
            if printed is None:
                row["status"] = "emit_error"
                row["error"] = err
                all_rows.append(row)
                print(f"    emit_error: {err}", file=sys.stderr)
                continue
            transformed_path = base / "emit" / "transformed.dl"
            transformed_path.write_text(printed)

            wd_source = base / "source"
            wd_emit = base / "emitted"
            r_source = run_souffle(dl_path, facts_dir, wd_source)
            r_emit = run_souffle(transformed_path, facts_dir, wd_emit)
            row["status_source"] = r_source["status"]
            row["status_emit"] = r_emit["status"]
            print(f"    source={r_source['status']} emit={r_emit['status']}", file=sys.stderr)

            if r_source["status"] != "ok" or r_emit["status"] != "ok":
                row["status"] = "DNF"
                all_rows.append(row)
                continue

            profile_source = tuple_analyze(wd_source / "prof.log")
            profile_emit = tuple_analyze(wd_emit / "prof.log")
            row["T_excl_copy_source"] = profile_source["T_excl_copy"]
            row["T_excl_copy_emit"] = profile_emit["T_excl_copy"]
            row["tuple_totals_identical"] = profile_source["T_excl_copy"] == profile_emit["T_excl_copy"]

            ans_source = sorted_lines(wd_source / f"{answer_rel}.csv")
            ans_emit = sorted_lines(wd_emit / f"{answer_rel}.csv")
            row["answer_relation"] = answer_rel
            row["answers_identical"] = ans_source == ans_emit
            if not row["answers_identical"]:
                row["diff"] = {
                    "source_only": sorted(set(ans_source) - set(ans_emit))[:20],
                    "emit_only": sorted(set(ans_emit) - set(ans_source))[:20],
                }
            row["status"] = "ok" if (row["tuple_totals_identical"] and row["answers_identical"]) else "MISMATCH"
            all_rows.append(row)

    attempted = len(all_rows)
    identical = sum(1 for r in all_rows if r.get("status") == "ok")
    summary = {
        "transformer": transformer,
        "attempted": attempted,
        "identical": identical,
        "identical_of_attempted": f"{identical}/{attempted}",
        "rows": all_rows,
    }
    MEASUREMENTS.mkdir(parents=True, exist_ok=True)
    (MEASUREMENTS / "summary.json").write_text(json.dumps(summary, indent=2))
    print(json.dumps({k: v for k, v in summary.items() if k != "rows"}, indent=2))
    for r in all_rows:
        if r.get("status") != "ok":
            print("NOTABLE:", json.dumps(r, indent=2)[:1500])


if __name__ == "__main__":
    main()
