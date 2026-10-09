# Partial labels and source-review acquisition

The [reviewed source-span study](reviewed-source-spans.md) failed the original
recall screen. Inspection of its raw review responses then found a supervision
problem: its historical normalization helper assigned `not_necessary` to each
intent the response had not named. Choosing one edit purpose does not establish
that every other purpose is absent.

The replacement [Python contract](../../research/annotation/source_review/README.md)
preserves unknown labels, supports several purposes per edit, and binds quoted
targets to original UTF-8 bytes. A separately registered acquisition completed
100 additional source reviews under that contract. These are data and contract
results. The subsequent [matched training study](partial-label-learning.md)
fits the corrected masks separately; it does not qualify a detector.

## Distinguishing the annotation tasks

A whole-unit decision about mandatory editing, the purpose of a particular
repair, and the need to detect a multi-component reference event are different
judgments. For example, `wordiness` and `empty_framing` can describe the same
removal. Selecting one does not supply a negative for the other.

The 233 historical reviews contained 1,629 normalized `not_necessary` intent
fields. Raw responses expressed selected necessary, optional, and unresolved
purposes; the helper supplied the other negatives. These normalized fields
must not be represented as individually explicit negative annotations.
Earlier packet hashes, labels, quarantines, fits, and evaluation events remain
unchanged.

A read-only comparison found 43 of 123 reference events with at least one
component whose normalized current intent was negative. Forty-two had every
component marked that way. These are adjudication candidates, not 43 proved
reference-label errors. The historical audit's use of the word `explicit` for
those normalized fields was inaccurate; inspecting the raw responses revealed
the default-negative path. A whole-unit review also cannot settle a relation
between several source blocks without reviewing that relation.

The 75 historical supervision conflicts divide into 43 binary-necessity
disagreements and 32 intent-only disagreements. Masking a disputed intent while
retaining an agreed binary label would preserve more cases than excluding the
whole case. That is a future packet decision, not a retroactive change to the
completed study. Count-only previews still leave insufficient calibration
coverage, so no new fit is justified merely by that distinction.

The replacement adapter requires a reason for a known intent negative.
Unmentioned, optional, and unknown dimensions remain masked. Retaining a unit
does not populate all eight intent negatives. A repair can have nonexclusive
purposes, and the revised text requires its own review. The trainer must skip
an all-unknown intent-loss row instead of averaging an empty tensor or filling
unknown entries with zeros. The adapter's 14 regression tests also reject
unlocalized mandatory decisions, malformed outputs, overlapping targets,
ambiguous repeated quotes, unchanged rewrites, and protected-source edits.

## Completed acquisition

The acquisition used five exposed development projects: etcd, fzf, imgui,
ripgrep, and tokio. They occur in calibration partitions that lacked enough
explicit necessary cases in the preceding study. Each project contributed ten
random-control units and ten units selected by feature coverage. The arms were
outside the mixed review packet during annotation.

Feature coverage used greedy maximum cosine distance from existing reviewed
target-feature anchors and the random selections within that project. It used
the cached frozen ModernBERT target features, seed 61093, and no classifier
scores, teacher explanations, new encoder fit, or prospective source.
Previously reviewed units were excluded.

An initial packet overrepresented short headings and labels. It was superseded
before annotation by a separately registered minimum of 16 eligible target
tokens. This is a tokenizer threshold, not a guarantee of a complete paragraph:
short link labels and changelog fragments can still satisfy it.

The root Codex assistant reviewed the complete target and adjacent original
units under [ADR 0041](../adr/0041-assistant-review-acceptance.md). The record
makes no independent-human or whole-document quality claim. Nonlocal repetition
and unsupported fact assumptions remain ungraded unless explicitly addressed.
All 100 original byte slices, labels, masks, and the selection sequence were
independently recomputed after review.

| Acquisition arm | Cases | Necessary | Retain | Optional | Unknown |
| --- | ---: | ---: | ---: | ---: | ---: |
| Random control | 50 | 8 | 29 | 12 | 1 |
| Feature coverage | 50 | 1 | 38 | 11 | 0 |
| Combined | 100 | 9 | 67 | 23 | 1 |

