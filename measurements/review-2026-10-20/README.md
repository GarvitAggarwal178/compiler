# Review 2 integration evidence

New local evidence; original measurements and the pass-one run were not overwritten.

Phase A passed `make verify`, `make rejection`, `make oracle` with Soufflé 2.5,
and the supplied `check_regressions.py --stage review2` (28 checks). The new
regressions include arithmetic cycles across Go/C/oracle, deterministic rejection,
signed queries, actual inspection, full-output metadata and exact output bytes.
Guarded Soufflé totals remained 974 / 6102 / 681 / 207 / 286 / 273 / 417 / 676
on the eight labelled configurations in `phase-a-oracle/summary.json`.

Phase B was runnable before phase C: `phase-b-verify.log`,
`phase-b-acceptance.json`, and `browser-review2-3/checks.json` record passing
repository, supplied black-box and 13-group real browser gates. Browser request
snapshots, responses and command outcomes are in `browser-review2-3/runs.json`.
The two screenshots in that run came from live Chromium execution.

Failures encountered and corrected remain explicit:

- Initial `make verify` failed because the extracted handoff was inside the Go
  tree and contained mixed reference packages. Moving it outside fixed the scan;
  `phase-a-verify-1.log` retains that failure.
- A focused initial planner cleanup failed after making unbound arithmetic variables
  panic: the optional index lookup legitimately probes for unavailable values.
  It now checks availability before evaluation; the subsequent backend tests pass.
- Initial HTTP tests could not open loopback sockets inside the sandbox. The
  authorized tests were rerun with local socket access; this was an environment
  failure, not a skipped test.
- A backend test reused a nonzero JSON target and retained an omitted pointer
  from an earlier response. Resetting that test target fixed the assertion;
  fresh HTTP response JSON already omitted the unavailable comparison.
- `browser-review2-1` exposed a narrow SVG path target. The UI now has a wider
  path hit area and keyboard activation. `browser-review2-2` passed the mutations
  but the test read visible C text before opening its collapsed panel. The test
  now opens the actual panel. Both failed-run checks/raw responses are retained.

No recorded answers substituted for live work. Source/data errors, timeout,
process cleanup, missing C/oracle and no-output cases were exercised explicitly.
The final-state browser recheck and further native lifecycle checks are recorded
under the separate `../final-2026-10-20/` run.
