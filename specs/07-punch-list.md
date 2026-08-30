# PUNCH-LIST — remaining work

Date: 2026-08-27. **This is a list, not a specification.** Five items, then the
presentation artifact, then done. Protocol as before: commit per item, report a
number, "what did not work" first.

---

## P1 — Multi-query seeding, then re-run task B

`magicset.FindQuery` seeds only the first bindable query candidate in source
order. Task B's constructions depend on a second `.output` branch being
demand-restricted while the first is declined; the limitation defeats the design,
which is why `T_guarded < T_none` held on 0/12 points.

**Change:** collect seeds from every bindable `.output` projection, not the
first. The adornment worklist already accepts multiple seeds — this is seed
collection, not algorithm.

**Gate:**
- Answer-identical on all six original programs plus all four B constructions.
  A change to seeding that changes an answer is a stop condition.
- Re-run B's measurement. Report `T_guarded` vs `T_none` per program per scale
  point.
- If `T_guarded < T_none` now holds anywhere, the guard's contribution is
  measured for the first time — report the margin.
- If it still holds nowhere, that is now a genuine finding about the guard, and
  §5/§6/§8 of FINAL.md can say so without the caveat.

Re-run the full 32-point headline afterward; multi-query seeding may change
numbers on shapes with more than one output.

---

## P2 — Consistency pass on FINAL.md

Three contradictions, all internal:

1. §4 "13 firing programs" vs §5 "10/16". Reconcile.
2. §5 blast-radius distribution `12×0 / 2×1 / 1×2` sums to 15 programs against a
   16-program corpus. Locate the missing one.
3. The p2 prediction: state the raw `T_dlc` integer and say explicitly whether
   56.9× is a reduction factor from 55,411 or `T_souffle/T_dlc`. Q12 predicted
   `T_dlc ≈ 300–700`; report predicted and measured side by side, no smoothing.

Then re-derive every ratio in §5 and §6 from the committed measurement JSON
rather than from earlier prose, and state that you did.

---

## P3 — `p2.dl` under excl-sup

One number already in the measurement data. If `T_dlc (excl-sup)` for `p2.dl`
lands near `p4prime.dl`'s 231, the demand relaxation reproduced the hand
transform exactly and the entire residual gap is supplementary accounting.

Report it, and the same figure for `reachability_complement` at n=250.

---

## P4 — Re-frame the counting convention in FINAL.md §6 item 3

`incl-sup` stays the headline. Supplementary predicates materialize real tuples
and are a real cost of the chosen strategy; excluding them to make a ratio larger
is choosing a convention after seeing the result.

`excl-sup` has exactly one job: isolating demand-restriction from implementation
strategy, which is what explains `transitive_closure_bound`. Use it there and
nowhere else. Delete the "understated relative to Soufflé's own accounting"
framing and the `~8,741×` figure from the findings section.

The replacement claim, which is stronger because both halves have a mechanism:

> `dlc` is measurably worse than Soufflé's own transform on the positive fragment
> (0.49×, stable across five scale points), because its supplementary chain
> materializes checkpoint relations Soufflé does not generate. It is better by
> orders of magnitude on stratified negation, because Soufflé does not
> demand-restrict negated relations at all.

---

## P5 — Two clarifications in FINAL.md

1. **Q8's mechanism (§7).** State *why* `dlc` beats v1, not just that it does.
   Hypothesis to confirm or refute against the emitted programs: v1 derives
   `nonancestor`'s restriction from `m_ancestor`'s propagated 50-member set,
   producing the 50 × 455 cross product measured in VERIFY-01 §V4, while `dlc`
   seeds `magic_nonancestor` from the query directly and lets `ancestor`
   propagate only where its recursion requires. If that is not the mechanism,
   state what is.
2. **Corpus size (§8).** The strictly blueprint-§4-compliant corpus is 19
   programs; 89/195 parse after the parens amendment. State which set the
   naive/semi-naive differential gate actually ran against, as a number.

---

## P6 — Presentation artifact

Per `m2 m3.md` §10 and `NIGHT-BATCH-04` G, unchanged. Only after P1–P5.

One static HTML file from a Python script in `harness/`, reading committed
measurement JSON plus `dlc explain` output. No server, no framework, no build
step. Two hours capped — ship what renders.

Four things visible:

1. The analyzer rejecting programs, four grounds, with spans.
2. Three-column metric per shape, `incl-sup` headline.
3. The guard firing and declining, reason named, cone shown.
4. The two findings with mechanisms: `bb`→`bf` demand relaxation before and
   after, and cone behaviour.

---

## Order

P1 first — it changes numbers P2 must reconcile. P2–P5 are a single documentation
pass and can be one commit. P6 last.

If P1 runs long, do P2–P5 anyway; FINAL.md is presentable today apart from the
three contradictions, and those are the only items that would embarrass anyone.