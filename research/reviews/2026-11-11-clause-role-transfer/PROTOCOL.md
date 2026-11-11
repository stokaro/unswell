# Complete-unit clause-role transfer for #349

Status: source selection frozen before retrieving or reviewing the selected full pages. This is a cue-enriched, page-disjoint development challenge for a new source-role mechanism, not a prevalence sample or prospective #349 confirmation. The actual source-only reviewer is the root Codex assistant under ADR 0041, not a human annotator. No hosted or local model inference is planned in this stage.

## Hypothesis

The earlier line-level source-role selector admitted mixed technical sentences as empty narration. A useful replacement must treat the complete source unit and neighboring section as recognition context, separate a removable authorial clause from material propositions, and target only the former. A sentence with a cue may still convey a fact, scope, caveat, attribution, reader step, link, or table interpretation. Mere presence of a cue or lack of nearby lexical overlap is insufficient evidence for a finding. A safe suggested edit must preserve every material proposition, condition, quantity, identifier, version, link, and code boundary.

## Frozen selection

`selection.json` pins the base main commit, six raw GitHub search snapshots by SHA-256, four selected query strings and quotas, the two prior transfer inventories, 20 distinct repository/path/commit/blob tuples, and each search rank. The raw search JSON bytes are stored as deterministic gzip files under `search/`; two exploratory searches were retained for transparency but were not used in selection. The `hope-guide` query overlapped the retired selector's sources, and `guide-explore` mostly returned style guides or duplicate repository files.

For each selected query in the listed order, take the first ranked result whose repository is not in either pinned prior transfer inventory or already selected, whose URL contains a 40-hex commit, and whose path does not contain one of the frozen exclusion fragments. The quotas are 6 `guide-help`, 6 `in-guide`, 4 `worth-noting`, and 4 `following-sections`. This rule was applied before full selected-page retrieval. Do not replace a selected page after inspecting its role or engine output.

Fetch each page at the pinned commit. Verify the Git blob SHA-1 against the search result, record its original byte length and SHA-256, and retain the unchanged UTF-8 bytes locally. If a pinned page cannot be fetched, is not valid UTF-8, or exceeds 120 KiB, mark it unavailable in the 20-page denominator. Do not silently substitute another page. Never interpret a Markdown code span, table cell, or quoted example as a deletion target without source-function review.

## Source-only review before any candidate run

Read each complete available page, then annotate the first exact case-insensitive occurrence of its selected query phrase in eligible prose, if present. A missing exact phrase or an occurrence only in protected content is an ineligible target, not a positive or control. For an eligible target, record original byte span, surrounding unit, section role, and one of:

- `removable`: empty narration or unsupported evaluation can be removed or replaced without losing a material proposition;
- `mixed`: a removable authorial clause is attached to a material proposition; identify the smaller removable span and the content that must be kept;
- `retain`: the target performs a useful source-specific function or is not an editorial burden;
- `uncertain`: source evidence does not support a safe role judgment.

Record a proposed minimal edit only when its retained material content can be checked against the page. A mixed sentence is not permission to delete the whole sentence. Freeze the page hashes and all source-only labels before inspecting the current engine's output or developing the candidate. Keep uncertain cases in the delivered-finding denominator and report them separately.

## Screen and next gate

The challenge is adequately powered only if it yields at least five `removable` or `mixed` opportunities from at least three repositories and at least 15 `retain` controls from at least eight repositories. If it does not, report an underpowered result and do not claim a pass. The unchanged current-main CLI must complete all available pages; operational errors and ineligible pages remain visible.

Freeze a new non-LLM candidate and its code hash before running it against this set. A directional pass requires at least 80% FULL coverage of the source-only opportunities, at least 85% accepted actual delivered diagnoses, zero `retain` control hits, no edit that drops a material proposition, and complete processing of all available pages. A FULL hit must identify the actual removable burden and source bytes; overlap or a different critique does not count. Report abstention, invalid-page coverage, and edit safety separately. Do not tune a failed candidate on these exposed labels and call a rerun transfer.

Even a pass here is only a small cue-enriched transfer. Before #349 qualification, replay the complete unchanged 36-page, 123-event development cohorts and their actual combined finding stream, meet each 80% FULL and 85% accepted-delivery gate, resolve suggestion safety, then run the separately reserved prospective confirmation once under its declared role. Preserve all old denominators and negative results.
