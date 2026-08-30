# Methodology

## The oracle

**Soufflé 2.5**, pinned. Installed from the `x86_64-ubuntu-2404-souffle-2.5-
Linux.deb` release asset (`record/DECISIONS.md`, 2026-08-20). Every claim
of correctness in this project is a differential check against Soufflé:
same `.dl` source, same `.facts` input, **set equality on output
relations** — sorted, compared, symmetric difference reported, never a
text diff and never a boolean pass/fail alone.

Version drift was checked directly, not assumed away: Soufflé master
(`a1303be3`, 42 commits past the `2.5` tag) was cloned and re-run against
the same programs — the negation-isolation behaviour this project depends
on is unchanged (`experiments/16-souffle-version-risk.md`). A secondary
oracle, Jatalog (Java, semi-naive, stratified negation, no magic sets),
was considered as an independent cross-check on stratification decisions
specifically, but did not end up load-bearing in what's reported.

## The two corpora, and why they're never reported together

- **Correctness corpus**: Soufflé's own `tests/` directory, in-grammar
  subset (`tests/corpus/IN_GRAMMAR.txt`). Job: differential set-equality
  against the oracle, on real files this project didn't construct. Scale
  is irrelevant here — a 40-tuple program tests the parser, checker, and
  evaluator exactly as well as a 40-million-tuple one.
- **Measurement corpus**: `tests/corpus/BENCHMARK_FAMILY/`, five shapes
  swept across pre-registered scale points, each citation-traceable to a
  published source or explicitly disclosed as constructed. Job: a scaling
  curve with a controllable, known-shape reachable set.

The split exists because conflating them once already produced a
misleading number: an early corpus mixed differential-testing programs
with tiny fixed-size unit tests, and a "recoverable fraction" computed
over the mix measured Soufflé's own unit tests being unit tests, not what
a guard would recover (`experiments/13-corpus-split-ruling.md`). An
OpenRuleBench-based measurement corpus was attempted first and abandoned
— no reachable rule files on either candidate mirror
(`experiments/14-openrulebench-unobtainable.md`).

## Counting conventions, and why both are reported

Two independent axes, both resolved by *reporting both sides* rather than
picking the flattering one:

- **Copy-relation inclusion.** Some magic-set-transformed programs produce
  a relation that Soufflé's own JSON profile marks as a plain copy of
  another relation's output view (`COPY_T`), rather than a join-driven
  fixpoint derivation. Early measurements excluded these from
  `T_souffle`'s total; the convention was later fixed to report the sum
  both ways rather than defend the exclusion (`record/DECISIONS.md`,
  2026-08-20).
- **Supplementary-relation inclusion (`incl-sup` / `excl-sup`).**
  `dlc`'s own transform materializes supplementary checkpoint relations
  (`sup_*`) that Soufflé's transform has no equivalent of — they're part
  of what makes `dlc`'s transform general (it works for any program in the
  grammar, not one hand-derived shape). `incl-sup` (counting every
  relation, `dlc`'s default) is the headline convention everywhere in
  `results/findings.md`; `excl-sup` has exactly one job — isolating
  demand-restriction itself from implementation-strategy cost — and is
  used only where that isolation is the actual question
  (`transitive_closure_bound`'s ~0.49× figure, which `excl-sup` resolves
  to exactly 1.00×). Choosing a convention after seeing which one looks
  better is exactly the failure mode this project's own documentation
  discipline (`docs/OPEN_QUESTIONS.md` Q12, `results/superseded.md`) was
  written to catch.

## Metrics

**Exact derived-tuple counts, never wall-clock.** Every number is Σ over
relations of tuples inserted into Δ across all fixpoint iterations, for a
bound query — deterministic, hardware-independent. No wall-clock timing
appears anywhere in this project: the hardware (WSL2, hybrid CPU, no PMU)
cannot support a timing claim, and none is made
(`docs/02-design.md` §5).

## Pre-registration

Fixture generators are seeded from a constant recorded in the fixture
file; two runs of the same command must produce byte-identical output.
Scale points, corpus subdirectories, and — for constructed test cases —
the predicted result are committed *before* the measurement runs, not
chosen after seeing it. Two concrete instances: `docs/OPEN_QUESTIONS.md`
Q13 predicted a 2×–5× ratio growth for a not-yet-built construction before
it was measured (measured 12.69×, `experiments/56-mass-ratio-
characterization-construction.md`); the corpus-admissibility predicate
used for grammar-coverage census numbers is the same mechanical regex
across every run it's cited from, corrected in place when found wrong
rather than re-tuned toward a result (`docs/design-history.md`).
