# NIGHT-BATCH-03 T5 — the `ancestor_nonancestor` seed prediction: pre-registered, then tested

Date: 2026-08-27. Order followed exactly: Q11 committed
(`docs/OPEN_QUESTIONS.md`, commit `caa4c0f`) before any measurement ran.

## What did not work

**The prediction's stated premise does not match the actual committed
rule, and this was caught before building v2, not after.** Q11 (quoting
V4/Q8's own diagnosis) says `ancestor(x,y):-ancestor(x,z),parent(z,y)` is
"first-argument-invariant, exactly like `reach` in `p4prime.dl`."
The **actual** rule in `tests/corpus/BENCHMARK_FAMILY/ancestor_nonancestor.dl`
is:

```
ancestor(x,y) :- parent(x,y).
ancestor(x,y) :- parent(x,z), ancestor(z,y).
```

This is a different shape than the one quoted: the recursive atom is
`ancestor(z,y)` — its bound position is `z`, not `x`. Contrast
`p4prime.dl`'s actual `reach_bf`:

```
reach_bf(x,y) :- m_reach(x), reach_bf(x,z), edge(z,y).
```

Here `x` is genuinely invariant — the recursive atom is `reach_bf(x,z)`,
same `x` as the head, extending only at the far end (`y`). `ancestor`'s
recursive atom is not of this shape: `x` is not carried into the recursive
call at all, `z` is. Given this, a mechanically-correct adornment of
`ancestor`'s actual rule is **forced** to propagate the seed along `parent`
edges (exactly what `ancestor_nonancestor_guarded.dl` (v1) already does) —
v1's construction was not a suboptimal choice; the recursion's own shape
requires it. This was verified by inspection before building v2 (not
inferred from v2's result), and is disclosed here as a defect in the
prediction's premise, not a defect in v1.

## What was built, exactly as pre-registered

Per instruction, the prediction was tested as written, not revised:
`ancestor_nonancestor_guarded_v2.dl` seeds `m_nonancestor(1).` as a single
static fact (mirroring `p4prime.dl`'s `m_reach(1).` literally) with **no**
propagation rule, applied to `ancestor`'s actual recursive structure:

```
.decl m_nonancestor(a:number)
m_nonancestor(1).

.decl ancestor_bf2(x:number, y:number)
ancestor_bf2(x,y) :- m_nonancestor(x), parent(x,y).
ancestor_bf2(x,y) :- m_nonancestor(x), parent(x,z), ancestor_bf2(z,y).
```

Given the premise mismatch above, this construction is expected to be
**unsound**: `m_nonancestor` only ever contains `{1}` (no propagation
rule feeds it), so `ancestor_bf2`'s second rule requires
`m_nonancestor(z)` for `z` reached after the first `parent`-hop — never
true. `ancestor_bf2` can therefore only ever capture direct
`parent(1,y)` facts, missing every indirect ancestor. The original file
(v1) and `p4prime.dl` were both left unmodified — v1 stays committed, v2
is a new, separate file.

## The numbers

Provenance: `measurements/night03-t5/summary.json`,
`measurements/night03-t5/<tag>/{orig_vs_v2,v1_vs_v2}/{original,candidate}/{cmd.txt,stdout.txt,prof.log}`.
Method: `harness/m2_accept.py` (T3's validated harness), reused unmodified.

| `n` | `T_none` | `T_guard` (v1) | `T_guard` (v2) | v2 answers == baseline? | v2 answers == v1? |
|---|---|---|---|---|---|
| 500 | 250,450 | 25,500 | 996 | **no** | no |
| 1,000 | 1,000,950 | 51,000 | 1,996 | **no** | no |
| 2,000 | 4,001,950 | 102,000 | 3,994 | **no** | no |
| 4,000 | 16,003,950 | 204,000 | 7,992 | **no** | no |
| 8,000 | 64,007,950 | 408,000 | 15,996 | **no** | no |

**Predicted vs. measured, stated as both numbers, no smoothing:** Q11
predicted `T_guard ≈ 3,200` at n=500 and `≈ 18,200` at n=8,000. Measured:
996 at n=500, 15,996 at n=8,000 — the predicted *magnitude* is roughly in
range (same order at both ends), but the prediction is **wrong on the
dimension that actually matters**: it implicitly assumed v2 would be a
*correct* (answer-preserving) transform whose tuple count could be
meaningfully compared to v1's. It is not correct. `q_nonancestor` diverges
from the untransformed baseline at **every** scale point — sampled diff at
n=4000: `candidate_only` includes person ids `1, 10, 11, 13, 14, ...`
(real ancestors of node 1, wrongly reported as non-ancestors because
`ancestor_bf2` never computed anything past depth 1), `original_only` is
empty (v2 never reports a *spurious* ancestor — it only ever
*under*-reports `ancestor`, consistent with the missing-propagation
mechanism exactly as predicted structurally).

## Reading this the required way

**Q11 is wrong, and it is wrong for a specific, verified reason: its
premise misidentified which recursive shape `ancestor_nonancestor.dl`
actually uses.** Per instruction, no third variant was constructed to
rescue the ratio, and the prediction is not adjusted after the fact. The
smaller `T_guard` v2 shows is not evidence of a better guard — it is
exactly what an incomplete, under-derived relation looks like: cheaper
because it computed less than the correct answer, not because it computed
the correct answer more efficiently. **Q8 is not dissolved.**
`ancestor_nonancestor`'s modest 4×–62× contribution
(`docs/reports/night02-T5-guarded.md`) stands as previously measured; v1
remains the correct hand guard for this shape.

This result is not treated as an M2-M3-BUILD §13 stop condition (that
clause governs `dlc`'s own transform's answer-equality, M2-M3-BUILD §4/§13
item 1) — T5 is explicitly a pre-registered exploratory prediction test,
and its own instructions treat a wrong prediction, including an unsound
construction, as a reportable result, not a batch-halting defect. It is,
however, directly relevant to M2's real transform: it is a concrete,
measured demonstration of why the general adornment algorithm (§2 of
`docs/m2 m3.md`) must derive propagation mechanically from the recursive
rule's actual variable-binding structure, never from a shape assumed by
analogy to a different relation — exactly the failure this file's v2
naively falls into.

## What a skeptic attacks first

- Only one "literal" v2 construction was tried. A cleverer sound
  construction that still avoids full propagation might exist (e.g.
  restructuring `ancestor`'s recursion to an equivalent invariant form by
  hardcoding the query constant into a rule head) — not attempted, since
  T5 explicitly prohibits constructing a rescue variant, and v1's own
  header comment already states hardcoding the query constant into a rule
  head is the practice this family's guards specifically avoid.
- The "predicted magnitude roughly in range" observation above is
  arithmetic curiosity, not a finding — reported for completeness, but the
  answer divergence makes the tuple count itself meaningless as a
  transform-quality metric here (a wrong program with fewer tuples is not
  an improvement).

## Verdict

**T5: DONE.** Prediction tested exactly as pre-registered — **WRONG**,
for a verified structural reason (premise misidentified the recursion
shape), not a smoothed-over near-miss. v2 diverges from both the
untransformed baseline and v1 at all 5 scale points. Q8 stands unchanged;
Q11 is closed as falsified. `ancestor_nonancestor_guarded_v2.dl` committed
alongside v1 (neither modified, both reported), rows appended to
`docs/MEASUREMENTS.md`.
