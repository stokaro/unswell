# Source-bound editorial claim accounting

The developer command `go run ./cmd/reviewclaims` reads one explicit JSON input
from stdin and writes a preparation or accounting result. Run it from
`research/annotation`. Its [request schema](request.schema.json) defines the
current experimental format. [ADR 0042](../../../../docs/adr/0042-editorial-claim-accounting.md)
records the boundaries. This is not a model or an authorship detector.

Provide `version: "unswell-editorial-claims-v1"`, one `source` containing its
path, format, original text, and SHA-256, and arrays `claims`, `stages`, and
`approvals`. `include_raw` is optional and defaults to false. The source is
parsed by the existing extractor with its conservative defaults. Output `units`
and `protected` retain the resulting block IDs, spans, and exclusions without
source text. Empty `claims` and `stages` can prepare this location inventory.

Each claim specification declares `origin` (`run`, `candidate`, and `key`),
category, diagnostic, reason, suggestion, `targets`, and `support`. References
contain the engine `block`, original byte `span`, and exact `quote`. Targets
are nonempty; support is ordered, disjoint context that is not itself an edit
target. References must lie in eligible prose and cannot intersect protected
regions. Protected code remains available in the supplied whole source as
context but cannot appear in target or support references.

Prepare the declared claims with empty `stages` to obtain stable opaque IDs.
Preparation returns `complete: false`; it performs no editorial review. Preserve
the original source and declarations when adding stages. Changing them changes
their IDs and makes old responses inapplicable. Source identity includes path
and format, so identical bytes in separate named sources remain distinct.

Every stage has an ID and arrays of decisions, edits, and duplicate groups.
Each original claim requires exactly one decision with a nonempty reason:

| Status | Accounting behavior |
| --- | --- |
| `retained` | Keep the original claim available for display |
| `resolved` | Cite a concrete `edit_id` and its separate claim-specific approval |
| `rejected` | Preserve the original claim and the reason for rejecting it |
| `uncertain` | Preserve the original claim and report incomplete accounting |

An edit names its ID, exact original target reference, and replacement text.
Edits in a stage must be disjoint, change the target, and resolve at least one
claim. Deletion is permitted only with the same explicit review requirements.
The alternative never becomes the original source quotation. An edit can fix
multiple claims only when each has its own separate approval.

Resolution approvals name `stage_id`, kind `resolution`, the singleton claim ID,
`edit_digest`, reviewer, reason, and `meaning_preserved: true`. `EditDigest`
hashes Go's compact JSON encoding of the complete `Edit` in declaration order
(`id`, `target`, `replacement`). The target order is `block`, `span`, `quote`;
span order is `start`, `end`. UTF-8 strings use Go's ordinary JSON escaping,
including HTML and U+2028/U+2029 escapes. Do not hash a pretty-printed response.
An approval for another stage, edit, or claim cannot be reused.

Duplicate groups declare claim members, a representative, and a reason.
Their separate `same_defect` approval names the exact member set, reviewer, and
reason, with an empty edit digest and `meaning_preserved: false`. All members
must be retained and may occur in only one group. The original claims stay in
the artifact; `displayed_claims` contains representatives. Review of semantic
equivalence is an explicit caller responsibility.

Treat approvals as trusted developer input, collected separately from model
stage responses. A model must not supply its own trusted approvals. Reviewer
identity is recorded, not authenticated. All accounting results remain
`editorial_qualified: false`. Neither retained nor resolved is an editorial
accuracy label, and complete accounting is not a passing product gate.

The command omits original quotations, replacements, diagnostics, and reasons
unless `include_raw: true`. Source spans, IDs, categories, disposition states,
and reviewer declarations remain. An explicit raw artifact may reconstruct
private prose and must be handled according to the data's existing permissions.
There is no automatic publication or external model call.

Input is capped at 32 MiB, source at 2 MiB, and inventory at 10,000 claims,
blocks, exclusions, total references, edits, or approvals. A claim has at most
64 target or support references; a request has at most 32 stages. Replacement
text is capped at 16,000 bytes, identity/explanation strings at 2,400 bytes.
These are engineering limits, not statistical applicability thresholds.

The command returns 0 for valid preparation or complete supplied stages, 2 for
invalid input, uncertain required stages, or operational errors, and 130 for
cancellation. Uncertain replay still emits its incomplete record. Other invalid
inventories produce no partial success result. Keep original research labels and
full-event denominators separate from this representation check.

The optional [downstream consumer](../../../reviews/2026-09-30-reviewed-claims/README.md)
uses this command for the full original inventory across audit, rewrite, and
selection. It binds separate caller approvals to the exact stage file, preserves
mixed parent outcomes, and refuses to treat an uncertain earlier stage as complete
because final selection succeeded. It adds no parser, model call, or product gate.

## Preserve the basis of existing rules

Run `go run ./cmd/explainrules < result.json > observations.json` from
`research/annotation`. The input is one complete canonical engine report,
validated by the existing report reader. Output uses
`unswell-rule-evidence-v1` and retains every original finding's ID, rule version,
source hash, primary/related locations, suppression state, and full evidence.
Incomplete reports, duplicate identities, derived threshold diagnostics, missing
measurements, invalid numbers, cancellation, and writer failures produce errors.

`reason` names the actual evidence kind, activation, source-occurrence count,
metric values, units, onset, and saturation, plus any original evidence message.
It records the observed rule condition; it does not assert that a revision is
necessary. `action` retains the explicit evidence suggestion when present.
Otherwise it copies the engine's existing diagnostic and identifies that basis
as `engine_diagnostic`. This is review guidance, not invented replacement prose.
Every `editorial_verdict` is `unreviewed` and `editorial_qualified` stays false.

The projection omits source bodies and location snippets. Existing diagnostic,
guidance, or evidence text can still contain excerpts, so the artifact inherits
the input report's data permissions. It is not published automatically.
Target/support roles and independently actionable claims still need explicit
review through the existing accounting contract. The
[original-baseline replay](../../../reviews/2026-09-30-rule-evidence/README.md)
keeps the frozen packet intact and rejects protected target ranges rather than
changing their quotations to obtain a valid result.

The [explicit migration consumer](../../../reviews/2026-09-30-reviewed-claims/README.md)
accepts separately supplied reviews of original parent criticisms and source
roles. It keeps original parent records unchanged, binds declared children through
this command, and leaves unreviewed parents in the original denominator. Adding
reviewed pages cannot rename unchanged claims. Retention is not semantic acceptance;
the complete migration and later review/selection stages remain separate work.
