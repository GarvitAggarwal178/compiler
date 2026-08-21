# OpenRuleBench measurement-corpus pre-registration — blocked at step 1

Date: 2026-08-21. Corpus ruling §2.2. Outcome: **blocked**, not completed. This
report documents what was found before hitting a genuine external-access
blocker, and is written instead of a silent gap or a fabricated file list.

## What the procedure asked for

1. Obtain the suite and its generators; record source and commit/version.
2. Coverage check: count programs with ≥1 negated IDB literal AND a query with a
   bound argument, reported before selecting anything.
3. Mechanical inclusion predicate, structural only; commit predicate + file list
   as `tests/corpus/MEASUREMENT_PREREG.txt`.
4. Pre-register scale points, committed before any run.
5. Do not run.

## Step 1: could not obtain the suite

Both named candidate sources (corpus ruling §2.2 step 1) are dead:

- **Original OpenRuleBench distribution.** The project's own site
  (`www3.cs.stonybrook.edu/~pfodor/openrulebench_web/`) is a live but ancient
  (2009) frameset page; its own manual PDF names the actual project host as
  `rulebench.projects.semwebcentral.org` — DNS does not resolve
  (`getaddrinfo ETIMEOUT`). SemWebCentral (a SourceForge-style host for semantic
  web projects) has been shut down for years; no working mirror found in 4
  targeted web searches.
- **RUBEN** (`github.com/kev-ang/RUBEN`, cloned locally, commit at clone time
  recorded in `/root/ruben-src/.git`). This is a Java/Docker *orchestration*
  framework for running multiple reasoning engines through their own APIs — it
  contains **no rule/program files and no generators of its own**.
  `src/main/resources/Benchmark_Configuration.json` points `testDataPath` at
  `/Volumes/T7/RuleEngineBenchmark/test_data` — a private local path on the
  original author's machine, not shipped in the repo. The README's referenced
  external dataset host, `dataset.sti2.at/RUBEN/`, refuses connections
  (`ECONNREFUSED`) from two independent network paths (WSL `curl`, and the
  WebFetch tool separately) — also dead.

No fabricated placeholder files were created. `tests/corpus/
MEASUREMENT_PREREG.txt` does **not** exist as of this report — steps 2-4 cannot
be honestly completed without the actual rule-program text.

## What was found anyway, and why it matters regardless of the access problem

RUBEN's `src/main/resources/all_tests.json` — a real, version-controlled
configuration file, not a guess — lists every test case OpenRuleBench/RUBEN
tracks: **34 total, across exactly 3 categories:**

| Category | Test names | Count |
|---|---|---|
| `large_joins` | `join1`, `join2`, `lubm_derived`, `mondial`, `dblp` | 5 distinct programs |
| `datalog_recursion` | `same_generation`, `transitive_closure`, `wine`, `wordnet` | 4 distinct programs |
| `negation` | `same_generation` | **1 distinct program**, at 3 scale points: `SameGeneration_6000`, `SameGeneration_12000`, `SameGeneration_24000` |

**The entire negation category is one program shape** (same-generation with
negation, the classic textbook magic-sets/negation benchmark), scaled to three
sizes. Even granting every possible benefit of the doubt — treating each scale
point as independently countable, and assuming (not yet verified, since the
actual rule text is unobtainable) that this program's query has a bound
argument — that is **3, not the floor of 8** the ruling's §2.2 sets.

This matches the concern the ruling itself named: "OpenRuleBench's negation
coverage is unverified — if it turns out to be thin, the measurement corpus
problem is not solved and that must surface immediately." It has surfaced
immediately, from the catalog alone, before the deeper access problem was even
fully chased down.

## Disposition

**Blocked, reported per instruction, not acted on further.** Two live findings,
independent of each other, both point the same direction:

1. Neither candidate source for the actual files is currently reachable.
2. Even the catalog-level metadata (no file contents needed) suggests the
   negation category is far too small-scale in *program diversity* — 1 shape —
   regardless of how well any single one of those programs scales in *data
   size*.

This is a human decision, not one for this session to resolve by substituting a
different benchmark suite unilaterally (corpus ruling §2.2's own framing: DOOP
was already considered and rejected, and picking a third option without
authorization would repeat exactly the "corpus selected after seeing what's
convenient" failure mode `docs/dlc-blueprint.md` §7 warns against). Options for
the human to choose among, not chosen here:

- Track down a working OpenRuleBench mirror through a channel this session
  doesn't have (direct author contact, an institutional archive, a colleague's
  local copy) and re-attempt steps 2-4 once the files are in hand.
- Accept that OpenRuleBench's negation coverage is inherently thin (1 program
  shape) and treat that as the answer to "is negation-bearing recursive Datalog
  with bound queries rare in real corpora" (blueprint §9 failure mode #8) rather
  than pursuing more scale points of the same shape.
- Select a different measurement corpus, following the same rejection-reasoned
  process the ruling used for DOOP.
