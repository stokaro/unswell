# Rule qualification on the assistant-panel corpus

This record supports the rule defaults and phrase lists that changed after
[ADR 0046](../adr/0046-assistant-panel-corpus.md) and [ADR 0047](../adr/0047-qualified-defaults-and-assistant-review.md).
Every reviewer here is an AI assistant. No number is a human judgment, and none
measures agreement with a native-speaking editor.

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

## Limits

- Reviewers and auditor are assistants. Models can share a preference that readers do not have.
- The phrase lists cover prose in paragraphs, comments, and strings. List items, headings, and table cells are not checked by the praise rule.
- Coverage of the annotators' agreement stays low. Most items they mark are contextual: a claim without scope, a sentence that repeats an earlier one, a frame that adds nothing. Fixed rules name a small part of that. The MCP review tools let the calling assistant judge the rest.
- Two real pages failed to parse and are tracked in issue 522.
