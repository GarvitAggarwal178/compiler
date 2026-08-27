#!/usr/bin/env python3
"""
M3.5 -- the headline run. Full pipeline over the 5 BENCHMARK_FAMILY shapes
at every SCALE_POINTS.json point, plus every CULPRIT_CANDIDATES program,
using the T2 protocol: dlc emit --transformer=guarded, Souffle evaluates.
Also runs --transformer=magicset (ungated) at the same points for the
guard-firing table, and records T_none/T_souffle via plain Souffle runs
(with/without --magic-transform=*).
"""
import json
import resource
import subprocess
import sys
from pathlib import Path

sys.path.insert(0, "/root/compiler/harness")
from tuple_report import analyze as tuple_analyze  # noqa: E402
from cone_metric import analyze_file as parse_dep_graph, cone_size  # noqa: E402

REPO = Path("/root/compiler")
DLC = REPO / "bin" / "dlc"
FAMILY_DIR = REPO / "tests" / "corpus" / "BENCHMARK_FAMILY"
SCALE_POINTS = json.loads((FAMILY_DIR / "SCALE_POINTS.json").read_text())
FIXTURES_ROOT = REPO / "fixtures" / "benchmark-family"
CANDIDATES_DIR = REPO / "tests" / "corpus" / "CULPRIT_CANDIDATES"
CANDIDATE_FIXTURES_ROOT = REPO / "fixtures" / "culprit_candidates"
MEASUREMENTS = REPO / "measurements" / "m3-5-headline-m4"

MEM_LIMIT_BYTES = 8 * 1024 * 1024 * 1024
TIMEOUT_S = 300

SHAPES = {
    "same_generation_negation": {
        "dl": "same_generation_negation.dl", "answer": "q_notsg",
        "points": [(f"d{pt['depth']}_b{pt['branching']}",
                    FIXTURES_ROOT / "same_generation_negation" / f"d{pt['depth']}_b{pt['branching']}")
                   for pt in SCALE_POINTS["same_generation_negation"]["points"]],
    },
    "transitive_closure_bound": {
        "dl": "transitive_closure_bound.dl", "answer": "q_tc",
        "points": [(f"n{pt['n']}", FIXTURES_ROOT / "transitive_closure_bound" / f"n{pt['n']}")
                   for pt in SCALE_POINTS["transitive_closure_bound"]["points"]],
    },
    "ancestor_nonancestor": {
        "dl": "ancestor_nonancestor.dl", "answer": "q_nonancestor",
        "points": [(f"n{pt['n']}", FIXTURES_ROOT / "ancestor_nonancestor" / f"n{pt['n']}")
                   for pt in SCALE_POINTS["ancestor_nonancestor"]["points"]],
    },
    "culprit_cycle": {
        "dl": "culprit_cycle.dl", "answer": "out",
        "points": [(f"n{pt['n']}", FIXTURES_ROOT / "culprit_cycle" / f"n{pt['n']}")
                   for pt in SCALE_POINTS["culprit_cycle"]["points"]],
    },
    "reachability_complement": {
        "dl": "reachability_complement.dl", "answer": "q_unreach",
        "points": [(f"n{n}", REPO / "fixtures" / f"p2-scale-{n}")
                   for n in SCALE_POINTS["reachability_complement"]["points_already_run"]],
    },
}

CANDIDATE_CASES = {
    "cc_arity3_twobound": "out", "cc_edb_negated": "out", "cc_longer_cycle": "out",
    "cc_neg_early": "out", "cc_query_bothbound": "out", "cc_third_relation": "out",
    "cc_mixed_fallback": "ans",
}


def _limit_mem():
    resource.setrlimit(resource.RLIMIT_AS, (MEM_LIMIT_BYTES, MEM_LIMIT_BYTES))


def run_souffle(dl_path, facts_dir, workdir, magic=False, log_name="prof.log"):
    workdir.mkdir(parents=True, exist_ok=True)
    cmd = ["souffle", "-F", str(facts_dir.resolve()), "-D", str(workdir.resolve()), "-p", log_name]
    if magic:
        cmd.append("--magic-transform=*")
    cmd.append(str(dl_path.resolve()))
    (workdir / "cmd.txt").write_text(" ".join(cmd) + "\n")
    try:
        proc = subprocess.run(cmd, cwd=str(workdir), capture_output=True, text=True,
                               timeout=TIMEOUT_S, preexec_fn=_limit_mem)
    except subprocess.TimeoutExpired:
        return {"status": f"DNF:timeout-{TIMEOUT_S}s"}
    except MemoryError:
        return {"status": "DNF:memcap-8gb"}
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


