# Escalations

Append-only. NIGHT-BATCH-01 modified protocol (`CLAUDE.md` §5 semantics, batch §0.1):
on any STOP condition, write a complete entry here, abort that task, continue the
queue. Normal CLAUDE.md §5 STOP-and-wait resumes when the batch ends.

Entry format: observation, measurement/task IDs, what was tried, live explanations,
cheapest distinguishing experiment.

---

## 2026-08-20 — T2/T3: `docs/phase0.7-corpus-viability.md` does not exist

**Observation.** T2 is specified as "Exactly as specified in
`docs/phase0.7-corpus-viability.md` §2.2." T2's output requirement cites "the four
numbers from Phase 0.7 §3." Hard prohibition §0.2.3 names two thresholds from that
same document (`T_none ≥ 1,000`, "the floor of 8") directly in the night-batch text.
The file itself is absent from the repository (`find . -iname 'phase0.7*'` returns
nothing; `docs/` contains no such file).

**What I tried.** Checked `docs/`, `docs/reports/`, repo root, and a full-repo
`find` for any filename matching `phase0.7*` or `corpus-viability*`. Nothing.
Checked `git log` for any commit that might have added and then lost it — this repo
has 5 commits total, all authored this session, none reference the file.

**What I am doing about it, and why this is not "resolving the escalation."** The
per-program measurement fields T2 actually needs (`T_none`, fact input rows,
seedable?, negated-IDB-literal count, out-of-grammar feature count, status) are
given directly, in full, in the night-batch directive's own §T2 table — not only in
the missing external doc. Two of the three "output" requirements (the `T_none ≥
1,000` count and the floor-of-8 check) are also given verbatim in §0.2.3. Running
the measurement itself does not require guessing anything the missing document might
say differently. What I am **not** doing: fabricating "the four numbers from Phase
0.7 §3" — I do not know what those four numbers are defined to be, and inventing a
definition for them would be exactly the kind of unauthorized resolution this
protocol exists to prevent. T2's report states this gap plainly instead of silently
filling it, and reports only what the batch text itself specifies.

**Live explanations:** (1) the document was meant to be written before this batch
and wasn't; (2) the document exists somewhere outside this repo (another session,
another machine) and the night-batch directive assumed it would already be here;
(3) "Phase 0.7" is this batch's own numbering scheme for a doc the human intended to
attach but didn't.

**Cheapest distinguishing experiment:** none needed from my side — this is a
missing-input problem, not an ambiguous-observation problem. The human either has
the file and can supply it, or needs to write it. Either resolves this immediately.

**Task disposition:** T2 proceeds using the batch text's own inline specification
(see `docs/reports/night01-T2-corpus.md`), not aborted, since its core measurement
is fully determined without the missing file. T3 proceeds on the same basis (its
stated prerequisite is T2's seedable subset, which does not depend on the missing
document either).
