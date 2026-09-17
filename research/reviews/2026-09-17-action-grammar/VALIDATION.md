# Local validation

Measured runtime: `0e0941ae26dcfa89a4b5d92fcb3080cf889fc596`.
The clean semantic-freeze checkout supplied the after binary. Later edits add
measurement records, reporting checks and expected catalog identities. They do
not change the matcher or frozen source-only labels.

## Product checks

Public blackbox API and compiled CLI tests passed for coordinated actions,
relative antecedents, capability and assurance layers, purpose/steps frames,
related methods and revisions. Controls cover another actor, uncertain need,
permissions, negation, protected vocabulary, ordinary capabilities and structural
boundaries. Existing source-coordinate, configuration, cancellation and budget
checks also passed.

`make check` used `CGO_ENABLED=0`, `GOMAXPROCS=4`, `GOFLAGS=-p=2`
and isolated Go and linter caches. Root tests passed except one old rule-version
assertion and ten feature golden scenarios. Updating the assertion from 6 to 7
and exactly 26 `ruleset_hash` fields resolved those failures. No finding, feature
value, source location or policy expectation changed in those goldens. All 11
affected tests passed on rerun; unrelated passing tests were not repeated.

Ordinary tests also passed in the consumer, goanalysis, MCP, annotation and
dependency research modules. Go lint, Bash lint and negative probes, repository
policy, mirror/SBOM self-tests, schema and generated catalog checks passed.
Race detection, active fuzzing and coverage remain deferred under #123.

The sandbox prevented resource reporting by `/usr/bin/time`. The remaining
resource and reproducibility targets ran outside that restriction. The resource
harness processed 2049 words in three documents in 0.336 seconds cold and 0.333
seconds warm; an exceeded limit exited 2. Its negative checks and the research
cost checks passed. This small harness check is not the 100,000-word acceptance
benchmark. All six Linux/macOS/Windows amd64/arm64 binaries reproduced byte for
byte in two builds on this host with the same toolchain.

## Evidence validation

Fourteen evidence tests pass. They reject missing or stale reviews, incidental
length credit, control credit, reuse of a source, invalid overlap evidence,
hidden partial exposure, and post-diagnostic recall credit. They retain partial
coverage separately and preserve the existing #312 budget abstention.

The renderer validates all 15 sets in both profiles and emits observed counts,
reviewed changes and every remaining confirmation miss. Older restricted review
scopes stay separate. All changed findings and all retained confirmation findings
have a disposition. No semantic adjustment followed confirmation output.

## Self-check

An initial CLI/MCP run agreed on 1066 documents and 873 findings, with a passing
gate and successful failure, rewrite and malformed-input probes. Its feedback
led to simpler active sentences in the report and a split explanation of the
two false positives. Frozen annotations and raw judgments remain unchanged.
The next staged check passed on 1067 documents and 867 findings. MCP matched
the complete CLI evidence, and its failure, rewrite and malformed-input probes
passed. This final status annotation then received a focused scan; it does not
change product code or measured research inputs.

These records describe local checks. They do not establish terminal remote CI,
merged-commit acceptance or deployment to the playground.
