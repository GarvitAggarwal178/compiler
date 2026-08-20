#!/usr/bin/env python3
"""Plain BFS over an edge.facts file from a given source node. Used to
independently cross-check reachable-set sizes reported elsewhere
(Soufflé JSON profile logs), per CLAUDE.md's determinism/provenance
discipline. No Soufflé, no dlc -- a completely separate implementation
of the same graph-reachability question."""
import sys


def main():
    if len(sys.argv) != 3:
        print(f"usage: {sys.argv[0]} <edge.facts> <source>", file=sys.stderr)
        raise SystemExit(2)

    edge_path, source = sys.argv[1], int(sys.argv[2])
    adj = {}
    with open(edge_path) as f:
        for line in f:
            a, b = line.split()
            adj.setdefault(int(a), []).append(int(b))

    seen = {source}
    stack = [source]
    while stack:
        u = stack.pop()
        for v in adj.get(u, ()):
            if v not in seen:
                seen.add(v)
                stack.append(v)

    print(f"reachable_from_{source}_incl_self={len(seen)}")


if __name__ == "__main__":
    main()
