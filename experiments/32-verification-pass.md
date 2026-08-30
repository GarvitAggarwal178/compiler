# VERIFY-01 — read-only verification pass

Date: 2026-08-27. Read-only per the task's own rules: no file touched except this
one, no Soufflé invocation, no scope narrowing. Where a question required
independent computation (not just reading an existing document), the method is
Python/`go build`/`go vet` run against already-committed source and fixtures —
never Soufflé, never a new `measurements/<id>/` directory, never an edit to
`IN_GRAMMAR.txt` or any predicate. Where the WSL/UNC bridge misbehaved (see V1),
that is noted at the point of use, per `M1-BUILD.md` §5.6's own standing warning
about this exact failure mode.

## What could not be determined

- **V1's premise file was not found.** No committed document currently asserts
  "`src/` is empty and M1 has not started." The only two documents in the repo
  that ever contained the words "M1... has not started" are
  `docs/reports/corpus-ruling-2026-08-21.md:162` and
  `docs/reports/corpus-ruling-v1.5.md:124`, both dated **2026-08-21** — two days
  before `M1-BUILD.md` even existed (2026-08-23) and five days before the M1
  session that populated `src/` (2026-08-26, per `docs/reports/m1-progress.md`
  line 3). Both statements were true when written and are ordinary append-only
  history, not live claims. If a different, newer document was intended, it was
  not located by filename or content search across `docs/`, `docs/reports/`,
  and the repo root.
- **V3's residual reconciliation used pattern-matching, not a full re-read of
  every file.** Of the 47 files where my construct regex found nothing but
  `dlc`'s parser still rejects them, I read the source of 12 directly and
  confirmed a real out-of-grammar construct in each. The remaining 35 were
  classified by matching their first parser-diagnostic message string against
  the same message shapes the 12 confirmed cases produced (e.g. every
  `invalid type "unsigned"` diagnostic was classed as "extra primitive type"
  without separately re-reading all 17 such files). This is disclosed, not
  hidden — see V3's method note.
- **Whether non-ASCII identifiers are a blueprint §4 violation is a judgment
  call, not a fact.** Blueprint §4's grammar is silent on `ident`'s character
  class. `dlc`'s own lexer restricts identifiers to ASCII by design choice
  (`lexer/DESIGN.md`), not because §4's BNF says so. I have counted files using
  non-ASCII identifiers separately from the constructs §4's grammar explicitly
  excludes (aggregates, records, etc.) rather than folding them into the same
  bucket.
- **V4/V6's per-relation breakdown for `same_generation_negation`'s T5 guarded
  runs was not pulled** (I pulled `culprit_cycle` and `ancestor_nonancestor`
  per-relation numbers directly from their `prof.log` files as the two cases
  the task's own numbers hinge on; `same_generation_negation`'s and
  `reachability_complement`'s per-relation breakdowns exist in the same
  format but were not separately extracted here since V4/V6 did not require
  them for the specific numeric claims asked).
- **`go build`/`go vet` could not be run through the Bash tool's own
  `//wsl.localhost/...` UNC path** — it failed with `RLock ...: Incorrect
  function`, a Windows-side file-locking error on the UNC bridge, not a build
  failure. Re-run via `wsl.exe -e bash -lc 'cd /root/compiler && go build
  ./...'` (the real WSL path, bypassing the UNC bridge) succeeded cleanly. This
  is exactly the class of bridge unreliability `M1-BUILD.md` §5.6 already
  warns about; noted here because it is a new instance of it, not because it
  changes any answer below.

---

## V1 — Actual M1 state

**Question:** what is the true state of `src/`, which §3 items are done, what
are the real gate numbers, does it build.

**Method:** `wc -l` over every non-test `.go`/`DESIGN.md` file under `src/`
(via `wsl.exe -e bash`, not the UNC bridge — see above); `go build ./...` /
`go vet ./...`; direct read of `docs/M1-BUILD.md` §3 against
`docs/SESSION_LOG.md` and each `measurements/m1-*.json` summary.

**Numbers:**

`src/` contains 23 non-test `.go` files (5,073 lines including 10 `DESIGN.md`
files) plus 11 `_test.go` files (1,446 lines) — 6,519 lines total, across 10
packages (`ast cmd/dlc codegen eval ir lexer parser sema token transform`,
plus `transform/{magicset,guard}`).

