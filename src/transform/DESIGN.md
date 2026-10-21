# Transform pipeline

`Transformer` accepts a validated source `*ast.Program` and its source
stratification. It returns another `*ast.Program`, so the same evaluators and
C generator can consume either an original or rewritten program.

`PassThrough` returns the input program unchanged. `magicset` computes
adornments, demand seeds, magic rules and supplementary predicates. `guard`
checks the candidate's stratification and retains full original extents for
affected dependency cones. See `magicset/DESIGN.md` and `guard/DESIGN.md` for
the algorithms and limits.

`cmd/dlc` shares one transformation path for `run`, `run-seminaive`, `emit`,
and `codegen`. It checks the resulting program again and uses its new schemas
and strata. `--transformer=guarded` selects the checked rewrite. Raw
`--transformer=magicset` emission is for inspection; execution and C
generation reject an unstratifiable result.

`ReportingTransformer` is the optional typed reporting interface used by inspection
and annotated C generation. The generator records explicit relation/rule roles
and source origins; guard retains the attempted candidate and its actual final
metadata. `guarded-compact` composes that validated final program with the pure
restricted pass in `compact/`. It never infers roles from relation-name prefixes.
