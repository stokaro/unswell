# ADR 0064: The command names the words that a revision leaves out

Status: accepted

## Context

[ADR 0063](0063-every-changed-sentence.md) sends every sentence that a
revision changes to the judge, beside the sentences of the revision from its
place. The command decides what the judge is shown, and the judge decides
what it reads there. The
[record of that decision](../../research/reviews/2026-11-16-changed-sentences/RESULTS.md)
showed where that is not enough:

- **A cut clause.** A sentence that lost its last clause was reported on 11
  of 16 copies with one of two judges, under the bar of 80%. The judge was
  shown "A when B." beside "A." and answered that the sentence is there, with
  a quote of "A". Both damaged copies that the check accepted were of this
  kind. The quote stood in the passage, so the command took it. But a quote
  of what is left of a sentence shows only that the rest is kept.
- **A sentence said twice.** When a page says one sentence twice and the
  revision changes one copy, the place that the judge was shown held the kept
  copy, and the judge found the sentence there. One damaged copy of 198 went
  unreported for that, with each judge.
- **A removed sentence still needed the judge.** The judge was shown the kept
  sentences around the gap and had to say that the sentence is not among
  them. The rule that its quote has to carry a word of the sentence's own
  protected the answer; the question was still asked.

## Decision

**The judge is shown what the revision wrote, not what it kept.** A sentence
of the revision whose words are all words of kept sentences of the page says
what those sentences say. It is left out of every passage. So the kept copy
of a sentence that the page says twice no longer stands in for the changed
one.

**A sentence with nothing at its place is gone, and no model is asked.** When
the revision wrote nothing where the sentence stood, there is nothing to
show. The command reports the sentence as missing on its own. A revision
that removes a sentence and leaves the sentences around it as they were is
rejected whatever the judge would say.

**A sentence takes its own place first.** A sentence that stands in the
revision word for word takes its words there. Those that stand within two
sentences of their place take theirs before the others, so a sentence does
not take the place of its twin further down, and the report names the copy
that was changed.

**The command names the words that a revision leaves out of a sentence.** A
sentence is cut when a sentence of the revision near its place has its words
in their order with some left out, and no other word. That holds when one
run of words is gone and at least three words are left, or when several runs
are gone and at least half of the words is left. The command knows exactly
which words are gone. It prints them in the report and gives them to the
judge on a line of their own.

**For a cut sentence the judge answers about those words.** They are padding
when they tell the reader nothing. They are kept when the passage says
elsewhere what they said. Otherwise the sentence is changed. The command
takes the answers with two rules of its own:

- A negation is never padding. A cut that holds "not", "no", "never", or
  another negation cannot be cleared that way.
- A quote that clears a cut sentence needs a word which the rest of the
  sentence does not have: one of the words that were left out, or a word that
  the sentence never had. A quote of what is left of the sentence clears
  nothing, and the command asks once more.

With these, a judge that answers "it is there" and quotes the remains, as the
judge of the earlier record did, does not clear a cut. To clear it, a judge
has to say of the named words that they are padding.

## What the fitting half showed

The rule was fitted on the 40 posts of the fitting half of the style-guide
set and on nothing else, with new copies. Two kinds of copies are new. One
cuts the last words of a sentence from a preposition, so that the result does
not rest on the words that open a clause. The other is harmless: the page has
a phrase of padding in front of one sentence, and the revision leaves it out.
The [record](../../research/reviews/2026-11-16-left-out-words/RESULTS.md)
recounts these numbers.

**Without a model.** The changed sentence is a suspect on all 229 damaged
copies. For all 40 removed sentences the revision wrote nothing at the place,
so the command reports them without a question. No harmless copy of 75 has a
suspect, and on all 40 trimmed copies the command names the padding as the
words that were left out.

**With a judge.** A clause cut off is reported on 32 of 32 copies with each
judge. Last words cut off are reported on 39 of 40 with one judge and on 40
of 40 with the other; the one that was cleared lost "in more detail", which
the judge called padding. No damaged copy is accepted. Of the 40 trimmed
copies the check reports none, with either judge.

