# NIGHT-BATCH-03

Date issued: 2026-08-27. Unattended overnight batch. Eleven tasks, ordered by
value — if the batch dies at 3am, T1–T4 are what must be done.

---

## 0. Batch protocol

### 0.1 Escalation semantics — read this first

`M1-BUILD.md` §6's STOP-and-wait is **suspended for this batch**, per its own
clause allowing a batch to authorise otherwise.

On any blocker: write the escalation to `docs/ESCALATIONS.md` with the full
diagnosis, mark the task `BLOCKED` in the batch report, and **move to the next
task**. Do not stop the batch. Do not work around a blocker by weakening a gate
or narrowing scope — record it and move on.

Exception: if `go build ./...` breaks and you cannot restore it within one
focused attempt, revert to the last clean commit and stop the batch entirely.
A broken build blocks every downstream task.

### 0.2 Hard prohibitions — unchanged, non-negotiable

1. **No Lane A code.** `src/transform/magicset/`, `src/transform/guard/`,
   `src/eval/fallback.go` stay as 2-line marker files. Not stubs, not
   signatures, not algorithm sketches, not comments describing what the
   algorithm will do. If a task appears to require Lane A code, it does not —
   re-read the task, and if it still appears to, mark it BLOCKED.
2. **Never regenerate a golden file from `dlc`.** Goldens come from Soufflé.
3. Do not edit `tests/corpus/PREREGISTERED.txt`, `BENCHMARK_FAMILY/*.dl`,
   `SCALE_POINTS.json`, or any committed `measurements/` directory.
4. Do not edit `tests/corpus/IN_GRAMMAR.txt` or `harness/night01_t5_grammar.py`.
   T7 creates new files alongside them; it does not modify them.
5. No wall-clock in any report. Tuple counts only.
6. All Soufflé invocations file-redirected, never an interactive stream read.
   `wsl.exe -e bash -lc '...'` against the real WSL path, not the UNC bridge —
   VERIFY-01 hit the UNC `RLock` failure again.
7. No `git push --force`, no history rewrite.

### 0.3 Reporting

Per task, append to `docs/SESSION_LOG.md`: task ID, status
(`DONE` / `BLOCKED` / `PARTIAL`), the gate number, commit SHA, what is now
unblocked.

At batch end, write `docs/reports/night03-summary.md`. **The "what did not work"
section comes first.** For every task, state its result as a number. For any
task where a check found nothing, say so explicitly — "T8: exercised 14 symbol
comparison cases, zero defects found" is required output, not an omission. A
task reported with no number is not done.

Commit per task, message `[B][night03][T<n>]`.

### 0.4 Standing instruction

Do not narrow scope. You have done this twice before and disclosed it both
times, which is the right behaviour, but the fix is not to disclose it faster.
If a task says "all five shapes," do all five. If you believe a task is
mis-scoped, do it as written and add a note.

---

## T1 — Printer output must be Soufflé-parseable

**Why first:** the M3 measurement protocol (T2) depends entirely on `dlc` being
able to emit a `.dl` file that Soufflé can execute. If `parser/printer.go`'s
output is not Soufflé-legal, the protocol needs a different design and every
task after T2 is built on sand. This is the highest-risk unknown in the batch.

**Do:**

For each of the 20 files that currently parse (`m1-3.3-gate1-parse-coverage-summary.json`),
plus all five `BENCHMARK_FAMILY/*.dl` shapes, plus `tests/programs/p4prime.dl`:

1. `dlc` parse → print → write to a temp `.dl`.
2. Run **Soufflé** on the printed file with the shape's own fixture.
3. Run Soufflé on the **original** file with the same fixture.
4. Compare answer relations by sorted set equality.

**Gate:** `soufflé-accepted/attempted` and `answer-identical/attempted`, both
reported. Anything below 100% on the shapes is a defect in the printer, not in
the corpus — diagnose it and, if it is a printer bug (Lane B), fix it and
re-run. If it is a grammar-coverage issue, record it and move on.

**Deliverable:** `harness/night03_t1_printer_souffle.py`,
`docs/reports/night03-T1-printer-souffle.md`.

