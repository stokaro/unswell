# Rule qualification on the assistant-panel corpus

This record supports the rule defaults and phrase lists that changed after
[ADR 0046](../adr/0046-assistant-panel-corpus.md) and [ADR 0047](../adr/0047-qualified-defaults-and-assistant-review.md).
Every reviewer here is an AI assistant. No number is a human judgment, and none
measures agreement with a native-speaking editor. The labels, reviews, and
manifests are in the
[corpus record](../../research/corpus/2026-11-11-assistant-panel/README.md).

## Documents

| Set | Role | Documents | Repositories | Items | Words |
| --- | --- | ---: | ---: | ---: | ---: |
| Long-prose study: 38 sources and 455 generated versions | development | 493 | 20 | 40,496 | 770,986 |
| Real pages, first development set | development | 164 | 80 | 10,893 | 249,460 |
| Real pages, second development set | development | 593 | 310 | 37,622 | 884,956 |
| Real pages, final test part 1 | final test | 127 | 60 | 8,471 | 198,989 |
| Real pages, final test part 2 | final test | 645 | 320 | 44,782 | 1,034,967 |

Real pages are README files and documentation pages of public repositories. For four
languages and two creation bands, repositories were taken in order of stars. A
repository named in any earlier research material was skipped. The final test was
selected before any phrase list existed and was opened once, for the lists in this record.
An item is a sentence, a paragraph without separate sentences, or a fragment of six or
more words, with byte ranges from the engine's own extraction.

## Reviewers

| Role | Model | Access |
| --- | --- | --- |
| Annotator and reviewer A | `gpt-6-astra`, medium effort | command-line client, tools disabled |
| Annotator and reviewer B | `claude-sonnet-5-5` | command-line client, tools disabled |
| Auditor | `private-host/Qwen3.8-27B` | chat completion endpoint |

## Two questions, two answers

Annotators read each complete document and marked the items that **need revision**
under rubric `unswell-editorial-v1`. Reviewers read each engine finding and judged
whether it is **justified under its rule's statement**. The two questions give
different numbers. An annotator asked for required changes leaves most sentences
with a word such as `comprehensive` unmarked; a reviewer asked about the rule accepts
most findings on it. Rule defaults are judged by the second question. Coverage is
reported against the first.

On the long-prose study the annotators marked 3,388 of 40,496 items;
both marked 941. On the final test both marked 620 of 53,253 items.
Cohen's kappa on the long-prose study is 0.43. Sentence-level necessity is a judgment
on which two strong models often differ, so every recall figure below is a share of
their agreement, not of a ground truth.

## Development: every rule enabled

A configuration enabled all 58 rules and extended the phrase lists with 321 candidate
words and phrases. Both reviewers judged each finding, at most eight per rule and
document. They agreed on yes, no, or uncertain for 79.0% of 7,306 findings.

| Rule | Reviewed | Both accept | At least one accepts | Both reject |
| --- | ---: | ---: | ---: | ---: |
| `filler.wordy-phrase` | 116 | 100.0% | 100.0% | 0.0% |
| `filler.announced-importance` | 58 | 96.6% | 100.0% | 0.0% |
| `filler.nominal-support` | 28 | 96.4% | 100.0% | 0.0% |
| `filler.evaluative-closure` | 21 | 95.2% | 100.0% | 0.0% |
| `filler.unscoped-assurance` | 88 | 68.2% | 87.5% | 12.5% |
| `filler.weak-intensifiers` | 22 | 68.2% | 90.9% | 9.1% |
| `hype.vague-praise` (praise candidates) | 738 | 65.9% | 88.5% | 11.2% |
| `hype.vague-praise` (nouns candidates) | 48 | 64.6% | 83.3% | 16.7% |
| `filler.document-metadiscourse` | 100 | 63.0% | 87.0% | 13.0% |
| `syntax.long-sentence` | 1,065 | 60.2% | 85.9% | 14.0% |
| `filler.instruction-scaffolding` | 169 | 55.6% | 84.6% | 15.4% |
| `repetition.adjacent-word` | 19 | 52.6% | 57.9% | 42.1% |
| `policy.banned-phrases` (framing candidates) | 600 | 44.8% | 72.0% | 28.0% |
| `policy.banned-phrases` (verbs candidates) | 700 | 41.3% | 54.4% | 45.3% |
| `policy.banned-phrases` | 36 | 36.1% | 80.6% | 19.4% |
| `repetition.ngram-density` | 28 | 17.9% | 25.0% | 75.0% |
| `syntax.noun-stack` | 57 | 10.5% | 29.8% | 70.2% |
| `readability.long-paragraph` | 33 | 6.1% | 100.0% | 0.0% |
| `repetition.sentence-openers` | 77 | 5.2% | 27.3% | 72.7% |
| `policy.banned-phrases` (transitions candidates) | 681 | 5.1% | 23.9% | 76.1% |
| `readability.grade-metric` | 903 | 3.7% | 27.1% | 70.7% |
| `syntax.passive-candidate-density` | 1,408 | 2.2% | 19.1% | 80.9% |
| `syntax.parenthetical-load` | 121 | 1.7% | 23.1% | 76.9% |
| `syntax.paired-contrast-density` | 47 | 0.0% | 0.0% | 100.0% |
| `repetition.paragraph-openers` | 39 | 0.0% | 17.9% | 82.1% |
| `format.em-dash-density` | 35 | 0.0% | 5.7% | 94.3% |

