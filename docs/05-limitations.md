> Current 20 October scope: arithmetic-cycle scheduling, deterministic diagnostics,
> signed query literals, strict CLI arguments and byte/LF answer comparisons were
> repaired. The local workbench and separate conservative guarded-compact pass are
> implemented. Historic experiments retain their original scope; current evidence
> and tradeoffs are in `docs/09-review2.md` and `docs/10-final-review.md`.

# Limitations

Ordered by how much a reviewer would push on it, not by when it was found.

1. **The shape the guard exists for was found in 0 of 817 upstream test
   files in the historic census.** A structural census over Soufflé's full test corpus (195
   in-grammar + 622 full tree) found the culprit-cycle shape — the guard's
   entire reason to exist — in zero files beyond the one program this
   project constructed to exhibit it. The guard's checked behavior agrees on
   the constructed cases; this is not a formal proof for all programs. Its
   *necessity* on a known application workload is not demonstrated (`results/findings.md` item 4).

2. **No cost-based SIPS, by design.** The shared planner preserves stable source
   scheduling where executable and lowers stalled arithmetic columns. There is no
   cardinality estimation or cost model. The following hand-guard comparison is
   historical and concerns the original guarded mode. `dlc`'s mechanical transform is beaten by a hand-written guard
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
   (`record/OPEN_QUESTIONS.md`, `results/superseded.md`).

5. **The C backend uses naive evaluation.** Guarded optimization is now wired
   to C generation and covered by backend/oracle checks; generated semi-naive
   execution and cost-based join planning remain unimplemented.

6. **No controlled runtime or memory result.** The performance evidence consists
   of materializations and candidate-head attempts. This project has not collected
   a controlled timing or peak-memory benchmark. WSL2 and
   missing PMU counters do not make ordinary elapsed-time benchmarks impossible;
   they require careful controls and reported variability — but it does mean no claim here answers "is
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

9. **Compaction is restricted and can be a no-op.** Source/I/O relations,
   shared non-alias checkpoints, negative non-alias uses, repeated/constant call
   arguments, positive wildcards and existential projections are protected.
   The selected positive control still materializes more than Soufflé's automatic
   transform; the culprit-cycle point has no reduction. No runtime benefit is
   established by the smaller counts in `docs/10-final-review.md`.

10. **The demo is bounded and local.** Its temporary snapshot, source/facts/output
    and subprocess limits are documented in `docs/08-demo.md`. Timeouts are explicit
    results. Chromium/Linux was exercised; other browser engines and native
    Windows/MSVC were not. No tuple derivation witnesses, global evaluation clock,
    or generated semi-naive C are available. Plain TSV does not unambiguously encode
    embedded symbol tabs/LF; output comparisons concern exact serialized row sets
    within that domain. Unicode U+2028 is preserved as data.
