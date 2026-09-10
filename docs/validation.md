# Race detection, fuzzing, and coverage

Alpha development deferred three checks to shorten the implementation cycle:
race detection, active fuzzing, and coverage collection. This page records how
each one runs now, what budget it uses, and what the restored checks found.

## Coverage

`make check` runs the module suite once with instrumentation instead of running
it twice, so measuring coverage costs no extra test run. `scripts/coverage.sh`
executes `go test -count=1 -coverprofile=... -coverpkg=./... ./...` for every
runtime and consumer module in `.gomodules` and writes one profile per module
under `artifacts/coverage/`. `-coverpkg=./...` credits a package when another
package's test executes its code, which is how the engine, the rule catalog, and
the source mapper are actually exercised.

`scripts/coverage-report.py` merges the profiles, treats a statement as covered
when any profile records a hit, and reports statement coverage per package, per
area, and overall. Packages without executable statements report no percentage
and are left out of the totals. The gate fails when an area misses its target.

| Area | Packages | Target |
| --- | --- | --- |
| Engine and scoring | `unswell` | 90% |
| Rules | `rule`, `ruleset`, `builtin` | 90% |
| Configuration | `config`, `internal/appconfig` | 90% |
| Source mapping | `extract`, `document`, `internal/mapping` | 90% |
| Overall | every measured package in every module | 85% |

The measurement on `82c26019` reached 91.9% for engine and scoring, 92.7% for
rules, 91.9% for configuration, 91.1% for source mapping, and 86.6% across every
measured package in every module.

The targets come from the September 7, 2026 specification through
[#27](https://github.com/stokaro/unswell/issues/27). Nothing is excluded to reach
them: command packages, research code, and the example consumer all count toward
the overall figure. `bash scripts/coverage.sh --self-test` builds two synthetic
profiles and proves the gate accepts the passing one and rejects the other.

## Race detection

`make race` runs `bash scripts/modules.sh test -race -timeout 40m` across the
same modules. Race instrumentation multiplies runtime. The root package alone
needs more than the default ten minute panic threshold on a shared runner, so
the target sets the timeout explicitly. A slow host then reports a slow run
rather than a hang.

## Active fuzzing

`make fuzz` runs `scripts/fuzz.sh`, which asks `go test -list '^Fuzz'` for the
targets in every runtime and consumer module and then fuzzes each one for a
fixed budget. Discovery is the point: a hand-written list drifted during alpha
development and missed two targets. The default budget is 10 seconds and two
workers per target, matching the budget CI used before the deferral. Pass
`--fuzztime` and `--parallel` for a longer campaign, and `--list` to print the
discovered targets without running them.

Executing seed corpora during an ordinary `go test` run is not fuzzing. Only a
run with `-fuzz` generates new inputs. When a target fails, Go writes the
minimized input under that package's `testdata/fuzz` directory; commit it as a
regression case together with the fix.

`bash scripts/fuzz.sh --self-test` generates a temporary module with one passing
and one failing target, checks that discovery finds both, and checks that the
failing target fails the run.

## Restored evidence

The deferral is over. Each check ran locally on the tree that restored it, and
CI runs all three on every pull request and every push to main.

Race detection found no data race. The suite ran on `34e84339` with the minimum
compiler from `go.mod` (Go 1.25.0, `CGO_ENABLED=1`, darwin/arm64). It covered
the 35 packages that have tests in the runtime and consumer modules, and took
1,202 seconds of test time. The root package alone took 469 seconds under
instrumentation, which is why `make race` sets its own timeout.

Active fuzzing found a panic. `scripts/fuzz.sh` discovered 15 targets, two of
which the hand-written list had never run: `FuzzComparisonPlan` in
`research/annotation/evaluation` and `FuzzResearchPredictionInputs` in
`research/annotation/training`. Within five minutes `FuzzSourceMap` produced a
two-byte Python comment, `#>`, whose opener and terminator shared their bytes.
Comment narrowing removed both and inverted the span, which panicked in
`directiveText`. [#145](https://github.com/stokaro/unswell/pull/145) fixes it,
keeps the minimized input as a seed, and adds readable cases for the same
mistake. A macOS run also exposed a shutdown race in the MCP server tests, fixed
in [#146](https://github.com/stokaro/unswell/pull/146).

The campaign then ran again on `82c26019` with Go 1.27.1 on darwin/arm64. Every
one of the 15 targets fuzzed for 300 seconds with two workers, 4,545 seconds in
total, and found nothing further.

Coverage meets the targets. [`validation/coverage.json`](validation/coverage.json)
records the measured areas, every measured package, and the commit, compiler and
platform the measurement used. `make cover` regenerates it as
`artifacts/coverage/summary.json`, and every CI run uploads that directory with
the raw profiles. [`validation/fuzz-campaign.json`](validation/fuzz-campaign.json)
records each fuzz target and the time it ran.
