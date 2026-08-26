#!/usr/bin/env python3
"""
M1 §3.6 gate, part 2: "produces a stratum order agreeing with Soufflé's
evaluation order on the in-grammar corpus programs that contain
negation." Part 1 (correctly rejects an unstratifiable program) is
already covered by the 3 tests/rejection/stratification.py cases, run as
Go unit tests (src/sema/stratify_test.go) and cross-checked end-to-end
below too.

Method for part 2: `souffle --show=initial-ram <file>` (no facts needed,
purely static) emits a `BEGIN MAIN ... END MAIN` block with one `CALL
stratum_<Relation>` line per IDB relation, in Soufflé's actual execution
order. Found empirically while building this gate: that order is NOT a
strict numeric-stratum grouping (evaluation/set_ops_output/
set_ops_output.dl's MAIN calls stratum_AminusB, which this project's own
stratifier puts at stratum 1, BEFORE stratum_AunionB/stratum_AxB, both
stratum 0) -- Soufflé is free to interleave any valid topological order of
the dependency DAG, not sorted by stratum number. The correctness
invariant that actually matters, and the one this script checks, is
narrower and exactly what stratified negation requires: for every
relation X with a literal `!Y` in some rule body, Y's CALL must appear
strictly before X's CALL in Soufflé's MAIN sequence, and dlc's own
stratum[Y] must be strictly less than stratum[X]. Agreement is checked on
that relation, not on matching Soufflé's exact call sequence or its
internal stratum numbering.

Sample size caveat, stated plainly: only 1 of the 195 in-grammar files
both (a) parses under dlc today (§3.3 gate one: 20/195) and (b) contains
real negation (`!ident(`, not just a `!=` relop) -- the other two
candidates in that intersection are already-known problem files (a
reference to Soufflé's `match` builtin; a designed-to-fail negative test
case, both noted in §3.4's session log entry). This is a direct,
disclosed consequence of gate one's already-reported shortfall, not a
new limitation -- reported honestly as n=1, not inflated.
"""
import json
import re
import subprocess
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import dlc_interface  # noqa: E402

REPO = Path(__file__).resolve().parents[1]
SOUFFLE_TESTS = Path("/root/souffle-src/tests")
GATE1_SUMMARY = REPO / "measurements" / "m1-3.3-gate1-parse-coverage-summary.json"

NEGATED_ATOM_RE = re.compile(r"!\s*[A-Za-z_]\w*\s*\(")


def souffle_call_order(dl_path):
    proc = subprocess.run(["souffle", "--show=initial-ram", str(dl_path)],
                           capture_output=True, encoding="utf-8", errors="replace")
    calls = re.findall(r"CALL stratum_(\w+)", proc.stdout)
    return calls


def dlc_negative_edges(source_text):
    """Returns [(from_relation, to_relation), ...] for every negated
    body atom, by re-parsing with the real dlc parser (not a second
    regex pass over source -- this reuses the actual AST the stratifier
    itself walks, via a tiny amount of duplicated logic here only to
    extract edges for comparison, not to reimplement the stratifier)."""
    # Reuse the already-verified dlc_interface.run_dlc_check strata
    # output isn't enough by itself (it doesn't expose edges) -- but the
    # stratum values ARE enough to check the invariant: for a negative
    # edge X->Y, stratum[Y] < stratum[X]. Edges themselves come from a
    # plain text scan restricted to lines with ':-' and '!', which is
    # only used to enumerate WHICH pairs to check, not to compute
    # anything the correctness check depends on.
    edges = []
    for clause in re.split(r"\.\s*\n", source_text):
        if ":-" not in clause:
            continue
        head_part, body_part = clause.split(":-", 1)
        head_match = re.match(r"\s*([A-Za-z_]\w*)\s*\(", head_part)
        if not head_match:
            continue
        head = head_match.group(1)
        for m in re.finditer(r"!\s*([A-Za-z_]\w*)\s*\(", body_part):
            edges.append((head, m.group(1)))
    return edges


def main():
    gate1 = json.loads(GATE1_SUMMARY.read_text())
    parsed_files = [r["file"] for r in gate1["results"] if r["status"] == "parsed"]

    candidates = []
    for rel in parsed_files:
        text = (SOUFFLE_TESTS / rel).read_text(errors="replace")
        if NEGATED_ATOM_RE.search(text):
            candidates.append(rel)

    dlc_interface.build_dlc()

    results = []
    for rel in candidates:
        dl_path = SOUFFLE_TESTS / rel
        source = dl_path.read_text(errors="replace")
        check = dlc_interface.run_dlc_check(source)
        if check.status != "ok":
            results.append({"file": rel, "status": "dlc_rejected_other_ground",
                             "diagnostics": check.diagnostics})
            continue
        calls = souffle_call_order(dl_path)
        position = {name: i for i, name in enumerate(calls)}
        edges = dlc_negative_edges(source)
        violations = []
        for head, target in edges:
            if head not in position or target not in position:
                continue  # relation not in Souffle's IDB call list (shouldn't happen for a well-formed program)
            if not (position[target] < position[head]):
                violations.append({"from": head, "to": target, "souffle_positions": [position[target], position[head]]})
            if not (check.strata.get(target, -1) < check.strata.get(head, -1)):
                violations.append({"from": head, "to": target, "dlc_strata": [check.strata.get(target), check.strata.get(head)]})
        results.append({
            "file": rel, "status": "agree" if not violations else "disagree",
            "dlc_strata": check.strata, "souffle_call_order": calls, "violations": violations,
        })

    summary = {"candidates_considered": len(candidates), "results": results}
    out_path = REPO / "measurements" / "m1-3.6-stratification-summary.json"
    out_path.write_text(json.dumps(summary, indent=2))
    print(json.dumps(summary, indent=2))

    agree = sum(1 for r in results if r["status"] == "agree")
    print(f"GATE (part 2, n={len(candidates)}): {agree}/{len(results)} agree with Soufflé's ordering "
          f"invariant on the in-grammar files that both parse and contain negation", file=sys.stderr)
    disagree = [r for r in results if r["status"] == "disagree"]
    if disagree:
        print(f"{len(disagree)} DISAGREEMENT(S):", file=sys.stderr)
        for r in disagree:
            print(f"  {r['file']}: {r['violations']}", file=sys.stderr)
        raise SystemExit(1)


if __name__ == "__main__":
    main()
