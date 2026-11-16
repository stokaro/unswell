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

**A sentence has the words of its code spans.** The engine gives the prose of
a paragraph in pieces: what stands between two code spans is a piece of its
own, and the code is in none. The command joins the pieces of one block again
with what stands between them. Otherwise a term that a reviser sets as code
would leave its sentence, and the judge would be shown a sentence with a hole
in it. Two such sentences of the fitting half were called missing for that
reason alone.

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

**A long list goes in questions of twenty.** An ordinary revision of the
fitting half has a median of FITMEDIAN suspects, and a heavy one more than a
hundred. A model errs at the end of a long list.

## What the fitting half showed

FITTING

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
- A sentence with fewer than three content words is not compared.
- A sentence that the revision keeps word for word and moves to another
  place near by is kept, also when its new place changes what it refers to.
- What the page says across several sentences, such as the order of steps,
  is compared sentence by sentence only.
- The words are English, and the judges of the measurement are models.
