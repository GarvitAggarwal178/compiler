# Open questions

Append-only. Things noticed and not acted on, with the date noticed.

## 2026-08-21 — `.type` support deferred to M4, not adopted now

68% of `.type` declarations account for grammar exclusions found by T5
(`docs/reports/night01-T5-grammar.md`). Corpus ruling (`docs/reports/
corpus-ruling-2026-08-21.md` §4.2) considered and rejected adding `.type` support
now: it would grow the correctness corpus from ~195 to ~400 programs (195 already
exceeds what M1 can consume), buys nothing for the measurement corpus (moving to
OpenRuleBench, §2.2), and the Filter 4 gain is marginal (`number`/`symbol` typing
already rejects programs; subtypes add roughly one rejection ground). Cost is 2-3
days of M1, coming directly out of M3 where the result lives. **M4 candidate,
revisit only if OpenRuleBench integration fails** and the Soufflé suite has to
carry measurement after all.

## 2026-08-20 — pre-registered corpus contains negative (rejection) tests

T2 (NIGHT-BATCH-01, `docs/reports/night01-T2-corpus.md`) found 4 of 36
pre-registered programs (`semantic/plan1`, `semantic/plan2`,
`semantic/rel_stratification`, `semantic/witness_check`) are Soufflé's own
negative tests — deliberately invalid programs used to test Soufflé's error
diagnostics, not executable programs. `corpus_predicate.py`'s mechanical predicate
(negated IDB literal + constant-bearing `.output`) has no way to distinguish these
from real programs. Not fixed (NIGHT-BATCH-01 prohibition #2, predicate frozen for
the batch's duration). Consider for daylight: either exclude programs whose
`.err`/`.out` sidecar files indicate an expected failure, or accept the ~11%
contamination rate as a known corpus property and report coverage as
`accepted / (total - known-negative)`.

## 2026-08-20 — `corpus_predicate.py` non-determinism in a diagnostic field only

`check_program()` iterates `output_names` (a Python `set`) and records the first
`.output` relation found to carry a literal as `matched_output_relation`; when
multiple qualify, which one is recorded depends on Python's per-process randomized
string-hash seed. Found by T1 (NIGHT-BATCH-01) re-running `probe0.6-q5-corpus` and
diffing `tests/corpus/detail.json` against the committed version
(`docs/reports/night01-T1-audit.md`). **`tests/corpus/PREREGISTERED.txt` and the
36-count are unaffected** — inclusion only needs "at least one" match, which is
order-independent. Not fixed: NIGHT-BATCH-01 §0.2.2 forbids editing the predicate
that produced the pre-registered corpus, without exception, for the batch's
duration. Fix when unlocked: iterate `sorted(output_names)` instead of the raw set,
or drop the field, or run with `PYTHONHASHSEED=0`.

**Resolved 2026-08-21** (corpus ruling §4.3, prohibition lifted narrowly for this
purpose): `check_program()` now iterates `sorted(output_names)`. Verified: 3
consecutive runs of `harness/build_corpus.py` produce byte-identical
`tests/corpus/detail.json` (measurement `corpus-ruling-t4.3-verify`); the 36-row
`tests/corpus/PREREGISTERED.txt` data content is unchanged from the pre-fix
version (only a superseding header comment was added, separately).

## 2026-08-20 — Q3 (blueprint §10), NIGHT-BATCH-01 T8: no artifact found

"Does *Extended Magic for Negation* (~2019) ship a downloadable artifact?"

**Citation confirmed:** K. Tuncay Tekle and Yanhong A. Liu (Stony Brook University),
*Extended Magic for Negation: Efficient Demand-Driven Evaluation of Stratified
Datalog with Precise Complexity Guarantees*, ICLP 2019.
[arXiv:1909.08246](https://arxiv.org/abs/1909.08246). The paper "presents a simple
extension to demand transformation to support stratified negation, and a simple
extension to an optimal bottom-up evaluation method" — existence and scope only,
per the Prior Art Register rule (`docs/dlc-blueprint.md` §11: cite, do not read
implementation source; not applicable here regardless, since none was found).

**No downloadable or installable artifact found.** Two targeted web searches (paper
title + "artifact"/"GitHub"; authors' names + "implementation"/"github") and one
fetch of Yanhong Liu's Stony Brook faculty page (which lists her other Datalog
tooling but nothing tied to this paper) turned up only the paper itself (arXiv,
DeepAI mirror, ResearchGate, DOAJ) — no repository, no release, no download link.

**Filter 1 verdict: unchanged.** `docs/dlc-blueprint.md` §11's "cite-and-avoid,
artifact status unresolved" becomes "cite-and-avoid, no artifact exists" —
resolved in the same direction the blueprint already assumed, not escalated
(NIGHT-BATCH-01 §T8: escalate only if an installable artifact is found; none was).

## 2026-08-20 — Q5 (blueprint §10), resolved

"Which Soufflé `tests/` subdirectory is the pre-registered corpus?" — **The whole
`tests/` tree, filtered by a mechanical predicate, not a hand-picked subdirectory.**
36 of 612 `.dl`-bearing directories qualify (`harness/corpus_predicate.py` via
`harness/build_corpus.py`, `tests/corpus/PREREGISTERED.txt`,
`docs/reports/probe0_6.md`). No test in it has been run yet — that's M1/M2/M3's job
as each becomes relevant, not Phase 0.6's.

## 2026-08-20 — P6: no counterexample to the (a)/(b) independence found (bounded search)

Five constructions, one ~1hr session (`docs/reports/probe0_6.md` §3): the given
starting point, a derived-binding case (two variants, one degenerate), an
incompatible-bf/fb-pattern case, and a structural argument that the fourth planned
shape can't be stratified at all. None produced a program where correctly-seeded
(including fixpoint-propagated) restriction of a negated relation gave a wrong
answer. Working hypothesis per blueprint v1.2 §6: clause (b) collapses into correct
seed collection + clause (a), M3 re-scopes to ~2 weeks. **Not enacted by this report**
— it's a recommendation for the human, and the search was bounded (five
constructions), not exhaustive. If M3 finds a sixth construction that breaks this,
the collapse hypothesis is wrong and the two-week re-scoping decision needs revisiting
before M3's schedule is finalized around it.

## 2026-08-20 — Q1 (blueprint §10), empirical answer

"Does Soufflé's magic transform still decline negation, empirically?" — **No, not on
either negation probe tested.** P2: `unreach` (body contains `!reach`) was
magic-transformed, `unreach.total` 13,596 → 30, answer bit-identical
(`probe0-p2-off-extract`, `probe0-p2-on-extract`, `probe0-p2-diff`). P3: `p` (body
contains `!s`) was magic-transformed with a real magic seed
(`@magic.p.{bf}(1)`), answer bit-identical (`probe0-p3-on-profile`,
`probe0-p3-diff`). See `docs/reports/probe0.md` STOP section — this is escalated, not
resolved here.

## 2026-08-20 — `--inline-exclude=q`: demoted from experiment to M3 prerequisite

Per the Phase 0.5 ruling, this is no longer "the cheapest experiment to run next" —
it's a stated prerequisite of the P5 culprit-cycle program (`docs/dlc-blueprint.md`
§12, v1.1): P5's `q` is self-recursive (`q(x,y):-q(x,z),base(z,y).`), which the
Probe 0 evidence suggests Soufflé's inliner would not collapse the way it collapsed
P3's non-recursive `q(x,y):-base(x,y).` — but this is not yet verified, only
argued. Verify it when P5 is run (M3), not before. Note from Soufflé's docs:
`--magic-transform-exclude` already implies `--inline-exclude` for the named
relations, so P5 may not need the flag explicitly if `--magic-transform-exclude` is
used instead.

