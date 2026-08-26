# package ast

Defines every node type before any parse function exists (§3.2's own
instruction) so the parser, sema (decl/type/allowedness/stratification),
naive/semi-naive eval, and the pretty-printer are all written against one
fixed shape from the start, instead of the shape drifting under whichever
pass gets written first.

**Key decision: sum types as Go interfaces with a private marker method**
(`literalNode()`, `termNode()`, `arithNode()`), not a single struct with a
`Kind` tag field and a union of optional pointers. This mirrors the
grammar's own `literal ::= atom | '!' atom | constraint` and
`arith ::= ... | var | number | string` productions directly — a
`Literal` value's *Go type* is which alternative it is, so a type switch
in sema or eval is exhaustive-checkable by the compiler in a way a
`switch node.Kind` over an untyped tag is not. The cost is one-line
marker methods on every concrete type; that is cheaper than the
alternative's cost (every struct carrying fields that are meaningless for
most of its own variants).

**Key decision: `Arith` embeds `Term`, not the reverse, and does not
duplicate `termNode()`.** Every `Arith` value is legally a `Term`
(`term ::= arith | '_'`), but not every `Term` is an `Arith` (`Wildcard`
isn't). Embedding lets `*BinaryExpr` etc. satisfy `Term` automatically
once they satisfy `Arith`'s own two methods, without a second explicit
`termNode()` implementation per arith type — the embedding *is* the
"arith is-a term" relationship, expressed once.

**`ast.Equal`, not `reflect.DeepEqual`, for round-trip comparison.**
Every node carries a `Span` (per `token.DESIGN.md`'s rationale, inherited
here), and parse→print→reparse necessarily produces different spans on
the second parse (the printed text isn't byte-identical to the source).
A field-by-field `DeepEqual` would therefore never accept even a
perfectly correct printer. `Equal` is a hand-written, one-case-per-type
comparison that checks everything *except* `Span` — more code than
`reflect.DeepEqual`, and the only way to skip exactly the right fields.
This is what `harness/round_trip_scaffold.py`'s gate (§3.3, gate two)
will actually call through the CLI once that JSON boundary exists; the
comparison itself lives here so it is exercised by Go's own type system,
not re-implemented in Python against a JSON dict.

**Deferred, not solved here: JSON serialization.** The CLI (§3.3) needs
to hand a `*Program` to Python as JSON for the parse-coverage and
round-trip harnesses. Interface-typed fields (`Literal`, `Term`, `Arith`)
don't marshal to a self-describing JSON shape via the default
`encoding/json` behavior — a discriminated-union wrapper (a `"kind"`
string field alongside the payload) will be needed at the CLI boundary.
Left to §3.3, which is where that boundary is actually built; `ast`
itself has no JSON tags and no marshaling code, on purpose, so this
package's shape is not constrained by its own eventual serialization.
