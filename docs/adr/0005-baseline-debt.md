# ADR 0005: explicit acceptance of existing editorial debt

Status: accepted.

## Problem

Existing finding fingerprints combine prose and evidence with the source name.
They can collide when the same fragment appears in different structural contexts.
Derived score findings also depend on document hashes and numeric unit IDs, which
change after unrelated edits. Neither identity is sufficient for accepting debt.

## Decision

Keep baseline storage and comparison in a public, pure Go `baseline` package.
It accepts completed in-memory snapshots from the shared engine. It performs no
extraction, scoring, filesystem access, process execution, or network access.
The CLI owns explicit create, update, and check commands and atomic file writes.
Checking never changes a baseline.

Version baseline files and structural fingerprints independently. An identity
includes a project-relative path, kind, rule behavior version where applicable,
structural context hash, normalized content hash, and evidence hash. Use typed,
length-delimited serialization before hashing. Line numbers, report messages, and
numeric unit IDs are not persistent identities.

Extraction supplies grammar-derived heading, declaration, and keyed-value context.
Canonical content preserves negation, numbers, identifiers, and protected source
boundaries. Reformatting and unrelated line movement can retain identity. A changed
paragraph is analyzed in full. Unit identity also includes scoring evidence, so a
new document-level repetition can invalidate acceptance of an unchanged paragraph.
Indistinguishable candidates are ambiguous; they cannot silently inherit debt.

Compatibility records identify effective analysis policy, rule behavior, NLP,
features, scoring, and calibration. Per-source records retain source hashes and
policy-sensitive suppression identities. Report presentation and gate selection
mode do not change analysis identity. Actual analysis-policy changes require an
explicit baseline update after complete analysis of the affected coverage.

Snapshots declare complete analysis of each included document. Partial snapshots
cannot create or update accepted debt. An ordinary comparison may cover a subset
of baseline paths: entries outside that coverage remain unobserved. Updates retain
unobserved debt rather than silently deleting it. Updating global compatibility
requires complete coverage of the previous baseline. Stale entries within observed
documents are reported and are removed only by an explicit update.

Baseline acceptance is separate from source suppression. Raw findings, raw and
effective scores, probability availability, and suppression records remain intact.
The `new` gate considers new findings and new or changed scored units. The `all`
gate retains ordinary behavior even when baseline state is displayed. Missing or
incompatible required baseline data is an operational error, including with
`--no-gate`.

Saved results carry comparison evidence for every reporter and MCP. Reporters do
not load baselines or recompute acceptance. No source prose is stored in baseline
files; structural and content hashes still require the same access controls as
other repository metadata.

## Acceptance

Verify exact and ambiguous matches, stale and unobserved debt, incomplete scans,
policy/model changes, semantic edits, moved lines, repeated evidence, and bounded
artifact parsing. Add annotated CLI fixtures for explicit creation, updates,
read-only checks, and negative gates, plus saved-report and MCP parity. Issue #13
remains open until integration, documentation, and merged-commit CI are complete.

Changed-unit Git selection and trusted base-branch policy remain the existing
follow-up issues #14 and #15. Baseline acceptance is not a substitute for either.
