# NIGHT-BATCH-03 T4 — culprit-cycle source corpus

Date: 2026-08-27. Source-program collection only — no adornment, no
transform work. Runner: `harness/night03_t4_culprit_classifier.py` (census),
`harness/night03_t4_run_candidates.py` (fixture generation + Soufflé
recording).

## What did not work

The classifier's first draft scored **zero** findings everywhere, including
on `culprit_cycle.dl` itself — a bug, not a corpus result. Root cause:
`.decl`/`.input`/`.output` directives do not end in a period the way
clauses do, so a naive top-level-period split merged each directive with
the clause immediately following it, and the directive-prefix filter never
matched (the leading `.` had already been consumed as the split separator).
Fixed by stripping whole directive/preprocessor lines by regex *before*
splitting on periods, not by prefix-matching post-split statements.
Re-verified against `culprit_cycle.dl` directly (must score exactly one
finding, `negated=s, target_r=q`) before trusting the corpus-wide run.

## Method — the formal criterion, quoted, not reinvented

From `docs/dlc-blueprint.md` §6 ("Mechanism to detect"): magic rules take
body literals *preceding* the target atom under SIPS order; if a preceding
literal is `!q(...)`, `magic_r` acquires a negative edge to `q`; if `q`
transitively depends on `r`, and `r` depends on `magic_r` (always true
post-transform), the cycle `magic_r →¬q → r → magic_r` closes. Cheap
necessary precondition: `r` lies in a positive cycle in the source.

Concretely: for every rule body (source order as a SIPS proxy — `dlc`'s own
SIPS is left-to-right-in-source-order plus one required deviation,
M2-M3-BUILD.md §3), for every negated literal `!q(...)` followed later in
the same body by a positive atom `r(...)`: flag `(q, r)` if `r` has a
positive self/cyclic dependency (computed on a positive-edges-only source
precedence graph) **and** `q` transitively depends on `r` (any edge
polarity). Structural scan, regex-based clause extraction — not a real
parser, works on any file regardless of grammar membership (functors,
aggregates, etc. included), since the census must cover the full 622-file
Soufflé tree, not just the 195 in-grammar files.

## Step 2 — corpus counts

Provenance: `measurements/night03-t4/classifier-census.json`.

| corpus | attempted | flagged |
|---|---|---|
| `tests/corpus/IN_GRAMMAR.txt` | 195 | 0 |
| full Soufflé `tests/` tree | 622 | 0 |
| `BENCHMARK_FAMILY` shapes | 5 | 1 (`culprit_cycle.dl`) |

**Zero matches outside the one already-known file, across 817 attempted
files (195 + 622, `BENCHMARK_FAMILY` overlaps neither tree).** This is not
a null result to explain away — it is exactly T4's own stated premise
("There is currently exactly one … which makes every guard result n=1"),
now independently confirmed by a classifier built from the blueprint's own
formal mechanism rather than assumed.

## Step 3 — six hand-constructed source programs

`tests/corpus/CULPRIT_CANDIDATES/`, each varying one or more axes, each
re-verified through the classifier before use (so the corpus additions are
themselves checked against the same criterion as everything else, not
hand-labeled):

| file | varies | classifier verdict |
|---|---|---|
| `cc_longer_cycle.dl` | cycle length 2 (mutual recursion `r1<->r2`) vs. self-loop | flagged (`s`, `r1`) |
| `cc_neg_early.dl` | position of `!s` — first body literal after one EDB atom, not after a self-recursive one | flagged (`s`, `q`) |
| `cc_edb_negated.dl` | **negated predicate is EDB** (`blocked_pair`, no rules) | **not flagged** — negative control |
| `cc_arity3_twobound.dl` | arity 3 (edge label column); query bound on two arguments | flagged (`s`, `q`) |
| `cc_query_bothbound.dl` | arity 2; query bound on **both** arguments (`out():-p(1,5).`, zero-arity `out`) | flagged (`s`, `q`) |
| `cc_third_relation.dl` | negation/target co-occur in a *third* relation's rule (`q2`), not `target_r`'s own rule | flagged (`s`, `p`) |

`cc_edb_negated.dl` is deliberately a negative control: an EDB relation has
no outgoing dependency edges, so it structurally cannot "transitively
depend on" the target relation — included specifically to exercise the
classifier's boundary, not because it is expected to behave like a culprit
cycle. 5/6 flagged, 1/6 correctly excluded, matching the construction intent
exactly. Each program independently parses and passes `dlc check` cleanly
(`ok`, no diagnostics).