## 2026-08-20 — Q2 (blueprint §10), empirical answer

"What is the blast radius — one relation, or the whole SCC?" — **Zero relations
declined on P3.** `p` was transformed. `s` was relabeled (`@neglabel.s`) but computed
identically to the untransformed run. `q` never materializes in either run (see next
entry). Nothing in P3 was skipped by the magic-transform pass. See
`docs/reports/probe0.md` STOP section.

## 2026-08-20 — metric definition: what counts as a "derived" tuple across a copy

P1's magic-on run produces `path` as a `COPY_T` (not `NREC_T`/`REC_T`) of
`@interm_out.path.{ff}` (`probe0-p1-on-profile`). I excluded copies from the
`T_magic` sum on the grounds that a rename isn't a fixpoint-driven join, and said so
in `docs/MEASUREMENTS.md`. This needs a real, stated definition before M2's
headline metric depends on it — "derived tuple" should be pinned to something in the
semi-naive Δ-rewrite itself (Lane A), not to whatever Soufflé happens to label
`COPY_T` in a given build.

## 2026-08-20 — Soufflé inlines pass-through IDB relations before magic-transform runs

`q(x,y):-base(x,y)` in P3 never appears as a materialized relation in either the
magic-off or magic-on profile; post-transform rule bodies reference `base` directly
(`probe0-p3-on-profile`). This means any culprit-cycle argument that runs *through* a
relation like this (blueprint §6a's `magic_q →¬ s → q → magic_q`) can be silently
defused by Soufflé's own optimizer before the magic-transform pass ever sees the
program — independent of whether the guard concept is sound. `souffle` has
`--inline-exclude=<relations>` (confirmed via `souffle --help`, not yet used). Needs
resolving before any culprit-cycle program is added to a corpus: either use
`--inline-exclude` on Soufflé's baseline runs so both engines see the same IR, or
construct corpus programs where the negated-dependency chain can't be trivially
inlined away (e.g. give `q` a second use that blocks inlining).

## 2026-08-20 — Phase 0.5 resolution of the Q1/Q2 escalation

The human ruling on `docs/reports/probe0.md` reinterprets both entries above rather
than treating them as kills: Soufflé transforms the *negating* relation but never
demand-restricts the *negated* one (`@neglabel.<rel>` isolates and fully
materializes it). P4 (`docs/reports/probe0_5.md`) demonstrates by hand that
restricting the negated relation is both sound and large (170 vs 26,404 on the
negated relation) on this program. Blueprint bumped to v1.1
(`docs/dlc-blueprint.md`) with the differentiator restated around this. Version-drift
was checked and rejected as an explanation: Soufflé 2.5 (2025-03-25) is current and
its documentation still states the blanket-skip behaviour that the observed behaviour
contradicts.

## 2026-08-20 — metric definition: resolved as "report both conventions"

Superseding the entry above (kept, append-only): rather than pick `excl-copy` or
`incl-copy`, `docs/MEASUREMENTS.md` now reports both for any row containing a
`COPY_T` relation. P1' (v1.1) turns out to have zero `COPY_T` relations in its
magic-on run at all — removing `.output path` didn't just make the copy irrelevant,
it made Soufflé never compute the unrestricted relation the copy would have been
made from. The two conventions only diverge on P1 (v1.0, superseded) — supporting
evidence that the defect was `.output path`, not the transform in general.

## 2026-08-20 — P2's fixture was not built with a bounded-reachability core

The Phase 0.5 directive predicted `reach_bf ≈ 50` for P4 by analogy with P1's
engineered fixture (`gen_p1_graph`, core size 50). P2's fixture
(`build_p2_fixture`) is a plain random digraph with no such construction; measured
`reach_bf.total = 170` (BFS-cross-checked, `docs/MEASUREMENTS.md`
`probe0.5-p2-bfs-check`). Not acted on: no fixture change needed since the check P4
was built for (sound + large reduction on a negated relation) still holds at 170.
Worth remembering if a future probe wants a *specific* reachable-set size on a
negation program — reuse `gen_p1_graph`'s core/rest construction, not
`build_p2_fixture`'s.

## 2026-08-20 — no known program where Soufflé's negation-transform selectivity is wrong

Both negation probes here (P2, P3) have the negated relation depend on nothing the
magic seed restricts — Soufflé's silent selectivity happens to agree with a
guard-shaped notion of "safe" in both cases tested. Filter-1/thesis defense needs at
least one candidate program where that selectivity is either wrong (unsound) or
needlessly conservative (declines something safe); none has been found or
constructed yet. This is a design question for whichever corpus subdirectory Q5
resolves to, not something to invent inside Phase 0.

## 2026-08-26 — IN_GRAMMAR.txt admits more out-of-grammar features than the zero-arity gap already found

NIGHT-BATCH-02 T8 (`docs/reports/night02-T8-grammar-census.md`) found 11 of the 195
`IN_GRAMMAR.txt` files use zero-arity relations, which blueprint §4's grammar (as
written then) did not admit -- flagged, not fixed, since zero-arity is now
explicitly admitted by M1-BUILD.md §3.3's amendment. Running the real `dlc` lexer
over the same 195 files (M1 §3.1, `docs/SESSION_LOG.md` 2026-08-26) found a larger
problem: 42/195 files produce lex-error tokens, the large majority from
aggregates (`sum`/`max`/`min`/`count`-style syntax), records
(`semantic/type_system_records{,2}`, 240 and 24 error tokens respectively -- by far
the worst offenders), `#include` preprocessor directives, and Soufflé pragma
directives (`.printsize`, source-location pragmas) -- all explicitly out of
blueprint §4's grammar ("No functors, no aggregates, no components, no records").
NIGHT-BATCH-01 T5's mechanical text-scan predicate did not catch any of these.
**Practical consequence:** M1-BUILD.md §3.3's gate one ("all 195 in-grammar files
parse with zero errors") is very likely unachievable as literally stated without
first correcting `IN_GRAMMAR.txt` itself -- a human decision (`IN_GRAMMAR.txt` is
not under M1-BUILD.md §5's edit prohibition, but fixing the mechanical predicate
that built it is scope beyond a single M1 work item). Not acted on here; the
lex-coverage gate that actually governs §3.1 is panics-only and passed cleanly
(0/234, `measurements/m1-3.1-lex-coverage-summary.json`) regardless of this.

