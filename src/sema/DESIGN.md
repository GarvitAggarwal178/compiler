# package sema

Three checkers so far: declaration/arity/type (`decltype.go`, §3.4),
allowedness (§3.5), source stratification (§3.6) — each independent,
each walking the AST its own way, sharing only the package and the
`Diagnostic`/`Category` shape. Not a single unified "checker" abstraction,
because the four rejection grounds genuinely don't share structure beyond
"they produce a span and a message" — forcing them through one interface
would buy nothing.

**Key decision: a relation has exactly one schema-defining declaration.**
`.decl name(...)` introduces a schema; `.input name` / `.output name`
only *mark* an existing schema as externally loaded or exposed. Two
`.decl name(...)` for the same name is `DuplicateDecl` (matches Soufflé's
own "Redefinition of relation" — `experiments/28-souffle-diagnostic-catalogue.md`);
an `.input`/`.output` with no matching `.decl` anywhere is
`UndeclaredRelation`, not folded into duplicate-detection — they are
different mistakes with different fixes, and the diagnostic catalogue's
own instruction ("the *classification* must" match) is exactly why they
stay separate categories rather than one generic "declaration problem."

**Key decision: type checking is per-clause, not per-program.** Two
different clauses can use the same variable name (`X`) with no relation
between the two uses — Datalog variables are clause-scoped. `checkClause`
builds a fresh `varTypes` map per clause and discards it afterward; there
is no global type environment. This mirrors Soufflé's own diagnostic for
`type_number_vs_symbol_across_rule`
(`experiments/28-souffle-diagnostic-catalogue.md`), which is specifically about
one variable's uses *within one rule* conflicting, not across rules.

**Non-obvious decision, pinned by a specific rejection case: arithmetic
operators force `number` on their operands; bare comparisons do not.**
`X = Y` (both bare terms) imposes no type requirement from the comparison
itself — whatever type `X`/`Y` have (if any) comes from an atom
occurrence elsewhere in the clause. `X = Y + 1` is different: the `+`
forces `number` onto both `Y` and the whole expression, because this
grammar defines arithmetic only over numbers. Getting this wrong either
way breaks one of the two `type_*` rejection-test cases:
treating every relop side as number-forcing would spuriously reject `X =
"foo"` (a legal symbol comparison); treating no relop side as ever
number-forcing would silently accept `type_symbol_in_arithmetic`'s
`y = x + 1` where `x` is declared `symbol`. `checkConstraintSide` and
`forceArithNumber` are two separate functions specifically so this
distinction is visible in the code, not folded into one function with a
boolean flag a reader would have to trace through to notice.

**Explicitly out of scope, and disclosed rather than silently
unhandled: symbol-vs-symbol or number-vs-number comparisons where
*neither* side is otherwise typed.** `X = Y` where neither `X` nor `Y`
appears in any typed atom position anywhere in the clause is accepted
without complaint — this checker has no opinion on what such a program
means (Soufflé likely defaults it somehow; not investigated). Only
arithmetic-forced and atom-occurrence-derived types are checked. Adding
full bidirectional relop type inference was not needed by any of the
three rejection-test type cases and was left out rather than guessed at.

**Arity mismatch short-circuits type-checking for that occurrence.** If
an atom's term count doesn't match its schema, `checkAtomOccurrence`
returns immediately rather than zipping mismatched-length slices (which
would either panic or silently ignore the tail) and reporting a second,
misleading type diagnostic on top of the real arity one.

## Allowedness (`allowedness.go`, §3.5)

Implements the fixpoint definition in `record/DECISIONS.md` literally --
`G0` from positive-atom arguments, then repeated `V = E` grounding until
no change, then "every variable in the clause must be in `G`." Four
asymmetries, each forced by a specific probe program
(`experiments/19-allowedness-probe.md`, `experiments/21-allowedness-derivation.md`), and
each visible as a separate, deliberate line of code rather than folded
into one clever-looking condition:

- **Only `=` contributes** (`constraint.Op != "="` is checked before
  anything else in the fixpoint loop) — case (c), `X > Y` grounds
  nothing.
- **The grounded side must be a bare `*ast.Var`** (`groundBareSide`'s
  type assertion on `side`, not `other`) — case (j), `X+1 = Y` grounds
  `Y` (already grounded anyway) but never `X`, because `X` never appears
  as a *bare* side of any equation.
- **No arithmetic inversion** — the same `groundBareSide` asymmetry
  applied to the other case: `Y = X+1` can ground `Y` (if `X` were
  already grounded) but can never ground `X`, because grounding a
  `groundBareSide` call only ever adds the *bare* side's variable, never
  solves an equation for a variable buried inside `other`.