**If this fails:** mark BLOCKED, state precisely what Soufflé rejected, and
**skip T2** — do not build a protocol on a printer that cannot round-trip.
Go to T3.

---

## T2 — M3 measurement protocol, end-to-end, on the identity transform

**Why:** M1 §4.3 hit 4 DNFs at a 120s cap because `dlc` is a tree-walking
interpreter with no query planning. M3's headline numbers live at scale points
`dlc` cannot reach. The fix is that `dlc` decides and emits; **Soufflé
evaluates**. Building and validating that path now, against `PassThrough`,
means M3 measures on day one instead of week two.

**Do:**

Build `harness/night03_t2_protocol.py` implementing:

```
source.dl
  → dlc (transform via transform.Transformer) → transformed.dl
  → souffle -F<fixtures> -D<out> -p prof.log transformed.dl
  → harness/parse_profile.py → exact per-relation tuple counts
  → compare answer relations against souffle(source.dl) — sorted set equality
```

`dlc` needs a subcommand that parses, runs sema, applies the configured
`Transformer`, and prints. Name it `dlc emit`. **`PassThrough` is the only
transformer wired in** — this is Lane B plumbing, and the identity transform is
the whole point: if the pipeline is correct, `T(emit) == T(source)` exactly, on
every shape, at every scale point.

**Gate:** all five shapes, every pre-registered scale point in
`SCALE_POINTS.json`: tuple totals identical between the source run and the
emitted run, answers set-equal. Report `identical/attempted`. Any deviation is a
bug in the emit path and must be diagnosed, not tolerated.

Include a `--transformer=` flag reading a name, so that when Lane A lands the
only change is registering a second implementation. Do not implement a second
implementation.

**Deliverable:** `harness/night03_t2_protocol.py`, `dlc emit` in
`cmd/dlc/main.go`, `docs/reports/night03-T2-protocol.md`, `DESIGN.md` updated.

**Note:** DNFs are now irrelevant to measurement but `dlc`'s own evaluator
remains the correctness path — §3.8/§3.9 gates are unaffected and must still
pass.

---

## T3 — M2 acceptance harness and Soufflé reference transforms

**Why:** this is the biggest single accelerator for M2. Without it, the human
writes an adornment/magic-rule generator and has no way to know it is right
until the whole thing is done. With it, every partial implementation gets a
number.

**Two parts. Part B is the deliverable; part A is supporting material.**

### T3a — Soufflé's own transformed programs, as reference

Verify the exact option name first (`souffle --help`; M1 §3.6 used the `--show=`
family, and `transformed-datalog` is expected but confirm rather than assume).

For each of the five `BENCHMARK_FAMILY` shapes plus `p1prime.dl`, `p2.dl`,
`p4prime.dl`, `culprit_cycle.dl`: dump Soufflé's magic-transformed program under
`--magic-transform=*`, and also without it, and commit both under
`tests/reference/souffle-transformed/`.

**These are reference material, not golden targets.** Soufflé applies inlining,
its own naming (`@magic.`, `@neglabel.`, `@poscopy_1`), and other transforms
`dlc` will not replicate. A syntactic diff against these is worthless — do not
build one. Their value is that the human can read what a correct transform of a
known program looks like.

### T3b — Semantic acceptance harness for any candidate transform

Build `harness/m2_accept.py`. Input: an original `.dl`, a candidate transformed
`.dl`, a fixture directory. Output:

- Answer relations set-equal between the two? (correctness)
- Exact tuple total for each, via `parse_profile.py`.
- Ratio, in the project's three-column convention where applicable.
- Per-relation breakdown for both.

This works on **any** candidate transform regardless of naming — Soufflé's,
`dlc`'s, or a hand-written one — because the check is semantic. That is why it
survives contact with Lane A's actual output.

**Gate:** demonstrate the harness on three known-good pairs already in the repo
(`p2.dl` / `p4prime.dl`; and two `BENCHMARK_FAMILY` shapes with their
`guarded/` counterparts). Report the numbers it produces and confirm they match
the committed T5 values exactly. If they do not match, the harness is wrong —
diagnose before proceeding.

**Deliverable:** `tests/reference/souffle-transformed/`, `harness/m2_accept.py`,
`docs/reports/night03-T3-m2-harness.md`.

