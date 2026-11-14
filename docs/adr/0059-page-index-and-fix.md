# ADR 0059: A page index of AI-sounding documentation, and a fix for its marks

Status: proposed on November 14, 2026. It takes effect when the maintainer
merges it.

## Context

The tool exists to find AI-sounding wording in documentation and to help
lower it. The rules of lean wording did not do the first:
[ADR 0058](0058-default-rules-quiet-on-human-pages.md) took six of them out
of the defaults because they report on pages that people wrote and do not
follow the edits of people who clean such wording out. What does follow those
edits is on the surface of a page: em dashes, emoji in headings and list
items, arrows, stock words, bold labels in lists.

A rule reports one construction and is switched on or off as a whole. A page
needs one reading that weighs these marks together, stays low on pages that
people wrote, and says where the marks are.

## Decision

1. **`unswell check` computes an index from 0 to 100 for every Markdown page
   of 150 prose words or more.** Six marks are counted per 1,000 prose words.
   The index is 100 times the logistic function of a constant plus the
   weighted natural logarithms of one plus each rate, as a whole number.

   | Signal | Weight |
   | --- | ---: |
   | `em-dash` | 0.94 |
   | `emoji-in-heading` | 0.69 |
   | `emoji-at-list-item` | 0.71 |
   | `arrow` | 0.81 |
   | `stock-word` | 0.62 |
   | `bold-label-in-list` | 0.44 |
   | constant | -1.23 |

   The weights were fitted on the development part of the reference only:
   pairs before and after a de-slop commit, and README files as they were
   before November 2022, both split by a hash of the repository.

2. **Levels.** `elevated` from 40, `high` from 70. The threshold of `high` is
   set so that at most one in twenty pages that people wrote reaches it.

3. **The result lists the marks.** Each mark has its signal, its place, and a
   mechanical replacement where one exists. Each signal has one line of
   advice. The text and Markdown reports list the marks of the pages above
   `low`.

4. **`unswell fix` applies the mechanical replacements.** An em dash becomes
   a comma, a period between two full clauses, a colon after the short label
   of a list item, or nothing after other punctuation. An emoji leaves a
   heading or the start of a list item.
   An arrow becomes `->`. A bold label at the start of a list item loses its
   asterisks. A stock word has no replacement. The command calls no model and
   writes only with `--write`. It leaves a page at the level `low` as it is
   unless `--include-low` is given: the check lists no marks there, and no
   measurement covers such pages.

5. **The index is an index.** It has no gate, does not change the exit code,
   and is never reported as a probability or as a statement about the author.
   [ADR 0037](0037-diagnostics-not-authorship.md) stands.

## What was measured

Each bar was written on issue 525 before its run.

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

The bars for the index were an AUC of at least 0.80, a share of at least 70%
on the held-out pairs and on the web-flow edits among them, and at most 5% of
the README files at `high`. All four hold. The web-flow share rests on
44 pairs that differ and holds by one pair.

With the rules that ship on, 47 of the same 566 README files have a
finding (8.3%). No rule reports on more than one page in twenty; the
most frequent is `filler.nominal-support` with 24 pages (4.2%).

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

**Two more runs on new pages.**

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

The bars for the fix: the index lower on at least 90% of the pages after
each of the two fixes; each of the two the less AI-sounding version against
the original on more than half of the decisive pages, with the lower end of
the interval above 50%; and the same for the report version against the
version of the best third-party linter. All of them hold on the versions of
the registered run, and the two bars of `unswell fix` hold again on the
second reading with the command as it ships. The third judge had the same
bar for its own answers in three comparisons and meets it in each.

On the second run the bar of the report version against write-good's version
does not hold, and on the third run the command has that bar and meets it.
The decision stands for the command. No claim is made that a model guided by
unswell's report beats one guided by a third-party linter.

## Consequences

`unswell check` prints the index lines for Markdown pages. A project that
reads the JSON report finds `ai_sounding` on each Markdown document.

The index and the default rule `format.em-dash-frequency` both read em
dashes. The rule reports a page; the index weighs the dashes with five other
marks and lists each one.

What the measurement does not cover:

- Pages below 150 prose words, other formats, and other languages.
- Who wrote a page. The pages before a de-slop commit are pages that someone
  found AI-sounding, not pages known to be generated.
- People as judges. The fixer and the two judges of the panel are models of
  one family; one judge of another family read part of the comparisons.
- Whether the text after `unswell fix` is good prose. The judges answered one
  question, which version reads more like a model's. The command can leave a
  comma between two clauses that should be two sentences.
- Most of what a cleaning edit changes. The index does not move in 84 of
  the 189 held-out pairs.

The [record](../../research/reviews/2026-11-12-human-reference/RESULTS-index.md)
holds the pages, the pairs, and the answers of the judges. Its check recounts
every number here, recomputes each page's index from its counts with the
weights above, and a test fails when the weights of the product and of the
record part.
