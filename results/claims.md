# Claims → evidence → strength → caveats

One row per substantive claim made anywhere in `README.md`,
`results/findings.md`, or `docs/`. "Strength" is honest, not
promotional: **measured** (direct measurement, real corpus or committed
run), **measured on constructed programs** (real measurement, but the
input was built to exhibit the property, not found in the wild),
**single data point** (one program, one configuration — no sweep), or
**inferred** (a conclusion drawn from measured facts, not itself a
directly measured number).

| Claim | Evidence | Strength | Caveats |
|---|---|---|---|
| Soufflé isolates negated relations but never demand-restricts them | `experiments/01-souffle-negation-behaviour.md`, confirmed unchanged on a later dev snapshot (`experiments/16-souffle-version-risk.md`) | Measured | Contradicts Soufflé's own documentation as literally worded; not reported upstream as a project finding, only cited. |
| `dlc`'s guarded transform reaches 46.0×–1,342.7× over Soufflé's own transform on `reachability_complement` | `experiments/49-demand-relaxation.md`, three-column table sourced from `measurements/m4-sips/` | Measured | One shape, one corpus family (`BENCHMARK_FAMILY`); `incl-sup` convention. |
| The mechanical transform closed from 194× worse to parity with a hand-written guard | `docs/reports` lineage → `experiments/49`; cited figures pulled directly from `experiments/48-guard-headline-pre-m4.md`'s pre-existing table | Measured | "Parity" is `excl-sup`, 252 vs. 285 — a 12% gap, called "matches" not "beats" deliberately. |
| `dlc` is worse than Soufflé's own transform on the positive fragment (0.49×) | `experiments/49-demand-relaxation.md` §5, `transitive_closure_bound`, 5 scale points | Measured | Stable across the 5 points measured; not swept beyond n=8,000. |
| The guard fires only on culprit-cycle-shaped programs, 11/16 | `bin/conecheck`'s committed JSON, re-derived for `experiments/57` | Measured on constructed programs | 16-program corpus is self-constructed (12 pre-existing + 4 built for task B), not sampled from real code. |
| The guard agrees with Soufflé's own stratifier, 7/7 where checked | `experiments/45-culprit-cycle-detection.md` | Measured | n=7; not the full 16-program corpus. |
| The guard's contribution shrinks with scale on the original cone constructions (1.75×→1.02×) | `experiments/55-mass-ratio-characterization-decomposition.md` | Measured on constructed programs | Root cause is a specific unscaled fixture parameter, not a property claimed to generalize — see next row. |
| The guard's contribution grows with scale when the transformed branch scales instead (2.14×→12.69×) | `experiments/56-mass-ratio-characterization-construction.md` | Measured on constructed programs | Single construction; magnitude at n=100 (12.69×) exceeded the pre-registered prediction (2×–5×) — reported as measured, not adjusted toward the prediction. |
| The culprit-cycle shape appears in 0/817 real-world Soufflé test files | `experiments/36-culprit-cycle-source-corpus.md` | Measured | Census tool is structural/regex-based, cross-checked against `dlc`'s real parser only for grammar compliance, not specifically re-verified for culprit-cycle detection at 817-file scale. |
| Zero-arity relations appear in 12/195 in-grammar Soufflé files | `experiments/32-verification-pass.md` (corrected from an initial 11) | Measured | Corpus-admissibility number, not a coverage or correctness claim — see `results/superseded.md`. |
| 19/195 Soufflé in-grammar files are strictly blueprint-grammar-compliant; 89/195 parse after the parenthesized-directive amendment | `experiments/50-four-reconciliations.md`, `experiments/38-grammar-amendment-optional-parens.md` | Measured | Answers "how much of a real corpus this grammar admits," not "how much was tested" — see `results/superseded.md`'s note on not conflating this with checking volume. |
| External program coverage of the naive/semi-naive differential gate: 5 programs, 16/16 points agree | `experiments/31-m1-front-end-and-evaluators.md` §4 item 3 | Measured | Small by design and disclosed as small; a coverage number, not a checking-volume number. |
| Total answer/verdict-identity checks against Soufflé across all corpora: 93 | Sum of four already-committed gates: 32 (`measurements/m3-5-headline-m4/summary.json`) + 9 (`measurements/punch-list/p1-gate/summary.json`) + 39 (`experiments/26-hostile-source-corpus.md`) + 13 (`experiments/28-souffle-diagnostic-catalogue.md`) | Measured | Checking-volume, not coverage — many repeated scale points on a small number of distinct shapes. |
| Q11 (ancestor-seed propagation shape) | `experiments/37-ancestor-seed-prediction-falsified.md` | Measured — **falsified** | Recorded as a failed prediction, not adjusted after the fact; the constructed "fix" (v2) is answer-wrong at every scale point tested. |
| Q12 (`p2.dl` predicted `T_dlc ≈ 300–700`) | `docs/OPEN_QUESTIONS.md`, measured 974/252 | Measured — **missed under both conventions** | Not recorded as correct under either convention; see `results/superseded.md`. |
| Q13 (growing-sibling ratio predicted 2×–5× by n=100) | `experiments/56-mass-ratio-characterization-construction.md` | Measured — **direction confirmed, magnitude underestimated** (12.69× measured) | Pre-registered before the construction was measured. |
| The `V_i` supplementary projection formula | `src/transform/magicset/DESIGN.md`, cross-checked against `experiments/43-magic-set-transform.md` | Inferred (derivation), verified by the M2 acceptance gate | An off-by-one in the first draft was caught before any generated code existed around it — see `docs/project-log.md`. |
| Demand relaxation is sound (relaxing a bound position only ever computes a superset) | `src/transform/guard/DESIGN.md`'s lemma; `guard.AssertNegationSeeding` checked in code | Inferred (proof) + measured (assertion never fires on the corpus run) | The proof is not machine-checked; the assertion is a runtime invariant check, not a formal verification. |
| No wall-clock claim appears anywhere in this project | `docs/03-methodology.md` | N/A (methodological commitment) | This is a stated scope limit, not a result — deliberately, given the hardware (WSL2, hybrid CPU, no PMU). |
