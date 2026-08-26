#!/usr/bin/env python3
"""
M1 §3.9 gate two, the milestone's headline number: "T_naive vs
T_semi-naive, exact tuple counts, on a fixed program set."

Fixed program set: the 5 tests/corpus/BENCHMARK_FAMILY/ shapes (already
pre-registered this project's own corpus ruling, not chosen after seeing
results -- CLAUDE.md section 9's "metric selection after the fact"
failure mode), each at its own smallest pre-registered scale point
(SCALE_POINTS.json), so runtime stays fast. All 5 already pass `dlc
check` cleanly (confirmed this session). Both dlc run (naive, §3.8) and
dlc run-seminaive (§3.9) are run against the same fixture and facts dir;
T is the sum, across every relation, of profile.json's num-tuples plus
every iteration delta (the same derived-tuple metric this project has
used in every measurement since Phase 0, harness/parse_profile.py's own
summation). Answer-set equality between the two engines is checked too
(not just the headline count) -- a matching T with a different answer
set would be a worse finding than a mismatched T, and must not be masked
by only reporting the number.
"""
import json
import subprocess
import sys
import tempfile
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import dlc_interface  # noqa: E402

REPO = Path(__file__).resolve().parents[1]
FAMILY_DIR = REPO / "tests" / "corpus" / "BENCHMARK_FAMILY"

SHAPES = {
    "same_generation_negation": {
        "dl": "same_generation_negation.dl",
        "facts": REPO / "fixtures" / "benchmark-family" / "same_generation_negation" / "d4_b4",
        "outputs": ["q_notsg"],
    },
    "transitive_closure_bound": {
        "dl": "transitive_closure_bound.dl",
        "facts": REPO / "fixtures" / "benchmark-family" / "transitive_closure_bound" / "n500",
        "outputs": ["q_tc"],
    },
    "ancestor_nonancestor": {
        "dl": "ancestor_nonancestor.dl",
        "facts": REPO / "fixtures" / "benchmark-family" / "ancestor_nonancestor" / "n500",
        "outputs": ["q_nonancestor"],
    },
    "reachability_complement": {
        "dl": "reachability_complement.dl",
        "facts": REPO / "fixtures" / "p2-scale-250",
        "outputs": ["q_unreach"],
    },
    "culprit_cycle": {
        "dl": "culprit_cycle.dl",
        "facts": REPO / "fixtures" / "benchmark-family" / "culprit_cycle" / "n20",
        "outputs": ["out"],
    },
}


def total_tuples(profile_doc):
    total = 0
    per_relation = {}
    for name, rel in profile_doc["root"]["program"]["relation"].items():
        t = rel.get("num-tuples", 0) or 0
        if rel.get("iteration"):
            t += sum(v.get("num-tuples", 0) or 0 for v in rel["iteration"].values())
        per_relation[name] = t
        total += t
    return total, per_relation


def run_one(subcommand, dl_path, facts_dir, out_dir):
    proc = subprocess.run(
        [str(dlc_interface.DLC_BINARY), subcommand, str(dl_path), str(facts_dir), str(out_dir)],
        capture_output=True, encoding="utf-8", errors="replace",
    )
    doc = json.loads(proc.stdout.strip().splitlines()[-1]) if proc.stdout.strip() else {}
    return doc


def main():
    dlc_interface.build_dlc()
    results = {}
    for shape, spec in SHAPES.items():
        dl_path = FAMILY_DIR / spec["dl"]
        with tempfile.TemporaryDirectory() as naive_dir, tempfile.TemporaryDirectory() as semi_dir:
            naive_doc = run_one("run", dl_path, spec["facts"], naive_dir)
            semi_doc = run_one("run-seminaive", dl_path, spec["facts"], semi_dir)
            if naive_doc.get("status") != "ok" or semi_doc.get("status") != "ok":
                results[shape] = {"status": "not_ok", "naive_status": naive_doc.get("status"),
                                   "semi_status": semi_doc.get("status")}
                continue

            naive_profile = json.loads((Path(naive_dir) / "profile.json").read_text())
            semi_profile = json.loads((Path(semi_dir) / "profile.json").read_text())
            t_naive, naive_per_rel = total_tuples(naive_profile)
            t_semi, semi_per_rel = total_tuples(semi_profile)

            answers_match = True
            mismatches = []
            for out_rel in spec["outputs"]:
                naive_csv = (Path(naive_dir) / f"{out_rel}.csv")
                semi_csv = (Path(semi_dir) / f"{out_rel}.csv")
                naive_lines = sorted(naive_csv.read_text().splitlines()) if naive_csv.is_file() else []
                semi_lines = sorted(semi_csv.read_text().splitlines()) if semi_csv.is_file() else []
                if naive_lines != semi_lines:
                    answers_match = False
                    mismatches.append(out_rel)

            attempts_naive = naive_doc.get("derivation_attempts", 0)
            attempts_semi = semi_doc.get("derivation_attempts", 0)
            results[shape] = {
                "status": "ok", "T_naive": t_naive, "T_semi_naive": t_semi,
                "ratio_naive_over_semi": (t_naive / t_semi) if t_semi else None,
                "derivation_attempts_naive": attempts_naive, "derivation_attempts_semi_naive": attempts_semi,
                "derivation_attempts_ratio": (attempts_naive / attempts_semi) if attempts_semi else None,
                "answers_match": answers_match, "mismatched_relations": mismatches,
                "naive_per_relation": naive_per_rel, "semi_per_relation": semi_per_rel,
            }

    out_path = REPO / "measurements" / "m1-3.9-gate2-headline-summary.json"
    out_path.write_text(json.dumps(results, indent=2))

    print("=== M1 headline: T_naive vs T_semi-naive, per shape (never aggregated) ===")
    for shape, r in results.items():
        if r["status"] != "ok":
            print(f"{shape}: NOT OK ({r})")
            continue
        print(f"{shape}: T_naive={r['T_naive']:,} T_semi_naive={r['T_semi_naive']:,} "
              f"ratio={r['ratio_naive_over_semi']:.2f}x answers_match={r['answers_match']} | "
              f"derivation_attempts naive={r['derivation_attempts_naive']:,} "
              f"semi_naive={r['derivation_attempts_semi_naive']:,} "
              f"ratio={(r['derivation_attempts_ratio'] or 0):.2f}x")

    any_bad = any(r["status"] != "ok" or not r.get("answers_match", True) for r in results.values())
    if any_bad:
        print("GATE ISSUE: at least one shape did not run cleanly or answers diverged between engines", file=sys.stderr)
        raise SystemExit(1)
    print("GATE PASSED: all 5 shapes ran cleanly, naive and semi-naive answer sets match exactly.", file=sys.stderr)


if __name__ == "__main__":
    main()
