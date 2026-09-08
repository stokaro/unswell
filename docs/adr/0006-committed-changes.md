# ADR 0006: committed changes with structural context

Status: accepted for the fixed-policy committed comparison in issue #14.

## Decision

The CLI resolves `--changed-from REF` to a merge base and HEAD, loads their source
versions, and verifies selected working-tree and index content against HEAD. Git
commands use bounded argv-based execution. The engine receives source bytes and
performs no Git or filesystem operations.

Compare complete before/after documents under one fixed engine policy. Reuse the
structural content and evidence calculations used by baselines, without treating
the previous commit as accepted debt. Rules inspect the entire current document;
findings and scores remain visible even when their units are unchanged. A changed
paragraph selects all of its sentences. Changes to repetition evidence select
affected older units too. Source-permission changes select the complete file.

Equal fingerprints identify unchanged findings and scored units. Repeated identical
identities in a modified document are ambiguous and remain selected. Exact equality
of a complete source can establish that its duplicates are unchanged. New paths and
changed input formats select the complete file.
Renames retain source provenance while the new logical path is checked as new.
Deleting content is compared through the remaining complete document, so joined
paragraphs and changed context are included.

Selection and baseline acceptance remain separate result fields. An unchanged gate
reason is retained for audit. Explicit baseline comparison still checks compatibility
and preserves its own accepted-debt state. Reporters consume the finished result.

Unavailable references, dirty selected sources, incomplete analysis, and incompatible
required baselines are operational errors. A content-only comparison cannot silently
use changed configuration, dictionaries, rulesets, or baseline permissions. Trusted
base-policy loading and policy-change handling are the next issue, #15; until then,
the committed adapter rejects changed policy resources and points to a full check.
Ordinary draft checks continue to accept uncommitted source content.

## Acceptance

Cover unrelated line movement, changes to a complete paragraph, deleted separators,
new and changed repetition clusters, ambiguous duplicates, source permissions,
baseline composition, and retained raw scores. Git process fixtures must exercise
merge-base semantics, staged and unstaged changes, missing references, path safety,
renames, deletions, and content changes hidden by Git status optimizations. Saved
reports must retain selection and revision evidence without rereading source.

Root library tests, CLI Git fixtures, report round trips, and the built-CLI
`TestCommittedWorkflow` cover these contracts. Policy changes and untrusted CI
workflows require the separate trusted-policy work; this ADR does not qualify them.
