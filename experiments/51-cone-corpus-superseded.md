> **SUPERSEDED (2026-08-30) — this report's own headline finding no
> longer holds.** §"Cone results" below concludes `T_guarded < T_none`
> fails on 12/12 points and reports that as a finding about the guard's
> practical value. `PUNCH-LIST.md` P1 traced that result to a bug —
> `magicset.FindQuery` seeding only the first bindable query candidate,
> leaving a sibling `.output` branch Untouched at full extent — and
> fixed it (`FindQueries`, seed collection from every candidate). After
> the fix, `T_guarded < T_none` holds on **9/12** points. See
> `docs/reports/punch-list-p1.md` for the corrected measurement and
> `docs/reports/FINAL.md` §5/§6 item 5 for the finding as currently
> understood. **The body below is left exactly as originally written**
> (append-only discipline applies to reports as much as to
> `MEASUREMENTS.md`) — read it as the record of what was measured and
> concluded before the bug was found, not as the project's current
> position.

# NIGHT-BATCH-04 B — the cone corpus

Date: 2026-08-27. Four new programs, `tests/corpus/CONE_CORPUS/`, scale
points pre-registered (`SCALE_POINTS.json`) before any measurement ran.

## What did not work

**The first construction swept the new relation into an ENLARGED CULPRIT
SET, not a cone.** Following the task's own suggestion literally
("replace `blocked` with an IDB relation... in `s`'s rule"), the first
draft read a new recursive `gate` relation from `s`'s rule
(`s(x):-q(x,y),gate(y)`). `guard.Decide` reported `culprit=[gate,p,q,s]`,
`cone=[]` — `gate` became a FOURTH culprit member, not a cone member.
Diagnosed directly (`CheckCulpritCycle`'s raw unstratifiable-SCC dump,
not inferred): `gate_b`, `gate_f`, `magic_gate_b`, `magic_gate_f`, and
every `sup_gate_*` checkpoint were genuinely present in the same 28-member
SCC as `p_bf`/`q_bf`/`s_b`. The mechanism, verified by construction, not
assumed: `s`'s (and `q`'s) rules ALL already sit on the transform-induced
negative cycle (`culprit_cycle.dl`'s own header derivation:
`magic_q ->~ s -> q -> magic_q`). Magic-rule generation always creates a
BACKWARD edge (`magic_callee <- caller's own supplementary checkpoint`,
`docs/m2 m3.md` §4). Any new IDB atom read from inside a rule that is
itself part of the cycle gets a magic seed sourced from a checkpoint that
is BOTH fed by and feeds back into that same cycle — closing a NEW loop
through the callee, pulling it into the SAME SCC. This holds for EVERY
rule of `s` and EVERY rule of `q` (both entirely cycle-participating,
confirmed by dumping the raw SCC with `gate` read from `q`'s base case
too — same result).

