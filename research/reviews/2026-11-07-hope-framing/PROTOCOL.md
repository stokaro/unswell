# Bounded hope-framing screen for #349

Status: exposed development hypothesis, November 7, 2026. The implementing
Codex assistant has already read the 36 original pages, their 123 reference
events, and the four occurrences of `hope` or `hopefully`. This is neither a
blind test nor independent confirmation. It introduces no model call.

## Hypothesis

Three missed editorial events use an uncommitted hope to carry a document
endorsement, an indirect design requirement, or a speculative future aside.
The existing `filler.evaluative-closure` already diagnoses a fourth, direct
document-help hope. Extend that rule with bounded *relations* rather than a
word blacklist:

1. An author hopes that a document, together with an abstract process it
   entails, will help achieve an anaphoric goal. Target only the hope clause;
   retain the preceding stated goal and the document's actual scope.
2. An impersonal `it is hoped that if ... would ...` clause frames a design
   requirement as a wish. Retain the condition, actor, modality, and desired
   capability when requesting direct wording.
3. `Hopefully, some day, ... will become unnecessary` speculates about removal
   of a named feature without an action, date, or commitment. Retain current
   feature status and warnings; target only the speculative sentence.

Require the complete affirmative construction in eligible prose. Quotation,
attribution, negation of the hope, explicit dates or measured conditions, and
protected words used as construction operators exclude the match. Protected
identifiers may remain opaque subjects or objects. Keep original UTF-8 spans.
Ordinary `hope` about a concrete user action, an attributed roadmap plan, and
technical conditions are close controls. Do not broaden these constructions
after observing the replay.

## Frozen inputs and comparator

Compare merged main `0b5433f8289d27f97c905590b44acfaa221792c6` with one
candidate rule version. Reuse the original 26 whole pages, six context pages,
and four historical confirmation pages, with unchanged event denominators
57/46/20. Use the current-main assistant finding decisions in
`../2026-11-07-current-main-editorial-review/` and the original source archives
referenced there. The post-merge report in `../2026-11-07-quality-claim-screen/`
is the immediate before stream. Preserve the previous accepted findings and
event credit; remap finding IDs only by exact rule, source, span, and evidence.

Record the complete technical and strict before/after streams, every added or
removed finding, original-source spans, accepted/uncertain/rejected dispositions,
FULL-event gains and losses, unavailable pages, and advice safety. Inspect every
changed warning in complete source context. Reuse the 18 external controls and
add close affirmative, conditional, attributed, quoted, dated, protected, and
repaired examples. Do not infer a population precision estimate from them.

Accept this bounded development candidate only if it gains at least three
previously non-FULL events, all new findings are accepted under ADR 0041, no
accepted finding or FULL witness is lost, all close controls pass, and advice
preserves every fact, condition, identifier, version, and modality. Otherwise
retain the negative result and do not ship it. Passing development permits
only a frozen untouched prospective confirmation and a separate default-policy
decision; it does not satisfy #349's 80%/85% complete-cohort gates.
