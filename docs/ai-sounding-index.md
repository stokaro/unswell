# AI-sounding index and `unswell fix`

`unswell check` gives every Markdown page of 150 prose words or more an index
from 0 to 100. The index is the higher of two readings of the page. The first
counts the marks that people remove when they clean AI-sounding wording out of
documentation. The second reads the rhythm of the sentences: runs of
sentences that look alike, which is what remains in a page that a model wrote
under a style guide. In effect it measures how many sentences of the page are
short plain statements. The report lists where each mark is. `unswell fix`
removes the marks that have a mechanical replacement.

The index is not a probability, and it does not say who wrote a page. Both
readings are counted from the text, so a page gets the same index whoever
wrote it. A page that a person wrote can be high, and a page that a model
wrote can be low.

## What the index reads

The first reading, the marks:

| Signal | A mark is |
| --- | --- |
| `em-dash` | an em dash in prose |
| `emoji-in-heading` | an emoji in a heading |
| `emoji-at-list-item` | a list item that starts with an emoji |
| `arrow` | the arrow `→` or `⇒` in prose |
| `stock-word` | a word of a fixed list of praise and stock verbs, such as `seamless`, `robust`, or `leverage` |
| `bold-label-in-list` | a list item that starts with a bold label |

The second reading, the rhythm:

| Signal | A mark is |
| --- | --- |
| `plain-sentence-run` | two or more plain statements in a row: five to twenty words each, with no question mark, exclamation mark, parenthesis, colon, semicolon, or em dash |
| `subject-opener-run` | three or more sentences in a row whose first word is an article, a demonstrative, or a personal pronoun, such as The, This, It, or You |

Code, URLs, front matter, and quoted blocks are outside the count. A run is
counted over the sentences of running paragraphs in the order of the page. A
heading, a list item, a table, a quote, and a code block end a run; a blank
line between two paragraphs does not.

Each count is taken per 1,000 prose words; for a run signal it is the
sentences in its runs. A reading is 100 times the logistic function of a
constant plus the weighted natural logarithms of one plus each rate, as a
whole number. The index is the higher reading.
[ADR 0059](adr/0059-page-index-and-fix.md) lists the weights of the first
reading and [ADR 0060](adr/0060-rhythm-reading.md) those of the second.

| Level | Index |
| --- | --- |
| `low` | below 40 |
| `elevated` | 40 to 69 |
| `high` | 70 or more |

The index has no gate and does not change the exit code. A file that is not
Markdown gets no index, and neither does a page with fewer than 150 prose
words.

Two rules of `unswell check` report two habits of a page that the index does
not count: paragraphs that end on a denial, and one formal register from the
first line to the last. They are findings of their own and take no part in
the index; see [Habits of a page](page-habits.md).

## In the report

The lines of the index in the text report of `unswell check docs/guide.md`,
for a page with the marks of the first reading:

```text
AI-sounding index: 100 of 100 (high; marks 100, rhythm 0) in docs/guide.md: 4 em dashes, 1 emoji in a heading, 1 list item that starts with an emoji, 2 arrows, 1 stock word, 2 list items that start with a bold label.
  docs/guide.md:1:3: emoji in a heading; remove it
  docs/guide.md:3:3: stock word "seamless"; say what the thing does, or drop the word
  docs/guide.md:3:37: em dash; write a comma
  docs/guide.md:3:67: em dash; write a comma
  docs/guide.md:7:3: list item that starts with a bold label; remove the bold
  docs/guide.md:7:16: em dash; write a colon
  docs/guide.md:8:3: list item that starts with a bold label; remove the bold
  docs/guide.md:8:15: em dash; write a colon
  docs/guide.md:9:3: list item that starts with an emoji; remove it
  docs/guide.md:13:46: arrow; write "->"
  docs/guide.md:13:54: arrow; write "->"
`unswell fix --write` makes the replacements named above. A stock word needs your own words.
The index is the higher of two readings: the marks that people remove when they clean AI-sounding wording out of documentation, and runs of sentences that look alike. It is not a probability and does not say who wrote a page.
```

And for a page with none of those marks, whose sentences have one length and
one shape:

