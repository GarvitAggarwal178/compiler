# CLAUDE.md — operating contract for `dlc`

> This file is the agent's live working file and keeps evolving here at the
> repo root. `specs/01-agent-contract.md` is an archived snapshot taken at
> restructure time (2026-08-30) — not kept in sync automatically. For the
> readable, narrative history of how this project actually went, see
> `docs/project-log.md`.

You are working on `dlc`, a Datalog compiler built as a semester compiler-design
project. Read `docs/02-design.md` before anything else. It is the source of truth
for scope, metrics, milestones and kill conditions. This file governs *how* you work.

If this file and the blueprint conflict, stop and ask. Do not resolve it yourself.

---

## 0. Non-negotiables

1. **You do not author Lane A code.** See §2. This is not a style preference — the
   project is graded on the author's ability to defend the headline pass line by line
   under hostile questioning. Code you write in Lane A is worse than no code.
2. **No number appears in any document unless it was produced by a committed command
   whose output is committed.** Every integer traces to `measurements/<id>/cmd.txt`
   and `measurements/<id>/stdout.txt`. Numbers from memory, from estimation, or from
   a previous session's context are forbidden.
3. **No wall-clock time in any report.** Not as a headline, not as confirmation, not
   in a footnote. The hardware cannot support timing claims (no PMU under WSL2,
   hybrid CPU, thermal drift). If you catch yourself writing "faster", rewrite it as
   a tuple count.
4. **Determinism.** Every generator is seeded from a constant recorded in the fixture
   file. Two runs of the same command produce byte-identical output or something is
   wrong and you stop.
5. **Scope is closed.** The grammar in blueprint §4 is fixed. No functors, no
   aggregates, no records, no components. If you believe a feature is needed, append
   to `record/OPEN_QUESTIONS.md` and continue with what exists.

---

## 2. Authorship lanes

### Lane A — human-authored. You may read, test, review, and attack. You may NOT write.

- Lexer and precedence parser
- Declaration / arity / type checker
- Allowedness (range restriction) check
- Precedence graph construction, SCC, stratification
- Adornment, SIPS, magic rule generation, supplementary predicates
- The transform-safety guard (both clauses) and the TRANSFORM/FALLBACK decision
- Semi-naive Δ-rewrite
- Join planning and evaluation

Files under `src/` are Lane A by default.

**What you are encouraged to do in Lane A:** write failing tests against it, find
counterexamples, argue that an implementation is wrong, propose a specific test case
that would expose a suspected bug, and say plainly when something looks incorrect. Be
adversarial about it. Just do not edit the file. If asked to "fix" a Lane A file,
respond with the diagnosis and the test that proves it, and stop.

### Lane B — yours. Write freely.

- `harness/` — fixture generation, Soufflé invocation, `souffleprof` parsing, TSV
  set-comparison, corpus runner, golden-file management
- `measurements/` — raw command outputs, provenance
- `tools/` — build scripts, CI, lint
- `docs/` — syntheses and narrative (except `02-design.md`, which is Lane A)
- `experiments/`, `specs/`, `results/` — renamed reports, build plans, and
  conclusions respectively (see `docs/rename-map.csv` for the 2026-08-30
  restructure that moved them out of `docs/reports/`)
- Test *bodies* anywhere, including tests targeting Lane A code
- Instrumentation plumbing that Lane A code calls into (counter sinks, JSON emit),
  provided the counting logic itself stays in Lane A

---

## 3. Language and layout

Go for `src/` (single binary, the author knows it, no build archaeology). Python only
in `harness/` and `tools/`. No third-party parser generators — the parser is a
learning objective.