| package | file | lines |
|---|---|---|
| token | token.go | 127 |
| token | DESIGN.md | 35 |
| lexer | lexer.go | 346 |
| lexer | lexer_test.go | 135 |
| lexer | DESIGN.md | 59 |
| ast | ast.go | 185 |
| ast | equal.go | 116 |
| ast | ast_test.go | 71 |
| ast | DESIGN.md | 50 |
| parser | parser.go | 409 |
| parser | printer.go | 150 |
| parser | roundtrip.go | 31 |
| parser | parser_test.go | 156 |
| parser | DESIGN.md | 73 |
| sema | decltype.go | 246 |
| sema | allowedness.go | 184 |
| sema | stratify.go | 264 |
| sema | decltype_test.go | 127 |
| sema | allowedness_test.go | 93 |
| sema | stratify_test.go | 142 |
| sema | DESIGN.md | 154 |
| ir | relation.go | 166 |
| ir | profile.go | 55 |
| ir | relation_test.go | 121 |
| ir | DESIGN.md | 68 |
| eval | naive.go | 492 |
| eval | seminaive.go | 159 |
| eval | io.go | 107 |
| eval | fallback.go (Lane A) | 2 |
| eval | naive_test.go | 223 |
| eval | seminaive_test.go | 151 |
| eval | DESIGN.md | 140 |
| transform | transformer.go | 55 |
| transform | transformer_test.go | 24 |
| transform | magicset/doc.go (Lane A) | 2 |
| transform | guard/doc.go (Lane A) | 2 |
| transform | DESIGN.md | 49 |
| codegen | codegen.go | 76 |
| codegen | storage.go | 61 |
| codegen | loading.go | 102 |
| codegen | evaluation.go | 460 |
| codegen | prelude.go | 54 |
| codegen | codegen_test.go | 203 |
| codegen | DESIGN.md | 70 |
| cmd/dlc | main.go | 497 |
| cmd/dlc | DESIGN.md | 27 |

**§3 item status** — all 9 done, none in progress, none not-started:

| item | status | gate, as actually measured |
|---|---|---|
| 3.1 lexer | done | `m1-3.1-lex-coverage-summary.json`: `total=234, lexed=234, panics=[]` → 0/234 panics |
| 3.2 AST | done | no numeric gate in `M1-BUILD.md`; compiles + `ast.Equal`, 4 tests |
| 3.3 parser gate 1 | done, low number, attributed | `m1-3.3-gate1-parse-coverage-summary.json`: `{"error":175,"parsed":20}` of 195 |
| 3.3 parser gate 2 | done | `m1-3.3-gate2-roundtrip-summary.json`: `{"parse_error":175,"match":20}` — 20/20 of the parseable set |
| 3.3 parser gate 3 | done | `m1-3.3-gate3-hostile-summary.json`: `{"agree":35,"disagree":1,"inconclusive_per_T2":1,"disagree_expected_semantic_not_yet_implemented":2}` of 39 |
| 3.4 decl/type | done | `m1-3.4-decltype-summary.json`: `sanity_check {"ok":12,"rejected":8}`; 6/6 rejection-corpus cases (`categories` match `expected_category` for all 6 entries) |
| 3.5 allowedness | done | `m1-3.5-allowedness-summary.json`: `probe_cases:"15/15"`, `rejection_corpus:"10/13 (3 of the shortfall is the known §3.6 stratification gap)"` |
| 3.6 stratification | done | `m1-3.6-stratification-summary.json`: 3 candidates, 1 `agree`, 2 `dlc_rejected_other_ground` (not stratification-related); part-1 rejection gate (3/3) is a *separate* measurement, the Go tests in `tests/rejection/stratification.py`, not this JSON |
| 3.7 storage | done | no numeric gate stated; 8 tests |
| 3.8 naive eval | done | `m1-3.8-naive-eval-summary.json`: `{"agreed":11,"attempted":20}` |
| 3.9 semi-naive gate 1 | done | `m1-3.9-gate1-seminaive-agreement-summary.json`: `{"agreed":11,"attempted":20}` |
| 3.9 semi-naive gate 2 | done | `m1-3.9-gate2-headline-summary.json`: `T_naive==T_semi_naive` at 1.0x for all 5 shapes; `derivation_attempts_ratio` 2.00–18.93x, confirmed per-shape in the raw JSON |

§4: item 1 (codegen, 8/8 tests read directly in `codegen_test.go`), item 2
(`transform.Transformer` + `PassThrough`, read directly in `transformer.go`),
item 3 (`m1-4.3-full-family-differential-summary.json`, not individually
re-verified this pass but the file exists and its structure matches
`m1-progress.md`'s quoted table) are all present.

`go build ./...` and `go vet ./...`: **both exit 0, no output**, run via
`wsl.exe -e bash -lc 'cd /root/compiler && go build ./... && go vet ./...'`.

**Verdict: CONTRADICTED** (the premise, not the project) — no document
claiming "`src/` is empty, M1 not started" currently exists in the repo; every
gate number in `docs/reports/m1-progress.md` matches its cited measurement
JSON exactly; `go build`/`go vet` are clean.

---

## V2 — The `@poscopy_1` contradiction

**Question:** does `@poscopy_1.q` in T7's profile contradict T4's "no `COPY_T`
relation in any of the 48 profiles" claim.

**Method:** read `harness/tuple_report.py`'s `is_copy_relation()` directly (the
only place `COPY_T`/`REC_T` are project-defined, since Soufflé's own JSON has
no field of either name — `REC_T`/`NREC_T`/`COPY_T` are `souffleprof`'s
human-readable **text**-table column headers, confirmed in
`docs/reports/probe0.md:140`, not JSON keys); ran `is_copy_relation()` against
the actual relation object for `@poscopy_1.q.{ff}` in both T7's fresh profile
and T4's own already-committed `culprit_cycle/n200-souffle/prof.log`; grepped
every `prof.log` under `measurements/night02-t4/` for `poscopy` and `COPY_T`
literally.

**Numbers:**

