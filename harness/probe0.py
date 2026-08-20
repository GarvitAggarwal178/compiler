#!/usr/bin/env python3
"""
Phase 0 probe harness (Lane B).

Generates the seeded fixtures for P1/P2/P3, writes the three .dl programs
verbatim from docs/dlc-blueprint.md section 12, runs Soufflé with and
without --magic-transform=*, profiles each run with souffleprof, and
records every command + its raw output under measurements/<id>/.

No Lane A code here. This script does not implement Datalog evaluation;
it only drives an external oracle (Soufflé) and parses its own profiler
output.
"""

import json
import random
import subprocess
import sys
import time
from pathlib import Path

REPO = Path(__file__).resolve().parents[1]
FIXTURES = REPO / "fixtures"
PROGRAMS = REPO / "tests" / "programs"
MEASUREMENTS = REPO / "measurements"

# Every generator is seeded from a constant recorded here (CLAUDE.md #4).
SEEDS = {
    "p1_graph": 20260820001,
    "p2_graph": 20260820002,
    "p3_facts": 20260820003,
}

P1_DL = """\
.decl edge(a:number, b:number)
.input edge
.decl path(a:number, b:number)
.output path
path(x,y) :- edge(x,y).
path(x,y) :- path(x,z), edge(z,y).
.decl q(b:number)
.output q
q(y) :- path(1,y).
"""

P2_DL = """\
.decl edge(a:number, b:number)
.input edge
.decl node(a:number)
.input node
.decl reach(a:number, b:number)
reach(x,y) :- edge(x,y).
reach(x,y) :- reach(x,z), edge(z,y).
.decl unreach(a:number, b:number)
unreach(x,y) :- node(x), node(y), !reach(x,y).
.decl q2(b:number)
.output q2
q2(y) :- unreach(1,y).
"""

P3_DL = """\
.decl base(x:number, y:number)
.input base
.decl e(x:number, y:number)
.input e
.decl q(x:number, y:number)
.decl s(x:number)
.decl p(x:number, y:number)

q(x,y) :- base(x,y).
s(x)   :- q(x,_).
p(x,y) :- e(x,y).
p(x,y) :- p(x,z), !s(z), q(z,y).

.decl out(y:number)
.output out
out(y) :- p(1,y).
"""


def write_dl_programs():
    PROGRAMS.mkdir(parents=True, exist_ok=True)
    (PROGRAMS / "p1.dl").write_text(P1_DL)
    (PROGRAMS / "p2.dl").write_text(P2_DL)
    (PROGRAMS / "p3.dl").write_text(P3_DL)


def write_facts(path: Path, rows):
    path.parent.mkdir(parents=True, exist_ok=True)
    with open(path, "w", newline="\n") as f:
        for row in rows:
            f.write("\t".join(str(x) for x in row) + "\n")


def bfs_reachable(edges, source):
    adj = {}
    for a, b in edges:
        adj.setdefault(a, []).append(b)
    seen = {source}
    stack = [source]
    while stack:
        u = stack.pop()
        for v in adj.get(u, ()):
            if v not in seen:
                seen.add(v)
                stack.append(v)
    return seen


