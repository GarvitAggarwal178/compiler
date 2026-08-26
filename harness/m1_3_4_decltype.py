#!/usr/bin/env python3
"""
M1 §3.4 gate (no numeric gate is stated in M1-BUILD.md for this item;
this is the completion criterion this session adopted, analogous to how
§3.2 (AST) had none either): runs the real dlc decl/arity/type checker
(harness/dlc_interface.run_dlc_check) over the 6 real arity+type cases in
tests/rejection/{arity,type}.py and confirms each is REJECTED with the
correct Category (docs/reports/night02-T9-diagnostics.md's
classification, not verbatim text -- §3.4's own instruction). Also spot-
checks that sema does not spuriously reject the files gate one (§3.3)
already found parse cleanly, as a sanity check against "a checker that
rejects everything" trivially passing the rejection-corpus half.
"""
import importlib.util
import json
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import dlc_interface  # noqa: E402

REPO = Path(__file__).resolve().parents[1]
REJECTION_DIR = REPO / "tests" / "rejection"
SOUFFLE_TESTS = Path("/root/souffle-src/tests")
GATE1_SUMMARY = REPO / "measurements" / "m1-3.3-gate1-parse-coverage-summary.json"

EXPECTED_CATEGORY = {"arity": "arity_mismatch", "type": "type_mismatch"}


def load_cases(ground):
    spec = importlib.util.spec_from_file_location(ground, REJECTION_DIR / f"{ground}.py")
    mod = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mod)
    return mod.CASES


def main():
    dlc_interface.build_dlc()

    results = []
    for ground in ("arity", "type"):
        for case in load_cases(ground):
            r = dlc_interface.run_dlc_check(case["program"])
            cats = {d["category"] for d in r.diagnostics}
            want = EXPECTED_CATEGORY[ground]
            ok = r.status == "rejected" and want in cats
            results.append({
                "ground": ground, "name": case["name"], "status": r.status,
                "categories": sorted(cats), "expected_category": want, "ok": ok,
            })

    n_ok = sum(1 for r in results if r["ok"])
    print(json.dumps({"rejection_corpus": f"{n_ok}/{len(results)}", "results": results}, indent=2))

    # Sanity check: sema should not reject every real-world file gate one
    # already found parses cleanly -- a checker that rejects everything
    # would trivially "pass" the rejection corpus above for the wrong
    # reason. Not a strict pass/fail gate (some of these 20 files may
    # have genuine declared-type mismatches dlc is right to flag,
    # something Soufflé's own type system might handle differently) --
    # reported as a number for a human to look at, not asserted on.
    clean_check = {"ok": 0, "rejected": 0, "parse_error": 0, "other": 0}
    if GATE1_SUMMARY.is_file():
        gate1 = json.loads(GATE1_SUMMARY.read_text())
        parsed_files = [r["file"] for r in gate1["results"] if r["status"] == "parsed"]
        for rel in parsed_files:
            source = (SOUFFLE_TESTS / rel).read_text(errors="replace")
            r = dlc_interface.run_dlc_check(source)
            clean_check[r.status] = clean_check.get(r.status, 0) + 1
    print(json.dumps({"sanity_check_of_gate1s_20_parsed_files": clean_check}, indent=2), file=sys.stderr)

    out_path = REPO / "measurements" / "m1-3.4-decltype-summary.json"
    out_path.write_text(json.dumps({"rejection_corpus_results": results, "sanity_check": clean_check}, indent=2))

    if n_ok != len(results):
        print(f"GATE FAILED: only {n_ok}/{len(results)} rejection-corpus cases classified correctly", file=sys.stderr)
        raise SystemExit(1)
    print(f"GATE PASSED: {n_ok}/{len(results)} arity+type rejection-corpus cases classified correctly", file=sys.stderr)


if __name__ == "__main__":
    main()
