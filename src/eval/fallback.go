// M3.4 -- fallback evaluation. Verified, per M2-M3-BUILD.md §8's own
// instruction, before writing anything: the existing SCC-ordered
// evaluator (RunNaive/RunSemiNaive, naive.go/seminaive.go) already
// handles a mixed program -- one where the guard (src/transform/guard/
// decide.go) declined some predicates to their original, untransformed
// form while others keep their magic-set-restricted form -- with ZERO
// new evaluation machinery. Nothing in this file adds a second
// evaluation path.
//
// Why it already works: a mixed program is a plain *ast.Program, no
// different in kind from any other. sema.CheckStratification, run on the
// TRANSFORMED/mixed program (never the pre-transform StratumResult --
// see src/transform/DESIGN.md's own note on why the old StratumResult is
// invalid once magic-seed relations change the precedence graph), yields
// a valid SCC topological order over the mixed program's ACTUAL
// dependency graph, whatever mix of adorned and original relation names
// it contains. RunNaive/RunSemiNaive already process one SCC at a time in
// that order (the bug fixed in §3.9, eval/DESIGN.md) -- a declined
// predicate's original rules read another relation's full extent
// correctly for the same reason any two ordinary relations in any
// ordinary program do: the SCC order guarantees whatever it reads is
// already fully materialized by the time its own SCC runs.
//
// Verified directly: tests/corpus/CULPRIT_CANDIDATES/cc_mixed_fallback.dl
// (a program combining a genuine culprit-cycle group with a genuine safe
// group so the guard actually declines a strict, non-trivial subset --
// every other corpus program declines either nothing or everything, which
// would not exercise mixed evaluation at all). src/transform/guard/
// fallback_test.go: TestFallbackEvaluationMatchesDlcOwnEvaluator runs
// dlc's own RunNaive on both the untransformed original and guard.
// Decide's mixed final program and confirms identical answers -- and
// cmd/dlc's `dlc run --transformer=guarded` end-to-end matches real
// Soufflé on the untransformed original too (docs/reports/
// m3-4-fallback.md), three independent confirmations of the same number.
//
// cmd/dlc's runRun (§3.8/§3.9's entry point) now applies the named
// Transformer and re-derives stratification on its output before
// evaluating -- the one piece of real wiring this task's own framing
// anticipated ("fallback evaluation wiring"), not a new evaluator.
package eval
