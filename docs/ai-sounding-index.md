# AI-sounding index and `unswell fix`

`unswell check` gives every Markdown page of 150 prose words or more an index
from 0 to 100. The index counts marks that people remove when they clean
AI-sounding wording out of documentation, and the report lists where each
mark is. `unswell fix` removes the marks that have a mechanical replacement.

The index is not a probability, and it does not say who wrote a page. A page
that a person wrote can be high, and a page that a model wrote can be low.

## What the index counts

| Signal | A mark is |
| --- | --- |
| `em-dash` | an em dash in prose |
| `emoji-in-heading` | an emoji in a heading |
| `emoji-at-list-item` | a list item that starts with an emoji |
| `arrow` | the arrow `→` or `⇒` in prose |
| `stock-word` | a word of a fixed list of praise and stock verbs, such as `seamless`, `robust`, or `leverage` |
| `bold-label-in-list` | a list item that starts with a bold label |

Code, URLs, front matter, and quoted blocks are outside the count. Each count
is taken per 1,000 prose words. The index is 100 times the logistic function
of a constant plus the weighted natural logarithms of one plus each rate, as
a whole number. [ADR 0059](adr/0059-page-index-and-fix.md) lists the weights.

| Level | Index |
| --- | --- |
| `low` | below 40 |
| `elevated` | 40 to 69 |
| `high` | 70 or more |

The index has no gate and does not change the exit code. A file that is not
Markdown gets no index, and neither does a page with fewer than 150 prose
words.

## In the report

The lines of the index in the text report of `unswell check docs/guide.md`:

```text
AI-sounding index: 100 of 100 (high) in docs/guide.md: 4 em dashes, 1 emoji in a heading, 1 list item that starts with an emoji, 2 arrows, 1 stock word, 2 list items that start with a bold label.
  docs/guide.md:1:3: emoji in a heading; remove it
  docs/guide.md:3:17: stock word "seamless"; say what the thing does, or drop the word
  docs/guide.md:3:50: em dash; write a comma
  docs/guide.md:3:84: em dash; write a comma
  docs/guide.md:7:3: list item that starts with a bold label; remove the bold
  docs/guide.md:7:15: em dash; write a colon
  docs/guide.md:8:3: list item that starts with a bold label; remove the bold
  docs/guide.md:8:16: em dash; write a colon
  docs/guide.md:9:3: list item that starts with an emoji; remove it
  docs/guide.md:13:21: arrow; write "->"
  docs/guide.md:13:29: arrow; write "->"
`unswell fix --write` makes the replacements named above. A stock word needs your own words.
The index counts marks that people remove when they clean AI-sounding wording out of documentation. It is not a probability and does not say who wrote a page.
```

The text and Markdown reports name each page above `low`, the most marked
first. They list the place of up to 60 marks in a run, and each line says what
to do with its mark: the replacement that `unswell fix` would make, or, for a
stock word, that only you can say what the thing does. The JSON report holds the index of every page under
`documents[].ai_sounding`:

| Field | Content |
| --- | --- |
| `status` | `available`, or `too_short` for a page below 150 prose words |
| `index`, `level` | the index and its level |
| `prose_words` | the words that the rates are taken over |
| `signals` | for each signal its count, its rate per 1,000 words, and the advice |
| `marks` | up to 200 marks in source order: the signal, the byte span, the line and column, the word of a `stock-word` mark, and the `fix` where a mechanical replacement exists |

The MCP tool `unswell_check` returns the same result, so an assistant gets the
places, the advice, and the fixes for a draft it checks.

## `unswell fix`

```sh
unswell fix README.md docs
unswell fix README.md docs --write
```

```text
docs/guide.md: AI-sounding index 100 -> 45; removed 4 em dashes, 1 emoji in a heading, 1 emoji at the start of a list item, 2 arrows, 2 bold labels. Left for you: 1 stock word.
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

The command calls no model, and no text leaves the machine. It analyzes the
page again after each pass, until nothing is left to replace, and reports the
index before and after.

A page at the level `low` is left as it is. `unswell check` lists no marks
there, and every measurement below is on pages with an index of 40 or more.
`--include-low` replaces the marks of such pages as well. Files change only with `--write`. The command keeps
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
bytes. The published binary of `v0.1.0-alpha.11` writes, byte for byte, every
fixed page that the judges read in the three runs.

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
the first run and did not in the second, so that claim is not made.

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

## Limits

- The index was fitted and measured on English documentation that someone set
  out to clean. It says how much a page carries the marks that such edits
  remove. It does not say who wrote the page.
- It counts six kinds of marks on the surface of a page. A sentence that a
  cleaning edit rewords is outside it, and so is most of what such edits
  change. Of the held-out pages before a commit, 41.8% are not `high`.
- Against a page fixed with a third-party linter's findings, the advantage is
  measured for `unswell fix`. For a model that follows unswell's report it
  held in one run of two.
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
