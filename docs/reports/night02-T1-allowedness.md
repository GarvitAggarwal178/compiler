# NIGHT-BATCH-02 T1 — allowedness probes h–o

Date: 2026-08-23. Soufflé 2.5, same protocol as J1
(`docs/reports/J1-allowedness-probe.md`): each program run standalone with an
empty `q.facts`, accept = returncode 0, reject = nonzero. Measurement IDs
`night02-t1-allowedness-probe-{h..o}`.

**No definition of allowedness is proposed here.** Observed behaviour only —
see CLAUDE.md non-negotiable and J1's report for why.

## Results

| # | Program | Outcome | Diagnostic |
|---|---|---|---|
| h | `p(X) :- q(X), Y > 3.` | **reject** | `Ungrounded variable Y` (+ warning: `Variable Y only occurs once`) |
| i | `p(X) :- q(Y), 5 = X.` | accept | warning only (`Variable Y only occurs once`) |
| j | `p(X) :- q(Y), X + 1 = Y.` | reject | `Ungrounded variable X` |
| k | `p(X) :- q(Y), Z = Y + 1, X = Z + 1.` | accept | none |
| l | `p(X) :- q(Y), X = Y, Y = X.` | accept | none |
| m | `p(X) :- !q(Y), X = 1.` | **reject** | `Ungrounded variable Y` (+ warning: `Variable Y only occurs once`) |
| n | `p(X) :- q(Y), X = -Y.` | accept | none |
| o | `p(X,Y) :- q(Z), X = Z, Y = X + 1.` | accept | none |

## Case h — the load-bearing case

`Y` in `p(X) :- q(X), Y > 3.` does not appear in the head. It appears only in
a constraint (`Y > 3`), never in a positive atom. **Soufflé rejects it**, with
the identical diagnostic form (`Ungrounded variable Y`) used for head-only
violations in J1 case (c)/(d) and this file's case (j)/(m).

Named question this answers: *does the allowedness condition Soufflé enforces
quantify over every variable appearing anywhere in the clause, or only over
variables appearing in the head?* Case h's rejection is evidence for the
former — Soufflé did not let `Y` pass merely because dropping it from the
head would have made the clause well-formed. No conclusion is drawn beyond
that; this is the human's definition to write.

## Other cases, as observed

- **i** — `5 = X` (constant on the left, target variable on the right)
  accepts. Combined with J1 case (a) (`X = Y + 1`, target on the left)
  and case (b) (equation literal ordered before the grounding atom, also
  accepted), the equation's left/right position of the *constant* does not
  by itself block grounding — but see (j), where reversing which side holds
  the *bare variable* does block it.
- **j** — `X + 1 = Y` rejects `X`, even though `Y` is grounded by `q(Y)` and
  the equation could in principle be solved for `X`. The rejected variable is
  the one that is not alone on either side as a bare identifier.
- **k** — a three-step chain (`Y` from `q(Y)`; `Z` from `Y`; `X` from `Z`)
  accepts. Grounding is not limited to a single equation hop.
- **l** — `X = Y, Y = X` accepts, but note `Y` is already grounded directly
  by `q(Y)` earlier in the body — this program does not isolate "mutual,
  neither independently grounded" as a case, since `Y` was never ungrounded
  to begin with. Recorded as run, not adjusted or re-derived.
- **m** — `Y` appears only inside `!q(Y)`, never in a positive atom; rejects
  with the same diagnostic shape as J1 case (f) (`!q(X)` with `X` the sole
  head variable). Negation does not ground under any binding position tried
  so far (sole occurrence, as here, or occurrence alongside a positive atom
  as in J1 (f)).
- **n** — unary minus on the right-hand side (`X = -Y`) accepts; grounding
  through an expression is not blocked by the presence of a unary operator.
- **o** — grounding chains through a *head* variable (`X` grounded from `Z`
  by `q(Z)`, then `Y` grounded from `X`) and both `X` and `Y` end up bound;
  accepts.

## Provenance

All 8 measurement IDs under `measurements/night02-t1-allowedness-probe-{h..o}/`
(`cmd.txt`, `stdout.txt`, `stderr.txt`, `env.txt`, `meta.json` each), full JSON
summary at `measurements/night02-t1-allowedness-probe-summary.json`. Source
programs: `tests/programs/allowedness_probe_{h..o}.dl`. Runner:
`harness/night02_t1_allowedness.py`. No cap fired (all 8 runs completed near-
instantly, well under the 30-minute task cap).
