# Session log

Append-only, one entry per working session.

## 2026-08-20 — Phase 0 probe

**What changed:** Repo laid out per CLAUDE.md §3 (`docs/`, `harness/`, `tests/`,
`fixtures/`, `measurements/`, `tools/`; no `src/`, per instructions). `architecture.md`
moved to `docs/dlc-blueprint.md`. Git repo initialized. Soufflé 2.5 installed from
release `.deb` (`docs/DECISIONS.md`). Wrote `harness/probe0.py` (seeded fixture
generation for P1/P2/P3, verifies P1's reachable-from-1 set before running anything,
drives Soufflé with and without `--magic-transform=*`, profiles each run),
`harness/parse_profile.py` (exact per-relation tuple counts from Soufflé's JSON
profile log, since `souffleprof`'s text table rounds), `harness/extract_exact.py`
(runs the parser against all six logs with full provenance). Wrote the three `.dl`
programs verbatim from blueprint §12 to `tests/programs/`.

**Measurement IDs produced:** `probe0-p{1,2,3}-{off,on}-{run,profile,extract}` (18
directories), each self-contained under `measurements/<id>/` with
`cmd.txt`/`stdout.txt`/`stderr.txt`/`env.txt`/`meta.json`, plus Soufflé's own raw
output (`.csv`, `.log`) for the `-run` IDs. Full table in `docs/MEASUREMENTS.md`.

**What is now blocked:** Phase 1. Three CLAUDE.md §5 STOP conditions fired: Soufflé's
`--magic-transform=*` transformed relations containing negation on both P2 and P3
(contradicting the blueprint's stated premise and its own worked STOP example);
`T_magic(P1) > T_base(P1)`; P1's magic/no-magic ratio (≈1) misses the blueprint's
stated ~10^3 expectation by more than one order of magnitude. Full escalation with
three live explanations and one proposed (unrun) distinguishing experiment
(`--inline-exclude=q`) is in `docs/reports/probe0.md`.

**Single next action:** Human decides among the three live explanations in
`docs/reports/probe0.md` (Soufflé version drift / Soufflé already guard-equivalent /
P3-specific inlining artifact) and whether to authorize the `--inline-exclude=q`
follow-up experiment before Phase 1 starts.

## 2026-08-20 — Phase 0.5: corrected probes, blueprint amended to v1.1

**What changed:** Applied the human's ruling on `docs/reports/probe0.md` (received as
a "Phase 0.5" directive). Amended `docs/dlc-blueprint.md` to v1.1: differentiator
restated (§2 — Soufflé restricts the negating relation, never the negated one),
guard clause (b) promoted to primary metric source and (a) demoted to a correctness
side-condition (§6), the "positive cycle" precondition flagged unverified (§6),
primary headline metric redefined around negated-relation materialization (§7),
Soufflé Prior Art Register entry corrected against observed 2.5 behavior (§11), P1
replaced by P1' and P3 replaced by P5, P4 added as a decisive hand-transform
experiment (§12), failure mode #6 added (§9). Wrote `tests/programs/p1prime.dl` and
`tests/programs/p4.dl`; ran both against the *existing* P1/P2 fixtures, unregenerated.
Wrote `harness/probe0_5.py` (reuses `probe0.py`'s run helpers). `docs/MEASUREMENTS.md`
gained a `convention` (`excl-copy`/`incl-copy`) treatment per the ruling in §3.

**Measurement IDs produced:** `probe0.5-p1prime-{off,on}-{run,profile,extract}`,
`probe0.5-p1prime-diff`, `probe0.5-p4-{run,profile,extract}`, `probe0.5-p4-diff`,
`probe0.5-p2-bfs-check`. Full table in `docs/MEASUREMENTS.md`.

**Result:** P1' ratio ≈1.51×10⁴ (predicted ~1.5×10⁴) — confirms the `.output path`
mechanism exactly. P4's `q2.csv` byte-identical to P2's in both configurations;
`reach_bf`=170 vs Soufflé's unrestricted `reach`=26,404 (BFS-cross-checked; the "≈50"
prediction assumed P1's engineered fixture, P2's fixture has no such construction —
logged, not acted on). Full writeup: `docs/reports/probe0_5.md`.

**What is now blocked:** Nothing — Phase 0.5 answered its four questions and the
escalation is resolved by the human ruling, not by further investigation.

**Single next action:** Human decides whether to start Phase 1 (M1: lexer, parser,
type/allowedness checks, naive + semi-naive fixpoint, Lane A) or to run P5 first as an
M3 pre-check. `--inline-exclude=q` is now a stated P5 prerequisite
(`docs/OPEN_QUESTIONS.md`), not a standalone experiment to schedule separately.

## 2026-08-20 — Phase 0.6: P4' fix, P6 counterexample hunt, Q5 pre-registration

**What changed:** Applied the Phase 0.6 directive. Blueprint amended to v1.2
(`docs/dlc-blueprint.md`): three-column headline metric (`T_none`/`T_souffle`/
`T_guard`, contribution = `T_souffle/T_guard`, `T_none/T_guard` prohibited as
headline), clause-(a)/(b)-independence note added to §6, Q5 deadline moved to now
(§10), M1/Phase-0.6-parallel note added to §8. Fixed `p4.dl`'s query-constant-in-
rule-head bug as `tests/programs/p4prime.dl`. Ran a bounded (~1hr) counterexample
search for P6 across five constructions (`tests/programs/p6start_*.dl`,
`p6a1_*.dl`, `p6a1b_*.dl`, `p6a2_*.dl`). Pre-registered the M1-M3 corpus: wrote
`harness/corpus_predicate.py` + `harness/build_corpus.py`, sparse-checked-out
Soufflé 2.5's `tests/` tree (not vendored), applied the predicate, committed
`tests/corpus/PREREGISTERED.txt` + `detail.json` + `SOURCE.md`.

**Did not touch `src/` or write any Lane A code.** M1 (lexer, parser, type/
allowedness checks, naive + semi-naive fixpoint) is the human's work, running in
parallel per §8 v1.2 — this session's scope was Phase 0.6 (Lane B) only.

**Measurement IDs produced:** `probe0.6-p4prime-{run,profile,extract,diff}`,
`probe0.6-p6start-{base,hand}-{run,profile}`, `probe0.6-p6a1-{base,hand-naive,
hand}-{run,profile}`, `probe0.6-p6a1b-{base,hand-naive,hand}-{run,profile}`,
`probe0.6-p6a2-{base,hand}-{run,profile}`, `probe0.6-q5-eval-only`,
`probe0.6-q5-corpus`. Full table in `docs/MEASUREMENTS.md`.

**Result:** P4' confirmed sound (byte-identical to P2 and to the original buggy P4);
`@neglabel.reach` confirmed `REC_T` (genuine re-derivation), not `COPY_T`. P2's
three-column table: 40,030 / 26,465 / 231 — contribution is 114.6×
(`T_souffle/T_guard`), not the 173× a `T_none/T_guard` framing would have claimed.
P6: no counterexample found in five bounded attempts (one initially degenerate,
fixed and rerun); working hypothesis is clause (b) collapses into correct seed
collection + clause (a), M3 candidate re-scope to ~2 weeks, pending human
confirmation. Q5: 36 of 612 tests pre-registered, full-tree scope (not the
narrower 11-of-149 first framing, which is disclosed not discarded). Full writeup:
`docs/reports/probe0_6.md`.

**What is now blocked:** Nothing on the Lane B side. M1 continues independently.

**Single next action:** Human confirms (or overturns) the M3 re-scope-to-~2-weeks
recommendation from the P6 result, and decides whether/when to run the 36-program
pre-registered corpus (not yet run — that's an M1/M2/M3-relevant future step, not
Phase 0.6's).

(NIGHT-BATCH-01, 2026-08-20 to -21, is tracked separately in `docs/NIGHT_LOG.md`
and `docs/reports/night01-summary.md`, per the batch's own protocol — not
duplicated here.)

## 2026-08-21 — Corpus ruling: two corpora, T6 correction, subsumption resolved, OpenRuleBench blocked

**What changed:** Applied the corpus ruling (`docs/reports/
corpus-ruling-2026-08-21.md`, saved verbatim as the document of record —
`docs/phase0.7-corpus-viability.md` never landed, this supersedes that reference
everywhere). Blueprint bumped to v1.4: §7 splits correctness corpus (Soufflé
`tests/`, unchanged) from measurement corpus (OpenRuleBench, new); §9 gets
failure mode #7 (fired, resolved by migration) and #8 (negation-bearing
recursive+bound-query programs may be rare, T3/T4's 34-37% zero rate is first
evidence); §10 closes Q3, closes Q5 for correctness/reopens-then-blocks it for
measurement; §12 adds the P2-scale named benchmark with the `T_none=n²`
disclosure attached permanently.

`tests/corpus/PREREGISTERED.txt` header-annotated as superseded for measurement
purposes (data rows unchanged). `corpus_predicate.py`'s non-determinism fixed
(prohibition lifted narrowly per the ruling §4.3) — verified 3 consecutive runs
byte-identical, `PREREGISTERED.txt` unaffected. `docs/reports/
night01-T6-scaling.md` corrected per §3 (`T_none=n²` is definitional, not a
finding; `T_souffle≈0.62n²`, `T_guard≈1.2n`, `T_souffle/T_guard=Θ(n)`).

Subsumption divergence (§4.1) minimized to 4 nodes/1 rule
(`tests/programs/subsumption_minimal2.dl`), found already reported and already
fixed upstream (souffle-lang/souffle#2322, #2323, PR #2567, merged 2025-12-07 —
8+ months after our installed 2.5). No new issue filed. `docs/reports/
subsumption-repro.md`.

OpenRuleBench pre-registration (§2.2) attempted and **blocked**: neither the
original distribution (dead DNS) nor RUBEN (github.com/kev-ang/RUBEN, cloned and
inspected directly — no rule files, private local dataset path, dead referenced
external host) yields the actual rule-program files. `tests/corpus/
MEASUREMENT_PREREG.txt` not created — no fabrication. Catalog metadata found
before the block (RUBEN's `all_tests.json`) independently suggests the floor of
8 may not be met regardless (`negation` category is 1 program shape,
`same_generation`, at 3 scale points). `docs/reports/
openrulebench-preregistration.md`, `docs/ESCALATIONS.md` (2026-08-21).

**What is now blocked:** OpenRuleBench measurement-corpus pre-registration —
needs either a working data source this session's tools can't reach, or a human
decision to route around it (accept the thin coverage as the answer to failure
mode #8, or select a different measurement corpus with the same rejection-reasoned
rigor DOOP got).

**Single next action:** Human decides how to unblock or route around the
OpenRuleBench access problem. M1 (Lane A) still hasn't started — the ruling's own
§6 says the next session report should lead with M1 or the schedule is fiction;
this session did not touch `src/` and did not change that.

## 2026-08-22 — Ruling v1.5: version risk closed, benchmark family pre-registered

**What changed:** Applied the "corpus closed, version risk opened" ruling
(`docs/reports/corpus-ruling-v1.5.md`, saved verbatim as the document of record).
Blueprint → v1.5, marked "last spec until M1 exists."

**T0 (highest priority, cap 1 hour, used well under):** cloned Soufflé master
(`a1303be3`, 42 commits past the `2.5` tag), confirmed via commit-title scan that
only one commit touched the magic-set transform path since 2.5 (the subsumption
fix already known). Installed missing build deps (`bison`, `flex`, `sqlite3`
CLI), built master successfully, re-ran P2 and P4' against it. **Result: master
behaves identically to 2.5** — `@neglabel.reach`=26,404, `T_souffle`=26,465,
`T_guard`=231, all answer relations sorted-set-equal to the committed 2.5 runs.
Differentiator holds; 2.5 pinned; new Q6 opened and closed same-day.
`docs/reports/T0-version-risk.md`.

**§2 corpus hunt closed:** blueprint §9 failure mode #8 promoted to a stated
finding (three independent sources agree: negation-bearing recursive Datalog
with bound queries, at scale, is rare — 37%/34%/OpenRuleBench's 1-shape zero
signal).

**§3.2 canonical benchmark family pre-registered, not run:** five shapes in
`tests/corpus/BENCHMARK_FAMILY/`, each with a citation (Ullman 1988,
Beeri-Ramakrishnan 1987, OpenRuleBench's category as motivation where the
original file was unobtainable) or explicitly disclosed as constructed
(`culprit_cycle.dl`, for clause (a)). Two new generators added to
`harness/fixtures_lib.py` (`gen_family_tree`, `gen_culprit_cycle_facts`),
validated in-memory against every pre-registered scale point
(`SCALE_POINTS.json`) before committing — no fixture files or measurement
numbers produced, per instruction. `harness/run_benchmark_family.py`'s
Soufflé-invocation half deliberately left unimplemented.

**§3.3 DOOP (optional, cap 3 hours, attempted after T0/§3.2 complete):**
cloned `plast-lab/doop` (the actively-maintained, CI-tested canonical repo —
252 real `.dl`/`.logic` files in `souffle-logic/`). Requires Java 17 + Gradle
build + a target `.jar` (+ JRE platform libs from a separate `doop-benchmarks`
repo for a full analysis) — a materially heavier pipeline than anything else
this project has run. **Outcome: abandoned, no profile produced.** `./doop -h`
built and ran successfully (confirms the driver works here); Maven-coordinate
input hung for 2+ hours on Ivy dependency resolution (confirmed genuinely
stuck, not slow — direct Maven Central access worked fine in parallel); local
jar input reached DOOP's analysis factory in seconds but failed needing a JRE
platform library whose fetch URL (from the `doop-benchmarks` companion repo)
returns "not a valid input" for both the auto-selected and an explicitly
requested standard platform. Three distinct, clearly diagnosed failure points,
none inside Soufflé itself. Abandoned per the 3-hour cap, not retried with a
higher cap or further variants. `docs/reports/doop-attempt.md`.

**What is now blocked:** Nothing required. DOOP was optional and remains so —
§3.2's benchmark family is the measurement corpus of record regardless.

**Single next action:** Per the ruling's §5 — no further specs until M1 exists.
Lexer → precedence parser → decl/type check → allowedness → naive fixpoint →
semi-naive, three weeks, Lane A, human's track. This session did not touch
`src/`.

## 2026-08-22 — M1 session: J1 allowedness probe, J2 gates, J3 review protocol

**Note:** session opened referencing `docs/ruling-v1.5-final.md`, which does not
exist (only `docs/reports/corpus-ruling-v1.5.md`, already applied last
session). Not blocking — this session's task list (J1/J2/J3) is fully
self-contained in the instructions themselves, nothing in it depends on the
missing file's content. Noted, not escalated.

`src/` still does not exist as of this session's start; M1 (Lane A) is the
human's track and untouched here per the hard boundary.

**J1 (done first, per instruction):** 7 allowedness probe programs
(`tests/programs/allowedness_probe_{a..g}.dl`) run against Soufflé 2.5.
Accept: a, b, e, g. Reject: c, d, f (`Ungrounded variable X`). No definition
of allowedness proposed — observed behaviour only, per instruction. Full
table and per-case notes: `docs/reports/J1-allowedness-probe.md`.

**J2 (M1 gates, finished from NIGHT-BATCH-01 T7):** `harness/dlc_interface.py`
(new shared stub module: `run_dlc_parse`, `run_dlc_pretty_print`, both
`not_implemented` until Lane A lands). `harness/parse_coverage.py` — 195/195
`tests/corpus/IN_GRAMMAR.txt` files correctly report `not_implemented`, exits
nonzero otherwise (the human's day-2 acceptance gate, exists before the
parser does). `harness/round_trip_scaffold.py` — parse→print→reparse→compare
harness, 195/195 `not_implemented`, printer left as a stub deliberately.
`tests/rejection/allowedness.py`'s 3 original cases cross-validated against
J1 (all match exactly); 1 new case added
(`allowedness_equation_rhs_not_bound`, J1's least obvious finding —
already-grounded var on the left of `=` doesn't ground an unbound var on the
right) and independently re-verified against Soufflé 2.5 directly. All 13
rejection cases + parse-coverage + round-trip + the T7 golden-guard tests
re-run and confirmed still correctly blocked — no vacuous green anywhere.
Full detail: `docs/reports/J2-m1-gates.md`.

**J3 (review protocol):** noted, not executable yet — it activates "when the
human pushes Lane A code," and `src/` still does not exist as of this
session. No Lane A pushes to review. Protocol as instructed: on any future
push, run the differential harness + parse-coverage runner and report
numbers; attack with adversarial inputs (precedence edge cases, malformed
input, unicode, deep nesting, empty bodies, comments in awkward positions),
write failing tests under `tests/`, commit them; never edit `src/`; if asked
to fix a Lane A file, diagnosis + failing test, then stop. Speculative
adversarial test files were not pre-written this session — a failing test
needs a real interface to fail against, and inventing one now risks locking
in assumptions about dlc's CLI/AST shape that Lane A hasn't decided yet.

## 2026-08-26 — M1 build start, per docs/M1-BUILD.md

CLAUDE.md §2 superseded by docs/M1-BUILD.md §1 (direct human instruction).
Lane A narrowed to three components: magic-set transform (src/transform/
magicset/), the guard (src/transform/guard/), fallback evaluation wiring
(src/eval/fallback.go). Everything else in src/ is now Lane B.

**Setup (not a numbered §3 item, prerequisite to it):** installed Go 1.26
(none was present), ran `go mod init dlc`, created the full src/ directory
layout per M1-BUILD.md §2, wrote the three Lane A marker files (doc.go x2
+ fallback.go), each containing only the package clause and the one-line
"// Lane A — human-authored. See docs/M1-BUILD.md §1." comment — no stubs,
no algorithm sketches. `go build ./...` and `go vet ./...` both clean.
Logged: docs/DECISIONS.md.

**§3.1 (token and lexer):** gate PASSED. `harness/lex_coverage.py` (new --
see report for why not a reuse of `parse_coverage.py`) ran the built `dlc
lex` over all 195 `IN_GRAMMAR.txt` files plus all 39 `tests/hostile/`
files: 0/234 panics. `token`/`lexer`/`cmd/dlc` packages written, each with
`DESIGN.md`; Go unit tests in `src/lexer/lexer_test.go` (9 tests, all
pass) cover the `.`-vs-directive disambiguation, multi-char operators,
underscore-vs-identifier, error-token recovery, and the two documented
disagreements with Soufflé's lexer (unterminated block comment, non-ASCII
bytes). Found and flagged (not fixed, not blocking): 42/195 "in-grammar"
files produce lex-error tokens -- `IN_GRAMMAR.txt` (NIGHT-BATCH-01 T5's
mechanical scan) missed aggregates, records, `#include`, and pragma
directives as exclusion factors, all explicitly out of blueprint §4's
grammar. This does not fail §3.1's gate (panics only) but means §3.3's
gate one ("all 195 parse with zero errors") is very likely unachievable
as literally stated without a corpus correction -- flagged now, ahead of
that gate, not discovered cold when it fails.

**§3.2 (AST):** no independently-stated numeric gate (M1-BUILD.md gives
none for this item); completion criterion taken as "compiles, has
DESIGN.md, and provides what §3.3-3.9 need" per the item's own framing
("the AST shape is what four downstream passes are written against").
`Program`/`Decl`/`Clause`/`Atom`/`Literal`(`Atom`/`NegatedAtom`/
`Constraint`)/`Term`/`Arith`(`BinaryExpr`/`UnaryExpr`/`Var`/`NumberLit`/
`StringLit`)/`Wildcard`, every node with a `Span()`. Also added
`ast.Equal`, a span-ignoring structural comparison §3.3 gate two's
round-trip check will need (a straight `reflect.DeepEqual` can never pass
that gate, since a reparse always produces different spans). 4 Go unit
tests, all pass. `go build`/`go vet`/`go test` all clean.

**§3.3 (parser):** all three gates run; none is a clean "pass," all three
are honestly reported per M1-BUILD.md §7 ("report it as parsed/195," not
"make it say 195/195"). Recursive-descent parser + precedence-climbing
arith (`src/parser/parser.go`), pretty-printer (`printer.go`), and a
Go-native `Roundtrip` (`roundtrip.go`) using `ast.Equal`. CLI gained
`parse` and `roundtrip` subcommands. 9 Go unit tests (precedence table
matches `docs/reports/night02-T2-hostile.md`'s oracle-verified values
exactly), all pass.

- **Gate one:** `parsed/195 = 20/195`
  (`measurements/m1-3.3-gate1-parse-coverage-summary.json`). NOT a parser
  bug: 105/175 failures (60%) are real Soufflé files using `.input Name()`
  / `.output Name()` with cosmetic trailing parens -- grammar-legal in
  full Soufflé, not admitted by blueprint §4 plus the one authorized
  amendment (term-list-in-atom only). Flagged prominently
  (`docs/OPEN_QUESTIONS.md` 2026-08-26) as a human decision point (worth
  authorizing as a second amendment, given how pervasive and cosmetic it
  is) rather than silently added. Remaining 70 failures independently
  confirmed genuinely out-of-grammar: 15 `unsigned`/`float` types, ~55
  aggregates/functors/pragmas (consistent with §3.1's lex-coverage finding
  and blueprint's own "no functors, no aggregates").
- **Gate two:** `match/195 = 20/195`
  (`measurements/m1-3.3-gate2-roundtrip-summary.json`) -- but the other
  175 are inherited `parse_error` from gate one, not new round-trip
  failures. **Of the 20 files that parse, 20/20 round-trip correctly**
  (100%) -- no printer/precedence bug found on real-world content, beyond
  the 9 synthetic unit tests.
- **Gate three:** 39 hostile files vs `docs/reports/night02-T2-hostile.md`'s
  Soufflé-established verdicts: 35/39 agree; 2 are expected gaps (Soufflé's
  rejection was semantic -- duplicate decl, underscore-in-head -- and
  sema doesn't exist until §3.4/§3.6); 1 is T2's own already-known
  inconclusive case (`lexical_4kb_identifier.dl`); **exactly 1 genuine
  disagreement**, `comment_unterminated_block.dl` -- dlc rejects an
  unterminated block comment, Soufflé silently swallows it to EOF. This
  is the deliberate, pre-documented disagreement from `lexer/DESIGN.md`
  (§3.1) surfacing again here, not a new finding. Reported, not adjusted
  to match, per the gate's own instruction.
  (`measurements/m1-3.3-gate3-hostile-summary.json`)

**§3.4 (sema: decl/arity/type):** no numeric gate stated in M1-BUILD.md
for this item; adopted completion criterion (analogous to §3.2): the 6
real `tests/rejection/{arity,type}.py` cases, run through the real `dlc
check` CLI end-to-end, correctly classified. `src/sema/decltype.go`:
symbol table (one schema-defining `.decl` per relation; `.input`/
`.output` only mark an existing schema, confirmed NOT the same as
duplicate declaration -- `TestInputAfterDeclIsNotDuplicate`), per-clause
(not per-program) variable type tracking, and the arithmetic-forces-
number-but-bare-comparison-doesn't distinction pinned by the
`type_symbol_in_arithmetic` vs a plain `X = "foo"` case
(`TestBareComparisonDoesNotForceNumber`). 13 Go unit tests, all pass.
CLI gained a `check` subcommand.

- **Gate (adopted): 6/6** arity+type rejection-corpus cases correctly
  classified (`measurements/m1-3.4-decltype-summary.json`).
- **Sanity check** (not a strict gate, a validation): ran sema over all
  20 files gate one (§3.3) found parse cleanly. 12/20 clean, 8/20
  rejected. Investigated all 8, none is a checker bug: 4 are multi-file
  Soufflé programs split across `#include`d fragments or an EDB-only
  file with no `.decl` of its own (`example/magic_pointsto/edb.dl`,
  `example/pointsto/edb.dl`, `syntactic/include_directive1/foo.dl`) or a
  reference to a Soufflé builtin (`match`, `example/not_match/
  not_match.dl`) -- genuine single-file-analysis limitations, disclosed,
  not fixed (out of scope). The other ~4
  (`semantic/error_deduce_type.dl`, `semantic/rule_undeclared_relation{,2}
  .dl`, `semantic/type_system7.dl`) are, by their own names, Soufflé's
  *own* designed-to-fail negative test cases -- dlc correctly rejecting
  them is validation, not a bug.

**§3.5 (allowedness):** gate PASSED, cleanly. `src/sema/allowedness.go`
implements the fixpoint definition in `docs/DECISIONS.md` literally --
G0 from positive-atom arguments, `V = E`/`E = V` grounding to a fixpoint,
every variable in the clause must end up grounded. All four asymmetries
(only `=` contributes; grounded side must be bare; no arithmetic
inversion; every variable not just head variables) implemented as
separate, visible code paths, each pinned by DESIGN.md to the specific
probe case that forces it.

- **Gate: 15/15** probe programs (`tests/programs/
  allowedness_probe_{a..o}.dl`) match their recorded Soufflé verdict,
  both as Go unit tests (`src/sema/allowedness_test.go`) and end-to-end
  through the built CLI (`harness/m1_3_5_allowedness.py`,
  `measurements/m1-3.5-allowedness-summary.json`).
- **Plus all 13 `tests/rejection/` cases: 10/13** correctly rejected.
  The remaining 3 are exactly the stratification-ground cases --
  attributed explicitly to §3.6 not existing yet, not a §3.5 shortfall
  (verified: all 10 non-stratification cases, across arity/type/
  allowedness, are 100% correct).

**§3.6 (source stratification):** gate PASSED on both parts.
`src/sema/stratify.go`: precedence graph over IDB relations only
(EDB/`.input`-only relations are never nodes), Tarjan SCC, reject iff a
negative edge closes within one SCC, else memoized-DFS stratum
assignment over the condensation. Explicitly did NOT generalize toward
the transformed-program culprit-cycle detector (Lane A) -- confirmed the
*source* `culprit_cycle.dl` shape stratifies cleanly
(`TestStratifiableCulpritCycleShape`), consistent with T7's finding this
project already had that the culprit cycle is transform-introduced, not
present in the source program.

- **Gate, part 1 (rejects unstratifiable programs): 3/3** — all
  `tests/rejection/stratification.py` cases reject, as Go unit tests.
- **Gate, part 2 (agrees with Soufflé's evaluation order): 1/1** usable
  sample. Only 1 of the 195 in-grammar files both parses under dlc today
  (§3.3: 20/195) and contains real negation — the other two candidates
  are already-known problem files (a Soufflé builtin reference; a
  designed-to-fail negative test). Oracle signal took a wrong turn before
  landing: `--show=initial-ram`'s `SUBROUTINE` list is alphabetical, not
  evaluation order; the real signal is the `BEGIN MAIN ... CALL
  stratum_X ... END MAIN` block, which is *also* not grouped by numeric
  stratum (a valid topological order, not stratum-sorted) — the actual
  comparable invariant is "every negated dependency's `CALL` precedes its
  dependent's, and `dlc`'s own stratum numbers agree with that same
  ordering," not exact-sequence or exact-stratum-number agreement.
  (`measurements/m1-3.6-stratification-summary.json`,
  `harness/m1_3_6_stratification.py`)
- Found and fixed a wrong assumption in a first draft of the determinism
  test: a relation referenced both positively and negatively by two
  *different*, non-mutually-recursive relations is perfectly
  stratifiable (not a bug) -- logged in DESIGN.md as a genuine, easy
  misconception, not just fixed silently.

19 Go unit tests across `stratify_test.go`; full `src/sema` package
(decl/type + allowedness + stratification): 45 tests, all pass.
`go build`/`go vet`/`go test ./...` all clean.

**End of §3 (M1 work items) for this session's continuous run: 3.1-3.6
complete, all gates reported (none silently weakened). 3.7-3.9 remain
(relation storage/indices, naive eval, semi-naive eval + M1's headline
number) -- substantial remaining scope, continuing per the original
instruction ("Execute M1 §3 in order... If §3 completes, continue with
§4").

**§3.7 (relation storage and indices):** no numeric gate stated;
completion criterion (as for §3.2/§3.4): compiles, has DESIGN.md, meets
the item's own stated requirements, exercised by unit tests since real
data doesn't exist until §3.8. `src/ir/relation.go`: `Value`
(int64-or-interned-symbol), `StringTable`, `Relation` with exactly one
index (column 0, naive by explicit permission) and dedup-on-insert (set
semantics). `src/ir/profile.go`: `EmitProfile` matches Soufflé's own `-p`
JSON shape field-for-field so `harness/parse_profile.py`/
`tuple_report.py` need zero changes to read `dlc`'s own output.

- **Instrumentation requirement, verified structurally, not just by
  convention:** `Insert` has no code path into `Stats` at all --
  `TestInsertNeverTouchesStats` confirms EDB loads (which only ever call
  `Insert`) cannot inflate a derived-tuple count by construction, not by
  every call site remembering a flag correctly.
- **"Both copy conventions" (§3.7) deferred, disclosed:** `dlc`'s
  evaluator has no code path yet that could produce a `COPY_T`-shaped
  relation, so excl/incl trivially coincide today -- noted in DESIGN.md,
  not fabricated ahead of there being anything to report.
- 8 Go unit tests, all pass. `go build`/`go vet`/`go test ./...` all
  clean.

**§3.8 (naive evaluation):** gate PASSED, zero disagreements.
`src/eval/naive.go`: per-stratum naive fixpoint, nested-loop join with
`ir`'s column-0 index used when the joined atom's first term is already
bound. `safeOrder` reorders each clause's body so negated atoms/
grounding constraints only get evaluated once their variables are bound
-- required because allowedness's own fixpoint (§3.5) is deliberately
order-independent but a left-to-right evaluator is not; pinned by
`TestEquationBeforeGroundingAtomSafeOrder` against probe case (b)'s exact
shape. 13 Go unit tests (transitive closure, stratified negation,
zero-arity, symbols, dedup, unary minus, instrumentation correctness --
one test's own wrong assumption caught and fixed, logged in DESIGN.md).
CLI gained a `run` subcommand (parse+check+load facts+evaluate+write
`.output` CSVs in Soufflé's own tab-separated shape+write a
Soufflé-profile-shaped `profile.json`, `src/ir.EmitProfile`).
`harness/differential.py`'s `run_dlc()` now calls the real binary --
everything downstream (`compare()`, reporting) needed zero changes,
exactly as that file's original stub-era comment said it would.

- **Gate: agreed/attempted = 11/20** on the in-grammar files dlc's front
  end accepts (§3.3 gate one's 20/195), against real Soufflé, empty
  facts dir (none of the 20 has an `.facts` file in Soufflé's own tree --
  every one defines its EDB via source fact clauses). **0
  disagreements** among the 11 comparable cases. The other 9 are
  correctly rejected by sema before evaluation is even attempted (8
  already known from §3.4/§3.5's findings -- multi-file limitations,
  Soufflé builtins, designed-to-fail negative tests; 1 new one,
  `semantic/var_single/var_single.dl`, itself another Soufflé designed
  negative test for exactly the ungrounded-head-variable shape
  allowedness already covers). (`measurements/
  m1-3.8-naive-eval-summary.json`)

**§3.9 (semi-naive evaluation) — M1's last work item.**
`src/eval/seminaive.go`: Δ-rewrite per SCC (not per stratum -- see the
bug below), one variant per same-SCC positive-atom *occurrence* (keyed
by body position, not relation name, so a self-join gets one variant per
occurrence rather than redirecting every occurrence of that relation at
once).

**Found and fixed a real evaluation-order bug, via the differential
harness, not by inspection:** the first version grouped by stratum
*number* (mirroring `RunNaive`), and `example/josephus/josephus.dl`
immediately disagreed with Soufflé -- `Josephus` (reads a self-recursive
`Relation` positively, no negation anywhere in the file so both land at
stratum 0) got evaluated during the seed round, before `Relation`'s own
recursion had produced anything beyond its 6 initial facts. Fixed by
adding `SCCOrder` to `sema.StratumResult` (`sema/stratify.go`, a plain
topological order of the SCC condensation) and processing one SCC at a
time in that order instead of one stratum-number batch at a time --
stratum number is only ever needed for the negation-safety *rejection*
check, never for driving evaluation order. `RunNaive` never showed this
bug (brute-force full-fixpoint repetition is insensitive to
over-coarse grouping; semi-naive's entire point is to not redundantly
repeat, which is exactly what the bug violated). Two regression tests
added (`sema/stratify_test.go`, `eval/seminaive_test.go`), both passing;
full account in `eval/DESIGN.md`.

- **Gate one (same set equality as 3.8, unchanged): 11/20**, re-matching
  §3.8's baseline exactly after the fix, 0 disagreements
  (`measurements/m1-3.9-gate1-seminaive-agreement-summary.json`,
  `harness/m1_3_9_gate1_seminaive_agreement.py`).
- **Gate two, M1's headline number: `T_naive` vs `T_semi-naive`, exact
  tuple counts, on a fixed program set** (the 5 `tests/corpus/
  BENCHMARK_FAMILY/` shapes, pre-registered, each at its own smallest
  scale point -- `harness/m1_3_9_gate2_headline.py`,
  `measurements/m1-3.9-gate2-headline-summary.json`). **`T_naive ==
  T_semi_naive` exactly, all 5 shapes, ratio 1.00x every time** -- this
  is mathematically expected, not a null result: both evaluators compute
  the identical minimal Herbrand model of the same program, and both
  count each distinct tuple once (this project's own T-metric convention
  since Phase 0), so the final counts cannot differ by construction. The
  actual optimization signal semi-naive provides is in avoided redundant
  *re-derivation*, not fewer final tuples -- added a second counter,
  `Evaluator.DerivationAttempts` (every candidate head tuple a clause
  match builds, counted before dedup, so naive's repeated full rescans
  show up and semi-naive's Δ-restriction doesn't), reported alongside per
  shape, never aggregated: `same_generation_negation` 2.00x,
  `transitive_closure_bound` 18.93x, `ancestor_nonancestor` 6.18x,
  `reachability_complement` 16.21x, `culprit_cycle` 6.05x. Answer-set
  equality between naive and semi-naive verified on all 5 (not just the
  count) before trusting any of this.

**M1 §3 complete: all 9 work items (3.1-3.9) done, every gate reported
as measured, never weakened to hit a target number.** `go build`/
`go vet`/`go test ./...` all clean throughout. Continuing per the
original instruction: "If §3 completes, continue with §4 in order."

**§4 item 2 (the Lane A transform interface):** `src/transform/
transformer.go` -- `Transformer` interface (`Transform(prog
*ast.Program, strata *sema.StratumResult) (*ast.Program, error)`) plus
`PassThrough`, a no-op Lane B placeholder. Deliberately NOT wired into
`cmd/dlc`'s `run`/`run-seminaive` yet -- routing through a no-op changes
nothing observable; the natural point to wire it in is when Lane A's real
implementation lands (one call between `CheckStratification` and
`RunNaive`/`RunSemiNaive`, nothing else in `cmd/dlc` changes). Documented,
not solved: a real transform invalidates the `strata` argument for its
own output (magic-seed relations change the precedence graph) --
`Transformer`'s own doc comment says the implementer re-runs
`sema.CheckStratification` on the result; deliberately not designed
further than that (would be resolving a Lane A design question). 2 Go
tests, both pass.

**§4 item 1 (C code generation):** `src/codegen/` -- emits a standalone C
program (naive evaluation only, not semi-naive's Δ-rewrite; §4 item 1's
own wording describes naive's shape, disclosed scoping in DESIGN.md) from
an already-checked program: fixed-arity tuple structs, a chained hash
index on column 0 per relation ("nested loops... with hash indices," the
task's own wording, same one-index choice `ir.Relation` makes and for
the same reason), fact loading matching Soufflé's tab-separated
convention, per-stratum naive fixpoint. CLI gained a `codegen`
subcommand. **8 real end-to-end tests** (generate C -> compile with `cc`
-> run the binary -> check output): facts+rules, transitive-closure
recursion, stratified negation, `.input`-loaded facts, arithmetic +
constraints, symbols/string literals, a genuine self-join, zero-arity --
all pass. Spot-checked independently outside the test suite too:
`dlc codegen` on `ancestor_nonancestor.dl` at n=500, compiled and run,
matches NIGHT-BATCH-02 T4's already-committed real-Soufflé baseline
exactly.

- **Found and fixed a real correctness risk before any generated C was
  ever compiled:** a first draft's column-0 lookup collected matching
  row indices into a fixed `1<<20`-entry static buffer; NIGHT-BATCH-02's
  own measurements found relations in the millions of tuples at this
  family's larger scale points, so a single hash bucket could plausibly
  overflow it. Replaced with an inline hash-bucket walk at each join
  site -- no fixed-size buffer anywhere, regardless of relation size.
- **Disclosed, not fixed:** ordering comparisons (`<`,`<=`,`>`,`>=`) on
  `symbol` columns compare interned ids (assignment order), not strings
  (lexicographic) -- `=`/`!=` are unaffected (id equality is string
  equality); fixing the ordering case needs per-variable type
  information sema computes internally but doesn't currently export.
  Neither the benchmark family nor the hostile corpus exercises this
  shape, so it hasn't been hit by anything tested, but it is a real,
  disclosed gap, not silently assumed correct.

**§4 item 3 (extend the differential harness to the full benchmark
family, every pre-registered scale point):** `harness/
m1_4_3_full_family_differential.py`, `dlc run` (naive) vs real Soufflé,
ascending per shape, 120s cap per point (this task's own choice, never
retried higher on a DNF). **16/16 comparable points matched, 0
disagreements**, across all 5 pre-registered shapes. 4 DNFs, one per
shape except `culprit_cycle` (completed all 5 of its own points) --
every DNF lands exactly where NIGHT-BATCH-02 T4's already-committed
measurements predicted it would (tens-of-millions-of-tuples territory,
no query planning in a tree-walking interpreter). Full table:
`measurements/m1-4.3-full-family-differential-summary.json`.

**M1 (all of §3 and §4) complete.** Final report per §7:
`docs/reports/m1-progress.md`.

## 2026-08-27 — NIGHT-BATCH-03 + M2-M3-BUILD, unattended overnight batch

Lane A retired per `docs/m2 m3.md` §0 (supersedes M1-BUILD.md §1 and
NIGHT-BATCH-03 §0.2.1) -- all of `src/` including `transform/magicset/`,
`transform/guard/`, `eval/fallback.go` is Lane B from this session on.

**T1: DONE.** Printer -> Soufflé round-trip. `harness/night03_t1_printer_souffle.py`.
Gate: 6/6 shapes (5 BENCHMARK_FAMILY + p4prime.dl) Soufflé-accepted and
answer-identical against the untransformed original. Full 26-case corpus:
17/26 accepted, 17/17 answer-identical among comparable, 0 printer-only
failures, 0 answer mismatches -- the 9 non-accepted cases fail identically
on the original file too (broken/include-only Soufflé fixtures, not a
printer defect). `docs/reports/night03-T1-printer-souffle.md`. Commit
`df652e1`. T2 unblocked.

**T2: DONE.** `dlc emit` subcommand + `transformerRegistry` added to
`cmd/dlc`. `harness/night03_t2_protocol.py`: full M3 measurement pipeline
(`dlc` emits, Soufflé evaluates) on `--transformer=passthrough`, all 5
`BENCHMARK_FAMILY` shapes, every `SCALE_POINTS.json` point (25/25). Tuple
totals + answers identical in all 25 (mathematically required for a no-op
transformer; confirms the plumbing is transparent).
`docs/reports/night03-T2-protocol.md`. Commit `e0d550d`.

**T3: DONE.** `harness/m2_accept.py` (semantic acceptance harness for any
candidate transform) validated on 3 known-good pairs
(p2/p4prime, ancestor_nonancestor/guarded, same_generation_negation/guarded)
-- exact match against committed T5/T6 numbers on all 3, 3/3
answer-identical. T3a: `souffle --show=transformed-ast` (confirmed name --
NOT `--show=transformed-datalog`, which does not exist) dumped for 8 files
under `tests/reference/souffle-transformed/`, reference-only.
`docs/reports/night03-T3-m2-harness.md`. Commit `a394848`.

**T4: DONE.** `harness/night03_t4_culprit_classifier.py` implements
blueprint section 6's formal culprit-cycle mechanism directly (found and
fixed a real directive/period-splitting bug before trusting the count --
`culprit_cycle.dl` itself scored zero findings until fixed). Census: 0/817
real-world files (195 in-grammar + 622 full Souffle tree) match beyond the
one already-known file. 6 new hand-constructed programs added to
`tests/corpus/CULPRIT_CANDIDATES/` (5 structurally matching + 1 deliberate
negative control, `cc_edb_negated`, EDB negated predicate correctly not
flagged), all classifier-verified, `dlc check`-clean, Souffle-recorded
(6/6 complete, 6/6 answer-identical, 0 stratification failures under
Souffle's own transform -- consistent with `culprit_cycle.dl`'s own
observed behaviour). `docs/reports/night03-T4-culprit-corpus.md`.

**T5: DONE.** Q11 pre-registered (`docs/OPEN_QUESTIONS.md`, commit
`caa4c0f`) before any measurement, per instruction. Built exactly as
predicted (`ancestor_nonancestor_guarded_v2.dl`, static seed only, no
propagation) -- **falsified at all 5 scale points**: the premise
misidentified `ancestor`'s actual recursive shape (not first-argument-
invariant, unlike `reach`/`p4prime.dl`), verified by inspection before
building v2. `docs/reports/night03-T5-ancestor-seed.md`.

**T6: DONE, amendment kept.** `.input`/`.output` optional-parens amendment
(`src/parser/parser.go`). Gate one: 20/195 -> 89/195; negation-bearing
parsed: 3 -> 18. Both legs of the decision rule pass. 0 round-trip
regressions. `docs/reports/night03-T6-parens.md`.

**T7: DONE.** `harness/night03_t7_grammar_v2.py` (23 categories, each
justified against a specific blueprint section 4 production). Converged
iteratively (43/622 admitted with 44% cross-check -> 21/622 admitted, 90%
cross-check) after investigating real discrepancies against real `dlc`
diagnostics. `IN_GRAMMAR_V2.txt` committed alongside the untouched
`IN_GRAMMAR.txt` (confirmed exact subset). `docs/reports/
night03-T7-grammar-v2.md`.

**T8: DONE, real defect found and fixed.** Codegen's `<`/`<=`/`>`/`>=` on
`symbol` columns compared interned ids, not strings -- confirmed real
(temporarily disabled the fix, all 4 new tests failed exactly as
predicted, then restored it). Fixed by exporting `sema.ClauseVarTypes`.
Cross-checked against real Souffle: set-identical. `docs/reports/
night03-T8-symbol-order.md`.

**T9: DONE.** `harness/cone_metric.py`: downward dependency closure over
the full (positive+negative) edge set. Validated exactly against
`culprit_cycle`'s already-observed `{q,s}` pattern. All 5 shapes reported:
3 partition-style shapes at `cone_fraction=1/3`, `culprit_cycle` at `1/2`,
`transitive_closure_bound` n/a (no negation). `docs/reports/
night03-T9-cone.md`.

**T10: DONE.** 17 rows appended to `docs/MEASUREMENTS.md` (one per M1
gate), 3 `docs/DECISIONS.md` entries (flat-JSON provenance format
accepted; Q7 closed, C codegen drops below M3 overflow in M4; T8
zero-arity correction 11->12 recorded).

**T11: DONE.** `docs/dlc-blueprint.md`'s header `Status:` line (the only
volatile milestone-status assertion found anywhere outside session
history/logs) replaced with a pointer to `docs/SESSION_LOG.md` and
`docs/reports/m1-progress.md`. Version number untouched, no new document.

**NIGHT-BATCH-03 (T1-T11): all 11 tasks DONE, no BLOCKED, no PARTIAL.**
Full report: `docs/reports/night03-summary.md`. Continuing per instruction
into `docs/m2 m3.md` sections 2-9 (Lane A retired, all of `src/` now Lane
B).

## 2026-08-27 (continued) — M2-M3-BUILD sections 2-9, Lane A fully retired

Same session as the NIGHT-BATCH-03 entries above. `docs/m2 m3.md` section
0 retires Lane A entirely -- `src/transform/magicset/`, `src/transform/
guard/`, `src/eval/fallback.go` all implemented this session, marker
files deleted.

**M2 (sections 2-4): DONE.** `src/transform/magicset/`: `adorn.go`
(worklist adornment, occurrence-level tracking, EDB atoms never pushed,
10,000-pair cap), `sips.go` (source-order + early-bound-constraint +
negated-literal-after-grounding, no cost model per instruction),
`rules.go` (magic rules + supplementary chain; the V_i projection's
off-by-one, re-derived correctly from first principles before any
generated code existed -- full derivation in DESIGN.md), `transformer.go`
(`--transformer=magicset`). Validated via `harness/m2_accept.py` on 5
`BENCHMARK_FAMILY` shapes + `p2.dl`: 5/5 comparable cases answer-identical.
Found and fixed a real relation-origin mistagging bug while adding
`RelationOrigin` export (an occurrence's magic relation was tagged with
the enclosing RULE's predicate instead of the occurrence's TARGET
predicate) -- caught by a dedicated test before it could corrupt M3's
cone computation. `ancestor_nonancestor`/`reachability_complement`/`p2.dl`
mechanically adorn `bb` (both grounding atoms precede the negation), a
real measured cost gap vs. the hand guards' `bf` (up to ~194x), disclosed
not hidden. `culprit_cycle.dl`'s mechanical transform is unstratifiable --
the exact designed-for signal for the guard. 12 Go unit tests. Report:
`docs/reports/m2-headline.md`. Commit `b096057`.

**M3.1 (section 5): DONE.** `src/transform/guard/seeding.go`:
`AssertNegationAllBound`, the required gate on the central soundness claim
(3-sentence argument in DESIGN.md). Counterexample search (`harness/
night_m3_1_counterexample.py`): 4 `p6*_base.dl` + 6 `CULPRIT_CANDIDATES`
programs, 0 disagreements (4 no-op pass-throughs, 1 agreed, 5 correctly
unstratifiable matching NIGHT-BATCH-03 T4's structural predictions
exactly). Report: `docs/reports/m3-1-seeding.md`. Commit `4fee827`.

**M3.2 (section 6): DONE.** `src/transform/guard/stratify.go`:
`HasPositiveCycle` (cheap O(V+E) precondition), `CheckCulpritCycle`
(builds the candidate transform, runs `sema.CheckStratification` on it
directly), `AllUnstratifiableSCCs` (every culprit SCC, not just the
first). Differential oracle vs real Soufflé: 7/7 agreed (6
`CULPRIT_CANDIDATES` + `culprit_cycle.dl`), 0 disagreements. Report:
`docs/reports/m3-2-culprit-detection.md`. Commit `a76921a`.

**M3.3 (section 7): DONE.** `src/transform/guard/decide.go`: `Decide`
(culprit SCC -> original predicates via `RelationOrigin`, `ConeClosure` ->
downward closure over the source graph). `ConeClosure` cross-checked
exactly against NIGHT-BATCH-03 T9's already-committed `harness/
cone_metric.py` result on `culprit_cycle.dl` `{p}` declined: both compute
`{q,s}` exactly. `magicset.GenerateMixed` (declined-aware variant of
`Generate`) builds the final mixed program. `--transformer=guarded`
registered. Gate: 7/7 answer-identical; on the 6 genuine culprit-cycle
programs `T_guarded==T_none` exactly (whole unstratifiable SCC already
equals every IDB relation in this family); 0 regression confirmed
byte-identical on the 4 non-culprit `BENCHMARK_FAMILY` shapes. Report:
`docs/reports/m3-3-decide.md`. Commit `94d4e4b` (prep: `745931b`).

**M3.4 (section 8): DONE.** Verified first, per instruction: a mixed
program is an ordinary `*ast.Program`, the existing SCC-ordered evaluator
needs no changes. No corpus program produced a genuine (non-trivial) mix
before this task; added `tests/corpus/CULPRIT_CANDIDATES/
cc_mixed_fallback.dl` specifically to exercise one. Wired
`--transformer=<name>` into `dlc run`/`run-seminaive` (re-derives
stratification on the transformed output, never reuses the pre-transform
one). `src/eval/fallback.go` documents the verification, contains no new
evaluation logic. Three independent confirmations agree (dlc's own
evaluator via Go test and CLI, Soufflé on the untransformed original,
Soufflé on the guarded-emitted program). Report: `docs/reports/
m3-4-fallback.md`. Commit `213061a`.

**M3.5 (section 9), the headline run: DONE.** `harness/
night_m3_5_headline.py`: full T2-protocol sweep, all 5 `BENCHMARK_FAMILY`
shapes at every `SCALE_POINTS.json` point (25) + 7 `CULPRIT_CANDIDATES`
programs. **32/32 points, 0 DNFs up to n=8,000, 32/32 answer-identical**
-- the first time this project's own evaluator is not the measurement
bottleneck. Guard fires only on culprit-cycle-shaped programs (6/12),
always clause (a), cone always empty. Blast-radius integer: **22 declined
relations** across the corpus, 0%-80% per firing program. Hand transform
beats `dlc`'s mechanical one on every shape/scale point measured (up to
~5,300x on `reachability_complement`) -- reported plainly. One real,
unexplained finding: `transitive_closure_bound` is consistently ~0.49x
worse than Soufflé's own transform at all 5 scale points, not
investigated to root cause. Every `T_none`/`T_souffle` value cross-
confirmed exactly against already-committed `night02-T5-guarded.md`
numbers. Report: `docs/reports/m3-headline.md`. Commit `853ddbc`.

**Stopped at §12's item 8 ("the whole thesis") deliberately.** Items 9-11
(presentation artifact, all-free duplication strategy, C codegen for
transformed programs) are explicitly marked droppable/optional in
M2-M3-BUILD.md's own order-of-work table; §10 (presentation artifact) is
explicitly gated on a blast-radius integer existing, which it now does,
but is scoped as a separate "two days capped" follow-on, not a
same-session continuation.

**What is now unblocked:** the presentation artifact (§10), if the human
elects to spend the two-day budget on it; the all-free-duplication cone
strategy as an alternative to measure alongside cone-untransform (§7,
droppable); the `transitive_closure_bound` regression is worth a root-
cause investigation before any further headline claims cite it.

**Single next action:** human decides whether to continue into §10-11 or
treat M3 as complete and move to writeup.
