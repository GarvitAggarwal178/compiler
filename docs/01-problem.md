# The problem

Magic sets rewrite a Datalog program so that a bottom-up evaluator only
derives the facts a top-level query actually demands, instead of computing
every relation's full extent. For a query like "is node `y` reachable from
node `1`," a naive evaluator computes reachability from every node; a
magic-set-transformed one computes it from `1` only. On a graph with a
small reachable set, the difference is orders of magnitude in the number
of tuples derived — the compiler's whole reason to have an optimization
pass at all.

**The specific gap this project measures:** Soufflé — a mature, widely
used Datalog engine, and this project's own oracle — applies its
magic-set optimizer to a relation whose body contains a negated literal,
but never demand-restricts the relation *inside* that negation. It
isolates the negated relation under a renamed label (`@neglabel.<rel>`)
and computes it in full, every time, regardless of what the query actually
demands. This discharges completeness-under-negation by brute force: the
negated relation can't be incomplete if it's never restricted. It also
forfeits every reduction reachable through a negated literal, on every
program where the query happens to touch one.

The gap is real and measured, not asserted: on `reachability_complement`
(`experiments/49-demand-relaxation.md`), Soufflé's own transform buys
0.75×–0.82× against no transform at all — barely anything — while `dlc`'s
guarded transform, once it correctly relaxes demand on the negated
occurrence, reaches 46.0×–1,342.7× over Soufflé's own number at the same
scale points. The three-column comparison (`T_none` / `T_souffle` /
`T_dlc`) behind that ratio, across every measured shape, is in
`results/findings.md` and `results/claims.md`.

**Why this isn't just "restrict everything":** restricting a negated
relation's extent can make a stratified program unstratifiable — a
negative cycle through the magic-seed/supplementary chain that didn't
exist in the source. `docs/02-design.md` §4 describes the two-clause
guard this project builds to detect that case and fall back safely when
it fires. The soundness argument, not the transform itself, is the
project's actual contribution — the transform is textbook.
