#!/usr/bin/env python3
"""
Phase 0.6: P4' (fixed hand-transform) and the P6 counterexample-hunt
programs. Reuses probe0.py's run helpers and the existing P2 fixture
unchanged (edge.facts, node.facts) -- no new fixtures needed.

Lane B.
"""
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import probe0  # noqa: E402

PROGRAMS = probe0.PROGRAMS
FIXTURES = probe0.FIXTURES
MEASUREMENTS = probe0.MEASUREMENTS


def run_one(mid, dl_name, facts_dir, log_name="prof.log", magic=False):
    dl_path = PROGRAMS / dl_name
    run_id = f"{mid}-run"
    workdir = MEASUREMENTS / run_id
    proc = probe0.run_souffle(run_id, dl_path, facts_dir, workdir, log_name, magic)
    print(f"{run_id}: returncode={proc.returncode}", file=sys.stderr)
    if proc.returncode != 0:
        print(proc.stderr, file=sys.stderr)
        raise SystemExit(f"souffle run failed: {run_id}")

    prof_id = f"{mid}-profile"
    proc2 = probe0.run_profile(prof_id, workdir / log_name, workdir)
    print(f"{prof_id}: returncode={proc2.returncode}", file=sys.stderr)
    if proc2.returncode != 0:
        print(proc2.stderr, file=sys.stderr)
        raise SystemExit(f"souffleprof failed: {prof_id}")
    return workdir


def main():
    p2 = FIXTURES / "p2"

    run_one("probe0.6-p4prime", "p4prime.dl", p2)

    run_one("probe0.6-p6start-base", "p6start_base.dl", p2)
    run_one("probe0.6-p6start-hand", "p6start_hand.dl", p2)

    run_one("probe0.6-p6a1-base", "p6a1_base.dl", p2)
    run_one("probe0.6-p6a1-hand-naive", "p6a1_hand_naive.dl", p2)
    run_one("probe0.6-p6a1-hand", "p6a1_hand.dl", p2)

    run_one("probe0.6-p6a2-base", "p6a2_base.dl", p2)
    run_one("probe0.6-p6a2-hand", "p6a2_hand.dl", p2)

    run_one("probe0.6-p6a1b-base", "p6a1b_base.dl", p2)
    run_one("probe0.6-p6a1b-hand-naive", "p6a1b_hand_naive.dl", p2)
    run_one("probe0.6-p6a1b-hand", "p6a1b_hand.dl", p2)

    print("probe0.6 measurements complete", file=sys.stderr)


if __name__ == "__main__":
    main()
