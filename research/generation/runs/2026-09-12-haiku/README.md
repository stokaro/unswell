# Stage C run 3, September 12, 2026: a second model

This directory is the record of the third run of stage C of the
[pattern protocol](../../../methods/llm-patterns-v1.md), after the
[pilot](../2026-09-11-pilot/README.md) and [run 2](../2026-09-11-run2/README.md).
It ran under amendment 1: one generator family, Anthropic Claude, through
session agents of the flat-rate subscription, with no paid call. The model
is `claude-haiku-4-5-20251001`; the two earlier runs used `claude-opus-5`.
Every number is a rule outcome under one policy. It is not a false-positive
rate, not recall, not a share of LLM text, and not a decision. The run adds
a second model of the same family to the controlled cohort. It confirms no
hypothesis.

## What ran

The task set is the union of the pilot's 60 tasks and run 2's 140 tasks:
the same 200 task IDs and texts, so every task now has responses from both
models. No new draw was made. The header of `tasks.json` repeats the
pilot's eligible frame, and its selected counts are the sums of the two
draws:

| Ecosystem | Eligible in the pilot's frame | Selected |
| --- | --- | --- |
| c | 43 | 3 |
| cpp | 11 | 2 |
| csharp | 392 | 9 |
| go | 774 | 87 |
| java | 12 | 1 |
| javascript | 906 | 41 |
| python | 100 | 8 |
| rust | 364 | 49 |

The 200 tasks span 25 repositories: 158 in the training partition and 42
in development.

Each task met the two operations, `generate` from its fact sheet and
`polish` of its original text, under the two frozen prompt conditions,
`neutral` and `plain`: 800 requests. Every request went to one
instruction-free agent with the request text and one line telling the agent
to use no tools. The requests went through four workflow runs of the same
agent type, 200 each. The 800 agent transcripts name the model
`claude-haiku-4-5-20251001`; each agent made exactly one call, the answer,
and used no tool. Decoding parameters, remote request identifiers, and cost
are `unavailable`, as the amendment states. The shards of this run carry the
suffix `-haiku` through `corpus generations --shard-suffix`.

A request ID is a hash of the task, the operation, and the prompt, so two
runs that answer the same task write the same response path,
`generated/<request>.md`. The paired analysis therefore matches a record to
its response by the run-scoped source ID instead of the path, and a table of
this run never counts the Opus response to the same task. The tables of the
two earlier runs were built before this run's shards existed and are
unchanged. The dataset and analysis limits rose from 512 to 1,024 shards and
finding artifacts, because the corpus now holds 519 shards.

## Coverage

| Operation | Prompt | Responses | Requested words | Realized words | Off length | Mean overlap | Overlap above 0.5 | Fact-sheet identifiers carried |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `generate` | `neutral` | 200 | 22.3 | 22.2 | 0 | 0.000 | 0 | 428 of 807 |
| `generate` | `plain` | 200 | 22.3 | 22.1 | 0 | 0.000 | 0 | 435 of 807 |
| `polish` | `neutral` | 200 | 22.3 | 22.7 | 5 | 0.564 | 121 | 237 of 807 |
| `polish` | `plain` | 200 | 22.3 | 22.4 | 4 | 0.494 | 110 | 222 of 807 |

All 800 responses are complete; none was refused, truncated, or failed.
Overlap is the share of response tokens inside word 8-gram matches with
the original. For `polish` a high overlap is the operation itself. For
`generate` no response shares an 8-gram with the original: this model
paraphrased the testify assertion comments that the Opus runs reproduced
from the signature, and 29 `generate` responses share a 4-gram with the
original. The record marks contamination by public training data as
unknown for every response.

Every `generate` response came back as one prose paragraph without a
heading or a fenced code block. The extraction found 1,999 units in the 800
controlled documents: 728 paragraphs, 1,078 sentences, and 193 fragments.
It found 53 findings in 17,423 prose words (3.0 per 1,000 words) in 43
documents. Thirteen of the findings are long sentences, twelve are reading
grade, twelve are passive-candidate density, and eight are noun stacks.