def run_point(dl_path, facts_dir, answer_rel, base):
    row = {}

    # T_none: plain Souffle, no transform.
    wd_none = base / "none"
    r_none = run_souffle(dl_path, facts_dir, wd_none, magic=False)
    row["status_none"] = r_none["status"]
    if r_none["status"] == "ok":
        prof_none = tuple_analyze(wd_none / "prof.log")
        row["T_none"] = prof_none["T_excl_copy"]
        row["T_none_excl_sup"] = prof_none["T_excl_copy_excl_sup"]
        row["ans_none"] = sorted_lines(wd_none / f"{answer_rel}.csv")

    # T_souffle: Souffle's own automatic transform.
    wd_souffle = base / "souffle"
    r_souffle = run_souffle(dl_path, facts_dir, wd_souffle, magic=True)
    row["status_souffle"] = r_souffle["status"]
    if r_souffle["status"] == "ok":
        prof_souffle = tuple_analyze(wd_souffle / "prof.log")
        row["T_souffle"] = prof_souffle["T_excl_copy"]
        row["T_souffle_excl_sup"] = prof_souffle["T_excl_copy_excl_sup"]

    # T_dlc: dlc emit --transformer=guarded, Souffle evaluates (T2 protocol).
    guarded_path, guarded_doc = dlc_emit(dl_path, "guarded", base / "emit_guarded")
    if guarded_path is None:
        row["status_dlc"] = "emit_error"
        row["dlc_emit_doc"] = guarded_doc
        return row
    wd_dlc = base / "dlc"
    r_dlc = run_souffle(guarded_path, facts_dir, wd_dlc, magic=False)
    row["status_dlc"] = r_dlc["status"]
    if r_dlc["status"] == "ok":
        prof_dlc = tuple_analyze(wd_dlc / "prof.log")
        # M4-SIPS.md section 5: report both supplementary-counting
        # conventions. T_dlc (this key, unchanged name/meaning) is
        # "incl-sup" -- dlc's existing default, every relation counted;
        # T_dlc_excl_sup additionally excludes every sup_*-named
        # checkpoint relation dlc's magic-set transform materializes and
        # Soufflé's own transform has no equivalent of.
        row["T_dlc"] = prof_dlc["T_excl_copy"]
        row["T_dlc_excl_sup"] = prof_dlc["T_excl_copy_excl_sup"]
        row["ans_dlc"] = sorted_lines(wd_dlc / f"{answer_rel}.csv")
        if "ans_none" in row:
            row["answers_identical"] = row["ans_none"] == row["ans_dlc"]
            del row["ans_none"]
            del row["ans_dlc"]

    # Guard-firing: does the guard's output differ from the ungated
    # magicset transform? (byte-diff on printed text is a reasonable
    # proxy -- identical output means the guard declined nothing.)
    magicset_path, _ = dlc_emit(dl_path, "magicset", base / "emit_magicset")
    if magicset_path is not None:
        row["guard_fired"] = (guarded_path.read_text() != magicset_path.read_text())

    return row


def main():
    results = {}

    for shape, spec in SHAPES.items():
        print(f"===== shape={shape} =====", file=sys.stderr)
        dl_path = FAMILY_DIR / spec["dl"]
        shape_rows = []
        for tag, facts_dir in spec["points"]:
            print(f"  --- {tag} ---", file=sys.stderr)
            base = MEASUREMENTS / "family" / shape / tag
            row = run_point(dl_path, facts_dir, spec["answer"], base)
            row["tag"] = tag
            shape_rows.append(row)
            print(f"    {json.dumps({k: v for k, v in row.items() if k != 'dlc_emit_doc'})}", file=sys.stderr)
            if row.get("status_none") != "ok" or row.get("status_dlc") != "ok":
                print(f"    {shape}/{tag}: DNF/error, stopping ascent for this shape", file=sys.stderr)
                break
        results[shape] = shape_rows

    candidate_rows = {}
    for name, answer_rel in CANDIDATE_CASES.items():
        dl_path = CANDIDATES_DIR / f"{name}.dl"
        facts_dir = CANDIDATE_FIXTURES_ROOT / name
        base = MEASUREMENTS / "candidates" / name
        row = run_point(dl_path, facts_dir, answer_rel, base)
        candidate_rows[name] = row
        print(f"CANDIDATE {name}: {json.dumps({k: v for k, v in row.items() if k != 'dlc_emit_doc'})}", file=sys.stderr)
    results["_candidates"] = candidate_rows

    MEASUREMENTS.mkdir(parents=True, exist_ok=True)
    (MEASUREMENTS / "summary.json").write_text(json.dumps(results, indent=2))
    print(json.dumps(results, indent=2))


if __name__ == "__main__":
    main()
