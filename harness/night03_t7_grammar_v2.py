#!/usr/bin/env python3
"""
NIGHT-BATCH-03 T7: corrected corpus predicate. VERIFY-01 established the
true blueprint section 4-compliant count is far below night01_t5_grammar.py's
195 -- that predicate checks 12 construct categories, section 4 implies at
least nine more. This is a NEW file (per instruction) implementing every
exclusion factor named directly against the section 4 production it
violates -- never against "which files failed."

Blueprint section 4 grammar (quoted in full, docs/dlc-blueprint.md):

    program    ::= decl* clause*
    decl       ::= '.decl' ident '(' param (',' param)* ')'
                 | '.input' ident
                 | '.output' ident
    param      ::= ident ':' type
    type       ::= 'number' | 'symbol'
    clause     ::= atom '.' | atom ':-' body '.'
    body       ::= literal (',' literal)*
    literal    ::= atom | '!' atom | constraint
    atom       ::= ident '(' term (',' term)* ')'
    term       ::= arith | '_'
    constraint ::= arith relop arith
    relop      ::= '=' | '!=' | '<' | '<=' | '>' | '>='
    arith      ::= arith ('+'|'-') mul | mul
    mul        ::= mul ('*'|'/'|'%') unary | unary
    unary      ::= '-' unary | primary
    primary    ::= var | number | string | '(' arith ')'

"No functors, no aggregates, no components, no records."

Structural text scan, not a real parser -- same caveats as
night01_t5_grammar.py and harness/night03_t4_culprit_classifier.py:
approximate, disclosed, validated by cross-checking every survivor against
`dlc`'s real parser (a file this predicate admits must actually parse, or
the predicate itself is wrong).
"""
import json
import os
import re
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from corpus_predicate import strip_comments  # noqa: E402
import dlc_interface  # noqa: E402

REPO = Path(__file__).resolve().parents[1]
SOUFFLE_TESTS = Path("/root/souffle-src/tests")
OUT_DIR = REPO / "measurements" / "night03-t7"
IN_GRAMMAR_V2_PATH = REPO / "tests" / "corpus" / "IN_GRAMMAR_V2.txt"

# --- Categories inherited unchanged from night01_t5_grammar.py's 12 checks ---
# (kept exactly as-is so V2's histogram is comparable to V1's on the same axes)
BASELINE_PATTERNS = {
    # violates: decl ::= '.decl' ident '(' ... -- '.type' is not one of the
    # three admitted decl forms.
    "type_decl": re.compile(r"^\s*\.type\b", re.MULTILINE),
    # violates: no '.plan' production anywhere in the grammar.
    "plan_directive": re.compile(r"\.plan\b"),
    # violates: decl ::= '.decl' | '.input' | '.output' only -- '.pragma' is
    # not admitted.
    "pragma_directive": re.compile(r"^\s*\.pragma\b", re.MULTILINE),
    # violates: no component/instantiation production ('.comp'/'.init') exists.
    "component": re.compile(r"\.comp\b|\.init\b"),
    # violates: no aggregate production (`count : `, `sum w : `, etc., the
    # aggregate variable name is optional between the keyword and ':') --
    # arith has no aggregate form.
    "aggregate": re.compile(r"\b(count|sum|max|min|mean)\b\s*[A-Za-z_]?\w*\s*:"),
    # violates: term ::= arith | '_' -- arith's primary has no function-call
    # form; '.functor' declares one, '@name(' invokes a user functor.
    "functor_call": re.compile(r"\.functor\b|@\w+\s*\("),
    # violates: type ::= 'number' | 'symbol' only -- no record/list type form.
    "record_or_list_term": re.compile(r"\.type\s+\w+\s*=\s*\["),
    # violates: no choice-domain production exists in body/literal.
    "choice_domain": re.compile(r"choice-domain"),
    # violates: type ::= 'number' | 'symbol' only -- no ADT/union type form.
    "adt": re.compile(r"\.type\s+\w+\s*=.*\{"),
    # violates: clause ::= atom '.' | atom ':-' body '.' -- no subsumption
    # ('<=') relation between clauses exists.
    "subsumption": re.compile(r"\)\s*<=\s*[A-Za-z_]\w*\s*\("),
    # violates: decl ::= '.decl' ident '(' param (',' param)* ')' -- nothing
    # may follow the closing paren. (V2 below generalizes this beyond the
    # three literal keywords night01 checked.)
    "decl_qualifier_named": re.compile(r"\.decl\s+\w+\s*\([^)]*\)\s*(btree_delete|overridable|inline)\b"),
    # violates: literal ::= atom | '!' atom | constraint -- no disjunction
    # operator exists in the body grammar.
    "disjunction_semicolon": re.compile(r";"),
}

