# NIGHT-BATCH-02 T2 — hostile source corpus

Date: 2026-08-23. Independent task, no Lane A dependency, text files only.
39 `.dl` files under `tests/hostile/`, each run against installed Soufflé
2.5 with a shared empty facts directory (`q.facts`, `r.facts`, `p.facts` —
empty, arity-agnostic; several files reuse relation name `q` at different
arities, so real sample data would arity-mismatch whichever file expects
the other shape). Runner: `harness/night02_t2_hostile.py`. Full data:
`measurements/night02-t2-hostile-summary.json`. **31 accept, 8 reject.**

Every value predicted below was independently confirmed against the actual
`p.csv` output of the corresponding run (`measurements/
_scratch_night02_t2/check_outputs.sh` for the 13 precedence/associativity
files sharing the empty-facts run; a separate one-off run with a real
`q.facts` row for the two files whose empty-facts run produces no rows to
check, `unary_double_minus.dl` and `unary_paren_double_neg.dl`) — not just
inferred from "it parsed."

## What did not work / harness bugs found

- **`harness/probe0.py`'s `run_cmd` crashed on non-UTF-8 stderr.** Soufflé's
  error output for `lexical_unicode_identifiers.dl` includes a raw
  mis-decoded byte (`0xc3` without its continuation byte, from the tool's
  own error-pointer rendering of the malformed token), and Python's default
  strict UTF-8 decode raised `UnicodeDecodeError`, killing the whole batch
  partway through. Fixed: `subprocess.run(..., encoding="utf-8",
  errors="replace")` instead of `text=True`. This is shared J1/T1/T2
  infrastructure — the fix is retroactively safe for previously-recorded
  runs (none of them had this problem) and now protects future ones.
- **First pass mis-classified two cases as "reject" for a harness reason,
  not a language reason**, both from a missing `.facts` file (returncode 1,
  `Error loading <relation> data: Cannot open fact file`) rather than an
  actual parse/semantic rejection: `semantic_head_also_input.dl` (needed
  `p.facts`, not initially provided — fixed, now correctly accepts) and
  `lexical_4kb_identifier.dl` (needs a facts file literally named after its
  own 4,091-byte identifier, which itself exceeds the filesystem's own
  255-byte filename limit — not fixable by providing the file, see below).
  Caught by reading each reject's actual diagnostic before trusting the
  accept/reject count, not by assuming rc≠0 always means "the language
  rejected this."

## Results

### Nesting

| file | outcome | diagnostic |
|---|---|---|
| `nesting_parens_20deep.dl` (20 nested parens) | accept | none |
| `nesting_arith_deep.dl` (20-term chained arithmetic) | accept | none |

### Precedence (grammar §4: `mul` binds tighter than `arith` add/sub; `*`/`/`/`%` share a level, `+`/`-` share a level; both levels left-associative; `unary` binds tightest)

