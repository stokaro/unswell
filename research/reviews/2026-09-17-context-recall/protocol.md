# Contextual wording recall

This iteration follows PR #306 at
`4fc6f0fdfccfa0a146e7e5581b66427a404eb0aa` and continues #299. Useful complete-page
recall remains the objective. The working target is at least 80% event recall
and 85% precision for soft editorial diagnostics on new pages, with uncertainty
and applicability reported. Known-example gains alone do not meet that target.

The preceding 53 reviewed pages are development material. Preserve their frozen
labels and source identities. The exposed sets have different annotation scopes;
do not pool their denominators or silently relabel a miss as acceptable.

An exploratory scan enables all 51 current rules with their existing parameters.
Keep that policy separate from shipped profiles. Inspect whether disabled rules
cover the missed constructions, including their added false-alarm load. Frequency
differences between historical and generated prose do not establish editorial
precision. Do not enable rules merely to increase the finding count.

## Candidate boundaries fixed before implementation

Extend the existing source-preserving clause analysis. Investigate these families
using the exposed source-bound defects and explicit technical counterexamples:

1. Document self-justification and editorial ownership: deliberate non-repetition,
   explanations of which document owns a detail, and a page announcing what its
   own explanation will mean. A document pronoun may inherit only an explicit
   nearby document antecedent within the same prose block. File/object ownership
   and useful scope/navigation instructions must stay distinct.
2. Evaluative framing: a statement announcing the whole value or point of a
   preceding distinction, cognitive worth announcements, and bare endorsements
   of the behavior just described. Preserve named mechanisms, conditions and
   useful instructions. A numerical value or a concrete purpose is not empty
   framing just because it uses the same noun.
3. Unscoped superiority or benefit claims: affirmative quality predicates with
   an extreme modifier and an abstract quality head, or a comparison with an
   unbounded class of alternatives. Look for a complete construction, not an
   isolated adjective. Keep actual measurements, declared comparison scope,
   quotation, negation and technical uses as negative controls.
4. Assumptions about readers: an unrestricted nobody/everybody claim about what
   people can do, know or want. Access-control requirements, explicit failure
   conditions and concrete actors remain outside this construction. A finding
   identifies the wording and requests a scoped statement; it does not assert
   that the factual claim is false.

These are hypotheses to qualify against the review, not a promise that every
construction will ship. Reuse the current extraction, NLP, source maps, budgets,
configuration and reporters. Do not add a second analyzer, external-model runtime,
authorship inference or a lower global threshold. Keep the existing rule weights
and gate settings; rule versions must identify changed matchers.

## Unexposed confirmation

Use minimum SHA-256 of `context-recall-confirmation-v1\n` plus the pinned source
reference, excluding all 53 preceding sources. Request two pages per
Ptah/historical and short/medium/long cell from the existing pinned frame.
The historical long cell has only one unused source: retain it, disclose that
deficit, and select eleven pages in total. Do not refill the cell with an already
reviewed page or select another length after inspecting findings. Later
iterations need a larger historical source frame.

Read every selected page's source prose across all seven editorial categories
before opening diagnostic output. Preserve exact UTF-8 targets, context, proposed
edits, uncertain cases and acceptable controls. Freeze the complete annotations
before implementing the candidates. Do not tune from confirmation wording or
outputs; these sources can become development material only in a later iteration
with new confirmation. Code, imports, image text and linked destinations are not
silently added to the prose being judged.

Maintainer-accepted review by the implementing Codex assistant is sufficient under
ADR 0041. Record that provenance accurately. This does not create human labels,
independent agreement, calibrated probabilities or population accuracy.

## Measurement and acceptance

Run unchanged technical and strict profiles on complete source pages before and
after. Keep all additions and removals with source-bound dispositions. Credit a
defect once and require the diagnostic to identify its actual construction;
incidental length overlap is not a detection. Record partial matches separately.
Review every confirmation diagnostic, including retained warnings. Report
unresolved judgments, missed events, resource costs, abstentions and actual gate
results without forcing a published page to fail.

Use paired page sensitivity analysis and show cohort, length, format and category
breakdowns. The small sample, exhausted historical cell, shared repositories and
single reviewer limit transfer. Defaults stay experimental unless independently
qualified under #26; that qualification does not require a human reviewer.

Public-API tests and CLI fixtures must preserve source segments, quotes, code,
conditions, quantities, cancellation and budget exhaustion. Keep ordinary
repository checks and CLI/MCP dogfood. Do not activate deferred race, fuzzing or
coverage. Freeze model/rule identities and retain any amendment or failed replay.
