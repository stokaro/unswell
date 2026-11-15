# The sentences of a page in its revision

The [earlier record](../2026-11-15-revision-check/README.md) measured
`unswell compare`, the check of one revision. Two of its bars failed for the
same part of the check, the search for statements:

1. On copies of a post with one sentence removed, or with the word "not"
   removed from one sentence, the search reported a statement as not kept on
   39 of 58 copies with one judge. The check accepted 3 damaged copies with
   that judge and 2 with the other.
2. One judge found 97.8% of the statements kept, under 98%. Of the 77
   statements that it did not count as kept, 53 were ones that it called kept
   with words that are not in the revision.

This record measures the change that followed. The command now compares the
sentences of the page with the revision itself, without a model, and asks the
judge only about the sentences that it could not find.

[RESULTS.md](RESULTS.md) holds the numbers, and `make check-sentence-check`
recounts every one of them from the data here. The bars were written on
[issue ISSUE](https://github.com/stokaro/unswell/issues/ISSUE) before the run
on the measured part.

## Outcome

OUTCOME

## What the product gained

- **A comparison of sentences without a model.** For each sentence of the
  page the command looks for a counterpart in the revision: one sentence near
  the same place, or two or three in a row, with at least six tenths of its
  content words. The share is counted over all its content words and over its
  own words, those that the two sentences before it and the two after it do
  not have. A sentence without a counterpart is a suspect, and so is one whose
  counterpart has fewer negations.
- **A question about the suspects.** The judge gets each suspect beside the
  sentences of the revision from its place. It clears a suspect only with a
  quote that stands in that place; for a missing sentence the quote also has
  to carry one of the sentence's own words. One suspect that is not cleared
  rejects the revision.
- **A second question about a statement.** Before a statement counts as not
  kept, the command asks once more about it, in a list of those statements
  only.

## Texts

The texts are those of the earlier records, prepared in the same way. The
style-guide set has 70 posts: 40 in the fitting half and 30 in the measured
half. It has no others.

## The fitting half

The rule was fitted on the 40 posts of the fitting half and on nothing else.

- **Damaged copies.** Each post got a copy with one sentence removed and a
  copy with the word "not" removed from one sentence, twice, with two salts,
  so that the second copies damage other sentences than the first.
- **Harmless copies.** Each post got a copy with two sentences of one
  paragraph joined with ", and", and a copy with its negations written as
  contractions. Nothing that the page says is gone from such a copy.
- **Ordinary revisions.** The revisions of the fitting half that two models
  made by the report, 40 for each.
- **The rule as it grew.** Each step of the rule was run on the first damaged
  copies and on the ordinary revisions: all content words of a sentence
  anywhere in the revision; its own words as well, near its place; words
  without their endings; paragraphs cut into sentences.
- **Two judges.** The whole check ran on the first damaged copies with each
  judge, and on the ordinary revisions with the judge of the other family
  than the reviser.

The damaged copies of the fitting half were checked with a build that did
not yet ask a clearing quote for a word of the sentence. One judge then
cleared a removed sentence with words of the sentence beside it, and the rule
was added. The rows of those copies have that rule applied to the kept
answers, so their sentences say what the final build decides. That build
also asked again only about a statement whose quote was not in the revision.
On the ordinary revisions one judge then called 22 of the 90 statements of
one revision changed, the last twenty of the list among them. The final build
asks again about every statement that was not found, and the ordinary
revisions of the fitting half were checked with it.

## The measured part, and what had been read before

The 30 posts of the measured half were read once, by the earlier record's run
with the earlier build. This change knew from that run its counts, the text
of 14 sentences that its copies had damaged, and a few statements that a
judge had not counted as kept. No threshold and no word list of the rule was
set by them. There are no posts that no run has read.

So the measurement has two kinds of pages, and every bar is read on each:

- **The posts.** The 30 posts of the measured half, with new copies. A
  sentence that the earlier record damaged in a post is not picked again.
- **Fresh texts.** 30 texts that no run had read: texts that two models wrote
  under a style guide on the subjects of human posts, for each writer the 15
  after the first 30 in the order of a hash of the id. They are not posts of
  the style-guide set.

**Copies.** Each page gets the four copies of the fitting half, with a third
salt.

**Without a model.** The comparison runs on every copy.

**With a judge.** The whole check runs on each damaged copy with each of two
judges, `gpt-6-astra` at low reasoning effort and
`private-host/Qwen3.6-35B-A3B`.

**Revisions.** The revisions of the earlier record are checked again: 86 of
`gpt-6-astra` with `private-host/Qwen3.6-35B-A3B` as the judge, and 57 of
`private-host/Qwen3.6-35B-A3B` with `gpt-6-astra`. Each reviser also revises
the 30 fresh texts by the report of the build, and the judge of the other
family checks those revisions.

## Data

| File | Rows |
| --- | --- |
| `data/copies.jsonl.gz` | each copy of a page: its kind, how many sentences were compared, how many suspects the comparison named, and for a damaged copy whether the damaged sentence is one of them |
| `data/damage.jsonl.gz` | each damaged copy and judge: the verdict, the sound with what the judge named in each order, the statements and how many were kept, the suspects and how many were not cleared, and whether the damaged sentence is one of those |
| `data/revisions.jsonl.gz` | each revision with its judge: the verdict, the sound, the statements with how the others failed, the suspects, and the sentences that were not cleared by their kind |
| `data/variants.jsonl.gz` | each step of the rule on the first damaged copies and on the ordinary revisions of the fitting half |

The record holds no text of any page. A post of the style-guide set has the
number that it has in the earlier records and nothing else that tells which
post it is. The copies, the revisions, the sentences, and the judges' quotes
stay with the maintainer.

## Not measured

People as judges. Posts of the style-guide set that no run had read. Damage other than a
removed sentence and a removed "not". Whether a suspect that a judge cleared
with a proper quote is in fact kept: the command verifies the quote and its
words, not the meaning. A language other than English.
