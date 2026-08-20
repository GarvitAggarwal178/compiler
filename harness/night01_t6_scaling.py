#!/usr/bin/env python3
"""
NIGHT-BATCH-01 T6: P2 scaling sweep, three columns (T_none, T_souffle,
T_guard) plus E_recoverable, at each n in {250,500,1000,2000,4000,8000}.
Ascending order so a cap truncates the top of the list, not the middle.
Density held at the existing P2 fixture's ratio (400 edges / 200 nodes =
2.0). Seeded deterministically per n, recorded.

Uses tests/programs/p4prime.dl unmodified as T_guard -- it already
generalizes to any edge/node fixture (no hardcoded size), that is the
whole point of moving the query constant into the magic seed only
(Phase 0.6).

Abort condition: q2.csv (sorted set-equality, CLAUDE.md section 6) not
identical across all three configurations at any n. Abort the whole
task immediately, do not continue to larger n.
"""
import json
import resource
import subprocess
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import fixtures_lib  # noqa: E402
from tuple_report import analyze as tuple_analyze  # noqa: E402

REPO = Path(__file__).resolve().parents[1]
PROGRAMS = REPO / "tests" / "programs"
FIXTURES = REPO / "fixtures"
MEASUREMENTS = REPO / "measurements" / "night01-t6"

DENSITY = 400 / 200  # edges per node, matches the existing P2 fixture
SIZES = [250, 500, 1000, 2000, 4000, 8000]
BASE_SEED = 20260820100  # distinct namespace from Phase 0/0.5/0.6 seeds

MEM_LIMIT_BYTES = 8 * 1024 * 1024 * 1024
TIMEOUT_S = 300


def _limit_mem():
    resource.setrlimit(resource.RLIMIT_AS, (MEM_LIMIT_BYTES, MEM_LIMIT_BYTES))


def build_fixture(n):
    seed = BASE_SEED + n
    target_edges = round(n * DENSITY)
    edges = fixtures_lib.gen_random_graph(seed, n=n, target_edges=target_edges)
    fdir = FIXTURES / f"p2-scale-{n}"
    fixtures_lib.write_facts(fdir / "edge.facts", edges)
    fixtures_lib.write_facts(fdir / "node.facts", [(i,) for i in range(1, n + 1)])
    meta = {"seed": seed, "n": n, "target_edges": target_edges, "actual_edges": len(edges)}
    (fdir / "meta.json").write_text(json.dumps(meta, indent=2) + "\n")
    return fdir, meta


def run(dl_name, facts_dir, workdir, magic, log_name="prof.log"):
    workdir.mkdir(parents=True, exist_ok=True)
    cmd = ["souffle", "-F", str(facts_dir), "-D", str(workdir), "-p", log_name]
    if magic:
        cmd.append("--magic-transform=*")
    cmd.append(str(PROGRAMS / dl_name))
    (workdir / "cmd.txt").write_text(" ".join(cmd) + "\n")
    try:
        proc = subprocess.run(
            cmd, cwd=str(workdir), capture_output=True, text=True,
            timeout=TIMEOUT_S, preexec_fn=_limit_mem,
        )
    except subprocess.TimeoutExpired:
        return {"status": "DNF:timeout-300s"}
    except MemoryError:
        return {"status": "DNF:memcap-8gb"}
    (workdir / "stdout.txt").write_text(proc.stdout)
    (workdir / "stderr.txt").write_text(proc.stderr)
    if proc.returncode != 0:
        return {"status": f"error:returncode-{proc.returncode}", "stderr": proc.stderr[:500]}
    return {"status": "ok", "workdir": workdir}


def sorted_lines(path: Path):
    return sorted(path.read_text().splitlines()) if path.is_file() else None


def main():
    results = []
    for n in SIZES:
        print(f"=== n={n} ===", file=sys.stderr)
        fdir, fmeta = build_fixture(n)
        row = {"n": n, "fixture": fmeta}

        r_none = run("p2.dl", fdir, MEASUREMENTS / f"n{n}-none", magic=False)
        r_souffle = run("p2.dl", fdir, MEASUREMENTS / f"n{n}-souffle", magic=True)
        r_guard = run("p4prime.dl", fdir, MEASUREMENTS / f"n{n}-guard", magic=False)

        for label, r in [("none", r_none), ("souffle", r_souffle), ("guard", r_guard)]:
            row[f"status_{label}"] = r["status"]
            print(f"  {label}: {r['status']}", file=sys.stderr)

        if not all(r["status"] == "ok" for r in (r_none, r_souffle, r_guard)):
            results.append(row)
            print(f"  n={n}: at least one config did not complete, stopping ascent here", file=sys.stderr)
            break

        q2_none = sorted_lines(r_none["workdir"] / "q2.csv")
        q2_souffle = sorted_lines(r_souffle["workdir"] / "q2.csv")
        q2_guard = sorted_lines(r_guard["workdir"] / "q2.csv")

        identical = q2_none == q2_souffle == q2_guard
        row["answers_identical"] = identical
        if not identical:
            row["q2_none_len"] = len(q2_none) if q2_none is not None else None
            row["q2_souffle_len"] = len(q2_souffle) if q2_souffle is not None else None
            row["q2_guard_len"] = len(q2_guard) if q2_guard is not None else None
            results.append(row)
            print(f"ABORT CONDITION at n={n}: q2.csv not identical across configs", file=sys.stderr)
            (MEASUREMENTS / "summary.json").write_text(json.dumps(results, indent=2))
            print(json.dumps({"aborted_at_n": n, "results": results}, indent=2))
            return

        tr_none = tuple_analyze(r_none["workdir"] / "prof.log")
        tr_souffle = tuple_analyze(r_souffle["workdir"] / "prof.log")
        tr_guard = tuple_analyze(r_guard["workdir"] / "prof.log")

        row["T_none"] = tr_none["T_excl_copy"]
        row["T_souffle"] = tr_souffle["T_excl_copy"]
        row["T_guard"] = tr_guard["T_excl_copy"]
        row["E_recoverable"] = tr_souffle["E_recoverable"]
        row["T_souffle_over_T_guard"] = (tr_souffle["T_excl_copy"] / tr_guard["T_excl_copy"]
                                          if tr_guard["T_excl_copy"] else None)
        results.append(row)
        (MEASUREMENTS / "summary.json").write_text(json.dumps(results, indent=2))
        print(f"  T_none={row['T_none']} T_souffle={row['T_souffle']} T_guard={row['T_guard']} "
              f"E_recoverable={row['E_recoverable']} ratio={row['T_souffle_over_T_guard']}", file=sys.stderr)

    (MEASUREMENTS / "summary.json").write_text(json.dumps(results, indent=2))
    print(json.dumps({"aborted_at_n": None, "results": results}, indent=2))


if __name__ == "__main__":
    main()