The nine necessary units contain 12 localized edits. Eight have a fluency
purpose and two a clarity purpose; one unit has both. They do not add a
necessary rhetorical-framing label. Across the 800 intent dimensions, 21 have
observed binary labels and 779 remain optional or unknown. Global necessity
has 76 binary labels, independently of the per-intent masks.

Feature diversity did not enrich mandatory-edit cases on this packet. These
selected-pool counts estimate neither population prevalence nor classification
accuracy. They do not prove that every feature-coverage acquisition strategy
fails. Repeating this selection as though it were a successful active learner
would be unsupported.

The preceding calibration rule required both known necessity classes, at least
ten selected known cases, and at least 85% decision precision. A partition with
fewer than nine known positives cannot meet that rule regardless of model
scores. Adding the new binary labels in a count-only preview reduces such
partitions from seven to five; the remaining test groups are bat, imgui,
prometheus, ripgrep, and tokio. Meeting a count bound does not establish a score
threshold, delivered-criticism usefulness, or model availability. The preview
does not silently merge the new reviews with the old flawed intent packet.

The [measurement record](partial-source-labels-result.json) preserves hashes,
native completion, group counts, and calibration previews. The private raw
annotation packet is retained locally. Publishing aggregate hashes and a
reusable contract does not claim public reproduction of every private judgment.

## Research and Python tools

Primary sources and implementation details were checked on October 9, 2026.

| Source | Application | Boundary |
| --- | --- | --- |
| [Multi-Label Learning from Single Positive Labels, CVPR 2021](https://openaccess.thecvf.com/content/CVPR2021/html/Cole_Multi-Label_Learning_From_Single_Positive_Labels_CVPR_2021_paper.html), [author Python implementation](https://github.com/elijahcole/single-positive-multi-label) | Study missing-label losses separately from the baseline that assumes every unmentioned label is negative. The author code was inspected at `89152d52ba59a65d75b19cbe32254a037b557a57`. | Its experiments concern image classes. Some losses regularize an expected positive count, which our selected editorial packet does not estimate. No loss from the paper was fit here. |
| [Revisiting Active Learning under Label Variation, NLPerspectives 2025](https://aclanthology.org/2025.nlperspectives-1.7/) | Distinguish annotation errors from plausible preference variation; preserve optional choices and disagreement in acquisition. | This is a conceptual framework, not an implemented editorial classifier or measured gain for Unswell. |
| [Distilabel custom tasks](https://distilabel.argilla.io/latest/sections/how_to_guides/basic/task/), [structured generation](https://distilabel.argilla.io/latest/sections/how_to_guides/advanced/structured_generation/) | A Python annotation pipeline can retain raw prompts, responses, model identities, and available token statistics while parsing structured labels. Context7 and the official documentation were inspected. | JSON-schema conformance supplies no semantic quality guarantee. Distilabel was not installed or run, and this work made no provider requests. |
| [Identifying Reliable Evaluation Metrics for Scientific Text Revision, ACL 2025](https://aclanthology.org/2025.acl-long.335/) | Its abstract reports complementary evidence from task-specific metrics and LLM judging, with instruction-following easier to judge than correctness. | Only the abstract was inspected. No paper result is used as technical-preservation evidence or an Unswell accuracy estimate. |

The subsequent [registered comparison](partial-label-learning.md) preserves
partial labels and separates binary necessity from nonexclusive intent and span
losses. It still fails the recall screen. Further data needs contextual-relation
review, explicit hard retention controls, and calibration support across
projects. Source/revision pairs can teach where and why wording changes, but a
plausible rewrite alone does not make the original incorrect. The
[earlier methods review](learning-data-methods.md) records other implemented and
untested Python candidates.

Published detection remains 37/123 FULL events and 94/154 useful actual
criticisms on all 36 original pages and 4,549 units. All 804 original judgments
and 123 events remain in scope. The requirements remain at least 80% FULL and
85% useful in each cohort, zero unsafe or unresolved supplied advice, followed
by separate untouched confirmation. This acquisition made no runtime, model,
threshold, gate, or public-site change and supplies no product acceptance
credit. [Issue #349](https://github.com/stokaro/unswell/issues/349) remains open.
