#!/usr/bin/env python3
"""
Driver for tests/corpus/BENCHMARK_FAMILY/ (2026-08-22 ruling section 3.2).
NOT RUN as of the commit that introduces this file -- generating fixtures
and invoking Soufflé against this family is a future, explicitly
authorized step (ruling section 3.2 step 5: "Do not run"). This script
exists so that step is a single reviewable command, not something
improvised later.

When authorized, this would:
  1. For each shape in SCALE_POINTS.json, generate fixtures at every
     pre-registered scale point via harness/fixtures_lib.py (deterministic,
     seeded -- values are already committed above, not chosen at run time).
  2. Run untransformed, --magic-transform=*, and (where a hand-guard
     exists) the guarded form, exactly as harness/night01_t6_scaling.py
     does for the reachability_complement shape already.
  3. Record T_none/T_souffle/T_guard/E_recoverable via
     harness/tuple_report.py, with full cmd.txt/stdout.txt/meta.json
     provenance per measurement, same convention as every prior phase.

Not implemented past the fixture-generation stub below, deliberately --
implementing the Soufflé-invocation half now would make it too easy to
"just try it" before authorization.
"""
import json
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import fixtures_lib  # noqa: E402

REPO = Path(__file__).resolve().parents[1]
FAMILY_DIR = REPO / "tests" / "corpus" / "BENCHMARK_FAMILY"
SCALE_POINTS = json.loads((FAMILY_DIR / "SCALE_POINTS.json").read_text())


def generate_fixtures_only():
    """Materializes the deterministic fixture files for every
    pre-registered scale point. Does NOT invoke Soufflé. Safe to run
    (fixture generation is data materialization from an already-committed
    spec, not "running the benchmark") but not invoked automatically by
    this module -- call explicitly, e.g. from a REPL or a future task."""
    out = {}

    sg = SCALE_POINTS["same_generation_negation"]
    for pt in sg["points"]:
        seed = sg["seed_base"] + pt["depth"]
        edges, persons = fixtures_lib.gen_family_tree(seed, pt["depth"], pt["branching"])
        tag = f"d{pt['depth']}_b{pt['branching']}"
        fdir = REPO / "fixtures" / "benchmark-family" / "same_generation_negation" / tag
        fixtures_lib.write_facts(fdir / "parent.facts", edges)
        fixtures_lib.write_facts(fdir / "person.facts", [(p,) for p in persons])
        out[f"same_generation_negation/{tag}"] = {"seed": seed, "n_persons": len(persons), "n_edges": len(edges)}

    for shape_name in ("transitive_closure_bound", "ancestor_nonancestor"):
        spec = SCALE_POINTS[shape_name]
        for pt in spec["points"]:
            seed = spec["seed_base"] + pt["n"]
            edges = fixtures_lib.gen_core_rest_graph(seed, n=pt["n"], core_size=spec["core_size"],
                                                       target_edges=pt["target_edges"])
            tag = f"n{pt['n']}"
            fdir = REPO / "fixtures" / "benchmark-family" / shape_name / tag
            rel_name = "edge" if shape_name == "transitive_closure_bound" else "parent"
            fixtures_lib.write_facts(fdir / f"{rel_name}.facts", edges)
            if shape_name == "ancestor_nonancestor":
                fixtures_lib.write_facts(fdir / "person.facts", [(i,) for i in range(1, pt["n"] + 1)])
            out[f"{shape_name}/{tag}"] = {"seed": seed, "n": pt["n"], "n_edges": len(edges)}

    cc = SCALE_POINTS["culprit_cycle"]
    for pt in cc["points"]:
        seed = cc["seed_base"] + pt["n"]
        base_edges, e_edges, blocked = fixtures_lib.gen_culprit_cycle_facts(
            seed, n=pt["n"], target_base=pt["target_base"], target_e=pt["target_e"],
            blocked_fraction=cc["blocked_fraction"],
        )
        tag = f"n{pt['n']}"
        fdir = REPO / "fixtures" / "benchmark-family" / "culprit_cycle" / tag
        fixtures_lib.write_facts(fdir / "base.facts", base_edges)
        fixtures_lib.write_facts(fdir / "e.facts", e_edges)
        fixtures_lib.write_facts(fdir / "blocked.facts", blocked)
        out[f"culprit_cycle/{tag}"] = {"seed": seed, "n": pt["n"]}

    return out


def main():
    print("This driver is not invoked automatically. Import and call "
          "generate_fixtures_only() explicitly, or extend this module with "
          "the Soufflé-invocation half, when the benchmark family run is "
          "authorized. See this file's module docstring.", file=sys.stderr)
    raise SystemExit(1)


if __name__ == "__main__":
    main()
