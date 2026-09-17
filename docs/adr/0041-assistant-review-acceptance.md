# ADR 0041: Accept assistant-reviewed diagnostic evidence

Status: accepted by the maintainer on September 17, 2026.

## Context

The whole-page audit in [#298](https://github.com/stokaro/unswell/pull/298)
records missed defects and false alarms with source spans and proposed repairs.
One Codex assistant supplied the judgments. The maintainer accepted that review
and confirmed that an independent reviewer will not be available. Diagnostic
work must proceed without waiting for a native speaker or a second annotator.

## Decision

Assistant review accepted by the maintainer is sufficient editorial evidence for
diagnostic development and rule acceptance. This updates the human-review
prerequisite for rules in ADR 0037, the roadmap and #26. Acceptance still depends
on the evidence for the particular rule and its intended default behavior.

- Identify the reviewer as an AI assistant. Record the rubric, source revisions,
  exact spans, rationale, proposed repair and unresolved cases.
- Keep acceptable technical counterexamples, including necessary reasons,
  qualifications, repeated identifiers and supported guarantees. Preserve their
  meaning when assessing a proposed edit.
- Measure misses as well as emitted findings. A warning earns recall credit only
  when it identifies the annotated defect at the relevant source locations.
- Freeze separate complete-page confirmation inputs before tuning. The same
  assistant may review them. After inspecting their outputs, retain them as
  exposed evidence; further tuning needs a new confirmation sample.
- Publish before/after findings and dispositions, sample sizes, uncertainty and
  unresolved judgments. Describe precision and recall as agreement with the
  named review on the stated sample. A single assistant's review cannot measure
  agreement between raters or establish population-wide accuracy.
- Retain regression tests, source integrity, resource bounds and ordinary CI.
  Review acceptance supplies no automatic change to a rule's severity or gate.
  Any default promotion needs its own recorded evidence under #26.

The existing whole-page annotations remain frozen. A later correction is a dated
amendment with its effect on the measurements, never an unrecorded label change.
The review's limitations remain visible, but independent human review is not an
open prerequisite for completing #299, #300 or diagnostic qualification in #26.

## Scope

This decision applies to editorial diagnostics. It does not supply the missing
training corpus, model calibration or held-out probability evaluation. The
deferred model-research protocol in #22/#25 retains its separate scope. Its
human-rater artifact checks do not govern acceptance of diagnostic review records.
Assistant labels must never be serialized as human labels to pass those checks.

No authorship claim follows from accepting an editorial judgment. The product
continues to report concrete wording problems under its selected policy.
