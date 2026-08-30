# Specs — index

The build plans this project actually ran against — evidence the
discipline described in `docs/03-methodology.md` was real, not asserted
after the fact. Renamed by content per `docs/rename-map.csv`; content
unedited from what directed the corresponding work.

| # | File | What it directed | When | What it produced |
|---|---|---|---|---|
| 01 | `01-agent-contract.md` | The operating contract every other spec below operates under — lane boundaries, escalation protocol, provenance discipline | Archived snapshot, 2026-08-30 | This restructure follows it; the **live** copy stays at the repo root (`CLAUDE.md`) and keeps evolving there — this archived copy is not kept in sync automatically |
| 02 | `02-m1-build.md` | Front end + first optimization: lexer, parser, semantic checks, naive/semi-naive evaluators, one authorized grammar amendment | Weeks 1–3 | `experiments/31-m1-front-end-and-evaluators.md` and everything M1-tagged in `experiments/` |
| 03 | `03-night-batch-03.md` | An 11-task unattended overnight batch — printer/Soufflé round-trip, M2 harness, culprit-cycle corpus, grammar amendment 2, fallback cone metric | One session | `experiments/33` through `experiments/42` |
| 04 | `04-m2-m3-build.md` | The magic-set transform and the guard — adornment, SIPS, supplementary predicates, the two-clause soundness condition, fallback evaluation | Weeks 4–11 | `experiments/43` through `experiments/48` |
| 05 | `05-night-batch-04.md` | An 8-task unattended overnight batch — M4 gate + headline re-run, cone corpus construction, reconciliations, `dlc explain`, whole-project report | One session | `experiments/49` through `experiments/57` |
| 06 | `06-m4-sips.md` | Demand relaxation on negated occurrences — two changes and one re-run, deliberately small | One session | `experiments/49-demand-relaxation.md` |
| 07 | `07-punch-list.md` | Six items closing out the project: multi-query seeding fix, a documentation consistency pass, three reframings, and this presentation artifact | One session | `experiments/53`, `experiments/54`, `docs/reports` consistency edits folded into `experiments/57` |

**No `03-verify-01.md` exists.** The user's original numbering for this
directory included a spec for the verification pass now at
`experiments/32-verification-pass.md`; no separate directing document for
that pass exists — it was Lane B initiative, self-directed rather than
run against a written build plan. Rather than commit an empty placeholder
or leave a numbering gap, `specs/` is numbered 01–07 without a gap, and
this row records why.

**The four-items follow-up work** (`experiments/55`, `56`, and the second
punch-list documentation pass folded into `experiments/57`) also has no
directing spec file here — it was given as a direct instruction, not
written down as a standalone build plan first.