```text
AI-sounding index: 99 of 100 (high; marks 23, rhythm 99) in docs/retry.md: 4 runs of plain sentences, 4 runs of sentences that open with their subject.
  docs/retry.md:3:1: 5 plain sentences in a row; say it as you would to a colleague: join two of them, use a contraction, put an aside in parentheses, or start one with And or But
  docs/retry.md:3:1: 5 sentences in a row open with their subject; open one of them with its condition, its time, or its reason
  docs/retry.md:7:1: 5 plain sentences in a row; say it as you would to a colleague: join two of them, use a contraction, put an aside in parentheses, or start one with And or But
  docs/retry.md:7:1: 5 sentences in a row open with their subject; open one of them with its condition, its time, or its reason
A run of sentences has no replacement that a command can make. Rewrite one or two of its sentences, and keep what they say.
Do not add an em dash anywhere: an em dash is one of the marks that this report counts.
The index is the higher of two readings: the marks that people remove when they clean AI-sounding wording out of documentation, and runs of sentences that look alike. It is not a probability and does not say who wrote a page.
```

The text and Markdown reports name each page above `low`, the most marked
first, with its two readings. They list the place of up to 60 marks in a run,
and each line says what to do with its mark: the replacement that
`unswell fix` would make, the remedy of a run of sentences, or, for a stock
word, that only you can say what the thing does.

A report lists the marks of a reading only when that reading is above `low`.
A few plain statements in a row are on nearly every page, and one em dash on a
long page is ordinary. The index line names runs of sentences under the same
condition.

The JSON report holds the index of every page under
`documents[].ai_sounding`:

| Field | Content |
| --- | --- |
| `status` | `available`, or `too_short` for a page below 150 prose words |
| `index`, `level` | the index and its level |
| `readings` | the two readings, `marks` and `rhythm`, that the index is the higher of |
| `prose_words` | the words that the rates are taken over |
| `signals` | for each signal its reading, its count, its rate per 1,000 words, and the advice; a run signal counts runs and gives the `sentences` in them, and its rate is of those sentences |
| `marks` | up to 200 marks in source order: the signal, the byte span, the line and column, the word of a `stock-word` mark, the `sentences` of a run, and the `fix` where a mechanical replacement exists |

`marks` holds the marks of both readings, also of a reading at `low`. In Go,
`AISounding.Listed` returns the ones that a report lists.

The MCP tool `unswell_check` returns the same result, so an assistant gets the
places, the advice, and the fixes for a draft it checks.

## `unswell fix`

```sh
unswell fix README.md docs
unswell fix README.md docs --write
```

```text
docs/guide.md: AI-sounding index 100 -> 48; removed 4 em dashes, 1 emoji in a heading, 1 emoji at the start of a list item, 2 arrows, 2 bold labels. Left for you: 1 stock word.
10 marks removed in 1 page of 1 document.
Nothing was written. Use --write to apply these replacements.
A replacement is mechanical. Read each one: a comma or a period is not always what the sentence needs, and a stock word needs your own words.
```

| Mark | Replacement |
| --- | --- |
| Em dash | A comma. A period and a capital letter between two full clauses. A colon after the short label of a list item. Nothing after other punctuation or at the start of a line. A dash between two numbers stays |
| Emoji in a heading or at the start of a list item | Removed |
| Arrow | `->`, or `=>` for `⇒` |
| Bold label at the start of a list item | The label without its asterisks |
| Stock word | None. It is left for you, because only you can say what the thing does |
| Run of sentences | None. It is left for you: it needs a rewrite of one or two sentences |

The command calls no model, and no text leaves the machine. It analyzes the
page again after each pass, until nothing is left to replace, and reports the
index before and after.

A page whose first reading is `low` is left as it is, also when its index is
high by the rhythm of its sentences. `unswell check` lists none of the marks
with a replacement there, and every measurement of the command below is on
pages whose marks are at 40 or more. `--include-low` replaces the marks of
such pages as well. Files change only with `--write`. The command keeps
the mode of a file, and it does not write a read-only file or a file that
changed during the run.

`--report text:path` and `--report json:path` write the report; the JSON
report has the schema version `unswell-fix-v1`. The exit code is 0 when every
selected document was handled and 2 when one failed or the run could not
start. A document without an index is counted and left as it is.

