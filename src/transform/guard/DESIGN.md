# package guard

M3. Transform-safety analysis: seed collection over negated occurrences
(`seeding.go`), culprit-cycle detection (`stratify.go`), the per-SCC
TRANSFORM/FALLBACK decision plus fallback cone (`decide.go`).

## The central technical claim, in three sentences (M2-M3-BUILD.md §5)

Allowedness requires every variable in a negated literal to be grounded
before it is evaluated, so the adornment `magicset` computes for a negated
atom is always all-bound. Therefore `magic_q^{b...b}` contains ground
tuples only, and demand-restricting `q^{b...b}` to exactly those tuples
decides exactly the membership questions the negation asks. Completeness
under negation holds by construction, given correct seeding — clause (b)
is not a separate check this package implements, it is a property of
`magicset` treating negated occurrences by the same uniform mechanism as
positive ones (never skipping them), verified here by an assertion, not
re-derived.

## `seeding.go` — the assertion, not a second algorithm

`AssertNegationAllBound` walks every negated occurrence `magicset.Adorn`
discovered and checks `Adorn.AllBound()`. There is no code path in
`magicset` that could produce a non-all-bound negated occurrence for an
allowedness-accepted program — the assertion exists to catch the day that
invariant breaks (an allowedness bug or an adornment bug, per §5's own
framing) loudly, at the exact point of violation, rather than downstream
as a silent wrong answer. `TestAssertNegationAllBoundFindsAtLeastOneOccurrence`
confirms the check inspects real occurrences, not an empty list.

## Counterexample search (M2-M3-BUILD.md §5, required)

`harness/night_m3_1_counterexample.py`: the 4 `tests/programs/p6*_base.dl`
constructions (Phase 0.6's original counterexample-hunt corpus) plus all 6
`tests/corpus/CULPRIT_CANDIDATES/` programs (NIGHT-BATCH-03 T4), `dlc`'s
real transform vs. the untransformed original, `harness/m2_accept.py`.

**10 attempted, 5 comparable-and-agreed, 5 correctly unstratifiable
(exactly the 5 programs NIGHT-BATCH-03 T4's structural classifier
predicted would trigger a culprit cycle), 0 disagreements.** Full
provenance and numbers: `docs/reports/m3-1-seeding.md`. This is the
counterexample search five bounded Phase 0.6 attempts already failed to
find — extending it with 6 more constructed programs specifically
designed to trigger the mechanism found no new one either.

Note on `p6*_base.dl` specifically: **all 4 are no-op pass-throughs**, not
a meaningful exercise of the real transform. `magicset.FindQuery` requires
the query rule's head to be directly the `.output` relation with a single
constant-bearing body atom; in every `p6*_base.dl` file the `.output`
relation's own rule either has more than one body atom
(`p6start_base.dl`: `ans(y):-target(y),orphan(y).`) or its single body
atom carries no constant of its own
(`p6a1_base.dl`/`p6a1b_base.dl`/`p6a2_base.dl`: `ans(y):-blocked_from(y).`
or similar) — the actual bound constant lives one or two projection steps
further in (e.g. `target(y):-reach(1,y).`). `FindQuery` does not chase
that chain, so all 4 report no bindable query and pass through unchanged
(confirmed directly: 0 occurrences of `magic_`/`sup_` in any of their
`dlc emit` output) — correctly, trivially answer-identical, but not a real
test of the mechanism. Disclosed as a known, literal-reading scope limit
of `FindQuery`, not generalized further here; the 6
`CULPRIT_CANDIDATES/` programs (all built to match the direct
`.output`-atom-with-constant pattern, NIGHT-BATCH-03 T4) are what actually
exercises the transform in this search.
