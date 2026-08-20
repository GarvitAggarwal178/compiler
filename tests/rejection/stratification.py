"""
Rejection ground: unstratifiable negation. Lane B test bodies -- see
arity.py's docstring for the shared discipline. `semantic/rel_stratification`
in Souffle's own test suite (found via NIGHT-BATCH-01 T2,
docs/reports/night01-T2-corpus.md) uses the same minimal self-cycle shape
as the first case here, confirming this is a real thing Soufflé itself
tests for, not an invented edge case.
"""

CASES = [
    {
        "name": "stratification_self_negative_cycle",
        "program": """\
.decl a(x:number)
.input a
a(x) :- a(x), !a(x).
""",
        "expected_ground": "stratification",
        "expected_diagnosis": (
            "a depends negatively on itself in a single rule "
            "(a(x):-a(x),!a(x).) -- a trivial 1-node cycle with a negative "
            "edge. Reject: unstratifiable."
        ),
    },
    {
        "name": "stratification_mutual_negative_cycle",
        "program": """\
.decl base(x:number)
.input base
.decl p(x:number)
.decl q(x:number)
.decl out(x:number)
.output out
p(x) :- base(x), !q(x).
q(x) :- base(x), !p(x).
out(x) :- p(x).
""",
        "expected_ground": "stratification",
        "expected_diagnosis": (
            "p negatively depends on q, and q negatively depends on p -- a "
            "2-node cycle with two negative edges, no valid stratum "
            "ordering exists. Reject: unstratifiable."
        ),
    },
    {
        "name": "stratification_cycle_through_positive_chain",
        "program": """\
.decl base(x:number)
.input base
.decl p(x:number)
.decl q(x:number)
.decl r(x:number)
.decl out(x:number)
.output out
p(x) :- base(x), !q(x).
q(x) :- r(x).
r(x) :- p(x).
out(x) :- p(x).
""",
        "expected_ground": "stratification",
        "expected_diagnosis": (
            "p negatively depends on q; q positively depends on r; r "
            "positively depends on p -- closing a cycle (p -> ~q -> r -> p) "
            "with exactly one negative edge inside it. This is the same "
            "'culprit cycle' shape blueprint section 6 clause (a) and P5 "
            "(section 12) are about, at the source-program level rather "
            "than a magic-transform-introduced one. Reject: unstratifiable."
        ),
    },
]
