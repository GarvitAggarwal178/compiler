# package eval

`fallback.go` is Lane A (specs/02-m1-build.md §1) and contains only the
marker comment. Everything else here — `naive.go` (§3.8), `seminaive.go`
(§3.9), `io.go` — is Lane B.

## The SCC-vs-stratum evaluation-order bug (§3.9, found by the
differential harness, not by inspection)

`RunSemiNaive`'s first version grouped clauses by **stratum number** for
its seed/recursive-round loop, mirroring `RunNaive`. §3.9's own gate one
("same set equality as 3.8, unchanged") caught a real disagreement on
`example/josephus/josephus.dl`: `Josephus` was missing a tuple
(`"e"`) that both naive evaluation and real Soufflé agree belongs there.

Root cause: `josephus.dl` has a self-recursive `Relation` and a second,
independent `Josephus` that reads it positively — no negation anywhere in
the file, so nothing forces `Josephus` to a strictly later stratum
*number*; both land at stratum 0. Grouping by stratum number put
`Josephus`'s (correctly seed-classified) clause in the same batch as
`Relation`'s clauses, so it ran during the **seed round** — before
`Relation`'s own recursive rule had produced anything beyond its 6
initial facts. `Josephus` silently got evaluated against a
not-yet-converged `Relation`.

**Why `RunNaive` never showed this bug:** naive evaluation repeats its
*entire* combined batch to a full joint fixpoint regardless of internal
structure — `Josephus`'s clause gets harmlessly re-evaluated on every
pass, and by the pass where `Relation` finally stops changing, `Josephus`
picks up everything it was missing for free. Semi-naive's entire point is
to *not* redundantly re-run a seed clause once its dependencies are
stable — which is exactly the case this bug violated.

**Fix:** `sema/stratify.go`'s `StratumResult` gained `SCCOrder`, a plain
topological order of the SCC condensation (every SCC a given SCC depends
on, by any edge, appears earlier). `RunSemiNaive` now processes one SCC
at a time in that order (`semiNaiveSCC`), not one stratum-number batch at
a time — stratum number is only ever needed for the negation-safety
*rejection* check in `stratify.go`, never for driving evaluation order.
`sameSCCAtomIndices` (renamed from `sameStratumAtomIndices`) classifies
"recursive vs seed" by SCC membership too, which was already the right
granularity for that decision (see below) — SCC-order processing is what
makes that classification actually *safe* to act on.
`TestSCCOrderRespectsPositiveDependencyEvenWithinOneStratum`
(`sema/stratify_test.go`) and
`TestSemiNaiveWaitsForDependencySCCEvenAtSameStratum`
(`eval/seminaive_test.go`) pin the exact shape that broke, both now
passing; `harness/m1_3_9_gate1_seminaive_agreement.py` re-confirmed
11/20 against real Soufflé, matching §3.8's baseline exactly, after the
fix.

**`Evaluator` bundles relation storage and the string interner into one
struct** purely so the many small helper functions (`evalArith`,
`tryUnify`, `groundTuple`, ...) don't each need both threaded through
their parameter list. Nothing here is concurrent; there is no locking
because there is nothing to protect against.

**Key decision, forced by a real bug an early test caught:
`safeOrder` reorders a clause's body before naive evaluation walks it
left to right.** `sema.CheckAllowedness`'s fixpoint (§3.5) is
deliberately order-independent — `X = Y + 1, q(Y).` is exactly as
allowed as `q(Y), X = Y + 1.` (probe case b). But a naive left-to-right
*evaluator* cannot process `X = Y + 1` before anything has bound `Y`: at
that point in the walk, `Y` isn't in `bindings` yet, and there is nothing
sensible to do with an equation whose right-hand side isn't computable.
`safeOrder` greedily schedules whatever literal is currently safe (a
positive atom always is; a negated atom or a non-grounding constraint
only once every variable it needs is already bound; a grounding `V = E`
constraint once `E`'s variables are bound, even if `V` itself isn't yet)
and repeats until the whole body is scheduled. `TestEquationBeforeGroundingAtomSafeOrder`
pins this directly against probe case (b)'s own shape.

