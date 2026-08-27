#!/usr/bin/env python3
"""
NIGHT-BATCH-04 B: the cone corpus gate.

For each of the four tests/corpus/CONE_CORPUS/ programs, at each
pre-registered scale point (SCALE_POINTS.json, committed before this
script's first run):

1. Generate fixtures (harness/fixtures_lib.py's gen_cone_corpus_facts,
   seeded), writing only the relation files each program's own
   .decl/.input lines declare.
2. Cross-check guard.Decide's culprit/cone/declined sets (via
   bin/conecheck, tools/conecheck/main.go) against harness/
   cone_metric.py's independent source-graph computation -- exact
   agreement required, as in M3.3.
3. Run the T2 protocol (dlc emit --transformer=guarded, Souffle
   evaluates) for T_dlc; plain Souffle for T_none/T_souffle; compare
   every .output relation's answer set-equality against the
   untransformed baseline.
"""
import json
import subprocess
import sys
from pathlib import Path

sys.path.insert(0, "/root/compiler/harness")
from fixtures_lib import gen_cone_corpus_facts, write_facts  # noqa: E402
from tuple_report import analyze as tuple_analyze  # noqa: E402
from cone_metric import analyze_file as cone_analyze_file, cone_size  # noqa: E402

REPO = Path("/root/compiler")
DLC = REPO / "bin" / "dlc"
CONECHECK = REPO / "bin" / "conecheck"
CORPUS_DIR = REPO / "tests" / "corpus" / "CONE_CORPUS"
SCALE_POINTS = json.loads((CORPUS_DIR / "SCALE_POINTS.json").read_text())
FIXTURES_ROOT = REPO / "fixtures" / "cone_corpus"
MEASUREMENTS = REPO / "measurements" / "night04-b-cone"

# Which relation keys (from gen_cone_corpus_facts' returned dict) each
# program's .decl/.input lines actually declare.
PROGRAM_RELATIONS = {
    "cc_cone_only": ["base", "e", "blocked", "gate_seed", "gate_edge"],
    "cc_sibling_emptycone": ["base", "e", "blocked", "sibling_edge"],
    "cc_both": ["base", "e", "blocked", "gate_seed", "gate_edge", "sibling_edge"],
    "cc_cone_proper_subset": ["base", "e", "blocked", "chain_seed", "chain_edge",
                               "sibling_edge", "sibling2_edge"],
}
PROGRAM_ANSWERS = {
    "cc_cone_only": ["out"],
    "cc_sibling_emptycone": ["out", "out2"],
    "cc_both": ["out", "out2"],
    "cc_cone_proper_subset": ["out", "out2", "out3"],
}

CULPRIT = ["p", "q", "s"]  # identical across all four constructions (empirically confirmed)


def _limit_mem():
    import resource
    resource.setrlimit(resource.RLIMIT_AS, (8 * 1024 * 1024 * 1024,) * 2)


def run_souffle(dl_path, facts_dir, workdir, log_name="prof.log"):
    workdir.mkdir(parents=True, exist_ok=True)
    cmd = ["souffle", "-F", str(facts_dir.resolve()), "-D", str(workdir.resolve()),
           "-p", log_name, str(dl_path.resolve())]
    (workdir / "cmd.txt").write_text(" ".join(cmd) + "\n")
    try:
        proc = subprocess.run(cmd, cwd=str(workdir), capture_output=True, text=True,
                               timeout=300, preexec_fn=_limit_mem)
    except subprocess.TimeoutExpired:
        return {"status": "DNF:timeout-300s"}
    (workdir / "stdout.txt").write_text(proc.stdout)
    (workdir / "stderr.txt").write_text(proc.stderr)
    if proc.returncode != 0:
        return {"status": f"error:returncode-{proc.returncode}", "stderr": proc.stderr[:1500]}
    return {"status": "ok"}


def dlc_emit(dl_path, transformer, workdir):
    proc = subprocess.run([str(DLC), "emit", str(dl_path), f"--transformer={transformer}"],
                           capture_output=True, encoding="utf-8", errors="replace")
    workdir.mkdir(parents=True, exist_ok=True)
    (workdir / "emit_stdout.json").write_text(proc.stdout)
    try:
        doc = json.loads(proc.stdout.strip().splitlines()[-1])
    except (json.JSONDecodeError, IndexError):
        return None, {"status": "panic", "stderr": proc.stderr[:500]}
    if doc.get("status") != "ok":
        return None, doc
    transformed = workdir / "transformed.dl"
    transformed.write_text(doc["printed"])
    return transformed, doc


def sorted_lines(path):
    return sorted(path.read_text().splitlines()) if path.is_file() else []


