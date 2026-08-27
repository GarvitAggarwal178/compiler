# NIGHT-BATCH-04 D — four reconciliations

Date: 2026-08-27.

## What did not work

Nothing new broke in this task — all four items were audits/reconciliations
of already-committed reports. One (D1) did require a real fix (a missing
predicate category); the other three were mis-stated *comparisons*
between two already-correct numbers, not defects in either number itself.

## D1 — T7's 2 residual cross-check gaps: identified, fixed

`night03-T7-grammar-v2.md` already **identified** both files exactly
(`syntactic/issue2408/issue2408.dl` = the single line `x`,
`syntactic/syntax10/syntax10.dl` = a leading bare `*`) and disclosed why
neither the predicate's existing 23 categories nor a "quick fix" applied:
both are Soufflé's own deliberately-broken error-location test fixtures,
not an attempt at any real construct, so no NAMED grammar-production
violation matches them.

**Verdict: neither the predicate nor the parser was wrong** in the sense
of producing an incorrect answer on any specific check — `dlc`'s parser
correctly rejects both (they are not valid programs), and the 23-category
predicate correctly found no NAMED violation (there isn't one to name).
The gap was a missing category: "matches no violation" was being treated
as "in-grammar," which is wrong for content that also matches no *valid
construct*.

**Fixed**, not left disclosed-only: added a 24th category,
`no_top_level_construct` — a non-whitespace file (after comment-stripping)
containing zero `.` characters. This is not an empirical pattern-match
against these two files; it is a direct consequence of blueprint §4's own
grammar (`program ::= decl* clause*`): every `decl` (`.decl`/`.input`/
`.output`/`.include`) and every `clause` contains at least one literal `.`
token, so non-empty content with none cannot be assembled from any decl
or clause. A genuinely empty/comment-only file is explicitly excluded
from this check (`decl* clause*` permits zero repetitions — an empty
program is grammatically valid and `dlc` does parse it), so this cannot
misfire on a legitimately trivial file.

**Result**: `harness/night03_t7_grammar_v2.py` re-run over the full
622-file tree (`measurements/night03-t7/summary.json`,
`measurements/night03-t7/detail.json`): `|V2| = 19` (was 21),
**cross-check 19/19 = 100%** (was 19/21 = 90%). Confirmed the new
category flags *exactly* the intended 2 files plus one already-excluded
file (`syntactic/unterminated_comment/unterminated_comment.dl`, which was
already excluded via the pre-existing `unterminated_block_comment`
category — its comment-swallowed content also has zero periods, but this
does not change the admitted count since it was never admitted). Net
change: exactly −2, both intended.

## D2 — T7 (21/622) vs VERIFY-01 §V3 (≤19): resolved, no contradiction remained

Both numbers were already reconciled *in prose* by `night03-T7-grammar-v2.md`'s
own "VERIFY-01 gap, named explicitly" section, which named the 2 extra
V2 members as exactly the residual garbage-fixture pair and argued
VERIFY-01 would also have excluded them. This session verified that
claim directly rather than trusting the argument: VERIFY-01 §V3's own
category table (`docs/reports/verify-01.md`) lists a category named
**"deliberately-broken fixture content"** whose two named examples are —
verbatim — `syntactic/syntax10/syntax10.dl` and
`syntactic/issue2408/issue2408.dl`, i.e. VERIFY-01 independently found
and excluded the identical 2 files by direct read, not by the argument
T7 constructed after the fact.

**Verdict: no contradiction ever existed** — both methods agree the true
count is **19**. T7's number is now 19/622 exactly (D1's fix), matching
VERIFY-01's ≤19 bound exactly rather than merely "being consistent with"
it. `docs/reports/night03-T7-grammar-v2.md` is left as originally written
(its own reconciliation argument was correct, just not yet verified
against VERIFY-01's primary evidence, and the predicate itself is now
fixed rather than merely argued-around) — this report is the update of
record.

## D3 — M2 count: "5/6" (`m2-headline.md`) vs "5/5" (`SESSION_LOG.md`)

**Both numbers are correct; they measure different things and were never
in conflict.** `m2-headline.md`'s own sentence states both explicitly:
"4/5 shapes + `p2.dl` (**5/6 total comparable cases**): **5/5**
answer-identical" (line 93). Six programs were tested (5
`BENCHMARK_FAMILY` shapes + `p2.dl`); `culprit_cycle` is excluded as
not-comparable (its mechanical transform is unstratifiable — the
guard's designed-for signal, not a defect), leaving 5 comparable; of
those 5, all 5 are answer-identical. "5/6" = comparable⁄total. "5/5" =
identical⁄comparable. `SESSION_LOG.md`'s "5/5 comparable cases
answer-identical" is a terser, equally correct restatement of the same
fact, dropping the total-of-6 context rather than contradicting it.

**Verdict: no correction needed to either document.** Named here so the
apparent conflict does not resurface as a real discrepancy in the final
report — `FINAL.md` should state both numbers together, exactly as
`m2-headline.md` already does, to avoid re-introducing the ambiguity.

## D4 — Q8 status

Recorded explicitly in `docs/OPEN_QUESTIONS.md` ("Q8, CLOSED"), cross-
referencing `docs/reports/m4-sips.md`'s direct v1-vs-`dlc` comparison —
see that entry and that report for the numbers. Verdict: **v1 (hand
guard) is suboptimal**, not the shape's inherent contribution ceiling.

## Verdict

**D: DONE**, all four items. D1 was the only one requiring a code change
(one new, grammar-derived category, not a corpus-specific patch);
19/19 = 100% cross-check confirms it. D2 and D3 required no code or
document changes — both apparent discrepancies were already-correct
numbers describing different things, verified rather than assumed.