```
dlc/
  CLAUDE.md                 # this file — live, keeps evolving here
  README.md                 # front door
  docs/
    02-design.md            # Lane A, source of truth (current-state; see below)
    01-problem.md, 03-methodology.md, 04-related-work.md, 05-limitations.md
    design-history.md       # blueprint's old version deltas
    project-log.md          # session-by-session narrative
    06-reproduce.md
    rename-map.csv          # old path -> new path, for the 2026-08-30 restructure
  specs/                    # build plans that directed each phase of work
  record/                   # append-only: SESSION_LOG.md, MEASUREMENTS.md,
                             # DECISIONS.md, OPEN_QUESTIONS.md, ESCALATIONS.md
  experiments/              # every experiment report, renamed by content
  results/                  # findings.md, claims.md, superseded.md, presentation.html
  src/                      # Lane A
  harness/                  # Lane B
  tests/
    programs/               # .dl files, one per behaviour
    golden/                 # expected outputs, generated by ORACLE only
  measurements/<id>/
    cmd.txt  stdout.txt  stderr.txt  env.txt  meta.json
  fixtures/                 # seeded generators + generated .facts
```

The layout above is current as of the 2026-08-30 restructure
(`docs/restructure-notes.md`). `docs/dlc-blueprint.md` no longer exists —
its content was split into `docs/02-design.md` (current-state) and
`docs/design-history.md` (version deltas). `docs/reports/` no longer
exists — its contents moved into `experiments/`, `specs/`, `record/`, and
`results/` per `docs/rename-map.csv`.

---

## 4. Phase gating

**You may not start a phase until the previous phase's report is written and the
human has said to continue.** At each gate you stop, write the report, and end your
turn. Do not begin the next phase in the same session.

### Phase 0 — Probe (do this first, nothing else)

Implement `harness/probe0.py` and the three programs P1, P2, P3 from blueprint §12.
Note the blueprint's fixture requirements: 2000 nodes, ~4000 edges, seeded, and the
reachable-set from node 1 must be verified to be roughly 50 before the run — if it is
not, regenerate rather than reporting a meaningless ratio.

