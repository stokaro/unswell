# The words that a revision leaves out of a sentence

The [record before this one](../2026-11-16-changed-sentences/README.md)
measured the rule that sends every changed sentence of a page to the judge of
`unswell compare`. Of its 32 bars, 31 held. The one that did not: a sentence
that lost its last clause was reported on 11 of 16 copies with one of two
judges. That judge was shown the sentence beside what was left of it, answered
that the sentence is there, and quoted the remains. One more damaged copy went
unreported with each judge because the page says the sentence twice, and the
judge was shown the kept copy.

This record measures the change that answers both.

[RESULTS.md](RESULTS.md) holds the numbers, and `make check-left-out-words`
recounts every one of them from the data here. The bars were written on
[issue 599](https://github.com/stokaro/unswell/issues/599) before the run on
the measured part.

## Outcome

All 38 bars hold.

- **Every changed sentence is a suspect.** Without a model, on all 112
  damaged copies of the posts and on all 148 of the new texts, for each of
  the six kinds of damage. No harmless copy of 56 and of 40 has a suspect.
- **A removed sentence needs no judge.** For all 60 removed sentences the
  revision wrote nothing at the place, and the command reported each of them
  without a question to a model.
- **Both judges report the damage.** With `gpt-6-astra` the check reports 111
  of 112 damaged copies of the posts and 148 of 148 of the new texts. With
  `private-host/Qwen3.6-35B-A3B` it reports 112 of 112 and 147 of 148.
  Neither accepts a damaged copy.
- **A clause cut off, the bar that did not hold before.** Reported on 23 of
  23 copies with each judge: 3 of the posts and 20 of the new texts.
- **Last words cut off.** Reported on 59 of 60 copies with each judge. Each
  judge called one cut padding, on two different copies.
- **Trimmed copies are left alone.** Of 60 copies whose revision only leaves
  out padding, the check reports none, with either judge. Each judge called
  the padding padding on all 60.
- **Statements are found kept.** 99.4% and 99.4% over the revisions of the
  posts, 99.5% and 99.1% over those of the fresh texts.
- **Verdicts.** Every pair got a verdict on the first attempt, and no check
  was stopped.
- **The price.** The rule is stricter on ordinary revisions. A sentence that
  was not cleared stands in 4 and in 15 of the 27 revisions of posts, where
  the record before this one had 5 and 13, and in 8 and in 24 of the 30
  revisions of fresh texts, where it had 7 and 19. No bar was set for this,
  and nobody checked whether each of those sentences did lose something.

## What the product gained

- **The judge is shown what the revision wrote.** A sentence of the revision
  that is a kept sentence of the page, word for word, is left out of every
  passage. A kept copy of a sentence no longer stands in for a changed one.
- **Gone without a question.** When the revision wrote nothing at the place of
  a sentence, the command reports the sentence as missing on its own. No model
  is asked, so no model can clear it.
- **Its own place first.** A sentence that stands in the revision word for
  word takes its words there, and one that stands at its own place goes
  first.
- **The left-out words are named.** A sentence is cut when a sentence of the
  revision has its words in their order with some left out, and no other
  word. The command prints the words that were left out and gives them to the
  judge on a line of their own.
- **The judge answers about those words.** They are padding, or the passage
  says them elsewhere, or the sentence is changed. The command refuses
  "padding" for a negation. It refuses a quote of what is left of the
  sentence: such a quote shows only that the rest is kept.

## Texts

The style-guide set has 70 posts: 40 in the fitting half and 30 in the
measured half. It has no others. The texts are prepared as in the earlier
records.

## Copies

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
| Last words cut off | Its last three to eight words are cut, from "for", "in", "on", "with", "from", "by", "at", "under", "over", "during", "through", "into", "across", "between", or "against" to the end. What is cut has no word that opens a clause and no negation. |
| Two sentences joined (harmless) | Two sentences of one paragraph are joined with ", and". |
| Contractions (harmless) | The negations of the page are written as contractions. |
| Trimmed (harmless) | The page gets a phrase of padding in front of one sentence: "It is worth noting that", "It should be noted that", "It is important to note that", "As a matter of fact,", or "Needless to say,". The revision is the page without it. |

"Last words cut off" is new. It is a cut that does not start with one of the
six words of "A clause cut off", so the result does not rest on a list of
words that the command could know.

"Trimmed" is new and is the other side of the change. The revision leaves
words out of a sentence, as a damaged copy does, and those words told the
reader nothing. A check that reports such a sentence would reject the edit
that the report of `unswell check` asks for most often.

## The fitting half

The rule was fitted on the 40 posts of the fitting half and on nothing else,
with new copies of every kind. The judged rows of the fitting half are the
copies where the answer is the judge's: a clause cut off, last words cut off,
and trimmed. They were made with a build before a refactoring; the build of
the measurement finds the same suspects, the same passages, and the same
left-out words on all 344 fitting copies.

The check of ordinary revisions on the fitting half showed one thing that is
not about sentences. The check had accepted 4 of the 40 revisions of one
reviser, and with the new build it accepted none. Those 4 fail only the blind
comparison of the sound, whose prompt did not change: the judge was asked
again and named the page in both orders for 1 revision, where it had done so
for 5. The answer of that judge to the blind comparison is not stable from
one run to the next.

## The measured part, and what had been read before

The 30 posts of the measured half have been read by the three earlier
records and by two stopped starts of the last one. The copies of this record
change other sentences. The posts are still not unread, and there are no
others. Few of them have a sentence left with a clause to cut off: 3.

So every bar is read on two kinds of pages:

- **The posts.** The 30 posts of the measured half, with new copies.
- **New texts.** 30 texts that no run had read. Of the texts of that kind
  that existed, 3 were unread, so these were written for this record: two
  models wrote them under a style guide from the outlines of human posts of
  the measured half, 15 each, as the earlier texts were written, under the
  other of the two style guides. They are second texts on subjects of texts
  that earlier runs read, in other words. A name in the data marks them with
  a 2.

**Without a model.** The comparison runs on every copy.

**With a judge.** The whole check runs on each damaged copy and each trimmed
copy with each of two judges, `gpt-6-astra` at low reasoning effort and
`private-host/Qwen3.6-35B-A3B`.

**Revisions.** No new revisions were made. The revisions of the 27 posts that
each reviser made for the first record, and the revisions of 30 fresh texts
that each reviser made for the record before this one, are checked again with
the new build, by the judge of the other family.

**A pair without a verdict** is run once more, with the same build and the
same settings. The issue set this before the run. Every bar is read after
that retry, and RESULTS.md shows the share of verdicts on the first attempt.

## The build behind the numbers

The measured part comes from the build of commit `22e32981` of the branch of
this change, and so do the fitting rows of the comparison without a model.
The judged rows of the fitting half come from the build of commit `60fb081f`,
before a refactoring; the build of the measurement finds the same suspects,
the same passages, and the same left-out words on all 344 fitting copies.

The released build differs from the build of the measurement in the help text
of the command.

The published binary of `v0.1.0-alpha.17` was run on the same texts after the
release. It finds the same suspects, the same passages, and the same left-out
words on all 760 copies; `scripts/audit.py` is that check. From the kept
answers of one judge it reaches the same verdict, sound, statements,
sentences, losses, and em dashes on the 57 revisions that this judge checked;
`scripts/reproduce.py` is that check. The damaged and the trimmed copies
cannot be replayed, since several copies of one page ask the same question
about it and a repeated question gets a new answer, and the answers of the
endpoint judge were not kept.

## Data

| File | Rows |
| --- | --- |
| `data/copies.jsonl.gz` | each copy of a page: its kind, how many sentences were compared, how many suspects the comparison named; for a damaged copy whether the changed sentence is one of them and whether a judge is asked about it; for a trimmed copy whether the command names the words that were left out |
| `data/damage.jsonl.gz` | each damaged copy and judge: the verdict, the sound with what the judge named in each order, the statements and how many were kept, the suspects and how many were not cleared, and whether the changed sentence is one of those |
| `data/trims.jsonl.gz` | each trimmed copy and judge: the same parts, whether the sentence that lost its padding is reported, and whether the judge called the left-out words padding |
| `data/revisions.jsonl.gz` | each revision with its judge: the verdict, the sound, the statements with how the others failed, the suspects, the sentences that were not cleared by their kind, how many were cleared as padding, how many were reported without a question, and the em dashes of the page and of the revision |

The record holds no text of any page. A post of the style-guide set has the
number that it has in the earlier records and nothing else that tells which
post it is. The copies, the revisions, the sentences, the left-out words, and
the judges' quotes stay with the maintainer.

## Not measured

People as judges. Posts of the style-guide set that no run had read. Damage
other than the six kinds above. A sentence that a revision cuts and also
rewords: it goes to the judge as a reworded sentence. Words that a revision
puts in around a kept sentence. What a page says across several sentences,
such as the order of steps. Whether a judge that calls left-out words padding
is right: the command names the words and refuses a negation, and the rest is
the judge's reading. A language other than English.
