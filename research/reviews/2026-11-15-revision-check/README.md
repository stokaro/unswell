# A check of a revision, and what a revision does to a post

The [earlier record](../2026-11-15-style-guide-text/README.md) showed that the
AI-sounding index sees a post that a model wrote under a style guide, and that
a model that revises such a post by the report makes it read less like a
model's than it did. Two questions were left:

1. After a revision by the report, can the post still be told from text that
   a person wrote?
2. Can the product check one revision: keep it when it helps and reject it
   when it does not, whichever model revised?

The index answers neither. It falls after almost any edit of a run of
sentences, also after one that makes the page read more like a model's.

[RESULTS.md](RESULTS.md) holds the numbers, and `make check-revision-check`
recounts every one of them from the data here. The bars were written on
[issue 583](https://github.com/stokaro/unswell/issues/583) before the run on
the measured half.

## Outcome

Of the 18 bars, 8 hold.

- **A revised post is still told from human text.** None of the 8 bars on
  this holds. The panel picks the post on 22 of 30 as it is and on 22 of 30
  when the accepted revisions are published; with every revision published
  it picks 19 and 22 of 30. One judge alone tells 29 or 30 of 30 in every
  case. The median index of the revised posts fell from 68 to 23. A revision
  by the report lowers the index and does not make a post read as a person's.
- **The check sorts revisions.** It reached a verdict on every pair and
  accepted 15 of 86 and 15 of 57. The third model named the page as more like
  a model's text on 21 of the 30 accepted revisions and on 57 of the 113
  rejected ones. On 30 pairs with a known difference the comparison found it
  on 26.
- **The search for statements is the weak part.** With one judge it found
  97.8% of the statements kept, under the bar of 98%; 53 of the 77 that it
  did not count as kept are ones that the judge called kept with words that
  are not in the revision. That judge noticed damage made on purpose on 39 of
  58 copies, under the bar of 80%, and the other judge on 48 of 58. The check
  accepted 3 and 2 damaged copies.
- **Em dashes.** None of 86 revisions of one reviser and 1 of 86 of the other
  added one.

## What the product gained

`unswell compare <page> <revision>` checks one revision with a model that the
user names. The judge reads the opening of both texts, about 450 words,
without a hint that one is a revision of the other, in both orders, and names
the one that reads more like text written by a language model. The revision
counts as better only when the judge names the page in both orders. The
command also has the model list the statements of the page, look for each in
the page itself, and look for those it found there in the revision; a
statement counts as kept only when the model quotes words that stand in the
revision. Without a model the command checks that no number, code, or link is
lost and that no em dash is added. It accepts or rejects the revision.

The report of `unswell check` gained one line of advice for runs of
sentences: do not add an em dash.

## Texts

The texts are those of the earlier record, prepared in the same way: the 70
posts of the style-guide set, the human blog posts, and the texts that two
models wrote under a style guide. The articles that people wrote under a
style guide are those of its control. The same hash splits every set into a
fitting half and a measured half.

## The fitting half

The 40 posts of the style-guide set in the fitting half were used to choose
what the product does. No bar was set there.

- **Against human text.** The posts as they are, and after a revision by
  three reports and by one free rewrite, each against three human blog posts.
- **A report with sentence lengths.** A report that states how long the
  sentences of the page are, how long those of people's articles are, and asks
  for about a third of the sentences to run over 20 words.
- **A line on em dashes.** The report of `v0.1.0-alpha.13` with one more line:
  do not add an em dash.
- **Two instructions.** The instruction of the earlier record told the reviser
  to keep every other sentence. A second one leaves those words out.
- **Both orders.** One judge asked about one pair twice, with the texts in
  each order, under two wordings of the question.
- **A vote by section.** A page and its revision cut at their headings, each
  section asked in both orders.
- **Damage on purpose.** Copies of a post with one sentence removed, or with
  the word "not" removed from one sentence.

## The measured half

**Pages.** The 30 posts of the style-guide set, and 60 texts that two models
wrote under a style guide: for each writer the first 30 in the order of a
hash of the id.

**Revision.** The report of the build of this change, with this instruction:
revise the page so that what the report points at is resolved; change nothing
else; keep every heading, list, table, link, and code block as it is, and
keep all facts. A page whose report lists no mark is not revised. The
revisers are `gpt-6-astra` at low reasoning effort and
`private-host/Qwen3.6-35B-A3B`.

**Check.** `unswell compare` of the same build on each page and its revision.
The judge is of the other family than the reviser:
`private-host/Qwen3.6-35B-A3B` for the revisions of `gpt-6-astra`, and
`gpt-6-astra` for the revisions of `private-host/Qwen3.6-35B-A3B`. For the
second reviser the check runs on the style-guide set and on the 30 texts of
the first writer.

**What is published.** The revision when the check accepts it, and the post
as it was otherwise.

**Human text.** For each post of the style-guide set, the three human blog
posts it was paired with in the earlier record, and three matched articles of
people under a style guide, picked by a hash. The judges read openings of
about 450 words.

**Judges.** The three of the earlier record. The third model,
`private-host/Qwen3.8-27B`, revised no page and judges in no check. It
answers one order for each page and its revision.

**Damage on purpose.** Each post of the style-guide set gets two damaged
copies, as on the fitting half. The check runs on a post and its copy with
`private-host/Qwen3.6-35B-A3B`.

**A known difference.** 30 pairs in which a text that `gpt-6-astra` wrote
under a style guide stands as the page, and the human post on the same
subject as its revision. Only the blind comparison of the check is read.

**Added after the bars were written.** Two runs have no bar.

Every revision stands in the place of its post, against the human blog
posts. The published text is the post as it was for all but 2 and 3 of the 30
posts, so the published rows say little about what a revision does. This
comparison does.

The damaged copies were checked a second time, with `gpt-6-astra` as the
judge. The bar names `private-host/Qwen3.6-35B-A3B`, and the first run showed
that it does not hold; the second run shows how much of that is the judge.

## The build behind the numbers

The revisions, the verdicts of the check, and the answers of the measured
half come from the build of commit `0eb2bf5b` of the branch of this change.

The final build differs from it in one thing. A code span written with two
backticks shifted the spans after it, and the build reported a code span as
lost that the revision had kept. That happened on one post, in the check of
both revisers. Both revisions were rejected for their sound as well, so no
verdict depends on it. The count of revisions that lost something includes
it.

The trials of the fitting half used earlier builds of the branch. The first
ones compared whole pages, where a judge named the second text on every pair
that was tried, and asked about "two versions of one page". The command now
reads openings and asks with the wording of the earlier record. The search
for statements had no pass over the page itself at first, and it did not list
instructions to the reader; both were added after trials on the fitting half.

## Data

| File | Rows |
| --- | --- |
| `data/published.jsonl.gz` | one answer of one judge on one pair: a post as it is published against a human text, with the reviser, or `none` for the post as it was; rows with `kept` set to `every revision` publish every revision without the check |
| `data/revisions.jsonl.gz` | each page that a reviser was given: the index and the em dashes before and after, the verdict of the check with the sound, what the judge named in each order, the statements, how many were kept and how the others failed, what was lost, and the answer of the third model |
| `data/damage.jsonl.gz` | each damaged copy and judge: the kind of damage, the sound, what the judge named in each order, and how many statements the check did not find |
| `data/known.jsonl.gz` | each pair with a known difference: the sound and what the judge named in each order |
| `data/fitting.jsonl.gz` | the runs of the fitting half: answers of the judges, both orders of one pair, votes by section, and the index and the em dashes of each revision |

The record holds no text of any page. A post of the style-guide set has the
number that it has in the earlier record and nothing else that tells which
post it is. The revisions, the statements that the check listed, and the
judges' reasons stay with the maintainer.

## Not measured

People as judges. Text other than engineering blog posts. A reviser other
than these two. Whether a statement that the check calls kept is kept in
every detail: the check is a model that quotes words of the revision, and the
command verifies the quote, not the meaning.
