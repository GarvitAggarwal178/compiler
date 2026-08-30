# package token

Defines the token vocabulary and source-position types shared by every
downstream pass: `Kind` (an enum covering every terminal in blueprint §4's
grammar plus the three directives and the ERROR kind), `Position`
(byte offset + 1-based line/col), `Span` (a half-open `[Start, End)`
range), and `Token` itself.

**Key decision: every `Token` carries a `Span`, no exceptions, including
`ERROR` tokens.** M1-BUILD.md §3.1 calls this out directly ("retrofitting
spans is miserable") and it is the right call for a second reason too:
`sema`'s diagnostics (arity, type, allowedness, stratification — see
`experiments/28-souffle-diagnostic-catalogue.md`) all point at a specific source
location, and threading that back to a token that never recorded one would
mean rebuilding position information from scratch at every later pass.
Paying the cost once, in the lexer, is cheaper than paying it four times
later.

**Key decision: `Col` counts bytes, not runes.** This looks like a
shortcut but isn't one for this grammar: `night02-T2-hostile.md` (real
Soufflé, `lexical_unicode_identifiers.dl`) shows Soufflé's own lexer
rejects any non-ASCII byte as an invalid token outright, and `dlc`'s
lexer makes the same choice (see `lexer/DESIGN.md`) — so by the time a
`Token` exists, its source text is guaranteed ASCII, and byte columns and
rune columns are the same number. This is a consequence of the ASCII-only
decision, not an independent shortcut; if that decision ever changes,
`Col` would need to change with it.

**Non-obvious decision: `ERROR` is a `Kind`, not a separate error-reporting
side channel.** An error token slots into the same `[]Token` stream the
parser already walks, at the exact position lexing found trouble, with
`Message` holding what went wrong and `Text` holding the raw offending
source. This is what lets the lexer "continue past an error so a file
yields all its errors" (§3.1) without the parser needing a second,
separate diagnostics list to reconcile against token positions.
