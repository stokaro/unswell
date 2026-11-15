# ADR 0060: A second reading of the page index, the rhythm of sentences

Status: accepted

## Context

The AI-sounding index of [ADR 0059](0059-page-index-and-fix.md) counts six
marks: em dashes, emoji in headings and at list items, arrows, stock words, and
bold labels in lists. Those are the marks that people remove when they clean a
model's text.

A model that writes under a style guide leaves almost none of them. On 76 posts
of one product blog, written by a model under the blog's style guide and
published without edits by a person, `v0.1.0-alpha.12` counts 2 em dashes, 5
arrows, 7 stock words, and no emoji in 81,128 prose words. It gives 71 of the
posts the level `low` and none the level `high`. On such text the six signals
even point the wrong way: human blog posts from before November 2022 carry more
of the marks than the model's posts.

Three blind model judges were asked whether such text still reads as a
model's. They compared the openings of a model's post and of a human blog post
and named the one that reads more like text written by a language model. The
record is
[text written under a style guide](../../research/reviews/2026-11-15-style-guide-text/README.md),
and its bars were written on issue 577 before each run. On the measured half:

- Text that a model wrote with no style guide is picked on 154 of 154 pairs
  about the same subject and on 76 of 77 pairs about different subjects. This
  admits the panel.
- Text that two models wrote under one of two style guides is picked on 153 of
  153 pairs.
- A post of the product blog is picked, over its three pairs, on 22 of 30
  posts, 73.3%. The judges disagree here: one picks the post on all 30, one on
  21, and one picks the human post on 25.

So there is something to find, most clearly in what a model writes from an
outline and less clearly in the blog posts.

On the fitting half the judges give one kind of reason: the sentences are
uniform. Counted with the engine's own sentences, a model's page has runs of
plain statements of five to twenty words, and runs of sentences that open with
their subject. A human blog post has 15 sentences per 1,000 prose words in runs
of two or more plain statements at the median; a post of the product blog has
44, and a guided text of one writer 57.

## Decision

The index gets a second reading, the rhythm of the sentences, from two signals:

| Signal | A mark is |
| --- | --- |
| `plain-sentence-run` | two or more plain statements in a row: five to twenty words each, with no question mark, exclamation mark, parenthesis, colon, semicolon, or em dash |
| `subject-opener-run` | three or more sentences in a row whose first word is an article, a demonstrative, or a personal pronoun |

A run is counted over the sentences of running paragraphs in the order of the
page. A heading, a list item, a table, a quote, and anything between two
paragraphs that is not prose, such as a code block, end a run. A run is one
mark; the signal weighs the sentences in its runs per 1,000 prose words.

Each reading is 100 times the logistic function of a constant plus the weighted
natural logarithms of one plus each rate. The first reading keeps the six
signals, the weights, and the constant of ADR 0059. The second has the weights
3.2 and 0.35 and the constant -12.3. **The index of a page is the higher of the
two readings.** The levels keep their thresholds, 40 and 70.

The readings are not added. The six signals point the wrong way on this text,
so in a sum they work against the new ones. On the fitting half the best sum
that kept human pages as quiet separated the product blog's posts from human
posts with an AUC of 0.78 and put 28% of them at `high`. The higher of two
readings gave 0.88 and 50%.

The weights of the second reading come from a logistic regression on the
fitting half, human blog posts and README files against the three kinds of
model text. The constant was then lowered until no human blog post and at most
0.5% of the README files of the fitting half reached `high` by rhythm.

A report lists the marks of a reading only when that reading is above the level
`low`. A few plain statements in a row are on nearly every page, and one em
dash on a long page is ordinary; neither is worth a line. `AISounding.Listed`
is that rule. `unswell fix` follows it: it leaves a page as it is when the
first reading, whose marks have a replacement, is `low`.

A run has no replacement that a command can make. The report names a remedy at
each run:

- for plain statements: say it as you would to a colleague: join two of them,
  use a contraction, put an aside in parentheses, or start one with And or But;
- for subject openers: open one of them with its condition, its time, or its
  reason.

## What was measured

