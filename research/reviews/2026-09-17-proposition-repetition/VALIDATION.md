# Local validation

Runtime measured: `e18f3dcc29bdc46cd5812979238d9e1e0aaa04d3`.
The measured binary came from a clean detached checkout. Later changes are
evidence tooling, reviewed reports and expected-output updates, not runtime.

## Product checks

`make check` was run with `CGO_ENABLED=0`, `GOMAXPROCS=4`, `GOFLAGS=-p=2`
and isolated Go and linter caches. The root test run passed every package except
one expected-output scenario: `TestCLI/local_repetition` still contained the old
explanatory-restart message. Both JSON and text goldens were updated to the new
message. The exact failed scenario then passed; already passed root packages
were not run again solely because an expected message changed.

The remaining ordinary tests passed in the consumer, goanalysis, MCP, annotation
and dependency research modules. Policy, module tidy, strict Go lint, Bash lint
and policy self-tests, image-mirror and SBOM self-tests, schema validation and
CLI/MCP dogfood checks passed. The shared CLI/MCP result covered 1043 documents
before the final evidence pages were staged. A final staged self-check is
recorded below, so untracked evidence pages cannot escape the claimed scope.

The sandbox prevented `/usr/bin/time` resource reporting. Only the residual
`check-performance`, `check-research-cost` and `check-reproducible` targets were
rerun outside that restriction, and all passed. The resource harness processed
2049 prose words in three documents in 0.377 seconds cold and 0.371 seconds
warm, and rejected an exceeded limit with exit 2. This is the harness self-test,
not the 100,000-word performance qualification. All six Linux/macOS/Windows
amd64/arm64 binaries reproduced byte-for-byte in the same-host two-build check.

Focused public API tests and `TestRestrictionReformulationRevision` passed.
The budget regression failed on the prototype and passed after the resource
fix. Ten feature goldens changed only their 26 ruleset identity fields; the
runtime rule-version changes explain those identities. Catalog generation
and its drift check passed.

## Research evidence checks

`tools/test_evidence.py` passes 14 positive and negative checks. They cover
missing finding reviews, advice drift, cross-page credit, incidental sentence-
length overlap, attempted credit for controls, duplicate source identities,
source-quote drift, partial-to-full circular-reason coverage, post-diagnostic
additional findings, the known budget abstention, prototype semantic drift and
retention of the original budget failure. `tools/render.py` validates all 13
sets in both profiles and regenerates the report, change list and miss ledger.

The full confirmation result is 0/46 before and after. All 46 misses remain in
the ledger. No absent or partial finding is counted as full detection. Both
prototype and corrected reports are retained; their findings, documents, gate
results and policy manifests match. Runtime commit identities differ, and the
manifest records the two removed abstentions consistently with their detail records.

Race detection, active fuzzing and coverage remain deferred under #123. No
remote CI, merged-commit, release or playground acceptance is claimed here.

## Final staged self-check

After staging the evidence pages, `make policy dogfood-mcp` passed on 1047
documents. The complete CLI run emitted 849 findings and passed its configured
gate; MCP matched the full CLI evidence, including failure, rewrite and malformed-
input probes. Existing findings remain visible. The final evidence validator
and all 14 evidence tests also passed. This is local proof only.
