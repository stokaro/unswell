# ADR 0063: Every sentence that a revision changes goes to the judge

Status: accepted

## Context

[ADR 0062](0062-sentences-of-a-revision.md) made `unswell compare` compare
the sentences of a page with its revision. A sentence without a counterpart
in the revision, or with a counterpart that has one negation less, went to
the judge. The
[record of that decision](../../research/reviews/2026-11-15-sentence-check/RESULTS.md)
damaged pages in two ways, a sentence removed and the word "not" removed, and
the check reported the damage on every copy.

That result is narrow. The rule looked for exactly those two things: the
content words of a sentence, and its negations. A revision that keeps the
content words and changes what the sentence says had its counterpart and
passed:

- "after" where the page said "before", or "may" where it said "must":
  these are function words and were not compared;
- a "not" that the revision adds: only a counterpart with fewer negations
  was a suspect;
- a condition that the revision cuts off, when enough of the sentence's
  words remain.

For those the check had only its search for statements, a model that reads
two whole pages, which the first measurement had shown to miss damage.

## Decision

**A sentence is kept only when it stands in the revision word for word.** The
command looks for each sentence of the page near its place in the revision,
with the same words in the same order. Case, punctuation, markup, and
contractions do not count, and the sentence may stand inside a longer one or
run over the end of one into the next. So the edits that the report asks for
most often keep a sentence: two sentences joined with "and", a contraction, a
word put in front, an aside set in parentheses.

**The command reads every line of the page outside fenced code.** The first
build of this rule compared the prose that the engine extracts, and two holes
showed. The engine gives the prose of a paragraph in pieces around its code
spans, so a term that a reviser set as code left its sentence, and the judge
was shown a sentence with a hole in it. And the engine reads no prose in the
text of an image, in an indented block, or in a heading, so a change there
was never compared. The command now cuts the source itself: every line
outside fenced code blocks and front matter belongs to a block, and a block
is cut into sentences. Fenced code is compared as code, as before.

**A short sentence is compared by all its words.** ADR 0062 did not compare a
sentence with fewer than three content words, because two words are in some
other sentence by accident. "We did not run it for this post." has two, and
without its "not" it passed. Such a sentence is now compared word for word
like any other, and its counterpart is looked for by all its words.

**Every other sentence is a suspect and goes to the judge.** The revision
changed at least one of its words, and the command does not decide what the
change means. It names the kind of suspect as before: missing, when no
counterpart carries enough of its words; a negation lost; and now reworded,
for everything else. The judge gets each suspect beside the sentences of the
revision from its place.

This turns the rule around. ADR 0062 named what makes a sentence a suspect,
and whatever it did not name passed. Now the command names what makes a
sentence kept, and whatever it does not name is shown to the judge. Nothing
that one sentence says can change without that sentence being read beside
what replaced it.

**The judge still clears a suspect only with a quote** that stands in the
place it was shown, and for a missing sentence one that carries a word of the
sentence's own. One suspect that is not cleared rejects the revision.

**A sentence found without the words of its place is asked about again.**
One judge of the fitting half often said that a reworded sentence was kept
and quoted the sentence of the page that it had been asked about. Those words
are not in the place it was shown, so the sentence counted as not cleared: 13
of the 34 sentences that this judge did not clear, and the only reason for 5
of 16 rejections. The command now asks once more about such a sentence, and
only about such a one. A sentence that the judge called changed or missing is
not asked about again.

**A long list goes in questions of twenty.** An ordinary revision of the
fitting half has a median of 9 suspects with one reviser and of 14 with the
other, and a heavy one more than a hundred. A model errs at the end of a long
list.

## What the fitting half showed

The rule was fitted on the 40 posts of the fitting half of the style-guide
set and on nothing else. The
[record](../../research/reviews/2026-11-16-changed-sentences/RESULTS.md)
recounts these numbers.

**Without a model.** Each post got a copy of each kind of damage: a sentence
removed, "not" removed, "not" put in, one word replaced by its opposite, and
the last clause of a sentence cut off. The changed sentence is a suspect on
all 183 copies. On harmless copies, with two sentences joined or with
contractions, the comparison named no suspect on any of 75.

**With a judge.** On 50 damaged copies the changed sentence is a suspect that
the judge does not clear on 48, with either judge, and neither accepted a
damaged copy. Both judges cleared "above" where the page said "below" in a
reference to another part of the page, and each cleared one sentence that had
lost a clause with "when". The search for statements alone found a statement
not kept on 32 and on 15 of the 50 copies.

**On ordinary revisions.** The price of the rule is here. Of 40 revisions of
each reviser, 36 and 34 have a suspect, and the judge cleared 499 of 525
suspects and 711 of 744. A sentence that was not cleared stands in 9 and in
12 revisions; under the rule of ADR 0062 it stood in 5 and in 6 of the same
revisions. Read by hand, most of those sentences did lose a detail in the
revision.

## What was measured

MEASURED

## Consequences

- The `review` package gains `SentenceReworded` and `SuspectsPerQuestion`.
  `Suspects` returns more suspects than before, of three kinds.
- The command asks the judge about sentences on most revisions, in one
  question for every twenty suspects. Before, most revisions had no suspect.
- The check is stricter. A revision that says a sentence in other words is
  accepted only when the judge finds it and quotes it.
- The JSON report keeps the schema `unswell-compare-v2`; a suspect can now
  have the reason `reworded`.

## Limits

- The judge decides whether a reworded sentence still says the same. The
  command guarantees that the judge is shown the sentence and what replaced
  it; it does not guarantee what the judge reads.
- A sentence that the revision keeps word for word and moves to another
  place near by is kept, also when its new place changes what it refers to.
- Fenced code is compared as code and is not shown to the judge: a block with
  one changed line is reported as lost.
- What the page says across several sentences, such as the order of steps,
  is compared sentence by sentence only.
- The words are English, and the judges of the measurement are models.
