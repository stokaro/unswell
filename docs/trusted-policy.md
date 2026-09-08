# Check with trusted policy

Use `--policy-from-base` with `--changed-from REF` when a branch must satisfy the
policy that existed at its merge base. Unswell reads that policy from immutable Git
objects. A candidate configuration, dictionary, ruleset, baseline, or source
permission cannot silently weaken the check.

```sh
unswell check --changed-from main --policy-from-base docs/ src/
unswell check --changed-from main --policy-from-base --config policy.yaml --ruleset rules/ docs/
unswell baseline check --changed-from main --policy-from-base --baseline accepted.json docs/
```

The baseline example requires `gate.mode: new` in the trusted configuration to
accept matching debt. The flag applies to `check` and `baseline check`; it requires
committed source paths and cannot capture or update a baseline. Profile, gate-mode,
and outside-root overrides are rejected. The invocation still controls paths,
formats, resource flags, `--allow-empty`, and `--no-gate`; a required check must fix
those arguments in its trusted workflow.

## Resource comparison

Configuration discovery starts in the invocation directory and stops at the
project root, using entries from the merge-base tree. Inheritance and dictionaries
use the existing bounded configuration loader. Explicit configuration, ruleset,
and baseline paths refer to that tree and must stay inside the project root.
Missing or invalid trusted resources return an operational error.

The result compares each trusted resource with HEAD using SHA-256 content hashes.
It also compares configuration discovery and the immediate YAML members of explicit
ruleset directories. Any external resource difference selects all units in the
current source selection under the base policy. Trusted include/exclude settings
control directory discovery; explicit file arguments retain ordinary selection
behavior. This may surface old debt
that an ordinary changed-unit check would omit. Compatible base debt remains
accepted when the trusted gate uses mode `new`.

The audit covers the trusted resource graph and discovery/directory membership.
Changed referring files expose candidate graph changes; new candidate dependencies
are not loaded or compiled. New directory members are recorded through membership
hashes and are not executed or applied. Candidate syntax should be reviewed with
the ordinary `config validate` command. If a candidate policy file is also selected
as prose, its extraction must still succeed.

Changes to policy do not automatically fail clean prose. Maintainers must review
the policy that future checks will trust. Builtin rules and NLP/model identities
belong to the pinned checker binary; Unswell does not compile historical source or
download models. The alpha continues to require `calibration.model: none`.

Selected sources and observed policy resources must match HEAD and the index.
Unswell verifies them again after analysis, including committed deletions. A dirty
resource is an operational error even when it would be excluded from prose scanning
and even with `--no-gate`. Reports cannot overwrite observed inputs.

## Source permissions

The existing directive syntax is unchanged. A permission can suppress current
findings only if its kind, reason, rule IDs, and structural target match the base
unambiguously. Moving a block within the same context may retain that permission.
Changing its content, scope, or structural context can make it untrusted. Duplicate
identities are ambiguous unless the complete source is unchanged. An existing
file-wide permission retains its explicit file scope when that mode is enabled.

New, expanded, retargeted, or ambiguous permissions remain visible in the audit but
do not suppress findings. A finding that requires multiple permissions needs every
contributing permission to be trusted. Changed bindings select the whole file.
Malformed and unused directives retain their ordinary error behavior.

## Results and library use

`policy_comparison` has version `unswell-policy-comparison-v1`, completion,
`full_scan`, and sorted resource changes with paths, kinds, and before/after hashes.
An absent side has no hash. Directory hashes identify sorted, NUL-separated member
names. Source permission hashes identify structural bindings; neither contains
raw policy text. Permission records add `trust_state` (`trusted`, `untrusted`, or
`ambiguous`) and `trust_fingerprint`. Their raw match audit remains available even
when the permission cannot affect effective scores.

JSON, SARIF, text, HTML, and Markdown consume the same completed result. Saved
reports render without Git or reanalysis. SARIF marks only effective source
suppressions as accepted. Source snippets remain excluded by default. Exit codes
remain 0 for a complete pass, 1 for a complete policy failure, 2 for an operational
error or incomplete check, and 130 for cancellation.

Library callers use `Engine.AnalyzeChangedWithOptions` with
`ChangeOptions{TrustedPolicy: true}` and optional `PolicyChanges`. They supply the
trusted engine configuration, baseline, before/after bytes, and resource identities.
The engine owns permission matching, scoring, selection, and the gate; it performs
no Git, filesystem, or network operations. Ordinary `AnalyzeChanged` and draft
analysis retain their behavior. MCP draft tools continue to use the same engine
without claiming that caller-supplied text came from a verified Git revision.

## CI trust boundary

A commit hash establishes identity, not approval. Use a checker built from a
trusted revision, an independently controlled invocation, and a fetched base ref
that the pull request cannot replace. Pin those inputs in the required CI check.
Repository protections and maintainer review must cover changes to the checker,
workflow, configuration, rules, baseline, and source permissions. If a pull request
can replace its workflow or checker, it can replace the check itself; this flag
cannot prevent that. Do not run candidate executables with privileged credentials.

The [trusted workflow test](../e2e/trusted_test.go) runs the built CLI against a real
Git fixture. Its [golden](../e2e/trusteddata/workflow.golden.json) shows a disabled
candidate rule and a new permission failing, followed by a passing prose revision.
Library and CLI tests cover dictionaries, baseline expansion, inherited resource
removal, ambiguous targets, permission unions, and dirty-input rejection.
