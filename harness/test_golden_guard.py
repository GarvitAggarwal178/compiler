#!/usr/bin/env python3
"""
NIGHT-BATCH-01 T7 item 2: test for the golden-file guard. Plain
assertions, no external test framework (none is installed, and the
project has no other dependency on one) -- run directly:
    python3 harness/test_golden_guard.py
Exits 0 on success, nonzero (with the failing assertion's message) on
failure.
"""
import shutil
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from golden import generate_golden, GoldenGuardError, GOLDEN_ROOT  # noqa: E402

REPO = Path(__file__).resolve().parents[1]


def test_guard_refuses_non_souffle_command():
    threw = False
    try:
        generate_golden(
            "should-not-exist",
            REPO / "tests" / "programs" / "p2.dl",
            REPO / "fixtures" / "p2",
            ["dlc", "--fake-flag"],
        )
    except GoldenGuardError:
        threw = True
    assert threw, "generate_golden did not refuse a non-souffle generating command"
    assert not (GOLDEN_ROOT / "should-not-exist").exists(), \
        "guard refused but still created output on disk"


def test_guard_refuses_python_wrapper_pretending_to_be_souffle_output():
    """A command that merely contains 'souffle' as a substring but whose
    argv[0] basename isn't literally 'souffle' must still be refused."""
    threw = False
    try:
        generate_golden(
            "should-not-exist-2",
            REPO / "tests" / "programs" / "p2.dl",
            REPO / "fixtures" / "p2",
            ["not_souffle_at_all", "-x"],
        )
    except GoldenGuardError:
        threw = True
    assert threw, "guard did not refuse a command whose basename isn't 'souffle'"


def test_guard_allows_real_souffle():
    out_dir = generate_golden(
        "night01-t7-guard-smoke-test",
        REPO / "tests" / "programs" / "p2.dl",
        REPO / "fixtures" / "p2",
        ["souffle"],
    )
    assert out_dir.is_dir(), "generate_golden did not create the output directory"
    assert (out_dir / "q2.csv").is_file(), "expected output relation q2.csv missing"
    assert (out_dir / "GOLDEN_MANIFEST.json").is_file(), "manifest missing"
    # cleanup -- this is a smoke-test artifact, not a real golden file
    shutil.rmtree(out_dir)


def main():
    tests = [
        test_guard_refuses_non_souffle_command,
        test_guard_refuses_python_wrapper_pretending_to_be_souffle_output,
        test_guard_allows_real_souffle,
    ]
    failures = 0
    for t in tests:
        try:
            t()
            print(f"PASS {t.__name__}")
        except AssertionError as e:
            failures += 1
            print(f"FAIL {t.__name__}: {e}")
    if failures:
        print(f"{failures}/{len(tests)} failed", file=sys.stderr)
        raise SystemExit(1)
    print(f"all {len(tests)} passed")


if __name__ == "__main__":
    main()