**On ordinary revisions.** A sentence that was not cleared stands in 12 and
in 16 of 40 revisions, where it stood in 9 and in 14 under ADR 0063. The
check of one reviser's 40 revisions accepted none, where it had accepted 4.
Those 4 fail only the blind comparison of the sound, whose prompt did not
change: asked again, the judge named the page in both orders for 1 revision,
where it had done so for 5.

## What was measured

The bars were written on issue 599 before the run. Every bar was read on two
kinds of pages: the 30 posts of the measured half of the style-guide set,
which earlier measurements had read, and 30 texts that no run had read, which
two models wrote for this measurement. All 38 bars hold.

| What | Posts | New texts |
| --- | --- | --- |
| The changed sentence is a suspect, without a model | 112 of 112 | 148 of 148 |
| Of them, reported without a question to a model | 30 | 30 |
| Harmless copies with a suspect, without a model | 0 of 56 | 0 of 40 |
| The check reports the damage, judge `gpt-6-astra` | 111 of 112 | 148 of 148 |
| The check reports the damage, judge `private-host/Qwen3.6-35B-A3B` | 112 of 112 | 147 of 148 |
| Damaged copies accepted, the two judges | 0 and 0 | 0 and 0 |
| Trimmed copies whose sentence is reported, the two judges | 0 and 0 of 30 | 0 and 0 of 30 |

By kind of damage, the two kinds of pages together:

| Judge | A sentence removed | "not" removed | "not" put in | A word turned around | A clause cut off | Last words cut off |
| --- | --- | --- | --- | --- | --- | --- |
| `gpt-6-astra` | 60 of 60 | 18 of 18 | 47 of 47 | 52 of 52 | 23 of 23 | 59 of 60 |
| `private-host/Qwen3.6-35B-A3B` | 60 of 60 | 18 of 18 | 47 of 47 | 52 of 52 | 23 of 23 | 59 of 60 |

**The bar that did not hold under ADR 0063 holds.** A sentence that lost its
last clause is reported on 23 of 23 copies with each judge. Only 3 of those
copies are of the posts: the earlier measurements used most of their
sentences that have a clause to cut.

**Each judge called one cut padding**, of 60 with the last words cut off, on
two different copies.

**The statements.** The check finds 99.4% and 99.4% of the statements kept
over the revisions of the posts, and 99.5% and 99.1% over revisions of fresh
texts.

**Ordinary revisions.** A sentence that was not cleared stands in 4 and in 15
of 27 revisions of posts, where it stood in 5 and in 13 under ADR 0063, and
in 8 and in 24 of 30 revisions of fresh texts, where it stood in 7 and in 19.
No bar was set for this.

The [record](../../research/reviews/2026-11-16-left-out-words/RESULTS.md)
recounts these numbers. The published binary of `v0.1.0-alpha.17` finds the
same suspects, passages, and left-out words on all 760 copies of the record,
and it reaches the recorded verdict on the 57 revisions that could be
replayed from kept answers of a judge; see the
[alpha.17 evidence](../release/v0.1.0-alpha.17.md).

## Consequences

- The `review` package gains `SentenceCut`, and `Suspect` gains `Cut` and
  `Padding`. `Suspects` sets the outcome of a sentence with nothing at its
  place to `FactMissing`; a caller asks a model only about the others.
- `RevisionProtocol` is `unswell-revision-v3`: the question about sentences
  has a line for the words that were left out, and "padding" is an answer.
- The JSON report keeps the schema `unswell-compare-v2`. A suspect can have
  the reason `cut`, the field `cut`, and the field `padding`.
- The command asks fewer questions: none about a removed sentence whose
  neighbors are kept.
- A padding phrase that holds a negation word, such as "it goes without
  saying that", is not accepted as padding. The report names the words, and
  the writer decides.

## Limits

- Whether left-out words are padding is the judge's call. The command names
  the words and refuses a negation; it does not know what a word tells.
- A sentence that the revision cuts and also rewords is not a cut sentence.
  It goes to the judge as a reworded one, as before.
- Words that a revision puts in around a kept sentence are not shown to the
  judge: a sentence is kept when it stands in the revision word for word,
  also inside a longer one.
- The limits of ADR 0063 about what a page says across several sentences,
  about fenced code, and about English stand.
