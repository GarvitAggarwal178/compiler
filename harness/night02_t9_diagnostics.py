#!/usr/bin/env python3
"""
NIGHT-BATCH-02 T9: Soufflé diagnostic catalogue. Runs every case in
tests/rejection/{arity,type,stratification,allowedness}.py plus one new
minimal program for the "undeclared relation" ground (not one of the
project's four rejection grounds, but one of the seven error classes T9
asks this catalogue to cover) against the installed Soufflé 2.5, and
records the exact diagnostic text, whether it carries a line/column, and
whether Soufflé's actual rejection reason is consistent with each case's
`expected_ground`/`expected_diagnosis` prose (a human-readable description
of the reasoning, not a verbatim-match target -- J2 already established
that convention for the allowedness cases).

"ungrounded variable", "syntax error", "duplicate declaration", and
"unstratifiable negation" (the transform-introduced kind) already have
directly-measured catalogue entries from earlier in this batch (J1/T1,
T2, T5) -- not re-run here, cited by measurement ID instead.
"""
import importlib.util
import json
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import probe0  # noqa: E402

REPO = probe0.REPO
MEASUREMENTS = probe0.MEASUREMENTS
REJECTION_DIR = REPO / "tests" / "rejection"


def load_cases(module_name):
    spec = importlib.util.spec_from_file_location(module_name, REJECTION_DIR / f"{module_name}.py")
    mod = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mod)
    return mod.CASES


def main():
    facts_dir = MEASUREMENTS / "_scratch_night02_t9" / "facts"
    facts_dir.mkdir(parents=True, exist_ok=True)
    for name in ("a", "base", "baz", "foo"):
        (facts_dir / f"{name}.facts").write_text("")

    results = []

    for module_name in ("arity", "type", "stratification", "allowedness"):
        for case in load_cases(module_name):
            mid = f"night02-t9-{module_name}-{case['name']}"
            workdir = MEASUREMENTS / f"{mid}-run"
            workdir.mkdir(parents=True, exist_ok=True)
            dl_path = workdir / "prog.dl"
            dl_path.write_text(case["program"])
            cmd = ["souffle", "-F", str(facts_dir), "-D", str(workdir), str(dl_path)]
            proc = probe0.run_cmd(mid, cmd, workdir)
            rejected = (proc.returncode != 0) or ("Error:" in proc.stderr)
            results.append({
                "ground_module": module_name,
                "name": case["name"],
                "expected_ground": case["expected_ground"],
                "outcome": "reject" if rejected else "accept",
                "returncode": proc.returncode,
                "stderr": proc.stderr.strip(),
                "has_line_col": ("at line" in proc.stderr),
            })
            print(f"{module_name}/{case['name']}: "
                  f"{'reject' if rejected else 'accept'} (rc={proc.returncode})", file=sys.stderr)

    # undeclared relation -- new, minimal, not part of the 4 rejection grounds
    dl_path = REPO / "tests" / "programs" / "diagnostic_undeclared_relation.dl"
    mid = "night02-t9-undeclared-relation"
    workdir = MEASUREMENTS / f"{mid}-run"
    workdir.mkdir(parents=True, exist_ok=True)
    cmd = ["souffle", "-F", str(facts_dir), "-D", str(workdir), str(dl_path)]
    proc = probe0.run_cmd(mid, cmd, workdir)
    rejected = (proc.returncode != 0) or ("Error:" in proc.stderr)
    results.append({
        "ground_module": "undeclared_relation", "name": "undeclared_relation_minimal",
        "expected_ground": "undeclared_relation",
        "outcome": "reject" if rejected else "accept",
        "returncode": proc.returncode, "stderr": proc.stderr.strip(),
        "has_line_col": ("at line" in proc.stderr),
    })
    print(f"undeclared_relation: {'reject' if rejected else 'accept'} (rc={proc.returncode})", file=sys.stderr)

    out_path = MEASUREMENTS / "night02-t9-diagnostics-summary.json"
    out_path.write_text(json.dumps(results, indent=2))
    n_reject = sum(1 for r in results if r["outcome"] == "reject")
    print(json.dumps({"total": len(results), "reject": n_reject, "accept": len(results) - n_reject}, indent=2))


if __name__ == "__main__":
    main()
