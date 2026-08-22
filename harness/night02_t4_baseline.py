#!/usr/bin/env python3
"""
NIGHT-BATCH-02 T4: baseline sweep. Soufflé only, no dlc, no hand-transforms.
Per shape (all 5 in tests/corpus/BENCHMARK_FAMILY/), per pre-registered
scale point, two configurations: untransformed and --magic-transform=*.
Records T_none/T_souffle (both copy conventions), E_recoverable with its
per-relation breakdown, and checks the answer (.output) relation is
sorted-set-equal between the two configurations.

reachability_complement reuses the fixtures NIGHT-BATCH-01 T6 already
generated (fixtures/p2-scale-<n>/, no new fixtures per SCALE_POINTS.json's
own note) but is re-run here against tests/corpus/BENCHMARK_FAMILY/
reachability_complement.dl (not tests/programs/p2.dl) so its row uses the
same instrumentation (excl/incl-copy, neglabel breakdown) as the other 4
shapes -- p2.dl and reachability_complement.dl are structurally identical
(same relations, same rules; only the .output relation name differs, q2 vs
q_unreach), confirmed by inspection before running.

Abort condition, per shape (not per batch): answer relations differ between
configurations at any scale point. Log in full, drop that shape, continue
with the others. Ascending scale-point order so a cap truncates the top of
a shape's curve, not the middle. timeout 300s, 8GB memory cap per
NIGHT-BATCH-02 protocol 0.3; a DNF is never retried with a higher cap.
"""
import json
import resource
import subprocess
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from tuple_report import analyze as tuple_analyze  # noqa: E402

REPO = Path(__file__).resolve().parents[1]
FAMILY_DIR = REPO / "tests" / "corpus" / "BENCHMARK_FAMILY"
SCALE_POINTS = json.loads((FAMILY_DIR / "SCALE_POINTS.json").read_text())
FIXTURES_ROOT = REPO / "fixtures" / "benchmark-family"
MEASUREMENTS = REPO / "measurements" / "night02-t4"

MEM_LIMIT_BYTES = 8 * 1024 * 1024 * 1024
TIMEOUT_S = 300

# (shape, dl filename, answer relation, [(tag, facts_dir)]) built below.
SHAPES = {
    "same_generation_negation": {
        "dl": "same_generation_negation.dl", "answer": "q_notsg",
        "points": [
            (f"d{pt['depth']}_b{pt['branching']}",
             FIXTURES_ROOT / "same_generation_negation" / f"d{pt['depth']}_b{pt['branching']}")
            for pt in SCALE_POINTS["same_generation_negation"]["points"]
        ],
    },
    "transitive_closure_bound": {
        "dl": "transitive_closure_bound.dl", "answer": "q_tc",
        "points": [
            (f"n{pt['n']}", FIXTURES_ROOT / "transitive_closure_bound" / f"n{pt['n']}")
            for pt in SCALE_POINTS["transitive_closure_bound"]["points"]
        ],
    },
    "ancestor_nonancestor": {
        "dl": "ancestor_nonancestor.dl", "answer": "q_nonancestor",
        "points": [
            (f"n{pt['n']}", FIXTURES_ROOT / "ancestor_nonancestor" / f"n{pt['n']}")
            for pt in SCALE_POINTS["ancestor_nonancestor"]["points"]
        ],
    },
    "culprit_cycle": {
        "dl": "culprit_cycle.dl", "answer": "out",
        "points": [
            (f"n{pt['n']}", FIXTURES_ROOT / "culprit_cycle" / f"n{pt['n']}")
            for pt in SCALE_POINTS["culprit_cycle"]["points"]
        ],
    },
    "reachability_complement": {
        "dl": "reachability_complement.dl", "answer": "q_unreach",
        "points": [
            (f"n{n}", REPO / "fixtures" / f"p2-scale-{n}")
            for n in SCALE_POINTS["reachability_complement"]["points_already_run"]
        ],
    },
}


def _limit_mem():
    resource.setrlimit(resource.RLIMIT_AS, (MEM_LIMIT_BYTES, MEM_LIMIT_BYTES))


