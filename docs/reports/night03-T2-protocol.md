# NIGHT-BATCH-03 T2 — M3 measurement protocol, end-to-end, on the identity transform

Date: 2026-08-27. Prerequisite for every M3 measurement: `dlc` decides and
emits, Soufflé evaluates. T1 cleared the printer gate; this validates the
full pipeline plumbing before any real transform exists.

## What did not work

Nothing. 25/25 scale points identical on the first run.

## Method

`harness/night03_t2_protocol.py <transformer>`. Added `dlc emit <file>
[--transformer=name]` to `src/cmd/dlc/main.go` (a `transformerRegistry` map
so a second implementation is a one-line addition, per instruction — only
`passthrough` is registered here). Pipeline per shape, per pre-registered
`SCALE_POINTS.json` point:

```
source.dl
  -> dlc emit --transformer=passthrough -> transformed.dl (parser.Print of transform.PassThrough's output)
  -> souffle -F<fixtures> -D<out> -p prof.log transformed.dl
  -> harness/tuple_report.py -> exact per-relation tuple counts (T_excl_copy)
  -> compare against souffle(source.dl) run the identical way
```

Compared: tuple totals (`T_excl_copy`, exact integer) and the shape's answer
relation, sorted-set equality. All 5 `BENCHMARK_FAMILY` shapes, every point
in `SCALE_POINTS.json` (25 points total: 4 + 5 + 5 + 5 + 6).

## The numbers

Provenance: `measurements/night03-t2/summary.json` (25 per-point rows,
`measurements/night03-t2/<shape>/<point>/{source,emitted,emit}/`).

| metric | value |
|---|---|
| attempted | 25 |
| tuple totals + answers identical | 25/25 |
| DNF / emit errors | 0 |

Since `PassThrough.Transform` returns the input program unchanged
(`src/transform/transformer.go`), `T(emit) == T(source)` is mathematically
required, not a discovery — this task validates that nothing in the new
`dlc emit` → print → Soufflé-reparse → Soufflé-evaluate plumbing silently
changes the program or drops output. 25/25 exact matches confirms the
plumbing is transparent.

## What a skeptic attacks first

- A no-op transformer proves the pipeline is transparent, not that it is
  correct for a *real* transform — the interesting failure modes (wrong
  `V_i` projection, missing negated-occurrence seeding) only show up once
  `--transformer=magicset` exists. This task's scope is plumbing, not the
  transform itself.
- `culprit_cycle` reached only n=500 (its own pre-registered ceiling, not a
  cap this task introduced) — every other shape's largest scale point
  completed too, so no scale-point truncation happened here.
- `dlc emit`'s own sema pipeline (`CheckDeclType`, `CheckAllowedness`,
  `CheckStratification`) runs before the transformer is invoked; a bug that
  caused `emit` to silently reject-and-pass-through would show up as
  `emit_error` in the summary, and there were none — but this was not
  separately stress-tested with a deliberately-rejected input in this task
  (T2's scope is the 5 shapes, which are all known-accepted programs).

## Verdict

**T2: DONE, gate cleared.** 25/25 identical. `dlc emit` exists, is wired to
a registry that takes one line to extend, and the full measurement pipeline
(`dlc` decides+emits, Soufflé evaluates) is validated end-to-end. Ready for
a real second `Transformer` implementation.
