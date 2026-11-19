# Check committed changes

Use `unswell check --changed-from main` to check the committed work on a branch.
Unswell resolves the merge base of the selected reference and HEAD, analyzes both
versions, and applies the gate to changed units (the manifest records both commit
IDs). This mode requires Git and a locally available, unambiguous merge base.

```sh
unswell check --changed-from main docs/ src/
unswell check --changed-from main --report json:result.json --report html:result.html
```

The reference may have advanced after the branch diverged; comparison still starts
at the merge base, rather than treating the reference's current tree as the base.
Git commands receive arguments directly, do not fetch missing objects or run
checkout filters, have output limits, and honor the scan timeout.

## What affects the gate

Unswell analyzes complete selected documents to retain repetition context. A changed
paragraph selects its sentences (including findings on lines that were not edited).
Removing a blank line can join paragraphs and select the resulting paragraph.
Changes to multi-location evidence can select older occurrences elsewhere in the
document. These remain part of the comparison contract: protected code, negation,
numbers, and structural context.

Some edits can leave existing units unchanged: unrelated line movement and equivalent
Markdown emphasis. If the document changed and duplicate identities cannot be matched
unambiguously, those units remain selected. An exact match of the complete source
can establish that its duplicates are unchanged. A renamed file is checked in full
under its new path; the old path is recorded as deleted. Changed source permissions
select the entire affected document.

All current findings and raw and effective scores remain in the result. Findings
and assessments gain `change_state` (`changed`, `unchanged`, or `ambiguous`) and a
structural `change_fingerprint`. `changes` records source hashes, document status,
full-file selection reasons, completion, and compared and selected unit counts.
(The counts include sentence and paragraph assessments separately.) `gate.unchanged`
retains reasons excluded from the changed-unit gate.

Baseline acceptance remains independent. Unchanged selection does not create
accepted debt, and baseline capture does not accept `--changed-from`. A check may
combine `--changed-from` with `--baseline accepted.json --gate-mode new`;
compatibility checks still apply.

## Source and policy verification

Selected current files must be regular files whose bytes and index entries match
HEAD. Verification uses Git object identities; it rereads the selected inputs after
analysis. It detects content hidden by `assume-unchanged` or `skip-worktree` and
rejects symlinks. A committed deletion must also be absent from two places: the
index and the worktree. Explicit untracked paths are errors. Directory selection
uses committed paths and detects selected staged additions; unrelated untracked
drafts are outside that selection.

The Git manifest's `clean` field describes selected sources and policy inputs at
verification time, not the entire worktree. Checkout transformations (such as
LF-to-CRLF conversion) are rejected when the bytes differ from the committed blob.
Use a checkout with byte-preserving attributes or run an ordinary draft check.
This restriction keeps report positions tied to the verified source.

Configuration, inherited policies, dictionaries, explicitly loaded rulesets, and
the selected baseline must be committed, clean, and identical between the merge
base and HEAD. Policy inputs outside the project root are unsupported here. Changes
to discovered configuration or membership of a ruleset directory also fail; use
[trusted policy](trusted-policy.md) with `--policy-from-base` to check those
changes under merge-base policy. An ordinary changed-unit check alone does not
secure an untrusted pull request's workflow or command-line options.

These return an operational error: missing objects, incomplete extraction in either
version, incompatible baselines, dirty inputs, or a moving HEAD. An empty current
scan (including a selection containing only deletions) requires `--allow-empty`.
`--no-gate` does not hide those errors. Ordinary `unswell check` continues to check
uncommitted drafts (without Git verification).

## Library and reports

`Engine.AnalyzeChanged(ctx, before, after)` accepts two complete sets of source
bytes under the engine's fixed policy; callers own revision discovery and verification.
The method performs no filesystem access, process execution, or network access.
It shares extraction, NLP, structural identities, and scoring with ordinary scans
and baseline comparison; one engine supports concurrent calls.

The MCP tools continue to analyze supplied drafts with the same engine; they do
not discover Git revisions or accept a caller's claim that a local checkout is
clean. JSON, SARIF, text, HTML, and Markdown retain selection evidence; saved
reports can be rendered without Git, source files, or repeated analysis.

The [committed e2e workflow](../e2e/changes_test.go) uses annotated source fixtures
and a golden report; it checks line movement, joined paragraphs, semantic changes,
and a dirty-source rejection through the built CLI. Library tests also cover
repetition dependencies, source permissions, ambiguity, and baseline composition.
