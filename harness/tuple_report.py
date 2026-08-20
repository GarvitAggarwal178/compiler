#!/usr/bin/env python3
"""
Generalized tuple-count extraction from a Soufflé JSON profile log (`-p`
flag). Supersedes the single-purpose logic duplicated across probe0.py-
era scripts. Emits, per relation: its total tuple count, whether it is a
`COPY_T`-shaped relation (a synthetic single-literal pass-through rule
Soufflé emits when a magic-restricted relation also has an unconditional
`.output`, first identified in Phase 0 -- see docs/reports/probe0.md),
and whether its name carries the `@neglabel.` prefix (Soufflé's
negation-isolation marker, first identified in Phase 0/0.5).

Program-level summary: T (excl-copy) = sum over all non-input relations
excluding COPY_T ones; T (incl-copy) = sum including them;
E_recoverable = sum over `@neglabel.`-prefixed relations only.

This is Lane B instrumentation reading Soufflé's own output format. It
implements no Datalog semantics itself.
"""
import json
import re
import sys
from pathlib import Path

COPY_RULE_RE = re.compile(
    r"^\s*[\w.@{}]+\([^)]*\)\s*:-\s*(@[\w.{}]+)\([^)]*\)\s*\.\s*$", re.DOTALL
)


def relation_total(rel: dict):
    base = rel.get("num-tuples", 0) or 0
    iterations = rel.get("iteration")
    if iterations:
        delta_sum = sum(v.get("num-tuples", 0) or 0 for v in iterations.values())
        return base + delta_sum
    return base


def is_copy_relation(rel: dict) -> bool:
    """A COPY_T-shaped relation: exactly one non-recursive rule, whose
    text is a single-literal pass-through *from an `@`-prefixed internal
    relation* (Soufflé's synthetic adorned/interm_out namespace) into a
    plain output name. The `@`-prefix requirement on the body is load-
    bearing: a plain single-literal rule like `@neglabel.s :- base(...).`
    has the same *shape* but is a genuine independent re-derivation
    under a new synthetic name, not a redundant copy-out of already-
    computed data -- conflating the two was a real bug caught while
    validating this script against P3's known numbers (see git history
    of this file / docs/reports/night01-T2-corpus.md)."""
    if rel.get("iteration"):
        return False
    nr = rel.get("non-recursive-rule")
    if not nr or len(nr) != 1:
        return False
    (rule_text,) = nr.keys()
    return bool(COPY_RULE_RE.match(rule_text.strip()))


def is_input_relation(rel: dict) -> bool:
    return "loadtime" in rel and "non-recursive-rule" not in rel and "iteration" not in rel


def analyze(log_path: Path):
    with open(log_path) as f:
        doc = json.load(f)
    input_names = set()
    cfg = doc["root"]["program"].get("configuration", {})
    relations = doc["root"]["program"]["relation"]

    per_relation = {}
    for name, rel in relations.items():
        total = relation_total(rel)
        per_relation[name] = {
            "total": total,
            "is_input": is_input_relation(rel),
            "is_copy": is_copy_relation(rel),
            "is_neglabel": name.startswith("@neglabel."),
        }

    t_excl_copy = sum(
        v["total"] for v in per_relation.values() if not v["is_input"] and not v["is_copy"]
    )
    t_incl_copy = sum(
        v["total"] for v in per_relation.values() if not v["is_input"]
    )
    e_recoverable = sum(
        v["total"] for v in per_relation.values() if v["is_neglabel"]
    )
    neglabel_relations = {
        k: v["total"] for k, v in per_relation.items() if v["is_neglabel"]
    }

    return {
        "log_path": str(log_path),
        "relation_count": len(cfg.get("relationCount", "")) and cfg.get("relationCount"),
        "T_excl_copy": t_excl_copy,
        "T_incl_copy": t_incl_copy,
        "E_recoverable": e_recoverable,
        "neglabel_relations": neglabel_relations,
        "per_relation": per_relation,
    }


def main():
    if len(sys.argv) != 2:
        print(f"usage: {sys.argv[0]} <profile.log>", file=sys.stderr)
        raise SystemExit(2)
    print(json.dumps(analyze(Path(sys.argv[1])), indent=2, sort_keys=True))


if __name__ == "__main__":
    main()
