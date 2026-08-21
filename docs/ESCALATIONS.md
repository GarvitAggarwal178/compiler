# Escalations

Append-only. NIGHT-BATCH-01 modified protocol (`CLAUDE.md` §5 semantics, batch §0.1):
on any STOP condition, write a complete entry here, abort that task, continue the
queue. Normal CLAUDE.md §5 STOP-and-wait resumes when the batch ends.

Entry format: observation, measurement/task IDs, what was tried, live explanations,
cheapest distinguishing experiment.

---

## 2026-08-20 — T2/T3: `docs/phase0.7-corpus-viability.md` does not exist

**Observation.** T2 is specified as "Exactly as specified in
`docs/phase0.7-corpus-viability.md` §2.2." T2's output requirement cites "the four
numbers from Phase 0.7 §3." Hard prohibition §0.2.3 names two thresholds from that
same document (`T_none ≥ 1,000`, "the floor of 8") directly in the night-batch text.
The file itself is absent from the repository (`find . -iname 'phase0.7*'` returns
nothing; `docs/` contains no such file).

**What I tried.** Checked `docs/`, `docs/reports/`, repo root, and a full-repo
`find` for any filename matching `phase0.7*` or `corpus-viability*`. Nothing.
Checked `git log` for any commit that might have added and then lost it — this repo
has 5 commits total, all authored this session, none reference the file.

**What I am doing about it, and why this is not "resolving the escalation."** The
per-program measurement fields T2 actually needs (`T_none`, fact input rows,
seedable?, negated-IDB-literal count, out-of-grammar feature count, status) are
given directly, in full, in the night-batch directive's own §T2 table — not only in
the missing external doc. Two of the three "output" requirements (the `T_none ≥
1,000` count and the floor-of-8 check) are also given verbatim in §0.2.3. Running
the measurement itself does not require guessing anything the missing document might
say differently. What I am **not** doing: fabricating "the four numbers from Phase
0.7 §3" — I do not know what those four numbers are defined to be, and inventing a
definition for them would be exactly the kind of unauthorized resolution this
protocol exists to prevent. T2's report states this gap plainly instead of silently
filling it, and reports only what the batch text itself specifies.

**Live explanations:** (1) the document was meant to be written before this batch
and wasn't; (2) the document exists somewhere outside this repo (another session,
another machine) and the night-batch directive assumed it would already be here;
(3) "Phase 0.7" is this batch's own numbering scheme for a doc the human intended to
attach but didn't.

**Cheapest distinguishing experiment:** none needed from my side — this is a
missing-input problem, not an ambiguous-observation problem. The human either has
the file and can supply it, or needs to write it. Either resolves this immediately.

**Task disposition:** T2 proceeds using the batch text's own inline specification
(see `docs/reports/night01-T2-corpus.md`), not aborted, since its core measurement
is fully determined without the missing file. T3 proceeds on the same basis (its
stated prerequisite is T2's seedable subset, which does not depend on the missing
document either).

---

## 2026-08-20 — T3: `semantic/subsumption_multiple_rules` diverges under `--magic-transform=*`

**Observation.** Of 31 candidates T3 swept (26 completed before this one fired the
abort condition, in alphabetical order), `semantic/subsumption_multiple_rules`'s
untransformed and magic-transformed runs produce **different sets** of tuples (not
just different row order) in at least 4 output relations: `A`, `E`, `F`, `rel`.
Concretely, relation `A` (a shortest-path tree over a fixed 7-node graph, built via
subsumptive `<=` rules with `btree_delete`): untransformed produces 6 rows, missing
`(5,6,3)`; magic-transformed produces 7 rows, including it.

