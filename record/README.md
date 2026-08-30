# Record — append-only audit trail

`DECISIONS.md`, `ESCALATIONS.md`, `MEASUREMENTS.md`, `OPEN_QUESTIONS.md`,
`SESSION_LOG.md`. These are never rewritten, reordered, deduplicated, or
tidied — a superseded row stays, with a new row added alongside it citing
what replaced it. Their messiness (partial answers, corrections logged
inline, entries that turned out not to matter) **is** the audit trail; a
cleaned-up version of any of these files would be a different, less
trustworthy document.

**These files predate this restructure and are not updated by it.** Every
path they cite (a `docs/reports/*.md` filename, a `docs/dlc-blueprint.md`
section number) is the path that file had *at the time the row was
written* — not the current path under `experiments/`, `specs/`, or
`docs/`. To translate an old path cited in one of these files to its
current location, see `docs/rename-map.csv`. Paths into `measurements/
<id>/` are unaffected either way, since that tree did not move.