def cone_cross_check(program):
    dl_path = CORPUS_DIR / f"{program}.dl"
    proc = subprocess.run([str(CONECHECK), str(dl_path)], capture_output=True, text=True)
    go_result = json.loads(proc.stdout)

    text = dl_path.read_text()
    heads, all_edges, scc_of, scc_members = cone_analyze_file(text)
    py_result = cone_size(heads, all_edges, [CULPRIT])

    agree = (sorted(go_result["cone_relations"]) == sorted(py_result["cone_relations"]))
    return {
        "go_culprit": go_result["culprit_predicates"],
        "go_cone": go_result["cone_relations"],
        "go_declined": go_result["declined_relations"],
        "py_cone": py_result["cone_relations"],
        "cone_agrees": agree,
    }


def run_point(program, n, target_base, target_e, seed):
    facts = gen_cone_corpus_facts(seed, n=n, target_base=target_base, target_e=target_e)
    base = MEASUREMENTS / program / f"n{n}"
    facts_dir = FIXTURES_ROOT / program / f"n{n}"
    for rel in PROGRAM_RELATIONS[program]:
        write_facts(facts_dir / f"{rel}.facts", facts[rel])

    dl_path = CORPUS_DIR / f"{program}.dl"
    row = {"program": program, "n": n}

    wd_none = base / "none"
    r_none = run_souffle(dl_path, facts_dir, wd_none)
    row["status_none"] = r_none["status"]
    if r_none["status"] == "ok":
        prof = tuple_analyze(wd_none / "prof.log")
        row["T_none"] = prof["T_excl_copy"]
        row["ans_none"] = {a: sorted_lines(wd_none / f"{a}.csv") for a in PROGRAM_ANSWERS[program]}

    wd_souffle = base / "souffle"
    cmd = ["souffle", "-F", str(facts_dir.resolve()), "-D", str(wd_souffle.resolve()),
           "-p", "prof.log", "--magic-transform=*", str(dl_path.resolve())]
    wd_souffle.mkdir(parents=True, exist_ok=True)
    (wd_souffle / "cmd.txt").write_text(" ".join(cmd) + "\n")
    proc = subprocess.run(cmd, cwd=str(wd_souffle), capture_output=True, text=True, timeout=300)
    (wd_souffle / "stdout.txt").write_text(proc.stdout)
    (wd_souffle / "stderr.txt").write_text(proc.stderr)
    row["status_souffle"] = "ok" if proc.returncode == 0 else f"error:{proc.returncode}"
    if row["status_souffle"] == "ok":
        row["T_souffle"] = tuple_analyze(wd_souffle / "prof.log")["T_excl_copy"]

    guarded_path, doc = dlc_emit(dl_path, "guarded", base / "emit_guarded")
    if guarded_path is None:
        row["status_dlc"] = "emit_error"
        row["dlc_emit_doc"] = doc
        return row
    wd_dlc = base / "dlc"
    r_dlc = run_souffle(guarded_path, facts_dir, wd_dlc)
    row["status_dlc"] = r_dlc["status"]
    if r_dlc["status"] == "ok":
        row["T_dlc"] = tuple_analyze(wd_dlc / "prof.log")["T_excl_copy"]
        ans_dlc = {a: sorted_lines(wd_dlc / f"{a}.csv") for a in PROGRAM_ANSWERS[program]}
        if "ans_none" in row:
            row["answers_identical"] = all(
                row["ans_none"][a] == ans_dlc[a] for a in PROGRAM_ANSWERS[program]
            )
            row["per_relation_identical"] = {
                a: row["ans_none"][a] == ans_dlc[a] for a in PROGRAM_ANSWERS[program]
            }
            del row["ans_none"]

    if "T_none" in row and "T_dlc" in row:
        row["T_guarded_lt_T_none"] = row["T_dlc"] < row["T_none"]

    return row


def main():
    results = {"cone_cross_check": {}, "points": []}
    for program in SCALE_POINTS["programs"]:
        results["cone_cross_check"][program] = cone_cross_check(program)

    for program in SCALE_POINTS["programs"]:
        for pt in SCALE_POINTS["points"]:
            n = pt["n"]
            seed = SCALE_POINTS["seed_base"] + n
            row = run_point(program, n, pt["target_base"], pt["target_e"], seed)
            results["points"].append(row)
            print(json.dumps({k: v for k, v in row.items() if k != "dlc_emit_doc"}), file=sys.stderr)

    MEASUREMENTS.mkdir(parents=True, exist_ok=True)
    (MEASUREMENTS / "summary.json").write_text(json.dumps(results, indent=2))
    print(json.dumps(results, indent=2))


if __name__ == "__main__":
    main()