**A first alarm on this same task was a false positive, corrected before this one
was trusted.** `example/orbits1` initially looked identical-but-diverging under a
raw byte comparison; sorting first (CLAUDE.md §6: "set equality on output
relations... sort, then compare") showed it was only row order. The harness
(`harness/night01_t3_envelope.py`) was fixed to sort before comparing — this is
bringing the tool into line with an existing project rule, not loosening a check.
`subsumption_multiple_rules` survives the corrected, sorted comparison: the sets
genuinely differ.

**Measurement IDs:** `measurements/night01-t2/semantic__subsumption_multiple_rules/`
(untransformed), `measurements/night01-t3/semantic__subsumption_multiple_rules/`
(magic-transformed).

**What I tried (read-only, no new commands, per "do not investigate further
tonight"):** read the `.dl` source and the sorted CSV diff directly. The relations
that diverge (`A`, `rel`, `F`, and one `E`/`AF` fragment) all use Soufflé's
subsumption operator (`<=`) with the `btree_delete` qualifier — a feature entirely
outside blueprint §4's grammar (`dlc` will never parse or execute it). `A`'s
specific divergent rule chain (`A(from,to,c+1):-A(_,from,c),graph(from,to).` plus
two `<=` subsumption rules) has **no negation at all** — this divergence is not
obviously a completeness-under-negation question in the sense this project cares
about.

**Live explanations:**
1. Genuine interaction bug between subsumption's deletion-by-priority semantics and
   magic-set demand-driven evaluation order: which facts get subsumed-and-deleted
   before others are even derived can legitimately depend on evaluation order, and
   magic-transform changes that order relative to naive/semi-naive untransformed
   evaluation. If so, this is a property of subsumption + magic-sets, not of
   negation, and outside what this project's guard is meant to address.
2. A known, possibly-documented limitation of combining `<=`/`btree_delete` with
   `--magic-transform=*` in Soufflé (subsumption is a comparatively newer feature).
3. Corpus contamination: this program was included by the mechanical predicate for
   an unrelated reason (it also has genuine negated-IDB literals and constant-
   bearing outputs elsewhere in the same file, e.g. `AF`'s rule), but the specific
   divergence found lives in a subsumption-only fragment the predicate never
   distinguished from the negation-relevant fragments — the file bundles many
   independent test cases under one `.dl`.

**Cheapest distinguishing experiment, not run:** isolate the `A`/`graph` fragment
(lines defining `A`, `graph`, and their rules only) into its own `.dl` file and
re-run both configurations. If the divergence persists in isolation, explanation 1
or 2 gains weight; if it disappears, some cross-fragment interaction in the bundled
file (explanation 3, or something else entirely) is implicated instead.

**Task disposition:** T3 aborted at this program per the batch's explicit
instruction. 26 of 31 candidates completed before the abort (results retained,
`measurements/night01-t3/summary.json`); the remaining candidates alphabetically
after this one were not run. This is out of scope for `dlc` regardless of outcome
(subsumption is not in blueprint §4's grammar) — logged for a human decision on
whether to file upstream, exclude subsumption-using programs from the corpus
predicate (prohibition #2 blocks that tonight), or leave it, not investigated
further tonight.

**Resolved 2026-08-21** (corpus ruling §4.1): minimized to 4 nodes/4 edges/1
subsumption rule, confirmed the divergence isolates cleanly (rules out
cross-fragment contamination). Already reported and already fixed upstream —
[souffle-lang/souffle#2322](https://github.com/souffle-lang/souffle/issues/2322),
[#2323](https://github.com/souffle-lang/souffle/issues/2323), fixed by
[PR #2567](https://github.com/souffle-lang/souffle/pull/2567) (merged 2025-12-07,
commit `7bb8e64`) — more than 8 months after our installed Soufflé 2.5. No new
issue filed. Full writeup: `docs/reports/subsumption-repro.md`.

---

## 2026-08-21 — OpenRuleBench pre-registration blocked: neither candidate source is reachable

**Observation.** Corpus ruling §2.2 step 1 named two candidate sources: the
original OpenRuleBench distribution, and the RUBEN repository. Neither yields the
actual rule-program files. Original distribution's named host
(`rulebench.projects.semwebcentral.org`, from OpenRuleBench's own manual PDF)
does not resolve (DNS). RUBEN (`github.com/kev-ang/RUBEN`, cloned locally) is a
Java/Docker orchestration framework with no rule files of its own; its
`Benchmark_Configuration.json` points at a private local path
(`/Volumes/T7/RuleEngineBenchmark/test_data`) not shipped in the repo, and its
referenced external dataset host (`dataset.sti2.at/RUBEN/`) refuses connections
from two independent network paths.

**What I tried.** 4 web searches (original site, GitHub mirrors, direct file
extensions, artifact archives), 3 page/PDF fetches (project manual, a related
2023 paper, RUBEN's GitHub page), 1 direct git clone + local file-tree
inspection, 2 direct connection attempts (`curl`, WebFetch) against the
referenced dataset host from separate network paths. Did not fabricate
placeholder rule files or guess at program content to force the procedure
through.

**What was found anyway:** RUBEN's own `all_tests.json` (a real config file, not
inference) shows OpenRuleBench/RUBEN's entire `negation` category is **one
program shape** (`same_generation`) at 3 scale points — regardless of the access
problem, this independently suggests the floor of 8 (ruling §2.2) will not be met
by program-shape diversity, even if data-scale variants are counted separately.

**Live explanations:** (1) OpenRuleBench's canonical distribution has simply gone
offline in the years since 2009 and no maintained mirror exists; (2) a working
copy exists somewhere this session's search tools can't reach (behind
authentication, on an institutional server not indexed, or held privately by
researchers who've used it); (3) RUBEN's authors also hit this and that's why
`testDataPath` points at a private local drive rather than something the repo
fetches automatically.

**Cheapest distinguishing experiment, not run:** none available to this session
— this is an external-reachability problem, not an ambiguous-observation one.
Resolving it needs either a human with a channel this session lacks (direct
author contact, institutional access) or a decision to route around it.

**Task disposition:** `docs/reports/openrulebench-preregistration.md` written in
full. `tests/corpus/MEASUREMENT_PREREG.txt` **not created** — no fabricated file
list. Blueprint v1.4 §10's Q5 reopening updated to reflect blocked status, not
silently left looking unresolved-but-in-progress.
