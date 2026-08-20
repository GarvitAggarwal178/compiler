"""
Rejection ground: type mismatch. Lane B test bodies -- see arity.py's
docstring for the shared discipline (Lane A checker doesn't exist yet,
these must currently fail "not implemented", never pass vacuously).
"""

CASES = [
    {
        "name": "type_number_vs_symbol_across_rule",
        "program": """\
.decl foo(a:number, b:symbol)
.input foo
.decl bar(x:number)
.output bar
bar(y) :- foo(x, y).
""",
        "expected_ground": "type",
        "expected_diagnosis": (
            "y is bound to foo's second argument, declared symbol; y then "
            "flows into bar's first argument, declared number. Reject: "
            "type mismatch, symbol used where number is required."
        ),
    },
    {
        "name": "type_symbol_in_arithmetic",
        "program": """\
.decl foo(a:symbol)
.input foo
.decl bar(x:number)
.output bar
bar(y) :- foo(x), y = x + 1.
""",
        "expected_ground": "type",
        "expected_diagnosis": (
            "x is declared symbol (foo's argument); x is used in an "
            "arithmetic expression (x + 1), which requires number. Reject: "
            "type mismatch, symbol used in arithmetic context."
        ),
    },
    {
        "name": "type_mismatched_fact_literal",
        "program": """\
.decl foo(a:number)
.output foo
foo("not_a_number").
""",
        "expected_ground": "type",
        "expected_diagnosis": (
            "foo's argument is declared number; the fact supplies a string "
            "literal. Reject: type mismatch on a fact literal, not just a "
            "variable's inferred type."
        ),
    },
]