- **Every variable, not just head variables** — `collectAllVars` walks
  the head and every body literal including negated atoms' inner atom
  and both constraint sides, not just `c.Head`. Case (h)
  (`p(X):-q(X),Y>3.`, `Y` body-only) and case (m) (`!q(Y)`, `Y` inside a
  negation) both fail allowedness precisely because this walk finds `Y`
  and `G` never contains it.

**The fixpoint is a `for { changed := false; ...; if !changed { break }
}` loop over the *entire* body every iteration, not a single left-to-right
pass.** Cases (b) (equation before its grounding atom, order reversed),
(k) (a three-step grounding chain), (l) (mutual-looking equation), and
(o) (grounding chains through a head variable) all depend on this — a
single pass in body order would get (b) wrong outright (the grounding
atom for the equation's dependency comes *after* the equation
textually) and would under-ground (k)/(o)'s multi-step chains.

**`collectVars` (for `G0`) and `varsSubsetOf`/`collectArithVarsWithSpan`
(for the fixpoint and the final "every variable" check) all recurse
through `BinaryExpr`/`UnaryExpr`.** Nothing in the 15 probe programs
exercises a positive atom argument that is itself a compound arithmetic
expression (e.g. `q(X+1)`), but the fixpoint definition says "variables
occurring as an argument," not "arguments that are bare variables" --
recursing is the literal reading, applied consistently even where no
probe case forces the question.

## Source stratification (`stratify.go`, §3.6)

Precedence graph over IDB relations only (a relation is a node iff it is
some clause's head; an `.input`-only relation with no rules is never a
node, and edges to it are dropped when building the graph -- confirmed by
a test that would otherwise have been a false positive,
`TestStratificationDeterministicAcrossRuns`'s first draft, see below).
Tarjan SCC, then reject iff a negative edge closes within one SCC,
else assign strata via memoized DFS over the (guaranteed-acyclic)
condensation: a relation's stratum is one more than the highest stratum
among its negative dependencies, and at least as high as its positive
ones.

**Found while writing this file's tests, not by inspection: a relation
referenced both positively and negatively by two *different*,
non-mutually-recursive relations is perfectly stratifiable.** A first
draft of the determinism test used `b(x):-a(x).  c(x):-b(x),!b(x).`
expecting rejection, on the mistaken assumption that "the same relation
appears in one program both positively and negatively" was itself
suspicious. It isn't: `b` and `c` aren't in each other's SCC (there's no
cycle at all), so this stratifies cleanly (`b` at stratum 0, `c` at
stratum 1) -- exactly what stratified negation is supposed to allow. The
test was wrong, not the stratifier; fixed to use the actual
self-negative-cycle case instead. Left here because it's a genuine,
easy-to-make misconception about what "unstratifiable" means.

**Gate two's oracle signal took a wrong turn before it worked.** The
plan was to compare against Soufflé's `SUBROUTINE <Relation>` emission
order in `souffle --show=initial-ram <file>` output. That order turned
out to be **alphabetical, not evaluation order** (confirmed on
`evaluation/set_ops_output/set_ops_output.dl`: subroutines are listed
`A, AintersectionB, AminusB, AunionB, AxB, B, BminusA` -- pure sort
order). The actual execution order lives in a separate
`BEGIN MAIN ... CALL stratum_<Relation> ... END MAIN` block, which is
**also not sorted by numeric stratum** (that same file calls
`stratum_AminusB`, stratum 1 by this project's own numbering, *before*
`stratum_AunionB`/`stratum_AxB`, stratum 0) -- Soufflé emits some valid
topological order of the dependency DAG, not relations grouped by
stratum number. The comparison this project's gate actually needs, and
the one implemented in `harness/m1_3_6_stratification.py`, is narrower
and is the actual correctness invariant of stratified negation: for
every negated-atom edge `X -> !Y`, `Y`'s `CALL` must precede `X`'s in
Soufflé's `MAIN` sequence, and `dlc`'s `stratum[Y] < stratum[X]` must
hold -- not that the two tools agree on an exact sequence or on stratum
numbering itself, which was never something they were expected to share.

**`ClauseVarTypes` (NIGHT-BATCH-03 T8), `decltype.go`.** Exports the same
per-clause variable-type environment `checkClause`'s internal
`clauseChecker` already built (`runClauseChecker` factors the shared
construction out so `checkClause` and `ClauseVarTypes` are two thin
callers of one function, not a duplicated copy). Added because `codegen`
needed to know a variable's declared type (`symbol` vs `number`) to fix
the symbol-ordering codegen bug and had no way to compute it without
either reimplementing this exact logic a second time or exporting it --
exporting was the smaller, more honest change.
