# M1-BUILD — revised lane boundary and full M1 specification

Date: 2026-08-23. **§1 supersedes `CLAUDE.md` §2.** Everything else in `CLAUDE.md`
stands: provenance discipline (§0.2), no wall-clock in reports (§0.3), oracle-first
testing (§6), reporting rules (§7), session hygiene (§8).

---

## 1. Revised authorship lanes

`CLAUDE.md` §2 put all of `src/` in Lane A. That was scoped for defence under hostile
viva questioning, which is not the assessment format. The boundary is narrowed.

### Lane A — human-authored. Do not write these. Ever.

Three components only:

1. **Magic-set transform** — adornment, SIPS, magic rule generation, supplementary
   predicates. `src/transform/magicset/`
2. **The guard** — culprit-cycle detection on the *transformed* precedence graph, the
   per-SCC TRANSFORM/FALLBACK decision. `src/transform/guard/`
3. **Fallback evaluation wiring** — mixed evaluation where untransformed relations are
   fully materialised before transformed ones consume them.
   `src/eval/fallback.go`

Create these directories with a single `doc.go` stating `// Lane A — human-authored.
See docs/M1-BUILD.md §1.` and nothing else. Do not stub the functions. Do not sketch
the algorithms in comments.

If a Lane B component needs to call into Lane A, **define the interface**, put the
interface definition in Lane B, and provide a Lane B implementation that returns
"not implemented". The human fills it in later. Say clearly in your report which
interfaces you defined and what each expects.

### Lane B — yours. Write freely.

Everything else in `src/`: lexer, parser, AST, symbol table, declaration and type
checking, allowedness, source-program stratification, relation storage and indexing,
naive and semi-naive evaluation, C code generation, CLI. Plus all of `harness/`,
`tests/`, `docs/`.

### New obligation: explainability

The human must be able to explain every component without having written it. For each
package you create, write `DESIGN.md` in that package directory covering: what it
does, the key data structures and why those, the one or two non-obvious decisions and
what the alternatives were, and where the tricky cases are. Two to three paragraphs.
Not API documentation — the reasoning.

This is a deliverable, not a nicety. A package without `DESIGN.md` is not done.

---

## 2. Language, layout, dependencies

Go, per `CLAUDE.md` §3. **No parser generator, no lexer generator, no third-party
parsing library.** Standard library only for `src/`. Python remains fine in
`harness/`.

```
src/
  cmd/dlc/main.go              CLI
  token/                       token types, spans
  lexer/
  ast/
  parser/
  sema/                        decl+type check, allowedness, stratification
  ir/                          relations, tuple storage, indices
  eval/                        naive, semi-naive; fallback.go is Lane A
  transform/
    magicset/doc.go            Lane A — empty
    guard/doc.go               Lane A — empty
  codegen/                     C emission
```

---

## 3. M1 work items, in order

Each item has an acceptance gate. **Do not start the next item until the current
gate passes.** Commit per item, message `[B][m1][<item>]`.

### 3.1 Token and lexer

Token carries: kind, literal text, and a source span (byte offset, line, column for
both start and end). Every token, no exceptions — every diagnostic downstream needs
it and retrofitting spans is miserable.

Token kinds: identifier, number, string, `:-`, `.`, `,`, `!`, `(`, `)`, `:`, `_`,
the six relops (`=`, `!=`, `<`, `<=`, `>`, `>=`), the arithmetic operators
(`+`, `-`, `*`, `/`, `%`), and the directives `.decl`, `.input`, `.output`.

The `.` problem is the real content here: `.decl` and a clause-terminating `.` start
the same way. Resolve it in the lexer by lookahead, not in the parser. Document the
choice in `DESIGN.md`.

Errors produce an **error token with a span**, never a panic, never an exception.
Lexing continues past an error so a file yields all its errors, not just the first.

**Gate:** `harness/parse_coverage.py` runs the lexer over all 195 files in
`tests/corpus/IN_GRAMMAR.txt` with zero panics. All 39 files in `tests/hostile/`
produce either a clean token stream or error tokens with positions — zero panics.
Both corpora already exist.

### 3.2 AST

Define the types before writing any parse function; the AST shape is what four
downstream passes are written against.

`Program`, `Decl`, `Clause` (fact / rule), `Atom`, `Literal` (atom / negated atom /
constraint), `Term`, `Arith` (binary op / unary op / variable / number / string /
wildcard). Every node carries a span.

### 3.3 Parser

Recursive descent for clause structure; precedence climbing or Pratt for `arith`.
Grammar is blueprint §4 with one amendment: **the term list in an atom is optional**,
so zero-arity relations parse. NIGHT-BATCH-02 T8 found 11 of the 195 in-grammar files
use them.

Precedence, tightest first: unary `-`; `*` `/` `%`; `+` `-`; relops. Relops do not
chain.

Error recovery: on a parse error, record it with a span and skip to the next `.` at
clause level, then continue. One malformed clause must not kill the file.

**Gate one:** all 195 in-grammar files parse with zero errors. This number is the
day's headline; report it as `parsed/195`.

**Gate two:** round-trip. `harness/round_trip_scaffold.py` already exists and expects
a pretty-printer. Write the printer, then: parse → print → reparse → assert the two
ASTs are structurally equal, on all 195. This catches precedence bugs that reading
the code never will.

**Gate three:** the 39 hostile files produce the accept/reject verdicts recorded in
`docs/reports/night02-T2-hostile.md`, which were established against Soufflé. Where
you disagree with the oracle, report the disagreement — do not adjust to match.

### 3.4 Semantic analysis — declaration and type checking

Symbol table from `.decl`. Check every atom occurrence for arity and argument type
against its declaration. Reject undeclared relations, duplicate declarations, arity
mismatches, type mismatches.