Read the change before you keep it. A comma or a period is not always what
the sentence needs, and the command cannot always tell two full clauses from
one.

## From Go

`Engine.Analyze` returns the index in `DocumentResult.AISounding`. Each
`AIMark` carries its `Fix`, and `unswell.RemoveAIMarks(source, marks)` applies
the fixes and returns the new source with the number applied. Analyze the
result again to see what is left: a page lists at most 200 marks, and a fix
that overlaps an earlier one waits for the next pass.

## Measured behavior

The two readings were measured apart: the first on pages that people later
cleaned, the second on text that a model wrote under a style guide.

### The marks

189 held-out pairs: one English Markdown file of 150 words or more, before
and after a commit whose message says that it removes AI-sounding wording,
from repositories that had no part in choosing a rule or fitting the index.
566 README files as they were before November 2022 stand for pages that
people wrote.

| Measure | Result |
| --- | --- |
| AUC, pages before such a commit against the README files | 0.867 |
| Index lower after the commit, among the pairs in which it differs | 82 of 105, 78.1% (69.3% to 84.9%) |
| The same on short one-file edits through GitHub's web flow, which a person most likely typed | 31 of 44, 70.5% (55.8% to 81.8%) |
| Pairs in which the index does not change | 84 of 189 |
| README files at level `high` | 18 of 566, 3.2% |
| README files at level `elevated` | 69 of 566, 12.2% |
| Pages at `high` before the commit, and after it | 58.2%, 43.4% |

These are the numbers of the first reading alone, which was the whole index
when they were measured. With both readings the AUC on the same pages is
0.874, the index is lower after the commit in 85 of 113 pairs that differ
(75.2%), and 22 of the 566 README files are `high` (3.9%).

83 of the held-out pages: index 40 or more, 150 to 1,500 words, one per
repository. Model judges compared two versions of a page without knowing
which was which, and named the one that reads more like text written by a
language model. Two judges of one model family form the panel; a page counts
for a version when the two together name the other one more often. A judge of
another family, `gpt-6-astra`, was added afterwards and read part of the
comparisons alone. On 80 held-out pairs before and after a de-slop edit by
people, the three judges name the version before the edit in 76.7%, 70.9%, 79.4%
of their decisive answers; 70% admits a judge. On the 38 of those pairs that are
short one-file edits through GitHub's web flow, which a person most likely
typed, the shares are 74.2%, 78.4%, 89.7%.

After `unswell fix` the index is lower on 83 of the 83 pages, and 0 are still
`high` where 65 were. After a model resolves what the report of
`unswell check` points at, it is lower on 82, and 4 are still `high`.

| Version | Against | Panel: the less AI-sounding one on, of the decisive pages | Panel: pages without a verdict | Third judge alone, of its decisive answers | Third judge: pages it calls the same |
| --- | --- | --- | ---: | --- | ---: |
| After `unswell fix` | the original | 72 of 73, 98.6% (92.6% to 99.8%) | 10 | 43 of 43 | 40 |
| Fixed by a model with unswell's report | the original | 64 of 68, 94.1% (85.8% to 97.7%) | 15 | 51 of 51 | 31 |
| Fixed by the same model with write-good's findings | the original | 45 of 71, 63.4% (51.8% to 73.6%) | 12 | 21 of 35 | 42 |
| Fixed by the same model with Vale's Google findings | the original | 21 of 55, 38.2% (26.5% to 51.4%) | 28 | not read |  |
| Fixed by the same model with Vale's Microsoft findings | the original | 16 of 68, 23.5% (15.0% to 34.9%) | 15 | not read |  |
| Fixed with unswell's report | fixed with write-good's findings | 48 of 70, 68.6% (57.0% to 78.2%) | 13 | 56 of 59 | 22 |
| Fixed with unswell's report | fixed with Vale's Google findings | 52 of 67, 77.6% (66.3% to 85.9%) | 16 | not read |  |
| Fixed with unswell's report | fixed with Vale's Microsoft findings | 56 of 68, 82.4% (71.6% to 89.6%) | 15 | not read |  |
| After `unswell fix` | fixed with write-good's findings | 45 of 72, 62.5% (51.0% to 72.8%) | 11 | not read |  |
| After `unswell fix` | fixed with Vale's Google findings | 51 of 60, 85.0% (73.9% to 91.9%) | 23 | not read |  |
| After `unswell fix` | fixed with Vale's Microsoft findings | 51 of 67, 76.1% (64.7% to 84.7%) | 16 | not read |  |

