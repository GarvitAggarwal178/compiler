# Restructure notes — 2026-08-30

What moved where, every reference fixed, what was archived (nothing), and
the seven verification checks, with results. Tag `pre-restructure` marks
the commit immediately before this work started.

## What moved where

Full mapping: `docs/rename-map.csv` (84 rows — every tracked documentation
file, plus the 6 planning docs this session had deliberately left
untracked, each now added to git as part of the move it's listed under).
Summary:

- **57 experiment reports** (`docs/reports/*.md` + `docs/NIGHT_LOG.md`) →
  `experiments/01..57-<content-name>.md`, chronological by git-add
  timestamp with three documented exceptions (see the CSV's `reason`
  column for `experiments/14`, `15`, `49`, `57`).
- **7 build plans** (`M1-BUILD.md` + 5 previously-untracked planning docs)
  → `specs/02..07-<content-name>.md`; `CLAUDE.md` copied (not moved) to
  `specs/01-agent-contract.md` as an archived snapshot — the live file
  stays at the repo root.
- **5 append-only files** (`DECISIONS.md`, `ESCALATIONS.md`,
  `MEASUREMENTS.md`, `OPEN_QUESTIONS.md`, `SESSION_LOG.md`) → `record/`,
  moved whole, content untouched.
- **The presentation artifact** (`docs/reports/presentation.html`) →
  `results/presentation.html`.
- **The blueprint** (`docs/dlc-blueprint.md`) → moved and rewritten as
  `docs/02-design.md` (current-state document); its version history
  extracted to `docs/design-history.md`.
- **New documents**: `README.md`, `docs/01-problem.md`,
  `docs/03-methodology.md`, `docs/04-related-work.md`,
  `docs/05-limitations.md`, `docs/06-reproduce.md`, `docs/project-log.md`,
  `results/findings.md`, `results/claims.md`, `results/superseded.md`,
  four directory `README.md` indexes, and this file. `Makefile` added at
  the root (build/vet/test/rejection/gates targets).

`src/`, `harness/`, `tests/`, `fixtures/`, `measurements/`, `tools/`,
`bin/`, `go.mod` were never moved — confirmed by `git diff --stat
pre-restructure..HEAD -- src/ harness/ tests/ fixtures/ measurements/`
showing zero changes to any of those trees (only `tools/` gained two new
scripts: `apply_rename_map.py`, `fix_design_md_refs.py`).

## What did not work

- **Phase 2's size gate fired but the fix it implies doesn't apply here.**
  The repo measured 243.8 MB, over the 100 MB threshold, but 176.8 MB of
  that is `prof.log` — explicitly the source of every reported integer,
  explicitly excluded from archiving. The only permitted archive target
  (raw `.csv`, 28.35 MB) would have left the repo at ~215 MB, not relieved
  the condition the gate exists to catch, and would have deleted the
  oracle-comparison output every set-equality check in this project was
  verified against, for zero size benefit (already pushed, so `git rm`
  shrinks the working tree only, not the pushed history). Put to the user
  directly rather than archived silently; **archived nothing**, per the
  user's own choice.
- **The pre-restructure `pre-restructure` tag needed a second look.** It
  was created before this session's Phase 0 content edits, and — in an
  unrelated but concurrent request this same session — every commit's
  message was rewritten to drop a `Co-Authored-By` trailer, which moved
  every commit hash including the tagged one. The tag was moved to match
  and force-pushed; documented here so a reader comparing against
  `pre-restructure` knows its target commit's hash changed once, for a
  reason unrelated to the restructure itself.
- **`experiments/README.md`'s question/answer summaries for the ~32
  reports not named in the original request are paraphrased from each
  report's own header and content, not quoted.** A skeptic should treat
  the index as a navigation aid and read the cited report directly before
  citing the index's own wording as a claim.

## Cross-references: what was fixed, and what was deliberately left alone

**Fixed** (functional paths and live navigational documents): every path
in `README.md`, `docs/*.md`, `results/*.md`, the four directory
`README.md` indexes, `CLAUDE.md`, `harness/build_presentation.py`
(including 4 citations embedded in its generated HTML output, plus its
`OUT_PATH`/`EXPLAIN_DIR` constants), `src/**/DESIGN.md` (11 files, via
`tools/fix_design_md_refs.py`), and `tests/corpus/SOURCE.md`,
`SOURCE_V2.md`, `BENCHMARK_FAMILY/README.md`.

**Deliberately not touched, each for a stated reason:**

- `record/*.md` — append-only, never rewritten. Their internal citations
  to old `docs/reports/*.md` paths are translated via `docs/rename-map.csv`,
  documented in `record/README.md`, not rewritten in place.
- `experiments/*.md` — reports move and are renamed, never edited.
  Several cite other reports by their old filename; a reader follows
  `docs/rename-map.csv` or `experiments/README.md` to the current name.
- `specs/*.md` — build plans, treated the same way as reports for the
  same reason (`specs/README.md` states this explicitly): they are
  historical directing documents, evidence of what was asked at the time,
  not living documentation.
- 14 `.go` source files' code comments citing old report paths — these
  are code comments, not documentation cross-references, and outside
  this phase's named scope (`harness/build_presentation.py` and
  `src/**/DESIGN.md` specifically, not every `.go` comment).
