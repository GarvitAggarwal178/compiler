#!/usr/bin/env python3
"""
M1 §3.1 gate: runs the real dlc lexer over all 195 files in
tests/corpus/IN_GRAMMAR.txt and all 39 files in tests/hostile/, and
asserts zero panics across all 234 files. A panic (Go panic surfaced via
cmd/dlc's recover, a nonzero exit with no parseable JSON, or non-JSON
stdout) fails the gate; error TOKENS (malformed input correctly reported
as an ERROR token with a span) are expected and do not fail it -- the
gate is "the lexer never crashes", not "every file lexes cleanly".

Named lex_coverage.py, not parse_coverage.py, deliberately: §3.1 is the
lexer only, and parse_coverage.py's existing contract (asserting
100% not_implemented, harness/parse_coverage.py's own gate today) is
about the *parser* stage, which does not exist until §3.3. Overloading
one script's meaning across two different gates seemed more likely to
confuse a future reader than writing a second small script -- noted in
the M1 progress report as a deliberate naming deviation from
M1-BUILD.md's literal text, not a silent one.
"""
import json
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import dlc_interface  # noqa: E402

REPO = Path(__file__).resolve().parents[1]
SOUFFLE_TESTS = Path("/root/souffle-src/tests")
IN_GRAMMAR = REPO / "tests" / "corpus" / "IN_GRAMMAR.txt"
HOSTILE_DIR = REPO / "tests" / "hostile"
MEASUREMENTS = REPO / "measurements"


def load_in_grammar_files():
    lines = IN_GRAMMAR.read_text().splitlines()
    return [(rel, SOUFFLE_TESTS / rel) for rel in lines if rel.strip() and not rel.startswith("#")]


def main():
    dlc_interface.build_dlc()

    files = load_in_grammar_files()
    files += [(f"tests/hostile/{p.name}", p) for p in sorted(HOSTILE_DIR.glob("*.dl"))]

    results = []
    for label, path in files:
        if not path.is_file():
            results.append({"file": label, "status": "missing_source"})
            continue
        r = dlc_interface.run_dlc_lex(path)
        results.append({
            "file": label, "status": r.status,
            "token_count": r.token_count, "error_count": r.error_count,
            "diagnostic": r.diagnostic,
        })

    counts = {}
    for r in results:
        counts[r["status"]] = counts.get(r["status"], 0) + 1

    panics = [r for r in results if r["status"] in ("panic", "read_error", "build_missing", "missing_source")]

    summary = {"total": len(results), "counts": counts, "panics": panics, "results": results}
    out_path = MEASUREMENTS / "m1-3.1-lex-coverage-summary.json"
    out_path.write_text(json.dumps(summary, indent=2))
    print(json.dumps({"total": len(results), "counts": counts, "panic_count": len(panics)}, indent=2))

    if panics:
        print(f"GATE FAILED: {len(panics)} file(s) triggered a panic/build/read problem", file=sys.stderr)
        for p in panics[:20]:
            print(f"  {p['file']}: {p['status']} -- {p.get('diagnostic', '')[:200]}", file=sys.stderr)
        raise SystemExit(1)
    print(f"GATE PASSED: 0/{len(results)} panics ({counts.get('lexed', 0)} lexed cleanly through the CLI, "
          f"{sum(r['error_count'] for r in results if r['status']=='lexed')} total lex-error tokens across all files)",
          file=sys.stderr)


if __name__ == "__main__":
    main()
