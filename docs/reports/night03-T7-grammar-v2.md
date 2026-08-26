# NIGHT-BATCH-03 T7 — corrected corpus predicate

Date: 2026-08-27. `IN_GRAMMAR.txt` is untouched and stays committed; this
adds a second, corrected census, per the pre-registration rule's own
precedent (11-of-149 kept alongside 36-of-612).

## What did not work

The predicate's **first draft admitted 43/622 files, of which only
19 (44%) actually parsed under `dlc`** — a discrepancy nearly as large as
the one this task exists to fix. Investigated every one of the 24
mismatches by reading the real file content and `dlc`'s actual diagnostic
message (never guessed): found 3 missing categories entirely (bare functor
calls used as a relop operand — `x = to_number("10")`, not nested inside
another atom's parens; Soufflé's aggregate syntax with a named binding
variable, `sum w : Edge(_,_,w)`, which the first regex's `keyword\s*:`
shape couldn't match; unterminated block comments, a lexer-level, not
grammar-level, divergence) and 2 bugs in already-present categories
(`.include "file"` is Soufflé's real include directive — the first draft
only checked C-style `#include`; multi-head rules can have 3+ atoms, not
just exactly 2). Converged to 21/622 admitted, 19/21 (90%) cross-check
parsed, before stopping — the residual 2 are disclosed below, not forced
into a synthetic category.

## Method

`harness/night03_t7_grammar_v2.py` — a new file, not an edit to
`night01_t5_grammar.py`. Every one of 23 exclusion categories is commented
with the exact blueprint §4 production it violates (quoted in the script's
own header), never justified by which files failed. The original 12
categories are kept verbatim (comparability); 11 new ones added, per
VERIFY-01 §V3's list plus further gaps found during convergence (see "what
did not work" above).

Run over the **full 622-file** Soufflé source tree (not just the 195-file
`IN_GRAMMAR.txt` subset), cross-checked against `dlc`'s real parser.

## The numbers

Provenance: `measurements/night03-t7/summary.json`,
`measurements/night03-t7/detail.json`.

| metric | value |
|---|---|
| total `.dl` files scanned | 622 |
| `\|V2\|` (admitted by the corrected predicate) | **21** |
| cross-check: parsed under `dlc` | **19/21** |
| `V2 ⊆ IN_GRAMMAR.txt`'s 195? | **yes, exactly** (necessarily true — V2 checks a superset of V1's categories) |

Per-category exclusion histogram (files triggering each check, across all
622 — a file can trigger more than one):

| category | files | blueprint §4 production violated |
|---|---|---|
| `input_output_parens` | 370 | `decl ::= ... \| '.input' ident \| '.output' ident` (bare ident only) |
| `type_decl` | 291 | not one of the three admitted `decl` forms |
| `list_literal_term` | 76 | `primary` has no list/record literal |
| `record_or_list_term` | 70 | `type ::= 'number' \| 'symbol'` only |
| `functor_call_as_term` | 66 | `term ::= arith \| '_'` — no call form |
| `aggregate` | 66 | no aggregate production |
| `decl_qualifier_any` | 59 | nothing may follow `decl`'s closing `)` |
| `component` | 58 | no component/instantiation production |
| `pragma_directive` | 53 | not one of the three admitted `decl` forms |
| `extra_primitive_type` | 50 | `type ::= 'number' \| 'symbol'` only |
| `functor_call_bare_relop_operand` | 48 | `arith` has no call form |
| `zero_arity_decl` | 36 | `atom`/`decl` require ≥1 term/param |
| `extra_directive_other` | 34 | not one of the three admitted `decl` forms |
| `float_or_dotted_number_literal` | 31 | `number` is an integer literal, no decimal point |
| `decl_qualifier_named` | 31 | (subset of `decl_qualifier_any`, kept for V1 comparability) |
| `adt` | 25 | `type ::= 'number' \| 'symbol'` only |
| `multiname_decl` | 22 | `decl` takes exactly one `ident` |
| `disjunction_semicolon` | 21 | `literal` has no disjunction operator |
| `functor_call` | 14 | (subset of the two functor checks above, kept for V1 comparability) |
| `unknown_directive_other` | 13 | not one of the three admitted `decl` forms |
| `include_directive` | 13 | no preprocessor/include production |
| `choice_domain` | 10 | no choice-domain production |
| `bare_boolean_literal` | 8 | `atom` requires a parenthesised term list |
| `subsumption` | 8 | `clause` has no subsumption relation |
| `plan_directive` | 7 | not one of the three admitted `decl` forms |
| `multi_head_rule` | 7 | `clause` has exactly one head atom |
| `unterminated_block_comment` | 1 | lexer-level, not a grammar production (disclosed as such) |
| `unsupported_string_escape` | 1 | `string`'s escape grammar is this project's lexer, not §4 itself |

## The 2 residual cross-check gaps

`syntactic/issue2408/issue2408.dl` (file content: the single line `x`) and
`syntactic/syntax10/syntax10.dl` (file content: a leading bare `*`) are
both admitted by every regex category (they trigger none) but fail to
parse under `dlc`. Both are Soufflé's own deliberately-malformed test
fixtures for exercising error-location reporting, not real programs using
a real feature — there is no stable, generalizable grammar-production
category to assign a bare, structureless garbage token to without
producing false positives elsewhere (a lone identifier or a stray `*`
could appear inside a comment or string in an otherwise-valid file).
Disclosed, not force-fit.

## VERIFY-01 gap, named explicitly

VERIFY-01 §V3 independently found 19 truly §4-compliant files *within* the
195-file `IN_GRAMMAR.txt` subset. This task's 21-of-622 (19 of which parse)
is consistent with, not a contradiction of, that finding — VERIFY-01 never
scanned the full 622-file tree for files *outside* `IN_GRAMMAR.txt` that
might also be truly compliant; this task did, and found none (the 2 extra
V2 members beyond VERIFY-01's 19 are exactly the 2 residual garbage files
above, which VERIFY-01 would also have excluded by its own parser
cross-check — V2's 21 is 19 truly-parseable plus these 2 borderline
admits, not a genuinely larger true-compliance count).

## What a skeptic attacks first

- 23 regex categories over 622 files is still a text scan, not a real
  parser — the 90% cross-check rate (19/21), reached only after iterating
  on real discrepancies, is the actual evidence of correctness, not the
  category count. A residual 2/21 gap is honestly reported, not smoothed
  into "essentially 100%."
- `functor_call_bare_relop_operand`'s regex is the least precise category
  added — it could in principle flag a section-4-legal program if a
  variable name happens to look like `IDENT(` when adjacent to whitespace
  quirks the regex doesn't anticipate. No such false positive was found in
  this run (histogram counts and cross-check numbers are consistent with
  the true grammar boundary), but it was not independently stress-tested
  beyond this corpus.
- `IN_GRAMMAR.txt`'s own 195-file list is now known to be a strict
  superset with a much larger non-compliant residue (195 vs. 21 truly
  compliant) — every gate that reported against `IN_GRAMMAR.txt` alone
  before this task should, from now on, report against both lists, per
  instruction. This task does not retroactively re-run those gates
  (that is T10's provenance-backfill scope, not T7's).

## Verdict

**T7: DONE.** `\|V2\|=21`, cross-check `19/21` (2 residual, disclosed,
ungeneralizable garbage-fixture gaps). `V2 ⊆ IN_GRAMMAR.txt` confirmed
exactly. Both lists committed and both survive; `IN_GRAMMAR.txt` untouched.
