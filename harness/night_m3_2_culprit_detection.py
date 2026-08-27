#!/usr/bin/env python3
"""
M2-M3-BUILD.md section 6's required gate: dlc-verdict vs Soufflé-verdict
on every program in tests/corpus/CULPRIT_CANDIDATES/ plus culprit_cycle.dl
(7 programs). This IS a differential-oracle cross-check, not
self-consistency testing -- Soufflé's own stratification checker was
written by someone else and is asked the SAME question (does the
candidate TRANSFORMED program stratify) about the SAME transformed
program dlc produced.
"""
import json
import subprocess
import sys
from pathlib import Path

REPO = Path("/root/compiler")
DLC = REPO / "bin" / "dlc"
CANDIDATES_DIR = REPO / "tests" / "corpus" / "CULPRIT_CANDIDATES"
FAMILY_DIR = REPO / "tests" / "corpus" / "BENCHMARK_FAMILY"
CANDIDATE_FIXTURES_ROOT = REPO / "fixtures" / "culprit_candidates"
FAMILY_FIXTURES = REPO / "fixtures" / "benchmark-family" / "culprit_cycle" / "n20"
MEASUREMENTS = REPO / "measurements" / "m3-2-culprit-detection"

FILES = sorted(CANDIDATES_DIR.glob("*.dl")) + [FAMILY_DIR / "culprit_cycle.dl"]


def dlc_emit(dl_path, workdir):
    proc = subprocess.run([str(DLC), "emit", str(dl_path), "--transformer=magicset"],
                           capture_output=True, encoding="utf-8", errors="replace")
    workdir.mkdir(parents=True, exist_ok=True)
    doc = json.loads(proc.stdout.strip().splitlines()[-1])
    transformed = workdir / "transformed.dl"
    transformed.write_text(doc.get("printed", ""))
    return transformed, doc


def dlc_check_verdict(dl_path):
    proc = subprocess.run([str(DLC), "check", str(dl_path)], capture_output=True, encoding="utf-8", errors="replace")
    doc = json.loads(proc.stdout.strip().splitlines()[-1])
    if doc["status"] != "rejected":
        return "stratifiable", doc
    unstrat = any(d["category"] == "unstratifiable" for d in doc.get("diagnostics", []))
    return ("unstratifiable" if unstrat else "rejected_other_ground"), doc


def souffle_verdict(dl_path, facts_dir, workdir):
    workdir.mkdir(parents=True, exist_ok=True)
    cmd = ["souffle", "-F", str(facts_dir), "-D", str(workdir), str(dl_path)]
    proc = subprocess.run(cmd, cwd=str(workdir), capture_output=True, text=True, timeout=60)
    (workdir / "stderr.txt").write_text(proc.stderr)
    if proc.returncode == 0:
        return "stratifiable"
    if "Unable to stratify" in proc.stderr:
        return "unstratifiable"
    return "rejected_other"


def main():
    results = []
    for f in FILES:
        name = f.stem
        base = MEASUREMENTS / name
        facts = FAMILY_FIXTURES if name == "culprit_cycle" else CANDIDATE_FIXTURES_ROOT / name

        transformed, emit_doc = dlc_emit(f, base / "emit")
        if emit_doc.get("status") != "ok":
            results.append({"name": name, "status": "emit_error", "doc": emit_doc})
            continue

        dlc_verdict, check_doc = dlc_check_verdict(transformed)
        souffle_v = souffle_verdict(transformed, facts, base / "souffle")

        dlc_says_unstrat = dlc_verdict == "unstratifiable"
        souffle_says_unstrat = souffle_v == "unstratifiable"
        row = {
            "name": name,
            "dlc_verdict": dlc_verdict,
            "souffle_verdict": souffle_v,
            "agree": dlc_says_unstrat == souffle_says_unstrat,
        }
        results.append(row)
        print(json.dumps(row), file=sys.stderr)

    attempted = len(results)
    agreed = sum(1 for r in results if r.get("agree") is True)
    disagreements = [r for r in results if r.get("agree") is False]

    summary = {
        "attempted": attempted, "agreed": agreed,
        "agreed_of_attempted": f"{agreed}/{attempted}",
        "disagreements": disagreements,
        "results": results,
    }
    MEASUREMENTS.mkdir(parents=True, exist_ok=True)
    (MEASUREMENTS / "summary.json").write_text(json.dumps(summary, indent=2))
    print(json.dumps(summary, indent=2))


if __name__ == "__main__":
    main()
