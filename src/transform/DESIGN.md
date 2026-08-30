# package transform (top level only — not magicset/ or guard/)

`magicset/` and `guard/` are Lane A (empty `doc.go` markers only,
specs/02-m1-build.md §1). This file (`transformer.go`, at the `transform`
package level, a sibling of those two subpackages, not inside either) is
Lane B: the interface Lane A will implement, plus a placeholder so the
pipeline runs without it.

**Key decision: `Transform` takes the already-*source*-stratified
program and its `*sema.StratumResult`, and returns a plain
`*ast.Program`, not some richer "transformed program" type.** The
evaluator (`eval.RunNaive`/`RunSemiNaive`) only knows how to run an
`ast.Program` — giving `Transform` a different return type would mean
either the evaluator needs a second code path for "transformed" programs
specifically, or every caller has to unwrap/convert immediately. Handing
back an ordinary `*ast.Program` (magic predicates and all, once Lane A
implements it) means `cmd/dlc`'s pipeline calls `Transform` once and
feeds the result to the exact same `RunNaive`/`RunSemiNaive` it would
have called on the untransformed program — no branching on which kind of
program it got.

**Documented, not solved here: a real transform invalidates the
`strata` argument for its own output.** Introducing magic-seed relations
changes the precedence graph, so a transformed program's true
stratification can differ from the pre-transform one passed in. This
interface doesn't pretend to solve that (returning a second
`*sema.StratumResult` was considered and rejected — it would commit Lane
A to running its own stratification pass with a fixed contract before
the guard's own design is settled, which is exactly the kind of
resolve-it-for-them overreach CLAUDE.md's Lane A boundary exists to
prevent). Documented in `Transformer`'s own doc comment instead: whoever
implements this re-runs `sema.CheckStratification` on the output if they
need strata for it.

**`PassThrough` returns the identical pointer, not a copy.** Verified by
`TestPassThroughReturnsInputUnchanged` — cheap to check, and confirms
"pass-through" means what it says (no accidental deep-copy-then-return-
equivalent-but-different-object, which would still be correct but would
make `PassThrough`'s cost non-zero for no reason).

**Not wired into `cmd/dlc`'s `run`/`run-seminaive` yet, on purpose.**
Wiring a no-op `Transformer` into the CLI pipeline today would add a
call that changes nothing observable — §4 item 2 asks for the interface
and a pass-through implementation to exist, not for the CLI to route
through it before there is a second implementation to justify the
indirection. The natural place to wire it in is exactly when Lane A's
real `Transformer` lands, at which point `runRun` gains one call between
`sema.CheckStratification` and `RunNaive`/`RunSemiNaive` and nothing else
in `cmd/dlc` changes.
