# package lexer

A single-pass, hand-written scanner (`Lexer.Next`) over raw source bytes,
producing one `token.Token` at a time; `Tokenize` drives it to completion.
No lexer generator, no regexp — M1-BUILD.md §2 requires standard library
only, and a hand-loop over bytes is simpler to make panic-free than
regexp-based scanning would be (no possibility of a pathological pattern,
no hidden allocation surprises to reason about under fuzzing-style input).

**The `.`-disambiguation (§3.1's "the real content here").** On seeing
`.`, the lexer peeks at the next byte only: if it is an ASCII letter, it
greedily consumes an identifier-shaped run right after the dot and checks
it against the three known directive names (`decl`, `input`, `output`).
A match emits the directive token; anything else (`.foo`, `.de`) emits a
single `ERROR` token spanning the whole `.foo` — not a `DOT` followed by a
stray `IDENT`, because nothing in the grammar ever has a bare identifier
immediately following a bare `.`, so splitting it that way would just
relocate the eventual parse error one token later for no benefit. If `.`
is *not* followed by a letter, it is unconditionally `DOT` (the common
case: a clause terminator). This is a one-byte lookahead, resolved
entirely in the lexer, exactly as §3.1 asks — the parser never sees a
`.` and has to guess what it meant.

**Two deliberate disagreements with Soufflé's own lexer**, both discovered
by NIGHT-BATCH-02 T2 (`experiments/26-hostile-source-corpus.md`) and kept
because gate three of §3.3 explicitly permits disagreement ("report it,
do not adjust to match"):

1. **Unterminated block comments are an `ERROR` token here**, not a silent
   swallow to end-of-file. T2 found real Soufflé accepts
   `comment_unterminated_block.dl` outright (with only a warning that the
   relation ends up with no rules) — it silently treats everything after
   an unclosed `/*` as comment text through EOF. `dlc` treats this as a
   mistake worth surfacing: a source file that most likely lost a rule to
   a typo should not compile silently.
2. **Non-ASCII bytes are always an `ERROR`, one Unicode rune at a time**
   (via `utf8.DecodeRune`, not byte-at-a-time — so a 2-byte character like
   `é` is one `ERROR` token, not two). This matches Soufflé's own
   ASCII-only identifier behavior (same report, `lexical_unicode_identifiers.dl`)
   but the *reason* is independent: `token.Position.Col` (see
   `token/DESIGN.md`) counts bytes, which only stays meaningful as long as
   every token's source text is guaranteed ASCII.

**Never panics.** Every scan step bounds-checks via `peekByte` before
indexing; there is no `l.src[l.offset]` reachable without a preceding
"is there a byte here" check except the one inside `advance`, which is
only ever called immediately after such a check (or, at the top of
`Next`, after the explicit `l.offset >= len(l.src)` EOF test). An
unterminated string and an unterminated block comment are the two "ran
off the end while inside a token" cases, and both produce an `ERROR`
token rather than any special control flow.

**Underscore is not always the wildcard.** `_` alone lexes as
`token.UNDERSCORE`; `_foo`, `_1`, etc. lex as ordinary `IDENT` (real
Soufflé accepts underscore-led identifiers as legal variable names, per
the same corpus scan). The rule is applied after maximal-munch identifier
scanning, not by special-casing `_` at the switch in `Next` — a bare `_`
is simply the case where the scanned identifier text happens to equal
`"_"`.
