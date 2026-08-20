#!/usr/bin/env python3
"""
Apply harness/corpus_predicate.py's mechanical inclusion predicate over
every .dl-bearing directory anywhere under a Soufflé `tests/` checkout,
and write the pre-registered corpus list. Structural only: no execution,
no Soufflé invocation.

Usage: python3 harness/build_corpus.py <souffle-tests-root>

Writes tests/corpus/PREREGISTERED.txt (one relative path per line,
sorted) and tests/corpus/detail.json (full predicate output per test,
for provenance -- includes why each test was or was not included).
"""
import json
import os
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from corpus_predicate import check_program  # noqa: E402

REPO = Path(__file__).resolve().parents[1]
CORPUS_DIR = REPO / "tests" / "corpus"


def main():
    if len(sys.argv) != 2:
        print(f"usage: {sys.argv[0]} <souffle-tests-root>", file=sys.stderr)
        raise SystemExit(2)

    root = Path(sys.argv[1]).resolve()
    results = {}
    for dirpath, _dirnames, filenames in os.walk(root):
        dl_files = sorted(f for f in filenames if f.endswith(".dl"))
        if not dl_files:
            continue
        test_dir = Path(dirpath)
        rel_name = str(test_dir.relative_to(root))
        included = False
        detail = None
        for fn in dl_files:
            r = check_program(test_dir / fn)
            if r["included"]:
                included = True
                detail = {"dl_file": fn, **r}
                break
            if detail is None:
                detail = {"dl_file": fn, **r}
        results[rel_name] = {"included": included, **detail}

    included_names = sorted(k for k, v in results.items() if v["included"])

    CORPUS_DIR.mkdir(parents=True, exist_ok=True)
    (CORPUS_DIR / "PREREGISTERED.txt").write_text(
        "\n".join(included_names) + "\n"
    )
    (CORPUS_DIR / "detail.json").write_text(
        json.dumps(
            {
                "corpus_root_scanned": str(root),
                "total_dl_bearing_dirs": len(results),
                "included_count": len(included_names),
                "results": results,
            },
            indent=2,
            sort_keys=True,
        )
        + "\n"
    )

    print(f"total_dl_bearing_dirs={len(results)}", file=sys.stderr)
    print(f"included_count={len(included_names)}", file=sys.stderr)


if __name__ == "__main__":
    main()