## Decisions

A phrase is listed when both reviewers accepted at least 80% of at least five reviewed
findings of its word family, and the phrase itself did not fall below 60% on three or
more findings. Inflections of one word form a family. A word counts toward a cluster
when at least one reviewer accepted 85% of its findings and both rejected at most 15%.
A rule that ships enabled becomes opt-in when both reviewers rejected at least 60% of
15 or more findings.

| Change | Content |
| --- | --- |
| `hype.vague-praise` ships enabled, list extended | `beautiful`, `beautifully`, `comprehensive documentation`, `comprehensive guide`, `easy to use`, `excellent`, `exceptional`, `exceptionally`, `intuitive`, `powerful`, `seamless`, `seamless integration`, `seamlessly`, `sophisticated`, `state-of-the-art`, `unprecedented` |
| `filler.wordy-phrase` list extended | `by adhering to`, `empower`, `empowering`, `empowers`, `facilitate`, `facilitates`, `facilitating`, `first and foremost`, `foster`, `fostering`, `fosters`, `leverage`, `leveraged`, `leverages`, `leveraging`, `the bottom line`, `utilize`, `utilizes`, `utilizing` |
| `hype.modifier-cluster` list extended | `beautiful`, `comprehensive`, `crucial`, `cutting-edge`, `delightful`, `easy-to-use`, `enterprise-grade`, `excellent`, `extensive`, `feature-rich`, `intuitive`, `powerful`, `robust`, `seamless`, `seamlessly`, `significantly`, `sophisticated`, `state-of-the-art`, `straightforward`, `streamlined`, `unprecedented`, `user-friendly` |
| `repetition.paragraph-openers` becomes opt-in | both reviewers rejected 32 of 39 findings and accepted 0 |
| `repetition.sentence-openers` becomes opt-in | both reviewers rejected 56 of 77 findings and accepted 4 |

On the development sets the listed phrases have 415 reviewed findings, and both reviewers
accepted 94.2%. That figure is optimistic, because the same findings selected the phrases.
The final test below is the unbiased measurement.

Candidates that failed include stock connectives and summary openers (both reviewers
accepted about 5%), `harness` (a technical term in every reviewed case), `ensure` and
its forms, and `this allows`. Surface measurements stayed opt-in: both reviewers
rejected most findings of the passive, readability, parenthetical, contrast, and
em-dash measurements.

## Final test

The engine ran with the technical profile on the 772 final pages, before and after the change.

| Measure | Before | After |
| --- | ---: | ---: |
| Findings | 1,489 | 1,617 |
| Findings other than `syntax.long-sentence` | 265 | 393 |
| Findings per 1,000 prose words | 1.67 | 1.82 |
| Items both annotators marked that a finding touches | 50 of 620 (8.1%) | 95 of 620 (15.3%) |
| The same without `syntax.long-sentence` | 20 (3.2%) | 68 (11.0%) |

Both reviewers then judged the findings of the changed engine on the final test.

| Rule | Reviewed | Both accept | At least one accepts | Both reject |
| --- | ---: | ---: | ---: | ---: |
| `filler.wordy-phrase` | 136 | 86.0% | 86.0% | 14.0% |
| `hype.vague-praise` | 64 | 82.8% | 95.3% | 4.7% |
| `filler.evaluative-closure` | 5 | 80.0% | 100.0% | 0.0% |
| `syntax.long-sentence` | 1,073 | 61.5% | 86.5% | 13.4% |
| `repetition.adjacent-word` | 11 | 54.5% | 54.5% | 45.5% |
| `filler.document-metadiscourse` | 53 | 37.7% | 77.4% | 22.6% |
| `filler.instruction-scaffolding` | 79 | 35.4% | 83.5% | 16.5% |
| `filler.unscoped-assurance` | 18 | 22.2% | 50.0% | 50.0% |

