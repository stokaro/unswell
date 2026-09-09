# ADR 0024: Bind annotation decisions to reproduced measurements

Status: accepted for the research data path; model and corpus qualification remain open.

## Problem

Annotation decisions and prepared feature collections have different identities.
A round's unit ID alone cannot establish that a measurement describes the same
text, context, source segments, or source group. Joining on a bounding interval can
also cross protected code. A hash on a saved artifact establishes integrity of its
declared contents, not reproduction from the original source.

## Decision

Extend the existing research `corpus` package with `Join` and the existing developer
command with `corpus join`. The operation receives the candidate artifact, a
validated annotation round, exact source and notice bytes, and an explicit set of
prepared feature IDs. It performs these steps:

1. Reproduce the candidate artifact through `corpus.Verify`, including source
   bytes, mappings, extraction exclusions, pipeline, groups, and partitions.
2. Match every round unit to an expected candidate through
   `annotation.Round.MatchTargets`. IDs, text, context, kind, role, source metadata,
   extraction declarations, and rights must agree. Allowed uses form a set.
3. Measure the frozen corpus with the public engine and its prepared collection.
   Unit kinds and extraction policy come from the plan. The configuration uses
   `builtin:custom`, with no rule selection or threshold-based sample filtering.
4. Match each candidate to exactly one measurement using source path and hash,
   unit kind, text/context hashes, and every original source segment. Reject
   missing, extra, or ambiguous targets.
5. Export separate decisions, measurements, and bindings with reproduced source
   IDs, group IDs, partitions, and feature input hashes.

The round can cover a subset. Every corpus target stays in the output, including
unannotated targets. Missing and uncertain decisions retain their existing status;
the join does not select training examples or turn accepted baseline debt into a
negative label. Origin claims are independently curated and do not participate in
editorial target matching. Their administrative input remains bound by the round
hash; no origin metadata is supplied as a numerical feature.

The annotation profile hash identifies the human instructions. Engine policy,
vocabulary, extraction, preparation, NLP, and feature identities keep their own
meanings. Configuration rejects null overrides, so nil extraction fields are
omitted when constructing it. An unset global context set uses defaults;
per-language overrides require an explicit set. Empty sets disable extraction.
The complete target comparison checks the resulting behavior.

`unswell-corpus-feature-bindings-v1` is a separate research artifact. Its status is
`verified_targets_with_measured_features`, and `human_corpus` remains
`not_qualified`. Tutorial rounds retain `simulation`; pilot/corpus declarations
remain declarations. The output does not claim permission to train or publish.
An annotation-only permission stays annotation-only.

## Limits and consequences

The implementation uses the common engine, prepared feature registry, and existing
source loader. It does not import engine internals, invoke reference runtimes, or
fetch resources. Sources are measured one at a time. Existing source, preparation,
and candidate limits apply; cumulative measurements are limited to 10,000 targets
and half of the 128 MiB artifact budget. The final compact artifact must fit the
full budget. Errors and cancellation return no partial artifact.

The joined output omits prose and annotation rationales. It retains source paths,
opaque group identifiers, ranges, feature values, and hashes, so publication still
requires reviewing the permitted uses of the research data. It is not a blinded
annotation packet. Reformatting the original round changes its input identity.

Consumers must reproduce the join from the original inputs before trusting a
saved file; recomputing its digest cannot establish honest labels or provenance.
This version does not add a generic saved-join loader, a trained model, corpus
qualification, an inference probability, or a probability gate. Those requirements
remain under #21–#25 and #56.

## Evidence

Root e2e runs the compiled command over the frozen eight-source Ptah fixture and
checks all 378 bindings against the existing original-segment golden. A tutorial
round selects five units without supplying any responses; every decision remains
missing, and all 378 measurements remain present. Changing the round's repository
identity makes the real command fail with exit 2 and no output.

Package tests reject rehashed candidate/partition/pipeline changes, source and
notice corruption, changed annotation provenance, bounding-span substitutions,
changed context, duplicate/missing/extra targets, and invalid feature selections.
They preserve disabled language contexts, empty-source records, source ownership,
separate instruction/policy hashes, and cancellation.
