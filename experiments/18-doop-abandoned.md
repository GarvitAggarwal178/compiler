# §3.3 DOOP — attempted, abandoned per cap

Date: 2026-08-22. Corpus ruling v1.5 §3.3, cap 3 hours, optional stretch item,
attempted only after T0 and §3.2 were complete (both were). **Outcome:
abandoned — no profile produced.** Real elapsed time on this subtask ran well
past the 3-hour cap, dominated by one hung step; recorded per instruction
("If DOOP is not producing a profile in that time, abandon and record it. This
is a stretch item, never a dependency.").

## What worked

- Cloned the real, actively-maintained, CI-tested repository —
  `github.com/plast-lab/doop` (not a stale mirror; 252 real `.dl`/`.logic`
  files under `souffle-logic/`).
- Installed Java 17 (`openjdk-17-jdk-headless`), DOOP's stated requirement.
- **`./doop -h` built and ran successfully** — full Gradle build, 25 actionable
  tasks, `BUILD SUCCESSFUL in 24m 53s` on first run (cold Gradle + dependency
  download), real usage text printed. Confirms the driver itself is sound in
  this environment.
- A **local** jar as input (`src/test/resources/java11-string-concat.jar`,
  bundled in the repo) resolves and reaches DOOP's actual analysis factory in
  seconds on a warm build — the driver, build, and local-file input path all
  work correctly.

## What did not work

1. **Maven-coordinate input (`-i org.apache.commons:commons-lang3:3.12.0`)
   hung indefinitely on Ivy dependency resolution.** Log stalled at
   `:: resolving dependencies :: org.apache.commons#commons-lang3-caller;working`
   for **over 2 hours** with constant, non-progressing CPU usage on the Gradle
   daemon — not slow, genuinely stuck. Direct network access to Maven Central
   was confirmed working (`curl` to `repo1.maven.org` succeeded instantly) at
   the same time, so this is specific to DOOP's Ivy-based resolution of its
   synthetic `<artifact>-caller` wrapper module, not a general network
   problem. Killed after confirming no progress across multiple checks.

2. **Local-jar input needs a JRE platform library DOOP fetches from a
   companion repository, and that fetch is broken.** Without `--platform`,
   DOOP auto-selected `jre1.25` (matching this environment's installed JDK)
   and failed: `Not a valid input: https://github.com/plast-lab/
   doop-benchmarks/raw/refs/heads/main//JREs/jre1.25/lib/rt.jar`. Retried with
   `--platform java_8` (the most standard, most likely to exist) — same class
   of failure, same URL pattern, `JREs/jre1.8/lib/rt.jar` also "not a valid
   input." **Both attempts completed in seconds** (not hung) — this is DOOP's
   `doop-benchmarks` companion repository not serving the expected file at the
   expected path (repo structure drift, a dead file, or a URL-construction bug
   in current DOOP against the current `doop-benchmarks` layout), not an
   environment problem on this session's side.

## Disposition

Three real, substantive attempts, three distinct and clearly diagnosed
failure points, none of them a T_guard/dlc-parsing problem (this stage never
even reached Soufflé — it failed earlier, in DOOP's own fact-generation
input-resolution stage). **Abandoned per the 3-hour cap.** Not retried with a
higher cap or a fourth variant, per this project's standing discipline
(NIGHT-BATCH-01 §0.3: "Never retry a DNF with a higher cap").

**This was optional and remains optional.** §3.2's benchmark family (five
cited/disclosed shapes, pre-registered, not run) is the measurement corpus of
record regardless of this outcome — DOOP was explicitly "motivation only,"
never a dependency of anything else in this ruling.

**If revisited later:** the fix for path 2 (local jar + explicit platform) is
almost certainly cheaper than fixing path 1 (Ivy hang) — either point
`DOOP_PLATFORMS_LIB` at a manually-obtained JRE `rt.jar`/module set instead of
letting DOOP fetch one, or find wherever `doop-benchmarks`' actual current
layout keeps its JRE libraries and adjust. Not attempted here — out of the
capped budget.
