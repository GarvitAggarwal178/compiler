# Escalations

Append-only. NIGHT-BATCH-01 modified protocol (`CLAUDE.md` §5 semantics, batch §0.1):
on any STOP condition, write a complete entry here, abort that task, continue the
queue. Normal CLAUDE.md §5 STOP-and-wait resumes when the batch ends.

Entry format: observation, measurement/task IDs, what was tried, live explanations,
cheapest distinguishing experiment.
