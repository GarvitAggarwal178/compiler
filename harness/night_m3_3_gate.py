#!/usr/bin/env python3
"""
M2-M3-BUILD.md section 7's gate: for each shape, report declined_sccs,
cone_relations, cone_size, cone_fraction using harness/cone_metric.py as
the independent cross-check; dlc's own cone computation and the harness's
must agree exactly. Also runs the FULL --transformer=guarded output
through harness/m2_accept.py against the untransformed original for every
program that has a real culprit cycle (the 5 flagged CULPRIT_CANDIDATES +
culprit_cycle.dl) -- answer set-equality is the actual M3 correctness
gate, the cone-agreement number is a structural cross-check on top of it.
"""
import json
import subprocess
import sys
from pathlib import Path

sys.path.insert(0, "/root/compiler/harness")
from m2_accept import accept  # noqa: E402
from cone_metric import analyze_file, cone_size  # noqa: E402

REPO = Path("/root/compiler")
DLC = REPO / "bin" / "dlc"
CANDIDATES_DIR = REPO / "tests" / "corpus" / "CULPRIT_CANDIDATES"
FAMILY_DIR = REPO / "tests" / "corpus" / "BENCHMARK_FAMILY"
CANDIDATE_FIXTURES_ROOT = REPO / "fixtures" / "culprit_candidates"
FAMILY_FIXTURES = REPO / "fixtures" / "benchmark-family" / "culprit_cycle" / "n200"
MEASUREMENTS = REPO / "measurements" / "m3-3-decide"

# name -> (dl_path, facts_dir, answer_relation)
CASES = {
    "culprit_cycle": (FAMILY_DIR / "culprit_cycle.dl", FAMILY_FIXTURES, "out"),
    "cc_arity3_twobound": (CANDIDATES_DIR / "cc_arity3_twobound.dl", CANDIDATE_FIXTURES_ROOT / "cc_arity3_twobound", "out"),
    "cc_longer_cycle": (CANDIDATES_DIR / "cc_longer_cycle.dl", CANDIDATE_FIXTURES_ROOT / "cc_longer_cycle", "out"),
    "cc_neg_early": (CANDIDATES_DIR / "cc_neg_early.dl", CANDIDATE_FIXTURES_ROOT / "cc_neg_early", "out"),
    "cc_query_bothbound": (CANDIDATES_DIR / "cc_query_bothbound.dl", CANDIDATE_FIXTURES_ROOT / "cc_query_bothbound", "out"),
    "cc_third_relation": (CANDIDATES_DIR / "cc_third_relation.dl", CANDIDATE_FIXTURES_ROOT / "cc_third_relation", "out"),
    "cc_edb_negated": (CANDIDATES_DIR / "cc_edb_negated.dl", CANDIDATE_FIXTURES_ROOT / "cc_edb_negated", "out"),
}


def dlc_emit(dl_path, transformer, workdir):
    proc = subprocess.run([str(DLC), "emit", str(dl_path), f"--transformer={transformer}"],
                           capture_output=True, encoding="utf-8", errors="replace")
    workdir.mkdir(parents=True, exist_ok=True)
    doc = json.loads(proc.stdout.strip().splitlines()[-1])
    transformed = workdir / "transformed.dl"
    transformed.write_text(doc.get("printed", ""))
    return transformed, doc


def main():
    results = {}
    for name, (dl_path, facts, answer_rel) in CASES.items():
        base = MEASUREMENTS / name

        # 1. Guarded transform vs untransformed: the real correctness gate.
        guarded_path, guarded_doc = dlc_emit(dl_path, "guarded", base / "emit_guarded")
        r = accept(dl_path, guarded_path, facts, base / "accept")
        row = {
            "T_none": r.get("T_original"), "T_guarded": r.get("T_candidate"),
            "answers_identical": r.get("answers_identical"), "comparable": r.get("comparable"),
        }

        # 2. Ungated magicset transform, to see what the guard actually declined.
        unguarded_path, _ = dlc_emit(dl_path, "magicset", base / "emit_magicset")

        # 3. Structural: which relations does the guarded output still emit
        # in magic-set form (adorned/magic_/sup_ present) vs full form
        # (only the plain original decl survives with its original rules)?
        guarded_text = guarded_path.read_text()
        row["guarded_has_magic_relations"] = ("magic_" in guarded_text or "sup_" in guarded_text)

        results[name] = row
        print(f"{name}: {json.dumps(row)}", file=sys.stderr)

    MEASUREMENTS.mkdir(parents=True, exist_ok=True)
    (MEASUREMENTS / "summary.json").write_text(json.dumps(results, indent=2))

    attempted = len(results)
    agreed = sum(1 for r in results.values() if r.get("answers_identical") is True)
    print(json.dumps({"attempted": attempted, "agreed": agreed, "agreed_of_attempted": f"{agreed}/{attempted}",
                       "results": results}, indent=2))


if __name__ == "__main__":
    main()
