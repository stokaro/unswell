# Supported actions with a possible purpose

The existing `filler.instruction-scaffolding` rule recognizes a supported
action followed by a relative modal helpfulness clause and a purpose nominal.
The diagnosis explains both layers. Advice preserves the action, possible
benefit, relative antecedent and any following method. It asks for direct
wording without turning a capability into an obligation or a possible benefit
into a guarantee.

Each complete sentence fits the existing 48-token candidate bound. The action
prefix independently meets the existing supported-action construction. Its
relative clause uses `that` or `which`, a capability modal, `be helpful` or
`be useful`, and `for` with a nominal phrase ending in `purpose` or `purposes`.
The nominal phrase has an action or object head. Numbers, protected operators,
negation, attribution, quotation, permission restrictions, questions, qualified
benefits and split sentences do not establish the specific diagnosis.
Helpfulness alone and a simple active action remain outside this construction.
An existing immediate method keeps its related source evidence. Mixed groups
retain the rule's general explanation.

## Paired measurement

The replay used all 36 original pages, their unchanged source bytes, 804
judgments and 123 reference events. The baseline is Unswell commit
`ff9f5dc269da15580b2ea4b9408a29a45857fc98`.

| Complete source set | Before | After | Changed diagnoses |
| --- | ---: | ---: | ---: |
| Original pages, technical | 146 | 146 | 1 |
| Original pages, strict | 167 | 167 | 1 |
| Separate controls, technical | 7 | 7 | 0 |
| Separate controls, strict | 8 | 8 | 0 |

The changed BuildKit sentence uses a modal passive association wrapper and a
second possible-helpfulness clause ending in a tracing-purpose phrase. The old
generic action warning earned partial credit for the complete original event.
The new explanation identifies both constructions. Its guidance retains the
original pronoun referent, vertex association, possible tracing benefit and
relative antecedent. The root Codex assistant accepted this diagnosis and
guidance under ADR 0041. The rule supplies no automatic replacement.

Every other criticism retains its messages, advice, source locations, related
evidence, metrics, policy and suppression fields. The changed criticism retains
its source evidence and score. Rule version and dependent identities change.
All scans completed without errors or abstentions.

The technical-profile census consequently changes from 28/123 to 29/123 FULL
original events, or 22.8% to 23.6%. The contextual cohort changes from 9/46 to
10/46. The whole-document and exposed confirmation cohorts remain 17/57 and
2/20. Useful actual deliveries remain 86/146, or 58.9%; 55 uncertain and 5
rejected complaints remain in the denominator. Old judgments and the old
partial-event credit are unchanged. The new actual diagnosis supplies the new
complete-event witness. No strict-profile usefulness qualification is claimed.

Six separate complete Ptah pages were selected by tracing, troubleshooting,
migration-log and verification topics before implementation or output inspection.
They are pinned to Ptah commit
`a641f2dd20a9c8596629f43fae0cc2f811d7f5e0` and contain 46,227 source bytes.
Known duplicate source hashes were excluded. They retain all seven technical
and eight strict complaints. No new positive confirmation was observed.
Project-wide exposure history is unestablished; these are separate controls.

Public-engine blackbox tests cover five positive variations, technical controls,
source coordinates in Markdown, Go and Python, policy exemptions, occurrence
allowance, candidate bounds, mixed explanations and retained method ownership.
A CLI blackbox checks both profiles and a preserving revision. Its original
adjacent-method regression fails before the ownership fix.

## Limits

This extension improves one exposed construction's diagnosis. It does not
establish unseen positive recall, population precision or completion of the
contextual quality goal. The original difficult reference census remains below
the 80% FULL recall and 85% useful-delivery requirements. No model calls,
probability calibration, authorship claims or gate changes are involved.
The broader work remains open in
[#349](https://github.com/stokaro/unswell/issues/349).
