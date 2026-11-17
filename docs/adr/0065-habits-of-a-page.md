# ADR 0065: Two rules for habits of a page

Status: accepted

## Context

A model that writes under a style guide leaves few of the marks that the
rules of `unswell check` count.
[ADR 0060](0060-rhythm-reading.md) gave the AI-sounding index a second
reading for such a page: runs of short plain sentences. The rules themselves
still had nothing to say about it. On one post of the style-guide set of the
earlier records, `unswell check` reported no finding, and one of three blind
judges took the post for a model's text in every pair.

The maintainer read that post and named what gives it away. Two of those
things can be counted, and no rule counted them:

- Paragraph after paragraph ends by saying what something is not, or what it
  does not do. One such sentence is an honest limit. Six of them are a
  device.
- The page keeps one formal register from its first line to its last: no
  contraction, no aside in parentheses, no question, and no sentence that
  opens with "And", "But", or "So".

A rule that ships on has to stay quiet on pages that people wrote
([ADR 0058](0058-default-rules-quiet-on-human-pages.md)): it may report on at
most one in twenty of them.

## Decision

1. **`syntax.closing-denial-density` is new and on by default.** It counts
   the paragraphs of two sentences or more whose last sentence denies, and
   reports the page once, with each such sentence, when more than two
   paragraphs end that way and they are at least 8% of those paragraphs.

2. **A denial has one of three forms**, read from the words of the last
   sentence: a last clause after a semicolon or a dash that opens with a
   negation; a tail ", not X"; a main clause with its negation among the
   first six words. A contraction of "not" is a negation, so a reviser who
   writes "doesn't" for "does not" changes nothing. A condition, a question,
   an instruction, a sentence that speaks to the reader, a limit such as "not
   only", and a status such as "not yet" are left alone.

3. **`syntax.uniform-register` is new and on by default.** It reports a page
   once, at its first sentence, when its running paragraphs have at least 25
   sentences and none of them has a casual mark: a question, an exclamation,
   a first word "And", "But", "So", "Or", or "Yet", a contraction, or
   parentheses that hold two words or more.

4. **Both are warnings with no gate and no weight in the score.** A page with
   either finding still passes the ordinary gate.

5. **Both report the page once.** The habit is a property of the page. One
   closing denial and one formal paragraph are fine, and a finding on each
   would teach a writer to remove the wrong thing.

6. **The evidence is a comparison of sets of pages.** The forms and the
   limits were chosen on the fitting halves. The bars were written on issue
   603 before the rules ran on the measured parts, and all 10 hold:

   | Pages of the measured parts | Pages | Paragraphs end on a denial | One formal register |
   | --- | ---: | ---: | ---: |
   | Posts of the style-guide set | 30 | 13 (43.3%) | 18 (60.0%) |
   | Blog posts that people wrote before November 2022 | 77 | 1 (1.3%) | 0 (0.0%) |
   | Articles that people wrote under a style guide | 205 | 1 (0.5%) | 4 (2.0%) |
   | Documentation pages that people wrote before November 2022 | 61 | 2 (3.3%) | 2 (3.3%) |
   | Texts that two models wrote under a style guide | 120 | 1 (0.8%) | 110 (91.7%) |

   At least one of the two reports 23 of the 30 posts.

## What was tried and dropped

**An echoed pair of words.** The post repeats a pair of words from the end of
one sentence at the start of the next. On the fitting halves people do that
more often than the posts of the style-guide set do: two such places or more
in 27% of the blog posts of people and in 18% of the posts, counted with a
prototype outside the engine. So it is no rule.

**The reference of what people remove.** [ADR 0056](0056-rules-tested-on-de-slop-commits.md)
admits a default rule when its findings are lower after a commit that says it
removes AI-sounding wording. These two rules are not confirmed there, and
they did not enter by that route. On the test part the first reports 54 of
985 files before the commit and 55 after it, and the second 41 and 42. The
files of that reference are mostly README pages; few have 25 sentences of
running prose or paragraphs that argue a case.

## Consequences

`unswell check` with no configuration now has something to say about a page
that a model wrote under a style guide: 9 in 10 of such texts get the
register finding, where the rules reported little before. It says more about
today's documentation too. Of 3,302 pages of public repositories, the first
rule reports 8.8% and the second 6.1%.

A project with reference pages or specifications turns
`syntax.uniform-register` off for them. People write such pages in one
register on purpose, and they are outside the measurement.

The catalog holds 62 rules. Revision 14 of the rule classes records both as
general style: the measurement compares sets of pages and does not say who
wrote one page.

## Not covered

Whether a post that is revised until both rules are quiet reads as a
person's. The [record of the blind check](../../research/reviews/2026-11-15-revision-check/RESULTS.md)
found that a revision by the report of the index lowers the index and does
not get there, and the same may hold here; a measurement with blind judges
has to say. People as judges. A language other than English. A denial whose
negation stands after the sixth word of its clause.

The [record](../../research/reviews/2026-11-17-page-habits/README.md) holds
the counts and recounts them, and [the documentation](../page-habits.md) says
what each rule reads.
