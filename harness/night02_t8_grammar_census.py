#!/usr/bin/env python3
"""
NIGHT-BATCH-02 T8: grammar usage census over the 195 files in
tests/corpus/IN_GRAMMAR.txt (NIGHT-BATCH-01 T5's mechanically-verified
in-grammar pool -- NOT the pre-registered corpus, see that file's own
header). Static analysis only, no Soufflé invocation, no dlc.

Method: a small hand-written tokenizer for exactly blueprint section 4's
grammar (comments stripped first), then simple structural counting over
the token stream -- not a full parser, but token-aware rather than pure
regex-on-raw-text, so operator counting doesn't confuse e.g. '<=' with a
lone '<', and body-literal splitting respects paren nesting. Known
limitation, disclosed rather than hidden: recursion detection is
DIRECT-recursion only (a rule's body mentions its own head relation by
name) -- mutual recursion through an intermediate relation is not
detected by this mechanical scan and would need real dependency-graph
construction (that is Lane A's job, not this census's).
"""
import json
import re
import statistics
import sys
from pathlib import Path

REPO = Path(__file__).resolve().parents[1]
SOUFFLE_TESTS = Path("/root/souffle-src/tests")
IN_GRAMMAR = REPO / "tests" / "corpus" / "IN_GRAMMAR.txt"
MEASUREMENTS = REPO / "measurements"

COMMENT_BLOCK_RE = re.compile(r"/\*.*?\*/", re.DOTALL)
COMMENT_LINE_RE = re.compile(r"//[^\n]*")

TOKEN_RE = re.compile(r"""
    (?P<string>"[^"]*")
  | (?P<number>\d+)
  | (?P<ident>[A-Za-z_]\w*)
  | (?P<op><=|>=|!=|[=<>+\-*/%!,.()])
  | (?P<ws>\s+)
""", re.VERBOSE)

ARITH_OPS = {"+", "-", "*", "/", "%"}
RELOPS = {"=", "!=", "<", "<=", ">", ">="}


def strip_comments(text):
    text = COMMENT_BLOCK_RE.sub(" ", text)
    text = COMMENT_LINE_RE.sub(" ", text)
    return text


def tokenize(text):
    toks = []
    for m in TOKEN_RE.finditer(text):
        kind = m.lastgroup
        if kind == "ws":
            continue
        toks.append((kind, m.group()))
    return toks


def load_file_list():
    lines = IN_GRAMMAR.read_text().splitlines()
    return [ln.strip() for ln in lines if ln.strip() and not ln.startswith("#")]


def split_top_level(tokens, sep_tok):
    """Split tokens at top-level (paren-depth 0) occurrences of a single
    literal token (e.g. ',')."""
    parts, cur, depth = [], [], 0
    for kind, tok in tokens:
        if tok == "(":
            depth += 1
        elif tok == ")":
            depth -= 1
        if tok == sep_tok and depth == 0:
            parts.append(cur)
            cur = []
            continue
        cur.append((kind, tok))
    parts.append(cur)
    return parts


def max_paren_depth(tokens):
    depth = maxd = 0
    for kind, tok in tokens:
        if tok == "(":
            depth += 1
            maxd = max(maxd, depth)
        elif tok == ")":
            depth -= 1
    return maxd


