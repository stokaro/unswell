# ADR 0029: Isolate LLMDet numerical compatibility

Status: accepted for research implementation; detector adoption remains open.

## Context

Issue #51 requires resource review and reference parity before proposing an
offline LLMDet pack. Its published Python code combines eleven proxy features
with a nine-class LightGBM model. The table archive is several gigabytes, and
component terms require review independent of the repository's MIT license.

## Decision

Use the existing `research/annotation` Go module, which also contains training
and evaluation tools, for an experimental `llmdet` numerical package and a
developer probe. Reuse its bounded JSON validation. Keep the product engine,
extraction, NLP, configuration, reports, and default gate unchanged.

Implement the pinned proxy calculation over explicit token IDs and probability
rows. Preserve its starting position, context precedence, log base, skipped
terms, and denominator for comparison. Return coverage and an absent usable
score when no position contributes a finite likelihood. The reference's raw
number remains separately labeled; compatibility does not validate its meaning.

Support an explicit data-only ensemble format with finite numeric splits and
leaves. Validate shape, dimensions, resource limits, connectivity, and depth
before inference. Reject missing or nonfinite features and unsupported tree
semantics. Sum leaves in original order and retain raw margins separately from
the multiclass softmax response. Neither output is calibrated for Unswell.

The probe consumes local numerical artifacts. It does not tokenize prose, scan
repositories, download resources, or make policy decisions. Reference scripts
use a separate pinned Python environment; ordinary Go builds and tests never
invoke it. Upstream notices accompany adapted calculation code.

## Evidence and limits

Keep authored numerical controls distinct from human-labeled prose. Reference
outputs over controlled rows check formulas. Outputs from the acquired published
classifier check tree execution. They do not establish full LLMDet reproduction,
tokenizer compatibility, dictionary coverage on prose, or editorial usefulness.

Do not distribute model tables or classifier weights under an assumed license.
Record observed metadata, bytes actually acquired, hashes, and unresolved terms
separately. Full detector adoption still requires permitted assets, compatible
tokenization, resource measurements, and the shared held-out evaluation protocol.

Promotion into product inference needs a separate API decision. Reuse this
numerical implementation when qualified; do not create another extraction or
scoring path beside the existing engine.
