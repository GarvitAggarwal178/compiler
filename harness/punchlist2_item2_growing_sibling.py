#!/usr/bin/env python3
"""
PUNCH-LIST-2 item 2: measure cc_growing_sibling.dl at the same scale
points as the original task-B constructions (n=20/50/100), testing the
Q13 prediction (docs/OPEN_QUESTIONS.md): does T_none/T_guarded GROW with
n instead of shrinking, once the sibling's own fixture is scaled with n
instead of held fixed?

Same T_none/T_souffle/T_dlc methodology as harness/night04_b_cone_gate.py,
reused here for one program rather than four.
"""
import json
import subprocess
import sys
from pathlib import Path

sys.path.insert(0, "/root/compiler/harness")
from fixtures_lib import gen_growing_sibling_facts, write_facts  # noqa: E402
from tuple_report import analyze as tuple_analyze  # noqa: E402
from cone_metric import analyze_file as cone_analyze_file, cone_size  # noqa: E402
from punchlist2_p1_decompose import origin_of  # noqa: E402

REPO = Path("/root/compiler")
DLC = REPO / "bin" / "dlc"
CONECHECK = REPO / "bin" / "conecheck"
DL_PATH = REPO / "tests" / "corpus" / "CONE_CORPUS" / "cc_growing_sibling.dl"
FIXTURES_ROOT = REPO / "fixtures" / "cone_corpus" / "cc_growing_sibling"
MEASUREMENTS = REPO / "measurements" / "punch-list-2" / "item2-growing-sibling"
SCALE_POINTS = [20, 50, 100]
SEED_BASE = 20260827600  # fresh seed base, distinct from CONE_CORPUS/SCALE_POINTS.json's

PREDS = ["p", "q", "s", "tc", "out", "out2"]
DECLINED = {"p", "q", "s"}


def _limit_mem():
    import resource
    resource.setrlimit(resource.RLIMIT_AS, (8 * 1024 * 1024 * 1024,) * 2)


def run_souffle(dl_path, facts_dir, workdir, magic=False):
    workdir.mkdir(parents=True, exist_ok=True)
    cmd = ["souffle", "-F", str(facts_dir.resolve()), "-D", str(workdir.resolve()), "-p", "prof.log"]
    if magic:
        cmd.append("--magic-transform=*")
    cmd.append(str(dl_path.resolve()))
    (workdir / "cmd.txt").write_text(" ".join(cmd) + "\n")
    proc = subprocess.run(cmd, cwd=str(workdir), capture_output=True, text=True, timeout=300, preexec_fn=_limit_mem)
    (workdir / "stdout.txt").write_text(proc.stdout)
    (workdir / "stderr.txt").write_text(proc.stderr)
    return "ok" if proc.returncode == 0 else f"error:{proc.returncode}"


def dlc_emit_guarded(workdir):
    proc = subprocess.run([str(DLC), "emit", str(DL_PATH), "--transformer=guarded"],
                           capture_output=True, encoding="utf-8", errors="replace")
    workdir.mkdir(parents=True, exist_ok=True)
    (workdir / "emit_stdout.json").write_text(proc.stdout)
    doc = json.loads(proc.stdout.strip().splitlines()[-1])
    if doc.get("status") != "ok":
        return None
    path = workdir / "transformed.dl"
    path.write_text(doc["printed"])
    return path


def sorted_lines(p):
    return sorted(p.read_text().splitlines()) if p.is_file() else []


def main():
    results = []
    for n in SCALE_POINTS:
        seed = SEED_BASE + n
        facts = gen_growing_sibling_facts(seed, n=n)
        facts_dir = FIXTURES_ROOT / f"n{n}"
        for rel in ["base", "e", "blocked", "sibling_edge"]:
            write_facts(facts_dir / f"{rel}.facts", facts[rel])

        base = MEASUREMENTS / f"n{n}"
        row = {"n": n}

        row["status_none"] = run_souffle(DL_PATH, facts_dir, base / "none")
        prof_none = tuple_analyze(base / "none" / "prof.log")
        row["T_none"] = prof_none["T_excl_copy"]
        ans_none = {a: sorted_lines(base / "none" / f"{a}.csv") for a in ["out", "out2"]}

        row["status_souffle"] = run_souffle(DL_PATH, facts_dir, base / "souffle", magic=True)
        row["T_souffle"] = tuple_analyze(base / "souffle" / "prof.log")["T_excl_copy"]

        guarded_path = dlc_emit_guarded(base / "emit_guarded")
        row["status_dlc"] = run_souffle(guarded_path, facts_dir, base / "dlc")
        prof_dlc = tuple_analyze(base / "dlc" / "prof.log")
        row["T_guarded"] = prof_dlc["T_excl_copy"]
        ans_dlc = {a: sorted_lines(base / "dlc" / f"{a}.csv") for a in ["out", "out2"]}
        row["answers_identical"] = all(ans_none[a] == ans_dlc[a] for a in ["out", "out2"])

        declined_total = 0
        transformed_total = 0
        for name, info in prof_dlc["per_relation"].items():
            if info["is_input"] or info["is_copy"]:
                continue
            origin = origin_of(name, sorted(PREDS, key=len, reverse=True))
            if origin in DECLINED:
                declined_total += info["total"]
            else:
                transformed_total += info["total"]
        row["declined_portion"] = declined_total
        row["transformed_portion"] = transformed_total
        row["ratio_T_none_over_T_guarded"] = row["T_none"] / row["T_guarded"] if row["T_guarded"] else None

        # Cone cross-check (expected empty, per conecheck already run).
        heads, all_edges, scc_of, scc_members = cone_analyze_file(DL_PATH.read_text())
        py_cone = cone_size(heads, all_edges, [sorted(DECLINED)])
        row["cone_cross_check_empty"] = (py_cone["cone_relations"] == [])

        results.append(row)
        print(json.dumps(row))

    MEASUREMENTS.mkdir(parents=True, exist_ok=True)
    (MEASUREMENTS / "summary.json").write_text(json.dumps(results, indent=2))


if __name__ == "__main__":
    main()
