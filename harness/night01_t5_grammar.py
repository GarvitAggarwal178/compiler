#!/usr/bin/env python3
"""
NIGHT-BATCH-01 T5: grammar coverage census. For every .dl file in
Soufflé's tests/ tree, mechanically detect features outside blueprint
v1.1/v1.2 section 4's grammar (the grammar itself is unchanged across
v1.0-v1.2). No execution, no dlc parser (that's M1, Lane A) -- a text
scan, same spirit and same caveats as corpus_predicate.py: this is not a
real parse, it is good enough to give M1 a differential test pool before
the parser exists.

Writes tests/corpus/IN_GRAMMAR.txt (NOT the pre-registered corpus --
header comment says so explicitly) and a full per-file detail dump for
provenance.
"""
import json
import os
import re
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from corpus_predicate import strip_comments  # noqa: E402

REPO = Path(__file__).resolve().parents[1]
SOUFFLE_TESTS = Path("/root/souffle-src/tests")
OUT_DIR = REPO / "measurements" / "night01-t5"
IN_GRAMMAR_PATH = REPO / "tests" / "corpus" / "IN_GRAMMAR.txt"

FEATURE_PATTERNS = {
    "type_decl": re.compile(r"^\s*\.type\b", re.MULTILINE),
    "plan_directive": re.compile(r"\.plan\b"),
    "pragma_directive": re.compile(r"^\s*\.pragma\b", re.MULTILINE),
    "component": re.compile(r"\.comp\b|\.init\b"),
    "aggregate": re.compile(r"\b(count|sum|max|min|mean)\s*:"),
    "functor_call": re.compile(r"\.functor\b|@\w+\s*\("),
    "record_or_list_term": re.compile(r"\.type\s+\w+\s*=\s*\["),
    "choice_domain": re.compile(r"choice-domain"),
    "adt": re.compile(r"\.type\s+\w+\s*=.*\{"),
    "subsumption": re.compile(r"\)\s*<=\s*[A-Za-z_]\w*\s*\("),
    "decl_qualifier": re.compile(r"\.decl\s+\w+\s*\([^)]*\)\s*(btree_delete|overridable|inline)\b"),
    "disjunction_semicolon": re.compile(r";"),
}


def classify(text: str):
    found = {}
    for name, pat in FEATURE_PATTERNS.items():
        n = len(pat.findall(text))
        if n:
            found[name] = n
    return found


def main():
    OUT_DIR.mkdir(parents=True, exist_ok=True)
    results = {}
    for dirpath, _dirnames, filenames in os.walk(SOUFFLE_TESTS):
        for fn in sorted(filenames):
            if not fn.endswith(".dl"):
                continue
            fp = Path(dirpath) / fn
            rel = str(fp.relative_to(SOUFFLE_TESTS))
            raw = fp.read_text(errors="replace")
            text = strip_comments(raw)
            feats = classify(text)
            results[rel] = {
                "in_grammar": len(feats) == 0,
                "features": feats,
            }

    total = len(results)
    in_grammar = sorted(k for k, v in results.items() if v["in_grammar"])
    out_grammar = [k for k, v in results.items() if not v["in_grammar"]]

    histogram = {}
    for k in out_grammar:
        for feat in results[k]["features"]:
            histogram[feat] = histogram.get(feat, 0) + 1

    (OUT_DIR / "detail.json").write_text(json.dumps(results, indent=2, sort_keys=True))
    (OUT_DIR / "summary.json").write_text(json.dumps({
        "total_dl_files": total,
        "in_grammar_count": len(in_grammar),
        "out_of_grammar_count": len(out_grammar),
        "histogram": dict(sorted(histogram.items(), key=lambda kv: -kv[1])),
    }, indent=2))

    IN_GRAMMAR_PATH.write_text(
        "# In-grammar file census (NIGHT-BATCH-01 T5), NOT the pre-registered corpus.\n"
        "# This is every .dl file under Souffle's tests/ tree that a mechanical text\n"
        "# scan found zero out-of-grammar features in, against blueprint section 4's\n"
        "# grammar (v1.0-v1.2, unchanged). It has NOT been run, NOT been checked for\n"
        "# negation/seedability, and is NOT tests/corpus/PREREGISTERED.txt. It exists\n"
        "# to give M1's parser a differential test pool before it is finished.\n"
        "# Predicate source: harness/night01_t5_grammar.py. Do not confuse the two files.\n"
        + "\n".join(in_grammar) + "\n"
    )

    print(f"total={total} in_grammar={len(in_grammar)} out_of_grammar={len(out_grammar)}", file=sys.stderr)
    print(json.dumps(histogram, indent=2))


if __name__ == "__main__":
    main()
