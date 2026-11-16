# Every sentence that a revision changes

The [earlier record](../2026-11-15-sentence-check/README.md) measured the
comparison of sentences of `unswell compare` on two kinds of damage: a
sentence removed, and the word "not" removed from a sentence. The check
reported both on every copy. The rule was built for those two kinds. It
looked for the content words of a sentence and counted its negations, so a
revision that keeps the content words and changes what the sentence says
passed it.

This record measures the rule that replaced it. A sentence is kept only when
it stands in the revision word for word, and every other sentence goes to the
judge beside what replaced it.

[RESULTS.md](RESULTS.md) holds the numbers, and `make check-changed-sentences`
recounts every one of them from the data here. The bars were written on
[issue 595](https://github.com/stokaro/unswell/issues/595) before the run
on the measured part.

## Outcome

Of the 32 bars, 31 hold.

- **Every changed sentence is a suspect.** Without a model, the changed
  sentence is a suspect on all 96 damaged copies of the posts and on all 102
  of the fresh texts, for each of the five kinds of damage. No harmless copy
  of 56 and of 41 has a suspect.
- **With one judge the check reports the damage almost always.** With
  `gpt-6-astra` the changed sentence stays a suspect that the judge does not
  clear on 96 of 96 copies of the posts and on 101 of 102 of the fresh texts,
  and the check accepts no damaged copy.
- **With the other judge one kind of damage gets through.** With
  `private-host/Qwen3.6-35B-A3B` it is 95 of 96 and 97 of 102, and the check
  accepts 2 damaged copies. A sentence removed, "not" put in, and a word
  turned around are reported on every copy, 60, 53, and 54. A sentence that
  lost its last clause is reported on 11 of 16, under the bar of 80%: the
  command showed the judge the sentence beside what was left of it, and the
  judge found the sentence there. Both accepted copies are of this kind.
- **One copy with "not" removed went unreported with each judge.** It is a
  copy of the text that says its sentences twice. The changed sentence had no
  counterpart near its place, and the place that the judge was shown holds
  the other copy of the sentence.
- **Statements are found kept.** 100.0% and 99.4% over the revisions of the
  posts, 99.7% and 98.9% over those of the fresh texts.
- **The rule rejects many ordinary revisions.** Of the 27 revisions of posts
  by each reviser, a sentence that the judge did not clear stands in 5 and in
  13; of the 30 revisions of fresh texts, in 7 and in 19. No bar was set for
  this, and nobody checked whether each of those sentences did lose
  something.
- **Verdicts.** Four checks hung on the endpoint of one judge and were
  stopped; each was run once more, as the issue had set before the run. Every
  pair has a verdict.

## What the product gained

- **Kept means word for word.** The command looks for each sentence of the
  page near its place in the revision, with the same words in the same order.
  Case, punctuation, markup, and contractions do not count, and the sentence
  may stand inside a longer one. Every other sentence is a suspect and goes
  to the judge. Before, the command named what makes a sentence a suspect,
  and whatever it did not name passed.
- **Every line outside fenced code.** The command cuts the page into
  sentences itself and no longer asks what the engine reads as prose: running
  prose, list items, rows of tables, headings, the text of an image, an
  indented block. A sentence has the words of its code spans, so a term that
  a reviser sets as code stays a word of its sentence.
- **Short sentences too.** A sentence with fewer than three content words is
  compared word for word like any other, and its counterpart is looked for by
  all its words.
- **One place for each sentence.** A sentence of the page takes the words
  that it matches in the revision, so a sentence that the page says twice
  needs two places there.
- **Questions of twenty.** A long list of suspects goes to the judge in
  several questions.

## Texts

The texts are those of the earlier records, prepared in the same way. The
style-guide set has 70 posts: 40 in the fitting half and 30 in the measured
half. It has no others.

## Damage and harmless copies

A copy changes one sentence of a page. The sentence is picked by a hash among
the sentences of running prose without a digit, code, or a link, so that the
checks of numbers, code, and links do not notice the change.

| Kind | What is done to the sentence |
| --- | --- |
| A sentence removed | The sentence is taken out. |
| "not" removed | The word "not" is taken out of it. |
| "not" put in | The word "not" is put after "is", "are", "was", "will", "can", "does", "do", "should", or "must". |
| A word turned around | One word is replaced by its opposite: before and after, above and below, must and may, always and sometimes, all and some, more and less, first and last, higher and lower, larger and smaller, faster and slower, newer and older, earlier and later, increases and decreases, enables and disables, required and optional, same and different. |
| A clause cut off | Its last clause is cut, from "if", "when", "unless", "because", "until", or "while" to the end. |
| Two sentences joined (harmless) | Two sentences of one paragraph are joined with ", and". |
| Contractions (harmless) | The negations of the page are written as contractions. |

## The fitting half

The rule was fitted on the 40 posts of the fitting half and on nothing else:
on copies of each kind, and on the revisions that two models made of those
posts by the report. Two things changed on the way. A first build compared
the pieces of prose as the engine gives them; two sentences were then called
missing because the revision had set one of their words as code. And one
judge often found a reworded sentence with the words of the sentence it had
been asked about; those words are not in the place it was shown, and that
alone rejected 5 of 16 revisions. The command now asks once more about a
sentence that was found without the words of its place.

The fitting rows of this record come from the build of the record: the copies
of all 40 posts without a model, 50 damaged copies with each judge, and the
40 revisions of each reviser.

## Two starts that were stopped

The bars were written on the issue, and the measured run started twice with
an earlier build. Each time its first stage, the comparison without a model,
showed damaged copies whose changed sentence was no suspect: a changed
sentence that the judge is never shown. Those are holes in the rule itself,
so each time the run was stopped and the rule was changed. The issue has a
comment for each stop, written before the next start, with what had been
seen.

**The first start** showed 4 such copies of 261, all of posts. Two were
short sentences: the rule did not compare a sentence with fewer than three
content words, and "We did not run it for this post." without its "not"
passed. Two stood where the engine reads no prose, in the text of an image
and in an indented block. It was stopped during its judged stage; the
verdicts on 66 and 56 damaged copies and on 39 revisions had been seen as
counts, and no answer of a judge. The command now reads every line outside
fenced code and compares a short sentence like any other.

**The second start** showed 3 such copies of 244, all of one fresh text. That
text says its sentences twice, and the second copy of a sentence vouched for
the changed one. It was stopped before its judged stage began. A sentence of
the page now takes its own words in the revision.

The third start is the measurement of this record. Its copies are new, with
another salt, and change no sentence that an earlier run or start changed.

## The measured part, and what had been read before

The 30 posts of the measured half have been read by the two earlier records
and by the two starts above. The copies of this record change other
sentences. The posts are still not unread, and there are no others. So every
bar is read on two kinds of pages:

- **The posts.** The 30 posts of the measured half, with new copies.
- **Fresh texts.** 30 texts that two models wrote under a style guide on the
  subjects of human posts, for each writer the 15 after the first 60 in the
  order of a hash of the id. They are not posts of the style-guide set. No
  judge had read them. The second start had: its model-free stage ran on
  copies of them, and one of them showed the hole that stopped it. Three
  texts of this kind are left that no run has read, too few for a
  measurement.

**Without a model.** The comparison runs on every copy.

**With a judge.** The whole check runs on each damaged copy with each of two
judges, `gpt-6-astra` at low reasoning effort and
`private-host/Qwen3.6-35B-A3B`.

**Revisions.** The revisions of the 27 posts that each reviser made for the
first record are checked again. Each reviser also revises the 30 fresh texts
by the report of the build, and the judge of the other family checks those
revisions.

**A pair without a verdict** is run once more, with the same build and the
same settings. The issue set this before the run. Every bar is read after
that retry, and RESULTS.md shows the share of verdicts on the first attempt.

## The build behind the numbers

The measured part and the fitting rows come from the build of commit
`77c51f13` of the branch of this change.

The final build differs from it in one thing that changes no result. A code
fence that is never closed runs to the end of a page, and the earlier build
read what follows it neither as sentences nor as code; the final one reads it
as code. No page and no revision of this record has a fenced code block, and
the final build names the same suspects on all 553 copies.

The published binary of `v0.1.0-alpha.16` was run on the same texts after the
release. It names the same suspects on all 553 copies; `scripts/audit.py` is
that check. From the kept answers of one judge it reaches the same verdict,
sound, statements, sentences, losses, and em dashes on the 57 revisions that
this judge checked; `scripts/reproduce.py` is that check. The damaged copies
cannot be replayed, since several copies of one page ask the same question
about it and a repeated question gets a new answer, and the answers of the
endpoint judge were not kept.

## Data

| File | Rows |
| --- | --- |
| `data/copies.jsonl.gz` | each copy of a page: its kind, how many sentences were compared, how many suspects the comparison named, and for a damaged copy whether the changed sentence is one of them |
| `data/damage.jsonl.gz` | each damaged copy and judge: the verdict, the sound with what the judge named in each order, the statements and how many were kept, the suspects and how many were not cleared, and whether the changed sentence is one of those |
| `data/revisions.jsonl.gz` | each revision with its judge: the verdict, the sound, the statements with how the others failed, the suspects, the sentences that were not cleared by their kind, and the em dashes of the page and of the revision |

The record holds no text of any page. A post of the style-guide set has the
number that it has in the earlier records and nothing else that tells which
post it is. The copies, the revisions, the sentences, and the judges' quotes
stay with the maintainer.

## Not measured

People as judges. Posts of the style-guide set that no run had read. Damage
other than the five kinds above. What a page says across several sentences,
such as the order of steps. Whether a suspect that a judge cleared with a
proper quote is in fact kept: the command verifies the quote, not the
meaning. A language other than English.
