# ADR 0066: `unswell compare` gives two results

Status: accepted

## Context

[ADR 0061](0061-blind-check-of-a-revision.md) made `unswell compare` accept a
revision only when two things hold. A judge names the page, in both orders of
a blind comparison, as the text that reads more like a model's. And nothing
is lost: the statements, the sentences, the numbers, code, and links are
kept, and no em dash is added. The command gave one verdict for both.

The two parts are not equally steady. Nothing lost rests on what the command
compares itself, and on a judge that has to quote the words of the revision.
The blind comparison rests on one question, asked in each order, and a
revision is better only when the judge gives the rarer answer both times.

The records of ADR 0063 and ADR 0064 ran the same revisions twice, and the
runs can be compared
([issue 596](https://github.com/stokaro/unswell/issues/596)). On posts, where
a judge mostly hears no difference, a "better" from one run came again 3
times of 8 with one judge and 0 of 3 with the other. So the same revision of
a post passes in one run and fails in the next, and the check accepted fewer
than one revision of a post in ten. A cleanup of many pages cannot be gated
on that.

## Decision

1. **The command gives two results.** "Nothing lost" passes when at least 98%
   of the statements are kept, every sentence that the comparison names is
   found, no number, code, or link is lost, and no em dash is added. "Reads
   less like a model's text" passes when the judge names the page in both
   orders. Each has its own line in the text report and its own field in the
   JSON report, with its reasons.

2. **`--require` says what a revision has to pass.** With `all`, the default,
   it is both results, as before. With `kept` it is nothing lost alone: the
   blind comparison neither rejects the revision nor gives a reason.

3. **The judge is asked about the sound in either case.** The report says
   what it heard. A command that skipped the question would hide a revision
   that reads worse.

4. **The default stays `all`.** A reader of the word "accepted" in an earlier
   record reads the same rule, and a caller who asks for nothing gets the
   stricter check.

5. **An unknown requirement fails before anything is sent.** It is an error
   of the options, with exit code 2, and no verdict.

6. **A quote is compared without regard to case, and a contraction of "not"
   is read as its words.** A statement is kept when the judge quotes words
   that stand in the revision. On the fitting half, 5 of 35 copies that only
   write negations as contractions were rejected for a statement whose quote
   had "does not" where the copy has "doesn't". The comparison of sentences
   already reads the two as the same words.

7. **The JSON report has the schema `unswell-compare-v3`.** The verdict holds
   `require`, `accepted` with its `reasons`, and `kept` and `sound`, each
   with `passed` and its own `reasons`. `review.Judge` takes the requirement
   and returns a `RevisionVerdict` with both results.

## Consequences

A caller that needs the same answer on the next run asks for `--require kept`
and gets a check of what a revision lost. A caller that wants the judge's ear
too keeps the default and knows that it may not repeat.

The blind comparison is no steadier than it was. This decision takes it out
of the gate for those who ask; it does not make it repeat.

With `kept` required, a revision that reads more like a model's text than the
page is accepted when it loses nothing. The report says so on its own line.

## Measurement

The bars were written on
[issue 605](https://github.com/stokaro/unswell/issues/605) before the run on
the measured part. Eleven of the 14 can be read, and all eleven hold.

- The result on nothing lost is the same in two runs on 55 and on 49 of 57
  revisions, with the two judges.
- Of 155 harmless copies, 144 lose nothing with the judge that saw them all.
  No damaged copy loses nothing: 0 of 219 with that judge.
- With `kept` required the check accepts 19 and 41 of 57 revisions in a first
  run, where it accepts 6 and 12 with `all` required.

The run of the copies with `gpt-6-astra` was stopped after 198 of 374 copies,
when that judge was no longer available. Three of its bars are not read, and
the issue keeps them open. The
[record](../../research/reviews/2026-11-17-two-results/README.md) has the
counts on what it did judge.

## Not covered

A blind comparison that repeats, for example one that asks several times.
Whether a revision that loses nothing is worth publishing. People as judges.

[The documentation](../compare.md#two-results) says how to choose between the
two requirements.
