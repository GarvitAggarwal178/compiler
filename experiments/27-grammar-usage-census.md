# NIGHT-BATCH-02 T8 — grammar usage census over `IN_GRAMMAR.txt`

Date: 2026-08-23. Independent task, no Lane A dependency, static analysis
only (no Soufflé invocation). Corpus: the 195 files in
`tests/corpus/IN_GRAMMAR.txt` (NIGHT-BATCH-01 T5's mechanically-verified
in-grammar pool — **not** the pre-registered corpus; that file's own header
says it "has NOT been run"). Runner: `harness/night02_t8_grammar_census.py`.
Full data: `measurements/night02-t8-grammar-census.json`.

**Method, disclosed plainly:** a small hand-written tokenizer for exactly
blueprint §4's grammar (comments stripped first), then structural counting
over the token stream — token-aware, not raw-text regex, so `<=` isn't
double-counted as `<`, and body-literal splitting respects paren nesting.
**Not a full parser.** One disclosed limitation with a concrete
consequence, found while writing this task (see below): recursion
detection here is *direct*-recursion only (a rule's body mentions its own
head relation by name); mutual recursion through an intermediate relation
is invisible to this scan.

## What did not work / findings that need the human's attention

**11 of the 195 "in-grammar" files declare a zero-arity relation**
(`.decl Name()`), which blueprint §4's own grammar does not admit — `decl`
requires `param (',' param)*`, at least one parameter. Confirmed by direct
grep, not just the tokenizer's arity count:

```
evaluation/cprog1/cprog1.dl        .decl Res()
evaluation/cprog2/cprog2.dl        .decl Res()
evaluation/cprog3/cprog3.dl        .decl R()
evaluation/cprog4/cprog4.dl        .decl R()
evaluation/cprog5/cprog5.dl        .decl R()
evaluation/existential/existential.dl        .decl a()  .decl b()
scheduler/functionality/functionality.dl     .decl R()
semantic/rel_nullary/rel_nullary.dl          .decl A()  .decl C()
semantic/string_substr/string_substr.dl      .decl Nullary()
syntactic/doc_comment_dangling1/doc_comment_dangling1.dl   .decl foo()
syntactic/include_directive3/include_directive3.dl         .decl ok()
```

This is a real gap in NIGHT-BATCH-01 T5's mechanical in-grammar predicate
(`harness/night01_t5_grammar.py`), not a contradiction of any blueprint
claim — `IN_GRAMMAR.txt` was built by a text scan, disclosed as
approximate, and T2 (this same batch) independently confirmed Soufflé
itself *accepts* zero-arity relations even though this project's own
restricted grammar does not. **Practical consequence:** once the real
parser exists and correctly enforces "at least one parameter," these 11
files will fail parse-coverage (`harness/parse_coverage.py`) even though
`IN_GRAMMAR.txt` currently lists them as in-grammar — worth knowing before
that gate is read as a parser-quality signal rather than a corpus-quality
one. Not fixed here: `tests/corpus/IN_GRAMMAR.txt` is not
`tests/corpus/BENCHMARK_FAMILY/` or `SCALE_POINTS.json`, so it is not
under night-02's prohibition 0.2, but this task's job is census, not
corpus maintenance — flagged for the human, not corrected in this commit.

**The tokenizer reports one wildcard occurrence in head position**
(`wildcard_positions.head = 1` out of 688 total wildcards) — interesting
given T2 (this same batch) found Soufflé outright *rejects* `_` in head
position (`Underscore in head of rule`). A crude follow-up grep to find
which specific file this is did not reliably isolate it (multi-line rule
bodies defeated the quick single-line heuristic used for the follow-up,
which is cruder than the tokenizer itself) — **left as an open item, not
resolved this task.** If genuine, that file would fail to run standalone
under real Soufflé despite being in the "in-grammar" pool, same class of
issue as the zero-arity finding above.

## Aggregate results

195/195 files analyzed, 0 missing.

### Arithmetic operators (body occurrences, token-counted)

