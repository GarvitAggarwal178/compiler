# Design history

`docs/02-design.md` describes the compiler as it stands today, with no
version deltas in its body. This document is the version history that used
to live inside it: `dlc-blueprint.md` went v1.0 → v1.5 with amendments
applied in place, and its founding premise was falsified on day one and
patched rather than rewritten. Reading the old file meant reconstructing
which parts still held; this table exists so nobody has to do that again.

Full irreversible-choice detail lives in `record/DECISIONS.md` (append-only,
one line per decision) — this table indexes it by design area rather than
duplicating it. For the narrative of *how* each of these was found, see
`docs/project-log.md`.

## Blueprint versions

| Version | Date | What changed | What prompted it | Established by |
|---|---|---|---|---|
| v1.0 | 2026-08-20 | Original design: `dlc` selected as the project; headline framed as "the guard," two-clause soundness condition; Probe 0's P1/P2/P3 specified. | Prior-art review; magic sets + negation identified as the one candidate with a strong external oracle, exact-integer metric, and non-vacuous headline. | `record/DECISIONS.md` (2026-08-20 rows); `docs/02-design.md` §4/§5 |
| v1.1 | 2026-08-20 | §2, §6, §7, §9, §11, §12 amended: differentiator restated (Soufflé restricts the *negating* relation but never the *negated* one — the founding premise "declines to transform anything touching negation" was wrong); guard clause (b) promoted to primary, clause (a) demoted to a side-condition; P1 replaced by P1′ (`.output path` was forcing full materialization, masking the ratio). | Probe 0's own measurement contradicted the founding premise on day one. | `experiments/01-souffle-negation-behaviour.md`, `experiments/02-output-forces-materialization.md` |
| v1.2 | 2026-08-20 | §6, §7, §8, §10 amended: P3 declared void (Soufflé's inliner removed the pivot relation and the recursive rule was independently dead) and replaced by P5, the corrected culprit-cycle program; headline metric changed from a pair to three columns (`T_none`/`T_souffle`/`T_guard`), contribution defined as `T_souffle/T_guard` rather than `T_none/T_guard` (the latter credits the guard with Soufflé's own transform's share). | Same probe run — P3's construction was defective independently of the guard question. | `experiments/03-completeness-counterexample-search.md` |
| v1.3 (implicit) | 2026-08-21 | Two-corpus split ruled: a correctness corpus (Soufflé's own `tests/`, in-grammar subset, scale-irrelevant) and a measurement corpus (scalable, external-shape benchmarks), never reported together. | NIGHT-BATCH-01 T2/T3's 31-program corpus had 3/31 clearing `T_none ≥ 1,000`; a recoverable-fraction ratio computed over it measured Soufflé's own unit tests being unit tests. | `experiments/13-corpus-split-ruling.md` |
| v1.4 | 2026-08-21 | P2-scale benchmark added: the P2 fixture family swept at `n ∈ {250..8000}`, with a permanent disclosure requirement that `T_none = n²` is definitional for this shape, not an observed property. | Wanted a scaling curve on a shape whose baseline growth rate was known and controllable. | `experiments/10-guard-scaling-first-measurement.md` |
| v1.5 | 2026-08-22 | §3 measurement-corpus adopted: OpenRuleBench route abandoned (no reachable rule files on either candidate source); replaced by `tests/corpus/BENCHMARK_FAMILY/`, five citation-traceable shapes with disclosed provenance per shape. | OpenRuleBench pre-registration blocked at the source-resolution stage. | `experiments/14-openrulebench-unobtainable.md`, `experiments/17-corpus-closed-version-risk-opened.md` |

## Post-blueprint architectural changes

The blueprint's own build-order section (§8) was superseded by
`specs/02-m1-build.md` before M1 started, and design changes after that
point were tracked as reports and decisions rather than blueprint version
bumps. The ones that changed what the compiler *does* (not just what was
measured):

| Change | What changed | What prompted it | Established by |
|---|---|---|---|
| Grammar amendment 1 | Zero-arity relations (`.decl foo()`) authorized beyond blueprint §4. | 11 (later corrected to 12) of 195 in-grammar Soufflé test files use them; blueprint §4's grammar would reject files the oracle itself accepts. | `specs/02-m1-build.md` §3.3; count correction in `experiments/32-verification-pass.md` |
| Grammar amendment 2 | Optional parentheses on `.input`/`.output` directive names. | Same motivation — corpus coverage without expanding the language beyond directive syntax. | `experiments/38-grammar-amendment-optional-parens.md` |
| `V_i` projection fix | The supplementary-chain projection's index arithmetic was re-derived from first principles after an off-by-one was found (a literal's own bound-position variables were being dropped from the checkpoint feeding its own magic rule). | Caught before any generated code existed around the wrong version. | `experiments/43-magic-set-transform.md`; full derivation in `src/transform/magicset/DESIGN.md` |
| Demand relaxation on negated occurrences | A bound position in a negated occurrence's adornment can be relaxed to free when its only binder is an unrestricted full-extent scan. | Measured cost of the naive `bb`-forced adornment: up to ~5,300× worse than a hand-written guard on one shape. | `experiments/49-demand-relaxation.md`; soundness lemma in `src/transform/guard/DESIGN.md` |
| Multi-query seeding | Adornment worklist seeded from every bindable query candidate (`FindQueries`), not just the first found in source order. | A program with two independent `.output` branches left the second one at full extent — the reason the guard's own practical contribution could not be demonstrated for most of a session. | `experiments/53-multi-query-seeding.md` |

## Corrections to previously reported numbers

Tracked in full, with what replaced them and why, in `results/superseded.md`
— not repeated here to avoid a third copy of the same list.