Produce `docs/reports/probe0.md` containing exactly the six answers in blueprint §12,
each with its measurement ID. Then **stop**. [Historical note, 2026-08-30: this
phase's actual report split into three files, now `experiments/01-souffle-
negation-behaviour.md`, `experiments/02-output-forces-materialization.md`,
`experiments/03-completeness-counterexample-search.md` — see `experiments/README.md`.]

Phase 0 has no Lane A code. If you find yourself writing a parser, you have gone
wrong.

### Phase 1 — M1, weeks 1–3

Lane A (human): lexer, parser, type check, allowedness, naive fixpoint, semi-naive.
Lane B (you): differential harness against Soufflé, corpus runner, tuple-count
extraction, rejection-test suite, `docs/reports/m1.md`. [Historical note,
2026-08-30: the actual M1 report is `experiments/31-m1-front-end-and-
evaluators.md`.]

Gate: `T_naive` vs `T_semi-naive` on a fixed program set, plus counts of programs
rejected by each of the four semantic checks, plus set-equality against Soufflé on
every accepted program.

### Phase 2 — M2, weeks 4–6. Phase 3 — M3, weeks 7–11. Phase 4 — M4, weeks 12–15.

Same structure. Scope per blueprint §8.

---

## 5. Escalation protocol

### STOP and report. Do not work around, do not proceed.

- An oracle disagreement (our output ≠ Soufflé's) that you have not explained after
  two focused attempts.
- A measured integer differs from the blueprint's stated expectation by more than one
  order of magnitude, in either direction. Including when it looks *better* than
  expected — especially then.
- Observed behaviour contradicts a blueprint claim. Example: Soufflé transforms P3.
  This does not mean "adjust the blueprint"; it means the project's differentiator may
  be dead and a human decides.
- A required tool, flag, or output field does not exist (e.g. `souffleprof` has no
  per-relation tuple counts).
- The fix requires editing a Lane A file.
- Anything that would change the grammar, a metric definition, the pre-registered
  corpus subdirectory, or a milestone boundary.
- The transform pessimizes: `T_magic > T_base` on any program.
- Two runs of the same command produce different output.

When you stop, write: what you observed, the measurement ID, what you tried, the two
or three explanations you consider live, and the single cheapest experiment that
would distinguish between them. Then end the turn. Do not also propose a redesign.

### CONTINUE, log to `SESSION_LOG.md`, do not ask.

- Build, lint, dependency, or CI plumbing failures with an obvious fix.
- Missing fixtures or test files you can generate within the fixed grammar.
- `souffleprof` output needs a different parsing approach than you first tried.
- Harness is slow, refactors within Lane B, renaming, file organisation.
- A Lane A bug you found: write the failing test, log it, continue with other work.

**Calibration:** stopping is cheap, silently reinterpreting the project is not. But
asking about a `pip install` failure wastes the human's attention and trains them to
skim your escalations. If you are unsure which bucket something falls in, ask — once —
and record the answer in `DECISIONS.md` so the class of question is settled.

---

## 6. Test discipline

**Oracle-first.** No feature is done until it has a differential test against Soufflé
on a `.dl` program that runs unmodified on both. Golden files in `tests/golden/` are
generated by Soufflé and never by `dlc`. If you ever regenerate a golden file from our
own output, you have destroyed the oracle and must say so loudly.

Comparison is **set equality on output relations**, not text diff. Sort, then compare.
Report the symmetric difference, not just a boolean.

Three test categories, all required:

1. **Agreement** — accepted program, our answer set == Soufflé's.
2. **Rejection** — malformed program, we reject with the correct diagnosis. One test
   per rejection ground (arity, type, allowedness, stratification). These are Filter 4
   evidence and carry weight in the report; treat them as first-class, not as
   error-handling.
3. **Metric** — derived-tuple counts under each configuration, recorded with
   provenance, checked for run-to-run stability.

---

## 7. Reporting

`record/MEASUREMENTS.md` is a table: measurement ID, date, command, program, config,
integer, and a one-line interpretation. Append-only. Never edit a past row; if a
number is superseded, add a new row and reference the old ID.

`experiments/NN-<content>.md` (or, before the 2026-08-30 restructure,
`docs/reports/<phase>.md`) is written for a hostile reader. Structure:

- What was measured and under what configuration, with IDs
- The integers, in a table, side by side across configurations
- What the integers mean, in at most three sentences per table
- What did **not** work, what remains unverified, and what a skeptic would attack first
- Open questions moved to `OPEN_QUESTIONS.md` with dates

Write the "what did not work" section before the results section. If it is empty, you
have either had an unusually clean run or you are not looking; say which.

Prose rules: no adjectives on results. "3 orders of magnitude" not "dramatic". No
claim without a measurement ID. If you cannot cite an ID, delete the sentence.

---

## 8. Session hygiene

Start every session by reading `SESSION_LOG.md` and `OPEN_QUESTIONS.md`. End every
session by appending to `SESSION_LOG.md`: what changed, measurement IDs produced, what
is now blocked, and the single next action.

Commits: one logical change, message states the lane (`[B] harness: parse souffleprof
per-relation counts`). Never commit a mix of lanes.

Never `git push --force`. Never rewrite history. Never delete a measurement directory.

---

## 9. Failure modes specific to this project — watch for these

- **Guard vacuity.** If the transform-safety guard ends up rejecting every program
  with negation, the project has no contribution. This is failure mode #1 in the
  blueprint. Measure the blast-radius integer early in Phase 3 and escalate the moment
  it looks like zero.
- **Oracle drift.** Golden files quietly regenerated from `dlc`. Check for this
  explicitly at each gate.
- **Metric selection after the fact.** The corpus subdirectory is pre-registered
  (blueprint Q5). If you find yourself choosing which programs to report on after
  seeing results, stop.
- **Report inflation.** A phase report longer than two pages usually means the numbers
  are thin and prose is compensating. Cut the prose, not the numbers.