# Local validation

Runtime freeze: `241c65bb77b23551064bd7e45c2ea57de3434bf3`.
The source-only selection and annotation freeze preceded runtime changes and
confirmation diagnostics. No semantic changes followed the confirmation result.

- All root ordinary tests passed except ten expected feature identity snapshots.
  Updated exactly 26 `ruleset_hash` fields in those ten files after checking
  that every other parsed value was unchanged. All ten affected CLI scenarios
  passed on rerun. Public API, compiled CLI, source mapping, controls, resource,
  configuration and cancellation tests passed in the root suite.
- Ordinary tests for every other module passed with `go test -count=1`.
- Strict Go lint passed for all modules. Bash checks covered 35 scripts and
  rejected syntax, quoting and formatting violations in negative probes.
- Repository policy, negative policy tests, module tidiness, generated catalog
  and report schema checks passed. Mirror and SBOM negative probes passed.
- The resource self-test analyzed 2,049 prose words across three documents:
  0.366 seconds cold, 0.610 seconds warm on this macOS host. An exceeded limit
  exited 2. This is not the 100,000-word or 2-vCPU qualification. The research
  cost harness rejected a failed stage.
- All six Linux/macOS/Windows architecture binaries reproduced byte for byte
  across two local builds with the same Go 1.27.1 toolchain and `CGO_ENABLED=0`.
  This does not establish cross-toolchain reproducibility.
- All 16 sets completed in technical and strict with source identities checked.
  The known exposed-instruction c05 budget abstention remains explicit; no new
  abstention, operational error or skipped rule was accepted.
- Evidence checks cover drift, missing findings, unsupported credit, additional
  repairs, lost detections, the confirmation regression and abstention handling.
- CLI/MCP dogfood matched 1,076 documents and 878 findings with gate PASS.
  Failure, rewrite and malformed-input probes passed. This closing validation
  record was checked separately after that replay.

`make check` stopped at the stale feature identities. I reran the ten affected
cases and completed the remaining targets separately. Race, active fuzzing and coverage remain
deferred to #123. Passing software checks does not qualify this experimental
candidate: confirmation recall fell from 2/44 to 1/44, and false positives remain.
Remote CI and exact-tree artifact acceptance are not verified. No merge, release
or playground deployment was performed.
