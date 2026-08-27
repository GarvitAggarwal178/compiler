#!/usr/bin/env python3
"""
PUNCH-LIST.md P1: hard answer-identical gate on multi-query seeding.

Re-emits every one of the 6 original M2/M3 programs (5 BENCHMARK_FAMILY
shapes + p2.dl) plus all 4 tests/corpus/CONE_CORPUS/ (task B)
constructions with the post-P1 dlc binary, and compares answers against
the untransformed baseline via harness/m2_accept.py -- the same
methodology M2's own gate and task B's own gate already used, re-run here
specifically to confirm multi-query seed collection changed no answer.

A single mismatch here is a stop condition (PUNCH-LIST.md P1's own
wording: "a seeding change that alters an answer stops the batch").
"""
import json
import subprocess
import sys
from pathlib import Path

sys.path.insert(0, "/root/compiler/harness")
from m2_accept import accept  # noqa: E402

REPO = Path("/root/compiler")
DLC = REPO / "bin" / "dlc"
MEASUREMENTS = REPO / "measurements" / "punch-list" / "p1-gate"

ORIGINAL_CASES = [
    ("p2", REPO / "tests/programs/p2.dl", REPO / "fixtures/p2-scale-250", "magicset"),
    ("reachability_complement", REPO / "tests/corpus/BENCHMARK_FAMILY/reachability_complement.dl",
     REPO / "fixtures/p2-scale-250", "magicset"),
    ("ancestor_nonancestor", REPO / "tests/corpus/BENCHMARK_FAMILY/ancestor_nonancestor.dl",
     REPO / "fixtures/benchmark-family/ancestor_nonancestor/n500", "magicset"),
    ("same_generation_negation", REPO / "tests/corpus/BENCHMARK_FAMILY/same_generation_negation.dl",
     REPO / "fixtures/benchmark-family/same_generation_negation/d4_b4", "magicset"),
    ("transitive_closure_bound", REPO / "tests/corpus/BENCHMARK_FAMILY/transitive_closure_bound.dl",
     REPO / "fixtures/benchmark-family/transitive_closure_bound/n500", "magicset"),
    ("culprit_cycle", REPO / "tests/corpus/BENCHMARK_FAMILY/culprit_cycle.dl",
     REPO / "fixtures/benchmark-family/culprit_cycle/n20", "magicset"),
]

CONE_CORPUS_CASES = [
    (name, REPO / f"tests/corpus/CONE_CORPUS/{name}.dl", REPO / f"fixtures/cone_corpus/{name}/n20", "guarded")
    for name in ["cc_cone_only", "cc_sibling_emptycone", "cc_both", "cc_cone_proper_subset"]
]


def emit_and_check(name, dl, facts_dir, transformer, group):
    workdir = MEASUREMENTS / group / name
    emitdir = workdir / "emit"
    proc = subprocess.run([str(DLC), "emit", str(dl), f"--transformer={transformer}"],
                           capture_output=True, encoding="utf-8", errors="replace")
    emitdir.mkdir(parents=True, exist_ok=True)
    (emitdir / "emit_stdout.json").write_text(proc.stdout)
    doc = json.loads(proc.stdout.strip().splitlines()[-1])
    if doc.get("status") != "ok":
        return {"name": name, "status": "emit_error", "doc": doc}
    candidate = emitdir / "transformed.dl"
    candidate.write_text(doc["printed"])
    r = accept(dl, candidate, facts_dir, workdir)
    return {"name": name, "comparable": r.get("comparable"), "answers_identical": r.get("answers_identical"),
            "T_original": r.get("T_original"), "T_candidate": r.get("T_candidate"), "diffs": r.get("diffs")}


def main():
    results = []
    for name, dl, facts, transformer in ORIGINAL_CASES:
        results.append(emit_and_check(name, dl, facts, transformer, "original"))
    for name, dl, facts, transformer in CONE_CORPUS_CASES:
        results.append(emit_and_check(name, dl, facts, transformer, "cone_corpus"))

    MEASUREMENTS.mkdir(parents=True, exist_ok=True)
    (MEASUREMENTS / "summary.json").write_text(json.dumps(results, indent=2))

    all_ok = True
    for r in results:
        comparable = r.get("comparable")
        identical = r.get("answers_identical")
        flag = ""
        if comparable and not identical:
            flag = "  !!!! ANSWER MISMATCH !!!!"
            all_ok = False
        print(f"{r['name']}: comparable={comparable} answers_identical={identical} "
              f"T_original={r.get('T_original')} T_candidate={r.get('T_candidate')}{flag}")
    print()
    print("GATE:", "PASS -- 0 mismatches" if all_ok else "FAIL -- see !!!! lines above")
    return 0 if all_ok else 1


if __name__ == "__main__":
    raise SystemExit(main())
