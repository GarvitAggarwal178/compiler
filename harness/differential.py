#!/usr/bin/env python3
"""
NIGHT-BATCH-01 T7 item 1: differential runner. Takes a .dl program and a
fact directory, runs both dlc (via a stub -- Lane A doesn't exist yet)
and Soufflé, compares output relations by set equality on sorted output
(CLAUDE.md section 6), reports the symmetric difference per relation,
not a boolean.

The dlc side is a deliberate stub, not an accident of a missing binary:
run_dlc() always returns status="not_implemented" right now. M1 (Lane A)
replaces its body with a real subprocess call to the built dlc binary;
everything downstream (comparison, reporting) is already wired and does
not change when that happens.
"""
import json
import subprocess
import sys
from dataclasses import dataclass, field
from pathlib import Path
from typing import Optional


@dataclass
class EngineResult:
    engine: str
    status: str  # "ok" | "error:<msg>" | "not_implemented"
    output_relations: dict = field(default_factory=dict)  # name -> sorted list of lines
    stderr: str = ""


@dataclass
class Comparison:
    relation: str
    dlc_only: list
    souffle_only: list
    match: bool


def run_dlc(dl_path: Path, facts_dir: Path) -> EngineResult:
    """STUB. M1 (Lane A) replaces this body with a real invocation of the
    built dlc binary, once it exists. Everything else in this module is
    already correct for that day and does not need to change."""
    return EngineResult(engine="dlc", status="not_implemented",
                         stderr="dlc is not implemented yet (M1, Lane A)")


def run_souffle(dl_path: Path, facts_dir: Path, workdir: Path) -> EngineResult:
    # Resolve to absolute paths before changing cwd for the subprocess --
    # relative paths here would resolve against `workdir`, not the
    # caller's cwd (the exact bug found and fixed in T2 tonight).
    dl_path = dl_path.resolve()
    facts_dir = facts_dir.resolve()
    workdir.mkdir(parents=True, exist_ok=True)
    output_names = _extract_output_names(dl_path)
    cmd = ["souffle", "-F", str(facts_dir), "-D", str(workdir), str(dl_path)]
    try:
        proc = subprocess.run(cmd, cwd=str(workdir), capture_output=True, text=True, timeout=300)
    except subprocess.TimeoutExpired:
        return EngineResult(engine="souffle", status="error:timeout")
    if proc.returncode != 0:
        return EngineResult(engine="souffle", status=f"error:returncode-{proc.returncode}", stderr=proc.stderr)

    relations = {}
    for name in output_names:
        fp = workdir / f"{name}.csv"
        if fp.is_file():
            relations[name] = sorted(fp.read_text().splitlines())
        else:
            relations[name] = []
    return EngineResult(engine="souffle", status="ok", output_relations=relations, stderr=proc.stderr)


def _extract_output_names(dl_path: Path):
    import re
    text = dl_path.read_text(errors="replace")
    return sorted(set(re.findall(r"^\s*\.output\s+(\w+)", text, re.MULTILINE)))


def compare(dlc: EngineResult, souffle: EngineResult):
    """Set equality per relation, symmetric difference reported. If dlc
    is not_implemented, every relation is reported as fully souffle_only
    (nothing to compare against) rather than silently skipped -- the
    caller must be able to tell "not comparable yet" from "compared and
    matched"."""
    if dlc.status != "ok":
        return {
            "comparable": False,
            "reason": dlc.status,
            "relations": list(souffle.output_relations.keys()),
        }
    all_relations = set(dlc.output_relations) | set(souffle.output_relations)
    comparisons = []
    for rel in sorted(all_relations):
        d = set(dlc.output_relations.get(rel, []))
        s = set(souffle.output_relations.get(rel, []))
        comparisons.append(Comparison(
            relation=rel,
            dlc_only=sorted(d - s),
            souffle_only=sorted(s - d),
            match=(d == s),
        ))
    return {
        "comparable": True,
        "all_match": all(c.match for c in comparisons),
        "relations": [c.__dict__ for c in comparisons],
    }


def main():
    if len(sys.argv) != 4:
        print(f"usage: {sys.argv[0]} <dl_path> <facts_dir> <workdir>", file=sys.stderr)
        raise SystemExit(2)
    dl_path, facts_dir, workdir = Path(sys.argv[1]), Path(sys.argv[2]), Path(sys.argv[3])

    dlc_result = run_dlc(dl_path, facts_dir)
    souffle_result = run_souffle(dl_path, facts_dir, workdir)
    result = compare(dlc_result, souffle_result)
    result["dlc_status"] = dlc_result.status
    result["souffle_status"] = souffle_result.status
    print(json.dumps(result, indent=2))


if __name__ == "__main__":
    main()
