# `dlc` — design, as it stands today

`dlc` is a from-scratch Datalog compiler (Go, single binary, `src/cmd/dlc`)
built as a semester compiler-design project. It compiles a fixed subset of
Soufflé's syntax to an executable evaluation plan, and implements a
magic-set transform that treats negated occurrences by the same
demand-driven mechanism as positive ones — guarded by a soundness check
that the transform itself can violate, with a per-relation fallback when it
does.

This document describes the compiler as it exists now. For how it got
here — the amendments, the false starts, the corrected predicates — see
`docs/design-history.md` and `docs/project-log.md`. For the numbers behind
every claim below, see `results/claims.md` and `experiments/`.

---

## 1. Source language

```
program    ::= decl* clause*

decl       ::= '.decl' ident '(' param (',' param)* ')'
             | '.decl' ident '(' ')'                  // amendment: zero-arity
             | '.input' ident
             | '.input' ident '(' ')'                 // amendment: optional parens
             | '.output' ident
             | '.output' ident '(' ')'                // amendment: optional parens
param      ::= ident ':' type
type       ::= 'number' | 'symbol'

clause     ::= atom '.'                       // fact
             | atom ':-' body '.'             // rule
body       ::= literal (',' literal)*
literal    ::= atom
             | '!' atom                       // negation
             | constraint
atom       ::= ident '(' term (',' term)* ')'
             | ident '(' ')'                  // amendment: zero-arity
term       ::= arith | '_'                    // wildcard
constraint ::= arith relop arith
relop      ::= '=' | '!=' | '<' | '<=' | '>' | '>='

arith      ::= arith ('+' | '-') mul | mul
mul        ::= mul ('*' | '/' | '%') unary | unary
unary      ::= '-' unary | primary
primary    ::= var | number | string | '(' arith ')'
```

A strict subset of Soufflé's syntax, closed by design: no functors, no
aggregates, no records, no components — every test program runs unmodified
on Soufflé, which is the compiler's sole oracle. The grammar admits exactly
two amendments beyond the original design, both narrow and both
independently authorized: zero-arity relations (`.decl foo()`,
`experiments/11-m1-harness-buildout.md`'s corpus census found 12 of 195
in-grammar Soufflé test files use them — corrected from an initial count
of 11, `experiments/32-verification-pass.md`) and optional parentheses on
`.input`/`.output` names (`experiments/38-grammar-amendment-optional-
parens.md`). No other feature has been added.

Against Soufflé's own `tests/` corpus (`tests/corpus/IN_GRAMMAR.txt`), 19
of 195 files are strictly compliant with the grammar above before either
amendment (`experiments/50-four-reconciliations.md`); 89 of the full 195
parse after the parenthesized-directive amendment
(`experiments/38-grammar-amendment-optional-parens.md`). These are corpus-
admissibility numbers — how much of a real-world Soufflé test suite this
grammar accepts — not a coverage or correctness metric; see
`results/claims.md` for how they're allowed to be cited.

## 2. Pipeline, pass by pass

1. **Lexer** (`src/lexer`) — hand-written, no generator (the parser is a
   learning objective for this course).
2. **Precedence parser** (`src/parser`) — produces `*ast.Program` via
   recursive descent for clause structure and precedence climbing for
   arithmetic. `parser.Print`/`parser.Roundtrip` keep the printer honest
   against re-parsing its own output.
3. **Declaration / arity / type check** (`sema.CheckDeclType`) — rejection
   grounds `arity_mismatch` and `type_mismatch`.
4. **Allowedness** (`sema.CheckAllowedness`) — every variable in a rule
   must be grounded by some positive body literal before it can appear in
   a negated literal, a constraint, or the head. Rejection ground
   `allowedness`.
5. **Stratification** (`sema.CheckStratification`) — Tarjan SCC over the
   full positive+negative precedence graph; a negative edge inside an SCC
   is rejected. Rejection ground `unstratifiable`.
6. **Magic-set transform** (`src/transform/magicset`) — §3 below.
7. **Transform-safety guard** (`src/transform/guard`) — §4 below.
8. **Evaluation** — naive and semi-naive (`src/eval`) fixpoint evaluators.
   A mixed program (partially transformed, partially fallback) uses the
   same evaluator unmodified.
9. **C codegen** (`src/codegen`) — naive evaluation to standalone C, for an
   *untransformed* program only. Never wired to the transform path; the
   pipeline story (parse → transform → evaluate) is complete only for the
   tree-walking evaluator.

All four rejection grounds are demonstrated live via `dlc explain`, one
committed sample program per ground (`experiments/explain-samples/`).

## 3. The magic-set transform

