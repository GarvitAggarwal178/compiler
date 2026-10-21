# Shared executable body planning

`Body` first runs the existing stable scheduling strategy. Positive bare columns
bind before their expression tests; negation and constraints require their inputs,
and a bare equality side can bind from a ground other side. SIPS supplies actual
head demand; runtime Go/C plans start with no bound variables.

When the stable schedule stalls, positive compound column terms become fresh
variables and deferred equality filters. Fresh names reserve all source variables
and initial bindings. This permits the valid joint arithmetic-binding cycle from
the review handoff without solving inverse expressions. Original AST nodes/spans
are retained separately; `Steps` records source literal indices and bound sets.
A remaining stall is an explicit error. Source range restriction is still checked
on original variables, so genuinely ungrounded inversion remains rejected.

Go evaluation, C emission and magic-set SIPS call this package. Generated schemas
and C variable types are inferred from the actual planned body. Division/modulo
zero remains a failed derivation. Missing bindings are compiler invariants rather
than implicit numeric zero. Tests cover cycles, forward equalities, same-atom
arithmetic, fresh-name collisions, source preservation and independent backends.