`filler.wordy-phrase` meets the 85% target of issue 26: both reviewers accepted 117 of 136 findings
(95% interval 79% to 91%). `hype.vague-praise` is just below it: 53 of 64
(interval 72% to 90%), with at least one reviewer accepting 95.3% and both rejecting 4.7%.
It ships enabled as a warning without a gate, and the point estimate is stated here.

The same table shows that several rules that already shipped enabled do worse on real
pages than on the development sets: `filler.instruction-scaffolding`,
`filler.document-metadiscourse`, and `filler.unscoped-assurance`. The final test reports;
it does not select. Their defaults are unchanged here and are the next subject of issue 26.
`syntax.long-sentence` produces three of four findings on real pages, and both reviewers
accept about six in ten.

### Third-assistant audit

The auditor compared each passage with a version in which the flagged phrase was
removed, or replaced by a plain word for a stock verb. The order was random. Placebo
items removed a random ordinary word elsewhere. An item is confirmed when the auditor
does not prefer the original and reports no lost information.

| Group | Audited | Confirmed | No information lost |
| --- | ---: | ---: | ---: |
| Findings | 209 | 67.9% | 90.9% |
| Placebo | 69 | 8.7% | 42.0% |
| `filler.wordy-phrase` | 145 | 74.5% | 88.3% |
| `hype.vague-praise` | 64 | 53.1% | 96.9% |

The audit separates findings from placebo edits, so the auditor is not accepting
everything. It confirms fewer findings than the 85% that ADR 0046 set for a fitted
model. About half of its rejections come from the mechanical edit: removing one
adjective can leave `most language models` or `for , API-key-free`, and the auditor then
prefers the grammatical original while stating that the removed word carried no
information. A later audit should let an assistant write the minimal edit.

## Second round: severity by evidence

[ADR 0048](../adr/0048-severity-follows-review.md) sets default severity from the reviews above:
a warning needs both reviewers to accept at least 85% of 50 or more findings on real pages.
A new set of 589 real pages from 316 repositories, selected by the same rule and used
for nothing before, reports the result.

| Severity | Findings | Reviewed | Both accept | At least one accepts | Both reject |
| --- | ---: | ---: | ---: | ---: | ---: |
| warning | 115 | 115 | 82.6% (75% to 88%) | 87.8% | 12.2% |
| note | 1,063 | 885 | 59.9% (57% to 63%) | 86.1% | 13.9% |

| Rule | Severity | Reviewed | Both accept | At least one accepts |
| --- | --- | ---: | ---: | ---: |
| `filler.wordy-phrase` | warning | 85 | 90.6% | 95.3% |
| `repetition.adjacent-word` | warning | 9 | 33.3% | 33.3% |
| `filler.nominal-support` | warning | 7 | 100.0% | 100.0% |
| `repetition.exact-sentence` | warning | 6 | 66.7% | 66.7% |
| `syntax.long-sentence` | note | 715 | 60.4% | 87.0% |
| `filler.instruction-scaffolding` | note | 70 | 51.4% | 88.6% |
| `hype.vague-praise` | note | 54 | 81.5% | 90.7% |
| `filler.document-metadiscourse` | note | 37 | 43.2% | 70.3% |
| `filler.unscoped-assurance` | note | 9 | 22.2% | 33.3% |

Findings are capped at eight per rule and page for review, so a page with many long
sentences contributes eight of them.
The third assistant, `private-host/Qwen3.8-27B`, judged the same findings as a check. It gave the
same answer as the second reviewer for 83.1% of 1,000 findings. It accepted
75.4% of them and the second reviewer 83.9%.

## Third round: thresholds, matcher forms, and two rules

[ADR 0049](../adr/0049-defaults-follow-reviewed-forms.md) reads the reviews inside a rule.
Split by sentence length, the long-sentence findings on real pages show where the rule holds.

