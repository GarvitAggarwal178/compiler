# NIGHT-BATCH-01 — T5: grammar coverage census

Date: 2026-08-20. Outcome: **done**, no prerequisite, full coverage (no execution
needed — pure mechanical text scan against blueprint §4's grammar, unchanged v1.0
through v1.2).

## Method

`harness/night01_t5_grammar.py`, every `.dl` file under Soufflé's `tests/` tree
(622 files — slightly more than the 612 `.dl`-*bearing directories* T2/T4 count,
since a handful of directories contain more than one `.dl` file). Detects 11
out-of-grammar feature categories by regex over comment-stripped source: `.type`
declarations, record/list-typed declarations, components (`.comp`/`.init`),
`.pragma`, aggregates (`count`/`sum`/`max`/`min`/`mean` `:`), `.decl` qualifiers
(`btree_delete`/`overridable`/`inline`), algebraic data types, disjunction (`;`),
functor calls, `choice-domain`, and subsumption (`<=` between two atoms — the exact
feature T3 found causing a genuine divergence, `docs/ESCALATIONS.md`).

**Caveat, same as `corpus_predicate.py`'s:** this is a text scan, not dlc's own
parser (M1, Lane A, doesn't exist yet). It can mis-classify at the margin. It is
good enough to fix a starting differential-test pool before the parser lands, not a
claim of parser-grade precision.

## Results

| | Count | Rate |
|---|---|---|
| Total `.dl` files | 622 | — |
| **Fully in-grammar** | **195** | **31.4%** |
| Out-of-grammar (≥1 feature detected) | 427 | 68.6% |

## Out-of-grammar feature histogram (of the 427 out-of-grammar files)

| Feature | File count | Share of out-of-grammar files |
|---|---|---|
| `.type` declaration | 291 | 68.1% |
| Record/list-typed declaration | 70 | 16.4% |
| Component (`.comp`/`.init`) | 58 | 13.6% |
| `.pragma` directive | 53 | 12.4% |
| Aggregate (`count`/`sum`/`max`/`min`/`mean`) | 33 | 7.7% |
| `.decl` qualifier (`btree_delete`/`overridable`/`inline`) | 31 | 7.3% |
| Algebraic data type | 25 | 5.9% |
| Disjunction (`;`) | 21 | 4.9% |
| Functor call | 14 | 3.3% |
| `choice-domain` | 10 | 2.3% |
| Subsumption (`<=`) | 8 | 1.9% |
| `.plan` directive | 7 | 1.6% |

**`.type` declarations are the single largest exclusion factor by nearly 4×** over
the next most common feature. This matters for M1's expectations: even the
mildest, most common real-world pattern — a plain subtype alias like `.type D <:
symbol` — is enough by itself to put a program outside blueprint §4's grammar,
since §4 has no `.type` production at all (only bare `number`/`symbol` inline in
`param`). A parser that only ever sees the 31.4% "fully in-grammar" pool will never
encounter this, the most common real-world grammar gap, in its own test corpus.
Worth a note for whoever decides M1's test-writing priorities, not acted on here.

## Output

`tests/corpus/IN_GRAMMAR.txt` — 195 file paths (relative to Soufflé's `tests/`
root), header comment makes explicit this is **not**
`tests/corpus/PREREGISTERED.txt` and has not been run, checked for negation, or
checked for seedability. It exists to give M1's parser something to differentially
test against before it's finished — a much larger and more grammar-representative
pool than the 36-program pre-registered corpus, which was filtered for a completely
different purpose (negation + seedability, not grammar coverage).

Full per-file detail (which features, if any, each file triggered):
`measurements/night01-t5/detail.json`. Summary: `measurements/night01-t5/summary.json`.
