# M3.2 — culprit-cycle detection, clause (a) (§6)

Date: 2026-08-27. `src/transform/guard/stratify.go`.

## What did not work

Nothing. The Go implementation (`sema.CheckStratification` run on the
candidate transformed AST) and the independent Soufflé cross-check agreed
on the first run, on all 7 programs.

## Method

**Cheap precondition:** `HasPositiveCycle`, a dedicated positive-edges-only
precedence graph over `prog.Clauses`, O(V+E). **The check:**
`CheckCulpritCycle` builds the candidate transformed program (`magicset.
Adorn`+`Generate`) and runs `sema.CheckStratification` on it directly — the
pre-existing source stratifier, applied to the transformed AST.
**Differential oracle:** `harness/night_m3_2_culprit_detection.py` prints
the transformed program, gets `dlc check`'s verdict, and separately feeds
the identical text to real Soufflé — not self-consistency testing, since
Soufflé's stratifier is independent code asked the same question.

6 Go unit tests (`src/transform/guard/stratify_test.go`):
`HasPositiveCycle` correctly identifies `q`/`p` (self-recursive) as
cyclic and `s`/`out` (non-recursive) as not, on the `culprit_cycle` shape;
`CheckCulpritCycle` correctly flags `culprit_cycle`'s mechanical
transform as unstratifiable, correctly accepts `ancestor_nonancestor`'s,
and correctly treats a no-bindable-query program as a trivially
stratifiable no-op.

## The gate

Provenance: `measurements/m3-2-culprit-detection/summary.json`. Every
program in `tests/corpus/CULPRIT_CANDIDATES/` (6) plus
`tests/corpus/BENCHMARK_FAMILY/culprit_cycle.dl` (1) — 7 total.

| program | `dlc` verdict | Soufflé verdict | agree |
|---|---|---|---|
| `cc_arity3_twobound` | unstratifiable | unstratifiable | yes |
| `cc_edb_negated` | **stratifiable** | **stratifiable** | yes |
| `cc_longer_cycle` | unstratifiable | unstratifiable | yes |
| `cc_neg_early` | unstratifiable | unstratifiable | yes |
| `cc_query_bothbound` | unstratifiable | unstratifiable | yes |
| `cc_third_relation` | unstratifiable | unstratifiable | yes |
| `culprit_cycle` | unstratifiable | unstratifiable | yes |

**7/7 agreed, 0 disagreements.** Every disagreement would have been
reported individually per instruction; there were none to report.

## What a skeptic attacks first

- `dlc check`'s verdict is derived from the exact same `sema.
  CheckStratification` call `CheckCulpritCycle` itself uses internally —
  the Go-level unit tests and the `dlc check`-based harness numbers are
  not two independent confirmations of `dlc`'s own logic, only one. The
  genuinely independent confirmation is the Soufflé cross-check, which is
  why it is the gate, not the Go tests.
- All 7 programs in this gate are drawn from a corpus (`CULPRIT_CANDIDATES`
  + `culprit_cycle.dl`) that was specifically constructed or selected to
  be relevant to this exact mechanism — this gate says nothing about the
  false-positive rate on a program the mechanism has no opinion about
  (that evidence is M3.1's counterexample search and M2's own 5/6 gate,
  both already reporting 0 unexpected rejections).
- `HasPositiveCycle`'s own correctness rests on 2 targeted assertions
  (`q`/`p` cyclic, `s`/`out` not) on one shape plus one non-recursive
  program exercising the precondition's fast-path skip
  (`TestCheckCulpritCyclePreconditionSkipsNonRecursiveProgram`) — not
  exercised on the other 4 `BENCHMARK_FAMILY` shapes or the 5 other
  `CULPRIT_CANDIDATES` constructions directly. `CheckCulpritCycle` does
  call it as a short-circuit (skip the full transform-and-check path when
  no relation anywhere in the source has a positive cycle) — wired in
  during this task after being caught missing in an earlier draft.

## Verdict

**M3.2: DONE.** 7/7 agreed between `dlc` and Soufflé, 0 disagreements.
Proceeding to M3.3 (per-SCC decision + fallback cone).
