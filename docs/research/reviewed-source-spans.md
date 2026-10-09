# Reviewed source edits and span learning

The completed comparison rejects both registered delivery configurations. Exact
edit annotations did not yield adequate source-only diagnostics: the unit-mean
arm had one available source group and proposed two criticisms; the token-span
arm had none. Neither stream supplied a byte-covering witness for any of the
123 unchanged reference events. A covering witness would still need semantic
review; no product improvement is claimed.

The [measurement record](reviewed-source-spans-result.json) preserves annotation
counts, registration and completion identities, every calibration group, and
the independent reconstruction. Published core `8a56d13f2ac0` remains at 37/123
FULL events and 94/154 useful actual criticisms. The requirements remain at
least 80% FULL recall and 85% useful actual criticisms in **each** cohort, zero
unsafe or unresolved supplied advice, and separate prospective confirmation.

A subsequent [raw-response audit](partial-source-labels.md) found that the
historical annotation helper defaulted unmentioned intent dimensions to
negative. Those fields must not be described as individually explicit negative
reviews. This completed failed experiment remains frozen; the replacement
contract and new acquisition are separate, with no retroactive label or fit
changes.

## Source review and aligned edits

The root Codex assistant separately reviewed the original and revised text for
25 pilot cases and 208 further cases under
[ADR 0041](../adr/0041-assistant-review-acceptance.md). These are the previously
exposed 233 source units, not 233 new independent examples. The reviewer used
complete original units, adjacent units, headings, and the original pages.
There was no independent human annotation or new provider request.

| Annotation item | Completed result |
| --- | ---: |
| Original source cases | 233 |
| Necessary / retain / unresolved current decisions | 109 / 122 / 2 |
| Reviewed original-byte replacements | 137 |
| ERRANT-suggested alignment operations | 358 |
| Cases conflicting with historical supervision | 75 |
| Quarantined cases, including unresolved cases | 77 |
| Admitted original cases: necessary / retain | 70 / 86 |
| Admitted unique before/after variants | 226 |
| Nonempty variants supplying training loss | 223 |
| Admitted positive edit spans | 87 |

The conflicting cases supply no training loss. Their older annotations and all
original evaluation events remain unchanged. Three empty repaired units are
excluded from encoding; their original units remain in training and evaluation.
Unknown and optional intent decisions remain masked. A retained construction,
an unchanged sentence, or an unmarked token is never an inferred clean label.