---

## T4 — Culprit-cycle source corpus

**Why:** M3's guard needs programs where it should fire. There is currently
exactly one (`culprit_cycle.dl`), which makes every guard result n=1.

**Scope carefully — this is source-program collection, not transform work.**

Structural criterion, from the literature and already recorded in the blueprint:
a culprit cycle requires a predicate in a **positive cycle** whose magic
predicate would acquire a negative edge. So:

1. Write a structural classifier (Lane B, pure graph analysis on the parsed
   source precedence graph — `sema.CheckStratification` already builds this):
   flag any program containing a predicate `r` such that (i) `r` is in a
   positive cycle, and (ii) some rule body defining `r`, or a rule body in `r`'s
   SCC, contains `!q(...)` where `q` transitively depends on `r`.
2. Run it over: all 195 `IN_GRAMMAR.txt` files, all 622 files in the Soufflé
   source tree (structural scan only, no execution), and the five benchmark
   shapes. Report counts.
3. Hand-construct **five to eight additional source programs** with this
   structure, varying: cycle length, position of the negated literal in the
   body, arity, whether the negated predicate is EDB or IDB, and whether the
   query is bound on one or both arguments. Each with a seeded fixture
   generator in `fixtures_lib.py`.
4. For each, record Soufflé's behaviour: untransformed result, result under
   `--magic-transform=*`, whether any relation is skipped, and the tuple totals.

**Do not adorn any of them.** Do not write a general adorned form. Deriving the
adorned program is the transform, and the transform is Lane A. If a task step
seems to need it, you have misread the step.

**Gate:** number of structurally-matching programs found in each corpus;
number of new programs constructed; Soufflé behaviour recorded for each.

**Deliverable:** `harness/night03_t4_culprit_classifier.py`,
`tests/corpus/CULPRIT_CANDIDATES/`, `docs/reports/night03-T4-culprit-corpus.md`.

---

## T5 — The `ancestor_nonancestor` seed prediction: pre-register, then test

**Why:** V4 established that `ancestor_nonancestor_guarded.dl`'s `T_guard` grows
2.00× per doubling of `n` while its seed `m_ancestor` stays pinned at 50, and
that `nonancestor_bf` is 89% of the total. The diagnosis is that the file
propagates its seed on a recursion where the bound argument never moves —
`ancestor(x,y) :- ancestor(x,z), parent(z,y)` is first-argument-invariant,
exactly like `reach` in `p4prime.dl`, whose own header comment states the rule
correctly and uses a static seed.

**Order matters. Do these in sequence and commit between steps.**

### Step 1 — pre-register, before running anything

Append to `docs/OPEN_QUESTIONS.md`, verbatim, with today's date:

> **Q11.** `ancestor_nonancestor_guarded.dl` propagates `m_ancestor` across a
> recursion whose bound argument is invariant, and derives `nonancestor`'s
> restriction from `m_ancestor` rather than seeding it from the query.
> Prediction, recorded before measurement: a variant seeding `m_nonancestor(1).`
> directly gives `T_guard ≈ 3,200` at n=500 and `≈ 18,200` at n=8,000, i.e.
> `T_souffle/T_guard ≈ 32×` and `≈ 1,400×` respectively, moving this shape into
> the same band as `reachability_complement` and `same_generation_negation` and
> dissolving Q8.

**Commit this before step 2.** A prediction recorded after the measurement is
not a prediction.

### Step 2 — build and measure the variant

Create `tests/corpus/BENCHMARK_FAMILY/guarded/ancestor_nonancestor_guarded_v2.dl`.
**Do not modify the original** — it is committed measurement history and the
`p4`/`p4prime` precedent applies: both files survive, both are reported.

Run it at every `ancestor_nonancestor` scale point, same fixtures, unregenerated.

**Gate:** `T_guard` per scale point for v2; answer set-equality against v1 and
against the untransformed baseline at every point; measured vs predicted, stated
as both numbers with no smoothing.

If the prediction is wrong, that is a result and reported as one. Do not adjust
the prediction. Do not construct a third variant to make a number come out.

