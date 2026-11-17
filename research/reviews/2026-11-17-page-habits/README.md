# Two habits of a page

November 17, 2026. A model that writes under a style guide leaves few of the
marks that the rules of `unswell check` count, and its page can still read as
a model's. This record measures two rules that name two reasons for that:

- `syntax.closing-denial-density`: paragraph after paragraph ends by saying
  what something is not, or what it does not do.
- `syntax.uniform-register`: a page of 25 sentences or more has no
  contraction, no aside in parentheses, no question, no exclamation, and no
  sentence that opens with "And", "But", or "So".

[RESULTS.md](RESULTS.md) holds the numbers, and `make check-page-habits`
recounts every one of them from the data here. The bars were written on
[issue 603](https://github.com/stokaro/unswell/issues/603) before the rules
ran on the measured parts. [Habits of a page](../../../docs/page-habits.md)
says what each rule reads, and
[ADR 0065](../../../docs/adr/0065-habits-of-a-page.md) has the decision.

## Outcome

All 10 bars hold.

- **Quiet on people.** Each rule reports at most 3.3% of the pages of each
  kind that people wrote: of 77 blog posts, 1 and 0; of 205 articles written under
  a style guide, 1 and 4; of 61 documentation pages, 2 and 2. The limit for a
  rule that ships on is one page in twenty.
- **The posts.** Of the 30 posts of the style-guide set, the first rule
  reports 13 and the second 18. At least one of them reports 23.
- **Model texts.** Of 120 texts that two models wrote under a style guide,
  the second rule reports 110. The first reports 1: the denials are a habit
  of the posts and not of those two models.

Two things came out without a bar.

- **Today's documentation.** Of 3,302 pages of public repositories, the first
  rule reports 8.8% and the second 6.1%. Who wrote those pages is unknown.
- **What people remove.** On the reference of files before and after a commit
  that says it removes AI-sounding wording, neither rule moves: on the test
  part the first reports 54 of 985 files before the commit and 55 after it,
  and the second 41 and 42. People who clean documentation that way don't
  take these habits out (most of those files are README pages, with few long
  stretches of prose).

## Where the rules came from

The maintainer read one post of the measured half of the style-guide set, a
post on which `unswell check` reported no finding and which one of three
blind judges had taken for a model's text in every pair. The maintainer named
what gives it away, and two of those things could be counted. So both rules report that
post, and that was known before the run. The other 29 posts of the measured
half had been read by earlier records for other questions and never with
these rules.

A third candidate was dropped on the fitting halves: a pair of words that the
next sentence repeats from the end of the one before. Counted with a
prototype outside the engine, people do that more often than the posts do.

## Pages

| Set | Fitting part | Measured part |
| --- | ---: | ---: |
| Posts of the style-guide set | 40 | 30 |
| Blog posts that people wrote before November 2022 | 74 | 77 |
| Articles that people wrote under a style guide, from four repositories | 228 | 205 |
| Documentation pages that people wrote before November 2022, from documentation pull requests | 60 | 61 |
| Texts that two models wrote, with and without a style guide | 295 | |
| Fresh texts that the same two models wrote under a style guide | | 120 |

The first three sets and their halves are those of the
[record of the second reading](../2026-11-15-style-guide-text/README.md). The
documentation pages are the 121 of the
[human reference](../2026-11-12-human-reference/README.md), split by
repository as there. The fresh texts were written for the three records of
`unswell compare` from the outlines of human posts; no fitting used them.

Two sets have no parts and carry no bar: 3,302 pages of public repositories
today and 455 pages of the panel corpus that three models wrote or polished.

The forms of a denial, the marks of a casual sentence, and the limits of both
rules were chosen on the fitting parts and on nothing else. The fitting rows
of this record come from the build of the measurement.

## Build

Commit `86cd7bf4` of the branch of this change, with both rules on and every
other rule off (`scripts/policy.yaml`). A page is one run of
`unswell check --stdin` on its text.

The measured rows were written a second time after the first reading. The
fresh texts come in four sets with the same ids, and the first file could not
tell them apart. The second run of the same build gave the same counts.

The final build differs from the build of the measurement in a refactoring of
two functions. It gives the same row for each of the 4,947 pages of this
record.

## Data

| File | Rows |
| --- | --- |
| `data/pages.jsonl.gz` | each page of the twelve sets: its set, its part, an id, its words, the number of closing sentences that the first rule lists (0 when it is quiet), and whether the second rule reports the page |
| `data/pairs.jsonl.gz` | each file of the reference of ADR 0056 without an assistant's co-author trailer: its part, an id, and for each rule whether it reports the file before the commit and after it |

The record holds no text, path, repository, or title. A post of the
style-guide set has the number that it has in the earlier records. Every
other page has a hash of its set and its id in the maintainer's copy.

`scripts/load.py` runs the rules, `scripts/export.py` writes the data, and
`scripts/deslop/score_pairs.py` of the human reference scores the pairs with
the same policy. They need the texts, which stay with the maintainer.

## Not measured

Whether a post that is revised until both rules are quiet reads as a
person's: the earlier record of a blind check found that a revision by the
report of the index does not get there. People as judges. Reference pages and
specifications. A language other than English. How many of the findings a
person would agree with, sentence by sentence.
