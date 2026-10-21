# Final review — conservative checkpoint compaction

The implemented final pass is **`guarded-compact`**, composed after the existing
validated guarded transform. `guarded` remains a separate mode with its historical
identity. The source language has not grown. This maintenance and demo work was
AI-assisted; individual team identities and contributions must be supplied by the
team, as described in `docs/09-review2.md`.

## Exact change and invariants

`src/transform/compact/compact.go` selects checkpoint relations using explicit
roles produced by `magicset.AllocateNames` and rule generation. A source predicate
called `sup_*` is still a source predicate. Input/output declarations and all
source relations are protected. `Apply` copies the AST and public metadata before
editing and records removed definitions, consumers, source-rule origins and reasons.

Only two rewrites are legal:

1. An exact identity alias `A(v0,...,vn) :- B(v0,...,vn)` with distinct bare
   variables in the same order, equal schemas, A != B, one definition and no I/O.
   Equality of the complete relations permits replacing all body references,
   including negative references. The generated A declaration and definition are
   removed; source B and its I/O designations remain.
2. A non-fact checkpoint with one definition and exactly one positive use in a
   different clause, no negative uses or I/O, distinct bare head/call arguments,
   and every body variable in the head. Positive body wildcards disqualify it.
   Simultaneous variable substitution unfolds that one call, preserving every
   projected column. General head unification and arithmetic inversion are absent.

Eligibility conditions are checked against the current AST after each removal,
in stable declaration order.
Each step removes one relation; the finite declaration set bounds termination.
Changed bodies use `plan.Body`, and the resulting program is checked again for
schemas, allowedness and stratification. The guard's attempted candidate and final
program remain distinct in inspection. No unstratifiable candidate is executed.

Identity substitution and narrow positive unfolding have local semantic arguments;
these arguments do not prove the entire compiler correct. The pass does not remove
existential projection/dedup barriers or promise lower runtime or memory.

## Why the restriction matters

`fixtures/demo/projection-control.dl` uses `p(x,z) :- edge(x,y),node(z)` with many
witnesses y for the same x. Materializing the projected prefix prevents those
witnesses from multiplying downstream work. The final pass retains that barrier.
Locally measured Go semi-naive attempts are **162 guarded / 141 compact**, with all
20 requested rows equal under Go, native C and Soufflé. See
`measurements/final-2026-10-20/compact-final/summary.json` and its
`projection-control/` commands and outputs.

The supplied audit also recorded an aggressive variant at 2,021 attempts. That
was supplied counterexample evidence, not a production pass implemented or rerun
here. Its focused raw artifacts are retained under
`measurements/final-2026-10-20/supplied-projection-evidence/`. The integrated code
never enables that transformation.

## Observed materialization on one engine

All three columns below were measured on **Soufflé 2.5**, using the existing
`T_excl_copy` convention: exclude input loads and the documented copy category;
include remaining recorded supplementary checkpoints. The production pass
reproduced the conservative reference values. Every requested output matched.

| Configuration | Original | Guarded | Guarded compact |
|---|---:|---:|---:|
| Reachability complement n250 | 62,534 | 974 | 752 |
| Ancestor/nonancestor n500 | 250,450 | 6,102 | 6,101 |
| Same generation d4/b4 | 51,301 | 681 | 679 |
| Positive transitive closure n500 | 86,618 | 207 | 151 |
| Culprit cycle n20 | 286 | 286 | 286 |
| Growing sibling n20 | 584 | 273 | 245 |
| Growing sibling n50 | 2,660 | 417 | 361 |
| Growing sibling n100 | 8,577 | 676 | 571 |

Evidence: `measurements/final-2026-10-20/supplied-compaction/` and
`compact-final/`, each with exact commands, profiles, emitted programs,
outputs and a summary. The integrated harness also records source/fact hashes,
actual compaction actions and Go counters.

