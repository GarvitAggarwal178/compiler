#!/usr/bin/env python3
"""
NIGHT-BATCH-03 T3b: semantic acceptance harness for any candidate transform.

Input: an original .dl, a candidate transformed .dl, a fixture directory.
Works on ANY candidate regardless of naming convention (Souffle's, dlc's, a
hand-written guard) because the check is purely semantic: run both programs
standalone against the same facts, compare answer relations by sorted set
equality, and report exact tuple totals via harness/tuple_report.py's
Souffle-profile parser.

Deliberately does NOT build a syntactic diff against tests/reference/
souffle-transformed/ -- Souffle's own naming/inlining makes that worthless
(NIGHT-BATCH-03.md T3a).

The ratio this script reports is plainly T_original/T_candidate, labeled as
exactly that -- it is the caller's job, not this script's, to know whether
`original` is a plain baseline (T_none) or an already magic-transformed
dump (T_souffle), and to name the ratio accordingly in a report. This
script never calls its own output "contribution" or assumes which
three-column slot either file fills.
"""
import argparse
import json
import re
import resource
import subprocess
import sys
from pathlib import Path

sys.path.insert(0, "/root/compiler/harness")
from tuple_report import analyze as tuple_analyze  # noqa: E402

MEM_LIMIT_BYTES = 8 * 1024 * 1024 * 1024
TIMEOUT_S = 300


def _limit_mem():
    resource.setrlimit(resource.RLIMIT_AS, (MEM_LIMIT_BYTES, MEM_LIMIT_BYTES))


def output_names(dl_path: Path):
    text = dl_path.read_text(errors="replace")
    return sorted(set(re.findall(r"^\s*\.output\s+(\w+)\s*(?:\(\s*\))?\s*$", text, re.MULTILINE)))


def run_souffle(dl_path: Path, facts_dir: Path, workdir: Path):
    workdir.mkdir(parents=True, exist_ok=True)
    cmd = ["souffle", "-F", str(facts_dir.resolve()), "-D", str(workdir.resolve()),
           "-p", "prof.log", str(dl_path.resolve())]
    try:
        proc = subprocess.run(cmd, cwd=str(workdir), capture_output=True, text=True,
                               timeout=TIMEOUT_S, preexec_fn=_limit_mem)
    except subprocess.TimeoutExpired:
        return {"status": f"DNF:timeout-{TIMEOUT_S}s"}
    except MemoryError:
        return {"status": "DNF:memcap-8gb"}
    if proc.returncode != 0:
        return {"status": f"error:returncode-{proc.returncode}", "stderr": proc.stderr[:1000]}
    return {"status": "ok"}


def sorted_lines(path: Path):
    return sorted(path.read_text().splitlines()) if path.is_file() else []


def accept(original: Path, candidate: Path, facts_dir: Path, workdir: Path):
    names = sorted(set(output_names(original)) | set(output_names(candidate)))
    wd_orig, wd_cand = workdir / "original", workdir / "candidate"
    r_orig = run_souffle(original, facts_dir, wd_orig)
    r_cand = run_souffle(candidate, facts_dir, wd_cand)

    result = {
        "original": str(original), "candidate": str(candidate), "facts_dir": str(facts_dir),
        "status_original": r_orig["status"], "status_candidate": r_cand["status"],
        "output_relations": names,
    }
    if r_orig["status"] != "ok" or r_cand["status"] != "ok":
        result["comparable"] = False
        if r_orig["status"] != "ok":
            result["original_stderr"] = r_orig.get("stderr", "")
        if r_cand["status"] != "ok":
            result["candidate_stderr"] = r_cand.get("stderr", "")
        return result

    profile_orig = tuple_analyze(wd_orig / "prof.log")
    profile_cand = tuple_analyze(wd_cand / "prof.log")

    per_relation_answer_match = {}
    diffs = {}
    for n in names:
        a = sorted_lines(wd_orig / f"{n}.csv")
        b = sorted_lines(wd_cand / f"{n}.csv")
        per_relation_answer_match[n] = (a == b)
        if a != b:
            diffs[n] = {"original_only": sorted(set(a) - set(b))[:20],
                        "candidate_only": sorted(set(b) - set(a))[:20]}

    result.update({
        "comparable": True,
        "answers_identical": all(per_relation_answer_match.values()) if names else True,
        "per_relation_answer_match": per_relation_answer_match,
        "diffs": diffs,
        "T_original": profile_orig["T_excl_copy"],
        "T_candidate": profile_cand["T_excl_copy"],
        "ratio_original_over_candidate": (
            profile_orig["T_excl_copy"] / profile_cand["T_excl_copy"]
            if profile_cand["T_excl_copy"] else None
        ),
        "per_relation_original": profile_orig["per_relation"],
        "per_relation_candidate": profile_cand["per_relation"],
    })
    return result


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("original", type=Path)
    ap.add_argument("candidate", type=Path)
    ap.add_argument("facts_dir", type=Path)
    ap.add_argument("--workdir", type=Path, default=None)
    args = ap.parse_args()

    import tempfile
    workdir = args.workdir or Path(tempfile.mkdtemp(prefix="m2accept_"))
    result = accept(args.original, args.candidate, args.facts_dir, workdir)
    print(json.dumps(result, indent=2, sort_keys=True))


if __name__ == "__main__":
    main()
