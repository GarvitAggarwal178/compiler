# In-grammar census V2 source (NIGHT-BATCH-03 T7)

Corrected version of `IN_GRAMMAR.txt`'s predicate — NOT a replacement.
**Both files survive**, exactly as this project kept 11-of-149 alongside
36-of-612 (`docs/reports/corpus-ruling-2026-08-21.md` precedent).
`IN_GRAMMAR.txt` is untouched and stays committed.

**Why a second predicate, not a fix to the first:** VERIFY-01 established
that the true blueprint §4-compliant count is far below the 195
`night01_t5_grammar.py` admits, because that predicate checks 12 construct
categories while §4 implies at least nine more. The exclusion criterion
(blueprint §4's grammar) was fixed in advance and never changed; the
mechanical predicate implementing it was incomplete. This is the same class
of fix as `corpus_predicate.py`'s sort-determinism fix
(`docs/SESSION_LOG.md`, 2026-08-21) — a bug in a predicate, not a
re-litigation of the criterion itself.

**Source tree:** same as `SOURCE.md` — Soufflé `2.5`, sparse-checked-out
`tests/` subtree, 622 `.dl` files total, `/root/souffle-src/tests/` on this
machine.

**Predicate:** `harness/night03_t7_grammar_v2.py` (a new file, not an edit
to `night01_t5_grammar.py`). 23 total exclusion categories: the original
12, each kept verbatim and re-commented with the exact blueprint §4
production it violates, plus 11 new ones (zero-arity decls, multi-name
decls, arbitrary decl qualifiers, functor calls nested inside atom args,
functor calls as a bare relop operand, multi-head rules (2+ atoms),
`.include`/`#include`, extra primitive types, `.input`/`.output` parens,
unsupported string escapes, list/record literal terms, float/dotted
number literals, bare boolean literals, unknown/arbitrary directive names,
and unterminated block comments as a disclosed lexer-level — not
grammar — divergence). See that file's own inline comments for the exact
production cited per check; never justified by which files failed.

**Result:** `IN_GRAMMAR_V2.txt`, 21 of 622 files. Confirmed a strict subset
of `IN_GRAMMAR.txt`'s 195 (necessarily so — V2 checks a superset of V1's
categories). Cross-checked against `dlc`'s real parser: 19/21 parse. The 2
residual files are hostile/deliberately-malformed test fixtures
(`syntactic/issue2408/issue2408.dl` is the single line `x`;
`syntactic/syntax10/syntax10.dl` starts with a bare `*`) with no
generalizable grammar-production category to assign them to — disclosed as
a known, accepted gap rather than force-fit into a synthetic category (see
`docs/reports/night03-T7-grammar-v2.md`).

**Process note:** the predicate converged iteratively — first pass:
43/622 admitted, 19/43 (44%) cross-check parsed. Investigated every
discrepancy by reading the actual file content and the real `dlc` parser's
diagnostic message (not guessed), found 3 more missing categories (bare
functor-as-relop-operand, aggregate syntax with a named binding variable,
unterminated block comments) plus 2 fixes to already-present categories
(`.include` directive, multi-head rules with 3+ atoms), converging to
21/622 admitted, 19/21 cross-check parsed, before stopping.