| file | expression | accept? | computed value | matches grammar-correct reading? |
|---|---|---|---|---|
| `precedence_add_mul.dl` | `2 + 3 * 4` | yes | 14 | yes (`2+(3*4)`, not `(2+3)*4=20`) |
| `precedence_sub_mul.dl` | `10 - 2 * 3` | yes | 4 | yes (`10-(2*3)`, not `(10-2)*3=24`) |
| `precedence_add_div.dl` | `2 + 20 / 4` | yes | 7 | yes (`2+(20/4)`, not `(2+20)/4=5`) |
| `precedence_add_mod.dl` | `1 + 10 % 3` | yes | 2 | inconclusive by itself — both readings give 2 (see `sub_mod`) |
| `precedence_sub_mod.dl` | `5 - 10 % 3` | yes | 4 | yes (`5-(10%3)`, not `(5-10)%3=-2`) — distinguishes the readings |
| `precedence_mul_div.dl` | `20 / 4 * 2` | yes | 10 | yes, left-assoc (`(20/4)*2`, not `20/(4*2)=2`) |
| `precedence_mul_mod.dl` | `7 % 5 * 3` | yes | 6 | yes, left-assoc (`(7%5)*3`, not `7%(5*3)=7`) |
| `precedence_div_mod.dl` | `20 / 4 % 3` | yes | 2 | yes, left-assoc (`(20/4)%3`, not `20/(4%3)=20`) |
| `precedence_left_assoc_add.dl` | `10 - 3 - 2` | yes | 5 | yes, left-assoc (`(10-3)-2`, not `10-(3-2)=9`) |
| `precedence_left_assoc_div.dl` | `100 / 10 / 2` | yes | 5 | yes, left-assoc (`(100/10)/2`, not `100/(10/2)=20`) |
| `precedence_unary_mul.dl` | `-2 * 3` | yes | -6 | degenerate — both groupings give -6 |
| `precedence_unary_sub.dl` | `-2 - 3` | yes | -5 | yes, unary binds only to `2` (`(-2)-3`, not `-(2-3)=1`) |
| `precedence_paren_override.dl` | `(2 + 3) * 4` | yes | 20 | yes, parens override default precedence |
| `precedence_constraint_both_sides.dl` | `Y+1 < Z*2-1` | yes | (constraint, no value) | accepts; relop sits outside arith on both sides per grammar |

**All 14 precedence/associativity files accept, and every numerically-
distinguishing case's computed value matches the grammar-correct reading.**
No case where Soufflé's actual evaluated value contradicted blueprint
grammar §4's stated precedence/associativity table.

### Unary minus

| file | expression | accept? | computed value |
|---|---|---|---|
| `unary_double_minus.dl` | `X = - -Y` (Y=5) | yes | 5 |
| `unary_sub_minus.dl` | `Z = X - -Y` | yes | (accepts; not numerically spot-checked, would need a 2-column `q` row without arity-conflicting the shared facts dir — see caveat above) |
| `unary_paren_double_neg.dl` | `X = -(-(Y))` (Y=5) | yes | 5 |

### Comments

| file | outcome | diagnostic |
|---|---|---|
| `comment_inside_body.dl` | accept | none |
| `comment_between_neck_and_body.dl` | accept | none |
| `comment_contains_period.dl` (`.` inside `//` comment) | accept | none |
| `comment_block_with_period.dl` (`.` inside `/* */` comment, mid-body) | accept | none |
| `comment_unterminated_block.dl` | **accept, but see below** | `Warning: No rules/facts defined for relation p` |

**`comment_unterminated_block.dl` is the interesting case.** The
unterminated `/* ...` does not produce a syntax error — it silently
consumes everything to end-of-file, including the `p(X) :- q(X).` rule
that followed it. The program still compiles (there is nothing left to
reject), but `p` ends up with zero rules, hence the warning. A parser that
reports "accept" here without also surfacing this warning would hide a
real authoring mistake; this is a case worth the rejection-corpus's
attention even though Soufflé itself does not reject it.

### Lexical

| file | outcome | diagnostic |
|---|---|---|
| `lexical_unicode_identifiers.dl` (non-ASCII `données`/`résultat`/`Élève`) | **reject** | `syntax error, unexpected invalid token, expecting ( or ","` + `unexpected <?>` (mis-decoded byte) |
| `lexical_4kb_identifier.dl` (4,091-byte identifier) | reject, but **inconclusive on the lexical question** | `Error loading <name> data: Cannot open fact file` — a *runtime* fact-loading failure, not a parse/lex error; no syntax or semantic error was reported, meaning the identifier itself compiled fine. The failure is that a facts filename that long exceeds the filesystem's own 255-byte limit — an artifact of this harness's `-F`-directory convention, not evidence Soufflé's lexer rejects long identifiers. |
| `lexical_mixed_line_endings.dl` (LF/CRLF/tab mixed) | accept | none |
| `lexical_no_trailing_newline.dl` | accept | none |
| `lexical_empty_file.dl` | accept | none |
| `lexical_only_comments.dl` | accept | none |

