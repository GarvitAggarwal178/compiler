# dlc

A from-scratch Datalog compiler (Go), with a magic-set transform guarded
by a soundness check the transform itself can violate.

## The headline result

Soufflé — this project's oracle — restricts a relation whose body
*contains* negation, but never restricts the relation *inside* the
negation, computing it in full every time. `dlc`'s guarded transform
closes that gap: **46.0×–1,342.7× fewer derived tuples than Soufflé's own
transform**, same stratified-negation programs, same scale points
(`results/findings.md`).

**The arc is the actual result.** Before this project's demand-relaxation
rule existed, `dlc`'s own mechanical transform was **194× worse than a
hand-written guard** at n=250, widening to **5,317× worse** at n=8,000.
One structural rule — a bound position whose only binder is an
unrestricted scan carries no demand information, and can be relaxed to
free, soundly — closed that to **parity** (252 vs. 285) at n=250. See
`results/presentation.html`.

## Quick start

```sh
go build ./...
go test ./...
# Show the analyzer rejecting a malformed program, with a diagnosis:
go run ./src/cmd/dlc explain experiments/explain-samples/rejection_allowedness.dl
# Run one measurement (untransformed vs. guarded):
go run ./src/cmd/dlc run      <file.dl> <factsDir> <outDir>
go run ./src/cmd/dlc run      <file.dl> <factsDir> <outDir> --transformer=magicset
```

`docs/06-reproduce.md` has the full path from a clean clone to
reproducing a committed headline number.

## Reading order

| Order | Document | What it answers |
|---|---|---|
| 1 | `docs/01-problem.md` | What's the gap, and why does it matter? |
| 2 | `docs/02-design.md` | What does the compiler do, right now? |
| 3 | `results/findings.md` | What was actually found, with numbers? |
| 4 | `results/claims.md` | For any specific claim: what's the evidence, and how strong is it? |
| 5 | `docs/05-limitations.md` | What would a reviewer push on first? |

`experiments/README.md` indexes every measurement behind every claim.
`docs/design-history.md` and `docs/project-log.md` cover how the design
and the sessions actually went, including what was wrong and how it was
caught.

## What this is not

- Not general magic sets under negation — a closed subset grammar, a
  detector plus a fallback, not a reimplementation of Balbin et al.
- Not validated against multiple oracles — Soufflé only.
- Not a timing result — every number is a tuple count, no wall-clock.
- Not shown to matter on real code — correctness is demonstrated on
  constructed programs; the guard's shape was found in **0 of 817**
  real-world files (`results/findings.md` item 4).
