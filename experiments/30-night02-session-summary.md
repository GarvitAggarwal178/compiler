# NIGHT-BATCH-02 — morning summary

Date: 2026-08-23. All 9 tasks reached a defined outcome; no aborts, no
DNFs, `src/` never touched.

## 1. Escalations

**None.** No CLAUDE.md §5 STOP condition fired this batch. One correction
was issued mid-batch (below) but it was fully diagnosed and resolved
within the same session, not left open — it does not meet the bar for an
`ESCALATIONS.md` entry (that file already exists from NIGHT-BATCH-01;
nothing was appended this batch).

**Correction, logged for transparency:** T5 originally reported "Soufflé
exits 0 on a stratification error," based on one interactive Bash-tool
reading through the `wsl.exe` bridge. T9 re-ran the same case 5 times via
reliable (file-redirected) invocation — all 5 returned rc=1. Root cause:
the bridge racing on live-streamed interactive output, not a real Soufflé
behavior. Retracted in `docs/reports/night02-T5-guarded.md` (marked, not
silently edited); commit `9df90dd`. No `T_guard` number was affected.

## 2. Task outcome table

| task | outcome | headline |
|---|---|---|
| T1 | done | allowedness probes h–o: h (body-only ungrounded var) rejects |
| T3 | done | 38 fixtures, idempotent, one bug caught (edge direction) before trusting numbers |
| T4 | done | 5 shapes × 24 scale points, no abort, `transitive_closure_bound` E_recoverable=0 (disclosed null result) |
| T5 | done | 4 shapes guarded, 20 points, no abort; `culprit_cycle`'s general derivation hit an unstratifiable cycle, discarded |
| T6 | done | 3/4 shapes share Θ(n) ratio growth class, differ 2 orders of magnitude in constant; `culprit_cycle` disagrees (flat) |
| T2 | done | 39 hostile files, 31 accept/8 reject; found a `probe0.py` UnicodeDecodeError bug, fixed |
| T8 | done | 195-file grammar census; found 11 "in-grammar" files use zero-arity relations blueprint §4 doesn't admit |
| T9 | done | 7-class diagnostic catalogue; 13/13 rejection-corpus cases cross-checked, all consistent |
| T7 | done | `q` confirmed to survive Soufflé's inliner; culprit cycle confirmed; dead-rule check fires at 4/5 points |

## 3. T1 — allowedness probes h–o

| case | program | outcome |
|---|---|---|
| **h** | `p(X):-q(X),Y>3.` | **reject** — `Ungrounded variable Y` |
| i | `p(X):-q(Y),5=X.` | accept |
| j | `p(X):-q(Y),X+1=Y.` | reject |
| k | `p(X):-q(Y),Z=Y+1,X=Z+1.` | accept |
| l | `p(X):-q(Y),X=Y,Y=X.` | accept |
| m | `p(X):-!q(Y),X=1.` | reject — `Ungrounded variable Y` |
| n | `p(X):-q(Y),X=-Y.` | accept |
| o | `p(X,Y):-q(Z),X=Z,Y=X+1.` | accept |

**h is the load-bearing case.** `Y` never appears in the head, only in a
constraint — Soufflé still rejects it. Answers the named question: the
allowedness condition Soufflé enforces quantifies over every variable in
the clause, not only head variables. No definition proposed (human's
decision). Full table: `docs/reports/night02-T1-allowedness.md`.

## 4. T5 — the headline, per shape, never aggregated

| shape | `T_none`→`T_souffle`→`T_guard` range | `T_souffle/T_guard` range |
|---|---|---|
| `same_generation_negation` | up to 212M → 127M → 29,134 | 68.4× → 4,370.6× |
| `ancestor_nonancestor` | up to 64M → 25.4M → 408,000 | 4.0× → 62.2× |
| `culprit_cycle` | up to 84,105 → 83,295 → 83,290 | ~1.0× (no contribution) |
| `reachability_complement` | up to 64M → 40.8M → 9,615 | 157.2× → 4,243.9× |
| `transitive_closure_bound` | up to 25.5M → 101 → *(excluded)* | E_recoverable=0, no guard needed |

Answer equality against the untransformed baseline held at every scale
point, every shape. Full tables: `docs/reports/night02-T5-guarded.md`.

## 5. T6 — cross-shape growth, including disagreement

`same_generation_negation`, `ancestor_nonancestor`, `reachability_complement`
all fit **T_none/T_souffle ≈ Θ(n²), T_guard ≈ Θ(n), ratio ≈ Θ(n)** — same
class, three structurally different recursion shapes (tree, chain, chain).
**They disagree on magnitude by up to 2 orders of magnitude** at
comparable `n` (`ancestor_nonancestor` 4×–62× vs. `reachability_complement`
157×–4,244× vs. `same_generation_negation` 251×–4,371×), unexplained by
this task. `culprit_cycle` disagrees on **class**, not just magnitude: flat
~1.0×–1.1× throughout, no growth, and its 5 points don't admit a clean
power-law fit at all. `transitive_closure_bound`'s `T_souffle` is flat
Θ(1) (no guard needed). Full account:
`docs/reports/night02-T6-scaling-crossshape.md`.

## 6. What did not work

- **T3:** same-generation reachability walk used the wrong edge direction
  on the first attempt (root has no outgoing `child→par` edge); caught
  before trusting the number.
- **T5:** general-adornment derivation for `culprit_cycle` recreated an
  unstratifiable negative cycle (`Unable to stratify {m_q,p_bf,q_bf,s_bf}`)
  — discarded per requirement, not fixed up. This is P5's own reason for
  existing, confirmed by construction.
- **T2:** `probe0.run_cmd` crashed on non-UTF-8 stderr (unicode-identifier
  case) — fixed (`errors="replace"`), retroactively safe.
- **T8:** found (not fixed — out of scope) that 11 of `IN_GRAMMAR.txt`'s
  195 "in-grammar" files declare zero-arity relations, which blueprint §4
  doesn't admit — a gap in NIGHT-BATCH-01 T5's mechanical predicate.
- **T9/T5 correction:** see §1 above.
- **T7:** first check of "does the second `p` rule fire" used only n=20
  and looked dead; checking the other 4 already-committed T4 points showed
  it fires at all of them — n=20 alone is the (likely fixture-sparsity)
  anomaly.

## 7. What a skeptic attacks first

- `same_generation_negation`'s 4,371× ratio is explained by a fixture
  accident (query node = tree root, which has no parent) — not evidence of
  a generally-superior transform. Untested: a fixture where the query node
  is not the root.
- `culprit_cycle`'s ~1.0× contribution and T7's finding that Soufflé's own
  automatic transform restricts only `p` (leaving `q`/`s` at full,
  untransformed size) are the same fact observed twice — the hand guard
  isn't underpowered, a cyclic-negation-safe guard for this shape provably
  can't do more with the general derivation attempted.
- The `ancestor_nonancestor` vs. `reachability_complement`/
  `same_generation_negation` magnitude gap (same Θ(n) class, ~50–70× lower
  constant) is reported, not explained.
- T7's `--inline-exclude=q` conclusion ("q was never inlinable") rests on
  the flag having zero observable effect, not on inspecting Soufflé's own
  transformed-program output — plausible, not independently confirmed.
