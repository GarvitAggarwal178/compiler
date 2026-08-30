# Experiments — index

Every experiment record in this directory, renamed by content from its
original chronological/task-ID filename (`docs/rename-map.csv` has the
full mapping). Numbered chronologically by when it was produced (three
documented exceptions — see the map's `reason` column). A reader should be
able to find the experiment behind any claim in `results/claims.md` from
this table alone.

**Status** column: `current` — its numbers stand as reported;
`superseded` — see `results/superseded.md` for what replaced it and why
(the file itself is left unedited, per append-only discipline for
reports); `historical` — a session wrapper or narrative record, not a
measurement.

| # | File | Question | Answer | Status |
|---|---|---|---|---|
| 01 | `01-souffle-negation-behaviour.md` | Does Soufflé decline to transform anything touching negation, as documented? | No — it transforms the negation-bearing relation, never demand-restricts the negated one | current |
| 02 | `02-output-forces-materialization.md` | Why did P1 (v1.0) measure a ~1× ratio instead of the predicted ~10³? | An unnecessary `.output path` forced full materialization regardless of the transform | superseded (P1 replaced by P1′) |
| 03 | `03-completeness-counterexample-search.md` | Does guard clause (b) collapse into correct seeding, or is a separate completeness check needed? | No counterexample found after a bounded search; working hypothesis confirmed | current |
| 04 | `04-night01-session-log.md` | — | Raw append-only line log, NIGHT-BATCH-01 | historical |
| 05 | `05-provenance-and-determinism-audit.md` | Does every measurement directory reproduce byte-identically? | 61/61 did, after 8 provenance gaps were found and backfilled | current |
| 06 | `06-corpus-scale-survey.md` | What fraction of a 31-program corpus clears meaningful scale? | 3/31 clear `T_none ≥ 1,000` — too small-scale to measure recovery on | superseded (led to the corpus split, see `13`) |
| 07 | `07-recoverable-envelope-sweep.md` | What's the recoverable-fraction envelope across the initial corpus? | Median 0.026 — measuring unit tests being unit tests | superseded (see `13`) |
| 08 | `08-whole-tree-exploratory-sweep.md` | What does a whole-tree exploratory sweep of Soufflé's tests/ look like? | Baseline characterization, informed later corpus decisions | historical |
| 09 | `09-grammar-coverage-census.md` | How much of Soufflé's grammar does blueprint §4 admit? | Initial grammar-coverage numbers | superseded (corrected in `32`) |
| 10 | `10-guard-scaling-first-measurement.md` | Does the P2-scale benchmark show Θ(n) contribution growth? | Yes, `T_souffle/T_guard = Θ(n)` on this shape | current |
| 11 | `11-m1-harness-buildout.md` | — | M1 harness build-out, groundwork for later diff/gate scripts | historical |
| 12 | `12-night01-session-summary.md` | — | NIGHT-BATCH-01 morning summary | historical |
| 13 | `13-corpus-split-ruling.md` | Should correctness and measurement corpora be reported together? | No — they answer different questions; split permanently | current |
| 14 | `14-openrulebench-unobtainable.md` | Is OpenRuleBench usable as the measurement corpus? | No reachable rule files on either candidate source | current (negative result) |
| 15 | `15-souffle-subsumption-bug.md` | Does Soufflé subsumption interact with magic-transform correctly? | A real pre-2.5-fix bug existed, confirmed already fixed upstream post-2.5; outside this project's grammar entirely | current (disclosure, not a project finding) |
| 16 | `16-souffle-version-risk.md` | Does Soufflé master still refuse to demand-restrict negated relations? | Yes, unchanged 42 commits past the 2.5 tag | current |
| 17 | `17-corpus-closed-version-risk-opened.md` | Is the corpus decision final as of this point? | Corpus closed; version risk opened as the next open question | historical (ruling) |
| 18 | `18-doop-abandoned.md` | Is DOOP usable as an alternative benchmark source? | No — input-resolution failed on two independent paths within a 3-hour cap | current (negative result) |
| 19 | `19-allowedness-probe.md` | Does a first allowedness-check draft behave as expected? | Initial probe, informed the M1 implementation | historical |
| 20 | `20-m1-gates-from-harness.md` | Are M1's harness-derived gates satisfied? | Yes, at the point measured | current |
| 21 | `21-allowedness-derivation.md` | Is allowedness correctly derived per Balbin et al.'s definition? | Yes, cross-checked | current |
| 22 | `22-benchmark-family-fixture-generation.md` | Do the five BENCHMARK_FAMILY fixture generators produce valid, seeded output? | Yes | current |
| 23 | `23-baseline-sweep.md` | What are `T_none`/`T_souffle`/`E_recoverable` at baseline, pre-guard? | Baseline numbers established for later comparison | current |
| 24 | `24-hand-guard-benchmark-family.md` | Does a hand-written guard demonstrate the recoverable reduction on BENCHMARK_FAMILY? | Yes — this is the `T_guard` baseline later compared against `dlc`'s own transform | current |
| 25 | `25-cross-shape-scaling-analysis.md` | Does the Θ(n) result generalize across shapes, or is it specific to one? | Mechanism, not just shape, confirmed across multiple constructions | current |
| 26 | `26-hostile-source-corpus.md` | Does `dlc`'s parser agree with Soufflé's own accept/reject verdicts on hostile input? | 35/39 agree; 2 expected semantic-gap disagreements, 1 documented lexer divergence — all 39 accounted for | current |
| 27 | `27-grammar-usage-census.md` | What grammar constructs does `IN_GRAMMAR.txt` actually use? | Initial census, including the zero-arity count later corrected (`32`) | superseded (zero-arity count) |
| 28 | `28-souffle-diagnostic-catalogue.md` | Does `dlc`'s rejection diagnosis match Soufflé's own for the 13 rejection cases? | 13/13 consistent | current |
| 29 | `29-cone-collapse-observed.md` | Is the fallback cone always empty in naturally-arising culprit-cycle programs? | Yes, in every program measured before deliberate construction (`51`) | current |
| 30 | `30-night02-session-summary.md` | — | NIGHT-BATCH-02 morning summary | historical |
| 31 | `31-m1-front-end-and-evaluators.md` | Is M1 (front end + naive/semi-naive evaluators) complete and gated? | Yes — full report, including the wsl.exe tooling artifact note | current |
| 32 | `32-verification-pass.md` | Do the project's own corpus/grammar numbers survive independent re-verification? | Several corrected (zero-arity 11→12, `IN_GRAMMAR.txt` 195→19 strict), one Q closed with a confirmed mechanism (Q8) | current (corrections applied elsewhere, see `results/superseded.md`) |
| 33 | `33-printer-output-souffle-parseable.md` | Does `dlc`'s pretty-printer output re-parse cleanly on Soufflé? | 17/26 accepted, 17/17 answer-identical among comparable | current |
| 34 | `34-m2-acceptance-harness.md` | Does the M2 acceptance harness correctly compare against Soufflé reference transforms? | Yes | current |
| 35 | `35-m3-measurement-protocol.md` | Does the M3 measurement protocol work end-to-end on the identity transform? | Yes | current |
| 36 | `36-culprit-cycle-source-corpus.md` | How prevalent is the culprit-cycle shape in real-world Soufflé source? | 0/817 files (195 in-grammar + 622 full tree) | current |
| 37 | `37-ancestor-seed-prediction-falsified.md` | Does the hand guard v1 propagate a seed the way Q11 predicted? | No — falsified; the constructed "fix" is answer-wrong at every scale point | current (falsified prediction, disclosed) |
| 38 | `38-grammar-amendment-optional-parens.md` | Should `.input`/`.output Name()` parens be authorized? | Yes — 89/195 in-grammar files then parse | current |
| 39 | `39-corrected-grammar-predicate.md` | Does the corrected grammar-compliance predicate match the real parser's verdict? | Yes, 100%, after three corrective passes | current |
| 40 | `40-codegen-symbol-ordering-gap.md` | Does the C codegen path have a symbol-ordering gap? | Yes, identified and documented | current |
| 41 | `41-fallback-cone-metric.md` | Is the fallback cone metric implementation correct? | Yes, cross-checked 4/4 against an independent implementation | current |
| 42 | `42-night03-session-summary.md` | — | NIGHT-BATCH-03 summary | historical |
| 43 | `43-magic-set-transform.md` | Does the magic-set transform (adornment, SIPS, supplementary chains) work correctly? | Yes, after the `V_i` off-by-one was caught and fixed pre-release | current |
| 44 | `44-seed-collection-negated-occurrences.md` | Is seed collection over negated occurrences correct (M3 §5)? | Yes | current |
| 45 | `45-culprit-cycle-detection.md` | Does culprit-cycle detection agree with Soufflé's own stratifier? | Yes, 7/7 where checked | current |
| 46 | `46-per-scc-decision-and-fallback-cone.md` | Is the per-SCC TRANSFORM/FALLBACK decision (M3 §7) correct? | Yes | current |
| 47 | `47-fallback-evaluation.md` | Does mixed (partially transformed, partially fallback) evaluation work (M3 §8)? | Yes, same evaluator, no re-plumbing needed | current |
| 48 | `48-guard-headline-pre-m4.md` | What is the guard's contribution before demand relaxation? | Sub-1×–16× on `reachability_complement`/`ancestor_nonancestor`; source of the 194×/5,317×-worse-than-hand-guard figures | current (baseline for `49`'s comparison) |
| 49 | `49-demand-relaxation.md` | What does demand relaxation on negated occurrences change? | Contribution jumps to 46×–1,343× / 17×–888× on two shapes; closes the mechanical-vs-hand-guard gap to parity on one | current — **the project's central result** |
| 50 | `50-four-reconciliations.md` | Do four previously-noted inconsistencies resolve? | Yes — including the strict-grammar corpus count fixed to 19/19 | current |
| 51 | `51-cone-corpus-superseded.md` | Is the fallback cone ever non-empty, and does it flip `T_guarded < T_none`? | Cone construction succeeded, but this report's own `T_guarded < T_none` conclusion (0/12) was later found to be a seeding bug | **superseded** — see its own header and `53` |
| 52 | `52-dlc-explain-rejection-transform-guard-modes.md` | Does `dlc explain` correctly report rejection/transform/guard decisions? | Yes, one sample per rejection ground plus transform/guard cases | current |
| 53 | `53-multi-query-seeding.md` | Does fixing `FindQuery`→`FindQueries` change any program's answer? | No (hard gate passed) — and it flips `T_guarded < T_none` from 0/12 to 9/12 | current |
| 54 | `54-presentation-artifact-build.md` | Does the static presentation artifact build correctly from committed JSON? | Yes | current |
| 55 | `55-mass-ratio-characterization-decomposition.md` | Why does the guard's margin shrink with scale on the original constructions? | Declined portion grows by design; transformed portion shrinks due to an unscaled fixture parameter | current |
| 56 | `56-mass-ratio-characterization-construction.md` | Does the margin grow instead, if the transformed branch is made to scale? | Yes — 2.14×→12.69×, confirming the mechanism on the opposite construction | current |
| 57 | `57-final-report-whole-project-synthesis.md` | What does the whole project add up to? | Whole-project synthesis; source for `results/findings.md`, `results/claims.md`, `results/superseded.md` | current (synthesis, superseded in form by `results/`, not in content) |
