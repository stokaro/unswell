# Three pilots of revising a post

November 17, 2026. The [record of the two page habits](../2026-11-17-page-habits/README.md)
left one question open: does a post read as a person's once a revision has
taken those habits out? These pilots asked it on the fitting half, in three
ways that had not been tried.

[RESULTS.md](RESULTS.md) holds the numbers, and `make check-revision-pilots`
recounts them from the data here. No bar was set (these are pilots on the
fitting half), so nothing here is a measurement on pages held back.

## Outcome

The answer is no, in all three ways.

- **By the report, with the closing sentences listed.** A model got the
  report of `unswell check` for the post and, beside it, the sentences that
  `syntax.closing-denial-density` lists. It took them out: 98 closing
  sentences became 4, 23 posts in one register became 1, and the median index
  fell from 68 to 24. Three blind judges then compared each of the 38 revised
  posts with three blog posts that people wrote. Together they pick 25 of the
  38, where they picked 26 of the same posts before. `gpt-6-astra` alone picks
  38 of 38 before and after.
- **In the manner of two posts of people.** The model got the openings of two
  human posts and was asked to say what the page says in their manner. By the
  product's own readings the page got worse: the median index rose from 70 to
  88. The one judge that was asked picks 33 of 39 revised posts, where it
  picked 27. The pilot stopped there.
- **In rounds, by the judge's own remarks.** Six posts were revised four
  times each. Before every round `gpt-6-astra` compared the post with two
  human posts and said why it took the post for a model's, and the reviser
  got those remarks. The judge named the post in all 60 judgments.

So the rules name habits that a reader notices, and a model can remove them
on request. Removing them does not make a post read as a person's to these
judges.

## Why one judge decides little and another everything

The three judges are far apart. On the 40 posts as they are, `gpt-6-astra`
picks 40, `private-host/Qwen3.8-27B` picks 28, and
`private-host/Qwen3.6-35B-A3B` picks 13; together they pick 28. A count of
the three together follows the middle judge. A revision that moves that judge
alone would pass a bar on the three together while the first judge still
tells every post, so RESULTS.md gives each judge beside the count.

The first judge is not reading the subject. When the fitting half was first
read, two of the judges compared 147 texts that two models wrote under a
style guide with the human post whose outline each text follows. `gpt-6-astra`
named the model's text in 147 of the 147 pairs, and the middle judge in 146.
Those answers are in the data of this record too.

## How the pilots were run

**Posts.** The 40 posts of the fitting half of the style-guide set, each with
the three human blog posts that the earlier records paired it with.

**The reviser** is `private-host/Qwen3.6-35B-A3B` in all three pilots. Two
posts of the first pilot and one of the second came back shorter than half
the page and were left out.

**The report** of the first pilot is the lines of `unswell check` of
`v0.1.0-alpha.18` about the page index, and then the two findings of a page
as a whole. The text report gives the place of each closing sentence; the
pilot gave the sentence itself.

**The judges** got the opening of the post and of a human post, about 450
words each, as post A and post B in an order fixed by the pair, and named the
one that reads more like a model's. In the third pilot the order changes
from round to round, and the third human post of each post took no part.

## Data

| File | Rows |
| --- | --- |
| `data/revisions.jsonl.gz` | each post and way of revising: whether a revision came back, and before and after it the index, the number of closing sentences that the first rule lists, and whether the second rule reports the page |
| `data/answers.jsonl.gz` | each judged pair: the text (the post as it is, or one of the two revisions), the post, the number of its human post, the judge, and what the judge named |
| `data/rounds.jsonl.gz` | each judgment of the third pilot: the post, the round, the pair, and what the judge named |
| `data/subjects.jsonl.gz` | each pair of a human post and a model's text on its subject: the writer, an id, the judge, and what the judge named |

The record holds no text. A post has the number that it has in the earlier
records. The revisions, the reports, and the judges' remarks stay with the
maintainer, and so do the texts that the scripts in `scripts/` need.

## Not measured

The measured half: every number here is on the fitting half. A second
reviser. A revision by a person, or a post with paragraphs that a person
wrote, which is the one thing these pilots give a reason to try. People as
judges.