def run(dl_path, facts_dir, workdir, magic, log_name="prof.log"):
    workdir.mkdir(parents=True, exist_ok=True)
    cmd = ["souffle", "-F", str(facts_dir), "-D", str(workdir), "-p", log_name]
    if magic:
        cmd.append("--magic-transform=*")
    cmd.append(str(dl_path))
    (workdir / "cmd.txt").write_text(" ".join(cmd) + "\n")
    try:
        proc = subprocess.run(
            cmd, cwd=str(workdir), capture_output=True, text=True,
            timeout=TIMEOUT_S, preexec_fn=_limit_mem,
        )
    except subprocess.TimeoutExpired:
        return {"status": f"DNF:timeout-{TIMEOUT_S}s"}
    except MemoryError:
        return {"status": "DNF:memcap-8gb"}
    (workdir / "stdout.txt").write_text(proc.stdout)
    (workdir / "stderr.txt").write_text(proc.stderr)
    if proc.returncode != 0:
        return {"status": f"error:returncode-{proc.returncode}", "stderr": proc.stderr[:500]}
    return {"status": "ok", "workdir": workdir}


def sorted_lines(path: Path):
    return sorted(path.read_text().splitlines()) if path.is_file() else None


def main():
    all_results = {}
    for shape, spec in SHAPES.items():
        print(f"===== shape={shape} =====", file=sys.stderr)
        dl_path = FAMILY_DIR / spec["dl"]
        answer_rel = spec["answer"]
        shape_results = []
        aborted = False
        for tag, facts_dir in spec["points"]:
            print(f"  --- {tag} ---", file=sys.stderr)
            row = {"tag": tag, "facts_dir": str(facts_dir)}
            wd_none = MEASUREMENTS / shape / f"{tag}-none"
            wd_souffle = MEASUREMENTS / shape / f"{tag}-souffle"

            r_none = run(dl_path, facts_dir, wd_none, magic=False)
            r_souffle = run(dl_path, facts_dir, wd_souffle, magic=True)
            row["status_none"] = r_none["status"]
            row["status_souffle"] = r_souffle["status"]
            print(f"    none={r_none['status']} souffle={r_souffle['status']}", file=sys.stderr)

            if r_none["status"] != "ok" or r_souffle["status"] != "ok":
                shape_results.append(row)
                print(f"    {shape}/{tag}: at least one config did not complete "
                      f"(DNF/error), stopping ascent for this shape here", file=sys.stderr)
                break

            a_none = sorted_lines(wd_none / f"{answer_rel}.csv")
            a_souffle = sorted_lines(wd_souffle / f"{answer_rel}.csv")
            identical = a_none == a_souffle
            row["answers_identical"] = identical
            if not identical:
                row["answer_none_len"] = len(a_none) if a_none is not None else None
                row["answer_souffle_len"] = len(a_souffle) if a_souffle is not None else None
                shape_results.append(row)
                print(f"ABORT for shape={shape} at {tag}: {answer_rel}.csv not "
                      f"identical between configs", file=sys.stderr)
                aborted = True
                break

            tr_none = tuple_analyze(wd_none / "prof.log")
            tr_souffle = tuple_analyze(wd_souffle / "prof.log")
            row["T_none_excl_copy"] = tr_none["T_excl_copy"]
            row["T_none_incl_copy"] = tr_none["T_incl_copy"]
            row["T_souffle_excl_copy"] = tr_souffle["T_excl_copy"]
            row["T_souffle_incl_copy"] = tr_souffle["T_incl_copy"]
            row["E_recoverable"] = tr_souffle["E_recoverable"]
            row["neglabel_relations"] = tr_souffle["neglabel_relations"]
            shape_results.append(row)
            print(f"    T_none={row['T_none_excl_copy']} T_souffle={row['T_souffle_excl_copy']} "
                  f"E_recoverable={row['E_recoverable']}", file=sys.stderr)

        all_results[shape] = {"aborted": aborted, "points": shape_results}
        MEASUREMENTS.mkdir(parents=True, exist_ok=True)
        (MEASUREMENTS / "summary.json").write_text(json.dumps(all_results, indent=2))

    (MEASUREMENTS / "summary.json").write_text(json.dumps(all_results, indent=2))
    print(json.dumps({k: v["aborted"] for k, v in all_results.items()}, indent=2))


if __name__ == "__main__":
    main()
