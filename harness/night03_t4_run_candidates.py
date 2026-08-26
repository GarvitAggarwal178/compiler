#!/usr/bin/env python3
"""
NIGHT-BATCH-03 T4 step 4: for each of the 6 hand-constructed
CULPRIT_CANDIDATES programs, generate a seeded fixture and record Soufflé's
behaviour: untransformed result, result under --magic-transform=*, whether
any relation is skipped, and tuple totals. Does NOT adorn anything by
hand -- both configurations are run by invoking Soufflé itself.
"""
import json
import sys
from pathlib import Path

sys.path.insert(0, "/root/compiler/harness")
import fixtures_lib as fl  # noqa: E402
from tuple_report import analyze as tuple_analyze  # noqa: E402
import subprocess

REPO = Path("/root/compiler")
CANDIDATES_DIR = REPO / "tests" / "corpus" / "CULPRIT_CANDIDATES"
FIXTURES_ROOT = REPO / "fixtures" / "culprit_candidates"
MEASUREMENTS = REPO / "measurements" / "night03-t4" / "candidates"

N = 50
SEED = 20260822450  # seed_base(culprit_cycle)=20260822400 + n=50, project convention
TARGET_BASE = 75
TARGET_E = 75
BLOCKED_FRACTION = 0.2


def write_fixture(name, kind):
    fdir = FIXTURES_ROOT / name
    if kind == "std":
        base_edges, e_edges, blocked = fl.gen_culprit_cycle_facts(
            SEED, n=N, target_base=TARGET_BASE, target_e=TARGET_E, blocked_fraction=BLOCKED_FRACTION)
        fl.write_facts(fdir / "base.facts", base_edges)
        fl.write_facts(fdir / "e.facts", e_edges)
        fl.write_facts(fdir / "blocked.facts", blocked)
    elif kind == "link":
        base_edges, e_edges, blocked = fl.gen_culprit_cycle_facts(
            SEED, n=N, target_base=TARGET_BASE, target_e=TARGET_E, blocked_fraction=BLOCKED_FRACTION)
        fl.write_facts(fdir / "base.facts", base_edges)
        fl.write_facts(fdir / "link.facts", e_edges)
        fl.write_facts(fdir / "blocked.facts", blocked)
    elif kind == "link_edb":
        base_edges, e_edges, blocked = fl.gen_culprit_cycle_facts(
            SEED, n=N, target_base=TARGET_BASE, target_e=TARGET_E, blocked_fraction=BLOCKED_FRACTION)
        fl.write_facts(fdir / "base.facts", base_edges)
        fl.write_facts(fdir / "link.facts", e_edges)
        fl.write_facts(fdir / "blocked_pair.facts", blocked)
    elif kind == "labeled":
        base_edges, e_edges, blocked = fl.gen_culprit_cycle_facts_labeled(
            SEED, n=N, num_labels=3, target_base=TARGET_BASE, target_e=TARGET_E, blocked_fraction=BLOCKED_FRACTION)
        fl.write_facts(fdir / "base.facts", base_edges)
        fl.write_facts(fdir / "e.facts", e_edges)
        fl.write_facts(fdir / "blocked.facts", blocked)
    return fdir


VARIANTS = [
    ("cc_longer_cycle", "std", "out"),
    ("cc_neg_early", "link", "out"),
    ("cc_edb_negated", "link_edb", "out"),
    ("cc_arity3_twobound", "labeled", "out"),
    ("cc_query_bothbound", "std", "out"),
    ("cc_third_relation", "std", "out"),
]


def run_souffle(dl_path, facts_dir, workdir, magic):
    workdir.mkdir(parents=True, exist_ok=True)
    cmd = ["souffle", "-F", str(facts_dir.resolve()), "-D", str(workdir.resolve()), "-p", "prof.log"]
    if magic:
        cmd.append("--magic-transform=*")
    cmd.append(str(dl_path.resolve()))
    (workdir / "cmd.txt").write_text(" ".join(cmd) + "\n")
    try:
        proc = subprocess.run(cmd, cwd=str(workdir), capture_output=True, text=True, timeout=120)
    except subprocess.TimeoutExpired:
        return {"status": "DNF:timeout"}
    (workdir / "stdout.txt").write_text(proc.stdout)
    (workdir / "stderr.txt").write_text(proc.stderr)
    if proc.returncode != 0:
        return {"status": f"error:returncode-{proc.returncode}", "stderr": proc.stderr[:1500]}
    return {"status": "ok"}


def sorted_lines(path):
    return sorted(path.read_text().splitlines()) if path.is_file() else []


def main():
    results = []
    for name, kind, answer_rel in VARIANTS:
        dl_path = CANDIDATES_DIR / f"{name}.dl"
        fdir = write_fixture(name, kind)
        base = MEASUREMENTS / name
        wd_none = base / "none"
        wd_magic = base / "magic"

        r_none = run_souffle(dl_path, fdir, wd_none, magic=False)
        r_magic = run_souffle(dl_path, fdir, wd_magic, magic=True)

        row = {"name": name, "kind": kind, "status_none": r_none["status"], "status_magic": r_magic["status"]}
        if r_none["status"] != "ok":
            row["none_stderr"] = r_none.get("stderr", "")
        if r_magic["status"] != "ok":
            row["magic_stderr"] = r_magic.get("stderr", "")

        if r_none["status"] == "ok":
            profile_none = tuple_analyze(wd_none / "prof.log")
            row["T_none"] = profile_none["T_excl_copy"]
            row["ans_none"] = sorted_lines(wd_none / f"{answer_rel}.csv")
        if r_magic["status"] == "ok":
            profile_magic = tuple_analyze(wd_magic / "prof.log")
            row["T_souffle"] = profile_magic["T_excl_copy"]
            row["ans_magic"] = sorted_lines(wd_magic / f"{answer_rel}.csv")
            # Which relations got skipped (0 tuples but non-input, non-empty in "none" run) --
            # a coarse proxy for "relation is a @neglabel-style full-extent skip."
            row["neglabel_relations"] = profile_magic.get("neglabel_relations", {})

        if "ans_none" in row and "ans_magic" in row:
            row["answers_identical"] = (row["ans_none"] == row["ans_magic"])
            del row["ans_none"]
            del row["ans_magic"]
        results.append(row)
        print(json.dumps(row, indent=2)[:800], file=sys.stderr)

    MEASUREMENTS.mkdir(parents=True, exist_ok=True)
    (MEASUREMENTS / "summary.json").write_text(json.dumps(results, indent=2))
    print(json.dumps(results, indent=2))


if __name__ == "__main__":
    main()
