#!/usr/bin/env python3
"""
NIGHT-BATCH-03 T4 item 1: structural culprit-cycle classifier.

Formal criterion, from docs/dlc-blueprint.md section 6 ("Mechanism to
detect"), NOT a reinvention: magic rules take the body literals preceding
the target atom under SIPS order. If a preceding literal is `!q(...)`, then
`magic_r` (r = the target atom immediately/later in that same body)
acquires a negative edge to `q`. If `q` transitively depends on `r`, and
`r` depends on `magic_r` (always true post-transform), the cycle
`magic_r ->(neg) q -> r -> magic_r` closes -- a *culprit cycle* (Chen 1997).

Cheap necessary precondition, checked first (O(V+E)): r must lie in a
POSITIVE cycle in the SOURCE precedence graph (self-loop or a nontrivial
SCC using only positive edges) -- this is what "the affected predicate must
lie in a positive cycle" (blueprint section 6) actually gates on.

Concretely, per rule body (source token order used as a SIPS proxy -- dlc's
own SIPS is left-to-right in source order plus one required deviation,
M2-M3-BUILD.md section 3, so source order is a reasonable structural proxy
for a precondition scan, not the real adornment): for every negated literal
`!q(...)` followed later in the same body by a positive atom `r(...)`,
flag (q, r) if r has a positive self/cyclic dependency AND q transitively
depends on r (any edge polarity) in the source precedence graph.

Structural scan only -- no dlc parse, no Soufflé execution, works on any
file whether or not it is in blueprint section 4's grammar (functors,
aggregates, etc. included; NIGHT-BATCH-03.md section 4 asks for a scan over
files that are NOT limited to the grammar). Literal extraction is a
regex-based approximation, not a real parser: it can miscount a functor
call nested inside an argument list as a spurious literal. Disclosed, not
silently assumed exact -- see the report's "what a skeptic attacks first."
"""
import json
import re
import sys
from pathlib import Path
from collections import defaultdict, deque

REPO = Path("/root/compiler")
SOUFFLE_TESTS = Path("/root/souffle-src/tests")

COMMENT_LINE_RE = re.compile(r"//.*")
COMMENT_BLOCK_RE = re.compile(r"/\*.*?\*/", re.DOTALL)
HEAD_RE = re.compile(r"^\s*!?\s*([A-Za-z_][A-Za-z0-9_.]*)\s*\(")
LIT_RE = re.compile(r"(!)?\s*([A-Za-z_][A-Za-z0-9_.]*)\s*\(")


def strip_comments(text):
    text = COMMENT_BLOCK_RE.sub(" ", text)
    text = COMMENT_LINE_RE.sub("", text)
    return text


DIRECTIVE_LINE_RE = re.compile(
    r"(?m)^\s*\.(decl|input|output|type|functor|pragma|comp|init|override)\b[^\n]*$"
)
PREPROCESSOR_LINE_RE = re.compile(r"(?m)^\s*#.*$")


def strip_directives(text):
    """Directives (.decl/.input/.output/...) and preprocessor lines do not
    end in a period the way clauses do -- splitting on top-level '.' before
    removing them would otherwise merge a `.decl` with the very next clause
    (the first real period encountered terminates the clause, not the
    decl). Stripped whole-line, single-line only -- a `.decl` whose
    parameter list wraps across lines is a known, disclosed approximation
    gap in this structural scan, not a real parser."""
    text = PREPROCESSOR_LINE_RE.sub("", text)
    text = DIRECTIVE_LINE_RE.sub("", text)
    return text


def split_top_level(text, sep="."):
    """Split text on `sep` only at paren-depth 0 and outside string
    literals -- a period inside `"..."` or inside `(...)` (e.g. a float or
    a nested call) never ends a statement."""
    out, buf, depth, in_str = [], [], 0, False
    i = 0
    while i < len(text):
        c = text[i]
        if in_str:
            buf.append(c)
            if c == "\\" and i + 1 < len(text):
                buf.append(text[i + 1])
                i += 2
                continue
            if c == '"':
                in_str = False
            i += 1
            continue
        if c == '"':
            in_str = True
            buf.append(c)
        elif c in "([":
            depth += 1
            buf.append(c)
        elif c in ")]":
            depth -= 1
            buf.append(c)
        elif c == sep and depth == 0:
            out.append("".join(buf))
            buf = []
        else:
            buf.append(c)
        i += 1
    if buf and "".join(buf).strip():
        out.append("".join(buf))
    return out