Of the 83 pages, the third judge calls the original and the page after `unswell fix` the same on 40, and the original and the page fixed with unswell's report the same on 31. Where it decides, it never names the unswell version as the more AI-sounding one against the original.

write-good is the best of the three third-party linters by this measure: its
page does best against the original. With Vale's two styles the same model
makes pages that the panel finds more AI-sounding than the original.

`unswell fix` changed after the registered run: it wrote a comma between two
full clauses and now writes a period there. The registered bars were read on
the earlier version (72 of 73 decisive pages against the original). The rows
of `unswell fix` above are a second reading of the same pages with the
version that ships.

### Two more runs on new pages

`unswell fix` changed after the first run, and its numbers above are a second
reading of pages that had been read once. Two more runs used pages that no
earlier run had read, from a new search of commits whose message states the
removal of AI-sounding wording, without any repository of an earlier set and
selected the same way. The bars were written on issue 525 before each run.
The second run used the published `v0.1.0-alpha.10` binary on 89 pages. The
third used the build of `main` that became `v0.1.0-alpha.11` on 250 pages; that
build leaves pages at the level `low` alone and otherwise writes the same
bytes. The published binary of `v0.1.0-alpha.12` writes, byte for byte, every
page after `unswell fix` that the judges read in the four runs, and prints the
same report that the fixer model was given in the fourth.

| Measure, for the panel | First run, 83 pages | Second run, 89 new pages | Third run, 250 new pages |
| --- | --- | --- | --- |
| Index lower after `unswell fix` | 83 of 83 | 85 of 89 | 247 of 250 |
| Page after `unswell fix`, against the original | 72 of 73, 98.6% (92.6% to 99.8%) | 67 of 68, 98.5% (92.1% to 99.7%) | 231 of 233, 99.1% (96.9% to 99.8%) |
| Page after `unswell fix`, against the page fixed with write-good's findings | 45 of 72, 62.5% (51.0% to 72.8%); no bar | 47 of 72, 65.3% (53.8% to 75.2%); no bar | 146 of 209, 69.9% (63.3% to 75.7%); **bar holds** |
| Page fixed by a model with unswell's report, against the original | 64 of 68, 94.1% (85.8% to 97.7%) | 60 of 72, 83.3% (73.1% to 90.2%) | not run |
| Page fixed by a model with unswell's report, against the page fixed with write-good's findings | 48 of 70, 68.6% (57.0% to 78.2%); bar holds | 39 of 72, 54.2% (42.7% to 65.2%); **bar does not hold** | not run |

Every bar against the original holds in every run. Against the best
third-party linter the result differs by version. For the command it holds
on the third run, where it had a bar, and the two earlier readings without a
bar point the same way. For a model that follows unswell's report it held in
the first run and did not in the second.

A fourth run tested the fix by the report again, after the report began to
name the replacement at each mark. On 131 more pages that nothing had fixed or
judged, with the bars written first: the index is lower after the fix by the
report on 131 of 131 pages; the panel finds the report version the less
AI-sounding one against the original on 111 of 113, 98.2% (93.8% to 99.5%),
and against the page fixed with write-good's findings on
73 of 108, 67.6% (58.3% to 75.7%). Both bars hold. All three judges
together give 87 of 122, 71.3% (62.7% to 78.6%) against
write-good's page, and the third judge alone 81 of 89 of its decisive answers.
On 61 of the 131 pages the model's page is the same as the command's. On the same
pages, without a bar, the page after `unswell fix` gives 73 of 107 against
write-good's page.

So both ways of fixing what unswell points at meet the bar against the best
third-party linter on pages that nothing had read: the command on the third
run, and a model that follows the report on the fourth. With the earlier
report, which gave the place of a mark and general advice, the bar held once
and failed once.