The one-tuple ancestor reduction is small. The culprit-cycle configuration is an
honest no-op. The positive control still exceeds Soufflé's own automatic transform
at this scale (151 versus its independently recorded 101 in
`measurements/review-2026-10-20/phase-a-oracle/positive_n500/`). These are adverse
or limited cases; the pass is not an always-beneficial optimizer.

## Work counters belong to their engine

Go's stored rule-derived insertion totals and candidate-head attempts are recorded
separately from Soufflé materialization. Soufflé may internally eliminate aliases:
on reachability n250, Go guarded stores 1,011 derived insertions while Soufflé's
profile records 974 under the convention above. The raw per-relation profiles show
that the latter omits the initial identity checkpoints and the aliased unreach
view. These engine totals are not interchangeable.

| Configuration | Go semi-naive guarded attempts | Go semi-naive compact attempts |
|---|---:|---:|
| Reachability n250 | 1,494 | 1,235 |
| Ancestor n500 | 11,752 | 11,051 |
| Same generation d4/b4 | 1,134 | 904 |
| Positive n500 | 315 | 257 |
| Culprit n20 | 443 | 443 |
| Growing n20 | 485 | 455 |
| Growing n50 | 642 | 584 |
| Growing n100 | 1,000 | 893 |

The same integrated summary also includes naive attempts and every relation's
insertions, including unfavorable baseline cases. None of these counters is an
elapsed-time, memory, or universal work guarantee. Generated C remains naive and
has no equivalent counter instrumentation.

## Acceptance and runnable review

The supplied final black-box script passed 29 checks. The supplied compaction
verifier passed all eight configurations and 24 deterministic graph cases against
independent reachability/complement answers; it compiled compact C for the eight
configurations and six graphs. The repository-owned harness adds original,
guarded and compact naive/semi-naive runs plus direct selected-mode C generation,
checks exact output manifests, and verifies the projection control with C/Soufflé.

The real Chromium workflow was rerun after phase C. All 14 browser groups passed,
including the earlier Review 2 route, compact mode, explicit skips, no-op behavior,
source-linked final graph selection and native compact execution. Raw submitted
snapshots, results, command outcomes, checks, and an actual compact screenshot are
under `measurements/final-2026-10-20/browser-release/`. Phase B was runnable and its
browser/repository gates passed before phase C began.

To demonstrate and reproduce:

```sh
make demo
# http://127.0.0.1:8080; select guarded or guarded-compact
make verify
make rejection
make oracle
make oracle-compact
bin/dlc inspect fixtures/demo/reachability.dl --transformer=guarded-compact
bin/dlc codegen fixtures/demo/reachability.dl /tmp/query.c --transformer=guarded-compact
cc -std=c11 -O2 /tmp/query.c -o /tmp/query
/tmp/query fixtures/demo/facts /tmp/query-out
```

If necessary prefix Make commands with `GOCACHE=/tmp/dlc-go-build-cache`.
`ORACLE_OUTPUT` and `COMPACT_OUTPUT` select fresh evidence directories; defaults
create unique run IDs. To repeat browser checks, use a Python interpreter with
Playwright and its Chromium installation outside this checkout:

```sh
python3 tests/demo/browser_smoke.py --url http://127.0.0.1:8080 --repo . \
  --output /tmp/dlc-final-browser-new --stage final
```

## Limits and remaining submission work

There are no unexplained oracle disagreements in the completed gates. Coverage is
finite: larger historical scales, a broad application-corpus study, other browser
engines, native Windows/MSVC, LeakSanitizer, controlled timing, peak memory and a
whole-compiler proof were not performed. Recursive arithmetic can be unbounded;
plain TSV has embedded-symbol-separator limits. Cost-based SIPS, parallelism,
more grammar and generated semi-naive C remain outside scope.

The final report still needs real team/guide/date fields and factual contribution
statements. Future optimization work would require new correctness and adverse-work
evidence. This implementation does not ship aggressive projection-removing inlining.
