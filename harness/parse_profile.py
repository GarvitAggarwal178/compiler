#!/usr/bin/env python3
"""
Parse a souffle -p profile log (JSON) and print exact per-relation
num-tuples. souffleprof's text tables round large counts (e.g. "1.52M");
the underlying log is JSON and carries exact integers under
root.program.relation.<name>.num-tuples.

Lane B: reads Soufflé's own output format, does not implement any
Datalog semantics.
"""
import json
import sys


def main():
    if len(sys.argv) != 2:
        print(f"usage: {sys.argv[0]} <profile.log>", file=sys.stderr)
        raise SystemExit(2)

    with open(sys.argv[1]) as f:
        doc = json.load(f)

    relations = doc["root"]["program"]["relation"]
    out = {}
    for name, rel in relations.items():
        base = rel.get("num-tuples", 0) or 0
        iterations = rel.get("iteration")
        if iterations:
            # Recursive relation: top-level "num-tuples" is only the
            # non-recursive (seed) rule's contribution. Each entry under
            # "iteration" carries the *new* tuples derived that semi-naive
            # round (a delta, not cumulative) -- summing every iteration's
            # delta plus the seed gives the total distinct tuples ever
            # inserted, which is exactly the blueprint's derived-tuple
            # metric under set semantics (each tuple inserted once).
            delta_sum = sum(v.get("num-tuples", 0) or 0 for v in iterations.values())
            out[name] = {"total": base + delta_sum, "seed": base, "delta_sum": delta_sum, "n_iterations": len(iterations)}
        else:
            out[name] = {"total": base}

    print(json.dumps(out, indent=2, sort_keys=True))


if __name__ == "__main__":
    main()
