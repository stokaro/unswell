# Which findings hold up: qualifying editorial rules by assistant review

This report covers the work of November 2026 on the assistant-panel corpus. It
follows [the first report](report-v1.md), which asked which constructions mark
generated technical English. This one asks a narrower question: which findings
of a local, rule-based linter do two independent reviewers accept on real
documentation pages, and what follows for the defaults.

Every annotator, reviewer, and auditor in this report is an AI assistant. No
number is a human judgment. The maintainer is not a native speaker of English
and accepted assistant review for this work in
[ADR 0041](../docs/adr/0041-assistant-review-acceptance.md).

## Questions

1. What does the shipped engine report on real documentation pages?
2. Can a small local model learn what two assistant annotators mark?
3. Can review of findings decide which phrases and rules ship by default?
4. Does a decision made on reviewed pages hold on new pages?

## Related work

Issue [#523](https://github.com/stokaro/unswell/issues/523) lists the sources
and tracks a full reading. The ones below were read from their abstracts and
public data, and nothing in them was replicated here.

- Shaib and others, *Measuring AI "Slop" in Text* (arXiv 2509.19163, 2025),
  have human editors mark spans. They report moderate agreement between
  editors and weak performance of models as span annotators. The agreement
  figure in this report is of the same order.
- Kobak and others, *Delving into LLM-assisted writing in biomedical
  publications through excess vocabulary* (arXiv 2406.07016), measure words
  whose frequency in abstracts rose abruptly in 2023 and 2024. Their
  style-word list was tested here as a source of phrase candidates.
- Wikipedia's page *Signs of AI writing* and several open linters publish
  phrase lists. The lists overlap with the candidates tested here.

Older sources on readability and plain language motivate the surface
measurements. The first report cites them.

## Design

**Pages.** Real pages are README files and documentation pages of public
repositories. For four languages and two creation bands, repositories were
taken in order of stars. Each repository gives its README and at most two
pages under its documentation directory. A repository named in any earlier
research material was skipped, and no repository appears under two set roles.

**Two questions.** Annotators read a complete document and mark the items that
need revision under one rubric. Reviewers read each finding of the engine and
judge whether it is justified under the statement of its rule. The two
questions give different numbers, and the report keeps them apart. Defaults
are decided by the second. Coverage is reported against the first.

**Panel.** The annotators and reviewers are `gpt-6-astra` at medium effort and
`claude-sonnet-5-5`, both without tools. The auditor and third reviewer is
`private-host/Qwen3.8-27B`. A finding counts as accepted when both reviewers
say yes.

**Decisions before tests.** Each decision rule was written in an ADR before the
test set for it was opened: ADR 0047 for phrases and opt-in rules, ADR 0048 for
severity, and ADR 0049 for thresholds and matcher forms. A test set reports
and does not select. Once opened, it becomes development data for the next
round.

## The corpus

| Set | Role | Documents | Labels | Reviews |
| --- | --- | ---: | --- | --- |
| Long-prose study: 38 sources and 455 generated versions | development | 493 | two annotators | yes |
| Real pages, development 1 | development | 164 | two annotators | yes |
| Real pages, development 2 | development | 593 | none | yes |
| Real pages, test 1 | final test 1 | 772 | two annotators | yes |
| Real pages, test 2 | final test 2 | 589 | none | yes |
| Real pages, test 3 | final test 3 | 578 | none | yes |

The sets hold 3,189 documents and 10,909 reviewed findings. The
[corpus record](corpus/2026-11-11-assistant-panel/README.md) stores every label
and review without page text.

## First result: the engine was close to silent

On the 772 pages of test 1 the released engine reported 1,489 findings, and
1,224 of them were long sentences. Six textbook sentences of machine-style
praise produced no finding. The phrase rules shipped with three to seven
example phrases each, and 24 of 58 rules were opt-in because no corpus
precision had been measured. The earlier research had added contextual rules
one at a time against a frozen screen of 123 events
([#349](https://github.com/stokaro/unswell/issues/349)). That screen measured
progress on a set too small to show that the lists themselves were empty.

## Second result: a fitted model does not carry to real pages

The annotators marked 3,388 of 40,496 items in the long-prose study, and both
marked 941. Cohen's kappa is 0.43. Two strong models often differ on which
sentence needs revision, so every recall figure is a share of their agreement.

A word n-gram logistic regression was trained on the items that both
annotators marked, with repositories held out. It reached 18.7% recall at 85%
precision on generated documents. Trained on those documents, it reached no
recall at that precision on 164 real pages. Sentence embeddings did worse than
word n-grams, with an average precision of 0.19 against 0.44. The learnable
signal is lexical, and it belongs to the generated documents. ADR 0047 stopped
the model under the stop rule that ADR 0046 had set.

## Third result: review of findings qualifies phrases and rules

A configuration enabled all 58 rules and added 321 candidate words and
phrases. Both reviewers judged 7,306 findings on the development sets and gave
the same answer for 79.0%.

- A phrase joined a list when both reviewers accepted at least 80% of five or
  more findings of its word family. Sixteen praise words and phrases and
  nineteen stock verbs and phrases passed.
- Stock connectives and summary openers failed: both reviewers accepted about
  5% of their findings.
- Two opener rules became opt-in. Both reviewers rejected 56 of 77 and 32 of
  39 of their findings.
- The surface measurements stayed opt-in. Both reviewers rejected most
  findings of the passive, readability, parenthetical, contrast, and em-dash
  measurements.

On test 1 both reviewers accepted 86.0% of 136 `filler.wordy-phrase` findings
and 82.8% of 64 `hype.vague-praise` findings. Fixed rules other than the
long-sentence rule touched 11.0% of the items that both annotators had marked,
up from 3.2%.

A third assistant compared each flagged passage with an edited version, in
random order, beside placebo edits. It confirmed 67.9% of 209 findings and
8.7% of 69 placebo items. About half of its rejections came from the
mechanical edit, which can break grammar.

## Fourth result: severity is a statement about evidence

Read by rule, test 1 showed that one severity covered very different evidence.
ADR 0048 made a warning mean that both reviewers accepted at least 85% of 50
or more findings on real pages. On the 589 new pages of test 2 both reviewers
accepted 82.6% of 115 warnings and 59.9% of 885 notes.

## Fifth result: the unit of qualification is smaller than a rule

The reviews were then read inside each rule.

| Sentence length in prose words | Reviewed | Both accept |
| --- | ---: | ---: |
| 36 to 39 | 1,051 | 47.6% |
| 40 to 44 | 760 | 64.1% |
| 45 to 49 | 356 | 75.8% |
| 50 to 59 | 288 | 84.4% |
| 60 or more | 175 | 71.4% |

In `filler.instruction-scaffolding`, two forms gave 138 findings that both
reviewers accepted 18.1% of the time, and the other forms 96 findings at
78.1%. In `repetition.adjacent-word`, every rejected finding differed in case,
repeated a capitalized word, or stood on two sides of Markdown markup. Two
rules had no form that separated accepted findings from rejected ones and
became opt-in.

Test 3 measured the change on 578 new pages. Both reviewers judged the
findings of the earlier defaults and of the changed ones on the same pages.

| | Earlier defaults | Changed defaults |
| --- | ---: | ---: |
| Findings | 1,174 | 517 |
| Both accept, all reviewed findings | 64.2% | 77.3% |
| Both accept, notes | 60.6% | 72.0% |
| Both accept, warnings | 90.1% | 92.3% |

On the pages that chose the changes, the share rose from 63.0% to 81.5%. The
difference between 81.5% and 77.3% is the price of choosing on the same pages.
Both reviewers accepted 54.2% of the 646 dropped findings they reviewed: the
changes remove findings that hold about half the time.

## Negative and null results, stated

- The fitted word model has no recall on real pages at 85% precision.
- Sentence embeddings add nothing over word n-grams for this label.
- External phrase lists and excess-frequency candidates left 51 phrases after
  a lenient filter, and most repeat entries already tested. Longer lists will
  add little on this register.
- Three contextual rules that were built on generated documents do worse on
  real pages. Both reviewers accepted 39.0% of `filler.instruction-scaffolding`
  findings there against 69.0% on generated documents.
- `syntax.long-sentence` stays the weakest default: 67.0% of 276 findings on
  test 3, and half of the 38 findings above 70 words.
- The paired audit confirms fewer findings than review does, and part of the
  gap is an artifact of mechanical edits.

## Limitations

- Assistants wrote every label. Models can share a preference that readers do
  not have, and agreement between two models is not agreement with an editor.
- Reviewers saw at most eight findings per rule and page.
- Tests 2 and 3 have reviews and no item labels, so they report precision and
  no recall.
- The pages are English technical documentation from popular repositories.
  Nothing here speaks for other registers.
- The share of annotator agreement that fixed rules touch stays near 11%. Most
  marked items are contextual: a claim without scope, a sentence that repeats
  an earlier one, a frame that adds nothing.

## What changed in the product

- Qualified phrase lists for `hype.vague-praise`, `filler.wordy-phrase`, and
  `hype.modifier-cluster`.
- Default severity that follows reviewed precision, so `--min-severity warning`
  shows the confirmed findings.
- Thresholds and matcher forms that follow review inside a rule, and four
  rules moved to opt-in over the three rounds.
- Review in context by the calling assistant through the MCP server. Its
  findings are advisory and never reach scores or the gate.
- One Markdown block that the grammar rejects no longer hides its file.

## Reproduction

`make check-assistant-panel-corpus` verifies the record offline. The record's
README lists the scripts in the order they ran. A rebuild does not reproduce
the files byte for byte: repository rankings change, and the models do not
return the same answers twice.

## Conclusion

Review of findings by two assistants gives a usable signal for defaults. It
separates rules, phrases, and forms inside a rule, and its decisions hold on
new pages with a measured loss. It does not replace a human editor, and it
does not make a fixed rule read context. On this register the fixed rules name
a small, confirmed part of what the annotators mark. The rest needs a reader,
and the product now hands that part to the assistant that calls it.
