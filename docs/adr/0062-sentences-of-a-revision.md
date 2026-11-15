# ADR 0062: The sentences of a page are compared with its revision

Status: accepted

## Context

[ADR 0061](0061-blind-check-of-a-revision.md) gave the product a check of
one revision. Its search for statements is a model that lists the statements
of the page and looks for each in the revision. The
[record of that decision](../../research/reviews/2026-11-15-revision-check/RESULTS.md)
showed two faults:

- **It misses damage.** On copies of a post with one sentence removed, or
  with the word "not" removed from one sentence, the search reported a
  statement as not kept on 39 of 58 copies with one judge and on 48 of 58
  with the other. The check accepted 3 and 2 damaged copies.
- **It counts a careless quote as a loss.** One judge called 53 statements
  kept with words that are not in the revision. They counted as not kept, and
  that judge found 97.8% of the statements kept, under the bar of 98%.

Both come from one cause. Whether a sentence of the page is still in the
revision was left to a model that reads two whole pages.

## Decision

**The command compares the sentences itself.** For each sentence of the page
it looks for a counterpart in the revision: one sentence near the same place,
or two or three in a row, with at least six tenths of the content words of
the sentence. The share is counted over all its content words and over its
own words, which are those that the two sentences before it and the two after
it do not have. Words are compared in lower case and without their common
endings; function words do not count. A sentence without a counterpart is a
suspect, and so is one whose counterpart has fewer negations. A sentence with
fewer than three content words is not compared.

This needs no model, and the same pair always gives the same suspects.

**Own words, because the sentences around a removed one share its subject.**
On the fitting half the comparison by all content words named the removed
sentence on 32 of 40 damaged copies. The eight it missed were sentences
whose words stand in the sentences around them. Counted over their own words
as well, and near their place, it named 37 of 40, and 39 of 40 when a
paragraph was first cut into its sentences.

**The engine's items are cut into sentences.** The prose of a paragraph up to
a code span is one item of the engine and can hold several sentences. One of
them can go and the item still carries most of its words.

**The model is asked about the suspects, and only about them.** Most suspects
in an ordinary revision are harmless. A reviser writes "rather than" where
the page said "not", or says a sentence in words of its own. Rejecting on a
suspect alone would reject 9 of 40 and 17 of 40 ordinary revisions of the
fitting half. So the judge gets each suspect beside the sentences of the
revision from its place and says whether they still say it. It sees two
short texts, where the search for statements shows it two pages.

**The check, not the model, clears a suspect.** A "yes" counts when the
quoted words stand in the place that the model was shown. For a missing
sentence they also have to carry one of the sentence's own words. On the
fitting half one judge cleared a removed sentence with words of the sentence
beside it; those words stand in the place and are no word of the sentence
that is gone.

**One suspect that is not cleared rejects the revision.** The search for
statements allows 2% of losses, because a model that lists and looks up
statements errs. A sentence that the command cannot find and the judge cannot
quote is a concrete loss, and the report shows it with its line.

**A statement that was not found is asked again.** The command repeats the
question for those statements only, and a statement that is quoted properly
the second time is kept. A model errs on a long list. One judge quoted words
that are not in the revision for 53 statements of the earlier record. On the
fitting half the same judge called 22 of the 90 statements of one revision
changed, most of them at the end of the list; the other 15 revisions it had
read by then kept every statement.

## What the fitting half showed

The rule was fitted on the 40 posts of the fitting half of the style-guide
set and on nothing else. The
[record](../../research/reviews/2026-11-15-sentence-check/RESULTS.md) recounts
these numbers.

**Without a model.** Each post got a copy with one sentence removed and a
copy with the word "not" removed from one sentence, twice, so that the second
copies damage other sentences than the first. The removed sentence is a
suspect on 39 of 40 and on 38 of 40 copies, and the sentence without its
"not" on 38 of 38 and on 34 of 34. On harmless copies the comparison named no
suspect: none of 40 copies with two sentences joined, and none of 35 with the
negations written as contractions.

**With a judge.** On the first damaged copies the damaged sentence is a
suspect that the judge does not clear on 77 of 78 copies with either judge,
and neither accepted a damaged copy. The search for statements alone reported
a statement as not kept on 75 of those copies with one judge and on 47 with
the other.

**On ordinary revisions.** The check finds 1,798 of 1,803 statements kept for
the 40 revisions of one reviser and 3,657 of 3,664 for the 40 of the other.
Of the 18 suspects in the revisions of the first reviser the judge cleared 7,
and of the 47 in those of the second it cleared 38. A sentence alone rejects
5 revisions of each.

## What was measured

The bars were written on issue 589 before the run. The style-guide set has no
post that no run had read: the 30 posts of its measured half were read once
by the record of ADR 0061. So every bar was read twice, on those 30 posts with
new copies and on 30 texts that no run had read, which two models wrote under
a style guide. All 22 bars hold.

| What | Posts of the style-guide set | Texts that no run had read |
| --- | --- | --- |
| Damaged copies where the sentence is a suspect, without a model | 56 of 56 | 41 of 41 |
| Harmless copies with a suspect, without a model | 0 of 56 | 0 of 41 |
| Damaged copies that the check reports, with `gpt-6-astra` as the judge | 56 of 56 | 41 of 41 |
| The same, with `private-host/Qwen3.6-35B-A3B` | 56 of 56 | 41 of 41 |
| Damaged copies that the check accepts, either judge | 0 | 0 |

The statements. Over the revisions of the earlier record the check finds
3,033 of 3,039 statements kept with `private-host/Qwen3.6-35B-A3B` as the
judge, 99.8% where the earlier build found 97.8%, and 3,201 of 3,216 with
`gpt-6-astra`, 99.5%. Over revisions of the fresh texts it finds 897 of 902
and 1,587 of 1,599.

The sentences of ordinary revisions. Of the 86 and 57 revisions of the
earlier record, 18 and 25 have a suspect; the judge cleared 18 of 25 suspects
and 49 of 68, and a sentence that was not cleared stands in 7 and 10
revisions. No bar was set for this.

During the run the program of one judge and the endpoint of the other failed,
and 171 pairs ended without a verdict. Issue 589 set one retry for such pairs
while the run was in progress. After it every pair has a verdict; the record
shows both counts.

The published binary of `v0.1.0-alpha.15` names the same suspects on all 421
copies of the record, and it reaches the recorded verdict on the 30 pairs
that could be replayed from kept answers of a judge; see the
[alpha.15 evidence](../release/v0.1.0-alpha.15.md).

## Consequences

- The JSON report of `unswell compare` has the schema `unswell-compare-v2`
  and a `sentences` part: how many sentences were compared, and each suspect
  with its line, the place it was looked for, its own words, and the judge's
  answer. The protocol is `unswell-revision-v2`.
- `review.Judge` takes the suspects as a third argument. The package gains
  `Suspects`, `SuspectsPrompt`, `ParseSuspects`, `BindSuspects`, and
  `LostSentences`.
- The command makes up to seven calls where it made five.

## Limits

- The comparison reads words. A revision that keeps the words of a sentence
  and changes what it says, with "before" where the page said "after", has
  its counterpart. The search for statements may notice it, and numbers are
  checked apart.
- A sentence whose words all stand in the sentences around it can go
  unnoticed. Then the page still says most of what the sentence said.
- A paraphrase that keeps none of a sentence's own words is not cleared,
  whatever the judge says. The revision is rejected and the report names the
  sentence.
- The words are English. Function words, negations, and endings are lists of
  English words.
- The judges of the measurement are models. No person took part.
