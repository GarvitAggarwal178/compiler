# package magicset

M2. Implements the magic-set transform: adornment (`adorn.go`), SIPS
(`sips.go`), magic rules + supplementary predicates (`rules.go`), and a
`transform.Transformer` wrapper (`transformer.go`).

**Algorithm and paper.** Beeri & Ramakrishnan, "On the Power of Magic"
(the base transformation) — generalized adornment plus supplementary
predicates. This package does **not** implement Balbin, Port, Ramamohanarao
& Meenakshi's per-negative-literal program-segment treatment; negated IDB
atoms are adorned and given magic rules by the **same uniform mechanism**
as positive ones. That choice is not a simplification made for lack of
time — M2-M3-BUILD.md §5's own central claim is that this is *sound*:
allowedness already forces every negated atom's adornment to be all-bound
(every variable in a negated literal must be grounded before it is
evaluated), so demand-restricting `q^{b...b}` to exactly the magic-seeded
tuples decides exactly the membership questions the negation asks.
Completeness under negation holds by construction, given correct seeding
— which is exactly what treating negated atoms identically to positive
ones in the worklist achieves. See `src/transform/guard/DESIGN.md` for
where the actual verifying assertion lives.

**Where the paper leaves a choice open, this package picks the simplest
defensible one and says so, rather than inventing a cost model:**
`sips.go`'s ordering (see that file's own doc comment) is left-to-right in
source order with early-bound constraints pulled forward — not a
cost-based SIPS, per M2-M3-BUILD.md §3's explicit instruction not to
implement one.

## A real, measured consequence of taking the uniform-negation rule literally

`ancestor_nonancestor.dl`'s `nonancestor(x,y):-person(x),person(y),
!ancestor(x,y).` grounds **both** `x` and `y` (via the two `person` atoms)
before the negated literal is even reachable under SIPS (a negated atom
cannot be scheduled before its own variables are bound — there is no
choice here). The mechanical adornment computed for `ancestor` at that
occurrence is therefore `bb`, not the `bf` a reader familiar with the
hand-guarded `ancestor_nonancestor_guarded.dl` (which restricts only the
first argument) might expect by analogy. This is not a bug: it is the
textbook algorithm's literal, correct answer, pinned by
`TestAdornNegatedOccurrenceIsAllBound`. The measured consequence is real
and reported without softening in `docs/reports/m2-headline.md`: `dlc`'s
mechanical transform is markedly less effective than the hand guard on
this specific shape (and on `reachability_complement`/`p2.dl`, which share
the identical `person(x),person(y),!reach(x,y)`-shaped pattern) — the hand
guard's choice to restrict only one argument and leave the other free is a
*deliberate, non-mechanical* simplification a human author made
specifically to avoid this cost, not something a faithful demand-driven
adornment derives on its own. `culprit_cycle.dl` produces a program
Soufflé cannot stratify at all under this same mechanical adornment
(`Unable to stratify {magic_q_bf, magic_s_b, p_bf, q_bf, s_b, ...}`) —
this is the expected, designed-for outcome (the whole reason M3's guard
exists), not an M2 defect; `src/transform/guard/` is what decides what to
do about it.

## `adorn.go` — key decisions

**Occurrence-level, not just predicate-level, tracking.** The same
predicate can be adorned with genuinely different bindings at different
call sites within a single rule (a self-join) or across different rules —
`reachability_complement.dl` adorns `reach` as both `bb` (from the
negation in `unreach`) and `bf` (from `reach`'s own recursive rule, where
only the first argument is carried forward into the recursive call). Each
`(predicate, adornment)` pair is one worklist item and one set of output
relations; two different adornments of the same predicate never collide.

**EDB atoms are never pushed.** A relation is IDB (adornable) iff it is
the head of at least one clause, fact or rule — the exact same test
`sema/stratify.go`'s `buildPrecedenceGraph` uses. An `.input`-only
relation is read directly wherever it appears; magic sets only ever
restricts IDB predicates.

**Deterministic worklist.** `push` dedupes by `(pred, adornment)`; a
predicate's defining clauses are processed in source order (the order
`prog.Clauses` was walked when building `rulesByHead`, never Go map
iteration order). Two runs of the same input always produce
byte-identical output (CLAUDE.md's determinism rule).

**Worklist cap (10,000).** Blueprint failure mode #3 (adornment blowup).
`Adorn` returns an error rather than silently truncating or raising the
cap; the caller (`transformer.go`) surfaces it as a `transform_error`
through `dlc emit`.

## `rules.go` — the `V_i` projection, worked precisely

M2-M3-BUILD.md §4 gives `V_i = (vars bound after L₁..L_i) ∩ (vars in
L_{i+1}..L_n, or in t̄)`. Applied literally at face value to the checkpoint
*feeding into* literal `L_i`'s own magic rule (`sup_r_{i-1}`), this
undercounts: `sup_r_{i-1}`'s own "later" range would exclude `L_i` itself,
even though `L_i`'s own bound-position variables must be available in
`sup_r_{i-1}` for its magic rule to type-check at all. Re-deriving the
formula from first principles (index shifted by one, `j = i-1`) resolves
this: the "later" range for checkpoint `j` is `L_{j+1}..L_n`, and setting
`i = j+1` recovers exactly `L_i..L_n` — **inclusive of `L_i` itself**. This
package's `V[k]` (0-indexed, `V[k]` = the checkpoint *before* processing
`OrderedBody[k]`) is therefore computed as `BoundAfter[k] ∩ (vars in
OrderedBody[k:] ∪ vars(t̄))` — the slice `OrderedBody[k:]` starting *at* `k`,
not `k+1`. Verified both by direct trace (`ancestor_nonancestor`'s magic
rule for `!ancestor(x,y)` correctly carries both `x` and `y` forward, not
just `x`) and by the generated output passing `sema.CheckStratification`/
`CheckAllowedness` end to end (`TestGenerateOutputPassesEverySemaCheck`).
This is the single most important line in M2 per the task's own framing,
and the one place a fencepost error would have silently reintroduced
naive-evaluation cost with extra relations layered on top — exactly
blueprint failure mode #2.

**Naming.** `<pred>_<adornment>` for an adorned relation, `magic_<pred>_
<adornment>` for its magic predicate, `sup_<pred>_<adornment>_r<ruleIndex>_
<k>` for supplementary predicates — plain identifiers throughout, no `@`
(§1's naming rule; `TestGenerateNoAtNames`). Every new relation's `.decl`
uses positional column names (`c0`, `c1`, ...) since a supplementary
predicate's "columns" are a set of original-program variables gathered
from potentially several literals, with no single natural name of their
own.

**Two known simplifications, not yet exercised by the validation corpus,
disclosed rather than silently assumed correct:**
- A fact clause (`Body == nil`) for an adorned predicate is handled by the
  same `n=0` path as a normal rule (no supplementary chain beyond `sup_0`),
  but no shape in `BENCHMARK_FAMILY`/`CULPRIT_CANDIDATES` defines an IDB
  relation via a bare fact clause (all EDB data arrives through `.input`),
  so this path has not been exercised end-to-end against Soufflé.
- A constant appearing directly in a bound head position of a *non-query*
  adorned rule (e.g. a hypothetical `p(1,y):-foo(y).`) is handled
  structurally (`boundTermsAt` preserves whatever term is there, constant
  or variable) but likewise does not occur anywhere in the validation
  corpus.