`src/transform/magicset`: adornment (`adorn.go`), sideways information
passing / SIPS (`sips.go`), magic rules and supplementary predicates
(`rules.go`). SIPS is left-to-right, source-order — no cost model, by
design. Negated IDB atoms are adorned and given magic rules by the exact
same mechanism as positive ones, never skipped.

**Multi-query seeding.** The adornment worklist is seeded from *every*
bindable query candidate in the program (`FindQueries`), not the first one
found. A program with two independent `.output` branches gets both
demand-restricted independently; seeding from only the first left a
sibling branch at full extent, which was this project's own bug for most
of a session (see `docs/design-history.md`).

**The `V_i` projection.** For an adorned rule with SIPS-ordered body
`L₁..Lₙ`, the supplementary chain `sup_r_i(V_i) :- sup_r_{i-1}(V_{i-1}),
Lᵢ` must project `V_i` down to only the variables still needed —
`(variables bound after processing L₁..Lᵢ) ∩ (variables occurring in
Lᵢ₊₁..Lₙ, or in the head)`. Skipping this projection reimplements naive
evaluation with extra relations.

**Demand relaxation on negated occurrences.** A bound position in a
negated occurrence's adornment can be relaxed to free if its variable's
only binder in the SIPS prefix is an unrestricted full-extent scan — sound
because a magic set with fewer bound positions demands a superset of what
the negation needs, and completeness under negation requires *covering*
the queried instantiations, not equalling them. This is a structural rule,
not a cost model: it collapses several benchmark shapes to a single
adornment each, matching what a hand-written guard would produce, and
moved this project's own measured contribution from sub-1×–16× to
46×–1,343× on one shape (`experiments/49-demand-relaxation.md`).

## 4. The soundness problem, and the guard

**Magic sets can destroy stratification.** Restricting a negated
relation's extent can introduce a negative cycle through the
magic-seed/supplementary chain that did not exist in the source program
(`tests/corpus/BENCHMARK_FAMILY/culprit_cycle.dl`'s own header comment
derives the exact cycle: `magic_q ->~ s -> q -> magic_q`).

**The two-clause guard.**

- **Clause (a) — stratification preservation.** Re-run stratification on
  the candidate transformed program (`src/transform/guard/stratify.go`).
  If unstratifiable, the implicated predicates and their full dependency
  cone (`decide.go`) fall back to reading their original, untransformed
  extent — a per-SCC decision, not all-or-nothing for the whole program.
- **Clause (b) — completeness under negation.** Collapses into correct
  seeding, given allowedness. Allowedness requires every variable in a
  negated literal to be grounded before it is evaluated, so the adornment
  computed for a negated atom is always all-bound *before* demand
  relaxation. `magic_q^{b...b}` therefore contains ground tuples only, and
  demand-restricting `q^{b...b}` to exactly those tuples decides exactly
  the membership questions the negation asks — completeness holds by
  construction, given correct seeding. Demand relaxation (§3) amends this
  without contradicting it: the pre-relaxation adornment is still always
  all-bound (asserted in code, `guard.AssertNegationSeeding`), and
  relaxation is proven separately sound as its own lemma (a relaxed
  position only ever computes a superset). Both halves of the invariant
  are checked in code, not just argued in prose.

**The fallback cone.** When clause (a) fires on relation `r`, the fallback
is not `r` alone — it is `r`'s full dependency cone in the transformed
program, computed exactly (cross-checked against an independent
implementation, `harness/cone_metric.py`). A relation in the cone is
evaluated against its original, untransformed extent; this must be fully
materialized before any still-transformed relation consumes it, so mixed
evaluation is an ordering constraint on the same evaluator, not a second
evaluation strategy.

**Blast radius.** Across a 16-program corpus, the guard fires only on
programs with a genuine culprit-cycle shape (11/16) and on none without
one — it is not vacuous, and it agrees with Soufflé's own stratifier
everywhere checked. See `results/claims.md` for the exact figures and
their provenance.

## 5. What the guard is not

- **No cost-based SIPS.** The transform's ordering is fixed left-to-right,
  source order, by design — `dlc`'s mechanical transform is beaten by a
  hand-written guard on every shape and scale point measured, narrower
  after demand relaxation than before it, but never absent.
- **No wall-clock timing anywhere.** Every number in this project is a
  derived-tuple count; the hardware (WSL2, hybrid CPU, no PMU) cannot
  support a timing claim.
- **C codegen is not wired to the transform.** The generated-C path only
  ever sees an untransformed program.
- **The guard's necessity is not demonstrated on real code.** Its
  correctness is; see `results/findings.md` item 4.