In the third run all three judges together find the page after `unswell fix`
the less AI-sounding one against write-good's page on 163 of 230, 70.9% (64.7% to 76.4%),
and the third judge alone in 144 of 174, 82.8% (76.5% to 87.6%) of its decisive answers,
calling the two the same on 67 pages. Against the original the third judge
names the original in 136 of 136 decisive answers and calls the two the same on
112 pages.

In the second run the command refused 1 of the 89 pages: two of its blocks do
not parse as Markdown.

The judges often give the marks that the index counts as their reasons. With
`unswell fix` applied to both versions of the pairs that admit a judge, so
that neither side has such a mark, the three judges still name the version
before the person's edit in 75.3%, 72.7%, 77.3% of their decisive answers.
They tell a cleaned page apart without the marks, and a person's cleaning
changes more than the marks. Against the page as the person left it, the
earlier version after `unswell fix` is still named the more AI-sounding one
in 59.5% and 64.1% of the panel's decisive answers. That comparison does
not admit a judge; it sets the command against a person, and the command goes
part of the way that a person's cleaning goes.

The [record](../research/reviews/2026-11-12-human-reference/RESULTS-index.md)
holds the pages, the pairs, and the answers of the judges, and its check
recounts these numbers. The bars were written on issue 525 before each run.

### The rhythm

The second reading was measured on text that a model wrote under a style
guide: 70 posts of one product blog, written by a model under the blog's style
guide and published without edits by a person, and posts that two models wrote
from the outlines of human blog posts under one of two style guides. 151 blog
posts from 11 public repositories, as they were before November 2022, stand
for posts that people wrote. A hash split every set into a fitting half and a
measured half, and the numbers below are on the measured half.

Three blind model judges tell such text from human posts. They compare the
openings of two posts and name the one that reads more like text written by a
language model:

| Comparison | The panel picks the model's text |
| --- | --- |
| Text written with no style guide, against the human post on the same subject | 154 of 154 |
| The same against a human post on another subject | 76 of 77 |
| Text written under a style guide, against the human post on the same subject | 153 of 153 |
| A post of the product blog, over its three pairs against human posts | 22 of 30, 73.3% |

The first two rows admit the panel. On the posts of the product blog the
judges disagree: one picks the post on all 30, one on 21, and one picks the
human post on 25.

The index separates such text from the 77 human blog posts:

| Set | AUC | With `v0.1.0-alpha.12` | `high`, `elevated`, `low` |
| --- | --- | --- | --- |
| Posts of the product blog, 30 | 0.870 | 0.247 | 13, 14, 3 |
| Guided texts of `gpt-6-astra`, 77 | 0.979 | 0.279 | 66, 11, 0 |
| Guided texts of `private-host/Qwen3.6-35B-A3B`, 76 | 0.988 | 0.265 | 74, 1, 1 |
| Texts of the two writers with no style guide, 154 | 0.841 and 0.861 | 0.481 and 0.776 | 32, 108, 14 |

Of 352 pages that people wrote before November 2022, the 77 blog posts and
275 README files, 14 are `high`, 4.0% (2.4% to 6.6%). The earlier release puts
10 of them at `high`; the four more are README files.

A model that revises a page by the report makes it read less like a model's.
`gpt-6-astra` revised the 179 pages of the measured half whose report lists a
mark, with the instruction to resolve what the report points at and to change
nothing else. The index is lower after the revision on 174, 97.2%. The judges
find the page more like a model than its revision on 153 of 178 decisive
pages, 86.0% (80.1% to 90.3%), and on 18 of 26 for the posts of the product
blog. Each of the three judges alone points the same way.

The reviser matters. `private-host/Qwen3.6-35B-A3B` revised the same pages by
the same report and lowered the index on 174 of 178. The judges found the
page more like a model than its revision on 124 of 177, 70.1% (62.9% to
76.3%), and on 12 of 26 for the posts of the product blog, where its revision
does not help. Its revisions add 192 em dashes, against 53 for the first
reviser.

The wording of the remedy decides this. On the fitting half a reviser that
was told only to join sentences lowered the index and made the page read more
like a model's: the judges then name stitched clauses and smoothed
transitions. [ADR 0060](adr/0060-rhythm-reading.md) has those numbers.