## Paired tables

The [paired tables](paired.json) pair each original unit with its response
document. All 200 tasks pair in every arm, across 42 provenance
components. Rules whose class admits the comment role enter; 32 of
40 such rules fired in no response of any arm. Rows with any
finding, original or response:

| Rule | Arm | Originals with finding | Responses with finding | Paired change (95% interval) |
| --- | --- | --- | --- | --- |
| `filler.weak-intensifiers` | `generate`/`neutral` | 1 | 0 | -0.5 points (-1.8 to +0.0) |
| `filler.weak-intensifiers` | `generate`/`plain` | 1 | 0 | -0.5 points (-1.8 to +0.0) |
| `filler.weak-intensifiers` | `polish`/`neutral` | 1 | 0 | -0.5 points (-1.8 to +0.0) |
| `filler.weak-intensifiers` | `polish`/`plain` | 1 | 1 | +0.0 points (+0.0 to +0.0) |
| `filler.wordy-phrase` | `generate`/`neutral` | 2 | 0 | -1.0 points (-2.5 to +0.0) |
| `filler.wordy-phrase` | `generate`/`plain` | 2 | 0 | -1.0 points (-2.5 to +0.0) |
| `filler.wordy-phrase` | `polish`/`neutral` | 2 | 1 | -0.5 points (-1.6 to +0.0) |
| `filler.wordy-phrase` | `polish`/`plain` | 2 | 0 | -1.0 points (-2.5 to +0.0) |
| `readability.grade-metric` | `generate`/`neutral` | 2 | 4 | +1.0 points (-1.0 to +3.3) |
| `readability.grade-metric` | `generate`/`plain` | 2 | 4 | +1.0 points (+0.0 to +2.6) |
| `readability.grade-metric` | `polish`/`neutral` | 2 | 2 | +0.0 points (-1.8 to +1.3) |
| `readability.grade-metric` | `polish`/`plain` | 2 | 2 | +0.0 points (-1.6 to +1.5) |
| `repetition.exact-sentence` | `generate`/`neutral` | 11 | 0 | -5.5 points (-9.0 to -1.8) |
| `repetition.exact-sentence` | `generate`/`plain` | 11 | 0 | -5.5 points (-9.0 to -1.8) |
| `repetition.exact-sentence` | `polish`/`neutral` | 11 | 0 | -5.5 points (-9.0 to -1.8) |
| `repetition.exact-sentence` | `polish`/`plain` | 11 | 0 | -5.5 points (-9.0 to -1.8) |
| `repetition.near-sentence` | `generate`/`neutral` | 2 | 0 | -1.0 points (-2.4 to +0.0) |
| `repetition.near-sentence` | `generate`/`plain` | 2 | 0 | -1.0 points (-2.4 to +0.0) |
| `repetition.near-sentence` | `polish`/`neutral` | 2 | 0 | -1.0 points (-2.4 to +0.0) |
| `repetition.near-sentence` | `polish`/`plain` | 2 | 0 | -1.0 points (-2.4 to +0.0) |
| `repetition.ngram-density` | `generate`/`neutral` | 2 | 0 | -1.0 points (-3.8 to +0.0) |
| `repetition.ngram-density` | `generate`/`plain` | 2 | 0 | -1.0 points (-3.8 to +0.0) |
| `repetition.ngram-density` | `polish`/`neutral` | 2 | 1 | -0.5 points (-1.9 to +0.0) |
| `repetition.ngram-density` | `polish`/`plain` | 2 | 0 | -1.0 points (-3.8 to +0.0) |
| `repetition.paragraph-overlap` | `generate`/`neutral` | 4 | 0 | -2.0 points (-4.2 to +0.0) |
| `repetition.paragraph-overlap` | `generate`/`plain` | 4 | 0 | -2.0 points (-4.2 to +0.0) |
| `repetition.paragraph-overlap` | `polish`/`neutral` | 4 | 0 | -2.0 points (-4.2 to +0.0) |
| `repetition.paragraph-overlap` | `polish`/`plain` | 4 | 0 | -2.0 points (-4.2 to +0.0) |
| `syntax.long-sentence` | `generate`/`neutral` | 9 | 1 | -4.0 points (-9.0 to -0.5) |
| `syntax.long-sentence` | `generate`/`plain` | 9 | 0 | -4.5 points (-9.6 to -1.2) |
| `syntax.long-sentence` | `polish`/`neutral` | 9 | 8 | -0.5 points (-2.8 to +1.9) |
| `syntax.long-sentence` | `polish`/`plain` | 9 | 4 | -2.5 points (-5.6 to -0.4) |
| `syntax.noun-stack` | `generate`/`neutral` | 0 | 6 | +3.0 points (+0.8 to +5.1) |
| `syntax.noun-stack` | `generate`/`plain` | 0 | 1 | +0.5 points (+0.0 to +1.8) |
| `syntax.noun-stack` | `polish`/`plain` | 0 | 1 | +0.5 points (+0.0 to +1.6) |
| `syntax.parenthetical-load` | `generate`/`neutral` | 6 | 0 | -3.0 points (-7.7 to -0.4) |
| `syntax.parenthetical-load` | `generate`/`plain` | 6 | 1 | -2.5 points (-7.3 to +0.5) |
| `syntax.parenthetical-load` | `polish`/`neutral` | 6 | 3 | -1.5 points (-3.7 to +0.0) |
| `syntax.parenthetical-load` | `polish`/`plain` | 6 | 1 | -2.5 points (-6.1 to -0.4) |
| `syntax.passive-candidate-density` | `generate`/`neutral` | 5 | 2 | -1.5 points (-3.6 to +0.6) |
| `syntax.passive-candidate-density` | `generate`/`plain` | 5 | 1 | -2.0 points (-4.5 to +0.0) |
| `syntax.passive-candidate-density` | `polish`/`neutral` | 5 | 4 | -0.5 points (-1.7 to +0.0) |
| `syntax.passive-candidate-density` | `polish`/`plain` | 5 | 5 | +0.0 points (-2.1 to +2.1) |

