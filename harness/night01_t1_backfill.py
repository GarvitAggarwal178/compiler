#!/usr/bin/env python3
"""
NIGHT-BATCH-01 T1: backfill provenance gaps found by the audit. Each
command here reproduces a check that was already reported (in
docs/MEASUREMENTS.md or docs/reports/probe0*.md) but never captured
through probe0.run_cmd, or was captured only partially by hand.

Lane B. Re-runs real commands; does not fabricate any output.
"""
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import probe0  # noqa: E402

REPO = probe0.REPO
M = REPO / "measurements"


def main():
    # --- Phase 0 diff checks, never captured ---
    probe0.run_cmd(
        "probe0-p1-diff",
        ["diff", "-q",
         str(M / "probe0-p1-off-run" / "path.csv"), str(M / "probe0-p1-on-run" / "path.csv")],
        REPO,
    )
    probe0.run_cmd(
        "probe0-p1-diff-q",
        ["diff", "-q",
         str(M / "probe0-p1-off-run" / "q.csv"), str(M / "probe0-p1-on-run" / "q.csv")],
        REPO,
    )
    probe0.run_cmd(
        "probe0-p2-diff",
        ["diff", "-q",
         str(M / "probe0-p2-off-run" / "q2.csv"), str(M / "probe0-p2-on-run" / "q2.csv")],
        REPO,
    )
    probe0.run_cmd(
        "probe0-p3-diff",
        ["diff", "-q",
         str(M / "probe0-p3-off-run" / "out.csv"), str(M / "probe0-p3-on-run" / "out.csv")],
        REPO,
    )

    # --- P1 fixture verification (reachable_from_1 check), never captured
    # as its own measurement dir -- re-verify via a standalone BFS script
    # rather than re-running the full generator (which would re-derive
    # the same fixture; the check itself is what needed provenance).
    probe0.run_cmd(
        "probe0-p1-fixture-bfs-check",
        ["python3", "harness/night01_bfs_check.py", "fixtures/p1/edge.facts", "1"],
        REPO,
    )

    # --- P2 BFS cross-check (Phase 0.5 §3), only run ad hoc before ---
    probe0.run_cmd(
        "probe0.5-p2-bfs-check",
        ["python3", "harness/night01_bfs_check.py", "fixtures/p2/edge.facts", "1"],
        REPO,
    )

    # --- P4' diff checks (Phase 0.6), only run ad hoc before ---
    probe0.run_cmd(
        "probe0.6-p4prime-diff-vs-p4",
        ["diff", "-q",
         str(M / "probe0.6-p4prime-run" / "q2.csv"), str(M / "probe0.5-p4-run" / "q2.csv")],
        REPO,
    )
    probe0.run_cmd(
        "probe0.6-p4prime-diff-vs-p2",
        ["diff", "-q",
         str(M / "probe0.6-p4prime-run" / "q2.csv"), str(M / "probe0-p2-off-run" / "q2.csv")],
        REPO,
    )

    print("backfill complete", file=sys.stderr)


if __name__ == "__main__":
    main()