The [record](../research/reviews/2026-11-15-style-guide-text/RESULTS.md) holds
the pages, the answers of the judges, and the revisions, and its check
recounts these numbers. The bars were written on issue 577 before each run.
The published binary of `v0.1.0-alpha.13` gives every page of the record the
counts, the readings, and the index that the record holds, and it prints the
report that a reviser was given on all but one of the revised pages; the
[alpha.13 evidence](release/v0.1.0-alpha.13.md) has the detail.

### People who write under a style guide

A style guide asks for plain statements, and the second reading counts runs
of them. A control, run after the release, asks whether the reading rises on
any text that follows a style guide. It reads 339 pages that people wrote
before November 2022 in nine public repositories whose projects publish a
style guide for their writers. Four hold articles: `GoogleChrome/web.dev`,
`GoogleChrome/developer.chrome.com`, `GoogleCloudPlatform/community`, and
`istio/istio.io`. Five hold documentation: `gitlabhq/gitlabhq`,
`mdn/content`, `kubernetes/website`, `dotnet/docs`, and `alphagov/gds-way`.

A run forms only in paragraphs, so a page of lists and code is quiet whoever
wrote it. The bars are therefore on the 128 articles that have as much of
their prose in paragraphs as nine in ten posts of the product blog.

| Measure | Result |
| --- | --- |
| Second reading at `high`, of the 128 matched articles | 1, 0.8% (0.1% to 4.3%) |
| Index of the 30 posts of the product blog against those articles | AUC 0.861 |
| Second reading at `high`, of all 339 pages | 1 |

The reading does not rise on these people's pages. Their sentences show why:

| Pages | Words per sentence | Sentences over 20 words | Plain sentences | Plain sentences in runs | In runs by chance |
| --- | --- | --- | --- | --- | --- |
| Posts of the product blog, 30 | 12.5 | 7.9% | 68.1% | 58.3% | 52.0% |
| Matched articles of people, 128 | 17.6 | 32.4% | 47.6% | 30.6% | 26.8% |
| Matched documentation of people, 48 | 17.1 | 28.5% | 53.8% | 37.5% | 32.1% |

Each figure is the median over the pages of a group, for the sentences of
running paragraphs. The last column is what a page would have if its plain
sentences stood at random places among its sentences.

People who write under these guides write longer sentences, and about a third
of their sentences have more than 20 words. The runs follow from the share of
plain sentences: a group has 1.07 to 1.11 times as many sentences in runs as
chance gives. So the second reading is in effect a measure of how many
sentences of a page are short plain statements. The average length of a
sentence alone tells the posts of the product blog from the matched pages
with an AUC of 0.933, where the second reading gives 0.939.

The control does not show that the reading tells a model from a person who
writes in that manner. The reading is counted from the sentences, so such a
person gets the same reading. Of the 176 matched pages, 28 have as many plain
sentences as the lower quartile of the product blog's posts. Their sentences
are still longer, 15.3 words at the median, and one of the 28 is `high`.

The same [record](../research/reviews/2026-11-15-style-guide-text/RESULTS.md)
holds these pages and recounts their numbers. The figures about the sentences
of the product blog are counted on its posts; the record states them and does
not recount them. The two bars were written on issue 577 before the run.

### A revised post against human text

The measurements above compare a page with its own revision. A later one
asks the harder question: after a revision by the report, can a post of the
product blog still be told from text that a person wrote? It also measures
[`unswell compare`](compare.md), which checks one revision.

On the measured half, each of the 30 posts stands against three human blog
posts and against three articles that people wrote under a style guide. Three
blind judges name the text that reads more like a model's. A writer who
follows the product publishes a revision only when `unswell compare` accepts
it; the check accepted the revision of 2 posts for one reviser and of 3 for
the other.

