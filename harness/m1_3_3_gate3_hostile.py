#!/usr/bin/env python3
"""
M1 §3.3 gate three: runs the real dlc parser over all 39
tests/hostile/*.dl files and compares its accept/reject verdict against
the Soufflé-established verdicts recorded in
measurements/night02-t2-hostile-summary.json (docs/reports/
night02-T2-hostile.md).

Important asymmetry, not a bug: dlc at this point has a parser only (§3.3)
-- no decl/type/allowedness/stratification checking (§3.4-3.6 do not
exist yet). Several of Soufflé's 8 historical rejections are SEMANTIC,
not syntactic (duplicate declaration, underscore-in-head) -- dlc cannot
be expected to reject those yet, and this script reports that
distinction explicitly rather than scoring it as a plain pass/fail.
"""
import json
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import dlc_interface  # noqa: E402

REPO = Path(__file__).resolve().parents[1]
HOSTILE_DIR = REPO / "tests" / "hostile"
SOUFFLE_VERDICTS = REPO / "measurements" / "night02-t2-hostile-summary.json"

# Files whose ONLY known Soufflé rejection reason is semantic (post-parse),
# not syntactic -- dlc's parser alone cannot be expected to reject these
# until §3.4/§3.6 exist. Determined by reading each file's diagnostic in
# docs/reports/night02-T2-hostile.md.
SEMANTIC_NOT_SYNTACTIC = {
    "malformed_duplicate_decl.dl",   # "Redefinition of relation p" -- symbol-table check
    "semantic_wildcard_in_head.dl",  # "Underscore in head of rule" -- a post-parse restriction
}
# Already known inconclusive at the language level (T2's own report):
# a runtime fact-loading failure, not a parse/lex rejection.
KNOWN_INCONCLUSIVE = {"lexical_4kb_identifier.dl"}


def main():
    dlc_interface.build_dlc()
    souffle = {r["file"]: r["outcome"] for r in json.loads(SOUFFLE_VERDICTS.read_text())}

    results = []
    for dl_path in sorted(HOSTILE_DIR.glob("*.dl")):
        r = dlc_interface.run_dlc_parse(dl_path.read_text(errors="replace"))
        dlc_verdict = "accept" if r.status == "parsed" else ("reject" if r.status == "error" else r.status)
        souffle_verdict = souffle.get(dl_path.name, "unknown")
        agree = (dlc_verdict == souffle_verdict)
        category = "agree" if agree else "disagree"
        if dl_path.name in SEMANTIC_NOT_SYNTACTIC and not agree:
            category = "disagree_expected_semantic_not_yet_implemented"
        if dl_path.name in KNOWN_INCONCLUSIVE:
            category = "inconclusive_per_T2"
        results.append({
            "file": dl_path.name, "dlc_verdict": dlc_verdict, "souffle_verdict": souffle_verdict,
            "category": category, "diagnostic_count": r.error_count,
        })

    counts = {}
    for r in results:
        counts[r["category"]] = counts.get(r["category"], 0) + 1

    summary = {"total": len(results), "counts": counts, "results": results}
    out_path = REPO / "measurements" / "m1-3.3-gate3-hostile-summary.json"
    out_path.write_text(json.dumps(summary, indent=2))

    print(json.dumps({"total": len(results), "counts": counts}, indent=2))
    genuine_disagree = [r for r in results if r["category"] == "disagree"]
    if genuine_disagree:
        print(f"{len(genuine_disagree)} GENUINE disagreement(s) (not attributable to missing sema):", file=sys.stderr)
        for r in genuine_disagree:
            print(f"  {r['file']}: dlc={r['dlc_verdict']} souffle={r['souffle_verdict']}", file=sys.stderr)
    print(f"agree={counts.get('agree',0)} expected-semantic-gap={counts.get('disagree_expected_semantic_not_yet_implemented',0)} "
          f"inconclusive={counts.get('inconclusive_per_T2',0)} genuine-disagree={len(genuine_disagree)}", file=sys.stderr)


if __name__ == "__main__":
    main()
