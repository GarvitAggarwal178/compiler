# Ruling: corpus closed, version risk opened

Date: 2026-08-22. Applied to blueprint v1.5. **Last spec until M1 exists.**

Correction to `docs/reports/corpus-ruling-2026-08-21.md` §6: it stated "Week 1 is
gone." It is day 3 of 105. The M1 point stands; the arithmetic was wrong and is
withdrawn.

## 1. T0 — Soufflé version risk · cap 1 hour · Lane B · HIGHEST PRIORITY

Soufflé 2.5 released 2025-03-25. PR #2567 merged 2025-12-07. Master therefore
contains changes 2.5 does not, and the project's entire differentiator is an
observed behaviour of 2.5.

**Question: does Soufflé master still refuse to demand-restrict negated
relations?**

If it does not — if master already restricts them — the gap closes at the next
release and this project's premise expires. That must be known now.

Procedure:

1. Clone `souffle-lang/souffle` at master. Record the commit SHA.
2. `git log --oneline` on the magic-set transform path since the 2.5 tag.
   **Commit titles and messages only** — the Prior Art Register's "do not read
   the source" rule holds for implementation. Titles are metadata.
3. Build master, or obtain a nightly/CI artifact if a build is available. Cap
   applies: if it does not build in 30 minutes, record that and skip to step 5.
4. Re-run P2 and P4' against master. Compare against the 2.5 numbers already
   committed: does `@neglabel.reach` still appear, and is it still 26,404
   (unrestricted)?
5. Report regardless of whether step 3 succeeded.

**Outcomes, pre-decided so neither is a surprise:**

- *Master behaves as 2.5.* Differentiator holds. Pin 2.5, document the pin and
  the master SHA checked, note the check in the report as due diligence.
- *Master restricts negated relations.* Escalate. The project's premise has an
  expiry date and the framing must change — likely to "we characterize a
  limitation that was live in the last release and independently fixed," which
  is weaker but honest, or to a pivot decided by a human.

Do not resolve outcome 2 unilaterally.

## 2. Corpus hunt: closed

Three attempts, three sources, one consistent signal:

| Source | Negation-bearing programs | Scale | Zero rate |
|---|---|---|---|
| Soufflé `tests/`, pre-registered 36 | 31 ran | 3/31 clear 1,000 tuples | 37% (n=27) |
| Soufflé `tests/`, exploratory whole tree | 86 | not measured | 34% (n=86) |
| OpenRuleBench / RUBEN catalog | **1 shape**, 3 scale points | good, but unreachable | n/a |

**Ruling:** the search stops. This is the answer to blueprint failure mode #8,
not an obstacle to finding an answer:

> Negation-bearing recursive Datalog with bound queries, at scale, is rare in
> public corpora. Three independent sources agree.

This finding goes in the abstract alongside the Θ(n) result, not in a footnote.
A result that states its own applicability bound is stronger than one that
doesn't, and an examiner who finds the 34–37% zero rate in your data will assume
it was hidden.

No fourth corpus hunt. A fourth hunt is prior-art scanning in a new costume.

## 3. Measurement corpus: canonical benchmark family

Phase 0.7 fallback (1), now adopted: **the rules are external, the fact
generators are ours, and both facts are disclosed.**

### 3.1 Shapes — every one needs a citation

Each program shape must be traceable to a published source. Candidates:

| Shape | Provenance to cite |
|---|---|
| Same-generation with negation | OpenRuleBench negation category; standard magic-sets benchmark |
| Transitive closure, bound query | Beeri–Ramakrishnan; OpenRuleBench `datalog_recursion` |
| Ancestor / non-ancestor | Textbook stratified-negation example |
| Reachability complement | The P2 shape; cite its lineage rather than presenting it as ours |
| Culprit-cycle shape (P5) | Constructed for clause (a); disclose as constructed, not canonical |

A shape without a citation is one we invented and must be labeled as such in the
report. Mixing the two silently is the Filter 2 defect.

### 3.2 Procedure — before any measurement run, one commit

1. Write each shape's `.dl` from its cited source. Record the citation in a
   comment at the head of each file.
2. Write seeded, parameterized generators. Reuse `gen_p1_graph`'s core/rest
   construction so reachable-set size is controllable — the P2-fixture lesson.
3. **Pre-register the scale points** (`n` values per shape) and commit them.
4. Commit shapes + generators + scale points as `tests/corpus/BENCHMARK_FAMILY/`.
5. Do not run.

### 3.3 DOOP — time-boxed, motivation only, optional

`dlc` cannot parse DOOP, so no `T_guard`. But Soufflé alone yields `T_souffle`
and `E_recoverable` on a real production workload, which is the motivating
sentence the project currently lacks: *on a real analysis, X% of derived tuples
sit in relations Soufflé refuses to restrict.*

**Cap 3 hours.** If DOOP is not producing a profile in that time, abandon and
record it. This is a stretch item, never a dependency. Do not attempt before T0
and §3.2 are complete.

## 4. Blueprint → v1.5

- **§7** — measurement corpus is the benchmark family of §3, with the
  external-rules/our-generators split stated explicitly wherever numbers from
  it appear.
- **§9 #8** — promoted from open failure mode to **finding**, with the
  three-source table from §2 as its evidence.
- **§10** — Q5 closed for both corpora. New Q6: *does Soufflé master still
  refuse to restrict negated relations?* — due immediately, per §1.
- **§11** — add: subsumption/magic-transform divergence confirmed real and
  fixed upstream in PR #2567 (post-2.5); our pin is 2.5 and this is a
  known-issue note, not a project finding.

## 5. Lane A — M1

Day 3 of 105. M1 is three weeks and has not started.

The measurement side of this project is in good shape: the mechanism is
characterized, the applicability bound is measured from three sources, the
oracle is validated, the harness exists, and the provenance discipline is real.
None of that is a compiler.

Lexer → precedence parser → decl/type check → allowedness → naive fixpoint →
semi-naive. Three weeks.

**No further specs will be issued until a parser exists.** The ratio of
documents to source files in this repo is currently seven to zero, and
continuing to feed it makes the problem worse rather than better.
