# NIGHT-BATCH-01 — T4: whole-tree exploratory sweep

**EXPLORATORY — NOT PRE-REGISTERED, NOT REPORTABLE AS A RESULT.**

Date: 2026-08-20. Outcome: done (within the exploratory scope defined below —
this task is capped by design, not by an abort condition). Structural
reconnaissance only, to inform a future human corpus decision. Nothing here
supersedes or feeds back into `tests/corpus/PREREGISTERED.txt`
(NIGHT-BATCH-01 prohibition #2).

## What did not work first

The slow pass's initial candidate filter accidentally reused the exact same
predicate condition (`has_negated_idb AND seedable`) that built the pre-registered
36-program corpus — so the first run's "exploratory" set was just T3's set again,
defeating the point of "not just the 36." Fixed: broadened to every program with
`≥1` negated IDB literal, dropping the `seedable` requirement (107 programs) — this
is the right reconnaissance scope, since "how common is the `@neglabel` shape at
all" doesn't require the pre-registration-specific seedability condition.

Separately, `tuple_report.analyze` crashed (`KeyError: 'relation'`) partway through
on a profile log with no `relation` key at all (a program whose Soufflé run
produces essentially no relation activity). Fixed defensively (`.get("relation",
{})`); full 107-candidate slow pass re-run from scratch after the fix.

## Method

**Fast pass (full coverage, no execution):** every `.dl`-bearing directory in
Soufflé's `tests/` tree (612 total), `corpus_predicate.check_program` (read-only
import) for `has_negated_idb` and `seedable` — pure text scan, no Soufflé run.

**Slow pass (capped at 150, actually needed 107):** every program with `≥1`
negated IDB literal, `--magic-transform=*` only (no untransformed baseline needed —
T4 doesn't do answer-comparison), lexicographic order, checkpointed to disk every
50 programs. Global caps applied (300s, 8GB).

## Fast-pass counts (full tree, n=612)

| Metric | Count | Rate |
|---|---|---|
| Total `.dl`-bearing directories | 612 | — |
| With `≥1` negated IDB literal | 107 | 17.5% |
| Seedable (constant-bearing `.output`) | 242 | 39.5% |
| Both (= the pre-registered corpus) | 36 | 5.9% |

## Slow-pass counts (n=107, negated-IDB programs, `--magic-transform=*`)

| Status | Count |
|---|---|
| `ok` | 86 |
| `error:returncode-1` (Soufflé-rejected) | 16 |
| `error:returncode--6` (crash) | 4 |
| `DNF:timeout-300s` | 1 |
| `DNF:memcap-8gb` | 0 |

**`≥1` `@neglabel.` relation produced: 57 of 86 ok (66%).** Soufflé isolates at
least one negated relation on two-thirds of programs that have a negated IDB
literal at all — the mechanism this project's differentiator depends on is common,
not a corner case, at this reconnaissance scope.

**`E_recoverable = 0`: 29 of 86 (34%)** — consistent with T3's 37% on the smaller
pre-registered subset; the two numbers agree well enough to suggest T3's 37% isn't
a fluke of the smaller sample.

**`E_recoverable / T_souffle` distribution (n=84, excludes 2 with `T_souffle=0`):**

| min | Q1 | median | Q3 | max |
|---|---|---|---|---|
| 0.0 | 0.0 | 0.068 | 0.152 | 0.733 |

Wider spread than T3's pre-registered-subset distribution (max 0.571) — expected,
since this set isn't filtered for seedability and includes more program shapes.

## What a skeptic attacks first

- 21 of 107 (20%) either errored or DNF'd — a meaningfully higher failure rate than
  T2's 5/36 (14%) on the pre-registered (seedable-filtered) set. Dropping the
  seedability filter pulled in more out-of-grammar/edge-case programs, as expected.
- This task counts programs, not weighs them — one large program and one trivial
  program each count once toward the 66%/34% rates above. No claim here is about
  "how much of the total derived-tuple mass" carries the gap, only "how many
  programs show it at all."
- n=107 is still a small fraction of `souffle-lang/souffle`'s real-world usage;
  this is one project's own test suite, self-selected by Soufflé's own authors to
  exercise Soufflé's own features, not a neutral sample of Datalog-with-negation
  programs in general.

Raw provenance: `measurements/night01-t4/fast_pass.json`,
`measurements/night01-t4/slow_pass.json`, `measurements/night01-t4/summary.json`,
per-program run directories under `measurements/night01-t4/runs/`.
