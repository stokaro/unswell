# ADR 0056: Rules are tested on what people remove when they de-slop documentation

Status: proposed on November 13, 2026. It takes effect when the maintainer
merges it.

[ADR 0058](0058-default-rules-quiet-on-human-pages.md) later turned
`format.decorative-emoji` off outside the strict profile: it reports on more
than one in twenty pages that people wrote and is not confirmed on
likely-manual edits. The measurements below stand as they were made.

## Context

The tool is for documentation that reads as written by a model. Until now no
measurement used such documentation. The human references were chat answers,
fiction, and pages that people wrote
([ADR 0052](0052-claims-follow-human-references.md),
[ADR 0054](0054-trim-measured-against-editors.md)).

Commit history has the missing case. People commit with messages such as
"de-slop README" or "remove AI-isms from the docs". The file before such a
commit is documentation that someone found AI-sounding, and the file after it
is what they made of it.

## Decision

1. **The reference.** A pair is one Markdown file before and after a commit
   whose message states the removal of AI-sounding wording and that carries no
   assistant's co-author trailer. Repositories are split by hash into a
   development part (959 pairs from 648 repositories) and a test part
   (597 pairs from 423).

2. **The measure.** For a set of rules, the findings per 1,000 words are
   compared between the two versions. Among the pairs that differ, the share
   in which the findings are lower after the commit is the measure. Rules
   unrelated to the edit give 50%.

3. **A rule enters the defaults on this reference** when that share is at
   least 65% on the test part, with the lower end of the 95% interval above
   55%. The bar was registered on issue 525 before the test part ran.

4. **`format.decorative-emoji` is new and on by default.** It reports emoji in
   Markdown headings and list items. On the test part its findings are lower
   after the commit in 76.1% (68.2% to 82.5%) of the pairs that differ,
   and they fall from 1,408 to 424. That is the rule as registered, which also
   counted check marks and crosses. Those were then taken out: people use them
   in checklists, and on the development part they are removed about as often
   as any word. The rule as it ships gives 77.7% (69.8% to 84.0%) on the same
   pairs, measured after the first reading.

5. **`format.em-dash-density` stays off outside the strict profile.** It gives
   62.5% (51.0% to 72.8%), below the bar. It looks at one paragraph, and
   few paragraphs hold enough em dashes to pass its limits.

6. **`format.em-dash-frequency` is new and on by default.** It looks at the
   page. It reports a document once when its paragraphs, headings, and list
   items hold at least three em dashes and at least three per 1,000 words, and
   it lists each of them. It was registered after the emoji rule had been read
   on the test part, so it is a second use of those pairs. Its measure is the
   em dashes it lists per 1,000 words: lower after the commit in
   70.4% (64.1% to 76.0%) of the pairs that differ, and they fall from
   3,790 to 2,618. Its bar had one more condition, registered with it: the
   rule may flag at most 5% of the pages that people wrote before November
   2022. It flags 2 of 121.

7. **Neither rule has a gate or a weight in the score.** Each is a warning.

## Result for the default rules

| Default rules | Test pairs where they fire | Findings before | Findings after | Lower after, among pairs that differ |
| --- | ---: | ---: | ---: | --- |
| v0.1.0-alpha.9 | 158 of 597 | 380 | 334 | 64.0% (56.1% to 71.2%) |
| With `format.decorative-emoji` | 253 of 597 | 1,666 | 709 | 71.0% (64.9% to 76.3%) |
| With both new rules | 365 of 597 | 1,887 | 857 | 72.7% (67.8% to 77.1%) |

In a set of rules the em dash rule counts as one finding per document. All
three rows are above 50% with an interval that excludes it. This is the first
measurement in which the default rules follow what people change, and it is on
documentation. The new rules widen the reach: the defaults fire in
365 pairs where they fired in 158.

## How often the new rules report

| Texts | Pages | Emoji: pages with a finding | Emoji: findings per 1,000 words | Em dash: pages flagged | Em dash: listed per 1,000 words |
| --- | ---: | ---: | ---: | ---: | ---: |
| Pages that people wrote before November 2022 | 121 | 5.0% | 0.16 | 1.7% | 0.06 |
| Pages of public repositories today | 3,302 | 11.0% | 0.59 | 24.1% | 2.40 |
| Pages that three models wrote or polished for the panel corpus | 455 | 0.0% | 0.00 | 2.2% | 0.35 |
| Texts before a de-slop commit | 1,608 | 19.8% | 1.45 | 37.8% | 5.10 |

The em dash rule reports 24.1% of the pages of today's public repositories.
Who wrote those pages is unknown. Their share lies between the pages that
people wrote before 2022 and the texts that someone later de-slopped. Stricter
limits were tried on the development part: they lower all of these shares in
step and leave the measure where it is, so the registered limits stay. A
project that wants fewer reports raises `onset` or `allowed_occurrences`.

Neither rule detects a model. On the pages of the panel corpus, which three
models wrote or polished, the em dash rule flags 2.2% and the emoji rule
reports nothing. The rules follow what a de-slop commit takes out of a page
that someone found AI-sounding.

For the emoji rule the limit on text that people wrote was not part of the
registered bar. It was checked afterwards, and it would have turned the rule
off had the two rates been close.

## Consequences

A rule can now be tried against what people do to AI-sounding documentation
before it ships. The development part also shows which existing rules follow
nothing there: passive density, noun stacks, parenthetical load, and paired
contrasts sit at 50% to 56%. They are off by default already.

The reference has limits.

Who typed a commit is unknown. A commit without an assistant's trailer can
still have been made by an assistant that was told to de-slop. Its message
states a person's intent, and a person kept the result. Commits made in
GitHub's web editor were most likely typed by a person, and there are only
78 such pairs over both parts. They confirm neither rule: the emoji count is
lower after the commit in 8 of 18 pairs that differ, and the em dash count in
14 of 22. So the result is about commits that say they remove AI-sounding
wording. It is not shown for edits that a person certainly typed.

The pairs were not selected by language, and the tool is for English. 97 of
the 597 test pairs are not English prose. On the 500 English test pairs both
new rules still meet the bar: the emoji rule gives 77.3% (67.5% to 84.8%)
and the em dash rule 72.0% (65.3% to 77.9%). The defaults
fire in 300 of those pairs where the defaults of v0.1.0-alpha.9 fire in 141,
and give 72.9% (67.5% to 77.7%). This cut was made after the first
reading. The emoji rule fires more often on the files in other languages, so
part of its reach in the table above lies outside the tool's scope.

The commits were found by their message, so they show what is removed when
someone sets out to de-slop. A rule that counts emoji or em dashes says
nothing about a text that has none: the defaults fire in
365 of 597 test pairs.

The [record](../../research/reviews/2026-11-12-human-reference/RESULTS-words.md)
holds the counts of every pair, and its check recounts these numbers.
