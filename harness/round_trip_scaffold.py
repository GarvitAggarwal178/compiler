#!/usr/bin/env python3
"""
M1 §3.3 gate two: round-trip scaffold. Runs the real `dlc roundtrip`
subcommand (parse -> print -> reparse -> ast.Equal, entirely in Go --
src/parser/roundtrip.go) over all 195 files in tests/corpus/IN_GRAMMAR.txt.

Rewritten from the original J2-era design (run_dlc_parse +
run_dlc_pretty_print, compared in Python) now that the parser and printer
are real: the comparison itself moved into Go (dlc_interface.py's module
docstring explains why -- ast.Equal is what src/ast/equal.go was built
for, and reimplementing it against a JSON AST dump in Python would just
be a second, divergence-prone copy of the same logic). Every file that
does not even *parse* (§3.3 gate one) is reported as "parse_error" here
too, not skipped -- gate two only meaningfully applies to files gate one
already accepts, but that is visible in the results, not hidden by
filtering them out upfront.
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
        r = dlc_interface.run_dlc_roundtrip(source)
        results.append({
            "file": rel, "status": r.status,
            "diagnostics": r.diagnostics, "diagnostic": r.diagnostic,
        })

    counts = {}
    for r in results:
        counts[r["status"]] = counts.get(r["status"], 0) + 1

    summary = {"total": len(results), "counts": counts, "results": results}
    out_path = REPO / "measurements" / "m1-3.3-gate2-roundtrip-summary.json"
    out_path.write_text(json.dumps(summary, indent=2))

    matched = counts.get("match", 0)
    print(json.dumps({"total": len(results), "counts": counts}, indent=2))
    print(f"GATE RESULT: match/{len(results)} = {matched}/{len(results)}", file=sys.stderr)

    panics = [r for r in results if r["status"] in ("panic", "read_error", "build_missing", "missing_source")]
    mismatches = [r for r in results if r["status"] == "mismatch"]
    if panics:
        print(f"{len(panics)} file(s) hit panic/build/read problems", file=sys.stderr)
    if mismatches:
        print(f"{len(mismatches)} file(s) round-tripped to a DIFFERENT AST -- a real printer/parser "
              f"precedence or shape bug, not a parse-error:", file=sys.stderr)
        for m in mismatches[:10]:
            print(f"  {m['file']}", file=sys.stderr)
    if panics or mismatches:
        raise SystemExit(1)


if __name__ == "__main__":
    main()
