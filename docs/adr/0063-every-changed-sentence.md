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

**Each sentence takes its own place.** A page can say one sentence twice; a
text that a model wrote does it often. With a search that only asks whether
the words stand near by, the second copy of the sentence vouches for the
first, and the revision can change one of them. So a sentence of the page
takes the words that it matches, and no other sentence can take them again.
A sentence of the revision whose words are all taken is the counterpart of no
other sentence either: the judge is shown the changed copy, not the kept one.

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
the judge does not clear on 48 with one judge and on 49 with the other. One
judge cleared "above" where the page said "below" in a reference to another
part of the page, and each cleared one sentence that had lost a clause with
"when". One of those copies was accepted: the judge also heard it as better.
The search for statements alone found a statement not kept on 32 and on 14 of
the 50 copies.

**On ordinary revisions.** The price of the rule is here. Of 40 revisions of
each reviser, 36 and 34 have a suspect, and the judge cleared 510 of 537
suspects and 734 of 773. A sentence that was not cleared stands in 9 and in
14 revisions; under the rule of ADR 0062 it stood in 5 and in 6 of the same
revisions. Read by hand, most of those sentences did lose a detail in the
revision. Four are lines that stand for a code block in the texts of the
measurement, which a revision has fewer times than its page.

## What was measured

The bars were written on issue 595 before the run. The run was started three
times. The first two starts were stopped when their model-free stage showed
damaged copies whose changed sentence was no suspect: short sentences and
text that the engine reads as no prose the first time, a page that says its
sentences twice the second. Each time the rule was changed, as the sections
above say, and the issue has a comment for each stop with what had been seen.
The numbers below are those of the third start, on new copies.

Every bar was read on two kinds of pages: the 30 posts of the measured half
of the style-guide set, which earlier measurements had read, and 30 texts
that two models wrote under a style guide, which no judge had read. Of the 32
bars, 31 hold.

| What | Posts | Fresh texts |
| --- | --- | --- |
| The changed sentence is a suspect, without a model | 96 of 96 | 102 of 102 |
| Harmless copies with a suspect, without a model | 0 of 56 | 0 of 41 |
| The check reports the damage, judge `gpt-6-astra` | 96 of 96 | 101 of 102 |
| The check reports the damage, judge `private-host/Qwen3.6-35B-A3B` | 95 of 96 | 97 of 102 |
| Damaged copies accepted, the two judges | 0 and 0 | 0 and 2 |

By kind of damage, the two kinds of pages together:

| Judge | A sentence removed | "not" removed | "not" put in | A word turned around | A clause cut off |
| --- | --- | --- | --- | --- | --- |
| `gpt-6-astra` | 60 of 60 | 14 of 15 | 53 of 53 | 54 of 54 | 16 of 16 |
| `private-host/Qwen3.6-35B-A3B` | 60 of 60 | 14 of 15 | 53 of 53 | 54 of 54 | 11 of 16 |

**The bar that does not hold.** Each kind of damage had to be reported on at
least 80% of its copies. A sentence that lost its last clause was reported on
11 of 16 with one judge. The command did its part: the sentence was a
suspect, and the judge was shown it beside what was left of it. The judge
found the sentence there and quoted the words that remained. Both damaged
copies that the check accepted are of this kind.

**The statements.** The check finds 100.0% and 99.4% of the statements kept
over the revisions of the posts, and 99.7% and 98.9% over those of the fresh
texts.

**Ordinary revisions.** Almost every revision has a suspect, and the judge
clears most: 260 of 265 and 471 of 507 on the posts, 606 of 618 and 852 of
917 on the fresh texts. A sentence that it did not clear stands in 5 and in
13 of 27 revisions of posts, and in 7 and in 19 of 30 revisions of fresh
texts. No bar was set for this.

The [record](../../research/reviews/2026-11-16-changed-sentences/RESULTS.md)
recounts these numbers. The published binary of `v0.1.0-alpha.16` names the
same suspects on all 553 copies of the record, and it reaches the recorded
verdict on the 57 revisions that could be replayed from kept answers of a
judge; see the [alpha.16 evidence](../release/v0.1.0-alpha.16.md).

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
  it; it does not guarantee what the judge reads. One of the two judges found
  a sentence in what was left of it after its last clause was cut, on 5 of 16
  copies.
- When a page says a sentence twice and the revision changes one copy far
  from the other, the place that the judge is shown can hold the kept copy.
  One damaged copy of 198 went unreported for that, with each judge.
- A sentence that the revision keeps word for word and moves to another
  place near by is kept, also when its new place changes what it refers to.
- Fenced code is compared as code and is not shown to the judge: a block with
  one changed line is reported as lost.
- What the page says across several sentences, such as the order of steps,
  is compared sentence by sentence only.
- The words are English, and the judges of the measurement are models.

[ADR 0064](0064-words-left-out-of-a-sentence.md) answers the first two
limits: the command names the words that a revision leaves out of a sentence,
and it no longer shows the judge a kept copy of a sentence.
