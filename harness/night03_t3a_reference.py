#!/usr/bin/env python3
"""
NIGHT-BATCH-03 T3a: Souffle's own transformed programs, as reference
material (not golden targets -- Souffle applies inlining and its own
@-prefixed naming that dlc will not replicate; no syntactic diff is built
against these).

Confirmed option name (souffle --help, this session): the transformed-AST
dump mode is `--show=transformed-ast`, NOT `--show=transformed-datalog` as
speculatively named in NIGHT-BATCH-03.md -- verified before use, not assumed.
"""
import subprocess
import sys
from pathlib import Path

REPO = Path("/root/compiler")
FAMILY_DIR = REPO / "tests" / "corpus" / "BENCHMARK_FAMILY"
OUT_DIR = REPO / "tests" / "reference" / "souffle-transformed"

FILES = [
    ("same_generation_negation", FAMILY_DIR / "same_generation_negation.dl"),
    ("transitive_closure_bound", FAMILY_DIR / "transitive_closure_bound.dl"),
    ("ancestor_nonancestor", FAMILY_DIR / "ancestor_nonancestor.dl"),
    ("culprit_cycle", FAMILY_DIR / "culprit_cycle.dl"),
    ("reachability_complement", FAMILY_DIR / "reachability_complement.dl"),
    ("p1prime", REPO / "tests" / "programs" / "p1prime.dl"),
    ("p2", REPO / "tests" / "programs" / "p2.dl"),
    ("p4prime", REPO / "tests" / "programs" / "p4prime.dl"),
]


def dump(dl_path: Path, magic: bool):
    cmd = ["souffle", "--show=transformed-ast"]
    if magic:
        cmd.append("--magic-transform=*")
    cmd.append(str(dl_path))
    proc = subprocess.run(cmd, capture_output=True, text=True, timeout=60)
    return proc.returncode, proc.stdout, proc.stderr


def main():
    OUT_DIR.mkdir(parents=True, exist_ok=True)
    report = []
    for name, dl_path in FILES:
        if not dl_path.is_file():
            report.append({"name": name, "status": "missing_source", "path": str(dl_path)})
            continue
        rc_plain, out_plain, err_plain = dump(dl_path, magic=False)
        rc_magic, out_magic, err_magic = dump(dl_path, magic=True)

        (OUT_DIR / f"{name}.plain.dl").write_text(out_plain)
        (OUT_DIR / f"{name}.magic.dl").write_text(out_magic)
        row = {
            "name": name, "source": str(dl_path.relative_to(REPO)),
            "plain_rc": rc_plain, "magic_rc": rc_magic,
            "magic_stderr_has_error": "Error:" in err_magic,
        }
        if rc_magic != 0 or "Error:" in err_magic:
            row["magic_stderr"] = err_magic[:1000]
        report.append(row)
        print(f"{name}: plain_rc={rc_plain} magic_rc={rc_magic}", file=sys.stderr)

    (OUT_DIR / "MANIFEST.md").write_text(
        "# Souffle reference transforms\n\n"
        "Reference material, not golden targets -- Souffle applies inlining and its\n"
        "own `@`-prefixed naming that `dlc` will not replicate. A syntactic diff\n"
        "against these is worthless and none is built. Value: shows what a correct\n"
        "transform of a known program looks like, for human reading.\n\n"
        "Generated via `souffle --show=transformed-ast [--magic-transform=*] <file>`,\n"
        "confirmed option name (not `--show=transformed-datalog`, which does not\n"
        "exist in Souffle 2.5 -- checked via `souffle --help` before use).\n\n"
        "| file | source | plain rc | magic rc |\n|---|---|---|---|\n" +
        "\n".join(f"| {r['name']} | {r.get('source','-')} | {r.get('plain_rc','-')} | {r.get('magic_rc','-')} |"
                  for r in report)
    )
    import json
    print(json.dumps(report, indent=2))


if __name__ == "__main__":
    main()
