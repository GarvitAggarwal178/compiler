# Phase 0.6 report

Date: 2026-08-20. Lane B, run in parallel with M1 (Lane A, human-authored, not
started by this report). Blueprint amended to v1.2 (`docs/dlc-blueprint.md`).
Fixtures reused unchanged from Probe 0/0.5.

## 1. P4' — fixed hand-transform

`q2.csv` still byte-identical: **yes**, to both the original (buggy) P4 and to P2's
own answer in both configurations (`probe0.6-p4prime-run`, diffs recorded in
`docs/MEASUREMENTS.md`).

`@neglabel.reach` in the P2 magic-on profile: **`REC_T`, not `COPY_T`.** Its JSON
profile entry (`probe0-p2-on-run/prof_on.log`) carries its own
`non-recursive-rule` (`@neglabel.reach(x,y):-edge(x,y).`) and its own independent
20-iteration semi-naive fixpoint — a full private re-derivation of `reach` under a
new name, not an alias of an existing relation. `T_souffle` is not understated on
this account.

## 2. P2 three-column table

| `T_none` | `T_souffle` | `T_guard` |
|---|---|---|
| 40,030 | 26,465 | **231** |

(`reach`=26,404 + `unreach`=13,596 + `q2`=30; `@neglabel.reach`=26,404 +
`unreach.{bf}`=30 + `q2`=30 + seed=1; `m_reach`=1 + `reach_bf`=170 +
`unreach_bf`=30 + `q2`=30 — the last using P4', not the buggy P4.) `T_none/T_souffle`
≈ 1.5×, `T_souffle/T_guard` ≈ 114.6× — **the contribution is 114.6×, not 173×.**
Measurement IDs: `probe0-p2-off-extract`, `probe0-p2-on-extract`,
`probe0.6-p4prime-extract`.

## 3. P6 — counterexample attempt

**No counterexample found.** Five constructions, one bounded search:

| Attempt | Construction | Base | Correctly-seeded hand-transform | Result |
|---|---|---|---|---|
| Starting point (given) | Two negated occurrences of `reach`, seeds `{1,2}` collected | 2 | 2 | **Match** |
| 1 (derived binding, ∃) | `!reach(x,y)` where `x` ranges over `seed(x):-reach(1,x)` | 200 | 200 (both naive and fixpoint-correct) | Match, but **degenerate** — see below |
| 1b (derived binding, ∀) | Same shape, rewritten so a single wrong `x` changes the answer | 31 | naive: 200 / correct: **31** | **Naive diverges; fixpoint-correct matches** |
| 2 (incompatible bf/fb) | `!reach(1,y)` and `!reach(x,5)` on the same relation, independently adorned | 9 | 9 | **Match** |
| 3 (transitive dependency) | Negated relation transitively depends on its own negator | — | — | **Not constructible** (argued, not run) |

**Attempt 1 was degenerate, not a match on its own terms.** `blocked_from(y) :-
seed(x), node(y), !reach(x,y)` is existentially quantified over `seed` (≈170
members on this fixture); with that many candidate witnesses, almost every `y` fails
to be reached by *some* `x`, so `ans` saturates to the full 200-node domain
regardless of whether `reach_bf` was seeded correctly for `x≠1`. Both the naive
(under-seeded) and fixpoint-correct hand-transforms landed on the same 200-line
output as the baseline — this construction cannot distinguish correct from incorrect
seeding and is not evidence either way.

**Attempt 1b fixes this** by making `y` blocked only if *no* seed member reaches it
(`aux(y):-seed(x),reach_bf(x,y).` then `blocked_from(y):-node(y),!aux(y).`) — a
single incorrectly-omitted `x` now changes the answer. Result: the naive hand-transform
(seeds only `m_reach(1)`, never propagates `seed`'s derived values back into
`m_reach`) diverges hard (200 vs. 31 — confirms careless seeding really can be wrong).
The fixpoint-correct hand-transform (`m_reach(x) :- seed(x).`, closing the
seed↔reach_bf↔m_reach recursion before `aux`/`blocked_from` consume it) reproduces
the baseline exactly (31 = 31, `diff -q` clean). Stratification held throughout
(`{m_reach, reach_bf, seed}` is a positive SCC; `aux`/`blocked_from`/`ans` sit in
strata above it).

**Attempt 2** adorns the same relation two ways (`reach_bf`/`m_reach_bf` for the
`bf` pattern seeded at 1, `reach_fb`/`m_reach_fb` for the `fb` pattern seeded at 5,
the latter built over a reversed-edge helper relation `radj`) and combines both
negated conditions with AND. 9-tuple non-degenerate answer, exact match.

**Attempt 3 is not run — argued as structurally impossible.** For a negated
relation `X` to transitively depend on the relation `Y` whose rule contains `!X`,
stratification requires `Y` after `X` (Y consumes X negated) *and* `X` after `Y`
(X positively depends on Y) simultaneously — a cycle with a negative edge in it by
construction. That is condition 1's negation (a stratified source cannot have this
shape at all), not a clause-(b) question. No program can satisfy both "stratified"
and "negated relation transitively depends on its own negator" — reported as a
failed attempt with a proof, not run.

**Working hypothesis, per §6 v1.2:** clause (b) collapses into correct seed
collection (including fixpoint propagation through derived bindings, attempt 1b)
plus clause (a). No case tested here needed anything beyond that. M3 re-scopes to
~2 weeks per the blueprint's stated branch — this is a recommendation for the human
to confirm, not something this report enacts unilaterally.

Measurement IDs: `probe0.6-p6start-{base,hand}-run`, `probe0.6-p6a1-{base,hand-naive,
hand}-run`, `probe0.6-p6a1b-{base,hand-naive,hand}-run`, `probe0.6-p6a2-{base,hand}-
run`, all diffed via `diff -q`/`diff` directly against each pair's `ans.csv`.

## 4. Q5 — pre-registered corpus

**Predicate:** `harness/corpus_predicate.py` (`check_program`) — mechanical text
scan, no parser, no execution. Included iff (1) the program contains `!ident(` for
some `ident` not declared `.input` in the same file, and (2) at least one `.output`
relation has a defining clause (fact or rule) containing a numeric or string
literal. Applied via `harness/build_corpus.py`.

**Source:** `souffle-lang/souffle` tag `2.5`, commit
`5682a9f12e2668ecdd26348fe63cc508bc0fcf47` (matches the installed binary), `tests/`
subtree only, sparse-checked-out (not vendored — 612 `.dl`-bearing directories, too
large to commit wholesale). Full provenance: `tests/corpus/SOURCE.md`.

**Count: 36** (of 612 `.dl`-bearing directories across the entire `tests/` tree —
well above the ~15 floor, no reopening needed). List: `tests/corpus/
PREREGISTERED.txt`. Per-directory verdicts for all 612 candidates (why each was or
was not included): `tests/corpus/detail.json`.

**Disclosed, not hidden:** the predicate was first run against `tests/evaluation`
only — 11 included, under the ~15 floor. That scope was my own unrequested
narrowing, not what the directive asked for ("over Soufflé's tests/ tree"). Rerunning
the *same, unmodified* predicate over the full tree (fixing only which directories
count as candidates, not the inclusion logic) gave 36. Both runs are committed
(`probe0.6-q5-eval-only`, `probe0.6-q5-corpus`); the 36-count, full-tree run is the
one pre-registered in `tests/corpus/`.

No test in the corpus has been run. Nothing here scores accepted/total.

---

Four answers. Phase 1 (M1, Lane A) continues on its own timeline, unaffected by this
report per §8 v1.2.
