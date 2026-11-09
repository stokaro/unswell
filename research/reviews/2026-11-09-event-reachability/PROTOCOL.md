# Current-engine source reachability audit for #349

This is a frozen offline audit of the 123 exposed editorial events, not a new
diagnostic method or product qualification. It asks whether the current engine
exports source-mapped eligible prose that overlaps each frozen target. The
answer determines whether extraction availability is a hard recall ceiling
before another semantic candidate experiment. No hosted model, new grade,
changed event, or revised threshold is allowed.

## Inputs and execution

Use the SHA-256-pinned `joint-component-sol-high-evaluation/originals.json`
archive: 36 original UTF-8 pages, 123 reference events, paths, and repository
groups. The frozen main commit and tree identify the existing `reviewtext
--source-inventory` exporter. Build that exporter from the frozen tree, pin the
binary SHA-256, and run it once for each page with its exact original bytes.
The exporter uses the engine's technical-v1 extraction and unit preparation;
no second Markdown parser is allowed. A page error is recorded as unavailable
and its events stay in the denominator. No retry or source substitution.

For every exported unit, verify the source SHA-256, input SHA-256, original raw
slice, byte segments, and absence of overlap between eligible segments and
excluded spans. For every frozen target, verify the original quote against its
byte range before scoring. The audited content bytes are Unicode alphanumeric
characters in each target's exact UTF-8 slice; Markdown delimiters and
whitespace cannot create a false reachability hit.

Report, by original whole-page, contextual, and historical cohort, the number
of events whose every target contains eligible content, whose some target
contains eligible content, whose every content byte is eligible, and whose no
target content is eligible. Also report unavailable pages/events and targets
intersecting excluded spans. Keep the 57/46/20 event denominators and the 80%
minimum counts (46/37/16). An event with no eligible target content cannot be
source-locally diagnosed by a unit-bound rule under this extraction contract.
An event with overlap may still be impossible to diagnose correctly; overlap
does not prove semantic recall, accepted delivery, or safe edits.

Freeze this protocol, script, archive identity, main tree, and exporter binary
hash before the first 36-page audit. Publish aggregates only. Keep per-event
coverage details in a local, untracked file. If a page fails, preserve the
error category without printing private source content. Do not adjust the
input, metric, or event labels after seeing the result.
