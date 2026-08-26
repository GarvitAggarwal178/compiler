#!/usr/bin/env python3
"""
NIGHT-BATCH-03 T9: fallback cone metric. FALLBACK is downward-closed over
the full dependency relation (positive and negative edges alike) --
M2-M3-BUILD.md section 7's decision, this is the independent Lane B
measurement of it: given a program and a declined set of SCCs, what is the
downward dependency closure (the "cone"), as a plain graph query.

The DECISION of which SCCs to decline is Lane A/guard territory
(src/transform/guard/decide.go once it exists); this module only answers
"given a declined set, what does it drag down with it" -- a pure function
of the program's dependency graph, reusable regardless of how the declined
set was chosen.

Reuses harness/night03_t4_culprit_classifier.py's structural parser
(parse_structure/build_graphs) for the dependency graph -- same
regex-based, disclosed-approximate extraction, not a second parser.
"""
import json
import sys
from pathlib import Path
from collections import defaultdict

sys.path.insert(0, str(Path(__file__).resolve().parent))
from night03_t4_culprit_classifier import parse_structure, build_graphs  # noqa: E402


def tarjan_scc(nodes, all_edges):
    """Standard Tarjan SCC over the FULL dependency graph (all_edges --
    positive and negative together), matching sema/stratify.go's own SCC
    computation, which also uses the full (not positive-only) edge set."""
    index_counter = [0]
    indices, lowlink, on_stack, stack = {}, {}, {}, []
    scc_of, scc_members = {}, []

    def strongconnect(v):
        indices[v] = index_counter[0]
        lowlink[v] = index_counter[0]
        index_counter[0] += 1
        stack.append(v)
        on_stack[v] = True
        for w in sorted(all_edges.get(v, ())):
            if w not in indices:
                strongconnect(w)
                lowlink[v] = min(lowlink[v], lowlink[w])
            elif on_stack.get(w):
                lowlink[v] = min(lowlink[v], indices[w])
        if lowlink[v] == indices[v]:
            members = []
            while True:
                w = stack.pop()
                on_stack[w] = False
                scc_of[w] = len(scc_members)
                members.append(w)
                if w == v:
                    break
            scc_members.append(sorted(members))

    for v in sorted(nodes):
        if v not in indices:
            strongconnect(v)
    return scc_of, scc_members


def cone_size(idb_relations, all_edges, declined_sccs):
    """declined_sccs: list of SCCs (each an iterable of relation names)
    that the guard declined. Returns the spec'd dict exactly:
    {declined_sccs, cone_relations, cone_size, cone_fraction}.
    cone_relations is the downward closure REACHED FROM the declined set
    (i.e. what gets dragged down with it), not including the declined
    relations themselves unless a declined relation is also reachable from
    a different declined relation (harmless double-membership)."""
    declined_relations = set()
    for scc in declined_sccs:
        declined_relations |= set(scc)

    cone = set()
    frontier = list(declined_relations)
    visited = set(declined_relations)
    while frontier:
        r = frontier.pop()
        for dep in all_edges.get(r, ()):
            if dep not in visited:
                visited.add(dep)
                cone.add(dep)
                frontier.append(dep)

    return {
        "declined_sccs": len(declined_sccs),
        "declined_relations": sorted(declined_relations),
        "cone_relations": sorted(cone),
        "cone_size": len(cone),
        "cone_fraction": (len(cone) / len(idb_relations)) if idb_relations else 0.0,
    }


def analyze_file(text):
    """Convenience wrapper: parse a .dl file's text, return (idb_relations,
    all_edges, scc_of, scc_members) for the caller to pick a declined set
    from."""
    heads, rules = parse_structure(text)
    _pos_edges, all_edges = build_graphs(heads, rules)
    scc_of, scc_members = tarjan_scc(heads, all_edges)
    return heads, all_edges, scc_of, scc_members


def main():
    if len(sys.argv) != 3:
        print(f"usage: {sys.argv[0]} <file.dl> <declined_relation_name>", file=sys.stderr)
        raise SystemExit(2)
    path, declined_name = Path(sys.argv[1]), sys.argv[2]
    text = path.read_text()
    heads, all_edges, scc_of, scc_members = analyze_file(text)
    declined_scc = scc_members[scc_of[declined_name]]
    result = cone_size(heads, all_edges, [declined_scc])
    print(json.dumps(result, indent=2))


if __name__ == "__main__":
    main()
