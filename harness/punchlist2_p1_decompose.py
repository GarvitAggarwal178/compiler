#!/usr/bin/env python3
"""
PUNCH-LIST-2 item 1: decompose T_guarded into its declined portion and
its transformed portion, per program per scale point, for the 3
sibling-bearing task-B constructions. Confirms or refutes the hypothesis
that the declined (full-extent) cone grows with n while the transformed
(demand-restricted) sibling stays roughly constant, so the ratio
T_none/T_guarded tends to 1.0 as n grows.

Reads already-committed prof.log files (measurements/night04-b-cone/) --
computes nothing new about the transform itself, only re-groups already-
measured per-relation totals.
"""
import json
import sys
from pathlib import Path

sys.path.insert(0, "/root/compiler/harness")
from tuple_report import analyze as tuple_analyze  # noqa: E402

REPO = Path("/root/compiler")
MEAS = REPO / "measurements" / "night04-b-cone"
OUT = REPO / "measurements" / "punch-list-2" / "p1-decompose"

# Declined predicate sets per program (culprit ∪ cone), from
# bin/conecheck's already-committed output -- declined predicates pass
# through as their ORIGINAL bare relation name (never adorned/magic'd),
# so classification is exact, not a heuristic.
DECLINED = {
    "cc_sibling_emptycone": {"p", "q", "s"},
    "cc_both": {"p", "q", "s", "gate"},
    "cc_cone_proper_subset": {"p", "q", "s", "gate1", "gate1b"},
}
# All predicate names appearing anywhere in each program (for
# longest-match prefix stripping on magic_/sup_/adorned relation names).
ALL_PREDS = {
    "cc_sibling_emptycone": ["p", "q", "s", "tc", "out", "out2"],
    "cc_both": ["gate", "p", "q", "s", "tc", "out", "out2"],
    "cc_cone_proper_subset": ["gate1b", "gate1", "p", "q", "s", "tc", "direct", "out", "out2", "out3"],
}
SCALE_POINTS = [20, 50, 100]


def origin_of(name: str, preds: list) -> str:
    """Longest-match prefix strip: 'sup_gate1b_b_r0_1' -> 'gate1b',
    'magic_tc_bf' -> 'tc', 'p_bf' -> 'p', bare 'p' -> 'p'. preds must be
    sorted longest-first by the caller."""
    body = name
    if body.startswith("magic_"):
        body = body[len("magic_"):]
    elif body.startswith("sup_"):
        body = body[len("sup_"):]
    for p in preds:
        if body == p or body.startswith(p + "_"):
            return p
    return None  # EDB input or unrecognized -- caller excludes it anyway


def decompose(program: str, n: int, run: str = "dlc"):
    preds_longest_first = sorted(ALL_PREDS[program], key=len, reverse=True)
    declined = DECLINED[program]
    prof_path = MEAS / program / f"n{n}" / run / "prof.log"
    prof = tuple_analyze(prof_path)
    declined_total = 0
    transformed_total = 0
    unclassified = {}
    for name, info in prof["per_relation"].items():
        if info["is_input"] or info["is_copy"]:
            continue
        origin = origin_of(name, preds_longest_first)
        if origin is None:
            unclassified[name] = info["total"]
            continue
        if origin in declined:
            declined_total += info["total"]
        else:
            transformed_total += info["total"]
    return {
        "program": program, "n": n,
        "T_guarded": prof["T_excl_copy"],
        "declined_portion": declined_total,
        "transformed_portion": transformed_total,
        "unclassified": unclassified,
    }


def main():
    OUT.mkdir(parents=True, exist_ok=True)
    results = []
    for program in DECLINED:
        for n in SCALE_POINTS:
            r = decompose(program, n, run="dlc")
            r["T_guarded"] = r.pop("T_guarded")  # (key already named T_guarded for the dlc run)
            assert not r["unclassified"], f"unclassified relations: {r}"
            assert r["declined_portion"] + r["transformed_portion"] == r["T_guarded"], (
                f"decomposition does not sum to T_guarded: {r}")

            rn = decompose(program, n, run="none")
            assert not rn["unclassified"], f"unclassified (none run): {rn}"
            r["T_none"] = rn["T_guarded"]
            r["T_none_declined_predicates"] = rn["declined_portion"]
            r["T_none_sibling_predicates"] = rn["transformed_portion"]
            r["ratio_T_none_over_T_guarded"] = r["T_none"] / r["T_guarded"] if r["T_guarded"] else None

            results.append(r)
            print(f"{program:24s} n={n:4d}  T_none={r['T_none']:6d}  T_guarded={r['T_guarded']:6d}  "
                  f"ratio={r['ratio_T_none_over_T_guarded']:.3f}  |  "
                  f"guarded: declined={r['declined_portion']:6d} transformed={r['transformed_portion']:5d}  |  "
                  f"none: declined-preds={r['T_none_declined_predicates']:6d} sibling-preds={r['T_none_sibling_predicates']:5d}")
    (OUT / "summary.json").write_text(json.dumps(results, indent=2))


if __name__ == "__main__":
    main()
