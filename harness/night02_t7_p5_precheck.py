#!/usr/bin/env python3
"""
NIGHT-BATCH-02 T7: P5 inlining prerequisite. OPEN_QUESTIONS.md records the
argument that P5's self-recursive `q` will survive Soufflé's inliner where
P3's non-recursive `q` did not -- argued, never verified until now.

Check 1 (|p| > |e|, does the second p rule fire) is evaluated across ALL
5 of T4's already-committed untransformed ("none") profiles
(measurements/night02-t4/culprit_cycle/n{20,50,100,200,500}-none/prof.log)
-- no new Soufflé run needed for this part. First attempt at this task
used only n=20 (T4's smallest point) and found |p|==|e| there (rule looks
dead) -- checking the other 4 points immediately after showed |p|>|e|
clearly at every one of them (n=50: 90 vs 76; n=100: 270 vs 150; n=200:
475 vs 300; n=500: 819 vs 750). n=20 is very likely a small-fixture
coincidence (few edges, 20% of nodes blocked -> every 1-hop extension
from node 1 happens to be blocked by chance at that size), not a
structural dead-rule/inliner problem -- but it is flagged, not silently
discarded, since a single anomalous point is exactly the kind of thing a
skeptic should be told about directly.

Checks 2 and 3 need actual --magic-transform=* / --inline-exclude=q
Soufflé runs (T4 never ran with --inline-exclude), done fresh here at
n=200 -- a mid-range point, deliberately not n=20 given the check-1
anomaly there and not n=500 given no particular reason to prefer the
largest either:
  2. Does `q` appear as a materialized relation in the profile (both with
     and without --inline-exclude=q)?
  3. Does the culprit cycle form -- is any relation left untransformed
     (unbound, @neglabel-isolated) the way negation forces in this
     project's differentiator?

Does not edit culprit_cycle.dl (pre-registered, prohibited) even where a
per-point anomaly is found -- report, don't fix the shape.
"""
import json
import subprocess
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from tuple_report import analyze as tuple_analyze  # noqa: E402

REPO = Path(__file__).resolve().parents[1]
DL_PATH = REPO / "tests" / "corpus" / "BENCHMARK_FAMILY" / "culprit_cycle.dl"
T4_MEASUREMENTS = REPO / "measurements" / "night02-t4" / "culprit_cycle"
CHECK1_SCALE_POINTS = [20, 50, 100, 200, 500]
CHECK23_N = 200
FACTS_DIR = REPO / "fixtures" / "benchmark-family" / "culprit_cycle" / f"n{CHECK23_N}"
MEASUREMENTS = REPO / "measurements" / "night02-t7"

TIMEOUT_S = 300


def run(workdir, extra_args, log_name="prof.log"):
    workdir.mkdir(parents=True, exist_ok=True)
    cmd = ["souffle", "-F", str(FACTS_DIR), "-D", str(workdir), "-p", log_name,
           "--magic-transform=*", *extra_args, str(DL_PATH)]
    (workdir / "cmd.txt").write_text(" ".join(cmd) + "\n")
    proc = subprocess.run(cmd, cwd=str(workdir), capture_output=True,
                           encoding="utf-8", errors="replace", timeout=TIMEOUT_S)
    (workdir / "stdout.txt").write_text(proc.stdout)
    (workdir / "stderr.txt").write_text(proc.stderr)
    return proc


def relation_names_containing(per_relation, needle):
    return sorted(k for k in per_relation if needle in k)


def main():
    result = {}

    # Check 1: does the second p rule fire (|p| > |e|)? Across all 5 of
    # T4's already-committed untransformed profiles -- no new run.
    check1 = {}
    for n in CHECK1_SCALE_POINTS:
        prof = T4_MEASUREMENTS / f"n{n}-none" / "prof.log"
        tr = tuple_analyze(prof)
        pr = {k: v["total"] for k, v in tr["per_relation"].items()}
        p_total, e_total = pr.get("p"), pr.get("e")
        check1[str(n)] = {
            "p_total": p_total, "e_total": e_total,
            "p_gt_e": (p_total is not None and e_total is not None and p_total > e_total),
        }
    result["check1_per_scale_point"] = check1
    n_fires = sum(1 for v in check1.values() if v["p_gt_e"])
    result["check1_fires_at_n_points"] = f"{n_fires}/{len(check1)}"
    result["check1_dead_at_smallest_only"] = (not check1[str(CHECK1_SCALE_POINTS[0])]["p_gt_e"]
                                               and n_fires == len(check1) - 1)
    # Task instruction: if dead (interpreted as: dead at EVERY point, not
    # firing at all), stop here without checks 2/3. It fires at 4/5
    # points, so this is not the P3 failure mode -- continue.
    stop_here = (n_fires == 0)
    result["stopped_after_check_1"] = stop_here

    if not stop_here:
        # Checks 2 and 3 need fresh --magic-transform=* runs (T4 never
        # ran --inline-exclude), done at n=200 (see module docstring for
        # why not n=20 or n=500).
        wd_default = MEASUREMENTS / f"n{CHECK23_N}-default"
        proc1 = run(wd_default, [])
        result["check23_default_returncode"] = proc1.returncode
        tr1 = tuple_analyze(wd_default / "prof.log")
        result["check23_default_per_relation"] = {k: v["total"] for k, v in tr1["per_relation"].items()}
        result["check23_default_q_relations"] = relation_names_containing(tr1["per_relation"], "q")

        wd_excl = MEASUREMENTS / f"n{CHECK23_N}-inline-exclude-q"
        proc2 = run(wd_excl, ["--inline-exclude=q"])
        result["check23_excl_returncode"] = proc2.returncode
        tr2 = tuple_analyze(wd_excl / "prof.log")
        result["check23_excl_per_relation"] = {k: v["total"] for k, v in tr2["per_relation"].items()}
        result["check23_excl_q_relations"] = relation_names_containing(tr2["per_relation"], "q")

        # Check 2: is q materialized (present as a real relation, not
        # inlined away) in both runs?
        result["q_materialized_default"] = len(result["check23_default_q_relations"]) > 0
        result["q_materialized_with_exclude"] = len(result["check23_excl_q_relations"]) > 0

        # Check 3: culprit cycle forms -- any relation left unbound
        # (isolated under @neglabel, this project's own marker for "the
        # negated relation was not restricted").
        result["neglabel_relations_default"] = tr1["neglabel_relations"]
        result["culprit_cycle_forms"] = bool(tr1["neglabel_relations"])

    out_path = MEASUREMENTS / "summary.json"
    MEASUREMENTS.mkdir(parents=True, exist_ok=True)
    out_path.write_text(json.dumps(result, indent=2))
    print(json.dumps(result, indent=2))


if __name__ == "__main__":
    main()