**The `=` constraint has two different runtime behaviors, matching its
two different roles in allowedness's own fixpoint.** `evalConstraint`
checks, in order: is one side a not-yet-bound bare `Var` and the other
side's variables all bound? If so, this is a *grounding* occurrence —
evaluate the other side and bind the variable (an assignment, not a
test). Otherwise it's a plain equality *test* — evaluate both sides
(now guaranteed possible, `safeOrder` already ensured it) and compare.
Every other relop (`!=`,`<`,`<=`,`>`,`>=`) is always a test, never a
grounding — mirrors sema's own "only `=` contributes" rule
(`sema/DESIGN.md`) exactly, because it has to: an evaluator that grounded
through `<` would accept programs allowedness had already rejected as
unsafe, and one that couldn't evaluate a same-side-bound `=` as a test
would reject some allowed programs outright.

**A found-by-testing correction, not a design decision made in
advance: source-level fact clauses (`edge(1,2).`) ARE counted as
rule-derived tuples, matching Soufflé's own convention** — only
`.input`-loaded data (`LoadFacts`, which never calls
`RecordSeedInsert`/`RecordIterationInsert`) is excluded. A first draft of the test now named
`TestInstrumentationExcludesInputLoadsButCountsSourceFacts` assumed fact
clauses should also be excluded and failed; checking `harness/tuple_report.py`'s actual
`is_input_relation` predicate (keyed on Soufflé's own `"loadtime"`
profile attribute, which only a real `.input` load produces) confirmed
the *test's* assumption was wrong, not the evaluator's behavior — a plain
source fact is, to Soufflé, just a non-recursive rule like any other, and
is counted. Fixed by correcting the test, not the code, and logged here
so the reasoning survives past the one commit that fixed it.

**Naive evaluation records every insertion into iteration bucket 0**
(`ir.RelationStats.IterationInserts[0]`) — there is no seed/delta
distinction to make without semi-naive's Δ-rewrite, so `RecordIterationInsert(0)`
is called uniformly for every rule-derived tuple regardless of which
pass of the fixpoint loop produced it. `Total()` (ir/DESIGN.md) still
gives the right number either way.

**A negated atom's terms are required to already be fully ground**
(`groundTuple` returns `ok=false`, silently skipping that branch, if any
term isn't `ast.Arith`-evaluable) — this can only happen for a `Wildcard`
inside a negated atom's argument list, a construction allowedness does
not itself forbid but that has no sensible ground-truth membership test
(`!q(_, X)` — is that "for all possible first columns" or "there exists
one"? Soufflé's own semantics for this shape were not investigated).
Disclosed as an unhandled edge case, not silently miscomputed as a
false-negative or false-positive.

**Δ-rewrite variants are keyed by body-literal *position*, not relation
name** (`evalBody`'s `overrideIdx map[int]*ir.Relation`, `naive.go`).
A self-join — the same relation appearing twice in one recursive rule's
body, e.g. `p(x,y):-p(x,z),p(z,y).` — needs one Δ-rewrite variant per
*occurrence*, each redirecting only that one occurrence to the delta
relation while the other occurrence still reads the full accumulated
relation. Keying by name instead would redirect every occurrence of that
relation at once, silently dropping the old-new/new-old combinations
semi-naive evaluation exists to still catch.
`TestSemiNaiveSelfJoinMatchesNaive` (`eval/seminaive_test.go`) checks
this against a 5-node chain computed via `p(x,y):-p(x,z),p(z,y).`
specifically (not the edge-driven `p(x,y):-p(x,z),edge(z,y).` shape,
which has no self-join to get wrong).

**A `Wildcard` in *head* position falls back to a fixed value
(`NumberValue(0)`) instead of panicking.** Real Soufflé rejects this
construct outright (`experiments/26-hostile-source-corpus.md`,
`semantic_wildcard_in_head.dl`) but sema (§3.4-§3.6, as implemented)
doesn't check for it — adding that check was out of scope creep for this
item. The fallback exists purely so a program containing this construct
degrades to "produces a wrong-but-harmless tuple" rather than a crash,
consistent with "never panic" as the higher-priority invariant when the
two goals conflict.
