# Final integration evidence

This is a fresh run, separate from Review 2 and historical measurements. Production
mode is `guarded-compact`. The Review 2 browser and repository gates passed before
this mode was implemented.

| Check | Observed result / raw evidence |
|---|---|
| `GOCACHE=/tmp/dlc-go-build-cache make verify` | PASS; `verify-final.log` (build, vet, Go/Python tests) |
| `GOCACHE=/tmp/dlc-go-build-cache make rejection` | PASS; `rejection.log`, all four source rejection categories |
| `GOCACHE=/tmp/dlc-go-build-cache make oracle ORACLE_OUTPUT=measurements/final-2026-10-20/oracle-final` | PASS; `oracle-final.log`, profiles/answers in `oracle-final/`; eight configurations, 30 executable hostile fixtures, eight explicit rejections and one parser-only long filename |
| Supplied `scripts/check_regressions.py --repo . --stage final` | PASS, 29 checks; `supplied-regressions-final.json` contains exact commands/stdout/stderr/exit codes |
| Supplied `scripts/verify_compaction.py --repo . --output … --require-oracle` | PASS; `supplied-compaction/commands.json`, profiles, source/answers and summary; eight configurations, 24 independent graph cases, native C on eight configurations and six graphs |
| `GOCACHE=/tmp/dlc-go-build-cache make oracle-compact COMPACT_OUTPUT=measurements/final-2026-10-20/compact-final` | PASS; `compact-final/`, exact byte/LF file-manifest reader, Go naive/semi-naive counters for all three modes, direct compact C, oracle and projection control |
| `GOCACHE=/tmp/dlc-go-build-cache make demo` | Launched successfully at the documented default URL; `demo-launch.log` |
| Real Chromium `tests/demo/browser_smoke.py --stage final` | PASS, 14 groups, zero page errors; `browser-release/checks.json`, losslessly compressed raw request/responses in `runs.json.gz`, one useful live compact screenshot |
| `GOCACHE=/tmp/dlc-go-build-cache go test -count=1 ./tests/demo` | PASS; `native-lifecycle-final.log`, also covers a controlled slow C compiler with a child and an actual compiled native infinite loop, both killed/reaped on timeout |

The browser interpreter was
`/tmp/ccreplay-normal-acceptance/.venv/bin/python`; its Playwright and the installed
Chromium lived outside this repository. Browser commands used actual source/facts,
actual Go runs and C compilation/execution, controlled missing-tool servers, custom
edits, failure recovery, timeout/cancellation, stale-state clearing, real compact
removals/skips and a no-op case. `browser-final/` is an earlier successful check;
`browser-release/` is the last check after readability formatting. Formatting was
checked to preserve the JS/CSS syntax trees and HTML tag structure/visible words.

`compact-final/summary.json` records each benchmark's source/fact hashes, actual
compaction actions, materializations and per-relation Go work counters. The final
source/fixture manifest and tool versions are alongside this file. Earlier
`integrated-compaction/` and supplied runs remain available without overwritten
results. Public docs use the same-engine Soufflé table and label Go counters
separately because Soufflé may internally remove aliases.

`supplied-projection-evidence/` is labelled imported audit evidence. The aggressive
variant was neither implemented nor rerun here. Locally executed conservative
projection control recorded 162 guarded / 141 compact semi-naive attempts, with
all 20 requested rows matching C/Soufflé.

No oracle check was skipped in these release gates. The optional absent-oracle and
missing-C cases were intentional controlled failure tests. LeakSanitizer, other
browser engines, native Windows/MSVC, the largest historical scales, broad
application prevalence, controlled timing/peak memory and a formal correctness
proof remain unperformed. Generated C is naive. Plain TSV has documented embedded
symbol-separator limits.

Temporary native binaries and C build sources are not part of this evidence
release. Profiles, answer files, source Datalog, commands, summaries and failed-run
records are retained. The compact verifier builds native artifacts outside the
checkout; its full recheck passed (`compact-clean-check.log`).

A subsequent `GOCACHE=/tmp/dlc-go-build-cache make verify` passed (exit 0);
see `handoff-restored-verify.log`. Earlier oracle and browser checks were not
rerun for that repository verification.
