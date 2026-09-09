# ADR 0026: Bind the rule baseline to complete annotated blocks

Status: accepted for implementation; comparison and qualification remain open.

## Decision

The rule-activation baseline uses the existing engine's raw block activations.
It must join a label to the complete extracted text that those values describe.
Counted token spans and a bounding interval cannot prove that relationship.

Add an optional `FeatureUnit.Binding` to the existing result. Its versioned
`mapped-block-v1` contract records the SHA256 of the complete mapped block text
and every mapped source segment, including punctuation and whitespace. Separate
trimmed fields record the result of removing outer whitespace with the same
shared boundary function used by corpus preparation. The original fields still
describe the actual rule input, including the leading space in a Go comment. The
existing context hash still identifies grammar scopes. Source, policy, NLP,
preprocessing, and ruleset identities remain on the enclosing feature source.
No source prose is serialized. The engine captures this data during its existing
collection pass, without extraction, NLP, or rule reevaluation.

The binding describes the original block, including any protected separators.
It does not split protected pieces or rename a block as a prepared paragraph.
A consumer must reproduce source bytes and compare target and context hashes
and complete segments with the whitespace-trimmed whole block. No other
normalization is allowed. Equal parent IDs, equal tokens, or containment inside a
block are insufficient. A target covering only part of a block has no matching
block activation, even when a finding appears inside that target.

Saved reports without the optional binding retain their original semantics.
Readers validate present bindings and reject malformed contracts, hashes, and
source maps. Consumers requiring exact target binding reject missing identity;
they do not infer it from counted tokens. This is an additive report/API change;
strict older readers require an update for the new field.

## Training integration

The research workflow will select block activations explicitly and reuse corpus
verification, annotation matching, partition selection, permissions, the existing
Go optimizer, and separate calibration. It will retain the supplied rule policy
and actual ruleset identity. Missing target matches and inapplicable rules stay
absent with reasons; a missing value is never zero.

Baseline comparisons must declare the target population and context. Document
rules may use surrounding source content; their values must not be described as
isolated-sentence measurements. The prepared-feature model and rule baseline
must compare the same annotated targets and report coverage as well as quality.
This work does not permit inherited labels, synthetic human judgments, product
probabilities, or changes to default CI policy.

Original and trimmed binding segments count against the collection resource budget.
Tests cover punctuation, multiple source segments, protected boundaries, saved
report compatibility, ownership, limits, and exact versus partial target matches.
The baseline remains unqualified until #23 and #57 have the required corpus and
comparison evidence.
