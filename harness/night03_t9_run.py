#!/usr/bin/env python3
import json, sys
from pathlib import Path
sys.path.insert(0, "/root/compiler/harness")
from cone_metric import analyze_file, cone_size

REPO = Path("/root/compiler")
FAMILY = REPO / "tests" / "corpus" / "BENCHMARK_FAMILY"

SHAPES = [
    ("same_generation_negation", "notsg"),
    ("transitive_closure_bound", None),  # no negation, E_recoverable=0 per T4 baseline
    ("ancestor_nonancestor", "nonancestor"),
    ("reachability_complement", "unreach"),
    ("culprit_cycle", "p"),
]

results = {}
for shape, neg_rel in SHAPES:
    path = FAMILY / f"{shape}.dl"
    text = path.read_text()
    heads, all_edges, scc_of, scc_members = analyze_file(text)
    if neg_rel is None:
        results[shape] = {"negation_bearing_relation": None, "note": "no negation in this shape (T4 baseline: E_recoverable=0)"}
        continue
    declined_scc = scc_members[scc_of[neg_rel]]
    result = cone_size(heads, all_edges, [declined_scc])
    result["negation_bearing_relation"] = neg_rel
    result["idb_relation_count"] = len(heads)
    result["all_idb_relations"] = sorted(heads)
    results[shape] = result
    print(shape, "declined=", declined_scc, "->", json.dumps(result, indent=2), file=sys.stderr)

out_dir = REPO / "measurements" / "night03-t9"
out_dir.mkdir(parents=True, exist_ok=True)
(out_dir / "cone-metric-summary.json").write_text(json.dumps(results, indent=2))
print(json.dumps(results, indent=2))
