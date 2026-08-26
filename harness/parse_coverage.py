#!/usr/bin/env python3
"""
M1 §3.3 gate one: parse-coverage runner. Runs the real dlc parser
(harness/dlc_interface.run_dlc_parse) over all 195 files in
tests/corpus/IN_GRAMMAR.txt and reports parsed/error/panic counts, per
docs/M1-BUILD.md §3.3 ("report it as parsed/195").

This file previously (J2, NIGHT-BATCH-01/02) asserted 100% not_implemented
-- that assertion is retired now that the parser is real; a parser that
still reported not_implemented today would itself be the bug. The
"expected" number here is NOT necessarily 195/195: §3.1's lex-coverage run
already found 42/195 files use aggregates/records/#include/pragma syntax
outside blueprint §4's grammar (docs/OPEN_QUESTIONS.md 2026-08-26), and
§3.3's parser additionally rejects a documented, deliberate second
category (`.input`/`.output` with a trailing `()`, real Soufflé syntax
but not in blueprint's grammar plus its one authorized amendment -- see
src/parser/DESIGN.md). The actual count is reported as measured, with a
breakdown by cause where possible -- never adjusted to hit 195.
"""
import json
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import dlc_interface  # noqa: E402

REPO = Path(__file__).resolve().parents[1]
SOUFFLE_TESTS = Path("/root/souffle-src/tests")
IN_GRAMMAR = REPO / "tests" / "corpus" / "IN_GRAMMAR.txt"


def load_file_list():
    lines = IN_GRAMMAR.read_text().splitlines()
    return [line.strip() for line in lines if line.strip() and not line.startswith("#")]


def main():
    dlc_interface.build_dlc()

    files = load_file_list()
    results = []
    for rel in files:
        dl_path = SOUFFLE_TESTS / rel
        if not dl_path.is_file():
            results.append({"file": rel, "status": "missing_source", "diagnostic": "source file missing"})
            continue
        source = dl_path.read_text(errors="replace")
        r = dlc_interface.run_dlc_parse(source)
        results.append({
            "file": rel, "status": r.status,
            "decl_count": r.decl_count, "clause_count": r.clause_count,
            "error_count": r.error_count, "diagnostics": r.diagnostics,
            "diagnostic": r.diagnostic,
        })

    counts = {}
    for r in results:
        counts[r["status"]] = counts.get(r["status"], 0) + 1

    summary = {"total": len(results), "counts": counts, "results": results}
    out_path = REPO / "measurements" / "m1-3.3-gate1-parse-coverage-summary.json"
    out_path.write_text(json.dumps(summary, indent=2))

    parsed = counts.get("parsed", 0)
    print(json.dumps({"total": len(results), "counts": counts}, indent=2))
    print(f"GATE RESULT: parsed/{len(results)} = {parsed}/{len(results)}", file=sys.stderr)

    panics = [r for r in results if r["status"] in ("panic", "read_error", "build_missing", "missing_source")]
    if panics:
        print(f"{len(panics)} file(s) hit panic/build/read problems (not just parse errors) -- see summary JSON",
              file=sys.stderr)
        raise SystemExit(1)


if __name__ == "__main__":
    main()
