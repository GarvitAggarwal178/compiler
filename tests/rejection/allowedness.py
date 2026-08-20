"""
Rejection ground: allowedness / range restriction. Lane B test bodies --
see arity.py's docstring for the shared discipline. Blueprint section 6
notes the definition of allowedness implemented in M1 is itself a
hypothesis of the M3 soundness theorem, not hygiene -- these cases exist
to pin down which definition M1 actually implements.
"""

CASES = [
    {
        "name": "allowedness_head_var_unbound",
        "program": """\
.decl foo(a:number)
.input foo
.decl bar(x:number, y:number)
.output bar
bar(x, y) :- foo(x).
""",
        "expected_ground": "allowedness",
        "expected_diagnosis": (
            "y appears in the head bar(x, y) but nowhere in the body -- no "
            "positive literal binds it. Reject: y is not range-restricted."
        ),
    },
    {
        "name": "allowedness_var_only_in_negation",
        "program": """\
.decl foo(a:number)
.input foo
.decl baz(a:number, b:number)
.input baz
.decl bar(x:number, y:number)
.output bar
bar(x, y) :- foo(x), !baz(x, y).
""",
        "expected_ground": "allowedness",
        "expected_diagnosis": (
            "y appears in the head and in a negated literal !baz(x,y), but "
            "in no *positive* literal. Negation cannot range-restrict a "
            "variable under any of the allowedness definitions blueprint "
            "section 6 discusses. Reject: y is not range-restricted."
        ),
    },
    {
        "name": "allowedness_var_only_in_constraint",
        "program": """\
.decl foo(a:number)
.input foo
.decl bar(x:number, y:number)
.output bar
bar(x, y) :- foo(x), y > 0.
""",
        "expected_ground": "allowedness",
        "expected_diagnosis": (
            "y appears in the head and in a constraint (y > 0), but "
            "constraints are not atoms and do not bind a variable's "
            "domain by themselves. Reject: y is not range-restricted."
        ),
    },
]
