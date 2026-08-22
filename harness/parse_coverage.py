#!/usr/bin/env python3
"""
J2 item 1: parse-coverage runner. Takes tests/corpus/IN_GRAMMAR.txt,
invokes dlc's parse entry point (harness/dlc_interface.run_dlc_parse --
currently a stub, must fail cleanly with "not_implemented", never pass
vacuously), reports parsed/failed/errored with per-file diagnostics.

This is the human's day-2 acceptance gate: it exists and is runnable
*before* the parser does, so the first real parser commit has something
to report numbers against immediately, per this session's J2 instructions.
"""
import json
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from dlc_interface import run_dlc_parse  # noqa: E402

REPO = Path(__file__).resolve().parents[1]
SOUFFLE_TESTS = Path("/root/souffle-src/tests")
IN_GRAMMAR = REPO / "tests" / "corpus" / "IN_GRAMMAR.txt"


def load_file_list():
    lines = IN_GRAMMAR.read_text().splitlines()
    return [line.strip() for line in lines if line.strip() and not line.startswith("#")]


def main():
    files = load_file_list()
    results = []
    for rel in files:
        dl_path = SOUFFLE_TESTS / rel
        if not dl_path.is_file():
            results.append({"file": rel, "status": "errored",
                             "diagnostic": "source file missing (souffle-src checkout not present or path stale)"})
            continue
        source = dl_path.read_text(errors="replace")
        r = run_dlc_parse(source)
        status = {"not_implemented": "not_implemented", "parsed": "parsed"}.get(r.status, "failed")
        results.append({"file": rel, "status": status, "diagnostic": r.diagnostic})

    counts = {}
    for r in results:
        counts[r["status"]] = counts.get(r["status"], 0) + 1

    summary = {"total": len(results), "counts": counts, "results": results}
    out_path = REPO / "measurements" / "j2-parse-coverage-summary.json"
    out_path.write_text(json.dumps(summary, indent=2))
    print(json.dumps({"total": len(results), "counts": counts}, indent=2))

    # Today, every file must report not_implemented -- a green (or even
    # partially-green) suite before the parser exists is a bug in the
    # suite, not progress.
    if counts.get("not_implemented", 0) != len(results):
        print("ERROR: not every file reported not_implemented -- "
              "something is faking a parse result before Lane A exists",
              file=sys.stderr)
        raise SystemExit(1)
    print(f"OK: all {len(results)} files correctly blocked on not-implemented dlc parser.", file=sys.stderr)


if __name__ == "__main__":
    main()