| Sentence length in prose words | Reviewed | Both accept |
| --- | ---: | ---: |
| 36 to 39 | 1,051 | 47.6% |
| 40 to 44 | 760 | 64.1% |
| 45 to 49 | 356 | 75.8% |
| 50 to 54 | 192 | 83.3% |
| 55 to 59 | 96 | 86.5% |
| 60 to 69 | 113 | 74.3% |
| 70 or more | 62 | 66.1% |

In table rows both reviewers accepted 45.0% of 109 findings,
against 62.5% of 2,521 elsewhere.
In `filler.instruction-scaffolding`, the conditional reader goal and the `Be sure to` reminder gave
138 reviewed findings, and both reviewers accepted 18.1% of them.
In `repetition.adjacent-word`, both reviewers accepted none of the 15 findings whose two
copies differ in case, repeat a capitalized word inside a sentence, or stand on two sides of Markdown markup.
They accepted 45.1% of 133 `filler.document-metadiscourse` findings and
29.2% of 48 `filler.unscoped-assurance` findings. Both rules are now opt-in.

The next table compares the earlier defaults with the changed ones on the 2,118 real pages
reviewed before this round. A finding keeps its verdict when its page, rule, and byte span are unchanged.

| Rule | Reviewed before | Both accept | Reviewed after | Both accept |
| --- | ---: | ---: | ---: | ---: |
| `syntax.long-sentence` | 2,630 | 61.8% | 733 | 78.7% |
| `filler.wordy-phrase` | 284 | 90.5% | 284 | 90.5% |
| `filler.instruction-scaffolding` | 234 | 42.7% | 96 | 78.1% |
| `hype.vague-praise` | 211 | 82.5% | 211 | 82.5% |
| `filler.document-metadiscourse` | 133 | 45.1% | 0 | not reported |
| `filler.unscoped-assurance` | 48 | 29.2% | 0 | not reported |
| `repetition.adjacent-word` | 32 | 46.9% | 17 | 88.2% |
| `filler.nominal-support` | 24 | 95.8% | 24 | 95.8% |
| `filler.evaluative-closure` | 13 | 92.3% | 13 | 92.3% |
| `repetition.exact-sentence` | 11 | 45.5% | 11 | 45.5% |
| all notes | 3,256 | 60.6% | 1,057 | 79.6% |
| all warnings | 380 | 83.9% | 348 | 87.4% |
| all findings | 3,636 | 63.0% | 1,405 | 81.5% |

The default profile reports 1,733 findings on these pages in place of 4,219.
These pages chose the changes, so the right columns overstate the effect.

A third set of 578 new real pages from 319 repositories, selected by the same rule and used
for nothing before, reports the result. Both reviewers judged the findings of the earlier
defaults and of the changed ones on the same pages.

| Rule | Reviewed before | Both accept | Reviewed after | Both accept |
| --- | ---: | ---: | ---: | ---: |
| `syntax.long-sentence` | 832 | 60.8% | 276 | 67.0% |
| `filler.wordy-phrase` | 96 | 96.9% | 96 | 96.9% |
| `filler.instruction-scaffolding` | 59 | 57.6% | 26 | 88.5% |
| `hype.vague-praise` | 55 | 85.5% | 55 | 85.5% |
| `filler.document-metadiscourse` | 36 | 36.1% | 0 | not reported |
| `filler.unscoped-assurance` | 17 | 29.4% | 0 | not reported |
| `filler.nominal-support` | 11 | 90.9% | 11 | 90.9% |
| `repetition.adjacent-word` | 11 | 63.6% | 7 | 100.0% |
| `repetition.exact-sentence` | 10 | 80.0% | 10 | 80.0% |
| `filler.evaluative-closure` | 5 | 100.0% | 5 | 100.0% |
| all notes | 999 | 60.6% | 364 | 72.0% |
| all warnings | 141 | 90.1% | 130 | 92.3% |
| all findings | 1,140 | 64.2% | 494 | 77.3% |

| Severity, changed defaults | Findings | Reviewed | Both accept | At least one accepts | Both reject |
| --- | ---: | ---: | ---: | ---: | ---: |
| warning | 130 | 130 | 92.3% (86% to 96%) | 95.4% | 4.6% |
| note | 387 | 364 | 72.0% (67% to 76%) | 85.4% | 14.6% |

The changed defaults report 517 findings on these pages in place of 1,174,
or 0.61 per thousand words in place of 1.38.
Both reviewers accepted 54.2% of the 646 dropped findings they reviewed.
The gain is smaller than on the pages that chose the changes, as expected. `syntax.long-sentence`
holds two findings in three above 45 words, and one in two above 70, where the sentence is often a list.

