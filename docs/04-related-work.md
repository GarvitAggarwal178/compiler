# Related work

**Read — these are papers, not code:**

- **Beeri & Ramakrishnan** — magic sets, the base transformation this
  project's adornment/SIPS/supplementary-predicate machinery implements.
- **Chen (1997), *Magic Sets and Stratified Databases*** — culprit cycles
  cause unstratification; describes a labeling algorithm claimed simpler
  and more efficient than Balbin et al., on the grounds that Balbin's
  analysis of the abnormal behaviours is incomplete.
- **Balbin, Port, Ramamohanarao & Meenakshi**, *JLP* 11(3–4):295–344, 1991
  — the canonical algorithm for magic sets under negation; introduces a
  less restrictive allowedness definition than the naive one; evaluates
  negative literals via per-literal program segments under an extra
  control mechanism, rather than uniform magic rules the way this project
  does.
- **Ross**, *Modular stratification and magic sets for Datalog programs
  with negation*, JACM.
- **Behrend**, *Soft stratification for magic set based query evaluation*,
  PODS 2003.
- ***Extended Magic for Negation*** (~2019) — considered as substitution
  risk; artifact status unresolved (no downloadable implementation found).

**Cite, not substitutable:**

- **Soufflé** — this project's oracle. Its own documentation states that
  relations with negation in their body, or in the body of a dependency,
  are not transformed. Observed behaviour in Soufflé 2.5 (and confirmed
  unchanged on a later development snapshot,
  `experiments/16-souffle-version-risk.md`) contradicts the documentation
  as literally stated: the negation-*bearing* relation **is** transformed;
  only the negat*ed* relation is left fully materialized behind
  `@neglabel.<rel>`. This documentation/behaviour discrepancy is noted as
  a low-priority side observation, not pursued as a project finding.
- **`travitch/datalog`** (Haskell) — `MagicSets.hs`'s own comments state
  that negated literals can break stratification and therefore decline the
  transform, and the author is explicitly unsure whether the restriction
  should cover only negated literals or everything defining them. That
  stated uncertainty is this project's own research question, from an
  independent engine's author.
- **Jatalog** — Java, semi-naive, stratified negation, no magic sets.
  Considered as a secondary stratification cross-check; not load-bearing
  in what's reported.
- **DLV** — disjunctive Datalog with magic-set extensions; different
  fragment.
- **RecStep / DDlog / FlowLog** — Datalog on other substrates; out of
  scope.

## What this project does not claim

**Not** "we extended magic sets to negation." This is not a
reimplementation of Balbin et al.'s algorithm, nor an attempt to match its
generality. `dlc` implements a **detector and a fallback**: a static check
for when magic-set restriction of a negated relation is safe, and a
per-relation fallback to full materialization when it is not — layered on
top of the same uniform magic-rule mechanism used for positive atoms,
which is a narrower design than Balbin et al.'s dedicated per-literal
negation-handling segments. The claim defended is: *the transform is
applied under a stated soundness guard; here is where the guard fires,
here is what it is conservative on, and here is the derived-tuple
reduction where it does not fire* — not a general theory of magic sets
under negation.
