#!/usr/bin/env python3
"""
Mechanical, structural inclusion predicate for the pre-registered corpus
(blueprint Q5; moved to Phase 0.6 per the directive on docs/reports/
probe0_5.md). No execution, no Soufflé invocation, no dlc parser (that's
M1, Lane A) -- a plain text-level scan of each .dl source, run once,
committed alongside its output.

Predicate (both must hold):
  1. At least one negated IDB literal: a `!` immediately preceding an
     identifier-and-open-paren, where that identifier is not itself
     declared `.input` in the same file (a mechanical proxy for "derived
     relation", since only .input relations are guaranteed extensional).
  2. At least one `.output` relation with a defining clause (fact or
     rule) whose head-argument list or body contains a numeric or string
     literal -- a mechanical proxy for "constant or bound argument",
     since our grammar has no other way to introduce a literal.

This is intentionally not a real parser: it can both over- and
under-match at the margin (e.g. a literal embedded in an unrelated
nested atom inside a long rule body). That is an accepted cost of fixing
the corpus set *before* any hand-probing continues, not a claim of
parser-grade precision.
"""
import json
import re
import sys
from pathlib import Path

COMMENT_BLOCK_RE = re.compile(r"/\*.*?\*/", re.DOTALL)
COMMENT_LINE_RE = re.compile(r"//[^\n]*")
INPUT_NAME_RE = re.compile(r"^\s*\.input\s+(\w+)", re.MULTILINE)
OUTPUT_NAME_RE = re.compile(r"^\s*\.output\s+(\w+)", re.MULTILINE)
NEG_ATOM_RE = re.compile(r"!\s*([A-Za-z_]\w*)\s*\(")
NUMBER_LIT_RE = re.compile(r"(?<![A-Za-z0-9_.])\d+(?![A-Za-z0-9_])")
STRING_LIT_RE = re.compile(r'"[^"]*"')


def strip_comments(text: str) -> str:
    text = COMMENT_BLOCK_RE.sub(" ", text)
    text = COMMENT_LINE_RE.sub(" ", text)
    return text


def clauses_for_head(text: str, rel: str):
    """Approximate every clause (fact or rule) whose head atom is `rel`."""
    pattern = re.compile(
        rf"\b{re.escape(rel)}\s*\([^)]*\)\s*(?::-[^.]*)?\.", re.DOTALL
    )
    return pattern.findall(text)


def has_literal(clause_text: str) -> bool:
    return bool(NUMBER_LIT_RE.search(clause_text) or STRING_LIT_RE.search(clause_text))


def check_program(dl_path: Path):
    raw = dl_path.read_text(errors="replace")
    text = strip_comments(raw)

    input_names = set(INPUT_NAME_RE.findall(text))
    output_names = set(OUTPUT_NAME_RE.findall(text))

    neg_targets = set(NEG_ATOM_RE.findall(text))
    neg_idb_targets = neg_targets - input_names
    has_negated_idb = bool(neg_idb_targets)

    output_with_literal = False
    matched_output = None
    # sorted(), not raw set iteration: Python's per-process randomized
    # string-hash seed made `matched_output` (a diagnostic field only --
    # `included` itself was always order-independent) non-deterministic
    # across runs. Found by NIGHT-BATCH-01 T1 (docs/reports/
    # night01-T1-audit.md); fix authorized narrowly by the 2026-08-21
    # corpus ruling section 4.3, ordering only, no other predicate logic
    # touched.
    for rel in sorted(output_names):
        for clause in clauses_for_head(text, rel):
            if has_literal(clause):
                output_with_literal = True
                matched_output = rel
                break
        if output_with_literal:
            break

    included = has_negated_idb and output_with_literal
    return {
        "included": included,
        "has_negated_idb": has_negated_idb,
        "negated_idb_targets": sorted(neg_idb_targets),
        "output_with_literal": output_with_literal,
        "matched_output_relation": matched_output,
        "input_names": sorted(input_names),
        "output_names": sorted(output_names),
    }


def main():
    if len(sys.argv) != 2:
        print(f"usage: {sys.argv[0]} <corpus-root-dir>", file=sys.stderr)
        raise SystemExit(2)

    root = Path(sys.argv[1])
    results = {}
    for test_dir in sorted(p for p in root.iterdir() if p.is_dir()):
        dl_files = list(test_dir.glob("*.dl"))
        if not dl_files:
            continue
        # One .dl per test directory in this corpus; if more than one,
        # check them all and include the test if any qualifies.
        included = False
        detail = None
        for dl_path in dl_files:
            r = check_program(dl_path)
            if r["included"]:
                included = True
                detail = r
                break
            if detail is None:
                detail = r
        results[test_dir.name] = {"included": included, **detail}

    included_names = sorted(k for k, v in results.items() if v["included"])

    print(json.dumps({
        "corpus_root": str(root),
        "total_tests": len(results),
        "included_count": len(included_names),
        "included": included_names,
    }, indent=2))


if __name__ == "__main__":
    main()
