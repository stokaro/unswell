# ADR 0023: Collect prepared target measurements in the engine

Status: accepted for implementation; model qualification remains open.

## Decision

The engine optionally collects sentence, paragraph, and fragment measurements
through `nlp.PrepareUnits` and `feature.MeasureUnit`. The same saved result feeds
CLI reports and MCP. Extraction runs once. Existing block enrichment and rule
evaluation retain their inputs; preparation analyzes eligible pieces separately
under `eligible-piece-v1`, sharing each piece between its sentence and paragraph
targets. Block rule activations are not assigned to these targets.

`Options.PreparedFeatures` and `Options.PreparedKinds` are explicit sets. Both
must be present to enable collection. Unknown or duplicate entries and missing
provider capabilities fail construction. Requested capabilities are independent
of rule requirements. CLI and MCP startup expose repeated `--prepared-feature`
and `--prepared-kind` flags; an MCP client cannot change them.

`RunResult.PreparedFeatures` uses `unswell-prepared-feature-collection-v1` and is
absent by default. Each source records its actual NLP capabilities, full policy
hash, extraction policy hash, extraction switches, and preparation hash. These
identities have different meanings. The extraction policy hash covers its JSON
representation; the preparation hash also covers the preparation contract and
quote/structure switches. It is supplied as the feature preprocessing identity.
The source hash, complete target/context binding, and counted-token segments are
retained separately. Prose and normalized tokens are omitted.

The parent block ID is not a target ID: a paragraph and its sentences can share
it. A target is identified by source, kind, complete segments, and its binding.
Grammar context and surrounding prose have separate hashes. Corpus joins must
verify actual extraction options and complete bindings before attaching labels.
The older corpus pipeline identity is not an alias for this collection.

Per-source candidate limits charge saved values, segments, and measured token
visits. Existing file/block/token limits bound input preparation, with additional
preparation ceilings. Errors discard the affected source collection and make the
run incomplete even with `--no-gate`. Earlier complete source records may remain
as partial evidence; they are not a complete training input.

All five reporters consume the common result. Saved JSON validation checks
contracts, source identities, requested values, ranges, and completeness. Reports
can be rendered without the original files or NLP provider. Default report fields
and the existing block collection v1 retain their meanings. This additive opt-in
field requires updating strict older readers.

## Remaining work

This is descriptive collection for #56. Exact annotation joins, qualified data,
model artifacts, training commands, applicability, and calibrated probabilities
remain separate acceptance work. Collection supplies no quality or origin label.
