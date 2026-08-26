#!/usr/bin/env python3
"""
M1 §3.5 gate: (1) all 15 tests/programs/allowedness_probe_*.dl produce
the verdicts recorded in docs/reports/J1-allowedness-probe.md (a-g) and
docs/reports/night02-T1-allowedness.md (h-o), run end-to-end through the
real `dlc check` CLI (already independently verified as Go unit tests,
src/sema/allowedness_test.go -- this script re-checks the same thing
through the actual binary boundary, not just in-process Go). (2) all 13
cases across all four tests/rejection/*.py grounds are looked at -- not
all 13 can be REJECTED yet: stratification checking is §3.6, not §3.5,
so the 3 stratification cases are expected to still report "ok" (falsely
accepted, from dlc's current point of view) until that item lands. This
is reported explicitly as an attributed gap, not silently passed over.
"""
import importlib.util
import json
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import dlc_interface  # noqa: E402

REPO = Path(__file__).resolve().parents[1]
PROGRAMS_DIR = REPO / "tests" / "programs"
REJECTION_DIR = REPO / "tests" / "rejection"

PROBE_VERDICT = {  # True = accept (allowed)
    "a": True, "b": True, "c": False, "d": False, "e": True, "f": False, "g": True,
    "h": False, "i": True, "j": False, "k": True, "l": True, "m": False, "n": True, "o": True,
}

EXPECTED_CATEGORY = {"arity": "arity_mismatch", "type": "type_mismatch", "allowedness": "allowedness",
                      "stratification": "unstratifiable"}  # stratification: §3.6, not yet implemented


def load_cases(ground):
    spec = importlib.util.spec_from_file_location(ground, REJECTION_DIR / f"{ground}.py")
    mod = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mod)
    return mod.CASES


def main():
    dlc_interface.build_dlc()

    probe_results = []
    for letter, expect_accept in PROBE_VERDICT.items():
        src = (PROGRAMS_DIR / f"allowedness_probe_{letter}.dl").read_text()
        r = dlc_interface.run_dlc_check(src)
        got_accept = (r.status == "ok")
        ok = got_accept == expect_accept
        probe_results.append({"case": letter, "expect_accept": expect_accept, "got_accept": got_accept,
                               "status": r.status, "ok": ok})
    n_probe_ok = sum(1 for r in probe_results if r["ok"])

    rejection_results = []
    for ground in ("arity", "type", "allowedness", "stratification"):
        for case in load_cases(ground):
            r = dlc_interface.run_dlc_check(case["program"])
            cats = {d["category"] for d in r.diagnostics}
            want = EXPECTED_CATEGORY[ground]
            rejected_correctly = r.status == "rejected" and want in cats
            rejection_results.append({
                "ground": ground, "name": case["name"], "status": r.status,
                "categories": sorted(cats), "rejected_correctly": rejected_correctly,
                "expected_gap": (ground == "stratification"),  # §3.6 not implemented yet
            })
    n_rejection_ok = sum(1 for r in rejection_results if r["rejected_correctly"])
    n_expected_gap = sum(1 for r in rejection_results if r["expected_gap"])

    summary = {
        "probe_cases": f"{n_probe_ok}/{len(probe_results)}",
        "probe_results": probe_results,
        "rejection_corpus": f"{n_rejection_ok}/{len(rejection_results)} "
                             f"({n_expected_gap} of the shortfall is the known §3.6 stratification gap)",
        "rejection_results": rejection_results,
    }
    out_path = REPO / "measurements" / "m1-3.5-allowedness-summary.json"
    out_path.write_text(json.dumps(summary, indent=2))
    print(json.dumps({"probe_cases": summary["probe_cases"], "rejection_corpus": summary["rejection_corpus"]}, indent=2))

    if n_probe_ok != len(probe_results):
        print(f"GATE FAILED: only {n_probe_ok}/{len(probe_results)} probe cases matched their recorded verdict",
              file=sys.stderr)
        for r in probe_results:
            if not r["ok"]:
                print(f"  case {r['case']}: expected accept={r['expect_accept']}, got accept={r['got_accept']} (status={r['status']})", file=sys.stderr)
        raise SystemExit(1)

    non_stratification_shortfall = [r for r in rejection_results if not r["rejected_correctly"] and not r["expected_gap"]]
    if non_stratification_shortfall:
        print(f"GATE FAILED: {len(non_stratification_shortfall)} rejection-corpus case(s) outside the known "
              f"§3.6 gap were not correctly rejected:", file=sys.stderr)
        for r in non_stratification_shortfall:
            print(f"  {r['ground']}/{r['name']}: status={r['status']} categories={r['categories']}", file=sys.stderr)
        raise SystemExit(1)

    print(f"GATE PASSED: {n_probe_ok}/15 probe cases correct; "
          f"{n_rejection_ok}/13 rejection-corpus cases correctly rejected "
          f"({n_expected_gap}/13 stratification cases are the known, attributed §3.6 gap, not a §3.5 failure)",
          file=sys.stderr)


if __name__ == "__main__":
    main()