**Soufflé's identifiers are ASCII-only** — confirmed, not assumed: the
`données` case fails at the lexer (`unexpected invalid token`), and the
4,091-byte all-ASCII identifier case fails only downstream, at fact
loading, with no lex/parse complaint at all.

### Malformed

| file | outcome | diagnostic |
|---|---|---|
| `malformed_headless_rule.dl` (`:- q(X), X < 0.`) | reject | `syntax error, unexpected :-` |
| `malformed_empty_body.dl` (`p(X) :- .`) | reject | `syntax error, unexpected .` |
| `malformed_missing_period.dl` | reject | `syntax error, unexpected end of file, expecting .` |
| `malformed_unbalanced_parens.dl` (`q(X.`) | reject | `syntax error, unexpected ., expecting )` |
| `malformed_decl_no_arity.dl` (`.decl p()`) | **accept** | `Warning: No rules/facts defined for relation p` |
| `malformed_duplicate_decl.dl` | reject | `Redefinition of relation p` ... `Previous definition in file ... at line 3` |

**`malformed_decl_no_arity.dl` is a genuine surprise relative to this
project's own restricted grammar.** Blueprint §4's `decl` production
(`param (',' param)*`) requires at least one parameter — under this
project's own grammar, arity-0 relations are out of scope. **Soufflé itself
accepts them** (a real, if unusual, feature of full Soufflé, not a bug in
the test). This is not a contradiction to escalate — the blueprint's
grammar is a deliberate strict subset (§4: "Deliberately a strict subset of
Soufflé's syntax") — but it means a rejection-test case built on "arity-0
`.decl`" would need to say clearly *which* grammar it is testing
conformance to (`dlc`'s restricted one, which should reject this) since
the oracle itself (full Soufflé) does not.

### Semantic edges

| file | outcome | diagnostic |
|---|---|---|
| `semantic_wildcard_in_head.dl` (`p(X, _) :- ...`) | reject | `Underscore in head of rule` |
| `semantic_repeated_var_in_atom.dl` (`p(X,X) :- q(X,X).`) | accept | none |
| `semantic_head_also_input.dl` (relation both `.input` and rule-derived) | accept | none (after the `p.facts` harness fix above) |

`_` is legal as a **term** generally (blueprint grammar §4 does not
restrict it) but Soufflé specifically forbids it in head position — a
sharper rule than the blueprint's grammar states, worth a rejection-test
case of its own distinct from allowedness (this is a syntactic ban, not a
range-restriction failure). Repeated variables within one atom, and mixing
`.input` with a derivation rule for the same relation, are both accepted
without complaint.

## What a skeptic attacks first

- `lexical_4kb_identifier.dl`'s verdict is explicitly flagged inconclusive
  above — a skeptic should not cite "4KB identifiers are rejected" from
  this corpus; that claim is not actually established here.
- `precedence_add_mod.dl` and `precedence_unary_mul.dl` are included for
  completeness but are individually uninformative (both groupings coincide
  numerically) — the informative pairs are `sub_mod` and `unary_sub`
  respectively; a skeptic reading only the file list without the "matches?"
  column could be misled into thinking more pairs were distinguished than
  actually were.
- `comment_unterminated_block.dl`'s "accept" is technically correct but
  semantically empty (zero rules for `p`) — a naive accept/reject count
  would file this under "comments are fine," which undersells the actual
  finding (silent, potentially confusing swallowing of the rest of the
  file).

## Provenance

`measurements/night02-t2-hostile-summary.json` (39 rows), per-file
`measurements/night02-t2-hostile-<stem>{,-run}/` directories (`cmd.txt`,
`stdout.txt`, `stderr.txt`, `env.txt`, `meta.json`). Corpus:
`tests/hostile/*.dl` (39 files). Runner: `harness/night02_t2_hostile.py`.
Completed inside the 90-minute cap.
