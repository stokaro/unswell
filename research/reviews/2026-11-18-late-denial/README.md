# The fourth form of a closing denial

November 18, 2026. `syntax.closing-denial-density` reports a page whose
paragraphs end by saying what something is not. It read a denial in three
forms, and one of them needs the negation among the first six words of the
main clause. A closing sentence with a long subject was not counted:

```text
The list of pages that the crawler reads at night is not a map of the whole site.
```

This record measures a fourth form, for a negation that follows its verb
later in the clause.

[RESULTS.md](RESULTS.md) holds the numbers, and `make check-late-denial`
recounts every one of them from the data here. The bars were written on
[issue 607](https://github.com/stokaro/unswell/issues/607) before the form
ran on the measured parts. [Habits of a page](../../../docs/page-habits.md)
says what the rule reads, and
[ADR 0065](../../../docs/adr/0065-habits-of-a-page.md) has the decision.

## Outcome

All 7 bars hold, two of them at their limit.

- **The form finds something.** On the 30 measured posts the rule counts 79
  closing denials where it counted 68, 16% more. It reports 14 of the posts
  where it reported 13.
- **Quiet on people.** With four forms the rule reports 1 of 77 blog posts, 3
  of 205 articles written under a style guide, and 3 of 61 documentation
  pages. Each is within one page in twenty, the limit for a rule that ships
  on. The documentation pages are at 4.9%: a fourth page would pass it.
- **What it costs.** Of the 343 pages of people, 3 are reported with four
  forms and were not with three. The bar allowed 3.
- **Nothing is lost.** No page that three forms report is quiet with four,
  and on all 4,947 pages of the record no closing denial that three forms
  count is missing with four.

Without a bar: of the 120 fresh model texts the rule reports 1, as before.
Of 3,302 pages of public repositories today it reports 357 where it reported
289, which is 10.8% and 8.8%. Of 455 pages of the panel corpus it reports 23
where it reported 20.

## The second use of the measured parts

The measured parts are those of the
[record of the two page habits](../2026-11-17-page-habits/README.md), and
that record ran the rule on them once, with three forms. This run is their
second use, with a changed rule. So a bar here says less than a bar on pages
that no run has read.

What the first run left behind is a count for each set: how many of its
pages the rule reports. No sentence of a measured page was read to choose
the fourth form. One thing was known before, as in the earlier record: the
post that gave the idea for the rule is in the measured half, and the issue
started from the closing sentences that the rule does not list on it.

## How the form was chosen

On the fitting parts and on nothing else, with the sentences that the form
adds read one by one.

The first definition is the one of the issue: a negation after an auxiliary
near the end of the clause, with no word of a condition before it. It
reported 7 of the 228 fitting articles of people, where three forms report
1. Most of what it added there was a negation in a smaller clause, of the
kinds "the rows that do not match", "to make sure that the file is not
changed", and "which may not be there".

So the form leaves alone a negation whose verb stands in a relative clause.
"That", "which", or "who" before the verb opens such a clause, and the clause
has closed only when a finite verb of its own stands between: in "the pages
that the crawler reads at night is not a map", "reads" closes it. For this
the rule reads the parts of speech of the words. The form also leaves alone a
negation after
"what", "why", "how", or "where", the verbs "may", "might", "should", and
"must", "not ... until", and a negation inside quotation marks.

With that, the rule reports 1 of the 228 articles, 2 of 74 blog posts where
it reported 1, and 24 of the 40 fitting posts where it reported 22.

## Builds

Four forms: commit `fba3e2ed` of the branch of this change. Three forms: the
published binary of `v0.1.0-alpha.19`. The rule is on and every other rule is
off. A page is one run of `unswell check --stdin` on its text, once with the
limits of the rule as they ship and once with no limit, which gives every
closing denial that the rule counts.

The counts with three forms are those of the earlier record on each of the
4,947 pages.

## Data

| File | Rows |
| --- | --- |
| `data/pages.jsonl.gz` | each page of the twelve sets: its set, its part, an id, and for three forms and for four the number of closing sentences that the rule lists (0 when it is quiet) and the number that it counts with no limit; and how many of the sentences that three forms count are counted with four |

The record holds no text, path, repository, or title. A post of the
style-guide set has the number that it has in the earlier records. Every
other page has the hash that it has in the record of the two page habits.

`scripts/forms.py` runs the two builds, and `scripts/export.py` writes the
data. They need the texts, which stay with the maintainer.

## Not measured

Pages that no run has read. Whether a person agrees, sentence by sentence,
that what the form adds is a denial: on the fitting posts one of the twelve
added sentences has its negation in a clause without "that", of the kind
"the files the runner does not read", which the form cannot see. Reference
pages and specifications. A language other than English.