Diagnostics carry a span and follow the catalogue in
`docs/reports/night02-T9-diagnostics.md`. Message text need not match Soufflé's
verbatim, but the *classification* must.

### 3.5 Semantic analysis — allowedness

Implement exactly this definition. It was derived empirically from fifteen probe cases
(a–o) against Soufflé and is recorded in `docs/DECISIONS.md`:

> `G₀` = the set of variables occurring as an argument of a positive body atom.
> `Gᵢ₊₁ = Gᵢ ∪ { V : the body contains a constraint `V = E` or `E = V` where `V` is a
> bare variable and `vars(E) ⊆ Gᵢ` }`.
> `G` = the limit of the sequence.
> **A clause is allowed iff every variable occurring anywhere in the clause is in
> `G`.**

Four asymmetries, each pinned by a probe case. Implement them deliberately:

- Only `=` contributes. Inequalities never ground anything (case c).
- The grounded side must be a **bare variable**, not an expression (case j).
- No arithmetic inversion: `Y = X + 1` with `Y` grounded does **not** ground `X`
  (case d).
- The quantifier is over **every** variable in the clause, not only head variables
  (cases h, m).

It is a fixpoint, not a left-to-right scan (cases b, k, l, o).

**Gate:** all 15 probe programs in `tests/programs/allowedness_probe_*.dl` produce the
verdicts recorded in `docs/reports/night02-T1-allowedness.md` and
`docs/reports/J1-allowedness-probe.md`. Plus all 13 cases in `tests/rejection/`.

### 3.6 Semantic analysis — source stratification

Precedence graph over IDB relations, edges labelled positive/negative. Tarjan SCC.
Reject if any SCC contains a negative edge. Output a stratum ordering.

This is the *source* program's stratifier and is Lane B. The culprit-cycle detector
that runs on the *transformed* program is Lane A and is a different thing — do not
write it, and do not generalise this one in anticipation of it.

**Gate:** correctly rejects an unstratifiable program; produces a stratum order
agreeing with Soufflé's evaluation order on the in-grammar corpus programs that
contain negation.

### 3.7 Relation storage and indices

Tuples of int64 and interned strings. Per-relation storage with at least one index
supporting prefix lookup on a bound argument prefix — that is what join evaluation
needs. Index selection can be naive for now; record in `DESIGN.md` what a better
selection would do.

**Instrumentation, required:** a counter of tuples inserted per relation per
iteration, emitted as structured JSON matching what `harness/parse_profile.py`
already produces for Soufflé. The entire measurement apparatus depends on this
being comparable. Count rule-evaluation insertions; exclude EDB loads; report both
copy conventions where a copy exists.

### 3.8 Naive evaluation

Per stratum, iterate to fixpoint. Nested-loop join with index lookup, driven by a
variable order per rule. Correctness before speed.

**Gate:** set equality against Soufflé on every in-grammar corpus program the front
end accepts. Use the existing differential harness. Report `agreed/attempted` with
the symmetric difference for any disagreement.

### 3.9 Semi-naive evaluation

Δ-rewrite per stratum. Each recursive rule becomes a set of rules, each with one
body atom replaced by its delta.

**Gate one:** same set equality as 3.8, unchanged.

**Gate two — M1's headline number:** `T_naive` vs `T_semi-naive`, exact tuple counts,
on a fixed program set. This is M1's optimization pass and the number that makes the
milestone submittable on its own.

---

## 4. If M1 completes

Do **not** start the magic-set transform. It is Lane A.

Useful and permitted, in this order:

1. C code generation (§2 layout `codegen/`) — nested loops over relations with hash
   indices, emitting a standalone C file. This is the target-code phase and may be
   rubric-mandatory; getting it early de-risks that.
2. The interface Lane A will implement: define `transform.Transformer` in Lane B with
   the signature the evaluator needs, and a pass-through implementation, so the
   pipeline runs end-to-end before the transform exists.
3. Extend the differential harness to run `dlc` against Soufflé across the full
   benchmark family at every pre-registered scale point.

---

## 5. Prohibitions

1. **Do not write Lane A code** (§1). Not stubs, not comments describing the
   algorithm, not "a simple version to be replaced".
2. Do not edit `tests/corpus/PREREGISTERED.txt`, `IN_GRAMMAR.txt`,
   `BENCHMARK_FAMILY/`, `SCALE_POINTS.json`, or the predicates that built them.
3. Do not edit `docs/dlc-blueprint.md`. Append-only docs may be appended to.
4. **Never regenerate a golden file from `dlc`.** Goldens come from Soufflé only. The
   guard for this already exists in the harness; do not disable it.
5. No wall-clock in any report. Tuple counts only.
6. All measurements go through file-redirected invocation, never an interactive
   stream read — NIGHT-BATCH-02 established that the `wsl.exe` bridge can race and
   misreport exit codes.
7. Do not `git push --force`, rewrite history, or delete a `measurements/` directory.

---

## 6. Escalation

`CLAUDE.md` §5 STOP-and-wait is in force. No night-batch continue-on-escalation
semantics unless a batch explicitly authorises it.

Additionally, stop and report if:

- A gate cannot be passed after two focused attempts.
- Soufflé and `dlc` disagree on a program and you cannot explain why.
- A gate would pass only by weakening the gate.
- The work requires writing Lane A code.

---

## 7. Reporting

Per work item, append to `docs/SESSION_LOG.md`: item, gate result as a number
(`195/195`, `13/13`, `T_naive`/`T_semi-naive`), commit SHA, what is now blocked.

At the end of the session, `docs/reports/m1-progress.md`: which items are complete,
each gate's number, which Lane A interfaces you defined and what they expect, and what
a skeptic attacks first.

The "what did not work" section comes before the results section, per `CLAUDE.md` §7.