def analyze_file(rel_path):
    dl_path = SOUFFLE_TESTS / rel_path
    raw = dl_path.read_text(errors="replace")
    text = strip_comments(raw)

    stats = {
        "arith_ops": {op: 0 for op in ARITH_OPS},
        "relops": {op: 0 for op in RELOPS},
        "max_expr_depth": 0,
        "expr_depths": [],
        "negation_count": 0,
        "negation_positions": {"first": 0, "middle": 0, "last": 0, "only": 0},
        "wildcard_count": 0,
        "wildcard_positions": {"head": 0, "body": 0},
        "arities": [],
        "body_lengths": [],
        "head_literal_clauses": 0,
        "body_literal_clauses": 0,
        "string_literals": 0,
        "number_literals": 0,
        "rules_per_relation": {},
        "recursive_relations": set(),
        "all_relations": set(),
    }

    # .decl arities: line-based regex, directives aren't part of the
    # clause grammar and are simplest handled on raw (comment-stripped)
    # text directly.
    for m in re.finditer(r"\.decl\s+(\w+)\s*\(([^)]*)\)", text):
        params = [p for p in m.group(2).split(",") if p.strip()]
        stats["arities"].append(len(params))
        stats["all_relations"].add(m.group(1))

    # Split into clauses on raw text (regex, since ':-' isn't a single
    # token in our tokenizer and clause boundaries are easiest found on
    # text): a clause is up to the next '.' that is not inside parens and
    # not part of a directive line.
    body_text = re.sub(r"^\s*\.(decl|input|output)\b[^\n]*$", "", text, flags=re.MULTILINE)
    for clause_match in re.finditer(r"([^.]*(?:\([^()]*\)[^.]*)*)\.", body_text, re.DOTALL):
        clause_raw = clause_match.group(1).strip()
        if not clause_raw:
            continue
        if ":-" in clause_raw:
            head_raw, sep, body_raw = clause_raw.partition(":-")
        else:
            head_raw, body_raw = clause_raw, ""

        head_toks = tokenize(head_raw)
        head_ident = next((t for k, t in head_toks if k == "ident"), None)
        if head_ident:
            stats["all_relations"].add(head_ident)
            stats["rules_per_relation"].setdefault(head_ident, 0)
            stats["rules_per_relation"][head_ident] += 1

        # wildcard: '_' tokenizes as an "ident" (matches [A-Za-z_]\w*) equal to "_"
        for k, t in head_toks:
            if t == "_":
                stats["wildcard_count"] += 1
                stats["wildcard_positions"]["head"] += 1

        if not body_raw:
            continue  # fact, no body to analyze further

        body_toks = tokenize(body_raw)
        for k, t in body_toks:
            if t == "_":
                stats["wildcard_count"] += 1
                stats["wildcard_positions"]["body"] += 1
            if t in ARITH_OPS:
                stats["arith_ops"][t] += 1
            if t in RELOPS:
                stats["relops"][t] += 1
            if k == "string":
                stats["string_literals"] += 1
            if k == "number":
                stats["number_literals"] += 1

        depth = max_paren_depth(body_toks)
        stats["expr_depths"].append(depth)
        stats["max_expr_depth"] = max(stats["max_expr_depth"], depth)

        # body literals: split on top-level commas
        literals = split_top_level(body_toks, ",")
        literals = [lit for lit in literals if lit]
        stats["body_lengths"].append(len(literals))
        n = len(literals)
        for i, lit in enumerate(literals):
            is_neg = lit[0][1] == "!"
            if is_neg:
                stats["negation_count"] += 1
                if n == 1:
                    stats["negation_positions"]["only"] += 1
                elif i == 0:
                    stats["negation_positions"]["first"] += 1
                elif i == n - 1:
                    stats["negation_positions"]["last"] += 1
                else:
                    stats["negation_positions"]["middle"] += 1
            lit_ident = next((t for k, t in lit if k == "ident" and t != head_ident), None)
            # direct recursion: body references the head's own relation name
            body_idents = {t for k, t in lit if k == "ident"}
            if head_ident and head_ident in body_idents:
                stats["recursive_relations"].add(head_ident)
            stats["all_relations"] |= {t for k, t in lit if k == "ident"}

        if any(t == "_" for k, t in head_toks) or head_raw:
            stats["head_literal_clauses"] += 1
        if body_raw:
            stats["body_literal_clauses"] += 1

    stats["recursive_relations"] = sorted(stats["recursive_relations"])
    return stats