**The fix**: read the new relation from `p`'s NON-recursive base-case
rule (`p(x,y):-e(x,y),gate(x)`) instead. `p`'s base case is the one
clause among the three predicates that does NOT itself sit on the cycle
(only `p`'s recursive rule, which contains the negation, does) — its
magic seed traces only to the query's own constant seed fact, never to a
cycle-fed checkpoint, so a callee read there has nowhere to loop back
through. Confirmed directly: `culprit=[p,q,s]` (unchanged, exactly 3, no
widening), `cone=[gate]` (genuinely separate). This placement rule is
documented in `cc_cone_only.dl`'s own header comment and used for all
three cone-bearing constructions.

**`T_guarded < T_none` did not hold on any of the 12 measured points, on
any of the 4 constructions** — reported below as a finding about the
guard, per instruction, not chased further with more variants. Two
independent, verified mechanisms explain why, not one:

1. **Declined relations (culprit and cone alike) are always read at full,
   untransformed extent by design** (`M2-M3-BUILD.md` §7's own
   definition of FALLBACK). `gate`/`gate1`/`gate1b`, despite being
   genuine cone members and not culprit members, are computed identically
   to plain, untransformed Souffle — they contribute nothing to
   `T_dlc`'s reduction.
2. **The "sibling" relations (`tc`, `direct`) are not actually
   magic-set-restricted at all — they are simply "Untouched," a
   previously undocumented limitation, confirmed by reading the emitted
   program directly** (`measurements/night04-b-cone/cc_sibling_emptycone/
   n20/emit_guarded/transformed.dl`): `tc`'s rules are copied through
   byte-for-byte, no `magic_tc`, no `tc_bf` adornment, no supplementary
   chain. Cause: `magicset.FindQuery` (per its own doc comment) returns
   only the FIRST bindable-query candidate found in source order; `out2`'s
   `tc(1,y)` and `out3`'s `direct(1,y)` are equally valid candidates but
   are never independently seeded, since only one query drives the whole
   worklist. A "sibling" as this task defines it — reached via a SEPARATE
   `.output` branch — can therefore never benefit from demand restriction
   under the current single-query implementation; it is always computed
   at full extent, identically to `T_none`, regardless of whether the
   guard fires elsewhere in the same program.

Given (1) and (2) together, **every relation in every one of these four
programs is computed at full extent** (declined ones by design, sibling
ones by the single-query limitation) — `T_dlc == T_none` is not a
near-miss, it is exact at all 12 points (verified, not approximated:
e.g. `cc_cone_only` n=20: `T_none=255`, `T_dlc=255`, bit-identical).

## Construction 1 — `cc_cone_only.dl`: nonempty cone, no sibling

`culprit={p,q,s}`, `cone={gate}`. `gate` read from `p`'s base case.

## Construction 2 — `cc_sibling_emptycone.dl`: sibling, empty cone

`p`/`q`/`s` unchanged from `culprit_cycle.dl` (still reads EDB `blocked`
directly) — `culprit={p,q,s}`, `cone={}`, same as every
`CULPRIT_CANDIDATES` program before this task. `tc`/`out2` is the
sibling branch, confirmed Untouched (full extent) per the mechanism
above.

## Construction 3 — `cc_both.dl`: nonempty cone AND a sibling

`culprit={p,q,s}`, `cone={gate}`, plus `tc`/`out2` (Untouched).

## Construction 4 — `cc_cone_proper_subset.dl`: cone ⊊ non-declined

Two-hop cone (`gate1` reads `gate1b`, both IDB): `culprit={p,q,s}`,
`cone={gate1,gate1b}`. Two independent sibling branches (`tc`, `direct`,
both Untouched). Non-culprit IDB relations total 4
(`gate1`,`gate1b`,`tc`,`direct`); cone is exactly 2 of them — a proper
subset with a non-trivial (size-2), not merely size-1, complement.
Exercises `ConeClosure`'s BFS beyond a single hop for the first time in
this project (every prior cone-bearing program, including construction 1
and 3, only ever needed one hop).

## Gate 1 — cone cross-check against `harness/cone_metric.py`

`harness/night04_b_cone_gate.py`, `bin/conecheck` (new, `tools/
conecheck/main.go` — a thin JSON dump of `guard.Decide`, not new guard
logic) vs. `cone_metric.py`'s independent source-graph computation, given
the SAME declined seed (`{p,q,s}`, the real culprit set, identical across
all four constructions):

| program | Go `cone` | Python `cone_metric.py` | agree |
|---|---|---|---|
| `cc_cone_only` | `{gate}` | `{gate}` | **yes** |
| `cc_sibling_emptycone` | `{}` | `{}` | **yes** |
| `cc_both` | `{gate}` | `{gate}` | **yes** |
| `cc_cone_proper_subset` | `{gate1,gate1b}` | `{gate1,gate1b}` | **yes** |

**4/4 exact agreement**, matching M3.3's own cross-check precedent
(`culprit_cycle.dl` `{p}`→`{q,s}`, this task's cross-check generalizes it
to a real, multi-relation declined seed and two genuinely non-empty
cones).

## Gate 2 — three-column table + answer-identical, all 12 points

`measurements/night04-b-cone/summary.json`. `T_none`/`T_souffle`/`T_dlc`
(`T_excl_copy` convention, unaffected by A's excl-sup change since these
programs have no negation-driven adornment collapse to test):

| program | n=20 | n=50 | n=100 |
|---|---|---|---|
| `cc_cone_only` `T_none`/`T_souffle`/`T_dlc` | 255 / 249 / **255** | 787 / 762 / **787** | 3,866 / 3,836 / **3,866** |
| `cc_sibling_emptycone` | 514 / 258 / **514** | 1,101 / 783 / **1,101** | 4,059 / 3,944 / **4,059** |
| `cc_both` | 520 / 275 / **520** | 1,036 / 766 / **1,036** | 3,935 / 3,838 / **3,935** |
| `cc_cone_proper_subset` | 489 / 265 / **489** | 1,011 / 793 / **1,011** | 3,926 / 3,963 / **3,926** |

**Answer-identical: 12/12 (every program, every scale point, every
`.output` relation).** Zero divergences —
`per_relation_identical` checked separately for `out`/`out2`/`out3`
where present, all `true`.

**`T_guarded < T_none` strictly: 0/12.** `T_dlc == T_none` exactly at
every point (not approximately — see "what did not work" for the
mechanism). Per instruction: **reported as a finding about the guard.**
The guard's cone/decline mechanism is demonstrably correct (gate 1's
exact cross-check) and demonstrably exercises real, non-trivial,
multi-hop dependency structures (construction 4) — but on this whole
family of programs, contributing a measured tuple-count reduction
requires either a declined relation to somehow cost less than full
extent (it cannot, by definition) or a same-program sibling to be
genuinely demand-restricted (it cannot be, under the current
single-query `FindQuery` limitation). Closing this would require EITHER
relaxing FALLBACK's always-full-extent semantics (a soundness question,
out of this task's scope) or extending `FindQuery`/`Adorn` to support
multiple independently-seeded queries per program (a real, scoped M2
extension, not attempted here — flagged for `OPEN_QUESTIONS.md`).

