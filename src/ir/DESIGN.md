# package ir

Runtime representation: `Value` (an int64 or an interned symbol id),
`Tuple` (`[]Value`), `StringTable` (interning), and `Relation` (tuple
storage plus one index and the instrumentation counters §3.7 requires).

**Key decision: `Value` holds an interned string id, not a `string`
directly.** Two reasons, not one: a fixed-size, comparable struct (usable
directly as a Go map key, needed for `idx0`) instead of a variable-length
string field, and repeated symbol comparisons become int32 comparisons.
The `StringTable` itself is intentionally the simplest possible
interner (a map plus a growing slice) — nothing here needs concurrent
access or eviction.

**Key decision: exactly one index, on column 0, chosen without looking
at any actual query pattern.** §3.7 explicitly permits this ("Index
selection can be naive for now"). A real index-selection pass would look
at the *adornments* the join-planning/evaluation code (§3.8/§3.9)
actually produces — e.g. `reach(x,y):-reach(x,z),edge(z,y)` only ever
probes `reach` with its first column bound, so a column-0 index is
exactly right there, but a different rule shape binding a different
argument position would want a different index, and a relation queried
under several different binding patterns would want more than one.
Building indices per-adornment, discovered from the rule set at compile
time, is the improvement a smarter pass would make; not attempted here.

**Key decision: `Insert` never touches `Stats`; the evaluator calls
`RecordSeedInsert`/`RecordIterationInsert` itself, only after a `true`
return from `Insert`.** This is what makes "exclude EDB loads" (§3.7) a
structural guarantee rather than a discipline someone has to remember:
fact-loading code that only ever calls `Insert` cannot accidentally
inflate a derived-tuple count, because there is no code path from
`Insert` alone into `Stats` at all. The alternative (an `Insert(t,
isEDB bool)` flag) would make the same guarantee only as strong as every
call site remembering to pass the right bool.

**`RelationStats.Total()` reuses this project's own established metric**
(`docs/MEASUREMENTS.md` throughout): seed-rule tuples plus every semi-
naive iteration's delta, summed once. Naive evaluation (§3.8) has no
notion of "iteration" in the semi-naive sense, so it is expected to
record everything through `RecordIterationInsert(0)` (a single bucket) —
`Total()` doesn't care which evaluator produced the counts, only that
each tuple was counted exactly once, at the moment it was newly derived.

**`EmitProfile`'s JSON shape is not a new format — it is Soufflé's own**
`-p` profile-log shape, field-for-field
(`root.program.relation.<name>.num-tuples` /
`.iteration.<i>.num-tuples`). `harness/parse_profile.py` and
`harness/tuple_report.py` already parse exactly this shape from real
Soufflé runs; matching it here means those two scripts work against
`dlc`'s own output with zero changes, which is the actual point of this
requirement ("The entire measurement apparatus depends on this being
comparable").

**Deferred, disclosed rather than silently skipped: "report both copy
conventions where a copy exists" (§3.7).** This project's
excl-copy/incl-copy distinction (`docs/MEASUREMENTS.md`,
`harness/tuple_report.py`'s `is_copy_relation`) exists because Soufflé's
own compiler sometimes emits a `COPY_T`-shaped synthetic relation (a
single-literal pass-through rule from an adorned/interm_out relation,
first found in P1's `.output`-forced case). `dlc`'s own evaluator (naive/
semi-naive, interpreted, no magic-set adornment machinery of its own)
has no code path that produces anything shaped like that yet — so
excl-copy and incl-copy trivially coincide for everything `dlc` runs
today, the same way they coincided for the large majority of this
project's own Soufflé measurements (only P1's magic-on run ever actually
diverged). Revisit if/when `dlc` grows something that could produce a
copy-shaped relation; not fabricated here ahead of that.
