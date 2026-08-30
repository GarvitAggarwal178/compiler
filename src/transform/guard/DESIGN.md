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

## `stratify.go` — M3.2, culprit-cycle detection (clause a)

**Cheap precondition.** `HasPositiveCycle` is a small, dedicated
positive-edges-only precedence graph built directly from `prog.Clauses`
(not reused from `sema/stratify.go`'s `buildPrecedenceGraph`, whose
internals are unexported and which mixes both edge polarities) — O(V+E),
checked before any transform work, per M2-M3-BUILD.md §6's own framing.

**The actual check.** `CheckCulpritCycle` builds the candidate transformed
program via `magicset.Adorn`+`Generate` and runs `sema.CheckStratification`
— the pre-existing SOURCE stratifier, applied to the TRANSFORMED AST, per
§6 ("closes the gap the `Transformer` interface documented: `strata`
reflects the pre-transform precedence graph"). A no-bindable-query program
is trivially stratifiable (the transform is a no-op).

**Differential oracle, not self-consistency.** `harness/
night_m3_2_culprit_detection.py` prints the candidate transformed program
(`dlc emit --transformer=magicset`), asks `dlc check` for `dlc`'s own
verdict, and separately feeds the identical text to real Soufflé —
Soufflé's stratification checker was written by someone else and is asked
the same question about the same program. **7/7 agreed** (all 6
`CULPRIT_CANDIDATES` programs plus `culprit_cycle.dl` itself), 0
disagreements — full numbers in `experiments/45-culprit-cycle-detection.md`.

## `decide.go` — M3.3, per-SCC decision and the fallback cone

**The decision.** `Decide` reuses `CheckCulpritCycle`; if the fully-
transformed program is unstratifiable, every original predicate with a
generated relation inside an unstratifiable SCC (`UnstratifiableSCCs`,
mapped back via `RelationOrigin`) is a **culprit**. `ConeClosure` computes
the downward dependency closure of the culprit set over the SOURCE
precedence graph, following the full (positive+negative) edge relation —
`declined = culprit ∪ cone`. Cross-checked exactly against `harness/
cone_metric.py`'s own already-validated NIGHT-BATCH-03 T9 result on
`culprit_cycle.dl` with `{p}` declined: **both compute `{q, s}`, exactly**
(`TestConeClosureMatchesHarnessCulpritCycle`) — the required "dlc's own
cone computation and the harness's must agree exactly" gate.

**Building the final mixed program.** `magicset.GenerateMixed` (a
declined-aware variant of `Generate`, sharing all of `rules.go`'s
machinery) skips a declined predicate's entire adorned/magic/supplementary
apparatus and emits its ORIGINAL clauses instead; any occurrence inside a
still-TRANSFORM'd predicate's rule that targets a declined predicate is
left referencing the original, unrenamed atom (`occ.atom` already carries
the correct name and terms — "declined" is implemented as *not rewriting*,
not as a separate code path). This is the mechanism M2-M3-BUILD.md §8
depends on: a mixed program's untransformed relations read their full,
correctly-materialized extent because they are read by their own original
name, the same name §3.9's evaluator (or Soufflé) already computes in
full wherever else the program uses it.

## M3.3's actual measured outcome, stated plainly

On `culprit_cycle.dl` (and every one of NIGHT-BATCH-03 T4's 5 structurally-
matching `CULPRIT_CANDIDATES` programs), the unstratifiable SCC of the
transformed program is **not** confined to one relation — it entangles
`p`'s, `q`'s, and `s`'s own adorned/magic/supplementary relations together
(they were mutually dependent through the negative cycle to begin with).
The culprit set computed directly from that SCC is therefore already
`{p, q, s}` — **all** of the shape's IDB predicates — and the cone adds
nothing further (`ConeClosure` returns empty on top of it). The guard's
own numbers, honestly reported (`experiments/46-per-scc-decision-and-fallback-cone.md`): on every
one of these 6 programs, `T_guarded == T_none` exactly — **zero
contribution over the untransformed baseline**, because the whole program
falls back. This is blueprint failure mode #1's own stated risk
("if the guard declines everything, the project has no contribution"),
observed narrowly and specifically on the culprit-cycle-shaped subset of
the corpus, not universally: the same guard, unmodified, produces the
identical benefit `magicset`'s own transform already delivers on every one
of the 5 `BENCHMARK_FAMILY` shapes that do NOT trigger a culprit cycle
(confirmed byte-identical `dlc emit` output, `--transformer=guarded` vs.
`--transformer=magicset`, on all 4 applicable shapes) and on the one
negative control (`cc_edb_negated.dl`, 870→54, unaffected by the guard).
The guard is not vacuous in general — it is exactly as generous as it can
safely be, and it happens that this particular family of hand-constructed
culprit-cycle shapes offers it nothing to save.

## Counterexample search (M2-M3-BUILD.md §5, required)

`harness/night_m3_1_counterexample.py`: the 4 `tests/programs/p6*_base.dl`
constructions (Phase 0.6's original counterexample-hunt corpus) plus all 6
`tests/corpus/CULPRIT_CANDIDATES/` programs (NIGHT-BATCH-03 T4), `dlc`'s
real transform vs. the untransformed original, `harness/m2_accept.py`.

**10 attempted, 5 comparable-and-agreed, 5 correctly unstratifiable
(exactly the 5 programs NIGHT-BATCH-03 T4's structural classifier
predicted would trigger a culprit cycle), 0 disagreements.** Full
provenance and numbers: `experiments/44-seed-collection-negated-occurrences.md`. This is the
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