def gen_p1_graph(seed, n=2000, core_size=50, target_edges=4000):
    """
    Directed graph on nodes 1..n. Node 1's reachable set is exactly the
    `core` (nodes 1..core_size): core nodes only ever point at other core
    nodes, so reachability from node 1 cannot leak past core_size. The
    remaining ~1950 nodes carry the bulk of the ~4000 edges (rest->rest,
    rest->core) to hit the target edge count without touching
    reachability from node 1.
    """
    rng = random.Random(seed)
    core = list(range(1, core_size + 1))
    rest = list(range(core_size + 1, n + 1))

    edges = []
    edge_set = set()

    # Spanning structure: random topological order over core with node 1
    # first, each later node gets one edge from a random earlier node.
    # This guarantees every core node is reachable from node 1.
    order = core[1:]
    rng.shuffle(order)
    order = [1] + order
    for i in range(1, len(order)):
        node = order[i]
        pred = order[rng.randint(0, i - 1)]
        e = (pred, node)
        if e not in edge_set:
            edge_set.add(e)
            edges.append(e)

    # Extra core->core edges for density; never core->rest.
    extra_core_target = min(100, core_size * (core_size - 1) // 2)
    attempts = 0
    core_edge_count = lambda: sum(1 for a, b in edges if a in core and b in core)
    while core_edge_count() < len(order) - 1 + extra_core_target and attempts < 5000:
        attempts += 1
        a, b = rng.choice(core), rng.choice(core)
        if a == b:
            continue
        e = (a, b)
        if e in edge_set:
            continue
        edge_set.add(e)
        edges.append(e)

    # Fill remaining edges from `rest`, never rest -> nothing that grows
    # reachability from 1 is a risk only via core -> rest, which we never
    # emit. rest -> core and rest -> rest are both safe.
    remaining = target_edges - len(edges)
    attempts = 0
    max_attempts = remaining * 20 + 10000
    while remaining > 0 and attempts < max_attempts:
        attempts += 1
        a = rng.choice(rest)
        b = rng.choice(rest) if rng.random() < 0.8 else rng.choice(core)
        if a == b:
            continue
        e = (a, b)
        if e in edge_set:
            continue
        edge_set.add(e)
        edges.append(e)
        remaining -= 1

    return edges


def build_p1_fixture():
    base_seed = SEEDS["p1_graph"]
    attempt = 0
    while True:
        seed = base_seed + attempt
        edges = gen_p1_graph(seed)
        reach = bfs_reachable(edges, 1)
        if 45 <= len(reach) <= 55:
            break
        attempt += 1
        if attempt > 20:
            raise SystemExit(
                f"p1 fixture: reachable-set size did not land in [45,55] "
                f"after {attempt} attempts (last size={len(reach)})"
            )
    write_facts(FIXTURES / "p1" / "edge.facts", edges)
    meta = {
        "seed": seed,
        "attempts": attempt + 1,
        "n_nodes": 2000,
        "n_edges": len(edges),
        "reachable_from_1": len(reach),
    }
    (FIXTURES / "p1").mkdir(parents=True, exist_ok=True)
    (FIXTURES / "p1" / "meta.json").write_text(json.dumps(meta, indent=2) + "\n")
    return meta


def build_p2_fixture(n=200, target_edges=400):
    seed = SEEDS["p2_graph"]
    rng = random.Random(seed)
    edges = []
    edge_set = set()
    attempts = 0
    max_attempts = target_edges * 20 + 5000
    while len(edges) < target_edges and attempts < max_attempts:
        attempts += 1
        a, b = rng.randint(1, n), rng.randint(1, n)
        if a == b:
            continue
        e = (a, b)
        if e in edge_set:
            continue
        edge_set.add(e)
        edges.append(e)
    write_facts(FIXTURES / "p2" / "edge.facts", edges)
    write_facts(FIXTURES / "p2" / "node.facts", [(i,) for i in range(1, n + 1)])
    meta = {"seed": seed, "n_nodes": n, "n_edges": len(edges)}
    (FIXTURES / "p2" / "meta.json").write_text(json.dumps(meta, indent=2) + "\n")
    return meta


def build_p3_fixture(n=20, target_base=30, target_e=30):
    seed = SEEDS["p3_facts"]
    rng = random.Random(seed)

    def rand_edges(count):
        edges = []
        edge_set = set()
        attempts = 0
        max_attempts = count * 20 + 2000
        while len(edges) < count and attempts < max_attempts:
            attempts += 1
            a, b = rng.randint(1, n), rng.randint(1, n)
            if a == b:
                continue
            e = (a, b)
            if e in edge_set:
                continue
            edge_set.add(e)
            edges.append(e)
        return edges

    base_edges = rand_edges(target_base)
    e_edges = rand_edges(target_e)
    # Guarantee node 1 has at least one outgoing e-edge so p(1,_) is
    # non-empty in the base case, otherwise `out` is trivially empty and
    # the recursive/negated rule never fires.
    if not any(a == 1 for a, _ in e_edges):
        e_edges.append((1, rng.randint(2, n)))

    write_facts(FIXTURES / "p3" / "base.facts", base_edges)
    write_facts(FIXTURES / "p3" / "e.facts", e_edges)
    meta = {
        "seed": seed,
        "n_nodes": n,
        "n_base_edges": len(base_edges),
        "n_e_edges": len(e_edges),
    }
    (FIXTURES / "p3").mkdir(parents=True, exist_ok=True)
    (FIXTURES / "p3" / "meta.json").write_text(json.dumps(meta, indent=2) + "\n")
    return meta


def run_cmd(mid: str, cmd, cwd: Path, extra_meta=None):
    """Run `cmd`, capture everything, write full provenance under
    measurements/<mid>/, and return the CompletedProcess."""
    outdir = MEASUREMENTS / mid
    outdir.mkdir(parents=True, exist_ok=True)

    (outdir / "cmd.txt").write_text(" ".join(cmd) + "\n")

    t0 = time.time()
    proc = subprocess.run(cmd, cwd=str(cwd), capture_output=True, text=True)
    dt = time.time() - t0

    (outdir / "stdout.txt").write_text(proc.stdout)
    (outdir / "stderr.txt").write_text(proc.stderr)

    ver = subprocess.run(["souffle", "--version"], capture_output=True, text=True)
    env_lines = [
        f"cwd: {cwd}",
        f"souffle --version:",
        ver.stdout.strip(),
        f"wall_seconds_not_a_metric: {dt:.3f}",
    ]
    (outdir / "env.txt").write_text("\n".join(env_lines) + "\n")

    meta = {
        "id": mid,
        "cmd": cmd,
        "cwd": str(cwd),
        "returncode": proc.returncode,
    }
    if extra_meta:
        meta.update(extra_meta)
    (outdir / "meta.json").write_text(json.dumps(meta, indent=2) + "\n")

    return proc


def run_souffle(mid, dl_path: Path, facts_dir: Path, workdir: Path, log_name: str, magic: bool):
    workdir.mkdir(parents=True, exist_ok=True)
    cmd = ["souffle", "-F", str(facts_dir), "-D", str(workdir), "-p", log_name]
    if magic:
        cmd.append("--magic-transform=*")
    cmd.append(str(dl_path))
    extra = {"log_file": str(workdir / log_name), "magic": magic}
    return run_cmd(mid, cmd, workdir, extra_meta=extra)


def run_profile(mid, log_path: Path, workdir: Path):
    cmd = ["souffleprof", str(log_path), "-c", "rel"]
    extra = {"profiles": str(log_path)}
    return run_cmd(mid, cmd, workdir, extra_meta=extra)


def main():
    write_dl_programs()

    p1meta = build_p1_fixture()
    p2meta = build_p2_fixture()
    p3meta = build_p3_fixture()

    print("fixtures:", file=sys.stderr)
    print("  p1:", p1meta, file=sys.stderr)
    print("  p2:", p2meta, file=sys.stderr)
    print("  p3:", p3meta, file=sys.stderr)

    if not (45 <= p1meta["reachable_from_1"] <= 55):
        raise SystemExit("REFUSING TO RUN: p1 reachable-set verification failed")

    runs = [
        ("probe0-p1-off", PROGRAMS / "p1.dl", FIXTURES / "p1", "prof_off.log", False),
        ("probe0-p1-on", PROGRAMS / "p1.dl", FIXTURES / "p1", "prof_on.log", True),
        ("probe0-p2-off", PROGRAMS / "p2.dl", FIXTURES / "p2", "prof_off.log", False),
        ("probe0-p2-on", PROGRAMS / "p2.dl", FIXTURES / "p2", "prof_on.log", True),
        ("probe0-p3-off", PROGRAMS / "p3.dl", FIXTURES / "p3", "prof_off.log", False),
        ("probe0-p3-on", PROGRAMS / "p3.dl", FIXTURES / "p3", "prof_on.log", True),
    ]

    for mid, dl_path, facts_dir, log_name, magic in runs:
        run_id = f"{mid}-run"
        workdir = MEASUREMENTS / run_id
        proc = run_souffle(run_id, dl_path, facts_dir, workdir, log_name, magic)
        print(f"{run_id}: returncode={proc.returncode}", file=sys.stderr)
        if proc.returncode != 0:
            print(proc.stderr, file=sys.stderr)
            raise SystemExit(f"souffle run failed: {run_id}")

        prof_id = f"{mid}-profile"
        proc2 = run_profile(prof_id, workdir / log_name, workdir)
        print(f"{prof_id}: returncode={proc2.returncode}", file=sys.stderr)
        if proc2.returncode != 0:
            print(proc2.stderr, file=sys.stderr)
            raise SystemExit(f"souffleprof failed: {prof_id}")

    print("all probe0 measurements complete", file=sys.stderr)


if __name__ == "__main__":
    main()
