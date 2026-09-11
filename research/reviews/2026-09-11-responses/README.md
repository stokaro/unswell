# Review on real text: the controlled responses, September 11, 2026

This directory records the second review of the kind that
[ADR 0037](../../../docs/adr/0037-diagnostics-not-authorship.md) asks
for. Every catalog rule ran on the 800 responses of the controlled
cohort. They are the 240 of the [pilot](../../generation/runs/2026-09-11-pilot/README.md)
and the 560 of the [second run](../../generation/runs/2026-09-11-run2/README.md).
Each arm, generate and polish, holds 400, and the responses hold 15,920
prose words. A reader judged every finding under the rule's own
statement. The
responses are known generations, so nothing here decides authorship. The
question is whether the diagnostics fire on constructions in them that a
reader could act on.

One reader judged the findings: the maintainer's agent session, on the
day of the run. The maintainer has not confirmed the judgments.

## What ran

The tool was the merged commit `a1c6d7a` of this repository. Each
response is one Markdown file under
`<owner>__<repo>/generated/<response_id>.md` in the controlled cohort's
work directory. The files were copied into a scratch root with
[config-all.yaml](config-all.yaml), the all-rules configuration of the
[Ptah review](../2026-09-11-ptah/README.md) without its file exclusion,
and scanned with `--no-gate --include-source --timeout 20m --jobs 4`. The
scanner read the responses as Markdown paragraphs, not as comments inside
source files. A JSDoc block that a response reproduces is prose to it.
The run completed: 800 documents, 39 findings.
[summary.json](summary.json) holds the counts by rule and by arm.

## Findings by rule

| Rule | Findings | Generate | Polish |
| --- | --- | --- | --- |
| `syntax.long-sentence` | 13 | 0 | 13 |
| `syntax.passive-candidate-density` | 7 | 3 | 4 |
| `syntax.parenthetical-load` | 6 | 0 | 6 |
| `filler.wordy-phrase` | 3 | 0 | 3 |
| `readability.grade-metric` | 3 | 0 | 3 |
| `repetition.ngram-density` | 3 | 0 | 3 |
| `repetition.sentence-openers` | 2 | 0 | 2 |
| `filler.weak-intensifiers` | 1 | 0 | 1 |
| `syntax.noun-stack` | 1 | 1 | 0 |

The other 31 rules fired nowhere. The polish arm holds 35 of the 39
findings. Eighteen of them sit in one repository, date-fns, whose source
comments are JSDoc blocks that the polish responses keep.

## The judged findings

[sample.json](sample.json) holds every finding with the judgment and its
reason. The table condenses it.

| Rule | Justified | Not | Why not |
| --- | --- | --- | --- |
| `syntax.long-sentence` | 4 | 9 | seven JSDoc example blocks and tag lists read as one sentence; two sentence pairs joined at `chi.Router.` |
| `syntax.passive-candidate-density` | 7 | 0 | |
| `syntax.parenthetical-load` | 0 | 6 | the parentheses of a call in a JSDoc example |
| `filler.wordy-phrase` | 3 | 0 | |
| `readability.grade-metric` | 3 | 0 | |
| `repetition.ngram-density` | 0 | 3 | example output inside a JSDoc block |
| `repetition.sentence-openers` | 0 | 2 | a JSDoc tag list; the repeated opening is the tag format |
| `filler.weak-intensifiers` | 1 | 0 | |
| `syntax.noun-stack` | 1 | 0 | |

Nineteen findings name a construction in prose a reader could act on:
be-plus-participle sentences, sentences of 35 to 67 words, `in order to`,
`very possible`, and a four-noun stack. Twenty do not. Eighteen of those
fire on JSDoc example code and tag lists that the polish responses
reproduce from their sources. Two fire on a pair of sentences that the
splitter joins because the period of `chi.Router.` reads as part of the
identifier.

## What it shows

The rules fire on almost nothing in comments of this length. The
generate arm writes from the code alone. It holds four findings in 400
responses: three passive paragraphs and one noun stack. The polish arm
holds the rest, and most of those come from the source text. The record
of the second run showed that no paragraph rate differs between the arms
and the sources by three points. This review reads the findings behind
that result.

Two limits of the tool appear. Example code and tag lists inside a
comment reach the sentence units, as the directive comments did in the
[frequency run](../../acquisition/runs/2026-09-11-frequencies/README.md).
And the sentence splitter takes the period after an identifier such as
`chi.Router.` as part of the name. Both are questions for the extractor,
recorded here and not acted on.

## What it is not

The judgments are one reader's, and the reader is an agent session. The
counts are not a precision figure, and the review makes no claim that
the rules see generation: they fire on four of 400 generated comments,
and the polish findings are mostly inherited. The two limits are
observations, not defects filed.
