#!/usr/bin/env python3
"""
NIGHT-BATCH-02 T2: hostile source corpus. Text files only
(tests/hostile/*.dl), no dlc dependency -- runs each through the installed
Soufflé 2.5 with an empty facts directory (any relation any file happens
to declare .input for) and records accept/reject plus the exact
diagnostic. Gives an oracle-backed expectation for every case before the
lexer exists.

Note found while writing T5: Soufflé can exit 0 on some rejections while
still printing "Error:" to stderr and producing no output -- so this
script treats a run as "reject" if either the return code is nonzero OR
stderr contains "Error:", same fix already applied in
harness/night02_t5_guarded.py.
"""
import json
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
import probe0  # noqa: E402

REPO = probe0.REPO
MEASUREMENTS = probe0.MEASUREMENTS
HOSTILE_DIR = REPO / "tests" / "hostile"


def main():
    facts_dir = MEASUREMENTS / "_scratch_night02_t2" / "facts"
    facts_dir.mkdir(parents=True, exist_ok=True)
    for name in ("q", "r", "p"):
        (facts_dir / f"{name}.facts").write_text("")
    # Deliberately empty, not a sample row: different hostile files reuse
    # "q" at different arities (q(a:number) vs q(a:number,b:number)), and
    # this single shared facts directory can't satisfy both with real data
    # without an arity mismatch on whichever file expects the other shape.
    # Empty is arity-agnostic and safe everywhere; it costs the ability to
    # numerically verify the unary_double_minus.dl /
    # unary_paren_double_neg.dl output values (checked separately, see
    # measurements/_scratch_night02_t2/ for the one-off spot check with
    # real data used for the other files that don't share the conflict).
    # the 4KB-identifier file declares its own oddly-named .input relation;
    # its facts filename would itself exceed the filesystem's own filename
    # length limit (confirmed: OSError, name too long, ext4 255-byte cap),
    # so it deliberately gets no matching .facts file -- Soufflé's handling
    # of a missing input file is itself part of what this case exercises.
    for name in ("données",):
        (facts_dir / f"{name}.facts").write_text("")

    files = sorted(HOSTILE_DIR.glob("*.dl"))
    results = []
    for dl_path in files:
        mid = f"night02-t2-hostile-{dl_path.stem}"
        workdir = MEASUREMENTS / f"{mid}-run"
        workdir.mkdir(parents=True, exist_ok=True)
        cmd = ["souffle", "-F", str(facts_dir), "-D", str(workdir), str(dl_path)]
        proc = probe0.run_cmd(mid, cmd, workdir)
        rejected = (proc.returncode != 0) or ("Error:" in proc.stderr)
        outcome = "reject" if rejected else "accept"
        results.append({
            "file": dl_path.name,
            "outcome": outcome,
            "returncode": proc.returncode,
            "stderr": proc.stderr.strip(),
            "stdout": proc.stdout.strip(),
        })
        print(f"{dl_path.name}: {outcome} (rc={proc.returncode})", file=sys.stderr)

    out_path = MEASUREMENTS / "night02-t2-hostile-summary.json"
    out_path.write_text(json.dumps(results, indent=2, ensure_ascii=False))
    accepted = sum(1 for r in results if r["outcome"] == "accept")
    rejected = sum(1 for r in results if r["outcome"] == "reject")
    print(json.dumps({"total": len(results), "accept": accepted, "reject": rejected}, indent=2))


if __name__ == "__main__":
    main()
