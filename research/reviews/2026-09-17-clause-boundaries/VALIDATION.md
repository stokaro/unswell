# Local validation and acceptance limits

Semantic freeze: `dbd11d798beabf030142aab5a11f120a8ef53e26`.
Source and label freezes preceded runtime edits. No semantic change followed
opening the confirmation output.

- Public blackbox regressions and compiled CLI checks pass in both profiles.
  They cover named/protected function heads, opaque operand counterexamples,
  relative restrictions, coordinated fragments, goal prefixes, conditions,
  Unicode, CRLF and Markdown source spans.
- The root ordinary suite passed its runtime tests. Ten CLI feature snapshots
  needed new identities: exactly 26 `ruleset_hash` fields changed, with every
  other parsed value unchanged. All ten affected scenarios passed on rerun.
- Ordinary tests for the consumer, goanalysis, MCP, annotation and dependency
  modules passed with `go test -count=1`. Strict Go lint passed for all modules.
- Repository policy and its negative tests, module tidiness, Bash lint and
  negative probes, mirror/SBOM checks, generated catalog and report schema passed.
- The resource self-test scanned 2,049 words in three documents: 0.384 seconds
  cold and 0.401 seconds warm on this host. Exceeding a limit exited 2. These
  observations do not qualify 100,000 words on a 2-vCPU host. The research-cost
  harness rejected a failed stage.
- All six Linux/macOS/Windows architecture binaries reproduced byte for byte
  across two local Go 1.27.1 builds with `CGO_ENABLED=0`. This is one host and
  one toolchain, not cross-toolchain reproducibility.
- All 17 sets completed in technical and strict with source, policy and tool
  identities validated. The existing #312 abstention remains explicit. No new
  abstention, skipped rule or operational failure was accepted.
- Fifteen evidence tests pass, including missing findings, invented credit,
  source drift, false overlap, confirmation losses, additional repairs and
  restored detections. A removed false positive cannot count as recall gain.
- CLI/MCP dogfood matched 1,082 documents and 887 findings with gate PASS.
  Failure, rewrite and malformed-input probes passed. This closing record was
  added afterward and checked separately.

`make check` stopped on the ten stale feature snapshots. The affected scenarios
were rerun, and the remaining targets completed separately. Race, active fuzzing
and coverage remain deferred to #123. No Docker resources were created.

Software checks do not qualify this candidate: new-page recall fell from 4/56
to 2/56, and Ptah remains 0/20. Both lost detections and all 54 misses are retained.

Remote CI and exact-tree artifacts for this branch are not accepted. At the time
of this record, parent PR #324's CI run 35278009093 passed native and Quality
jobs but failed both container jobs during MCP self-check. The AMD64 log reports
an error in batch 25 without retaining its structured MCP error evidence. This
local CLI/MCP pass does not resolve that container failure. No merge, release
or playground deployment was performed.
