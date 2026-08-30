# M4-SIPS — demand relaxation on negated occurrences

Date: 2026-08-27. Two changes and one re-run. Small; do not expand it.

**Amends `docs/m2 m3.md` §3 and §5.** Nothing else in either document changes.

---

## 0. Why

`m3-headline.md` reports the hand transform beating `dlc` at every shape and
scale point, up to 5,317× on `reachability_complement`. The M2 report attributes
this to `bb`-vs-`bf` adornment and concludes a fix would require cost-based SIPS,
which §3 prohibits.

That conclusion is wrong. The fix is structural and sound, and it derives the
hand transforms mechanically.

---

## 1. The soundness lemma — put this in `guard/DESIGN.md`

`docs/m2 m3.md` §5 argues that a negated occurrence adorns all-`b`, and derives
completeness from that. **That argument is sufficient, not necessary**, and was
read as a requirement.

> **Lemma (relaxation).** For a negated occurrence `!q(t̄)` with adornment `α`,
> replacing any `b` with `f` is sound. A magic set with fewer bound positions
> demands a superset of the instantiations the negation queries. Completeness
> under negation requires the magic set to *cover* those instantiations, not to
> equal them. Computing more can never decide `!q(t̄)` incorrectly; only
> computing less can.

Worked instance, `p2.dl`: adorning `!reach(x,y)` as `bf` yields
`magic_reach_bf(1)`, demanding all of `reach(1,·)`, computed completely as 170
tuples. Deciding `(1,y) ∉ reach` against a complete `reach_bf(1,·)` is correct
for every `y`.

`AssertNegationAllBound` (§5's gate) must be **replaced**, not deleted. The new
invariant: every negated occurrence's adornment is all-`b` *before* relaxation,
and relaxation only ever turns `b` into `f`, never the reverse. Assert both.

---

## 2. The rule

> **A variable whose only binder in the SIPS prefix is an unrestricted
> full-extent scan carries no demand information.** Adorn its position `f`.

**Full-extent scan:** a positive body atom scheduled at a point where all of its
own argument positions are unbound. It enumerates its relation's entire extent
and restricts nothing downstream.

**Implementation:** track binding provenance per variable through the SIPS walk.
A variable is `restricting` if its binder is the magic atom, a query constant, or
a join with an already-restricting variable; `non-restricting` if its only binder
is a full-extent scan. At a negated occurrence, adorn `b` only at positions whose
variables are all restricting.

This is not a cost model. No cardinality estimate, no statistics, no benchmark
selection — a syntactic property of the SIPS prefix.

**Expected effect, all three affected shapes:**

| shape | `x` binder | `y` binder | derived adornment | hand transform uses |
|---|---|---|---|---|
| `p2` / `reachability_complement` | magic ← query constant | `node(y)`, full scan | `reach^bf` | `reach_bf` |
| `ancestor_nonancestor` | magic ← query constant | `person(y)`, full scan | `ancestor^bf` | `ancestor_bf` |
| `same_generation_negation` | magic ← query constant | full scan | `sg^bf` | `sg_bf` |

If the rule does not reproduce all three, it is wrong. Report which one fails.

---

## 3. Pre-register before measuring

Commit to `docs/OPEN_QUESTIONS.md` **before** implementing. Q11 just demonstrated
why this ordering matters.

> **Q12.** The mechanical `bb` adornment on negated occurrences forces a cross
> product in the supplementary chain: on `p2.dl`,
> `sup_reach_bb_r1_1(x,y,z) :- sup_reach_bb_r1_0(x,y), reach_bf(x,z)` is
> 200 × 170 ≈ 34,000 tuples against a total `T_dlc` of 55,411, and `y` is unused
> by `reach_bf`. Prediction, recorded before measurement: applying §2's
> relaxation collapses `reach` to a single `bf` adornment, eliminates the cross
> product, and brings `T_dlc` on `p2.dl` to within 3× of `p4prime.dl`'s 231 —
> i.e. `T_dlc ≈ 300–700`, a reduction of roughly 80–180× from 55,411. Same
> mechanism predicted on `reachability_complement` and `ancestor_nonancestor`;
> on the latter, `dlc` is predicted to approach v1's `T_guard` (25,500 at n=500)
> rather than its current 105,552.

---

## 4. Gate

1. `harness/m2_accept.py`, all 5 shapes + `p2.dl`, smallest scale point.
   **Answer-identical is a hard requirement** — a relaxation that changes an
   answer means the lemma was misapplied. Stop and report.
2. Emitted adornment sets for all six, before and after. Confirm the three
   shapes collapse to one adornment each.
3. Re-run `harness/night_m3_5_headline.py` in full: 32 points.
4. Predicted vs measured, both numbers, no smoothing. If the prediction is wrong,
   report it wrong and do not construct a variant to rescue it.

---

## 5. Second change — supplementary counting convention

`transitive_closure_bound` shows `T_dlc` = 207 vs `T_souffle` = 101, flat at all
five scale points, on a shape with no negation and one adornment. Constant, so
overhead, not scaling. `T_excl_copy` counts every supplementary predicate `dlc`
materializes; Soufflé does not generate an equivalent set.

**This systematically handicaps `dlc` against Soufflé on every shape**, which
means `same_generation_negation`'s reported 2,185× is understated.

Do not choose a convention. Report both, exactly as the project already reports
`excl-copy` / `incl-copy`:

- `T_dlc (incl-sup)` — current behaviour, every relation counted.
- `T_dlc (excl-sup)` — supplementary relations excluded, magic relations still
  counted.

Add both columns to `m3-headline.md`'s per-shape tables and to `m2_accept.py`'s
output. State which convention each contribution ratio uses.

**Gate:** the `transitive_closure_bound` open question closes with a number. If
`excl-sup` brings `T_dlc` to ≈ 101, the "anomaly" was the convention and should
be reported as such. If it does not, there is a real defect and it needs a root
cause.

---

## 6. What this does not touch

- No cost-based SIPS. No cardinality estimates. §3's prohibition stands.
- No changes to the guard, the cone, or fallback evaluation.
- No new corpus. No new benchmark shapes.
- No presentation work.

If this exceeds one session, ship §5 alone — the convention fix is independent
and closes an open question by itself.