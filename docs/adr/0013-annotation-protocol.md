# ADR 0013: independent editorial annotation

Status: accepted for the protocol and tools; corpus and model qualification remain open.

## Decision

Use the [versioned rubric](../editorial-annotation.md) and the separate
`research/annotation` Go module for annotation artifacts, blinded packets, and
agreement measurement. Preserve source metadata, origin evidence, original human
answers, and adjudication as distinct records. The tool cannot manufacture or
certify human participation.

The quality target has three labels: acceptable, needs revision, and uncertain.
Seven reason categories describe required wording changes. An uncertain answer
is not missing data, an origin label, or a negative example for every category.

Use nominal Krippendorff alpha because rounds may have more than two raters and
missing responses. Also report raw pair agreement, uncertain share, support, and
kind/role/category slices. Preserve undefined and negative results. Point estimates
do not establish sampling precision; corpus publication requires grouped intervals
under the preregistered evaluation protocol.

## Alternatives and consequences

A single human/AI label would confound origin and editorial quality. A classifier's
prediction would create circular ground truth. Replacing initial answers with
consensus would conceal disagreements. A two-rater-only coefficient would not
describe rounds with additional reviewers under one contract.

The separate module keeps research administration out of product inference and
the public engine API. It reuses source span and syntax contracts without
duplicating extraction. It participates in normal repository tests and lint.
The initial JSON format uses a crossed assignment within each round; more complex
assignment designs need explicit evolution of the protocol.

Packet fields are an allowlist. This reduces metadata leakage but cannot remove
origin clues inside prose or prove that reviewers stayed independent. Source
references and permission records require curator verification. The loader checks
their structure and declared scope, not the truth of those assertions.
Responses and adjudications carry the digest of the exact blinded packet. This
prevents accidental reuse across changed targets or context with the same unit ID.

Scripted tutorial responses are isolated from pilot/corpus records and cannot
count as human evidence. The next work is real pilot collection, corpus manifests,
leakage checks, and split assignment in #22. General feature contracts (#56),
comparative experiments (#57), Go training (#23), and calibration (#24) retain
their existing responsibilities. This ADR does not close the broader methodology
registry in #55 or make a revision probability available.
