# ADR 0007: trusted policy for committed analysis

Status: accepted. Merged-CI acceptance is recorded in issue #15.

## Required behavior

A pull request must not pass a required editorial check by weakening its own
configuration, replacing a dictionary or ruleset, expanding a baseline, or adding
an unreviewed source permission. The existing committed comparison in ADR 0006
rejects changed policy resources; this work adds an explicit trusted source and
visible policy-change handling to that engine.

The application must resolve the trusted revision once, load its configuration
graph and rules from immutable Git objects, and record the resulting resource and
engine identities. Config inheritance and resource limits must use the existing
appconfig/config implementation. Working-tree discovery must not determine which
trusted resources are loaded. A missing trusted dependency must be an error.

The analysis library continues to receive source bytes and policy. It must own
scoring, selection, and gate decisions; the CLI cannot restore or remove gate
reasons as an independent policy implementation. Policy changes require an explicit
audit and the full scan required by the specification. Source permissions and
baseline expansion must remain visible even when the trusted rules would otherwise
produce a passing result.

## Decision

Add `--policy-from-base` to committed checks. It requires `--changed-from REF` and
uses that comparison's merge-base commit as the policy source. Configuration
discovery, inheritance, dictionaries, explicit ruleset paths/directories, and an
explicit baseline are loaded from that immutable tree. Profile and gate-mode
overrides are rejected in this mode. Source selection uses the trusted policy's
include/exclude settings. Positional paths and execution flags remain owned by the
trusted invocation.

Compare HEAD identities for the trusted resource footprint and discovery/directory
membership. Record additions, modifications, and removals without applying HEAD
policy. Any such change requests a full scan of the selected current documents
under the trusted policy. A policy edit does not automatically fail clean prose;
it cannot weaken the gate used for that run. Reviewing the policy that a later run
will trust remains a maintainer responsibility. Changed referring resources cover
new HEAD dependency graphs; this mode does not compile an unused candidate policy
or claim to validate its syntax. Run ordinary config validation when reviewing it.

Source permissions require their own check inside the engine. Match each current
permission against the base by rule IDs, reason, kind, and structural target
identities. Only unambiguous matching permissions can affect the current gate.
Exact complete-source equality can retain existing duplicate identities. New,
expanded, retargeted, or ambiguous permissions remain in the audit but cannot
suppress evidence. A finding requiring several permissions is allowed only when
every contributing permission is trusted. Changed bindings select the whole file.

Expose an additive `AnalyzeChangedWithOptions` method with explicit trusted-policy
options and caller-supplied resource-change metadata. Ordinary `AnalyzeChanged`
retains its behavior. Both methods share extraction, NLP, identities, scoring, and
gate code. Report trust state and policy differences in the completed result; no
reporter or CLI reconstructs a policy decision. The engine treats supplied policy
as a caller assertion and performs no Git or filesystem access.

Baseline bytes also come from the base. A larger HEAD baseline cannot accept new
debt. Existing compatibility checks still apply. Embedded rules, NLP, and model
identities belong to the pinned checker; this feature does not build or execute
historical source code. The current alpha still requires `calibration.model: none`.

Cover removed or invalid HEAD policy files, changed references and ruleset
directory membership, baseline compatibility and expansion, permission unions,
duplicate targets, and unchanged ordinary drafts. Verification must include the
built CLI, saved reports, and a read-only real-repository example.

The checker binary and invoking workflow must themselves be pinned to a trusted
source. A Git object ID records identity, not organizational approval. A library
cannot stop a pull request from replacing the workflow or expected policy hash when
repository protections permit those changes.