The third assistant, `private-host/Qwen3.8-27B`, judged findings of this round as a check. It gave the
same answer as the second reviewer for 90.8% of 546 findings. It accepted
80.0% of them and the second reviewer 83.3%.

## Review in context through the MCP server

[ADR 0047](../adr/0047-qualified-defaults-and-assistant-review.md) lets the calling assistant review a
document in context: `unswell_review_request` returns the rubric and the items, and `unswell_review_submit`
binds each flag to source bytes or rejects it. A client drove the server with two calling assistants on
pages of final test 1. It passed the server's instructions and items to the assistant unchanged and
submitted the answer unchanged. The assistant of the other family then judged each bound flag under the
statement of its category, the way it judges a rule finding. A flag that it accepts has two models behind
it, the one that raised it and the one that judged it.

| Calling assistant | Pages | Flags | Bound by the server | Per 1,000 words | Reviewer | Reviewed | Accepted | Rejected |
| --- | ---: | ---: | ---: | ---: | --- | ---: | ---: | ---: |
| `claude-sonnet-5-5` | 120 | 351 | 331 | 1.77 | `gpt-6-astra` | 329 | 62.9% | 37.1% |
| `gpt-6-astra` | 120 | 378 | 359 | 1.92 | `claude-sonnet-5-5` | 356 | 76.7% | 23.0% |

Of the 690 bound flags, 675 lie outside every finding of the default rules on the same page.
On these pages the other annotator had marked 472 items for revision. The flags touch 224 of them
(47.5%); the default rules touch 20 (4.2%).
The third assistant, `private-host/Qwen3.8-27B`, judged 80 of the flags. It gave the same answer as the
reviewer for 81.2% and accepted 60.0%.

| Category | Reviewed | Accepted |
| --- | ---: | ---: |
| `unjustified_intensifiers` | 263 | 80.6% |
| `vague_claims` | 226 | 70.8% |
| `needless_repetition` | 143 | 48.3% |
| `wordiness` | 99 | 92.9% |
| `empty_framing` | 52 | 75.0% |
| `needless_complexity` | 8 | 62.5% |
| `formulaic_transitions` | 7 | 42.9% |

Review in context reports about three times as many items as the default rules and reaches items that no
rule names. Its flags are accepted less often than rule findings, and repetition is the weakest category:
whether a sentence repeats an earlier one is where two models differ most. These flags are judgments of
the calling assistant. They carry its label, stay advisory, and reach neither scores nor the gate. The two
assistants were also the annotators of these pages, under a rubric with the same categories, so the
coverage figure measures agreement between two models and not recall against an editor.

## Phrase lists read again on all opened pages

ADR 0047 lists a phrase when both reviewers accept at least 80% of five or more reviewed findings of
its word family. A test set becomes development data once it is opened, so the criterion was read again
over every reviewed finding of a listed phrase, on generated documents and on real pages.

| Rule | Word family | Reviewed | Both accept | On real pages | Decision |
| --- | --- | ---: | ---: | ---: | --- |
| `filler.wordy-phrase` | `in order to` | 269 | 99.6% | 99.5% of 218 | keep |
| `filler.wordy-phrase` | `leverage` | 155 | 87.1% | 82.9% of 117 | keep |
| `filler.wordy-phrase` | `utilize` | 108 | 98.1% | 98.0% of 50 | keep |
| `filler.wordy-phrase` | `facilitate` | 62 | 91.9% | 86.5% of 37 | keep |
| `filler.wordy-phrase` | `empower` | 18 | 100.0% | 100.0% of 11 | keep |
| `filler.wordy-phrase` | `foster` | 10 | 50.0% | 44.4% of 9 | drop |
| `filler.wordy-phrase` | `first and foremost` | 7 | 85.7% | 50.0% of 2 | keep |
| `filler.wordy-phrase` | `by adhering to` | 5 | 80.0% | none | keep |
| `filler.wordy-phrase` | `the bottom line` | 5 | 80.0% | none | keep |
| `filler.wordy-phrase` | `due to the fact that` | 3 | 100.0% | 100.0% of 1 | keep |
| `hype.vague-praise` | `powerful` | 108 | 86.1% | 83.0% of 88 | keep |
| `hype.vague-praise` | `seamless` | 47 | 97.9% | 96.7% of 30 | keep |
| `hype.vague-praise` | `excellent` | 38 | 81.6% | 74.1% of 27 | keep |
| `hype.vague-praise` | `intuitive` | 32 | 93.8% | 92.6% of 27 | keep |
| `hype.vague-praise` | `state-of-the-art` | 26 | 57.7% | 57.7% of 26 | drop |
| `hype.vague-praise` | `beautiful` | 20 | 95.0% | 94.4% of 18 | keep |
| `hype.vague-praise` | `easy to use` | 17 | 82.4% | 82.4% of 17 | keep |
| `hype.vague-praise` | `comprehensive documentation` | 15 | 86.7% | 90.0% of 10 | keep |
| `hype.vague-praise` | `sophisticated` | 14 | 92.9% | 90.0% of 10 | keep |
| `hype.vague-praise` | `exceptional` | 9 | 88.9% | 87.5% of 8 | keep |
| `hype.vague-praise` | `unprecedented` | 7 | 100.0% | 100.0% of 2 | keep |
| `hype.vague-praise` | `comprehensive guide` | 6 | 66.7% | 33.3% of 3 | drop |