All numbers are on the measured half, with the bars written first.

The index of each set against the 77 human blog posts:

| Set | AUC with this change | With `v0.1.0-alpha.12` |
| --- | --- | --- |
| Posts of the product blog, 30 | 0.870 | 0.247 |
| Guided texts of `gpt-6-astra`, 77 | 0.979 | 0.279 |
| Guided texts of `private-host/Qwen3.6-35B-A3B`, 76 | 0.988 | 0.265 |
| All of them, 183 | 0.965 | 0.268 |

13 of the 30 posts of the product blog are `high` and 14 `elevated`. On the
posts as they are published, without the preparation that the judges' texts
went through, the AUC is 0.882.

Human pages stay quiet: 14 of 352 pages written before November 2022 are
`high`, 4.0% (2.4% to 6.6%), against 10 with the earlier release. The four more
are README files; the 4 of 77 human blog posts at `high` are the same four as
before.

The earlier result stands: on the held-out pages of ADR 0059, the pages before
a de-slop commit separate from the README files with an AUC of 0.874, against
0.867, and the index is lower after the commit in 85 of 113 pairs that differ.

A revision by the report was measured with `gpt-6-astra` as the reviser on the
179 pages whose report lists a mark. The index is lower after the revision on
174, 97.2%. The panel finds the page more like a model than its revision on
153 of 178 decisive pages, 86.0% (80.1% to 90.3%), and on 18 of 26 for the
posts of the product blog. Each of the three judges alone points the same way.

Beside the bars, `private-host/Qwen3.6-35B-A3B` revised the same pages by the
same report. The index is lower on 174 of 178. The panel finds the page more
like a model than its revision on 124 of 177, 70.1% (62.9% to 76.3%), and on
12 of 26 for the posts of the product blog: there its revision does not help.
Its revisions add 192 em dashes, where the first reviser adds 53.

## What the fitting half showed about the remedy

The wording of the remedy decides whether a revision helps. A reviser that only
joins sentences lowers the index and makes the page read more like a model's:
the judges then name stitched clauses and smoothed transitions. On 24 pages of
the fitting half, each cell gives the pages where the judges find the page more
like a model, then the revision, then the ties:

| Remedy at a run of plain sentences | Reviser `gpt-6-astra` | Reviser `private-host/Qwen3.6-35B-A3B` |
| --- | --- | --- |
| join two of them, or add the condition or the reason that one leaves out | 9, 6, 9 | 6, 18, 0 |
| join two of them with and, but, or so | 5, 9, 10 | 3, 15, 5 |
| say it as you would to a colleague: join two of them, use a contraction, put an aside in parentheses, or start one with And or But | 18, 2, 4 | 13, 5, 6 |

Two judges read each pair, and three the second cell of the first row. The
last remedy ships.

## Consequences

- A page can be `high` with none of the six marks. Its report lists runs of
  sentences, and `unswell fix` leaves it as it is.
- The JSON result gains `ai_sounding.readings`, `signals[].reading`,
  `signals[].sentences`, and `marks[].sentences`. Two signals are added to
  `signals`.
- The text report names both readings in the index line, and its closing
  sentence says that the index is the higher of two readings.
- The index line names runs of sentences only when the second reading is above
  `low`.

## Limits

- The second reading was fitted on blog posts and README files. A person who
  writes short plain statements one after the other, as a careful style guide
  asks, can be `high`. The human pages of the record are from before November
  2022 and mostly follow no such guide.
- The judges are models. On the posts of the product blog they disagree, and
  with 30 posts the interval of the panel reaches down to 55.6%.
- Whether a revision by the report helps depends on the reviser. The bar was
  set and met for one model. With the other the posts of the product blog did
  not read less like a model's after the revision.
- A reviser adds em dashes, which the first reading counts. The remedy asks
  for parentheses and does not forbid the dash.
- Nobody checked that a revision keeps every fact of its page. A reviser can
  change what a sentence says.
- A model that writes with no style guide makes longer sentences. Its text
  separates from human posts too (AUC 0.841 and 0.861), with about one in five
  at `high`.
