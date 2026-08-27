#!/usr/bin/env python3
"""
M2-M3-BUILD.md section 5's required counterexample search: the 4
tests/programs/p6*_base.dl constructions plus every program in
tests/corpus/CULPRIT_CANDIDATES/, dlc's real transform vs untransformed,
answers compared via harness/m2_accept.py. A disagreement here is the
counterexample five bounded (Phase 0.6) attempts failed to find --
reported loudly, not patched around.
"""
import json
import subprocess
import sys
from pathlib import Path

sys.path.insert(0, "/root/compiler/harness")
from m2_accept import accept  # noqa: E402

REPO = Path("/root/compiler")
DLC = REPO / "bin" / "dlc"
MEASUREMENTS = REPO / "measurements" / "m3-1-counterexample"

P6_FILES = sorted((REPO / "tests" / "programs").glob("p6*_base.dl"))
CANDIDATE_FILES = sorted((REPO / "tests" / "corpus" / "CULPRIT_CANDIDATES").glob("*.dl"))

P6_FIXTURES = REPO / "fixtures" / "p2"
CANDIDATE_FIXTURES_ROOT = REPO / "fixtures" / "culprit_candidates"


def dlc_emit(dl_path, workdir):
    proc = subprocess.run([str(DLC), "emit", str(dl_path), "--transformer=magicset"],
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


def run_one(name, dl_path, facts_dir):
    base = MEASUREMENTS / name
    transformed, doc = dlc_emit(dl_path, base / "emit")
    if transformed is None:
        return {"name": name, "status": "emit_error", "doc": doc}

    # Printing always reformats (full parenthesization, stripped
    # comments, canonical spacing), so byte-equality against the raw
    # source is not a valid no-op test even for a genuine pass-through.
    # A real transform always introduces "magic_"/"sup_" relations;
    # their absence is the reliable signal instead.
    no_op = "magic_" not in doc["printed"] and "sup_" not in doc["printed"]
    r = accept(dl_path, transformed, facts_dir, base / "accept")
    row = {
        "name": name,
        "no_op_passthrough": no_op,
        "comparable": r.get("comparable"),
        "answers_identical": r.get("answers_identical"),
        "T_none": r.get("T_original"), "T_dlc": r.get("T_candidate"),
    }
    if not r.get("comparable"):
        # m2_accept.py's own return value truncates stderr to 1000 chars;
        # read the full file it wrote to disk instead, since "Unable to
        # stratify" can appear after that truncation point.
        stderr_path = base / "accept" / "candidate" / "stderr.txt"
        full_candidate_stderr = stderr_path.read_text() if stderr_path.is_file() else r.get("candidate_stderr", "")
        row["unstratifiable"] = "Unable to stratify" in full_candidate_stderr
        row["candidate_stderr"] = full_candidate_stderr[-800:]
        row["original_stderr"] = r.get("original_stderr", "")[:500]
    return row


def main():
    results = []
    for f in P6_FILES:
        results.append(run_one(f.stem, f, P6_FIXTURES))
    for f in CANDIDATE_FILES:
        results.append(run_one(f.stem, f, CANDIDATE_FIXTURES_ROOT / f.stem))

    attempted = len(results)
    agreed = sum(1 for r in results if r.get("answers_identical") is True)
    unstratifiable = sum(1 for r in results if r.get("unstratifiable"))
    disagreements = [r for r in results if r.get("answers_identical") is False]

    summary = {
        "attempted": attempted, "agreed": agreed,
        "agreed_of_attempted": f"{agreed}/{attempted}",
        "unstratifiable": unstratifiable,
        "disagreements": disagreements,
        "results": results,
    }
    MEASUREMENTS.mkdir(parents=True, exist_ok=True)
    (MEASUREMENTS / "summary.json").write_text(json.dumps(summary, indent=2))
    print(json.dumps({k: v for k, v in summary.items() if k != "results"}, indent=2))
    for r in results:
        print(json.dumps(r, indent=2)[:600], file=sys.stderr)


if __name__ == "__main__":
    main()
