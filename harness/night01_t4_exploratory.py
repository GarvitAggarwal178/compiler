#!/usr/bin/env python3
"""
NIGHT-BATCH-01 T4: whole-tree exploratory sweep. EXPLORATORY -- NOT
PRE-REGISTERED, NOT REPORTABLE AS A RESULT. Structural reconnaissance
only, to inform a future human corpus decision.

Two passes:
  1. Fast, full-coverage, no execution: has_negated_idb / seedable for
     every .dl-bearing directory in the whole tests/ tree (reuses
     corpus_predicate.check_program, read-only import).
  2. Slow, execution-requiring, capped: --magic-transform=* on the
     negated_idb+seedable subset, in lexicographic path order,
     checkpointed to disk every 50 programs so a cap leaves usable
     partial data. Records whether >=1 @neglabel. relation appeared and
     E_recoverable.

Writes to measurements/night01-t4/ only. Never touches tests/corpus/
(prohibition #2 -- this task's own candidate list is NOT the
pre-registered corpus and must not be confused with it or fed back into
it).
"""
import json
import os
import resource
import subprocess
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from corpus_predicate import check_program  # noqa: E402
from tuple_report import analyze as tuple_analyze  # noqa: E402

REPO = Path(__file__).resolve().parents[1]
SOUFFLE_TESTS = Path("/root/souffle-src/tests")
OUT_DIR = REPO / "measurements" / "night01-t4"

MEM_LIMIT_BYTES = 8 * 1024 * 1024 * 1024
TIMEOUT_S = 300
SLOW_PASS_CAP = 150  # exploratory cap, not a pre-registered corpus size


def _limit_mem():
    resource.setrlimit(resource.RLIMIT_AS, (MEM_LIMIT_BYTES, MEM_LIMIT_BYTES))


def fast_pass():
    results = {}
    for dirpath, _dirnames, filenames in os.walk(SOUFFLE_TESTS):
        dl_files = sorted(f for f in filenames if f.endswith(".dl"))
        if not dl_files:
            continue
        test_dir = Path(dirpath)
        rel_name = str(test_dir.relative_to(SOUFFLE_TESTS))
        # match build_corpus.py's own candidate-selection convention
        included = False
        detail = None
        for fn in dl_files:
            r = check_program(test_dir / fn)
            if detail is None or r["included"]:
                detail = {"dl_file": fn, **r}
            if r["included"]:
                included = True
                break
        results[rel_name] = {"included": included, **detail}
    return results


def run_magic(test_rel_path: str, dl_name: str):
    test_dir = SOUFFLE_TESTS / test_rel_path
    dl_path = test_dir / dl_name
    facts_dir = test_dir / "facts" if (test_dir / "facts").is_dir() else test_dir

    safe_name = test_rel_path.replace("/", "__")
    workdir = OUT_DIR / "runs" / safe_name
    workdir.mkdir(parents=True, exist_ok=True)
    log_name = "prof.log"

    cmd = ["souffle", "-F", str(facts_dir), "-D", str(workdir), "-p", log_name,
           "--magic-transform=*", str(dl_path)]

    try:
        proc = subprocess.run(
            cmd, cwd=str(workdir), capture_output=True, text=True,
            timeout=TIMEOUT_S, preexec_fn=_limit_mem,
        )
    except subprocess.TimeoutExpired:
        return {"path": test_rel_path, "status": "DNF:timeout-300s"}
    except MemoryError:
        return {"path": test_rel_path, "status": "DNF:memcap-8gb"}

    if proc.returncode != 0:
        return {"path": test_rel_path, "status": f"error:returncode-{proc.returncode}"}

    log_path = workdir / log_name
    if not log_path.is_file():
        return {"path": test_rel_path, "status": "error:no-log"}

    tr = tuple_analyze(log_path)
    return {
        "path": test_rel_path,
        "status": "ok",
        "has_neglabel": len(tr["neglabel_relations"]) > 0,
        "E_recoverable": tr["E_recoverable"],
        "T_souffle_excl_copy": tr["T_excl_copy"],
    }


def main():
    OUT_DIR.mkdir(parents=True, exist_ok=True)

    print("fast pass (structural, no execution)...", file=sys.stderr)
    fast = fast_pass()
    (OUT_DIR / "fast_pass.json").write_text(json.dumps(fast, indent=2, sort_keys=True))
    n_total = len(fast)
    n_neg_idb = sum(1 for v in fast.values() if v["has_negated_idb"])
    n_seedable = sum(1 for v in fast.values() if v["output_with_literal"])
    n_both = sum(1 for v in fast.values() if v["included"])
    print(f"fast pass: total={n_total} neg_idb={n_neg_idb} seedable={n_seedable} both={n_both}", file=sys.stderr)

    # Broader than T3 on purpose: T3 already covers the pre-registered
    # 36 (has_negated_idb AND seedable). T4's point is to look beyond
    # that -- every program with a negated IDB literal, seedable or not,
    # since "how common is the @neglabel shape at all" is the question,
    # not "does it qualify for the pre-registered corpus".
    candidates = sorted(k for k, v in fast.items() if v["has_negated_idb"])
    slow_candidates = candidates[:SLOW_PASS_CAP]
    print(f"slow pass: {len(slow_candidates)} of {len(candidates)} candidates (capped at {SLOW_PASS_CAP})",
          file=sys.stderr)

    slow_results = []
    for i, path in enumerate(slow_candidates, 1):
        dl_file = fast[path]["dl_file"]
        r = run_magic(path, dl_file)
        slow_results.append(r)
        if i % 10 == 0:
            print(f"  [{i}/{len(slow_candidates)}] {path}: {r['status']}", file=sys.stderr)
        if i % 50 == 0:
            (OUT_DIR / "slow_pass.json").write_text(json.dumps(slow_results, indent=2))
            print(f"  checkpoint written at {i}", file=sys.stderr)

    (OUT_DIR / "slow_pass.json").write_text(json.dumps(slow_results, indent=2))

    summary = {
        "fast_pass": {
            "total_dl_bearing_dirs": n_total,
            "with_negated_idb": n_neg_idb,
            "seedable": n_seedable,
            "both_negated_idb_and_seedable": n_both,
        },
        "slow_pass": {
            "cap": SLOW_PASS_CAP,
            "candidates_available": len(candidates),
            "candidates_run": len(slow_results),
            "ok": sum(1 for r in slow_results if r["status"] == "ok"),
            "with_neglabel": sum(1 for r in slow_results if r["status"] == "ok" and r["has_neglabel"]),
            "errors_or_dnf": sum(1 for r in slow_results if r["status"] != "ok"),
        },
    }
    (OUT_DIR / "summary.json").write_text(json.dumps(summary, indent=2))
    print(json.dumps(summary, indent=2))


if __name__ == "__main__":
    main()