def merge(agg, stats, fname):
    for op in ARITH_OPS:
        agg["arith_ops"][op] += stats["arith_ops"][op]
    for op in RELOPS:
        agg["relops"][op] += stats["relops"][op]
    agg["expr_depths"].extend(stats["expr_depths"])
    agg["negation_count"] += stats["negation_count"]
    for k in agg["negation_positions"]:
        agg["negation_positions"][k] += stats["negation_positions"][k]
    agg["wildcard_count"] += stats["wildcard_count"]
    for k in agg["wildcard_positions"]:
        agg["wildcard_positions"][k] += stats["wildcard_positions"][k]
    agg["arities"].extend(stats["arities"])
    agg["body_lengths"].extend(stats["body_lengths"])
    agg["string_literals"] += stats["string_literals"]
    agg["number_literals"] += stats["number_literals"]
    for rel, cnt in stats["rules_per_relation"].items():
        agg["rules_per_relation"][rel] = agg["rules_per_relation"].get(rel, 0) + cnt
    agg["recursive_relations"].update(stats["recursive_relations"])
    agg["all_relations"].update(stats["all_relations"])
    agg["files_with_negation"] += 1 if stats["negation_count"] else 0
    agg["files_with_wildcard"] += 1 if stats["wildcard_count"] else 0


def main():
    files = load_file_list()
    agg = {
        "arith_ops": {op: 0 for op in ARITH_OPS},
        "relops": {op: 0 for op in RELOPS},
        "expr_depths": [],
        "negation_count": 0,
        "negation_positions": {"first": 0, "middle": 0, "last": 0, "only": 0},
        "wildcard_count": 0,
        "wildcard_positions": {"head": 0, "body": 0},
        "arities": [],
        "body_lengths": [],
        "string_literals": 0,
        "number_literals": 0,
        "rules_per_relation": {},
        "recursive_relations": set(),
        "all_relations": set(),
        "files_with_negation": 0,
        "files_with_wildcard": 0,
        "files_analyzed": 0,
        "files_missing": 0,
    }
    per_file = {}
    for rel_path in files:
        dl_path = SOUFFLE_TESTS / rel_path
        if not dl_path.is_file():
            agg["files_missing"] += 1
            continue
        stats = analyze_file(rel_path)
        per_file[rel_path] = {
            "arith_ops": stats["arith_ops"], "relops": stats["relops"],
            "max_expr_depth": stats["max_expr_depth"],
            "negation_count": stats["negation_count"],
            "wildcard_count": stats["wildcard_count"],
            "recursive_relations": stats["recursive_relations"],
        }
        merge(agg, stats, rel_path)
        agg["files_analyzed"] += 1

    agg["recursive_relations"] = sorted(agg["recursive_relations"])
    agg["all_relations"] = sorted(agg["all_relations"])
    agg["recursive_relation_count"] = len(agg["recursive_relations"])
    agg["nonrecursive_relation_count"] = len(agg["all_relations"]) - len(agg["recursive_relations"])
    agg["expr_depth_median"] = statistics.median(agg["expr_depths"]) if agg["expr_depths"] else None
    agg["expr_depth_max"] = max(agg["expr_depths"]) if agg["expr_depths"] else None
    agg["body_length_median"] = statistics.median(agg["body_lengths"]) if agg["body_lengths"] else None
    agg["body_length_max"] = max(agg["body_lengths"]) if agg["body_lengths"] else None
    agg["arity_median"] = statistics.median(agg["arities"]) if agg["arities"] else None
    agg["arity_max"] = max(agg["arities"]) if agg["arities"] else None
    agg["arity_distribution"] = {str(a): agg["arities"].count(a) for a in sorted(set(agg["arities"]))}
    agg["body_length_distribution"] = {str(b): agg["body_lengths"].count(b) for b in sorted(set(agg["body_lengths"]))}
    agg["rules_per_relation_median"] = statistics.median(agg["rules_per_relation"].values()) if agg["rules_per_relation"] else None
    agg["rules_per_relation_max"] = max(agg["rules_per_relation"].values()) if agg["rules_per_relation"] else None

    out = {"aggregate": agg, "per_file": per_file}
    out_path = MEASUREMENTS / "night02-t8-grammar-census.json"
    out_path.write_text(json.dumps(out, indent=2, default=list))
    summary = {k: v for k, v in agg.items() if k not in ("expr_depths", "arities", "body_lengths", "rules_per_relation", "all_relations")}
    print(json.dumps(summary, indent=2, default=list))


if __name__ == "__main__":
    main()
