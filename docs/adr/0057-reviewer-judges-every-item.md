# ADR 0057: The reviewer judges every item

Status: proposed on November 13, 2026. It takes effect when the maintainer
merges it.

## Context

`unswell review` asked the reviewing model for the items that need revision
and nothing else. A model that answers this way picks part of what it could
flag, and picks differently on each run. Two runs of `claude-sonnet-5-5` on
the same 40 files repeated 77.4% of each other's flagged sentences.

That is a defect for a tool that people run again after they edit.

It was first taken for the cause of a failed measurement: on files before and
after a commit that removes AI-sounding wording, the review reports about as
many findings after the commit as before it. It is not the cause; see the
consequences.

## Decision

1. The reviewer answers with a verdict for every numbered item: `ok` or
   `revise`. An item marked `revise` carries the categories, the quote, and
   the reason as before.

2. An answer that judges fewer than four in five items is a failed answer and
   is asked again. A list that stops early is not a clean document.

3. `Document.ParseFlags` replaces `review.ParseFlags`: it needs the items to
   count the verdicts. `Document.Bind` and the report are unchanged.

4. The MCP tools keep the list form. There the calling assistant submits
   flags through a tool call, and the tool cannot ask it for every item.

## Evidence

The comparison was registered on issue 525 before it ran, on 40 files that
no earlier run of these forms with this reviewer had touched. Each form ran
twice on every file.

| Answer asked for | Sentences flagged, two runs | In both | Repeat rate | Changed words flagged | Their share of the flagged words | Same share of all words | Difference |
| --- | ---: | ---: | ---: | ---: | ---: | ---: | --- |
| A verdict for every item | 181 and 179 | 151 | 83.9% | 636 | 40.4% | 27.6% | +12.8 points (+3.4 to +21.9) |
| The same, "revise" only when the defect is clear | 99 and 102 | 85 | 84.6% | 349 | 42.0% | 26.9% | +15.1 points (+4.7 to +24.9) |
| Only the items that need revision | 121 and 122 | 94 | 77.4% | 439 | 39.3% | 27.6% | +11.7 points (+2.3 to +21.0) |

The verdict form repeats +6.5 points (+1.4 to +13.3) more of its flagged
sentences than the list form. Its flags fall on words that the commit changed
more often than chance, and it flags more of those words: 636 against 439.

A stricter form was registered first. It tells the reviewer to use `revise`
only when the defect is clear. It repeats as well (+7.2 points (+0.7 to +16.1)
against the list form) and finds less than the list form does: 349 changed
words. It was dropped for the form above, which was registered as a third arm
on the same files before it ran.

## Consequences

The answer grows with the document, about a line per sentence: about 490
tokens where it was about 190, against about 3,500 tokens of prompt.

The gain is partial. About one flagged sentence in six still does not come
back on a second run.

The findings are still not shown to fall after a de-slop commit (21 pairs
lower, 15 higher), and repeating better did not change that. The record shows
why. Of the 174 sentences flagged before the commit, the commit changed or
removed 100 (57.5%), where it changed 33.0% of all sentences. Of the 139
sentences flagged after the commit, 64 are new or changed text, 56 stood
unchanged and were flagged before, and 19 stood unchanged and were not. The
review points at the sentences that people go on to change. The text they
write in their place is flagged again, under a rubric of lean wording that a
de-slop edit does not aim at.

The review finds a small part of the edit: 100 of the 644 sentences that the
commits changed or removed (15.5%).

The tables in `docs/review.md` that compare the review with human edits and
with another model's annotation were measured with the list form. They stay,
marked as such, until they are measured again.

One model carries the result. The same runs with a second, weaker model,
`private-host/Qwen3.8-27B`, show no gain: the verdict form repeats 67.0%
of its flagged sentences where the list form repeated 67.4% (-0.4 points (-15.2 to +11.6)),
and it flags 285 changed words where the list form flagged 319. The registration
set its bars for the first model and reported the second without one. The
change goes in on the first model's result; a project that reviews with a
weaker model pays the longer answer and may get nothing for it.
