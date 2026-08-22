"""
Rejection ground: allowedness / range restriction. Lane B test bodies --
see arity.py's docstring for the shared discipline. Blueprint section 6
notes the definition of allowedness implemented in M1 is itself a
hypothesis of the M3 soundness theorem, not hygiene -- these cases exist
to pin down which definition M1 actually implements.

Cross-validated against J1 (docs/reports/J1-allowedness-probe.md,
2026-08-22), which ran the equivalent minimal shapes directly against
Soufflé 2.5 and recorded its exact diagnostic text. All three original
cases below match a J1 outcome exactly (head-var-unbound generalizes
J1's ungrounded-variable rejections; var-only-in-negation is J1 case f;
var-only-in-constraint is J1 case c). One case added
(allowedness_equation_rhs_not_bound) directly from J1's least obvious
finding: case (d) `Y = X + 1` rejects even though `X` sits on the other
side of an equation from an already-grounded `Y` -- equality is not
solved/inverted to find what it grounds, only case (a)'s shape (`X =
<expr over grounded vars>`, target variable on the left) grounds.
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
            "domain by themselves. Reject: y is not range-restricted. "
            "Matches J1 case (c) exactly: `p(X) :- q(Y), X > Y.` rejects "
            "with `Ungrounded variable X`."
        ),
    },
    {
        "name": "allowedness_equation_rhs_not_bound",
        "program": """\
.decl foo(a:number)
.input foo
.decl bar(x:number, y:number)
.output bar
bar(x, y) :- foo(x), x = y + 1.
""",
        "expected_ground": "allowedness",
        "expected_diagnosis": (
            "y appears in the head and on the right-hand side of an "
            "equation (x = y + 1) whose left-hand side (x) is already "
            "grounded by foo(x) -- but grounding does not flow backward "
            "through an equation from an already-grounded left-hand side "
            "to an unbound variable on the right. Reject: y is not "
            "range-restricted. Direct restatement of J1 case (d); this "
            "exact program independently re-verified against Soufflé "
            "2.5, which rejects it with `Ungrounded variable y`."
        ),
    },
]
