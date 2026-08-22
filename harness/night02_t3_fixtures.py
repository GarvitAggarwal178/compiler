#!/usr/bin/env python3
"""
NIGHT-BATCH-02 T3: benchmark-family fixture generation. For every shape in
SCALE_POINTS.json (except reachability_complement, which reuses
NIGHT-BATCH-01 T6's already-generated, already-verified fixtures per the
JSON's own note), materializes the pre-registered fixtures via
run_benchmark_family.generate_fixtures_only(), then:

  - records seed, SHA-256 of each generated fact file, node/edge counts
  - computes the reachable-set size from each shape's query constant,
    where that is a structural (fixture-only) property -- see per-shape
    notes in the report for where it is NOT (same_generation_negation,
    culprit_cycle's true p-reachability) because that requires running
    Souffle, which is T4's job, not this one
  - regenerates once more and confirms byte-identical SHA-256 (idempotency)

Does not touch tests/corpus/BENCHMARK_FAMILY/ or SCALE_POINTS.json (both
pre-registered, prohibition 0.2 of the night-02 issue). Does not invoke
Soufflé.
"""
import hashlib
import json
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import fixtures_lib  # noqa: E402
import run_benchmark_family as rbf  # noqa: E402

REPO = rbf.REPO
FAMILY_DIR = rbf.FAMILY_DIR
SCALE_POINTS = rbf.SCALE_POINTS
FIXTURES_ROOT = REPO / "fixtures" / "benchmark-family"


def sha256_file(path):
    h = hashlib.sha256()
    h.update(path.read_bytes())
    return h.hexdigest()


def hash_tree(root):
    out = {}
    for p in sorted(root.rglob("*.facts")):
        out[str(p.relative_to(FIXTURES_ROOT))] = sha256_file(p)
    return out


def main():
    gen_info = rbf.generate_fixtures_only()
    hashes_run1 = hash_tree(FIXTURES_ROOT)

    gen_info2 = rbf.generate_fixtures_only()
    hashes_run2 = hash_tree(FIXTURES_ROOT)

    idempotent = (hashes_run1 == hashes_run2) and (gen_info == gen_info2)
    mismatches = []
    if not idempotent:
        for k in set(hashes_run1) | set(hashes_run2):
            if hashes_run1.get(k) != hashes_run2.get(k):
                mismatches.append(k)

    structural = {}

    sg = SCALE_POINTS["same_generation_negation"]
    for pt in sg["points"]:
        tag = f"d{pt['depth']}_b{pt['branching']}"
        fdir = FIXTURES_ROOT / "same_generation_negation" / tag
        edges = [tuple(map(int, ln.split("\t"))) for ln in (fdir / "parent.facts").read_text().splitlines() if ln]
        persons = [int(ln) for ln in (fdir / "person.facts").read_text().splitlines() if ln]
        # parent.facts is (child, par); to reach descendants from root 0 the
        # walk must follow par -> child, i.e. the reverse of the stored tuples.
        descendant_edges = [(par, child) for (child, par) in edges]
        reach = fixtures_lib.bfs_reachable(descendant_edges, 0)
        structural[f"same_generation_negation/{tag}"] = {
            "n_persons": len(persons), "n_edges": len(edges),
            "reachable_from_0_note": "root 0 reaches the whole tree by construction "
                                      "(it IS the root; walk follows par->child, the "
                                      "reverse of the stored child->par tuples) -- "
                                      "equals n_persons, not a partition test the way "
                                      "core_size is for the graph shapes. sg/notsg "
                                      "cardinality at 0 requires running Soufflé (T4), "
                                      "not computable from the fixture alone.",
            "reachable_from_0": len(reach),
        }

    for shape_name, rel_name in (("transitive_closure_bound", "edge"), ("ancestor_nonancestor", "parent")):
        spec = SCALE_POINTS[shape_name]
        for pt in spec["points"]:
            tag = f"n{pt['n']}"
            fdir = FIXTURES_ROOT / shape_name / tag
            edges = [tuple(map(int, ln.split("\t"))) for ln in (fdir / f"{rel_name}.facts").read_text().splitlines() if ln]
            reach = fixtures_lib.bfs_reachable(edges, 1)
            entry = {"n": pt["n"], "n_edges": len(edges), "reachable_from_1": len(reach)}
            if shape_name == "ancestor_nonancestor":
                persons = [int(ln) for ln in (fdir / "person.facts").read_text().splitlines() if ln]
                entry["n_persons"] = len(persons)
            structural[f"{shape_name}/{tag}"] = entry

    cc = SCALE_POINTS["culprit_cycle"]
    for pt in cc["points"]:
        tag = f"n{pt['n']}"
        fdir = FIXTURES_ROOT / "culprit_cycle" / tag
        base_edges = [tuple(map(int, ln.split("\t"))) for ln in (fdir / "base.facts").read_text().splitlines() if ln]
        e_edges = [tuple(map(int, ln.split("\t"))) for ln in (fdir / "e.facts").read_text().splitlines() if ln]
        blocked = [int(ln) for ln in (fdir / "blocked.facts").read_text().splitlines() if ln]
        reach_e_only = fixtures_lib.bfs_reachable(e_edges, 1)
        structural[f"culprit_cycle/{tag}"] = {
            "n": pt["n"], "n_base_edges": len(base_edges), "n_e_edges": len(e_edges),
            "n_blocked": len(blocked),
            "reachable_from_1_via_e_only_note": "base-case-only lower bound on p's "
                                                 "true reachable set from node 1 -- p "
                                                 "also derives through q/base and is "
                                                 "gated by !s(z); the true reachable "
                                                 "set requires running Soufflé (T4), "
                                                 "not computable from the fixture alone.",
            "reachable_from_1_via_e_only": len(reach_e_only),
        }

    summary = {
        "idempotent": idempotent,
        "mismatches": mismatches,
        "generation_info": gen_info,
        "structural": structural,
        "hashes": hashes_run1,
    }
    out_path = REPO / "measurements" / "night02-t3-fixtures-summary.json"
    out_path.write_text(json.dumps(summary, indent=2))
    print(json.dumps({"idempotent": idempotent, "mismatches": mismatches, "structural": structural}, indent=2))

    if not idempotent:
        print("ABORT: fixture generation is not idempotent under the same seed", file=sys.stderr)
        raise SystemExit(1)


if __name__ == "__main__":
    main()