# --- New categories, VERIFY-01 section V3 plus further gaps found here ---
NEW_PATTERNS = {
    # violates: atom ::= ident '(' term (',' term)* ')' -- at least one term
    # is required; decl ::= '.decl' ident '(' param (',' param)* ')' --
    # at least one param is required. (dlc's own parser has an AUTHORIZED
    # amendment admitting this -- M1-BUILD.md section 3.3 -- but that
    # amendment is a dlc-specific deviation, not part of section 4 itself,
    # which this predicate holds to literally.)
    "zero_arity_decl": re.compile(r"\.decl\s+\w+\s*\(\s*\)"),
    # violates: decl ::= '.decl' ident '(' ... -- exactly one ident, not a
    # comma-separated list of names sharing one param list.
    "multiname_decl": re.compile(r"\.decl\s+\w+(?:\s*,\s*\w+)+\s*\("),
    # violates: decl ::= '.decl' ident '(' param (',' param)* ')' -- nothing
    # may follow the closing paren. Generalizes decl_qualifier_named to ANY
    # trailing identifier, not just the three Soufflé keywords night01 knew
    # to check for.
    "decl_qualifier_any": re.compile(r"\.decl\s+\w+\s*\([^)\n]*\)[ \t]*[A-Za-z_]"),
    # violates: term ::= arith | '_'; arith's primary has no call form at
    # all -- an atom nested inside another atom's argument list (two levels
    # of '(' before the matching ')') is a functor call used as a term,
    # whether user-defined ('.functor') or a Souffle builtin (strlen, cat,
    # to_number, ord, range, ...). Approximate (regex, not a real parser):
    # catches the common one-level-of-nesting shape.
    "functor_call_as_term": re.compile(
        r"[A-Za-z_]\w*\s*\([^()]*[A-Za-z_]\w*\s*\([^()]*\)[^()]*\)"
    ),
    # violates: term ::= arith | '_'; arith ::= ... | mul; mul ::= ... |
    # unary -- a function call is not reachable from arith at all, whether
    # nested inside an atom's args (functor_call_as_term above) or bare as
    # a relop operand (`X = to_number("10")`, `y = range(1,5)`).
    "functor_call_bare_relop_operand": re.compile(
        r"(?:=|!=|<=|>=|<|>)\s*[A-Za-z_]\w*\s*\(|[A-Za-z_]\w*\s*\(\s*[^()]*\)\s*(?:=|!=|<=|>=|<|>)"
    ),
    # violates: clause ::= atom ':-' body '.' -- exactly one head atom, not
    # a comma-separated disjunctive head (two or more atoms before ':-').
    "multi_head_rule": re.compile(
        r"^\s*(?:[A-Za-z_]\w*\s*\([^)\n]*\)\s*,\s*)+[A-Za-z_]\w*\s*\([^)\n]*\)\s*:-", re.MULTILINE
    ),
    # violates: no preprocessor/include production exists in the grammar.
    # Souffle's real form is the dot-directive '.include "file"', not a
    # C-style '#include' (that pattern is also kept, in case a file uses
    # a preprocessor literally).
    "include_directive": re.compile(r"^\s*(?:#include|\.include)\b", re.MULTILINE),
    # violates: type ::= 'number' | 'symbol' -- exactly these two, not
    # Souffle's wider primitive-type set.
    "extra_primitive_type": re.compile(r":\s*(unsigned|float)\b"),
    # violates: decl ::= ... | '.input' ident | '.output' ident -- a bare
    # ident only, no parenthesised option list. (NIGHT-BATCH-03 T6
    # authorized this as a second dlc-parser amendment; strict section 4,
    # unamended, still excludes it -- the same deliberate separation as
    # zero_arity_decl above.)
    "input_output_parens": re.compile(r"\.(input|output)\s+\w+\s*\("),
    # violates: no other directive form exists beyond '.decl'/'.input'/
    # '.output' -- catches the less-common named directives not already
    # covered by a dedicated check above.
    "extra_directive_other": re.compile(r"^\s*\.(override|printsize|index)\b", re.MULTILINE),
    # violates: decl ::= '.decl' | '.input' | '.output' only -- a catch-all
    # for ANY other '.identifier' directive-shaped token, including ones
    # with no real Souffle meaning at all (deliberately-invalid directive
    # names used to test error-path reporting, e.g. '.bad'/'.notexpected').
    # This is the general form every specific directive check above is a
    # named special case of.
    "unknown_directive_other": re.compile(
        r"^\s*\.(?!decl\b|input\b|output\b|type\b|pragma\b|comp\b|init\b|functor\b|plan\b|include\b|override\b|printsize\b|index\b)[A-Za-z_]\w*",
        re.MULTILINE,
    ),
    # violates: primary ::= var | number | string | '(' arith ')' -- no
    # list/record literal ('[' ... ']') production exists as a term, in
    # addition to record_or_list_term above (which only catches the
    # '.type X = [...]' declaration form, not a list literal used directly
    # as a term, e.g. `[a] = [1]`).
    "list_literal_term": re.compile(r"\[[^\[\]]*\]"),
    # violates: number is an integer literal (arith's `unary`/`primary`
    # give no decimal-point production) -- a bare integer immediately
    # followed by '.' and more digits (a float literal, or a dotted
    # multi-segment literal like an IPv4 address) is not representable.
    "float_or_dotted_number_literal": re.compile(r"\b\d+\.\d+\b"),
    # violates: literal ::= atom | '!' atom | constraint -- a bare
    # identifier with no argument list used directly as a body literal
    # (Souffle's `true`/`false` boolean-literal sugar) is not an `atom`
    # (which requires a parenthesised term list).
    "bare_boolean_literal": re.compile(r"(?::-|,)\s*(?:true|false)\s*[,.]"),
}