[ERRANT](https://github.com/chrisjbryant/errant) aligned the reviewed pairs in an
isolated Python environment: ERRANT 3.0.2, spaCy 3.8.16, rapidfuzz 3.14.6, and
`en_core_web_sm` 3.8.0. Independent coordinate checks rebuilt revisions from the
original UTF-8 bytes and checked protected spans. The expansion took 6.25 seconds
after the earlier pilot had loaded the language model. Neither the operation
types nor the number of differences establishes editorial necessity or detects
rhetorical patterns.

The usable spans remain sparse: 37 empty-framing, 25 fluency, 12 wordiness,
five vague-claim, four intensifier, two clarity, one repetition, and one
complexity target. This inventory is not exhaustive annotation of the
4,549-unit population.

## Matched offline comparison

Both arms use the same frozen ModernBERT-base revision
`8949b909ec900327062f0ebf497f51aef5e6f0c8`, eight intent decisions, seed 61092,
120 epochs, and a 1536-to-64-to-8 neural head. The source input contains the
complete target and its complete preceding and following units. All input
tokens are covered by windows of at most 1,024 tokens with 128-token overlap
when needed. No whole-document vector, teacher reason, proposed repair, label,
or project identity enters the predictor.

The unit arm applies the head to pooled target-token features. The span arm
applies it to individual target tokens and max-pools for unit-intent supervision.
Both use masked intent loss. The span arm adds positive supervision on tokens
overlapping the reviewed edit; unmarked tokens remain unknown. Separate review
of the revised unit supplies resolved intent negatives. Its whole-unit
necessity remains null, including when a required edit was made successfully.

The unchanged 13 source groups define nine training groups, three calibration
groups, and one outer test group per fold. Original units and their revisions
stay together. The groups are not 13 independent repositories. Calibration uses
only explicitly graded original units, so repaired versions cannot inflate its
sample count. Each arm accounts for every original unit exactly once.

Calibration selects a point with at least ten selected known originals and
85% decision precision, requiring both necessity classes. The highest observed
positive count wins, with fewer false positives and a higher cutoff breaking
ties. The fixed cutoff grid is -4 to 4 in steps of 0.5. A missing feasible
point is unavailable, rather than a successful gate or a probability.

Delivery emits at most two trained-positive intents per unit. The unit arm
quotes the complete original unit; the span arm proposes contiguous token
ranges above its cutoff, excluding protected tokens and any range crossing
protected content. Neither generates replacement text. This difference in
delivery geometry is part of the comparison, not evidence of production safety.

## Complete result and limits

| Arm | Available source groups | Original units accounted for | Available original units | Proposed criticisms | Byte-covering witnesses | Witnesses with matching intent |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Unit mean | 1/13 | 4,549 | 226 | 2 | 0/123 | 0/123 |
| Source span | 0/13 | 4,549 | 0 | 0 | 0/123 | 0/123 |

All 46 context events, 57 whole-document events, and 20 historical confirmation
events remain in the evaluation, including joint and unavailable cases.
Historical confirmation has already been exposed; it cannot substitute for an
untouched prospective test. The proposed streams have no completed semantic
review, so actual usefulness and semantic FULL are recorded as unavailable.

Seven calibration groups fail the count screen before model quality is
considered: `bat`, `buildkit`, `containerd`, `imgui`, `prometheus`, `ripgrep`,
and `tokio` have fewer than nine necessary originals. Even a perfect classifier
cannot reach 85% among ten selected cases there. The other groups have enough
positives by count, but only the unit-mean `fzf` fold reaches the registered
score requirement. The result therefore combines insufficient annotation
coverage with failed learned operating points; it does not prove all models
intrinsically incapable of the task.

All 26 head fits completed in 148.26 seconds. Frozen feature extraction covered
4,549 originals and 67 reviewed revisions, with 112,987 target tokens, in
105.71 seconds and a 1.77 GiB peak GPU allocation. An initial registration
failed at unavailable Metal in the filesystem sandbox before extraction or
fitting. The separately sealed execution amendment changed that environment,
preserving the algorithm, labels, and hyperparameters.

Independent reconstruction verified the source bytes, masked supervision,
derivative holdouts, token coverage, 26 saved-head calibration grids, native
terminal exits, and all event witnesses. Two audit mistakes, an exclusion-count
assertion and a receipt filename, were retained and corrected outside the frozen
experiment. No source review, model fit, threshold, or result was rewritten.
The public record contains aggregates and provenance hashes; the private review
packets and local feature files are not published. Complete public reproduction
is not claimed.

## Further primary research and Python code

The following sources were checked on October 9, 2026. None was installed or
fit as an additional Unswell candidate in this comparison.

| Source | Useful next experiment | Boundary |
| --- | --- | --- |
| [Re3-Sci2.0, EMNLP 2024](https://aclanthology.org/2024.emnlp-main.839/) and [author Python code](https://github.com/UKPLab/emnlp2024-llm-classifier) | Train an edit-intent classifier, then use its qualified predictions to expand revision annotations. The authors label over 94,000 edits after training on a smaller reviewed corpus. | Its input is an original/revised pair. Grammar, clarity, claims, and facts are edit purposes; they are not source-only necessity labels. The [dataset record](https://tudatalib.ulb.tu-darmstadt.de/items/469ec002-60a1-4fa7-8183-e783ad9a81ff) specifies CC BY-NC 4.0, separate from the code's Apache 2.0 license. |
| [Revision requirements in wikiHow, EMNLP 2020](https://aclanthology.org/2020.emnlp-main.675/) and [author Python code](https://github.com/irshadbhat/wikiHow_MoRR) | Separate source-sentence necessity classification from pairwise revision ranking. This is closer to the product decision than authorship classification. | Historical edited/unedited sentences are proxy labels; optional edits and unnoticed defects are not resolved by that history. The old implementation also truncates inputs by default. |
| [Discourse-connective edits, CODI 2025](https://aclanthology.org/2025.codi-1.19/) and [author corpus](https://github.com/berfingit/connective-deletion-replacement) | Teach contextual retention, necessary edits, and optional preferences for connectives, using both adjacent sentences. | In the reviewed replacement subset, most changes were optional. Surprisal alone did not reliably establish necessity. The corpus specifies CC BY-NC-SA 3.0; it is research input, not a stock phrase ban. |
| [Unified edit-intent taxonomy, ACL 2025](https://aclanthology.org/2025.findings-acl.1180/) | Audit mappings before transferring another corpus's edit labels to our eight intents. | A label-name match does not establish equivalent definitions or annotation scope. |

The next comparison should separate a source-only necessity objective from
edit-purpose and localization objectives, using only explicit source decisions
for necessity. Reviewed revision pairs can support ranking and auxiliary
intent training without making every revised sentence a clean negative.
Annotation acquisition must first provide adequate calibration coverage for
each source holdout. Adding another head on the same count-infeasible packet
would not test that improvement.

External revision corpora can supply auxiliary supervision after license,
domain, split, and intent-mapping checks. Their annotations must remain
distinguishable from our mandatory-edit labels. Any eventual diagnostic still
needs complete-source semantic review, information-preservation checks, and the
same unchanged complete evaluation before prospective confirmation.

The [source-unit](source-student-learning.md),
[encoder-adaptation](encoder-adaptation.md), and present span configurations are
completed development comparisons. This result adds no runtime model, changes
no gate, and leaves the detection objective open in
[#349](https://github.com/stokaro/unswell/issues/349).
