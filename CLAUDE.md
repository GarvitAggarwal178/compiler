# Working on dlc

This is the current repository working guide. The former milestone and authorship
rules are historical: see `specs/01-agent-contract.md`, the later ruling in
`specs/04-m2-m3-build.md`, and Git history. The compiler's core code received an
authorized AI-assisted maintenance pass; do not describe those changes as
exclusively human-authored.

Read `README.md` for setup and commands, `docs/02-design.md` for the supported
language and pipeline, `docs/05-limitations.md` for boundaries, and
`docs/07-pass1-review.md` for the maintenance record. `docs/06-reproduce.md`
describes current verification and the historical harness caveats.

Keep the source language and tuple-count conventions stable unless the user asks
to change them. Preserve `measurements/` raw evidence behind reported claims.
Do not treat tuple reductions as runtime speedups or differential tests as a
formal proof. Generated C currently evaluates by naive fixed point.

Keep generated artifacts out of the checkout, except the documented
`measurements/` evidence. Use `make verify` for build, vet and tests, and
`make oracle` for Souffle 2.5 comparisons. Report unavailable checks and their
limits explicitly. `src/` contains compiler code, `tests/` regression programs,
`fixtures/` inputs, `harness/` verification scripts, and `experiments/`,
`specs/`, `record/`, and `results/` dated historical material.

The 20 October review handoff authorizes phase A repairs/inspection, then the
runnable phase B local demo, then conservative phase C guarded-compact. Preserve
guarded behavior except correctness repairs. Current demo/review docs are
`docs/08-demo.md` and `docs/09-review2.md`. The shared planner is in `src/plan`;
compiler-owned reporting is in `src/inspect` and `src/transform/report.go`.
Use a fresh evidence run directory. No aggressive projection-removing inlining,
new language features, pushes, or branch/remote changes are authorized.

The review phases are implemented; `docs/10-final-review.md` documents the
production conservative pass and observed final evidence. `make oracle-compact`
verifies that mode separately. The workbench remains available through `make demo`.
Keep engine-specific counts labelled and source/candidate/final reports distinct.
