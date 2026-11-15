# ADR 0061: A blind check of a revision

Status: accepted

## Context

[ADR 0060](0060-rhythm-reading.md) gave the page index a second reading and
the report a remedy for each run of sentences. Its record compared a page
with its own revision: over many pages, three blind judges more often named
the page than the revision as the text that reads more like a model's.

That left two things open.

The index cannot tell whether one revision helped. It falls after almost any
edit of a run of sentences. On the fitting half of the
[record of this decision](../../research/reviews/2026-11-15-revision-check/README.md)
a report that asked for longer sentences took the median index of 40 posts
from 69 to 23, and the panel of three judges found the revision more like a
model's text than the post it came from on 24 of the 40. A writer who watches
the index would have kept those revisions.

And nobody had compared a revised post with text that a person wrote. A post
can read less like a model's than it did and still be told from a person's.

## Decision

**A command checks one revision.** `unswell compare <page> <revision>` asks a
model that the user names. A revision is accepted only if all of this holds:

1. Sound. The judge reads the opening of both texts, whole blocks up to about
   450 words, as text A and text B. The question does not say that one is a
   revision of the other. It is asked in both orders, and the judge has to
   name the page as the text that reads more like a model's both times.
2. Statements. The model lists the facts and the instructions of the page. It
   then looks for each in the page itself and has to quote the words that
   state it; a statement that it does not find there is left out. It then
   looks for the others in the revision, again with a quote. A statement is
   kept only when the quoted words stand in the revision, which the command
   checks itself. At least 98% have to be kept.
3. No number of the prose, no code, and no link of the page is missing from
   the revision. This needs no model.
4. The revision has no more em dashes in its prose than the page.

The command exits with 0 for an accepted revision, 1 for a rejected one, and
2 when the comparison did not complete. The index plays no part.

**Both orders, because one answer says little.** On the fitting half one
judge was asked about 40 posts and their revisions in both orders. It named
the second text both times on 30 of them: the page once and the revision
once. Another judge did so on 27 of 40. A judge that follows the order of
the texts gives an answer that looks like a preference. The command reads
two answers that name different texts as no difference.

**The report says not to add an em dash.** A reviser that rewrites a run
often joins two clauses with an em dash, and the first reading counts it.
With the report of `v0.1.0-alpha.13` the two revisers added 27 and 94 em
dashes to 36 and 34 posts. With one more line of advice, "Do not add an em
dash anywhere: an em dash is one of the marks that this report counts", they
added 0 and 1 to 40 posts each.

**The report does not ask for longer sentences.** People's articles have
longer sentences than the posts of the product blog, and a report can say so
with the figures of the page. A reviser follows it: the share of sentences
over 20 words rose from 8% to 36%. All three judges then found the revision
more like a model's text than the post, on 23, 30, and 24 of 40.

**A vote by section is not used.** A page and its revision were cut at their
headings, and each section was asked in both orders. One judge was unsure on
217 of 245 sections and found 11 revisions better and 11 worse. For the
other reviser, the vote called 8 of 20 revisions better; the panel of three
judges had found 5 of those 8 more like a model's text than the post.

## What the fitting half showed about a post against human text

Each of the 40 posts of the product blog against three human blog posts, as
it is and after a revision of every post:

| Revision | Panel picks the post | `gpt-6-astra` alone |
| --- | --- | --- |
| None | 28 of 40 | 40 of 40 |
| By the report of `v0.1.0-alpha.13`, reviser `gpt-6-astra` | 26 of 40 | 40 of 40 |
| The same, reviser `private-host/Qwen3.6-35B-A3B` | 28 of 40 | 40 of 40 |
| By a report with sentence lengths | 28 of 40 | 40 of 40 |
| A free rewrite "as an engineer writes for colleagues" | 22 of 40 | 40 of 40 |

One judge tells every post from a human post, whatever was done to it.
Against articles that people wrote under a style guide it tells 39 of 40.
MEASURED

## Consequences

- A fourth command sends text to a model that the user names, beside
  `review` and `trim`. Its JSON report has the schema `unswell-compare-v1`.
- The `review` package gains the prompts, the parsers, and the rule of the
  check: `NewRevision`, `SoundPrompt`, `ParseSound`, `SoundVerdict`,
  `FactsPrompt`, `ParseFacts`, `ControlPrompt`, `KeptPrompt`, `ParseKept`,
  `Confirmed`, `BindFacts`, `Lost`, `EmDashes`, `Judge`, and `Opening`.
- The text report of `unswell check` has one more line when it lists a run of
  sentences.

## Limits

LIMITS
