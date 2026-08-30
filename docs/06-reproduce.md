# Reproducing this project, from a clean clone

## 1. Toolchain

- **Go 1.25+** (`go.mod` declares `go 1.25.0`; this project was built and
  measured under Go 1.26.0, installed via `apt install golang-go`).
- **Soufflé 2.5**, exactly. Install from the release `.deb`, do not build
  from source and do not take a later version — this project's own
  version-risk check (`experiments/16-souffle-version-risk.md`) confirmed
  the relevant negation-isolation behaviour is unchanged on a later
  development snapshot, but 2.5 is the pinned oracle every committed
  number was measured against:

  ```sh
  wget https://github.com/souffle-lang/souffle/releases/download/2.5/x86_64-ubuntu-2404-souffle-2.5-Linux.deb
  sudo apt install ./x86_64-ubuntu-2404-souffle-2.5-Linux.deb
  souffle --version   # confirm: Version: 2.5 (2.5)
  ```
- **Python 3** for `harness/` and `tools/` (standard library only, no
  `requirements.txt` — nothing outside it is imported).

This project was developed on WSL2 (Windows host, Ubuntu guest); the
harness scripts assume a POSIX shell and are invoked with absolute paths
rooted at `/root/compiler` in several places. Adjust the path prefix if
cloning elsewhere.

## 2. Build and test

```sh
go build ./...
go vet ./...
go test ./...
make rejection   # the four rejection grounds, live, one sample each
```

All three `go` commands must be clean and `make rejection` must print a
`REJECT ground=...` line for each of the four grounds
(`experiments/explain-samples/rejection_*.dl`).

## 3. Fixtures

Every generator in `harness/fixtures_lib.py` is seeded from a constant
recorded in the fixture file or the calling script — two runs of the same
command must produce byte-identical `.facts` output. Regenerate a fixture
by re-running the harness script that produced it (named in the relevant
row of `record/MEASUREMENTS.md` or the experiment record that cites it);
never hand-edit a `.facts` file.

## 4. The gate commands

- **Correctness, differential against Soufflé:** `harness/differential.py`
  runs both engines on the same `.dl`/`.facts` pair and reports set
  equality on every output relation (symmetric difference on mismatch,
  never a bare pass/fail).
- **Rejection corpus:** the 13 cases under `tests/rejection/`, one per
  ground per malformed-input shape, checked against Soufflé's own
  accept/reject verdict (`experiments/28-souffle-diagnostic-catalogue.md`).
- **Guard blast radius:** `bin/conecheck` (built from `src/`, gitignored)
  reports culprit/cone membership per program in the cone corpus
  (`tests/corpus/CONE_CORPUS/`).

## 5. One worked example, end to end

Reproduces the growing-sibling construction behind
`results/findings.md` item 2 (`experiments/56-mass-ratio-
characterization-construction.md`):

```sh
cd /root/compiler   # or wherever this repo was cloned
python3 harness/punchlist2_item2_growing_sibling.py
```

This regenerates `cc_growing_sibling.dl`'s facts at n=20/50/100 from the
recorded seed, runs `dlc` untransformed and guarded plus Soufflé, and
prints a `T_none`/`T_guarded`/ratio table. The expected output matches
`measurements/punch-list-2/item2-growing-sibling/summary.json`: ratios
**2.14×, 6.38×, 12.69×** at n=20/50/100 respectively, with the declined
portion of `T_guarded` bit-for-bit identical to `T_none`'s own cost for
those same relations at every point. A different ratio means either the
toolchain versions above don't match, or something in `src/`
(Lane A) changed — check `git diff` against the tag `pre-restructure`
first.

## 6. What is not reproducible from this repo alone

Soufflé itself is not vendored (its `tests/` directory was sparse-checked
out for corpus construction, `record/DECISIONS.md`, and is not committed
— it is not this project's IP to redistribute). Programs cited from that
corpus by name (`tests/corpus/IN_GRAMMAR.txt`, `PREREGISTERED.txt`) need a
matching Soufflé source checkout at tag `2.5` to re-derive the same file
set; the corpus-admissibility *numbers* themselves are committed and don't
require it.
