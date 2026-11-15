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

## Data

| File | Rows |
| --- | --- |
| `data/published.jsonl.gz` | one answer of one judge on one pair: a post as it is published against a human text, with the reviser, or `none` for the post as it was |
| `data/revisions.jsonl.gz` | each page that a reviser was given: the index and the em dashes before and after, the verdict of the check with the sound, what the judge named in each order, the statements and how many were kept, what was lost, and the answer of the third model |
| `data/damage.jsonl.gz` | each damaged copy: the kind of damage and how many statements the check did not find |
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
