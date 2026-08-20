#!/usr/bin/env python3
"""
NIGHT-BATCH-01 T7 item 3: rejection-test scaffolding runner. Loads every
case from tests/rejection/{arity,type,allowedness,stratification}.py and
runs it through the stubbed dlc interface (harness/differential.py's
run_dlc). Every case must currently report "not_implemented" -- this
script exists to prove the scaffolding is wired correctly (cases load,
programs are well-formed enough to reach the stub, the stub is reached
honestly) without pretending to know what dlc's checker will say. It
FAILS LOUDLY if any case does anything other than not_implemented --
including if one accidentally "passes", which would mean something is
silently faking a Lane A result.

Each of the four case files' representative cases was independently
cross-checked against Soufflé itself tonight (docs/reports/
night01-T7-harness.md) -- Soufflé rejects all four for the matching
reason (arity, type, ungrounded/allowedness, stratification), which is
evidence the hand-written programs are genuinely malformed, not typos.
"""
import importlib.util
import sys
import tempfile
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from differential import run_dlc  # noqa: E402

REJECTION_DIR = Path(__file__).resolve().parents[1] / "tests" / "rejection"
GROUNDS = ["arity", "type", "allowedness", "stratification"]


def load_cases(ground: str):
    spec = importlib.util.spec_from_file_location(ground, REJECTION_DIR / f"{ground}.py")
    mod = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mod)
    return mod.CASES


def main():
    total = 0
    unexpected = 0
    with tempfile.TemporaryDirectory() as tmp:
        tmp_path = Path(tmp)
        empty_facts = tmp_path / "facts"
        empty_facts.mkdir()
        for ground in GROUNDS:
            cases = load_cases(ground)
            print(f"=== {ground}: {len(cases)} cases ===")
            for case in cases:
                total += 1
                dl_path = tmp_path / f"{case['name']}.dl"
                dl_path.write_text(case["program"])
                result = run_dlc(dl_path, empty_facts)
                ok = result.status == "not_implemented"
                if not ok:
                    unexpected += 1
                print(f"  {case['name']}: status={result.status} "
                      f"{'OK (blocked on Lane A, as expected)' if ok else 'UNEXPECTED'}")
                assert case["expected_ground"] == ground, \
                    f"case {case['name']} filed under {ground} but expected_ground={case['expected_ground']}"

    print(f"\n{total} cases loaded across {len(GROUNDS)} grounds, "
          f"{total - unexpected} correctly blocked on not-implemented dlc.")
    if unexpected:
        print(f"{unexpected} cases did NOT report not_implemented -- "
              f"something is faking a Lane A result, this must not happen", file=sys.stderr)
        raise SystemExit(1)


if __name__ == "__main__":
    main()
