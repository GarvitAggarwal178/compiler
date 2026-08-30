# Project log

A narrative of how this project actually went, session by session — what
was attempted, what was learned, what was wrong and how it was caught.
Assembled from `record/SESSION_LOG.md`, `experiments/04-night01-session-
log.md` (the original `NIGHT_LOG.md`), and the three morning summaries
(`experiments/12`, `30`, `42`). Those source files are left exactly as
written in `record/` and `experiments/`; this is the readable version, not
a replacement for them.

## Probe 0 — the premise breaks on day one

The project's founding premise was "production engines decline to
transform anything touching negation." Probe 0 tested it directly against
Soufflé 2.5, and it was wrong: Soufflé transforms the relation whose body
*contains* a negated literal, it just never demand-restricts the relation
*inside* the negation, isolating it under `@neglabel.<rel>` and computing
it in full. The premise was patched the same day (blueprint v1.1) rather
than the project being re-scoped — the corrected differentiator
(restricting the negated side, safely, is the actual gap) turned out to be
a stronger result than the original guess. Two of Probe 0's own three
programs also turned out to be defective by construction (P1 forced full
materialization via an unnecessary `.output`; P3's pivot relation was
inlined away by Soufflé before any transform question was reachable) —
both replaced, both documented as replacements rather than quietly fixed
in place. See `docs/design-history.md` for the full version table.

## NIGHT-BATCH-01 — corpus discipline, and a first false alarm

The first unattended batch built the provenance/determinism audit
(8 gaps found and backfilled, 61/61 measurement directories reproducing
byte-identically), a first grammar census, and a first scaling sweep. It
also produced the ruling that killed a bad number before it shipped: a
31-program corpus mixed differential-testing shapes with tiny fixed-size
unit tests, and a recoverable-fraction ratio computed over the mix was
measuring Soufflé's own unit tests being unit tests. Rather than patch the
ratio, the corpus itself was split in two — a correctness corpus (scale
irrelevant, real Soufflé test files) and a measurement corpus (scalable,
known-shape benchmarks) — never reported together again.

## NIGHT-BATCH-02 — a claim that had to be retracted mid-batch

T5 reported that Soufflé could exit with status 0 on a genuine
stratification error. It couldn't. T9, a few tasks later in the same
batch, re-ran the same program five times with output redirected to files
instead of streamed live to the terminal, and got a consistent exit code
of 1 every time. The original `rc=0` reading was a Bash-tool/`wsl.exe`
bridge artifact — a race on live-streamed multi-command terminal output,
not real Soufflé behaviour. It was corrected the same batch, marked rather
than silently edited in both the report and the two harness files whose
comments had repeated the wrong claim, and logged as a correction with no
`T_guard` number or other measurement affected. The batch's own retrospective
is explicit about this being the difference between a real anomaly and a
tooling artifact — and about checking which one it was before treating it
as either.

## M1 — the front end, and two caught-before-they-mattered bugs

M1 built the lexer, parser, semantic checks, and both fixpoint evaluators.
Two implementation mistakes were caught here, both before any dependent
measurement could be corrupted by them: a C codegen column-0 lookup used a
fixed-size buffer that could plausibly overflow at this project's own
larger scale points, replaced with an inline hash-bucket walk before any
generated C was ever compiled against it; and this session's own
Bash-tool/`wsl.exe` bridge corrupted two heredoc-based file writes
(backtick spans inside double-quoted shell content were interpreted as
command substitution by the outer shell) — caught via `Edit`, and all
later multi-line content went through `Write`/`Edit` instead of inline
heredocs.

## M2 — the magic-set transform, and an off-by-one caught before it shipped

The supplementary-chain projection (`V_i`, `docs/02-design.md` §3) got its
index arithmetic wrong on the first draft — a literal's own bound-position
variables were being dropped from the checkpoint feeding its own magic
rule. Re-derived from first principles before any generated code existed
around the wrong version. Separately, while exporting `RelationOrigin` for
later cone computation, a relation-origin mistagging bug was found and
fixed: an occurrence's magic relation was tagged with the enclosing rule's
own predicate instead of the occurrence's target predicate — caught by a
dedicated test before it could corrupt M3's cone computation months later.

## NIGHT-BATCH-03/04 and M3 — the guard, culprit cycles, and a real cone

M3 built the two-clause guard and the fallback cone. Constructing a
program where the fallback cone was genuinely non-empty (rather than the
culprit SCC always equalling the whole reachable IDB set, which is what
every naturally-arising culprit-cycle program before this session had)
took direct construction and measurement, not assumption: a new relation
read from a cycle-participating rule gets swept into an *enlarged culprit
set*, while the same relation read from a non-cycle-participating rule of
the same predicate becomes a genuine, separate cone member. The first
attempt at this construction got the placement wrong and discovered the
enlarged-culprit-set mechanism as a side effect; the second attempt, read
from the right rule, produced the intended cone.

The grammar-coverage predicate used for corpus census numbers needed three
corrective passes before it reproduced the real parser's own verdict
exactly: missing categories, a directive/period-splitting bug, and finally
two Soufflé error-fixture files with no generalizable violation category.
Each failure was diagnosed and fixed in turn, not smoothed over — the
final number (19/195 strictly compliant) is cross-checked against `dlc`'s
own parser, not asserted from the regex alone. A related correction
surfaced later, in the verification pass: the zero-arity relation count
that grammar census had reported as 11 was actually 12 — a file the
regex-based census had missed on direct read.

## M4 and the punch list — closing the gap, then finding what it revealed

M4-SIPS added demand relaxation on negated occurrences: a bound position
whose only binder is an unrestricted full-extent scan can be relaxed to
free, soundly. This closed the largest measured gap in the project — one
shape went from 194× worse than a hand-written guard at n=250 to parity,
and from 5,317× worse to a small, named residual at n=8,000
(`results/findings.md` item 1). It also produced this project's one
pre-registered prediction that missed under both counting conventions at
once (Q12, `results/superseded.md`) — recorded as a miss, not
reinterpreted after the fact toward whichever convention looked better.

The punch list found and fixed a second real bug: the adornment worklist
seeded from only the first bindable query candidate, leaving a second
independent `.output` branch at full extent. This was the reason the
guard's own practical contribution — `T_guarded < T_none` — had measured
0/12 for most of a session, reported as a finding about the guard rather
than a bug in seeding. Fixing it (seed from every candidate) flipped the
result to 9/12, and a follow-up characterization then found that the
guard's contribution shrinking with scale on those particular constructions
was itself a fixture artifact (`sibling_edges` never scaled with `n`), not
a property of demand restriction — confirmed by building the deliberately
opposite construction, which showed the contribution growing instead. See
`results/findings.md` item 2 for both curves side by side.

## What this arc says about the project's own process

Every correction above was caught by re-running something and checking
the output against an independent signal — a second re-run, an
independent implementation, the real parser, a dedicated test — not by
noticing something looked wrong on inspection. The two genuinely wrong
claims that shipped before being caught (the `wsl.exe` rc=0 artifact, the
zero-arity undercount) were both caught within the same or a following
session, both marked rather than silently corrected, and neither affected
a measurement number that had already been reported as a finding.
