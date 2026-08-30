# Pre-registered corpus source

Pre-registered per the Phase 0.6 directive (blueprint Q5, moved from "end of week 3"
to "before any hand-probing continues"). Structural predicate only — nothing in this
directory has been run.

**Source:** `https://github.com/souffle-lang/souffle`, tag `2.5`, commit
`5682a9f12e2668ecdd26348fe63cc508bc0fcf47` (matches the installed `souffle --version`
= 2.5, see `record/DECISIONS.md`). Only the `tests/` subtree was fetched:

```
git clone --depth 1 --branch 2.5 --filter=blob:none --sparse \
  https://github.com/souffle-lang/souffle.git <dir>
cd <dir> && git sparse-checkout set tests
```

Not vendored into this repo (612 `.dl`-bearing directories across the whole `tests/`
tree, not just the 36 that pass the predicate) — too large and not ours to redistribute
wholesale. `PREREGISTERED.txt` is the reproducible output of running
`harness/build_corpus.py` against that checkout; `detail.json` records the predicate's
verdict (and why) for all 612 candidate directories, not just the 36 included ones.

**Predicate:** `harness/corpus_predicate.py` (`check_program`), applied via
`harness/build_corpus.py`, see that file's docstring for the exact two-part rule and
its known imprecisions (a mechanical text scan, not dlc's own parser).

**Process note, disclosed rather than hidden:** the predicate was first run against
`tests/evaluation` only (11 included, under the ~15 threshold) before it was rerun
over the entire `tests/` tree (36 included) once the narrower scope was recognized as
an unrequested restriction, not something the Phase 0.6 directive asked for. The
predicate logic (`check_program`) was not changed between the two runs — only which
directories were treated as candidates. Both counts and the reasoning are in
`experiments/03-completeness-counterexample-search.md`.
