# NIGHT-BATCH-03 T3 — M2 acceptance harness and Soufflé reference transforms

Date: 2026-08-27. T3b (`harness/m2_accept.py`) is the gate for every M2/M3
work item from here on — built and validated before any transform code was
written, per instruction.

## What did not work

Nothing. `--show=transformed-datalog`, the option name NIGHT-BATCH-03.md
speculatively named, **does not exist** in Soufflé 2.5 (`souffle --help`
lists no such mode). The correct mode is `--show=transformed-ast`, confirmed
by direct invocation before any reference file was generated — exactly the
verification T3a asked for ("confirm the exact option name first ... rather
than assume").

## T3a — Soufflé's own transformed programs, as reference

`harness/night03_t3a_reference.py`. For the 5 `BENCHMARK_FAMILY` shapes plus
`p1prime.dl`, `p2.dl`, `p4prime.dl` (8 files — `culprit_cycle.dl` has no
separate `tests/programs/` copy, so it is not double-counted): dumped
`souffle --show=transformed-ast <file>` (plain) and
`souffle --show=transformed-ast --magic-transform=* <file>` (magic), both to
`tests/reference/souffle-transformed/<name>.{plain,magic}.dl`.

**8/8 dumped cleanly, return code 0, no errors, both configurations.**
`tests/reference/souffle-transformed/MANIFEST.md` records the exact command
and the confirmed option name. These are reference material only — no
syntactic diff is built against them (Soufflé's `@magic.`, `@neglabel.`,
`@poscopy_1` naming and its inlining make a diff worthless, per T3a's own
instruction); their value is that a human can read what a correct transform
of a known program looks like, e.g. `culprit_cycle.magic.dl` shows Soufflé's
own `@poscopy_1.q.{ff}` duplication strategy referenced in M2-M3-BUILD §7.

## T3b — `harness/m2_accept.py`

Input: an original `.dl`, a candidate transformed `.dl`, a fixture
directory. Runs both standalone against the same facts (`-F`), compares
every `.output` relation by sorted set equality, and reports exact tuple
totals via `harness/tuple_report.py`'s Soufflé-profile parser
(`T_excl_copy`), per-relation breakdown for both, and the plain ratio
`T_original / T_candidate` — labeled exactly that, not "contribution": this
script has no way to know whether `original` is a plain baseline (`T_none`)
or an already-transformed dump (`T_souffle`), so it never assumes a
three-column slot for either input. That labeling decision belongs to
whoever calls it and writes a report (M2-M3-BUILD §4 does this explicitly
for `dlc`'s own transform).

### Validation: three known-good pairs, numbers vs. committed history

| pair | fixture | `T_original` (harness) | committed | `T_candidate` (harness) | committed | answers identical |
|---|---|---|---|---|---|---|
| `p2.dl` / `p4prime.dl` | `fixtures/p2-scale-250` | 62,534 | `T_none`=62,534 (`docs/reports/night01-T6-scaling.md`) | 285 | `T_guard`=285 (same table) | yes |
| `ancestor_nonancestor.dl` / `..._guarded.dl` | `fixtures/benchmark-family/ancestor_nonancestor/n500` | 250,450 | `T_none`=250,450 (`docs/reports/night02-T5-guarded.md`) | 25,500 | `T_guard`=25,500 (same table) | yes |
| `same_generation_negation.dl` / `..._guarded.dl` | `fixtures/benchmark-family/same_generation_negation/d4_b4` | 51,301 | `T_none`=51,301 (`docs/reports/night02-T5-guarded.md`) | 452 | `T_guard`=452 (same table) | yes |

**All three: exact match, to the integer, against already-committed
measurement history.** 3/3 answer-set-equal. Provenance:
`measurements/night03-t3/{p2-p4prime,ancestor,same-gen}.json` (raw
`m2_accept.py` output, one file per pair).

## What a skeptic attacks first

- The harness's correctness rests on `tuple_report.analyze`'s `T_excl_copy`
  definition, which is reused unmodified from earlier sessions and was
  itself validated against these same numbers before — this is not an
  independent cross-check of `tuple_report.py`, it is confirmation that
  `m2_accept.py` calls it correctly. A bug shared between `tuple_report.py`
  and this script would not be caught by this validation.
- Only 3 pairs were checked, all pre-existing and all already known-correct.
  The harness has not yet been exercised on a genuinely novel candidate
  (that happens starting at M2.3 / §4's real gate).
- The ratio field is deliberately unlabeled as `T_none`/`T_souffle`/`T_dlc`
  to stay generic — this pushes a correctness-relevant labeling decision
  onto every caller, which is a real place a future report could
  mislabel a column if the caller is not careful. Documented in the
  script's own docstring as a deliberate tradeoff, not an oversight.

## Verdict

**T3: DONE, gate cleared.** `harness/m2_accept.py` reproduces every
committed number exactly on all 3 validation pairs — it is trusted as M2's
grading harness from here on. `tests/reference/souffle-transformed/` is
committed (8 files + `MANIFEST.md`) as human-readable reference material,
not a golden target.
