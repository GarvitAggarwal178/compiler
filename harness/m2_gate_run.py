#!/usr/bin/env python3
"""
M2-M3-BUILD.md section 4's real gate: harness/m2_accept.py on all 5
BENCHMARK_FAMILY shapes plus p2.dl, dlc's real magic-set transform vs the
untransformed original, at the SMALLEST pre-registered scale point per
shape (a full-scale-point sweep is section 9's headline run, not this
gate). Also reports T_dlc vs T_guard (the hand transform) where a guarded/
file exists, including BOTH ancestor_nonancestor variants (v1 and v2).
"""
import json
import subprocess
import sys
from pathlib import Path

sys.path.insert(0, "/root/compiler/harness")
from m2_accept import accept  # noqa: E402

REPO = Path("/root/compiler")
FAMILY = REPO / "tests" / "corpus" / "BENCHMARK_FAMILY"
DLC = REPO / "bin" / "dlc"

CASES = [
    ("same_generation_negation", FAMILY / "same_generation_negation.dl",
     REPO / "fixtures" / "benchmark-family" / "same_generation_negation" / "d4_b4",
     FAMILY / "guarded" / "same_generation_negation_guarded.dl"),
    ("transitive_closure_bound", FAMILY / "transitive_closure_bound.dl",
     REPO / "fixtures" / "benchmark-family" / "transitive_closure_bound" / "n500",
     None),
    ("ancestor_nonancestor", FAMILY / "ancestor_nonancestor.dl",
     REPO / "fixtures" / "benchmark-family" / "ancestor_nonancestor" / "n500",
     FAMILY / "guarded" / "ancestor_nonancestor_guarded.dl"),
    ("reachability_complement", FAMILY / "reachability_complement.dl",
     REPO / "fixtures" / "benchmark-family" / "reachability_complement",  # placeholder, fixed below
     FAMILY / "guarded" / "reachability_complement_guarded.dl"),
    ("culprit_cycle", FAMILY / "culprit_cycle.dl",
     REPO / "fixtures" / "benchmark-family" / "culprit_cycle" / "n20",
     FAMILY / "guarded" / "culprit_cycle_guarded.dl"),
    ("p2", REPO / "tests" / "programs" / "p2.dl",
     REPO / "fixtures" / "p2-scale-250",
     REPO / "tests" / "programs" / "p4prime.dl"),
]
# reachability_complement reuses p2-scale fixtures (SCALE_POINTS.json's own note)
CASES[3] = ("reachability_complement", FAMILY / "reachability_complement.dl",
            REPO / "fixtures" / "p2-scale-250",
            FAMILY / "guarded" / "reachability_complement_guarded.dl")

ANCESTOR_V2 = FAMILY / "guarded" / "ancestor_nonancestor_guarded_v2.dl"

MEASUREMENTS = REPO / "measurements" / "m2-gate"


def dlc_emit(dl_path, workdir):
    proc = subprocess.run([str(DLC), "emit", str(dl_path), "--transformer=magicset"],
                           capture_output=True, encoding="utf-8", errors="replace")
    workdir.mkdir(parents=True, exist_ok=True)
    (workdir / "emit_stdout.json").write_text(proc.stdout)
    doc = json.loads(proc.stdout.strip().splitlines()[-1])
    if doc.get("status") != "ok":
        return None, doc
    transformed = workdir / "transformed.dl"
    transformed.write_text(doc["printed"])
    return transformed, doc


def main():
    results = {}
    for name, orig, facts, guarded in CASES:
        base = MEASUREMENTS / name
        transformed, doc = dlc_emit(orig, base / "emit")
        if transformed is None:
            results[name] = {"status": "emit_error", "doc": doc}
            print(f"{name}: EMIT ERROR {doc}", file=sys.stderr)
            continue

        r = accept(orig, transformed, facts, base / "vs_original")
        row = {
            "T_none": r.get("T_original"), "T_dlc": r.get("T_candidate"),
            "answers_identical": r.get("answers_identical"), "comparable": r.get("comparable"),
        }
        if not r.get("comparable"):
            row["original_stderr"] = r.get("original_stderr", "")[:500]
            row["candidate_stderr"] = r.get("candidate_stderr", "")[:500]

        if guarded and guarded.is_file():
            rg = accept(guarded, transformed, facts, base / "vs_guarded")
            row["T_guard"] = rg.get("T_original")
            row["dlc_vs_guard_answers_identical"] = rg.get("answers_identical")

        if name == "ancestor_nonancestor" and ANCESTOR_V2.is_file():
            rv2 = accept(ANCESTOR_V2, transformed, facts, base / "vs_v2")
            row["T_guard_v2"] = rv2.get("T_original")
            row["dlc_vs_v2_answers_identical"] = rv2.get("answers_identical")

        results[name] = row
        print(f"{name}: {json.dumps(row, indent=2)}", file=sys.stderr)

    MEASUREMENTS.mkdir(parents=True, exist_ok=True)
    (MEASUREMENTS / "summary.json").write_text(json.dumps(results, indent=2))
    print(json.dumps(results, indent=2))


if __name__ == "__main__":
    main()
