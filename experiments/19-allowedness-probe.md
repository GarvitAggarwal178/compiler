# J1 — allowedness probe

Date: 2026-08-22. Soufflé 2.5 (installed, `docs/DECISIONS.md`). Observed
behaviour only — **no definition of allowedness is proposed here**; that is
the human's decision and the hypothesis the M3 lemma is stated against.

## Method

Each case is its own `.dl` file (`tests/programs/allowedness_probe_{a..g}.dl`),
`q` declared `.input` (arity matched to what each case needs), `p` declared
`.output`. Run via `harness/j1_allowedness_probe.py` against an empty
`q.facts` (allowedness is a compile-time check, independent of data). Full
provenance: `measurements/j1-allowedness-probe-{a..g}-run/`,
`measurements/j1-allowedness-probe-summary.json`.

## The 7 outcomes

| # | Program | Outcome | Diagnostic |
|---|---|---|---|
| a | `p(X) :- q(Y), X = Y + 1.` | **accept** | none |
| b | `p(X) :- X = Y + 1, q(Y).` | **accept** | none |
| c | `p(X) :- q(Y), X > Y.` | **reject** | `Error: Ungrounded variable X ... p(X) :- q(Y), X > Y.` |
| d | `p(X) :- q(Y), Y = X + 1.` | **reject** | `Error: Ungrounded variable X ... p(X) :- q(Y), Y = X + 1.` |
| e | `p(X) :- q(Y), X = 5.` | **accept** | `Warning: Variable Y only occurs once ...` (warning, not rejection) |
| f | `p(X) :- !q(X).` | **reject** | `Error: Ungrounded variable X ... p(X) :- !q(X).` |
| g | `p(X,Y) :- q(X), Y = X.` | **accept** | none |

Exact stderr text for every reject, verbatim from Soufflé, in
`measurements/j1-allowedness-probe-summary.json`.

## Observations, not a definition

Stated as plainly as the data supports, without generalizing beyond what was
tested:

- **(a) vs (b):** identical binding structure, body literals in opposite
  order — both accept. Whatever grounding analysis Soufflé runs, it is not a
  strict single left-to-right pass over the body; reordering the constraint
  before the atom it depends on made no difference.
- **(a) vs (d):** `X = Y + 1` (X on the left) accepts; `Y = X + 1` (X on the
  right, Y already grounded, same equation up to algebraic rearrangement)
  rejects. Whatever grounds X in (a), it does not extend to "X appears
  anywhere in an equation whose other variables are grounded" — position
  relative to `=` matters, at least in this case.
- **(c):** `X > Y` alone does not ground X, even though `Y` is grounded.
  Inequality is not treated as a grounding source in this case.
- **(e):** `X = 5` grounds X with no other variable involved — a constant
  assignment alone is sufficient, and the accompanying warning is about `Y`'s
  single occurrence (an unrelated style diagnostic), not about `X`'s
  groundedness.
- **(f):** negation never grounds, confirmed directly — the single-literal
  case with nothing else in the body to fall back on.
- **(g):** `Y = X` grounds Y when X is grounded by a preceding positive atom
  — equality chaining through an already-grounded variable works, at least
  for this direct-variable-to-variable case (contrast with (d), where the
  chain went through an arithmetic expression rather than a bare variable).

Seven data points, not a specification. Case (d) in particular narrows what
"grounds via equality" can mean — it is not simply "appears in an equation
with a grounded variable on the other side."