STRING_LIT_RE = re.compile(r'"((?:[^"\\]|\\.)*)"')
ESCAPE_RE = re.compile(r"\\(.)")


def has_unsupported_string_escape(text):
    """primary ::= ... | string -- section 4 does not define an escape
    grammar for the `string` terminal at all. This project's own lexer
    (src/lexer/lexer.go: "lexString ... Supports \\" and \\\\ as escapes")
    is the closest thing to an operative definition of what `string` means
    here -- any other backslash escape (\\n, \\t, \\xNN, ...) is therefore
    outside what this project's `string` terminal can represent faithfully,
    a distinct and disclosed category, not folded into a vague "other"."""
    for m in STRING_LIT_RE.finditer(text):
        content = m.group(1)
        for esc in ESCAPE_RE.finditer(content):
            if esc.group(1) not in ('"', "\\"):
                return True
    return False


def has_unterminated_block_comment(raw_text):
    """Not a grammar-production violation -- a lexer-level disagreement
    (src/lexer/DESIGN.md's own documented, deliberate divergence: dlc
    rejects an unterminated '/*', real Souffle silently swallows it to
    EOF). Counted on the RAW (pre-strip_comments) text, since a properly
    terminated comment is removed before this predicate's other checks
    ever see it, but an unterminated one is not (strip_comments' own
    regex requires a matching '*/' to match at all)."""
    return raw_text.count("/*") > raw_text.count("*/")