**Deliverable:** the v2 file, `docs/reports/night03-T5-ancestor-seed.md`, rows
appended to `docs/MEASUREMENTS.md` in the standard format.

---

## T6 — Grammar amendment 2: optional parens on `.input` / `.output`

**Authorized.** `m1-progress.md` attributes 105 of 175 parse failures (60%) to
`.input Name()` / `.output Name()` with cosmetic trailing parens. This is the
same production as `.input ident` with syntax noise — no semantics, no new
construct — and is the identical class as the zero-arity amendment
`M1-BUILD.md` already authorized.

**Do:**

1. Implement: `.input`/`.output` accept an optional parenthesised list, which is
   parsed and discarded. Unit test it.
2. Re-run §3.3 gate one and gate two against `IN_GRAMMAR.txt`.
3. Report **four** numbers: total parsed before, total parsed after,
   negation-bearing files parsed before, negation-bearing files parsed after.
   The last two matter most — M3's differential corpus needs negation.

**Mechanical decision rule, apply it yourself, no human needed:**

- Keep the amendment if `total parsed ≥ 80` **or** `negation-bearing parsed ≥ 5`.
- Otherwise revert the amendment commit, and keep the measurement report.

Report the rule's inputs and its outcome either way.

**Do not touch the corpus predicate in this task.** T7 does that. Two variables,
two measurements, two commits.

**Deliverable:** parser change (or its revert), `docs/reports/night03-T6-parens.md`.

---

## T7 — Corrected corpus predicate

**Why:** VERIFY-01 established the true blueprint-§4-compliant count is 19/195,
and that `night01_t5_grammar.py` checks 12 construct categories while §4 implies
at least nine more. The exclusion criterion — blueprint §4's grammar — was fixed
in advance; the predicate failed to implement it. That is a bug fix, on the same
grounds as the `corpus_predicate.py` sort fix, and does not violate the
pre-registration rule.

**Do:**

1. Write `harness/night03_t7_grammar_v2.py` — a **new** file. Justify each
   exclusion factor by reference to a blueprint §4 production it violates, in a
   comment on that check. Never by reference to which files failed.
2. Categories to add, from VERIFY-01 §V3: zero-arity decls, multi-name
   `.decl A,B,C(...)`, `.decl` qualifiers beyond the three checked, bare and
   built-in functor calls used as terms, comma-separated multi-head rules,
   directives beyond the three §4 admits, `#include`, extra primitive types
   (`unsigned`/`float`), `.input`/`.output` parens, string escape sequences.
   Add any further category §4 excludes that you find.
3. Emit `tests/corpus/IN_GRAMMAR_V2.txt` over the full 622-file source tree.
4. Cross-check against `dlc`'s parser: every file in V2 must parse; report
   `parsed/|V2|`. Two independent methods agreeing is the validation.

**Both lists survive.** `IN_GRAMMAR.txt` is untouched and stays committed. Every
gate that used it must from now on report against both, exactly as the project
kept 11-of-149 alongside 36-of-612.

**Gate:** `|V2|`, `parsed/|V2|`, and the per-category exclusion histogram.

**Deliverable:** the predicate, `IN_GRAMMAR_V2.txt`, `SOURCE_V2.md`,
`docs/reports/night03-T7-grammar-v2.md`.

---

## T8 — Codegen symbol-ordering gap

**Why:** `m1-progress.md` flags it under "what a skeptic attacks first" —
`<`, `<=`, `>`, `>=` on `symbol` columns compare interned ids, not string
values, and nothing currently run through codegen exercises it. Untested is not
verified.

**Do:** write generate→compile→run tests covering ordering comparisons on symbol
columns, including cases where interning order and lexicographic order differ
(intern `"zebra"` before `"apple"`). Compare against Soufflé on the same
program. If it is wrong, fix it — `codegen/` is Lane B.

**Gate:** `cases/cases` passing, and an explicit statement of whether a defect
was found. If none, say "14 cases, zero defects" — do not report silence.

**Deliverable:** tests in `codegen_test.go`, `docs/reports/night03-T8-symbol-order.md`.

---

## T9 — Fallback cone metric

