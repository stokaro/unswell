# ADR 0042: preserve editorial claims independently of revisions

Status: accepted for experimental research accounting; not detector qualification.

An earlier source-bound prototype accepted overlap between a new finding and its
parent as structural lineage. That relationship cannot establish that a revision
addresses every criticism in the parent. A grammar edit and a rhetorical criticism
can concern the same paragraph. A retained antecedent can also support a finding
about later repetition without being a revision target. See
[#367](https://github.com/stokaro/unswell/issues/367).

Keep the accounting tool in the existing `research/annotation` consumer module,
under `internal/claimreview` with the `cmd/reviewclaims` application boundary.
It uses `extract.Parse` and `document.Document`; it adds no parser, model client,
core result field, or reporter computation. Source bytes, existing block IDs,
protected exclusions, and exact original ranges govern binding.

The `unswell-editorial-claims-v1` inventory has separately declared criticisms,
their original targets, support-only context, and run/candidate/claim provenance.
The caller declares claim boundaries. The tool does not infer semantic atomicity
or turn a compound legacy diagnostic into independently accepted claims.
IDs hash the contract version, source identity, original specification, and exact
references. Later stage decisions never overwrite that specification.

Every stage must account for every original claim as retained, resolved, rejected,
or uncertain. Missing, repeated, or invented claims fail validation. Resolution
requires a concrete edit within that claim's original target and separate
caller-supplied review of the exact edit digest for that exact claim. Approval of
a grammar correction cannot approve a rhetorical claim just because their source
ranges overlap. Support-only context cannot become the edit target.

Deduplication requires an explicit same-defect member set and a separate review
of that relationship. All original claims and dispositions survive; only display
representatives change. No overlap, generic similarity, category match, or edit
quality score infers this relationship.

Approvals are trusted application input, distinct from untrusted stage output.
The tool records reviewer declarations and reasons but cannot authenticate a
reviewer or establish the semantic truth of their approval. Assistant reviews
remain explicitly identified under [ADR 0041](0041-assistant-review-acceptance.md).
Structural validity never upgrades existing rejected or uncertain research labels.

The optional command prepares source-bound IDs or replays supplied stages. Its
strict local schema rejects unknown fields, case aliases, duplicate keys, invalid
Unicode, unsupported versions, and incomplete inventories. Uncertain mandatory
accounting emits `complete: false` and exits 2. Preparation without stages is
explicitly incomplete and exits 0. Cancellation exits 130. A writer error fails.

The library receives an existing extracted document; the developer command
extracts the explicitly supplied source once. Loading and replay call no external
process, network, model, configuration endpoint, or current-directory discovery.
The command excludes quotes, replacements, and review explanations unless
`include_raw` is explicitly true. Its output is research evidence, not the
product `RunResult`; default CLI/MCP gate and report contracts stay unchanged.

The old 804-candidate pool is a mechanical migration test, not new annotation.
Undeclared support roles or independent claim splits are not invented. Unsupported
old ranges remain explicit invalid records and stay in the original denominator.
The 123-event development objective and prospective confirmation role remain
unchanged. This contract cannot close the contextual-recall qualification issue.