These are small counts over 200 pairs. Eight intervals exclude zero, and
all of them are negative: `repetition.exact-sentence` in every arm,
`syntax.long-sentence` under `generate`, and `syntax.parenthetical-load`
under `generate`. An original paragraph carries a repetition finding when
another part of its document repeats it; a one-paragraph response has no
other part, so those rows measure the context of the original, not the
model. The long-sentence and parenthetical rows under `generate` are the
operation writing a shorter or a plainer sentence, as in run 2. A `polish`
response keeps most of the same findings.

Three rules fire in a response and not in its original.
`repetition.sentence-openers` fires once under each `polish` prompt.
`syntax.noun-stack` fires in eight responses against no original: six
under `generate`/`neutral`, one under `generate`/`plain`, and one under
`polish`/`plain`. The six `generate`/`neutral` cases string fact-sheet
identifiers into one noun phrase, such as `dirty date parameter value`,
`clock sequence field value`, `assertion failure output messages`, and
`case-first naming convention transformation rules`. Run 2 had one such
response. The tables also hold a `difference_from_h0` field. It compares a
one-paragraph response with whole H0 documents of the comment role, so
document length dominates it. It stays because the protocol names it. The
paragraph-level comparison below is the fair one.

## Paragraph-level comparison

The [run tables](tables-paragraph-haiku.json) count the controlled cohort
from this run's shards only: 722 documents with a paragraph
unit, 728 paragraphs, 25 provenance components,
15,755 prose words. Controlled paragraphs of all arms
against all historical paragraphs:

