#!/usr/bin/env python3
import json, sys
from pathlib import Path
sys.path.insert(0, "/root/compiler/harness")
from m2_accept import accept

REPO = Path("/root/compiler")
FAMILY = REPO / "tests" / "corpus" / "BENCHMARK_FAMILY"
ORIG = FAMILY / "ancestor_nonancestor.dl"
V1 = FAMILY / "guarded" / "ancestor_nonancestor_guarded.dl"
V2 = FAMILY / "guarded" / "ancestor_nonancestor_guarded_v2.dl"
FIXTURES = REPO / "fixtures" / "benchmark-family" / "ancestor_nonancestor"
OUT = REPO / "measurements" / "night03-t5"

points = ["n500", "n1000", "n2000", "n4000", "n8000"]
rows = []
for tag in points:
    fdir = FIXTURES / tag
    wd_base = OUT / tag / "orig_vs_v2"
    wd_v1v2 = OUT / tag / "v1_vs_v2"
    r_base = accept(ORIG, V2, fdir, wd_base)
    r_v1 = accept(V1, V2, fdir, wd_v1v2)
    row = {
        "tag": tag,
        "T_none_orig": r_base.get("T_original"),
        "T_guard_v2": r_base.get("T_candidate"),
        "answers_v2_vs_baseline_identical": r_base.get("answers_identical"),
        "T_guard_v1": r_v1.get("T_original"),
        "answers_v2_vs_v1_identical": r_v1.get("answers_identical"),
        "comparable_base": r_base.get("comparable"),
        "comparable_v1": r_v1.get("comparable"),
    }
    if r_base.get("diffs"):
        row["diff_sample_vs_baseline"] = {k: v for k, v in list(r_base["diffs"].items())[:1]}
    rows.append(row)
    print(json.dumps(row, indent=2), file=sys.stderr)

OUT.mkdir(parents=True, exist_ok=True)
(OUT / "summary.json").write_text(json.dumps(rows, indent=2))
print(json.dumps(rows, indent=2))