def parse_structure(text):
    """Returns (heads, rules) where rules is a list of (head_name, [(neg, name), ...])
    in source body order, restricted to non-directive statements containing ':-'."""
    text = strip_comments(text)
    text = strip_directives(text)
    statements = split_top_level(text, ".")
    heads = set()
    rules = []
    for stmt in statements:
        s = stmt.strip()
        if not s:
            continue
        if ":-" not in s:
            # a fact -- still registers the relation as an IDB head
            m = HEAD_RE.match(s)
            if m:
                heads.add(m.group(1))
            continue
        head_part, body_part = s.split(":-", 1)
        m = HEAD_RE.match(head_part)
        if not m:
            continue
        head = m.group(1)
        heads.add(head)
        lits = [(bool(neg), name) for neg, name in LIT_RE.findall(body_part)]
        rules.append((head, lits))
    return heads, rules


def build_graphs(heads, rules):
    pos_edges = defaultdict(set)
    all_edges = defaultdict(set)
    for head, lits in rules:
        for neg, name in lits:
            if name in heads:
                all_edges[head].add(name)
                if not neg:
                    pos_edges[head].add(name)
    return pos_edges, all_edges


def reachable(edges, src, dst):
    if src == dst:
        return True
    seen = {src}
    q = deque([src])
    while q:
        u = q.popleft()
        for v in edges.get(u, ()):
            if v == dst:
                return True
            if v not in seen:
                seen.add(v)
                q.append(v)
    return False


def has_positive_cycle(pos_edges, r):
    """r lies in a positive cycle iff r is reachable from itself via >=1
    positive edge (covers both a direct self-loop and a longer positive
    mutual-recursion cycle)."""
    for v in pos_edges.get(r, ()):
        if reachable(pos_edges, v, r):
            return True
    return False


def classify(text):
    heads, rules = parse_structure(text)
    pos_edges, all_edges = build_graphs(heads, rules)
    findings = []
    for head, lits in rules:
        for i, (neg_i, name_i) in enumerate(lits):
            if not neg_i or name_i not in heads:
                continue
            for name_j_neg, name_j in lits[i + 1:]:
                if name_j_neg or name_j not in heads:
                    continue
                r = name_j
                q = name_i
                if r == q:
                    continue
                if has_positive_cycle(pos_edges, r) and reachable(all_edges, q, r):
                    findings.append({"rule_head": head, "negated": q, "target_r": r})
    return {"heads": sorted(heads), "n_rules": len(rules), "findings": findings}


def scan_corpus(file_paths, label):
    flagged = []
    errors = 0
    for fp in file_paths:
        try:
            text = fp.read_text(errors="replace")
        except OSError:
            errors += 1
            continue
        result = classify(text)
        if result["findings"]:
            flagged.append({"file": str(fp), "findings": result["findings"]})
    return {"label": label, "attempted": len(file_paths), "flagged": len(flagged),
            "read_errors": errors, "details": flagged}


def main():
    gate1 = json.loads((REPO / "measurements" / "m1-3.3-gate1-parse-coverage-summary.json").read_text())
    in_grammar_files = [SOUFFLE_TESTS / line.strip() for line in
                        (REPO / "tests" / "corpus" / "IN_GRAMMAR.txt").read_text().splitlines()
                        if line.strip() and not line.strip().startswith("#")]

    all_dl_files = sorted(SOUFFLE_TESTS.rglob("*.dl"))

    family_dir = REPO / "tests" / "corpus" / "BENCHMARK_FAMILY"
    family_files = sorted(family_dir.glob("*.dl"))

    results = {
        "IN_GRAMMAR_195": scan_corpus(in_grammar_files, "IN_GRAMMAR.txt (195)"),
        "SOUFFLE_TREE_622": scan_corpus(all_dl_files, "full Souffle tests/ tree"),
        "BENCHMARK_FAMILY_5": scan_corpus(family_files, "BENCHMARK_FAMILY shapes"),
    }
    for key, r in results.items():
        print(f"{r['label']}: attempted={r['attempted']} flagged={r['flagged']} read_errors={r['read_errors']}", file=sys.stderr)

    out_dir = REPO / "measurements" / "night03-t4"
    out_dir.mkdir(parents=True, exist_ok=True)
    (out_dir / "classifier-census.json").write_text(json.dumps(results, indent=2))
    print(json.dumps({k: {kk: vv for kk, vv in v.items() if kk != "details"} for k, v in results.items()}, indent=2))


if __name__ == "__main__":
    main()