- `.dl`/`.py`/`.json`/`.txt` test fixture files whose header comments cite
  old paths — fixture content is frozen for determinism; only
  documentation files were in scope.
- `harness/*.py` module docstrings (other than `build_presentation.py`)
  citing old report/spec paths in their "why this exists" preamble — these
  are historical citations describing what motivated the script when it
  was written, analogous to a commit message; the paths they cite were
  correct at the time.

## The seven checks

1. **`go build ./...`, `go vet ./...`, `go test ./...`** — all clean, run
   against the real toolchain (Go 1.26.0, native WSL2 Ubuntu, not the
   Windows-side git-bash shim). Zero errors, zero vet findings, every
   package's tests pass.
2. **Four rejection grounds, 13/13.** `harness/night02_t9_diagnostics.py`
   re-run: 14/14 reject as expected (the 13 rejection-corpus cases plus
   one additional `undeclared_relation` case in the same suite), 0
   accept. (`harness/run_rejection_tests.py`, an older scaffold script
   from before `dlc`'s rejection logic existed, reports the same 13 cases
   as "UNEXPECTED" — its hardcoded expectation is that `dlc` has *not yet*
   implemented rejection, a pre-M1 negative control. It has not been run
   since M1 landed and was not treated as authoritative here;
   `night02_t9_diagnostics.py` is the report `results/claims.md` and
   `experiments/28` actually cite.)
3. **One headline measurement re-run, byte-identical.**
   `harness/punchlist2_item2_growing_sibling.py` re-run in place, its
   output backed up first: `measurements/punch-list-2/item2-growing-
   sibling/summary.json` reproduced **byte-for-byte identical** —
   `diff` returned no output. The three `prof.log` sets underneath it
   changed (only in `loadtime` nanosecond timestamps — every `num-tuples`
   field is unchanged), and re-running the rejection check in step 2 also
   touched several `night02-t9-*/env.txt` files' `wall_seconds_not_a_metric`
   line — both are the explicitly-disclaimed non-metric timing fields
   this project's own methodology (`docs/03-methodology.md`) says carry no
   evidentiary weight; every measured integer is unchanged.
4. **No dangling citations.** Verified by a script checking that every
   backtick-quoted path-like string in `README.md`, `docs/*.md`,
   `results/*.md`, the four directory indexes, `CLAUDE.md`, and
   `src/**/DESIGN.md` resolves to an existing file (bare filenames inside
   a directory's own index table resolve relative to that directory,
   matching how the tables are written). Two non-path prose mentions
   (`p2.dl`, `cc_growing_sibling.dl` — fixture names, not literal
   root-relative paths) are not dangling references; `IN_GRAMMAR.txt`
   and similar bare corpus-file mentions are the same pattern.
5. **Fresh-clone test.** `git clone` to `/tmp/clone-test`: `docs/06-
   reproduce.md` present, `go build ./...` succeeded with no setup beyond
   what that document describes.
6. **Read-the-README test, honestly.** From `README.md` alone: what was
   built (a Datalog compiler with a guarded magic-set transform) is
   stated in two sentences; the headline number (46.0×–1,342.7× over
   Soufflé's own transform) and the arc (194×→5,317×→parity) are both
   present with citations; the reading-order table points at
   `results/findings.md` and `results/claims.md` for evidence. What it
   does **not** hand the reader directly is *how much of that evidence is
   on constructed vs. real-world programs* — that distinction is one
   click away (`docs/05-limitations.md`, item 1) but not in the README
   itself, by design (the README is capped under 400 words and the
   limitations file exists precisely so the README doesn't have to carry
   every caveat).
7. **Repo size, before and after, same measurement tool
   (PowerShell `Get-ChildItem -Recurse`, since the WSL9p-mounted `du`
   proved inconsistent between runs on this filesystem — noted here so a
   future measurement doesn't trust `du` on this mount without
   cross-checking):**

   | | Before (`pre-restructure`) | After |
   |---|---|---|
   | Total (working tree + `.git`) | 243.80 MB | 239.66 MB |
   | `measurements/` | 207.52 MB | 207.51 MB |
   | `.git` | 26.67 MB (loose, ~5,000 objects) | 22.47 MB (packed, 1 packfile) |

   `measurements/` is unchanged (confirming Phase 2's "archive nothing"
   and confirming step 3's reproduction added no net bytes). The `.git`
   size drop is from the same-session commit-message rewrite's `git gc`
   (unrelated to this restructure's own content), not from anything moved
   here. Total working-tree content (excluding `.git`) grew by
   documentation added in Phases 4–6, on the order of tens of KB — a
   rounding-level change against a 217 MB base.
