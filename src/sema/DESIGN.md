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
own "Redefinition of relation" — `docs/reports/night02-T9-diagnostics.md`);
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
(`docs/reports/night02-T9-diagnostics.md`), which is specifically about
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