The lists lose `comprehensive guide`, `foster`, `fostering`, `fosters`, `state-of-the-art`. On real pages both reviewers accepted 83.1% of 266
`hype.vague-praise` findings; without the dropped families the share is 86.5% of 237.
The same pages chose what to drop, so that share is optimistic, and the rule keeps its note until
new pages confirm it. Findings on generated documents are accepted more often than on real pages,
which is why a family can pass on all pages and sit below 80% on real ones.

## Fourth test: the same defaults on new pages

A fourth set of 606 new real pages from 318 repositories was selected by the same rule. The
decisions that could follow from it were written in issue 525 before any page was read. The defaults
under test are those of the third round with the phrase lists read again.

| Severity | Findings | Reviewed | Both accept | At least one accepts | Both reject | Both accept, third test |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| warning | 133 | 133 | 85.7% (79% to 91%) | 88.0% | 12.0% | 92.3% |
| note | 288 | 280 | 76.4% (71% to 81%) | 93.2% | 6.8% | 72.0% |
| all | 421 | 413 | 79.4% | 91.5% | 8.5% | 77.3% |

| Rule | Severity | Reviewed | Both accept | Both accept, third test |
| --- | --- | ---: | ---: | ---: |
| `syntax.long-sentence` | note | 190 | 81.1% | 67.0% of 276 |
| `filler.wordy-phrase` | warning | 109 | 95.4% | 96.9% of 96 |
| `hype.vague-praise` | note | 56 | 67.9% | 85.5% of 55 |
| `filler.instruction-scaffolding` | note | 25 | 60.0% | 88.5% of 26 |
| `repetition.adjacent-word` | note | 9 | 77.8% | 100.0% of 7 |
| `repetition.exact-sentence` | warning | 7 | 0.0% | 80.0% of 10 |
| `filler.nominal-support` | warning | 6 | 83.3% | 90.9% of 11 |

The defaults report 0.47 findings per thousand words on these pages.
The share over all findings holds between the two tests. Single rules move by ten to twenty points
between sets of this size, in both directions, so one set does not settle a rule. The third test ran
before the phrase lists were read again, so its `hype.vague-praise` figure covers a slightly longer list.

The decisions fixed in advance:

- `hype.vague-praise`: both reviewers accepted 67.9% of 56 findings on the new pages. It stays below the 85% of ADR 0048 and remains a note.
- `syntax.long-sentence` with eight or more commas: 57.1% of 21 on the new pages and 52.8% of 180 over all reviewed real pages, against 80.6% of 1,019 with fewer commas. That is not below half, so the rule keeps them.
- `repetition.exact-sentence`: 46.4% of 28 reviewed findings on real pages. That is still fewer than 30, so its severity is unchanged.

## Limits

- Reviewers and auditor are assistants. Models can share a preference that readers do not have.
- The phrase lists cover prose in paragraphs, comments, and strings. List items, headings, and table cells are not checked by the praise rule.
- Coverage of the annotators' agreement stays low. Most items they mark are contextual: a claim without scope, a sentence that repeats an earlier one, a frame that adds nothing. Fixed rules name a small part of that. The MCP review tools let the calling assistant judge the rest.
- The corpus tool left out 17 of 2,713 real pages that it could not parse; issue 522 tracks them. The engine now analyzes such a page and excludes the rejected block.
- The third round matches findings between defaults by page, rule, and byte span. Review stays capped at eight findings per rule and page.
