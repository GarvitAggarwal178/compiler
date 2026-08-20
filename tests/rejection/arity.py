"""
Rejection ground: arity mismatch. Lane B test *bodies* -- the checker
that must actually reject these is Lane A (decl/arity check, M1) and
does not exist yet. See harness/run_rejection_tests.py: every case here
must currently report "blocked: dlc not implemented", never a silent
pass and never a fabricated pass/fail verdict from this side.
"""

CASES = [
    {
        "name": "arity_too_few_args",
        "program": """\
.decl foo(a:number, b:number)
.input foo
.decl bar(x:number)
.output bar
bar(x) :- foo(x).
""",
        "expected_ground": "arity",
        "expected_diagnosis": (
            "foo is declared with arity 2 (a:number, b:number); the body "
            "literal foo(x) supplies only 1 argument. Reject: arity mismatch "
            "on foo at the bar rule."
        ),
    },
    {
        "name": "arity_too_many_args",
        "program": """\
.decl foo(a:number)
.input foo
.decl bar(x:number, y:number)
.output bar
bar(x, y) :- foo(x, y).
""",
        "expected_ground": "arity",
        "expected_diagnosis": (
            "foo is declared with arity 1 (a:number); the body literal "
            "foo(x, y) supplies 2 arguments. Reject: arity mismatch on foo."
        ),
    },
    {
        "name": "arity_mismatch_between_fact_and_decl",
        "program": """\
.decl foo(a:number, b:number)
.output foo
foo(1, 2, 3).
""",
        "expected_ground": "arity",
        "expected_diagnosis": (
            "foo is declared with arity 2; the fact foo(1,2,3) supplies 3 "
            "arguments. Reject: arity mismatch on a fact, not just a rule body."
        ),
    },
]