| op | count |
|---|---|
| `-` | 164 |
| `+` | 63 |
| `*` | 44 |
| `/` | 8 |
| `%` | 6 |

### Relational operators

| op | count |
|---|---|
| `=` | 283 |
| `<` | 79 |
| `>` | 46 |
| `!=` | 35 |
| `<=` | 26 |
| `>=` | 25 |

`=` alone (283) outnumbers all five comparison/inequality operators
combined (211) — most constraint use in this pool is equality/binding, not
ordering.

### Expression nesting depth (paren depth within a clause body)

Median: **1**. Max: **3**. Over 195 real files, deeply nested arithmetic
essentially does not occur — `nesting_parens_20deep.dl` (T2, this batch) is
purely a stress case, not a realistic shape.

### Negation

69 negated literals across 28/195 files (14.4% of files use negation at
all). Position within the body, among files with a multi-literal body:

| position | count |
|---|---|
| middle | 32 |
| last | 26 |
| first | 8 |
| only literal | 3 |

Negation appears in the middle or at the end of a body far more often than
at the start — consistent with SIPS-friendly authoring style (bind
variables via positive literals before consuming them in a negated check),
though this census does not verify that connection, only the position
count.

### Wildcards

688 total, 687 in body position, 1 in head position (see open item above).
47/195 files (24.1%) use `_` at all.

### Arity distribution

Median **1**, max **14**.

| arity | count |
|---|---|
| 0 | 25 |
| 1 | 607 |
| 2 | 217 |
| 3 | 65 |
| 4 | 16 |
| 5 | 2 |
| 6 | 6 |
| 7 | 10 |
| 8 | 3 |
| 11 | 2 |
| 14 | 1 |

Unary relations dominate (607 of 954 declared relations, 63.6%). The 25
arity-0 declarations correspond to the 11 files flagged above (some files
declare more than one nullary relation).

### Rule body length (literal count per rule)

Median **2**, max **12**.

| length | count |
|---|---|
| 1 | 536 |
| 2 | 366 |
| 3 | 109 |
| 4 | 41 |
| 5 | 29 |
| 6 | 18 |
| 7 | 9 |
| 8 | 5 |
| 9 | 5 |
| 11 | 1 |
| 12 | 2 |

Single-literal bodies (536) are the single largest bucket — many clauses
in this pool are simple facts or near-trivial pass-through rules.

### Literals: string vs number

677 number literals, 32 string literals — numeric constants dominate by a
wide margin in this pool (21:1).

### Rules per relation, recursion

Median rules-per-relation: **1**. 954 distinct relation names seen across
the pool; **61 relations (6.4%) are directly recursive**, 893 (93.6%) are
not, by this scan's direct-recursion-only detection (disclosed limitation
above — mutual recursion is not counted, so the true recursive fraction is
a lower bound, not an exact count).

## What a skeptic attacks first

- The zero-arity-relation finding is the most consequential thing in this
  report and is not this task's to fix — a skeptic should ask why
  `IN_GRAMMAR.txt` was trusted as a parser-testing pool at all before this
  gap was known, not just note the gap now.
- The recursive-relation count (61) is a **lower bound**, stated as such —
  mutual recursion is real in Datalog test suites (SG-style shapes, this
  project's own `same_generation_negation` included) and this scan cannot
  see it.
- The one head-wildcard occurrence was not tracked down to a specific
  file; a skeptic should treat it as "reported by the tokenizer, not
  independently confirmed" rather than a settled fact.
- This is a text-level scan, not the real parser; anywhere Soufflé's full
  grammar differs subtly from blueprint §4 in ways this tokenizer doesn't
  model (e.g., string literal escaping, comment edge cases) could bias
  individual counts by a small amount. Not expected to change any of the
  medians/max values materially, but not proven either.

## Provenance

`measurements/night02-t8-grammar-census.json` (full aggregate + per-file
breakdown). Runner: `harness/night02_t8_grammar_census.py`. Follow-up greps:
`measurements/_scratch_night02_t8/`. Completed inside the 60-minute cap.