## Blast radius, recomputed over the enlarged (16-program) corpus

`measurements/night04-b-cone/blast_radius/` (`bin/conecheck` over all 12
previously-measured programs + these 4 new ones):

| program | IDB | culprit | cone | declined | declined fraction |
|---|---|---|---|---|---|
| `culprit_cycle` | 4 | 3 | 0 | 3 | 0.750 |
| `cc_arity3_twobound` | 4 | 3 | 0 | 3 | 0.750 |
| `cc_edb_negated` | 3 | 0 | 0 | 0 | 0.000 |
| `cc_longer_cycle` | 5 | 4 | 0 | 4 | 0.800 |
| `cc_neg_early` | 4 | 3 | 0 | 3 | 0.750 |
| `cc_query_bothbound` | 4 | 3 | 0 | 3 | 0.750 |
| `cc_third_relation` | 4 | 3 | 0 | 3 | 0.750 |
| `cc_mixed_fallback` | 7 | 3 | 0 | 3 | 0.429 |
| `ancestor_nonancestor` | 3 | 0 | 0 | 0 | 0.000 |
| `reachability_complement` | 3 | 0 | 0 | 0 | 0.000 |
| `same_generation_negation` | 3 | 0 | 0 | 0 | 0.000 |
| `transitive_closure_bound` | 2 | 0 | 0 | 0 | 0.000 |
| **`cc_cone_only`** | 5 | 3 | **1** | 4 | 0.800 |
| **`cc_sibling_emptycone`** | 6 | 3 | 0 | 3 | 0.500 |
| **`cc_both`** | 7 | 3 | **1** | 4 | 0.571 |
| **`cc_cone_proper_subset`** | 10 | 3 | **2** | 5 | 0.500 |

**Total declined relations across the enlarged corpus: 38** (up from the
previously-reported 22 — the 4 new programs add 16, all self-consistent
with the numbers above; 22+16=38, checked arithmetically). **Cone-size
distribution** (the number this session was asked for specifically, not
just a total): of 16 programs, **12 have cone size 0** (unchanged from
before this task — every pre-existing program's culprit SCC already
equalled its full IDB set reachable from the query), **2 have cone size 1**
(`cc_cone_only`, `cc_both`), **1 has cone size 2** (`cc_cone_proper_subset`).
Cone as a fraction of each firing program's OWN declined set: 0.25
(`cc_cone_only`), 0.25 (`cc_both`), 0.40 (`cc_cone_proper_subset`), 0.0
everywhere else. The distribution is bimodal in the sense that matters
for the thesis — either a program's ENTIRE unstratifiable core equals its
IDB set (cone=0, the pre-existing 12) or a constructed program can push
some of that mass into a genuinely separate cone (up to 40% of the
declined set, construction 4) — but no naturally-occurring
(non-constructed) program in this corpus has ever landed anywhere but
cone=0.

## What a skeptic attacks first

- All four cone-corpus constructions are CONSTRUCTED, not found in a
  real corpus — same disclosure the original `culprit_cycle.dl` already
  carries (2026-08-22 ruling §3.1). The placement rule (`gate` must be
  read from `p`'s non-cyclic rule) is itself now a documented artifact of
  a specific implementation mechanism (backward magic-seed edges), not a
  general Datalog fact — a different guard implementation might not have
  this restriction.
- `T_guarded < T_none` failing on 12/12 points is a genuinely negative
  result for the "cone recovers value" half of the thesis — reported as
  such, not softened. The two named causes (always-full-extent decline,
  single-query limitation) are structural, not incidental, so a fifth or
  sixth construction along the same lines would not be expected to
  succeed either; more constructions were not attempted for this reason
  (per instruction: report the finding, don't chase it further).
- The blast-radius total (38) mixes NATURALLY-arising declines (the
  original 12 programs, cone always 0) with DELIBERATELY-constructed ones
  (cone up to 2) — reported side by side, not merged into one misleading
  average.

## Verdict

**B: DONE.** Four constructions built, all four cross-check exactly
against `harness/cone_metric.py` (gate 1: 4/4), all 12 scale points
answer-identical (gate 2, zero divergences), blast radius recomputed with
cone reported as a distribution (0/0/1/1/2 across 16 programs), and the
`T_guarded < T_none` question is answered directly and honestly: **it
does not hold on any measured point**, for two identified, verified
structural reasons, reported as a finding about the guard rather than a
task failure, per instruction.