def classify(text: str, raw_text: str = None):
    found = {}
    for name, pat in {**BASELINE_PATTERNS, **NEW_PATTERNS}.items():
        n = len(pat.findall(text))
        if n:
            found[name] = n
    if has_unsupported_string_escape(text):
        found["unsupported_string_escape"] = 1
    if raw_text is not None and has_unterminated_block_comment(raw_text):
        found["unterminated_block_comment"] = 1
    return found


def main():
    OUT_DIR.mkdir(parents=True, exist_ok=True)
    dlc_interface.build_dlc()

    results = {}
    for dirpath, _dirnames, filenames in os.walk(SOUFFLE_TESTS):
        for fn in sorted(filenames):
            if not fn.endswith(".dl"):
                continue
            fp = Path(dirpath) / fn
            rel = str(fp.relative_to(SOUFFLE_TESTS))
            raw = fp.read_text(errors="replace")
            text = strip_comments(raw)
            feats = classify(text, raw_text=raw)
            results[rel] = {"in_grammar_v2": len(feats) == 0, "features": feats}

    total = len(results)
    in_grammar_v2 = sorted(k for k, v in results.items() if v["in_grammar_v2"])
    out_grammar = [k for k, v in results.items() if not v["in_grammar_v2"]]

    histogram = {}
    for k in out_grammar:
        for feat in results[k]["features"]:
            histogram[feat] = histogram.get(feat, 0) + 1

    # Cross-check: every file admitted by V2 must actually parse under dlc.
    parse_results = {}
    parsed_count = 0
    for rel in in_grammar_v2:
        fp = SOUFFLE_TESTS / rel
        r = dlc_interface.run_dlc_parse(fp.read_text(errors="replace"))
        parse_results[rel] = r.status
        if r.status == "parsed":
            parsed_count += 1

    (OUT_DIR / "detail.json").write_text(json.dumps(results, indent=2, sort_keys=True))
    (OUT_DIR / "summary.json").write_text(json.dumps({
        "total_dl_files": total,
        "in_grammar_v2_count": len(in_grammar_v2),
        "out_of_grammar_count": len(out_grammar),
        "histogram": dict(sorted(histogram.items(), key=lambda kv: -kv[1])),
        "cross_check_parsed": parsed_count,
        "cross_check_total": len(in_grammar_v2),
        "cross_check_parse_results": parse_results,
    }, indent=2))

    IN_GRAMMAR_V2_PATH.write_text(
        "# In-grammar file census V2 (NIGHT-BATCH-03 T7), corrected predicate.\n"
        "# Superset check for tests/corpus/IN_GRAMMAR.txt's 12-category predicate,\n"
        "# extended with the categories VERIFY-01 section V3 found missing (zero-arity\n"
        "# decls, multi-name decls, arbitrary decl qualifiers, functor calls as terms,\n"
        "# multi-head rules, #include, extra primitive types, .input/.output parens,\n"
        "# unsupported string escapes), each justified against the specific blueprint\n"
        "# section 4 production it violates -- see harness/night03_t7_grammar_v2.py's\n"
        "# own comments, never against which files failed.\n"
        "# Both files survive: IN_GRAMMAR.txt is untouched and stays committed.\n"
        "# Predicate source: harness/night03_t7_grammar_v2.py.\n"
        + "\n".join(in_grammar_v2) + "\n"
    )

    print(f"total={total} in_grammar_v2={len(in_grammar_v2)} out_of_grammar={len(out_grammar)}", file=sys.stderr)
    print(f"cross-check: parsed/{len(in_grammar_v2)} = {parsed_count}/{len(in_grammar_v2)}", file=sys.stderr)
    print(json.dumps(histogram, indent=2))


if __name__ == "__main__":
    main()