**Why:** M3's blast-radius integer is the project's only number that a
hand-transform cannot produce. It is a cone size over the dependency graph, not
an SCC count — one declined component drags its downward dependency closure with
it, confirmed empirically in `night02-T7` check 3.

**Scope, precisely:** the **decision** of which SCCs to decline is Lane A. The
**measurement**, given a declined set, is a graph query and is Lane B.

Implement in `harness/` (not `src/`) a function:

```
cone_size(program, declined_sccs) -> {
    declined_sccs: int,
    cone_relations: [names],
    cone_size: int,
    cone_fraction: float   # of all IDB relations
}
```

where the cone is the downward dependency closure of the declined set over the
**full** dependency relation, positive and negative edges alike.

Validate on `culprit_cycle.dl` with `{p}` declined: the cone must contain `q`
and `s`, matching the observed full-extent relations in both the automatic and
hand-guarded runs. Report the number for all five shapes under a hypothetical
declined set of each shape's own negation-bearing SCC.

**Gate:** validated against `culprit_cycle`'s observed behaviour; numbers
reported for all five shapes.

**Deliverable:** `harness/cone_metric.py`, `docs/reports/night03-T9-cone.md`.

---

## T10 — Provenance backfill

VERIFY-01 §V7: M1's numbers are real `dlc` output but were never appended to
`docs/MEASUREMENTS.md` and use a flat JSON format rather than the
`cmd.txt`/`stdout.txt`/`env.txt` triple.

**Do not retrofit per-invocation triples.** That is busywork.

1. Append **one row per gate** to `docs/MEASUREMENTS.md` — approximately fifteen
   rows — each citing its `measurements/m1-*.json` file and its integer.
2. Append to `docs/DECISIONS.md`: flat per-gate JSON is an accepted provenance
   format for `dlc`-produced numbers, alongside the `cmd.txt` triple for Soufflé
   runs. Reason: `dlc` runs are reproducible from committed source plus a
   committed harness; Soufflé runs are not.
3. Append to `docs/DECISIONS.md`: Q7 closed — no rubric-mandated compiler phases;
   assessment is a presentation of what the project does. Consequence: C codegen
   carries no requirement, drops to the bottom of M4 below M3 overflow. Weeks
   10–12 are M3 overflow, released to M4 only if M3 closes on time.
4. Append the T8 zero-arity correction: `night02-T8` reported 11 zero-arity
   files; VERIFY-01 found 12 (`syntactic/doc_comment_dangling3/other.dl`).

---

## T11 — De-stale the project state document

The state document asserts `src/` is empty and M1 has not started. Both were
false on the day it was written.

Edit it to stop asserting volatile state: replace the `src/`-state and
milestone-status assertions with a pointer to `docs/SESSION_LOG.md` and
`docs/reports/m1-progress.md`. Leave everything non-volatile — thesis, filters,
grammar, metric definitions, prior art, dead ends, process rules — untouched.

Do not create a new document. Do not bump a version number on anything else.

---

## Task summary

| T | Task | Blocks | Drop if short on time |
|---|---|---|---|
| T1 | Printer → Soufflé | T2, all of M3 measurement | never |
| T2 | M3 protocol on PassThrough | M3 measurement | never |
| T3 | M2 acceptance harness | M2 velocity | never |
| T4 | Culprit-cycle corpus | M3 guard testing (currently n=1) | never |
| T5 | Ancestor seed prediction | Q8 | last resort |
| T6 | Parens amendment | corpus size | yes |
| T7 | Corrected predicate | corpus honesty | yes |
| T8 | Symbol ordering | a known untested gap | yes |
| T9 | Cone metric | M3 headline definition | yes |
| T10 | Provenance backfill | nothing | yes |
| T11 | De-stale state doc | nothing | yes |

---

## What this batch explicitly does not do

- **No Lane A code.** M2 and M3 are the human's, and this batch does not begin
  them. Every task here removes an obstacle in front of that work; none of them
  is that work.
- **No demo, no frontend, no visualisation.** Parked until the guard produces
  its first blast-radius integer. Nothing on screen tonight would show anything
  a hand-written `.dl` file did not already show in August.
- **No new corpus hunt.** Closed. Three independent sources agree.
- **No new specification documents.** This file is the last one until M2 exists.