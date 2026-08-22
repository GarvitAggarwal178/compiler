#!/usr/bin/env python3
"""
J2 item 2: round-trip scaffold. parse -> pretty-print -> reparse ->
assert structural equality. The pretty-printer (and the parser) are
Lane A and do not exist yet -- both are stubs in dlc_interface.py. This
file is the test harness and the comparison logic, which are the real
Lane B deliverable; the printer itself is left as a stub, deliberately.

Every file in the corpus must report "not_implemented" today, at the
first stub call reached (parse). Never a silent skip, never a vacuous
"match".
"""
import json
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from dlc_interface import run_dlc_parse, run_dlc_pretty_print  # noqa: E402

REPO = Path(__file__).resolve().parents[1]
SOUFFLE_TESTS = Path("/root/souffle-src/tests")
IN_GRAMMAR = REPO / "tests" / "corpus" / "IN_GRAMMAR.txt"


def load_file_list():
    lines = IN_GRAMMAR.read_text().splitlines()
    return [line.strip() for line in lines if line.strip() and not line.startswith("#")]


def round_trip_check(source_text: str):
    r1 = run_dlc_parse(source_text)
    if r1.status != "parsed":
        return {"status": r1.status, "stage": "parse", "diagnostic": r1.diagnostic}

    r2 = run_dlc_pretty_print(r1.ast)
    if r2.status != "printed":
        return {"status": r2.status, "stage": "pretty_print", "diagnostic": r2.diagnostic}

    r3 = run_dlc_parse(r2.text)
    if r3.status != "parsed":
        return {"status": r3.status, "stage": "reparse", "diagnostic": r3.diagnostic}

    equal = (r1.ast == r3.ast)
    return {"status": "match" if equal else "mismatch", "stage": "compare"}


def main():
    files = load_file_list()
    results = []
    for rel in files:
        dl_path = SOUFFLE_TESTS / rel
        if not dl_path.is_file():
            results.append({"file": rel, "status": "errored", "stage": "read",
                             "diagnostic": "source file missing"})
            continue
        source = dl_path.read_text(errors="replace")
        r = round_trip_check(source)
        r["file"] = rel
        results.append(r)

    counts = {}
    for r in results:
        counts[r["status"]] = counts.get(r["status"], 0) + 1

    summary = {"total": len(results), "counts": counts, "results": results}
    out_path = REPO / "measurements" / "j2-round-trip-summary.json"
    out_path.write_text(json.dumps(summary, indent=2))
    print(json.dumps({"total": len(results), "counts": counts}, indent=2))

    if counts.get("not_implemented", 0) != len(results):
        print("ERROR: not every file reported not_implemented -- "
              "the round-trip scaffold produced a real verdict before Lane A exists",
              file=sys.stderr)
        raise SystemExit(1)
    print(f"OK: all {len(results)} files correctly blocked on not-implemented dlc parser/printer.",
          file=sys.stderr)


if __name__ == "__main__":
    main()