## 2026-08-26 — `.input`/`.output NAME()` (trailing parens) is the dominant cause of §3.3 gate one's shortfall, not a parser bug

M1 §3.3's parser gate one ("all 195 in-grammar files parse with zero errors")
measured 20/195, far below what §3.1's lex-coverage finding alone predicted
(`docs/reports/m1-progress.md` has the full breakdown). Root cause of the largest
single bucket (105/175 failures, 60%): real Soufflé test files overwhelmingly write
`.input Name()` / `.output Name()` with cosmetic empty parentheses after the
relation name -- e.g. `evaluation/access1/access1.dl`'s `.output Low()`. Blueprint
§4's grammar says `.input ident` / `.output ident`, no parens, and M1-BUILD.md
§3.3 authorizes exactly one amendment (the term-list-in-an-atom one) -- not this.
`dlc` therefore rejects the parenthesized form as specified, correctly, per the
grammar as actually authorized. **This is not a bug and was not silently
"fixed"** by adding an unauthorized second amendment -- flagged here instead,
since a human may want to authorize it explicitly (the idiom is cosmetic,
near-universal in real Souffle code, and costs nothing semantically to admit).
Remaining causes of the 175 failures, all independently confirmed as genuinely
out-of-grammar rather than parser bugs: 15 files use `unsigned`/`float` types
(2-type grammar is number/symbol only); the remaining ~55 use aggregates
(`sum w : ...`), functors (`cat(...)`, `strlen(...)`, `range(...)`, etc. --
blueprint's own words: "No functors, no aggregates"), or the pragma/`#include`
directives already flagged in the entry above. Not acted on beyond reporting;
gate two (round-trip) and gate three (hostile-corpus oracle agreement) are
unaffected and both passed cleanly on everything that did parse.

## 2026-08-27 — Q11, pre-registered before measurement (NIGHT-BATCH-03 T5)

**Q11.** `ancestor_nonancestor_guarded.dl` propagates `m_ancestor` across a
recursion whose bound argument is invariant, and derives `nonancestor`'s
restriction from `m_ancestor` rather than seeding it from the query.
Prediction, recorded before measurement: a variant seeding `m_nonancestor(1).`
directly gives `T_guard ≈ 3,200` at n=500 and `≈ 18,200` at n=8,000, i.e.
`T_souffle/T_guard ≈ 32×` and `≈ 1,400×` respectively, moving this shape into
the same band as `reachability_complement` and `same_generation_negation` and
dissolving Q8.

## 2026-08-27 — Q12, pre-registered before implementation (NIGHT-BATCH-04 / M4-SIPS)

**Q12.** The mechanical `bb` adornment on negated occurrences (M2's own
finding, `docs/reports/m2-headline.md`) forces a cross product in the
supplementary chain: on `p2.dl`,
`sup_reach_bb_r1_1(x,y,z) :- sup_reach_bb_r1_0(x,y), reach_bf(x,z)` is
200 × 170 ≈ 34,000 tuples against a total `T_dlc` of 55,411, and `y` is
unused by `reach_bf`. Prediction, recorded before implementing
`M4 sips.md` §2's demand-relaxation rule: applying the relaxation
collapses `reach` to a single `bf` adornment, eliminates the cross
product, and brings `T_dlc` on `p2.dl` to within 3× of `p4prime.dl`'s
231 — i.e. `T_dlc ≈ 300–700`, a reduction of roughly 80–180× from
55,411. Same mechanism predicted on `reachability_complement` (currently
two adornments, `reach_bb`+`reach_bf`, per
`measurements/m4-sips/before/reachability_complement.transformed.dl`)
and `ancestor_nonancestor` (currently `ancestor_bb`+`ancestor_bf` per
`measurements/m4-sips/before/ancestor_nonancestor.transformed.dl`); on
the latter, `dlc` is predicted to approach v1's `T_guard` (25,500 at
n=500) rather than its current 105,552 (`docs/reports/m2-headline.md`).
`same_generation_negation` currently has `sg_bb`+`sg_bf`
(`measurements/m4-sips/before/same_generation_negation.transformed.dl`)
and is predicted to collapse to `sg_bf` alone by the identical
mechanism. If the rule does not reproduce all three collapses, report
which one fails rather than adjusting the rule to rescue the prediction
(`M4 sips.md` §2: "If the rule does not reproduce all three, it is
wrong").

Answer set-equality (`harness/m2_accept.py`, all 5 `BENCHMARK_FAMILY`
shapes + `p2.dl`, smallest scale point) is a hard gate on this change,
checked before any of the above numbers are trusted: a relaxation that
changes an answer means the lemma (`M4 sips.md` §1) was misapplied.

## 2026-08-27 — Q8, CLOSED (NIGHT-BATCH-04 D4 / A's ancestor comparison)

**Q8** asked whether `ancestor_nonancestor`'s modest measured contribution
band (4×–62×, `docs/reports/night02-T5-guarded.md`) is inherent to the
shape's recursive structure, or a sign that `ancestor_nonancestor_guarded.dl`
(v1)'s hand-written guard is itself suboptimal. T5 (NIGHT-BATCH-03) tested
one specific hypothesis (Q11: a first-argument-invariant-style seed) and
found it unsound (answer-wrong), leaving Q8 explicitly unresolved --
"Q8 stands unchanged" (`docs/reports/night03-T5-ancestor-seed.md`).

**Resolved by `docs/reports/m4-sips.md`'s direct comparison**, same
fixture (n=500), same methodology, both SOUND (answer-identical, unlike
Q11's v2): v1's hand guard measures `T_guard = 25,500`
(`measurements/m4-sips/v1_check/`, exactly reproducing the number already
cited above). `dlc`'s mechanical transform, after `M4 sips.md` §2's
demand relaxation (a sound rule, not a rescue construction), measures
`T_dlc = 6,102` (incl-sup) / `2,752` (excl-sup) at the same point --
**4.2×–9.3× below v1, not "near" it.** Per NIGHT-BATCH-04 A's own
pre-stated decision rule ("if it lands well below, v1 is suboptimal after
all"): **v1 is suboptimal.** The modest 4×–62× band was not inherent to
`ancestor`'s recursion shape; it was a cost of v1's choice to propagate
`m_ancestor` across the recursion (the same choice Q11 already flagged)
rather than seed `nonancestor`'s restriction directly the way `dlc`'s
correct mechanical SIPS relaxation now does. `dlc` vs v1's full-scale
comparison (n=500..8,000: `T_souffle/T_dlc` grows from 16.6× to 887.8×)
is in `m4-sips.md` section "Q8, closed" and its headline table.