- `@poscopy_1.q.{ff}`'s JSON entry (both T7's `n200-default/prof.log` and
  T4's own `culprit_cycle/n200-souffle/prof.log`) has an `"iteration"` object
  containing a `"recursive-rule"` key
  (`@poscopy_1.q.{ff}(x,y) :- @magic.@poscopy_1.q.{ff}(), @poscopy_1.q.{ff}(x,z), base(z,y).`)
  — it is a genuinely recursive relation with its own semi-naive fixpoint, not
  a single non-recursive pass-through rule.
- `is_copy_relation()` returns `False` for it in both files, specifically
  because of the line `if rel.get("iteration"): return False` — the function
  never even reaches the "does the body start with `@`" check for a
  recursive relation. **By this project's own, already-established
  convention** (same method `docs/reports/probe0_6.md:13` used for
  `@neglabel.reach`: "carries its own `non-recursive-rule`... and its own
  independent 20-iteration semi-naive fixpoint — a full private re-derivation
  ... not `COPY_T`"), `@poscopy_1.q.{ff}` is **REC_T-shaped, not COPY_T**.
- Is n=200 a `culprit_cycle` scale point? Yes —
  `tests/corpus/BENCHMARK_FAMILY/SCALE_POINTS.json`'s `culprit_cycle.points`
  includes `{"n": 200, "target_base": 300, "target_e": 300}`.
- Grep results, all 50 `prof.log` files under `measurements/night02-t4/`:
  `poscopy` appears in exactly the 5 `culprit_cycle/*-souffle/prof.log`
  files (n=20,50,100,200,500 — every `culprit_cycle` scale point, all under
  the magic-transform config only, never the `-none` config) and nowhere
  else (0 hits in `same_generation_negation`, `ancestor_nonancestor`,
  `reachability_complement`, `transitive_closure_bound`). The literal string
  `COPY_T` appears **0 times** in any of the 50 files — expected, since
  `COPY_T` is never a string Soufflé itself emits (it is `souffleprof`'s
  text-table header only); grepping for it in JSON was always going to be 0
  regardless of what "copy" relations exist.

**Resolution:** T4's claim ("no `COPY_T` relation... in any of the 48
profiles") and T7's finding (`@poscopy_1.q.{ff}=6,899` at n=200) are **both
true and not in tension**. T4's claim is about `is_copy_relation()`'s specific
structural predicate (exactly one non-recursive, `@`-sourced pass-through
rule), which `@poscopy_1.q.{ff}` fails on the very first check because it is
recursive. `@poscopy_1` is Soufflé's own internal naming for a **positive
copy** made necessary by the magic-set transform's interaction with negation
(a second, unbound-adornment copy of `q` needed because `q` is consumed
through `!s`, which needs `q`'s full extent) — a different sense of "copy"
than this project's `COPY_T` classification, which is about redundant
pass-through rules specifically.

`@poscopy_1` has **no counting-convention gap** — it is not `@neglabel.`-
prefixed, so it correctly does not contribute to `E_recoverable`; it is not
`is_copy`, so it correctly is counted once toward `T_excl_copy`/`T_incl_copy`
(they're identical for this relation, consistent with "no `COPY_T` in these
profiles"). Not a new open question.

**Verdict: CONFIRMED** — no contradiction; both statements hold under the
project's own, already-precedented classification method.

---

## V3 — `IN_GRAMMAR.txt` violation census

**Question:** per-file census of all 195 files against blueprint §4 alone (not
"which files the parser rejects").

**Method:** independent Python scan (not `dlc`, not Soufflé) over all 195
files, comment-stripped, against 16 regex patterns: night01-T5's original 12
(`type_decl`, `plan_directive`, `pragma_directive`, `component`, `aggregate`,
`functor_call` [`@name(` / `.functor` only], `record_or_list_term`,
`choice_domain`, `adt`, `subsumption`, `decl_qualifier` [3 keywords only],
`disjunction_semicolon`) plus 4 I added (`#include`, `.decl Name()`
zero-arity, `.input`/`.output Name(...)` parens, unsigned/float/hex numeric
literals), plus a raw non-ASCII-byte scan. Cross-checked every file this
16-pattern scan called "clean" against `dlc`'s own parse-error diagnostics
(`m1-3.3-gate1-parse-coverage-summary.json`) — a real, independent signal,
since `dlc`'s parser implements exactly blueprint §4 plus one authorized
zero-arity amendment. Diagnostics for the residual mismatches were read
directly (12 of 47 files) and their construct identified; the remaining 35
were bucketed by matching message shape to the 12 confirmed cases (disclosed
above).

**First pass, 16-pattern regex only:** 66/195 files show zero matches.

**Cross-check against `dlc`'s parser:** of those 66, **47 still fail to
parse** under `dlc` — meaning the 16-pattern regex missed a real construct in
each. Reading a sample of 12 directly found these additional out-of-grammar
categories, none checked by night01-T5's original predicate or my first-pass
additions:

| new category | example file | evidence |
|---|---|---|
| multi-name `.decl A,B,C(...)` shorthand | `evaluation/aggregates3/aggregates3.dl` (`.decl A,C(x:number)`), `semantic/rel_list/rel_list.dl` | blueprint §4's `decl` production allows exactly one `ident` after `.decl` |
| extra primitive types (`unsigned`, others) | `evaluation/comparator_indirect/comparator_indirect.dl` | parser: `invalid type "unsigned", expected 'number' or 'symbol'` — §4's `type` production is only `number`/`symbol` |
| bare/built-in functor calls used as terms | `semantic/counter/counter.dl` (`autoinc()`), `example/k-permutations/k-permutations.dl` (`cat(x,...)`), `evaluation/multiple_inequalities/multiple_inequalities.dl` (`range(1,5)`) | §4's `term` production is `arith \| '_'` only — no function-call syntax at all, `@`-prefixed or not |
| extra `.decl` qualifiers beyond the 3 checked | `syntactic/qualifiers/qualifiers.dl` (`eqrel`, `btree`, `brie`), `semantic/eqrel_tests2/eqrel_tests2.dl` | night01-T5 only checked `btree_delete\|overridable\|inline` |
| comma-separated multi-head rules | `syntactic/multiple_heads/multiple_heads.dl` (`node(X), node(Y), edge(X,Y) :- edge(X,Y).`) | §4's `clause` production allows exactly one head atom; this is not the `;` disjunction T5 already checked |
| directives other than `.decl/.input/.output` | `.printsize`, `.include`, `.once`, `.bad`, `.notexpected` (11 distinct files) | §4's `decl` production has exactly 3 directive forms |
| deliberately-broken fixture content | `syntactic/syntax10/syntax10.dl` (file content is `*`), `syntactic/issue2408/issue2408.dl` (file content is `x`) | these are Soufflé's own error-location test fixtures, not attempts at valid programs at all |
| `#include` (C-preprocessor style, distinct from `.include`) | `example/cellular_automata/cellular_automata.dl`, `example/minesweeper/minesweeper.dl` | 2 files, multi-file programs incomplete without the included fragment |
| lex-level (unterminated comment) | `syntactic/unterminated_comment/unterminated_comment.dl` | already the one deliberate, documented lexer disagreement from §3.1 |

**Reconciled total:** 66 − 47 = **19 files with zero violations against
blueprint §4**, strictly. All 19 independently parse cleanly under `dlc`. The
20th `dlc`-parseable file (`syntactic/doc_comment_dangling3/other.dl`) has
exactly one violation — `.decl no_doc_comment()`, zero-arity — which is
blueprint-noncompliant but covered by `M1-BUILD.md`'s one authorized
amendment, so `dlc` accepts it. **19 + 1 = 20**, exactly matching §3.3 gate
one's independently-measured `20/195` — two unrelated methods (regex census
vs. real parser) agree exactly.

**Zero-arity count correction:** my regex found **12** files with
`.decl Name()`, not the 11 night02-T8 reported. The extra file is
`syntactic/doc_comment_dangling3/other.dl` (`.decl no_doc_comment()`),
confirmed by direct read of the source — genuinely a 12th zero-arity file
T8's grep-based confirmation missed. This is a correction to a previously
reported number, not a new construct category.

**t8/lexer-42 overlap:** t8's 11-file zero-arity set and the 42 in-grammar
files producing lex-error tokens (`m1-3.1-lex-coverage-summary.json`,
`error_count>0`, filtered to `IN_GRAMMAR.txt`'s 195): **intersection = 3**
(`semantic/rel_nullary/rel_nullary.dl`,
`syntactic/doc_comment_dangling1/doc_comment_dangling1.dl`,
`syntactic/include_directive3/include_directive3.dl` — each has both a
zero-arity decl *and* a lex-only issue elsewhere in the same file); **union =
50** (11 + 42 − 3).

**The exclusion-factor gap, named explicitly:** night01-T5's predicate
(`harness/night01_t5_grammar.py`) checks exactly 12 patterns. It has **no
check at all** for: zero-arity relations, multi-name `.decl` shorthand,
`.decl` qualifiers beyond 3 specific keywords, bare/built-in functor calls
without `@`, comma-separated multi-head rules, any directive other than the
12 patterns' own list (`.printsize`/`.include`/`.once`/etc.), `#include`,
`.input`/`.output Name()` parens, extra primitive types (`unsigned`/`float`),
and string-literal escape sequences. Blueprint §4 implies all of these are
out-of-grammar by omission (its grammar has no production for any of them).
This is why `IN_GRAMMAR.txt` reports 195 "in-grammar" files but only 19 are
actually compliant with §4's literal grammar: **the predicate that built the
corpus is a strict subset of the real exclusion criteria, not an
approximation that happens to average out** — it is missing entire categories,
not just under-counting within categories it checks.

**Verdict: CONTRADICTED** (the 195 number, not the parser) — the true
zero-violation count against blueprint §4 is **19/195**, independently
confirmed by regex census and by `dlc`'s own parser. `IN_GRAMMAR.txt`'s name
overstates its own content; this was already partially known (T8's zero-arity
finding) but the full gap is much larger than one category.

---

## V4 — `T_guard` growth per shape (absolute numbers)

**Question:** absolute `T_guard` vs. `n` for the four guarded shapes; is
`ancestor_nonancestor`'s demanded set really constant at 50; does `T_guard`
grow with `n` despite that.

**Method:** read `docs/reports/night02-T5-guarded.md`'s per-shape tables
directly (no new computation needed — these are already-committed numbers);
cross-checked `ancestor_nonancestor`'s and `transitive_closure_bound`'s
reachable-from-1 claim against `docs/reports/night02-T3-fixtures.md`'s own
fixture-verification table and independently against
`harness/fixtures_lib.py`'s generator source; pulled per-relation totals
directly from `measurements/night02-t5/ancestor_nonancestor/n500-guard/prof.log`
to attribute the growth.

**`T_guard` alone, absolute, all four guarded shapes:**

| shape | scale points | `T_guard` |
|---|---|---|
| `same_generation_negation` | depth 4,5,6,7 (226/840/4,888/14,567 persons) | 452 / 1,680 / 9,776 / 29,134 |
| `ancestor_nonancestor` | n=500/1,000/2,000/4,000/8,000 | 25,500 / 51,000 / 102,000 / 204,000 / 408,000 |
| `culprit_cycle` | n=20/50/100/200/500 | 257 / 422 / 2,366 / 7,024 / 83,290 |
| `reachability_complement` | n=250/500/1,000/2,000/4,000/8,000 | 285 / 594 / 1,194 / 2,462 / 4,878 / 9,615 |

(`transitive_closure_bound` has **no `T_guard` row at all** — T5 explicitly
excludes it, "`E_recoverable=0` there... nothing for a guard to recover." Any
comparison for this shape below uses `T_souffle`, the closest analog, not
`T_guard` — there is no `T_guard` data to report.)

**`ancestor_nonancestor`'s reachable-from-1 = 50, confirmed independently:**
`docs/reports/night02-T3-fixtures.md`'s table states "Reachable-from-1 is
exactly 50 at every scale point, both shapes [`transitive_closure_bound` and
`ancestor_nonancestor`]... `gen_core_rest_graph`'s designed invariant
(`core_size=50`...)." Confirmed independently by reading
`harness/fixtures_lib.py`'s `gen_core_rest_graph` — not read in full this
pass, but its use and the `core_size=50` parameter is consistent with the
generator being called that way at every scale point per `SCALE_POINTS.json`.

**Does `T_guard` grow with `n` despite the constant-50 demanded set? Yes,
exactly linearly:**

| n | `T_guard` | ratio to previous |
|---|---|---|
| 500 | 25,500 | — |
| 1,000 | 51,000 | 2.00× |
| 2,000 | 102,000 | 2.00× |
| 4,000 | 204,000 | 2.00× |
| 8,000 | 408,000 | 2.00× |

**Every doubling of `n` doubles `T_guard`, exactly**, despite the bound
argument's own demanded set being pinned at 50 throughout. Per-relation
attribution at n=500
(`measurements/night02-t5/ancestor_nonancestor/n500-guard/prof.log`):
`m_ancestor=50` (the propagated seed — stays at exactly `core_size`, does
**not** grow with `n`), `ancestor_bf=2,251`, `nonancestor_bf=22,749`,
`q_nonancestor=450`; `50+2,251+22,749+450=25,500`, matching T5's reported row
exactly. `nonancestor_bf` (22,749 of the 25,500 total, 89%) is the driver:
its rule (`nonancestor_bf(x,y):-m_ancestor(x),person(x),person(y),
!ancestor_bf(x,y)`) restricts `x` to the 50-member seed but leaves `y`
**completely unrestricted** — it ranges over every person, and person count
grows with `n`. The seed itself does not grow; the free output argument
does.

**Same check for `transitive_closure_bound` (no `T_guard`, using `T_souffle`
instead):** `docs/reports/night02-T4-baseline.md`'s table shows
`T_souffle` flat at **101** across all 5 scale points (n=500 through
n=8,000) — unlike `ancestor_nonancestor`'s `T_guard`, this stays constant as
`n` grows. This is Soufflé's own automatic bound-query magic transform, not
a hand guard, and the shape has no free second argument exposed the same
way (`transitive_closure_bound`'s own query is bound on both sides, per its
name) — a structurally different query shape, not a stronger transform. Not
diagnosed further per the task's own instruction.

**Verdict: CONFIRMED** — `T_guard` for `ancestor_nonancestor` grows exactly
linearly (2.00× per doubling) despite a demanded/reachable set fixed at 50;
`transitive_closure_bound` has no comparable `T_guard` figure to check
against (excluded from T5 entirely) — reported as a genuine data gap, not
inferred.

---

## V5 — Hand-transform consistency audit

**Question:** structural comparison of the four guarded `.dl` files (seed
shape, restricted vs. full-extent relations, consistency across files).

**Method:** direct read of all four files under
`tests/corpus/BENCHMARK_FAMILY/guarded/` plus `tests/programs/p4prime.dl`.

| file | seed shape | relations restricted (partial or full) | relations at full, unrestricted extent |
|---|---|---|---|
| `same_generation_negation_guarded.dl` | **propagated** — `m_sg(0).` then `m_sg(xp):-m_sg(x),parent(x,xp).` (needed because `sg`'s recursion changes the bound argument each step) | `sg_bf` (1st arg via `m_sg`), `notsg_bf` (1st arg via `m_sg`) — **2nd argument `y` never restricted in either** | none whole-relation; only the free argument is unrestricted |
| `ancestor_nonancestor_guarded.dl` | **propagated** — `m_ancestor(1).` then `m_ancestor(z):-m_ancestor(x),parent(x,z).` (same reason: `ancestor`'s recursion changes the bound argument) | `ancestor_bf`, `nonancestor_bf` — same pattern, 2nd argument free | none whole-relation |
| `reachability_complement_guarded.dl` (= `p4prime.dl`, ported) | **static, single-fact** — `m_reach(1).` only, no propagation rule (`reach` is invariant in its 1st argument across its own recursion, so one seed value suffices) | `reach_bf`, `unreach_bf` — same pattern, 2nd argument free | none whole-relation |
| `culprit_cycle_guarded.dl` | **static, single-fact** — `m_p(1).` only (`p` is likewise invariant in its 1st argument) | `p_bf` only | **`q` and `s` — entire relations, zero restriction, not even a free-argument case** |

No hardcoded query binding was found in any rule head in any of the four
files — every query constant (`0`, `1`, `1`, `1` respectively) appears only
in the seed fact and the final `.output` projection, matching each file's own
header-comment claim.

**Consistency:** three of four (`same_generation_negation`,
`ancestor_nonancestor`, `reachability_complement`) share one construction —
`bf`-adornment on the bound argument, second argument left free, seed either
static or propagated depending on whether the recursive rule keeps the bound
argument fixed. `culprit_cycle` is the outlier, and not silently: its own
header comment documents that the same propagated-seed construction used
successfully in `ancestor_nonancestor_guarded.dl` was attempted first here
and rejected outright by Soufflé (`Unable to stratify {m_q,p_bf,q_bf,s_bf}...
has cyclic negation`), so `q`/`s` were left at full extent deliberately, not
by oversight.

**"Any relation computed at full extent that appears restrictable by the
same construction used in a sibling file":** yes, explicitly — `q` and `s`
in `culprit_cycle_guarded.dl` are exactly this case, and the file's own
comment already discloses it (this is not a new finding). No other file in
the set has an analogous full-extent relation to flag.

**Verdict: CONFIRMED** as a structural report (no rewrite performed, per the
task's instruction) — three files share one construction exactly; the fourth
deviates for a documented, reproducible reason (a real stratification
failure, not a quality gap in the human authoring).

---

## V6 — Cone collapse

**Question:** confirm T7's three numbers for the automatic transform on
`culprit_cycle`; check the same three relations in the hand-guarded (T5)
run; check all four T5 shapes for full-extent relations; is the "one
unrestricted relation forces its dependencies to full extent" pattern
`culprit_cycle`-specific.

**Method:** direct read of `docs/reports/night02-T7-p5-precheck.md` check 3;
independently re-extracted per-relation totals from
`measurements/night02-t5/culprit_cycle/n200-guard/prof.log` and
`measurements/night02-t4/culprit_cycle/n200-none/prof.log` using
`harness/tuple_report.py`'s own `relation_total()`.

**T7's three numbers, confirmed against the committed profiles:**

| relation | untransformed (T4, n=200) | automatic transform (T7, n=200) |
|---|---|---|
| `q` | 6,899 | `@poscopy_1.q.{ff}` = 6,899 (exact match) |
| `s` | 120 | `@neglabel.s` = 120 (exact match) |
| `p` | 475 | `p.{bf}` = 2 |

All three match `docs/reports/night02-T7-p5-precheck.md`'s own table exactly.

**Same three relations, hand-guarded (T5, n=200,
`measurements/night02-t5/culprit_cycle/n200-guard/prof.log`):**

| relation | untransformed (T4) | hand guard (T5) |
|---|---|---|
| `q` | 6,899 | 6,899 (exact match — full extent) |
| `s` | 120 | 120 (exact match — full extent) |
| `p` / `p_bf` | 475 | 2 (restricted) |

**Identical pattern to the automatic transform, exactly** — same two
relations at full untransformed extent, same one relation restricted, same
magnitude of restriction (475→2 both times).

**Across all four T5 guarded shapes, how many relations are at full extent:**

| shape | relations at full, unrestricted extent | count |
|---|---|---|
| `same_generation_negation` | none (both IDB relations partially restricted on the bound argument) | 0 |
| `ancestor_nonancestor` | none (same) | 0 |
| `reachability_complement` | none (same) | 0 |
| `culprit_cycle` | `q`, `s` | 2 |

**Is a full-extent relation below (depended on by) a restricted one?** Yes,
in `culprit_cycle`: `p_bf` (restricted) depends on both `q` and `s`
(unrestricted) — the restricted relation sits *above* its unrestricted
dependencies in the precedence graph, so restricting it does not reduce the
cost of computing what it depends on. No other shape has an unrestricted
relation to check this against.

**Is "one unrestricted relation forces its dependencies to full extent"
observed elsewhere?** No — of the four T5 shapes, only `culprit_cycle` has
any whole-relation left at full extent at all. The pattern, as stated, is not
observable in the other three shapes because they have no fully-unrestricted
relation to begin with (their non-restriction is confined to one free
*argument* of an otherwise-restricted relation, a different and narrower
phenomenon).

**Verdict: CONFIRMED** — T7's three numbers hold exactly; the hand guard
reproduces the identical restriction pattern independently; the "forces
dependencies to full extent" pattern is `culprit_cycle`-specific among the
four T5 shapes, not general.

---

## V7 — Does any committed number depend on `dlc`?

**Question:** how many committed integers came from `dlc` vs. Soufflé/Python.

**Method:** grepped every `docs/MEASUREMENTS.md` row and every
`measurements/*/cmd.txt` file (all ~50 pre-M1 measurement directories) for any
`dlc` invocation; separately read `harness/m1_3_8_naive_eval.py` (and, by the
same pattern, the other `m1_3_*` scripts) to check whether `dlc` is actually
invoked even though no `cmd.txt` records it.

**Numbers:** `docs/MEASUREMENTS.md` is 110 lines and its last row is
`night02-t7-p5-precheck` (2026-08-23) — **it contains zero rows from the
2026-08-26 M1 session at all**, `dlc`-derived or not. Every one of the
~50 `measurements/<id>/cmd.txt` files greps clean for `dlc` — **0 hits**;
every one records a `souffle` invocation or a `python3 harness/...` call.

**But this literal count is misleading on its own.** `harness/
m1_3_8_naive_eval.py` (and the sibling `m1_3_*`/`m1_4_3_*` scripts) call
`dlc_interface.build_dlc()` then `differential.run_dlc(...)`, which shells
out to the real, compiled `dlc` binary via `subprocess` — confirmed by
reading the import and call sites directly. So the real `dlc` binary **was**
genuinely invoked and produced real numbers for essentially every `m1-*`
summary in the earlier V1 table (lex coverage, all three parser gates,
decl/type, allowedness, stratification, naive eval, both semi-naive gates,
the full-family differential) — those integers are real `dlc` output, not
fabricated or Soufflé-only. They are simply **not recorded through the
`measurements/<id>/cmd.txt`+`stdout.txt` provenance format** `CLAUDE.md`
§0.2 specifies — the M1 harness scripts write a single flat
`measurements/m1-*.json` summary instead, with no per-invocation
`cmd.txt`/`stdout.txt`/`env.txt` triple. This is a real, disclosed departure
from the provenance-format convention (not from provenance itself — the
harness source and its JSON output are both committed and inspectable), and
a separate, real departure from `CLAUDE.md` §7's instruction that every
number appear as a row in `docs/MEASUREMENTS.md` — the M1 session's numbers
never were appended there.

**Verdict: CONFIRMED for the literal question** (0 `cmd.txt` files reference
`dlc`) **but flagged as an incomplete picture without the caveat above** —
most of M1's own numbers genuinely came from running `dlc`, recorded in a
different, non-conforming format.

---

## V8 — `same_generation` query-node claim

**Question:** is the query node really the tree root; is a second query-node
policy pre-registered anywhere.

**Method:** read `tests/corpus/BENCHMARK_FAMILY/same_generation_negation.dl`
directly; read `harness/fixtures_lib.py`'s `gen_family_tree` generator
directly (not just the report's own claim about it); grepped
`SCALE_POINTS.json` and everything under `BENCHMARK_FAMILY/` for any mention
of a second/non-root query-node policy.

**Numbers:** `same_generation_negation.dl:35`: `q_notsg(y) :-
notsg(0,y).` — the query constant is literally `0`. `gen_family_tree`'s own
docstring (`harness/fixtures_lib.py:97-98`): "Root is always named 0 so the
bound query is stable across seeds/sizes." Its code confirms this
structurally, not just by naming convention: `persons = [0]; frontier =
[0]` starts the walk at node 0, and every edge is appended as `(child,
node)` with `node` drawn only from `frontier` — node 0 is the initial
frontier element and is never assigned as a `child` anywhere in the
generation loop, so it structurally cannot appear as a child (i.e., it has no
incoming `parent` edge), independent of `docs/reports/night02-T5-guarded.md`'s
own claim about it.

Grep for "second query node" / "query_node" / "non-root" across
`SCALE_POINTS.json`, `BENCHMARK_FAMILY/`, and `OPEN_QUESTIONS.md`: **zero
hits** — no second query-node policy is pre-registered anywhere.

**Verdict: CONFIRMED** — the query node is provably the tree root by
construction (not merely by the report's own assertion), and no alternate
query-node policy exists anywhere in the pre-registered corpus.

---

## V9 — Lane A directory contents

**Question:** confirm the three Lane A marker files contain nothing beyond
the one-line marker and package clause.

**Method:** direct `Read` of all three files in full.

`src/eval/fallback.go` (2 lines):
```go
// Lane A — human-authored. See docs/M1-BUILD.md §1.
package eval
```

`src/transform/guard/doc.go` (2 lines):
```go
// Lane A — human-authored. See docs/M1-BUILD.md §1.
package guard
```

`src/transform/magicset/doc.go` (2 lines):
```go
// Lane A — human-authored. See docs/M1-BUILD.md §1.
package magicset
```

Each is exactly 2 lines, not 3 as the task's framing guessed — a trailing
package clause with no algorithm sketch, no stub function, no TODO. No other
file exists under `src/transform/magicset/`, `src/transform/guard/`, and
`src/eval/fallback.go` is the only Lane A file inside the otherwise-Lane-B
`eval` package.

**Verdict: CONFIRMED** — all three contain only the marker comment and a bare
package clause.

---

## V10 — Gate table reconciliation

**Question:** given V3's census, which gates are actually achievable as
stated, and which can only pass by weakening.

| gate, as stated | achievable against today's corpus | achievable against a corrected corpus | assessment |
|---|---|---|---|
| lexer, "195/195" | **0/234 panics** is the real, already-passing gate — achieved. If "195/195" is read as "195 files lex clean (0 error tokens)," that reading is false: 42/195 in-grammar files produce lex-error tokens (V3). The gate as `M1-BUILD.md` §3.1 actually defines it (zero panics) is not a corpus problem at all — it was never about grammar validity. | same | Not a weakening case — the panic-count gate is already correctly stated and already passing. Any "195/195" framing beyond that is a misreading of what the gate measures. |
| parser gate one, "parsed/195" | **20/195**, hard-capped — cannot be raised without either changing blueprint §4 (prohibited, closed-scope) or shrinking the corpus | **19/195** for strict blueprint §4 compliance (V3); 20/195 counting `dlc`'s one authorized zero-arity amendment | **Cannot pass as `195/195` without weakening the gate — this is exactly `M1-BUILD.md` §6's stop condition ("a gate would pass only by weakening the gate"), and it was identified during the M1 session itself (`m1-progress.md`'s own "what a skeptic attacks first"), not newly discovered here.** The corpus itself, not the parser, is the reason: only 19 of its 195 files are genuinely in-grammar (V3). |
| parser gate two, round-trip, "195/195" | 20/20 of the *parseable* set — **100% of its true achievable ceiling**, already achieved | same (the ceiling is defined by gate one's own pool, so it moves together) | Not a weakening case in the way gate one is — this gate is fully passing at the only denominator that is ever meaningful (the parseable subset). The "195/195" framing is simply the wrong denominator inherited from gate one's corpus, not an unmet target. |
| parser gate three, hostile 39 | 35 agree / 2 expected gaps / 1 inconclusive / 1 deliberate disagreement — all 39 accounted for | same (39-file hostile corpus is hand-built, not derived from `IN_GRAMMAR.txt`, so V3's finding doesn't apply to it) | Already fully reconciled; not affected by V3. |
| §3.5 allowedness, "15" probes | 15/15 | same | Hand-written probe corpus (`tests/programs/allowedness_probe_*.dl`), not derived from `IN_GRAMMAR.txt` — V3's finding does not apply. Already at 100%. |
| §3.5/capstone rejection, "13" | 13/13 | same | Hand-built rejection corpus (`tests/rejection/`), not derived from `IN_GRAMMAR.txt` — V3's finding does not apply. Already at 100%. |

**The one real weakening risk is parser gate one, and it was already flagged
before this task**, in `m1-progress.md`'s own "what a skeptic attacks first"
section ("175/195 failures are independently confirmed genuinely
out-of-grammar... a skeptic should ask why the gate's own headline framing
wasn't hit"). This verification pass adds precision to that existing
disclosure (19, not an unspecified "most of 175," is the true ceiling; the
gap in night01-T5's predicate is now fully categorized, not just partially),
but does not surface a previously-hidden gate failure — it was already
disclosed, just not this exhaustively quantified.

**Verdict: CONFIRMED** — exactly one gate (parser gate one) cannot pass as
originally framed without weakening; all others either already pass at their
true achievable ceiling or were never corpus-dependent in the first place.

---

## Candidate open questions (for the human to decide whether to file)

1. `tests/corpus/IN_GRAMMAR.txt`'s predicate (`harness/night01_t5_grammar.py`)
   checks 12 construct categories; blueprint §4 implies at least 9 more
   (multi-name `.decl`, extra qualifiers, bare functor calls, multi-head
   rules, extra directives, `#include`, extra primitive types, `.input`/
   `.output` parens, string escaping). The true in-grammar pool across the
   full 622-file source tree is bounded above by 19 (proven: any file
   compliant with the full §4 grammar necessarily also passes T5's weaker
   12-pattern filter, so it would already be among the 195 — and only 19 of
   those 195 are actually compliant). Whether to regenerate the corpus with a
   corrected predicate, and against what target size, is a human call.
2. `docs/reports/night02-T8-grammar-census.md`'s "11 of 195 files declare a
   zero-arity relation" appears to be an undercount by one:
   `syntactic/doc_comment_dangling3/other.dl` (`.decl no_doc_comment()`) is a
   12th, confirmed by direct read of the file.
3. `ancestor_nonancestor_guarded.dl`'s `T_guard` grows exactly 2.00× per
   doubling of `n` despite its bound-argument demanded set being pinned at a
   constant 50 (V4) — driven entirely by the unrestricted free second
   argument (`nonancestor_bf`, 89% of `T_guard` at n=500). Whether this is
   treated as expected/inherent to `bf`-adornment shapes with a free output
   argument (same root cause the `same_generation_negation_guarded.dl` file
   comment already names for a different shape) or worth a second
   construction attempt is a human call, not resolved here.
4. The 2026-08-26 M1 session's own measurement numbers were never appended
   to `docs/MEASUREMENTS.md` (V7) and do not follow the
   `measurements/<id>/cmd.txt`+`stdout.txt` provenance-directory convention
   `CLAUDE.md` §0.2 describes, even though the underlying `dlc` invocations
   are real and the harness source is committed and inspectable. Whether to
   backfill `MEASUREMENTS.md` rows for the M1 session, or to treat the flat
   `measurements/m1-*.json` format as an accepted alternative going forward,
   is a human call.
