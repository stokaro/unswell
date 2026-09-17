# Local validation

Runtime measured: `4bac934fd7085d5a53f15ad2a71a8f77bd493649`.
The CLI was built from the clean task checkout after committing runtime,
focused tests and frozen source-only annotations. Later edits contain evidence
tooling, reports and expected ruleset identities, not runtime changes.

## Product checks

Focused public API and compiled CLI regressions passed. Strict builtin-package
lint passed after extracting bounded parser helpers. Tests exercise narrated
prerequisites, same and distinct actors, optional capabilities, permission,
negative conditions, named operations, gerund and queried-value complements,
source coordinates, settings, cancellation and exhausted work budgets.

`make check` ran with `CGO_ENABLED=0`, `GOMAXPROCS=4`, `GOFLAGS=-p=2`
and isolated Go and linter caches. The root tests passed except ten e2e feature
report scenarios whose ruleset identity changed with instruction rule version 6.
The reviewed goldens change only 26 `ruleset_hash` fields; their parsed contents
otherwise match. Those exact ten scenarios passed after updating the identities.
Passed root tests were not repeated just to collect coverage.

Ordinary tests passed in the consumer, goanalysis, MCP, annotation and dependency
research modules. Strict Go lint, Bash lint with negative policy probes,
mirror/SBOM policy self-tests, schema validation and catalog drift checks passed.

The sandbox blocked resource reporting, so only the residual performance,
research-cost and reproducibility targets ran outside that restriction. All
passed. The resource harness processed 2049 prose words in three documents in
0.350 seconds cold and 0.354 seconds warm, and rejected a limit breach with
exit 2. This is a harness self-test, not the 100,000-word qualification. All six
Linux/macOS/Windows amd64/arm64 binaries reproduced byte-for-byte in the same-
host two-build check.

CLI/MCP self-checks matched on 1053 documents before all evidence was staged.
The next staged self-check caught duplicate review paragraphs and a discussed
modifier formatted as ordinary prose. The Markdown renderer now references the
already reviewed replacement for a same-location change, prints general advice
once, and quotes the discussed modifier as a literal. Exact frozen source bytes
and judgment JSON remain unchanged. Final staged results are recorded below.
This document does not establish remote CI, merged-commit or deployment evidence.
Race detection, active fuzzing and coverage remain deferred under #123.

## Research evidence checks

`tools/test_evidence.py` passes 14 checks, including rejection of missing reviews,
changed advice, cross-page credit, incidental length overlap, control credit,
reused source identities, source-quote drift and post-diagnostic recall credit.
It also preserves partial-event exclusion, the known #312 budget abstention,
two lost detections and the three removed capability controls.

`tools/render.py` validates all 14 complete-page sets in both profiles and
regenerates the report, diagnostic-change list and confirmation miss ledger.
All added, changed and removed findings are reviewed; all retained confirmation
findings are reviewed. Older limited repetition/framing scopes are retained.
There was no post-output matcher adjustment or new resource exception.

## Final staged self-check

After the rendering repairs, `make policy dogfood-mcp` passed on 1057 documents.
The complete CLI result contained 861 findings and passed its configured gate.
MCP matched the full CLI evidence; failure, rewrite and malformed-input probes
passed. Frozen-input validation and all 14 evidence tests also passed. Existing
nonblocking findings remain visible. These are local checks only.