| What is published | Human text | Panel picks the post | `gpt-6-astra` alone |
| --- | --- | --- | --- |
| The post as it is | human blog posts | 22 of 30 | 30 of 30 |
| The post as it is | articles under a style guide | 19 of 30 | 30 of 30 |
| Accepted revisions of `gpt-6-astra` | human blog posts | 22 of 30 | 29 of 30 |
| Accepted revisions of `gpt-6-astra` | articles under a style guide | 21 of 30 | 30 of 30 |
| Accepted revisions of `private-host/Qwen3.6-35B-A3B` | human blog posts | 22 of 30 | 29 of 30 |
| Accepted revisions of `private-host/Qwen3.6-35B-A3B` | articles under a style guide | 21 of 30 | 30 of 30 |
| Every revision of `gpt-6-astra`, without the check | human blog posts | 19 of 30 | 29 of 30 |
| Every revision of `private-host/Qwen3.6-35B-A3B`, without the check | human blog posts | 22 of 30 | 29 of 30 |

The bars were a panel rate of at most 55% and at most 65% for `gpt-6-astra`
alone, the judge that told every post before. Neither holds for any reviser
or kind of human text.

**A revision by the report lowers the index and does not make a post read as
a person's.** The median index of the 27 revised posts fell from 68 to 23
with either reviser, and one judge still tells 29 of 30 posts from a human
blog post. On the fitting half the same judge told 40 of 40 posts after a
revision by the report of `v0.1.0-alpha.13` with either reviser, after a
revision by a report with sentence lengths, and after a free rewrite that was
asked to sound like an engineer writing for colleagues.

The index cannot tell whether one revision helped. On the fitting half a
report that asked for longer sentences took the median index of 40 posts from
69 to 23, and all three judges found the revision more like a model's text
than the post, on 23, 30, and 24 of the 40. That advice is not in the report.

The [record](../research/reviews/2026-11-15-revision-check/RESULTS.md) holds
these answers and recounts the numbers, and
[ADR 0061](adr/0061-blind-check-of-a-revision.md) has the decisions. The bars
were written on issue 583 before the run.

Three later pilots on the fitting half tried what the report did not have
then: a revision that takes out the paragraphs' closing denials, which
[a rule](page-habits.md) now lists; a rewrite in the manner of two human
posts; and four rounds of revision by the judge's own remarks. The judge that
told every post before tells every post after each of them, and the three
judges together pick 25 of 38 revised posts where they picked 26. The
[record of the pilots](../research/reviews/2026-11-17-revision-pilots/README.md)
has the counts. The same judge names a model's text in 147 of 147 pairs with
a human post on the same subject, so it does not go by the subject of a post.

## Limits

- The first reading was fitted and measured on English documentation that
  someone set out to clean. It says how much a page carries the marks that
  such edits remove. The second was fitted and measured on blog posts. Neither
  says who wrote the page.
- The first reading counts six kinds of marks on the surface of a page. A
  sentence that a cleaning edit rewords is outside it, and so is most of what
  such edits change. Of the held-out pages before a commit, 41.8% are not
  `high`.
- The second reading counts sentences that look alike. In effect it measures
  how many sentences of a page are short plain statements, and a person who
  writes in that manner gets the same reading as a model. People who write
  under nine published style guides do not write in that manner: the reading
  is `high` on one of 339 of their pages. The control holds few pages of
  people who do.
- A run of sentences has no mechanical replacement. Whether a revision by the
  report helps depends on who revises: with one model it did on every kind of
  page, with another it did not on the posts of the product blog.
- A lower index after a revision does not show that the revision helped. The
  index falls after almost any edit of a run. [`unswell compare`](compare.md)
  checks one revision blind and looks for the statements and the sentences of
  the page in it.
- A revised post is still told from human text: one judge tells 29 of 30.
- Against a page fixed with a third-party linter's findings, the advantage is
  measured for `unswell fix` and for a model that follows the report as it is
  now. With the earlier report, which did not name the replacement, it held
  in one run of two.
- The fixer and the two judges of the panel are models of one family. A judge
  of another family read part of the comparisons and heard no difference
  between the original and the page after `unswell fix` on 40 of the 83 pages.
  No person took part.
- `unswell fix` replaces marks one by one and leaves stock words. It writes a
  period only where it can tell two full clauses apart; elsewhere a comma can
  still join two clauses that should be two sentences. On 400 development
  pages it wrote a period for 184 dashes and a comma for 2,553.
- A page that the engine cannot parse as Markdown fails in `unswell fix` as
  it does in `unswell check`.
