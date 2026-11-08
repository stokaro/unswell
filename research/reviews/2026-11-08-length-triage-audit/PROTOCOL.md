# Length-only triage feasibility audit for #349

Status: post hoc arithmetic on exposed assistant review, November 8, 2026.
This audit was devised after the editorial decisions and source text were read.
It cannot qualify a policy change or estimate prospective performance.

## Question

The technical whole-page stream has 82 accepted findings among 126 delivered,
while 28 of 57 frozen events are FULL. Thirty-eight of the 44 nonaccepted
findings are `syntax.long-sentence`, but 46 findings from that rule are accepted.
Could raising only that rule's prose-word onset remove enough nonaccepted
findings to help the unchanged 85% delivery gate without discarding many
accepted findings or an existing FULL-event witness?

Use the source-hashed complete technical report from the quality-claim screen,
its aggregate cohort result, and the earlier complete assistant review's
finding/event ledgers. `freeze.json` pins these files. No reference event,
finding decision, source page, or denominator is changed. Confirm that the
engine and policy inputs affecting this rule have not changed since the report.

For every `syntax.long-sentence` finding, take only its recorded prose-word
length, cohort, ID, and earlier review status. Treat both `uncertain` and
`nonactionable` as nonaccepted, as the gate does. Report pairwise AUROC for
length predicting `actionable` only as descriptive rank separation. Simulate
each integer onset from the existing 35 through the largest observed length:
the rule emits only when sentence words are strictly greater than onset.
Count removed accepted and nonaccepted findings and any originally FULL event
whose complete credited-finding set would disappear. Keep all 123 original
event records in the audit.

For whole pages, also show an intentionally optimistic projection: add exactly
one perfectly accepted finding for every currently missing FULL event, plus
one to replace any lost FULL event, with no other changes. This is an arithmetic
ceiling for a hypothetical future recall mechanism, not measured performance.
An onset that crosses 85% by hiding useful findings is not an acceptable
policy. No threshold is selected or deployed from this exposed audit.

## Decision boundary

If a length-only onset must remove many accepted findings to approach 85%,
retire onset tuning as the next policy step. Investigate source-local editorial
necessity and multiple missed event categories with separate technical
controls. Any candidate needs a frozen full-stream comparison and untouched
prospective confirmation; #349 keeps the 80% FULL recall and 85% accepted
delivery targets in every complete cohort.
