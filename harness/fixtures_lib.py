#!/usr/bin/env python3
"""
Reusable fixture-generation library, factored out of probe0.py (T7 item
4). Two generator shapes:

- gen_core_rest_graph: probe0.gen_p1_graph's construction, generalized.
  Gives exact control over the reachable-from-node-1 set size (the core)
  while still hitting a target total edge count via a much larger `rest`
  partition. Use this whenever a fixture needs a *specific* reachable-set
  size -- the Phase 0.5 lesson (docs/OPEN_QUESTIONS.md, 2026-08-20) is
  that plain random digraphs give no such control.

- gen_random_graph: probe0.build_p2_fixture's construction, generalized.
  Plain uniform-random directed edges, no reachability control. Use this
  when only density/scale matters, not a specific reachable-set size
  (e.g. NIGHT-BATCH-01 T6's scaling sweep).

Both are deterministic given (seed, n, ...): same inputs always produce
the same edge list.
"""
import random


def gen_core_rest_graph(seed, n=2000, core_size=50, target_edges=4000):
    rng = random.Random(seed)
    core = list(range(1, core_size + 1))
    rest = list(range(core_size + 1, n + 1))

    edges = []
    edge_set = set()

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

    extra_core_target = min(100, core_size * (core_size - 1) // 2)
    attempts = 0
    core_edge_count = lambda: sum(1 for a, b in edges if a in core and b in core)  # noqa: E731
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

    remaining = target_edges - len(edges)
    attempts = 0
    max_attempts = remaining * 20 + 10000
    while remaining > 0 and attempts < max_attempts and rest:
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


def gen_random_graph(seed, n=200, target_edges=400):
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
    return edges


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


def write_facts(path, rows):
    path.parent.mkdir(parents=True, exist_ok=True)
    with open(path, "w", newline="\n") as f:
        for row in rows:
            f.write("\t".join(str(x) for x in row) + "\n")
