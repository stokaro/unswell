# Audit the representation before fitting another detector

This audit uses existing, exposed whole-page reviews. It introduces no new
editorial judgments, fresh confirmation sample, runtime rule or model. Its
purpose is to measure whether the shared extracted blocks and prepared model
units retain the prose needed to represent the annotated editing operations.

The starting runtime is merged main
`fa1743891a3e82d2d5361f41762604767b497b8f`. The preceding structural-wordiness
study reported 150/947 full detections on exposed pages and 2/63 on its new
pages. Those labels and outputs are now exposed development evidence. A
representation audit cannot establish improved diagnostic recall.

## Inputs and denominators

Use the original full-page review and all subsequent complete-category reviews
through structural-wordiness. Exclude the narrow framing and repetition-scope
confirmation reviews from aggregate results. Keep every included source,
including pages without defects, code-heavy pages and long references. Validate
source hashes and every annotation quote against the original UTF-8 bytes.
Preserve reviewers, uncertain judgments, controls and multi-location events.
Never turn an unmarked fragment into a confirmed clean training example.

Record the selected manifests, annotations and archives by digest. Reject
duplicate repository/path identities rather than counting a new revision of an
exposed page as another observation. Do not use origin metadata as a target.

## Representation measurements

Run the current shared extraction and English NLP provider, not a regex-only
replacement. Export extracted blocks and `nlp.PrepareUnits` with sentence,
paragraph and fragment kinds. Keep source segments and the identities of the
provider and unit contract. Record protected gaps and resource errors; do not
silently omit a page or treat an error as a negative result.

For each frozen defect, collect source bytes belonging to unprotected Word
tokens that overlap its annotated targets. Measure whether one extracted block,
one prepared sentence, or one prepared enclosing unit contains all those bytes.
Also report the number of blocks and prepared pieces required. Count a
multi-location event once. Events with no eligible prose tokens are a separate
category, not automatically representable or detected. Retain their source
locations for inspection.

This is an optimistic representability bound: a unit containing the target may
still lack a necessary reason, earlier definition or neighboring example. It
does not diagnose the defect, verify the proposed edit, or establish that a
model using that unit can learn it. Conversely, a multi-unit event may be
diagnosable by a relational rule that links several source locations.

Record contextual controls and uncertain targets independently. A unit can
contain both a defect and a useful technical statement. Such co-location must
not silently turn the statement into defective prose or the whole paragraph
into an accepted negative example.

## Decision after measurement

Use the measured failure categories to choose the next experiment: repair
extraction or context assembly where it loses necessary evidence; otherwise
test a broader representation against the existing rules baseline. Any learned
candidate remains a separate research result until a new source-grouped
confirmation sample and diagnostic-specific precision/recall support it.
Assistant review stays assistant review under ADR 0041; no classifier response
becomes a calibrated editorial probability or an authorship claim.