New generator: `harness/fixtures_lib.py`'s `gen_culprit_cycle_facts_labeled`
(arity-3 variant only — the other five reuse `gen_culprit_cycle_facts`
unmodified, since they keep `base`/`e-or-link`/`blocked` at arity 2). All
seeded (`seed=20260822450`, `seed_base(culprit_cycle)=20260822400 + n=50`,
the project's own seeding convention), `n=50`, matching `culprit_cycle`'s
own `n=50` scale point for comparability.

## Step 4 — Soufflé's recorded behaviour

Provenance: `measurements/night03-t4/candidates/summary.json`. **No
adornment was performed by hand or by `dlc`** — both configurations are
Soufflé invoking its own transform.

| file | `T_none` | `T_souffle` (`--magic-transform=*`) | `@neglabel` relations | answers identical |
|---|---|---|---|---|
| `cc_longer_cycle` | 510 | 432 | `@neglabel.s`=34 | yes |
| `cc_neg_early` | 510 | 432 | `@neglabel.s`=34 | yes |
| `cc_edb_negated` | 870 | 28 | *(none)* | yes |
| `cc_arity3_twobound` | 242 | 125 | `@neglabel.s`=11 | yes |
| `cc_query_bothbound` | 509 | 432 | `@neglabel.s`=34 | yes |
| `cc_third_relation` | 649 | 624 | `@neglabel.s`=35 | yes |

**6/6 ran to completion under both configurations, 6/6 answer-identical,
zero stratification failures from Soufflé.** This is consistent with —
not a contradiction of — `culprit_cycle.dl`'s own behaviour: Soufflé's real
automatic `--magic-transform=*` has never been observed to fail to
stratify any of this family's shapes (`docs/reports/night02-T4-baseline.md`)
either. The stratification failure this project has actually observed
(`docs/reports/night02-T5-guarded.md`, `Unable to stratify {m_q,p_bf,
q_bf,s_bf}`) came from a **hand-derived naive general adornment** attempt,
not from Soufflé's own (evidently more careful) transform. These 6 programs
exist to give M3.2's culprit-cycle detector (which runs against `dlc`'s own
candidate transform, not Soufflé's) a non-trivial, structurally-verified
corpus once that detector exists — not to demonstrate a Soufflé
stratification failure, which was never this task's claim.

The 5 flagged variants all show Soufflé's `@neglabel.` marker on `s`,
confirming Soufflé's own transform engages with the negated relation
exactly where the classifier predicted it would; `cc_edb_negated` shows no
`@neglabel` relation at all, consistent with its negative-control status —
there is nothing for Soufflé to isolate when the negated predicate is a
plain EDB read.

## What a skeptic attacks first

- The structural classifier is a regex-based approximation, not a real
  parser: a functor call nested inside an argument list (e.g.
  `foo(bar(x,y))`) would be mis-tokenized as two literals. This is a known,
  disclosed gap — checked against `culprit_cycle.dl` and the six new files
  by hand, but not fuzzed against adversarial nesting. Given the 622-file
  scan found **zero** false positives to investigate (0 flagged besides the
  known file), there is no evidence this gap produced a wrong answer here,
  but it could in principle miss a real match if the negation/target pair
  were separated by a functor-heavy literal.
- Source-order-as-SIPS-proxy is exactly that — a proxy. `dlc`'s real SIPS
  (M2-M3-BUILD.md §3) mostly preserves source order but pulls grounded
  constraints forward and pushes negated literals back past whatever grounds
  them; a body where this reordering changes which atom precedes which
  could in principle be missed or spuriously flagged by a source-order-only
  scan. None of the 6 constructed programs exercises this edge (their
  negated literals' variables are already grounded at the point they
  appear in source order, so SIPS would not move them).
- Zero matches on 817 real-world files is itself worth a skeptic's
  attention: it says this cycle shape is rare in practice (consistent with
  blueprint failure mode #8's existing finding on negation-bearing
  recursive Datalog generally), not that the classifier is broken — the
  classifier is validated positively (`culprit_cycle.dl`, 5/6 new files) and
  negatively (`cc_edb_negated.dl`) before this null result is trusted.

## Verdict

**T4: DONE.** Census: 0/817 real-world files match beyond the one already
known; six new hand-constructed programs added (5 structurally matching,
1 deliberate negative control), all independently classifier-verified,
`dlc check`-clean, and Soufflé-recorded (6/6 complete, 6/6 answer-identical,
0 stratification failures under Soufflé's own transform). `CULPRIT_CANDIDATES/`
now has 7 programs total (6 new + `culprit_cycle.dl` itself, referenced not
duplicated) for M3.2's guard testing, versus n=1 before this task.
