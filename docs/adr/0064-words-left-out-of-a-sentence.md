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

FITTING

## What was measured

MEASURED

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
