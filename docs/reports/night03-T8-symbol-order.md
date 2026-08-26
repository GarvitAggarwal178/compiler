# NIGHT-BATCH-03 T8 — codegen symbol-ordering gap

Date: 2026-08-27. Named as a known, disclosed gap in `m1-progress.md`'s
own "what a skeptic attacks first" section; untested is not verified.

## What did not work

Confirmed the bug was real, not hypothetical, before fixing it: with the
fix's condition forced off (`if false && isOrderingOp(...)`), all 4 new
tests fail, and fail with exactly the predicted wrong answer set (id-based
`<` on intern order `zebra=0,apple=1,mango=2` wrongly answers
`{zebra<apple, zebra<mango, apple<mango}` instead of the correct
`{apple<mango, apple<zebra, mango<zebra}`). Restored the fix immediately
after confirming this; the failing-then-passing pair is what makes these
regression tests, not just tests that happen to pass.

## The defect, and the fix

`src/codegen/evaluation.go`'s `emitConstraint` compared every relational
operator's operands as raw `int64_t`, including `<`,`<=`,`>`,`>=`. Since
`symbol` columns are represented as their interned id (`src/codegen/
DESIGN.md`'s own "types erased to `int64_t` uniformly" design, correct and
deliberate for `=`/`!=`), an ordering comparison on a `symbol` column
compared *assignment order*, not *lexicographic order* — exactly the
disclosed gap.

Fixed by exporting `sema.ClauseVarTypes` (`src/sema/decltype.go`, a thin
wrapper factored out of `checkClause`'s existing internal
`clauseChecker.varTypes` — no duplicated logic, no behavior change to any
existing sema check) and using it in `emitConstraint`: when an ordering
operator's operand is symbol-typed (`isSymbolArith`: a bare string literal,
or a `Var` whose `ClauseVarTypes` entry is `"symbol"` — arithmetic
operators are always number-typed in this grammar, so no other `ast.Arith`
shape can be symbol-typed), emit `strcmp(str_lookup(x), str_lookup(y)) OP 0`
using the prelude's existing `str_lookup` un-intern function, instead of a
raw integer comparison. `=`/`!=` are unchanged (id equality is string
equality regardless of intern order — never affected by this bug).

## Method

4 new end-to-end tests (`src/codegen/codegen_test.go`,
`TestCodegenSymbolOrdering{LessThan,LessOrEqual,GreaterThan,
GreaterOrEqual}`): generate → compile with `cc` → run the binary → check
output, the same real end-to-end pattern every other `codegen_test.go` case
uses. Facts (`"zebra"`, `"apple"`, `"mango"`, in that source order — intern
order is therefore the *reverse* of lexicographic order) are chosen
specifically so a regression produces a *different answer set*, not merely
a different row ordering. Cross-checked against real Soufflé on the
identical program (`.decl q(a:symbol) ... p(x,y):-q(x),q(y),x<y.`):

```
$ souffle -D out sym.dl && sort out/p.csv
apple	mango
apple	zebra
mango	zebra
$ dlc codegen sym.dl sym.c && cc sym.c -o sym && ./sym . out && sort out/p.csv
apple	mango
apple	zebra
mango	zebra
```

Set-identical (sorted lines match exactly, both engines agree on all 3
tuples).

## The numbers

| metric | value |
|---|---|
| cases (4 operators: `<`,`<=`,`>`,`>=`) | 4/4 pass with the fix |
| cases confirmed to fail without the fix | 4/4 (verified directly, fix temporarily disabled) |
| Soufflé cross-check on the shared example program | set-identical |
| **A defect was found: yes.** | The disclosed gap was real, not merely theoretical — this is the required explicit statement, not silence. |

`go build`/`go vet`/`go test ./...` all clean after the fix (full suite,
not just `codegen`).

## What a skeptic attacks first

- `isSymbolArith` only recognizes a bare `Var` or `StringLit` as
  symbol-typed. This is exhaustive for this grammar (arithmetic operators
  force `number` on every operand they touch — `sema/decltype.go`'s
  `forceArithNumber`, already-tested code, unmodified here) but the
  argument rests on that invariant holding, not on an independent proof
  reconstructed in `codegen` itself.
- `ClauseVarTypes` is only correct for an already-accepted clause (a
  clause with sema diagnostics can return an incomplete map, per
  `checkAtomOccurrence`'s early-return on an arity mismatch) — safe here
  because `codegen.Generate` is only ever called on a program that already
  passed every sema check (`cmd/dlc`'s `runCodegen`), but this is a
  documented precondition, not independently enforced by `ClauseVarTypes`
  itself.
- Only one cross-check program was run against Soufflé, not a sweep. The
  4 Go unit tests are the actual regression coverage; the Soufflé
  cross-check is corroborating evidence that the *fixed* semantics
  (lexicographic string comparison) is the right target to fix toward, not
  an independently-sized test suite of its own.

## Verdict

**T8: DONE. A real defect was found and fixed, not merely tested for and
confirmed absent.** 4/4 new cases pass after the fix, 4/4 confirmed to fail
before it, Soufflé cross-check set-identical. `src/codegen/DESIGN.md` and
`src/sema/DESIGN.md` updated to record the fix and the new export.
