# ADR 0017: Collect shared measurements through RunResult

Status: accepted for implementation.

## Decision

An explicit `Options.Features` set selects implemented block feature IDs from the
existing `feature.Catalog`. Unknown IDs and duplicates are errors. The engine
canonicalizes order and validates every requested NLP capability before analysis.
An advertised but unrequested capability is not evidence that it was computed.
The request does not add a model, threshold, or gate rule.

The engine creates its existing shared feature set once after extraction and NLP.
Rules and result collection read that same set. Collection never extracts prose
or calls NLP again. Per-file engines preserve the request while retaining their
effective source/extraction/vocabulary policy. Returned slices and numeric values
belong to the caller; concurrent calls share no mutable result buffers.

`RunResult.Features` is absent when collection was not requested. When present it
records the collection version, block contract, canonical IDs, and source records
with source/policy/vocabulary hashes, NLP identity, actual capabilities, and
preprocessing identity. Each block records its kind, context hash, original span,
counted source segments, input hash, and the requested values. Unsupported block
kinds receive absent values with `unsupported_unit`; excluded prose is never
restored. Empty ratios keep their existing absence reasons instead of zero.

Grammar context strings can contain heading prose. The collection stores only
the SHA-256 of their JSON representation, including order and boundaries; it never
copies those strings into a report. The full measured input hash still binds the
actual context used by the feature set.

The enclosing RunResult completion state governs the collection. Values obtained
before a later operational failure remain partial evidence, not a complete model
input. Training and inference consumers must reject incomplete runs. Policy
failure is independent of collection: a complete run may contain findings and
valid measurements. Baseline, changed-unit selection, suppressions, and trusted
policy retain the measurements from the full source/context actually analyzed.

The CLI requests IDs with repeated `--feature` flags. MCP accepts the same flags
at startup; checks use the fixed engine and describe exposes the selected IDs.
Requests do not become client-controlled policy overrides. All report formats
consume the same collection. JSON and SARIF preserve it, while text, Markdown,
and HTML display numeric values or absence reasons. Saved reports can be rendered
without a model or source files. No normalized words or template keys are stored.

## Compatibility and remaining work

The field is additive and omitted by default. Existing saved reports remain
readable; strict older readers must be updated for explicitly requested feature
collections. Existing feature identities and the scoring/baseline contract do not
change. The collection has its own version; schema and API snapshots document it.

This implements block collection within #56. Raw rule activations still need a
separate post-evaluation contract covering applicability, disabled/skipped rules,
incomplete evaluation, and evidence aggregation. No missing activation may be
invented as zero. Repetition/model-specific vectors and optional LLMDet contracts
must use this pipeline when implemented. No calibrated probability or qualified
model is claimed by collecting descriptive measurements.
