#!/usr/bin/env python3
"""
NIGHT-BATCH-02 T5: hand-guarded transforms, the headline. Runs the 4
committed guards under tests/corpus/BENCHMARK_FAMILY/guarded/ against every
scale point T4 already ran, records T_guard, and checks sorted-set-equality
of the answer relation against T4's untransformed baseline CSV (not a fresh
untransformed re-run -- T4 already produced and committed those).

transitive_closure_bound is excluded: T4 found E_recoverable=0 there (no
negation in the shape), so there is nothing to guard.

culprit_cycle's committed guard is the SAFE partial-restriction version
(see its own header comment); the unsafe general-adornment attempt that
Soufflé rejected for cyclic negation is scratch evidence only
(measurements/_scratch_night02_t5/culprit_cycle_unsafe_cyclic.dl), not run
by this script.

Abort condition, per shape: answer mismatch at any scale point, or a DNF.
Log fully, drop the shape, continue with the others. Ascending order,
300s/8GB caps, same as T4.
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
GUARDED_DIR = FAMILY_DIR / "guarded"
SCALE_POINTS = json.loads((FAMILY_DIR / "SCALE_POINTS.json").read_text())
FIXTURES_ROOT = REPO / "fixtures" / "benchmark-family"
T4_MEASUREMENTS = REPO / "measurements" / "night02-t4"
MEASUREMENTS = REPO / "measurements" / "night02-t5"

MEM_LIMIT_BYTES = 8 * 1024 * 1024 * 1024
TIMEOUT_S = 300

SHAPES = {
    "same_generation_negation": {
        "guarded_dl": "same_generation_negation_guarded.dl", "answer": "q_notsg",
        "points": [
            (f"d{pt['depth']}_b{pt['branching']}",
             FIXTURES_ROOT / "same_generation_negation" / f"d{pt['depth']}_b{pt['branching']}")
            for pt in SCALE_POINTS["same_generation_negation"]["points"]
        ],
    },
    "ancestor_nonancestor": {
        "guarded_dl": "ancestor_nonancestor_guarded.dl", "answer": "q_nonancestor",
        "points": [
            (f"n{pt['n']}", FIXTURES_ROOT / "ancestor_nonancestor" / f"n{pt['n']}")
            for pt in SCALE_POINTS["ancestor_nonancestor"]["points"]
        ],
    },
    "culprit_cycle": {
        "guarded_dl": "culprit_cycle_guarded.dl", "answer": "out",
        "points": [
            (f"n{pt['n']}", FIXTURES_ROOT / "culprit_cycle" / f"n{pt['n']}")
            for pt in SCALE_POINTS["culprit_cycle"]["points"]
        ],
    },
    "reachability_complement": {
        "guarded_dl": "reachability_complement_guarded.dl", "answer": "q_unreach",
        "points": [
            (f"n{n}", REPO / "fixtures" / f"p2-scale-{n}")
            for n in SCALE_POINTS["reachability_complement"]["points_already_run"]
        ],
    },
}


def _limit_mem():
    resource.setrlimit(resource.RLIMIT_AS, (MEM_LIMIT_BYTES, MEM_LIMIT_BYTES))


def run(dl_path, facts_dir, workdir, log_name="prof.log"):
    workdir.mkdir(parents=True, exist_ok=True)
    cmd = ["souffle", "-F", str(facts_dir), "-D", str(workdir), "-p", log_name, str(dl_path)]
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
    # Belt-and-suspenders: rc!=0 already catches Soufflé errors reliably
    # (re-verified under T9, docs/reports/night02-T9-diagnostics.md -- an
    # earlier claim here that rc could be 0 on a stratification error was
    # a Bash-tool/wsl.exe bridge artifact from streamed interactive
    # output, not a real Soufflé behavior; retracted). Checking stderr too
    # costs nothing and stays as a second line of defense.
    if proc.returncode != 0 or "Error:" in proc.stderr:
        return {"status": f"error:returncode-{proc.returncode}", "stderr": proc.stderr[:800]}
    return {"status": "ok", "workdir": workdir}


def sorted_lines(path: Path):
    return sorted(path.read_text().splitlines()) if path.is_file() else None


def main():
    all_results = {}
    for shape, spec in SHAPES.items():
        print(f"===== shape={shape} =====", file=sys.stderr)
        dl_path = GUARDED_DIR / spec["guarded_dl"]
        answer_rel = spec["answer"]
        shape_results = []
        aborted = False
        for tag, facts_dir in spec["points"]:
            print(f"  --- {tag} ---", file=sys.stderr)
            row = {"tag": tag, "facts_dir": str(facts_dir)}
            baseline_csv = T4_MEASUREMENTS / shape / f"{tag}-none" / f"{answer_rel}.csv"
            wd_guard = MEASUREMENTS / shape / f"{tag}-guard"

            r_guard = run(dl_path, facts_dir, wd_guard)
            row["status_guard"] = r_guard["status"]
            print(f"    guard={r_guard['status']}", file=sys.stderr)

            if r_guard["status"] != "ok":
                shape_results.append(row)
                print(f"    {shape}/{tag}: guard did not complete cleanly, "
                      f"stopping ascent for this shape here", file=sys.stderr)
                aborted = True
                break

            a_baseline = sorted_lines(baseline_csv)
            a_guard = sorted_lines(wd_guard / f"{answer_rel}.csv")
            identical = a_baseline == a_guard
            row["answers_identical_vs_t4_baseline"] = identical
            if not identical:
                row["baseline_len"] = len(a_baseline) if a_baseline is not None else None
                row["guard_len"] = len(a_guard) if a_guard is not None else None
                shape_results.append(row)
                print(f"ABORT for shape={shape} at {tag}: {answer_rel}.csv not "
                      f"identical to T4 baseline", file=sys.stderr)
                aborted = True
                break

            tr_guard = tuple_analyze(wd_guard / "prof.log")
            row["T_guard_excl_copy"] = tr_guard["T_excl_copy"]
            row["T_guard_incl_copy"] = tr_guard["T_incl_copy"]
            shape_results.append(row)
            print(f"    T_guard={row['T_guard_excl_copy']}", file=sys.stderr)

        all_results[shape] = {"aborted": aborted, "points": shape_results}
        MEASUREMENTS.mkdir(parents=True, exist_ok=True)
        (MEASUREMENTS / "summary.json").write_text(json.dumps(all_results, indent=2))

    (MEASUREMENTS / "summary.json").write_text(json.dumps(all_results, indent=2))
    print(json.dumps({k: v["aborted"] for k, v in all_results.items()}, indent=2))


if __name__ == "__main__":
    main()
