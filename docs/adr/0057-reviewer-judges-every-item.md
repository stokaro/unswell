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
   the reason as before. The prompt says to use `revise` only when the defect
   is clear.

2. An answer that judges fewer than four in five items is a failed answer and
   is asked again. A list that stops early is not a clean document.

3. `Document.ParseFlags` replaces `review.ParseFlags`: it needs the items to
   count the verdicts. `Document.Bind` and the report are unchanged.

4. The MCP tools keep the list form. There the calling assistant submits
   flags through a tool call, and the tool cannot ask it for every item.

## Evidence

The comparison was registered on issue 525 before it ran, on 40 files that
no run of either form with this reviewer had touched. Each form ran twice on
every file.

| Answer asked for | Sentences flagged, two runs | In both | Repeat rate | Flagged words that the commit changed | Same share of all words | Difference |
| --- | ---: | ---: | ---: | ---: | ---: | --- |
| A verdict for every item | 99 and 102 | 85 | 84.6% | 42.0% | 26.9% | +15.1 points (+4.7 to +24.9) |
| Only the items that need revision | 121 and 122 | 94 | 77.4% | 39.3% | 27.6% | +11.7 points (+2.3 to +21.0) |

The verdict form repeats +7.2 points (+0.7 to +16.1) more of its flagged sentences. Its
flags fall on words that the commit changed more often than chance. Both bars
of the registration hold.

## Consequences

The answer grows with the document, about a line per sentence, so a review
costs more output tokens.

The gain is partial. About one flagged sentence in six still does not come
back on a second run.

The findings are still not shown to fall after a de-slop commit (17 pairs
lower, 15 higher), and repeating better did not change that. The record shows
why. Of the 95 sentences flagged before the commit, the commit changed or
removed 57 (60.0%), where it changed 33.1% of all sentences. Of the 72
sentences flagged after the commit, 37 are new or changed text, 27 stood
unchanged and were flagged before, and 8 stood unchanged and were not. The
review points at the sentences that people go on to change. The text they
write in their place is flagged again, under a rubric of lean wording that a
de-slop edit does not aim at. It also finds a small part of the edit: 57 of
the 642 sentences that the commits changed or removed (8.9%).

The tables in `docs/review.md` that compare the review with human edits and
with another model's annotation were measured with the list form. They stay,
marked as such, until they are measured again.

One reviewer was measured. The same comparison with a second model was
started and is reported on issue 525.
