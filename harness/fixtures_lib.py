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


def gen_family_tree(seed, depth=6, branching=4):
    """Deterministic family tree for the same-generation benchmark shape
    (tests/corpus/BENCHMARK_FAMILY, 2026-08-22 ruling section 3). Root is
    always named 0 so the bound query is stable across seeds/sizes.
    Returns (parent_edges, person_ids) where parent_edges are
    (child, parent) pairs and person_ids covers every node including the
    root. Branching factor jitters +-1 per node (seeded) so the tree is
    not perfectly regular, matching real family-tree irregularity without
    losing determinism."""
    rng = random.Random(seed)
    parent_edges = []
    persons = [0]
    frontier = [0]
    next_id = 1
    for _level in range(depth):
        new_frontier = []
        for node in frontier:
            b = max(1, branching + rng.randint(-1, 1))
            for _ in range(b):
                child = next_id
                next_id += 1
                parent_edges.append((child, node))
                persons.append(child)
                new_frontier.append(child)
        frontier = new_frontier
        if not frontier:
            break
    return parent_edges, persons


def gen_culprit_cycle_facts(seed, n=20, target_base=30, target_e=30, blocked_fraction=0.2):
    """Facts for the P5 culprit-cycle shape (constructed, not from a
    published source -- disclosed in tests/corpus/BENCHMARK_FAMILY/
    culprit_cycle.dl's header comment). Generalizes probe0.build_p3_fixture
    with a `blocked` relation and a controllable size."""
    rng = random.Random(seed)

    def rand_edges(count):
        edges, edge_set = [], set()
        attempts, max_attempts = 0, count * 20 + 2000
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
    if not any(a == 1 for a, _ in e_edges):
        e_edges.append((1, rng.randint(2, n)))
    blocked = [(i,) for i in range(1, n + 1) if rng.random() < blocked_fraction]
    return base_edges, e_edges, blocked


def gen_culprit_cycle_facts_labeled(seed, n=20, num_labels=3, target_base=30, target_e=30, blocked_fraction=0.2):
    """NIGHT-BATCH-03 T4: arity-3 variant of gen_culprit_cycle_facts, for
    the CULPRIT_CANDIDATES construction that varies arity (label is a
    third column on base/e; blocked gets a matching second column). Query
    connectivity is guaranteed at label 1 specifically (not just node 1)
    since the arity-3 candidate's query binds both the node and the label
    argument."""
    rng = random.Random(seed)

    def rand_labeled_edges(count):
        edges, edge_set = [], set()
        attempts, max_attempts = count * 20 + 2000, count * 20 + 2000
        tries = 0
        while len(edges) < count and tries < attempts:
            tries += 1
            a, b, l = rng.randint(1, n), rng.randint(1, n), rng.randint(1, num_labels)
            if a == b:
                continue
            key = (a, b, l)
            if key in edge_set:
                continue
            edge_set.add(key)
            edges.append(key)
        return edges

    base_edges = rand_labeled_edges(target_base)
    e_edges = rand_labeled_edges(target_e)
    if not any(a == 1 and l == 1 for a, _, l in e_edges):
        e_edges.append((1, rng.randint(2, n), 1))
    blocked = [(i, l) for i in range(1, n + 1) for l in range(1, num_labels + 1) if rng.random() < blocked_fraction]
    return base_edges, e_edges, blocked


def gen_cone_corpus_facts(seed, n=20, target_base=30, target_e=30,
                           blocked_fraction=0.2, gate_fraction=0.15,
                           gate_edges=20, sibling_edges=40):
    """NIGHT-BATCH-04 B: one generator serving all four
    tests/corpus/CONE_CORPUS/ programs -- they share the same
    culprit-cycle core (base/e/blocked, identical construction to
    gen_culprit_cycle_facts) and differ only in which of the extra
    relations below their own .decl/.input lines actually use. Returns a
    dict of relation name -> rows; each program's own fixture directory
    is written from the subset of keys its own .dl file declares.

    - base, e, blocked: the culprit-cycle core, identical construction to
      gen_culprit_cycle_facts (same rand_edges shape, same blocked_fraction
      convention).
    - gate_seed, gate_edge: feed the recursive `gate` IDB relation
      (cc_cone_only.dl, cc_both.dl) -- read from `p`'s non-recursive
      base-case rule (see cc_cone_only.dl's header comment for why that
      placement, not `s`'s rule, is required for a genuine cone).
    - chain_seed, chain_edge: a SECOND, independent recursive relation
      (`gate1b`, feeding `gate1`) for the two-hop cone variant
      (cc_cone_proper_subset.dl).
    - sibling_edge, sibling2_edge: plain edge relations for the `tc`/
      `direct` sibling branches (cc_sibling_emptycone.dl, cc_both.dl,
      cc_cone_proper_subset.dl) -- deliberately their own relation
      names, sharing nothing with the culprit core, so no dependency
      edge can accidentally form between a sibling branch and the
      declined one.
    """
    rng = random.Random(seed)

    def rand_edges(count):
        edges, edge_set = [], set()
        attempts, max_attempts = count * 20 + 2000, count * 20 + 2000
        tries = 0
        while len(edges) < count and tries < max_attempts:
            tries += 1
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
    if not any(a == 1 for a, _ in e_edges):
        e_edges.append((1, rng.randint(2, n)))
    blocked = [(i,) for i in range(1, n + 1) if rng.random() < blocked_fraction]

    gate_seed_rows = [(i,) for i in range(1, n + 1) if rng.random() < gate_fraction]
    gate_edge_rows = rand_edges(gate_edges)

    chain_seed_rows = [(i,) for i in range(1, n + 1) if rng.random() < gate_fraction]
    chain_edge_rows = rand_edges(gate_edges)

    sibling_edge_rows = rand_edges(sibling_edges)
    if not any(a == 1 for a, _ in sibling_edge_rows):
        sibling_edge_rows.append((1, rng.randint(2, n)))

    sibling2_edge_rows = rand_edges(sibling_edges)
    if not any(a == 1 for a, _ in sibling2_edge_rows):
        sibling2_edge_rows.append((1, rng.randint(2, n)))

    return {
        "base": base_edges, "e": e_edges, "blocked": blocked,
        "gate_seed": gate_seed_rows, "gate_edge": gate_edge_rows,
        "chain_seed": chain_seed_rows, "chain_edge": chain_edge_rows,
        "sibling_edge": sibling_edge_rows, "sibling2_edge": sibling2_edge_rows,
    }


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
