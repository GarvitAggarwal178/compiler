# Limitations

Ordered by how much a reviewer would push on it, not by when it was found.

1. **The shape the guard exists for was found in 0 of 817 real-world
   files.** A structural census over Soufflé's full test corpus (195
   in-grammar + 622 full tree) found the culprit-cycle shape — the guard's
   entire reason to exist — in zero files beyond the one program this
   project constructed to exhibit it. The guard's *correctness* is
   thoroughly demonstrated on constructed programs; its *necessity* on any
   known real-world program is not (`results/findings.md` item 4).

2. **No cost-based SIPS, by design.** The transform's literal ordering is
   fixed left-to-right, source order — no cardinality estimation, no cost
   model. `dlc`'s mechanical transform is beaten by a hand-written guard
   on every shape and scale point measured: up to several thousand-fold
   before demand relaxation existed, narrower (parity on demand
   restriction, a supplementary-chain-attributable gap on total
   materialization) after it, but never eliminated
   (`results/findings.md` item 1).

3. **Worse than Soufflé's own transform on the positive fragment.**
   `excl-sup`, isolating demand-restriction from implementation strategy:
   `dlc` is measurably worse than Soufflé on `transitive_closure_bound`
   before that isolation (0.48×–0.49×, stable across 5 scale points),
   because `dlc`'s supplementary chain materializes checkpoint relations
   Soufflé's transform does not generate. This is the cost of generality,
   named precisely, not smoothed into the headline convention
   (`results/findings.md` item 3).

4. **Two pre-registered predictions failed, and are recorded as
   failures.** Q11 predicted a specific propagation shape in a
   hand-written guard's recursion; tested exactly as written, the actual
   recursion was not of that shape, and the constructed "fix" was
   answer-wrong at every scale point. Q12 predicted `T_dlc ≈ 300–700` on
   one program; measured 974 (incl-sup) and 252 (excl-sup) — neither falls
   inside the stated range, a miss under both conventions in opposite
   directions, not a partial success under either
   (`docs/OPEN_QUESTIONS.md`, `results/superseded.md`).

5. **C codegen was never wired to the transform path.** The pipeline
   story (parse → transform → evaluate) is complete only for the
   tree-walking evaluator; the generated-C backend only ever sees an
   untransformed program.

6. **No wall-clock timing anywhere in this project.** Every number is a
   tuple count; the hardware (WSL2, hybrid CPU, no PMU) cannot support a
   timing claim, and none is made. This is a methodological choice stated
   up front, not an omission — but it does mean no claim here answers "is
   it actually faster."

7. **Structural/regex census tools, not real parsers, for corpus
   characterization.** Grammar-coverage and cone-metric census tools are
   regex- or structure-based, not `dlc`'s real parser. Always cross-checked
   against the real parser or a from-scratch Go implementation wherever a
   correctness claim depends on them (100% agreement on grammar
   compliance, 4/4 on cone-metric cross-check) — never trusted standalone.

8. **Single-query seeding — fixed, noted for the record.** For most of one
   session, the magic-set adornment worklist seeded from only the first
   bindable query candidate in source order, leaving a second independent
   `.output` branch at full extent. This was the reason the guard's own
   practical contribution could not be measured at all for that period —
   fixed by seeding from every candidate (`docs/design-history.md`,
   "Multi-query seeding"). Left here because it shaped how long it took
   to produce the guard-contribution numbers in `results/findings.md`,
   not because it is still live.
