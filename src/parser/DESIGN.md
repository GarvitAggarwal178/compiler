# package parser

Recursive descent over `clause`/`decl`/`atom`/`literal` structure;
precedence climbing over `arith` (tightest first: unary `-`; `* / %`;
`+ -`). Both by hand, no generator, per M1-BUILD.md §2.

**Non-obvious decision: `decl* clause*` is parsed as unordered/
interleaved, not strict staging.** Read as literal BNF, blueprint §4's
grammar requires every `.decl`/`.input`/`.output` to precede every fact/
rule in the file. This project's own `culprit_cycle.dl` (pre-registered,
already validated against real Soufflé this project's own measurements
depend on) declares `q`, writes `q`'s two rules, declares `s`, writes
`s`'s rule, declares `p`, writes `p`'s two rules, declares `out`, writes
`out`'s rule — interleaved throughout, not staged. A parser that enforced
strict staging would reject that file and, by extension, almost every
realistic Datalog program (declaring a relation right where it is first
needed is the normal way to write one). Read the grammar's `*`s as "any
number of, in any order" rather than literal Kleene sequencing — resolved
by evidence from an already-oracle-validated program already in this
repo, not by inventing a new grammar amendment (the one amendment
M1-BUILD.md §3.3 authorizes is the optional term list; this is a reading
of ambiguous BNF shorthand, not a second amendment).

**Non-obvious decision: atom-vs-constraint disambiguation is one token of
lookahead.** `literal ::= atom | '!' atom | constraint`, and both `atom`
and `constraint` can start with what looks like a bare identifier (`X` in
`X > 0` and `foo` in `foo(X)`). Resolved the same way most recursive-
descent parsers resolve this class of ambiguity: `IDENT` immediately
followed by `(` is an atom; anything else is parsed as `arith` and must
be followed by a relop to close out a `constraint`. `!` unconditionally
starts a negated atom (the grammar has no `! constraint`).

**Non-obvious decision: `.input`/`.output` do not accept a trailing
`()`.** Real Soufflé test files (e.g. `evaluation/access1/access1.dl`,
found while implementing this item) commonly write `.input Foo()` with
cosmetic empty parens. Blueprint §4's grammar says `.input ident` — no
parens — and M1-BUILD.md §3.3 authorizes exactly one amendment (the term-
list-in-an-atom one), not this. `dlc` therefore rejects the parenthesized
form; this is a second, independent reason (beyond the aggregate/record/
`#include`/pragma gap already found at the lexer stage, §3.1,
`docs/OPEN_QUESTIONS.md` 2026-08-26) that gate one's literal "195/195"
will not be hit — see `docs/reports/m1-progress.md` for the actual count
and its breakdown by cause.

**Error recovery is per top-level construct, not per-literal.** §3.3
specifies "skip to the next `.` at clause level" — applied uniformly to a
failed `decl` as well as a failed `clause`, even though a `decl`
production has no `.` terminator of its own. The tradeoff: a badly
malformed `.decl` can, in the worst case, swallow a subsequent valid
clause's period along with it before recovery finds the next real one.
Accepted because the actual guarantee needed ("one malformed clause must
not kill the file") holds regardless — a run of well-formed clauses after
one bad one still parses — and per-literal recovery was not what was
asked for.

**The pretty-printer over-parenthesizes on purpose.** Every nested
`BinaryExpr`/`UnaryExpr` operand is wrapped in parens unconditionally,
regardless of whether real operator precedence would require it (e.g.
`2 + 3 + 4` prints as `(2 + 3) + 4`). This makes the printer trivially
round-trip-correct — there is no precedence table to get subtly wrong
inside the printer itself, which would be exactly the kind of bug §3.3
gate two ("this catches precedence bugs that reading the code never
will") exists to catch, except self-inflicted by the printer instead of
the parser. A minimal-parens printer is future work, not a §3.3
requirement.

**`ast.Equal`, not the printed text, decides gate two.** Two different
correct programs can print to different-looking-but-equivalent text (this
printer doesn't even attempt minimality), so byte-comparing `printed`
against the original source was never going to work as the actual check;
`Roundtrip` (in `roundtrip.go`) reparses the printed text and compares
the two ASTs structurally instead, which is what `ast.Equal` was built
for (`ast/DESIGN.md`).
