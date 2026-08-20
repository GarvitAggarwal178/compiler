#!/usr/bin/env python3
"""
NIGHT-BATCH-01 T3: recoverable-envelope sweep over the pre-registered
corpus. For each program T2 marked ok+seedable, run --magic-transform=*
(T2 already has the untransformed T_none) and record E_recoverable = sum
of every @neglabel.-prefixed relation's tuple count -- exactly what
Soufflé forfeits by isolating negated relations instead of restricting
them, mechanical, no hand-transform, no program selection.

Abort condition: any program whose answer relations differ between
untransformed and magic-transform configurations. That is a Soufflé
soundness bug, not something to investigate further tonight -- log to
ESCALATIONS.md and abort the whole task per NIGHT-BATCH-01 protocol.
"""
import json
import resource
import subprocess
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from tuple_report import analyze as tuple_analyze  # noqa: E402

REPO = Path(__file__).resolve().parents[1]
SOUFFLE_TESTS = Path("/root/souffle-src/tests")
T2_SUMMARY = REPO / "measurements" / "night01-t2" / "summary.json"
OUT_DIR = REPO / "measurements" / "night01-t3"

MEM_LIMIT_BYTES = 8 * 1024 * 1024 * 1024
TIMEOUT_S = 300


def _limit_mem():
    resource.setrlimit(resource.RLIMIT_AS, (MEM_LIMIT_BYTES, MEM_LIMIT_BYTES))


def run_magic(test_rel_path: str, dl_name: str, output_names):
    test_dir = SOUFFLE_TESTS / test_rel_path
    dl_path = test_dir / dl_name
    facts_dir = test_dir / "facts" if (test_dir / "facts").is_dir() else test_dir

    safe_name = test_rel_path.replace("/", "__")
    workdir = OUT_DIR / safe_name
    workdir.mkdir(parents=True, exist_ok=True)
    log_name = "prof.log"

    cmd = ["souffle", "-F", str(facts_dir), "-D", str(workdir), "-p", log_name,
           "--magic-transform=*", str(dl_path)]
    (workdir / "cmd.txt").write_text(" ".join(cmd) + "\n")

    try:
        proc = subprocess.run(
            cmd, cwd=str(workdir), capture_output=True, text=True,
            timeout=TIMEOUT_S, preexec_fn=_limit_mem,
        )
    except subprocess.TimeoutExpired:
        return {"path": test_rel_path, "status": "DNF:timeout-300s"}
    except MemoryError:
        return {"path": test_rel_path, "status": "DNF:memcap-8gb"}

    (workdir / "stdout.txt").write_text(proc.stdout)
    (workdir / "stderr.txt").write_text(proc.stderr)

    if proc.returncode != 0:
        result = {"path": test_rel_path, "status": f"error:returncode-{proc.returncode}",
                  "stderr_snippet": proc.stderr[:500]}
        (workdir / "meta.json").write_text(json.dumps(result, indent=2))
        return result

    log_path = workdir / log_name
    tr = tuple_analyze(log_path) if log_path.is_file() else None

    # answer-relation identity check against T2's untransformed CSVs.
    # CLAUDE.md section 6: "Comparison is set equality on output
    # relations, not text diff. Sort, then compare." -- a raw byte
    # comparison is wrong here because Soufflé does not guarantee row
    # order is stable across different execution plans (untransformed vs
    # magic-transformed can legitimately emit the same set of tuples in
    # different orders).
    t2_workdir = REPO / "measurements" / "night01-t2" / safe_name
    answers_identical = True
    diffs = []
    for rel in output_names:
        f_none = t2_workdir / f"{rel}.csv"
        f_souffle = workdir / f"{rel}.csv"
        if not f_none.is_file() or not f_souffle.is_file():
            diffs.append(f"{rel}: missing csv (none={f_none.is_file()}, souffle={f_souffle.is_file()})")
            answers_identical = False
            continue
        lines_none = sorted(f_none.read_text().splitlines())
        lines_souffle = sorted(f_souffle.read_text().splitlines())
        if lines_none != lines_souffle:
            diffs.append(f"{rel}: set diff (sorted content differs, not just order)")
            answers_identical = False

    result = {
        "path": test_rel_path,
        "status": "ok",
        "T_souffle_excl_copy": tr["T_excl_copy"] if tr else None,
        "T_souffle_incl_copy": tr["T_incl_copy"] if tr else None,
        "E_recoverable": tr["E_recoverable"] if tr else None,
        "neglabel_relations": tr["neglabel_relations"] if tr else {},
        "answers_identical_to_untransformed": answers_identical,
        "answer_diffs": diffs,
    }
    (workdir / "meta.json").write_text(json.dumps(result, indent=2))
    return result


def main():
    t2 = json.loads(T2_SUMMARY.read_text())
    candidates = [x for x in t2 if x["status"] == "ok" and x["seedable"]]
    print(f"candidates (ok+seedable): {len(candidates)}", file=sys.stderr)

    OUT_DIR.mkdir(parents=True, exist_ok=True)
    results = []
    aborted = False
    for x in candidates:
        r = run_magic(x["path"], x["dl_file"], x["output_names"])
        r["T_none_excl_copy"] = x["T_none_excl_copy"]
        r["T_none_incl_copy"] = x["T_none_incl_copy"]
        print(f"{x['path']}: {r['status']}"
              + (f" answers_identical={r.get('answers_identical_to_untransformed')}" if r["status"] == "ok" else ""),
              file=sys.stderr)
        results.append(r)
        if r["status"] == "ok" and not r["answers_identical_to_untransformed"]:
            print(f"ABORT CONDITION: {x['path']} answer relations diverged: {r['answer_diffs']}", file=sys.stderr)
            aborted = True
            break

    (OUT_DIR / "summary.json").write_text(json.dumps(results, indent=2))
    print(json.dumps({"aborted": aborted, "results": results}, indent=2))


if __name__ == "__main__":
    main()