| Rule | Historical | Controlled |
| --- | --- | --- |
| `syntax.long-sentence` | 2,251 of 54,223 (4.2%) | 11 of 728 (1.5%) |
| `readability.grade-metric` | 1,844 of 54,223 (3.4%) | 12 of 728 (1.6%) |
| `syntax.passive-candidate-density` | 1,406 of 54,223 (2.6%) | 12 of 728 (1.6%) |
| `syntax.parenthetical-load` | 977 of 54,223 (1.8%) | 5 of 728 (0.7%) |
| `repetition.exact-sentence` | 596 of 54,223 (1.1%) | 0 of 728 (0.0%) |
| `repetition.paragraph-overlap` | 344 of 54,223 (0.6%) | 0 of 728 (0.0%) |
| `repetition.sentence-openers` | 270 of 54,223 (0.5%) | 0 of 728 (0.0%) |
| `filler.wordy-phrase` | 172 of 54,223 (0.3%) | 1 of 728 (0.1%) |
| `repetition.near-sentence` | 168 of 54,223 (0.3%) | 0 of 728 (0.0%) |
| `repetition.paragraph-openers` | 135 of 54,223 (0.2%) | 0 of 728 (0.0%) |
| `repetition.ngram-density` | 122 of 54,223 (0.2%) | 1 of 728 (0.1%) |
| `readability.long-paragraph` | 75 of 54,223 (0.1%) | 0 of 728 (0.0%) |
| `syntax.noun-stack` | 59 of 54,223 (0.1%) | 8 of 728 (1.1%) |
| `filler.weak-intensifiers` | 20 of 54,223 (0.0%) | 1 of 728 (0.1%) |
| `repetition.syntax-template` | 15 of 54,223 (0.0%) | 0 of 728 (0.0%) |
| `syntax.nominalization-chain` | 3 of 54,223 (0.0%) | 0 of 728 (0.0%) |
| `filler.announced-importance` | 1 of 54,223 (0.0%) | 0 of 728 (0.0%) |

With 728 controlled paragraphs, a rule at 1% prevalence fires
about seven times. No difference reaches the protocol's minimum useful
difference of three points. The largest positive difference is
`syntax.noun-stack`, one point above the historical rate; the largest
negative ones are `syntax.long-sentence`, `readability.grade-metric`, and
`syntax.parenthetical-load`, each one to three points below it. The whole
controlled cohort, three runs together, has 1,257
paragraphs in 1,233 documents across
57 components; the
[six-cohort tables](tables-paragraph.json) count it that way.

## Limits

- One family, two models. Every other family is `untested`; protocol
  version 2 needs three families, and a second model of one family is a
  model stratum, not a family.
- The agent harness ran the search agent type with an instruction-free
  system prompt; the agents used no tools, but the harness cannot show that
  the model saw nothing else.
- The tasks are the Opus runs' tasks, so a comparison between the models
  is paired by task. The runs are a day apart, and the historical frame was
  measured again in between; the task texts are unchanged.
- Requested lengths are short (about 22 words), because the originals are
  comment paragraphs; a one-sentence response has no paragraph unit.
- The overlap flag is the only contamination control, and no `generate`
  response raised it.

## What is in this directory

- `tasks.json`, `requests.json`, `responses.json`, `records.json`: the
  saved records, with every input and output text and its hash.
- `shards/`: the controlled cohort's 25 shard manifests of this run.
- `dataset-plan.json`, `tables.json`, `tables-paragraph.json`,
  `tables-sentence.json`, `paired.json`: the six-cohort plan and tables
  after this run joined the controlled cohort.
- `tables-paragraph-haiku.json`: the paragraph tables with the controlled
  cohort restricted to this run's shards.
- `digests.json`: SHA-256 of every record, shard, pinned copy, finding
  artifact, and output of the measurement, which are reproducible from the
  records and the checkouts.
