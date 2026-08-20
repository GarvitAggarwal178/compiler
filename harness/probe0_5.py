#!/usr/bin/env python3
"""
Phase 0.5: run P1' (P1 with `.output path` removed) and P4 (hand-transformed
P2), reusing the existing seeded P1/P2 fixtures unchanged -- no
regeneration, per the Phase 0.5 instructions. Same provenance shape as
probe0.py.

Lane B. Reuses probe0.py's run_cmd/run_souffle/run_profile helpers.
"""
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import probe0  # noqa: E402

REPO = probe0.REPO
FIXTURES = probe0.FIXTURES
PROGRAMS = probe0.PROGRAMS
MEASUREMENTS = probe0.MEASUREMENTS


def main():
    runs = [
        ("probe0.5-p1prime-off", PROGRAMS / "p1prime.dl", FIXTURES / "p1", "prof_off.log", False),
        ("probe0.5-p1prime-on", PROGRAMS / "p1prime.dl", FIXTURES / "p1", "prof_on.log", True),
        ("probe0.5-p4", PROGRAMS / "p4.dl", FIXTURES / "p2", "prof.log", False),
    ]

    for mid, dl_path, facts_dir, log_name, magic in runs:
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

    print("probe0.5 measurements complete", file=sys.stderr)


if __name__ == "__main__":
    main()
