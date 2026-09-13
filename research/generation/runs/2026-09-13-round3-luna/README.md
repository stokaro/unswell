# Stage C run 10, September 13, 2026: the OpenAI family on the same tasks

This directory is the record of the tenth run of stage C of the
[pattern protocol](../../../methods/llm-patterns-v1.md). It ran under
amendment 6, the third acquisition round, on the same 150 tasks as the
[Haiku round-three run](../2026-09-13-round3-haiku/README.md), so the two
families stay paired by task. The family is OpenAI through the Codex CLI of
the flat-rate subscription, with no API call, and the model is
`gpt-5.6-luna`. Every number is a rule outcome under one policy. It is not
a false-positive rate, not recall, not a share of LLM text, and not a
decision. It confirms no hypothesis.

## What ran

The task set is the Haiku round-three run's task set: the same 150 task IDs
and texts, drawn once by `corpus tasks` from the historical cohort of the
`development` partition with the comment role and the seed
`unswell-llm-patterns-v1`. The candidate artifacts were restricted to the
four development repositories that carry comment paragraphs and held no
controlled response. The draw is stratified by ecosystem:

| Ecosystem | Eligible | Selected |
| --- | --- | --- |
| cpp | 18 | 3 |
| csharp | 117 | 16 |
| java | 1,055 | 131 |

httpie/cli appears in neither row, because none of its 183 comment
paragraphs is an eligible task. The 150 tasks span assertj/assertj-core
131, AutoMapper/AutoMapper 16 and google/googletest 3.

Each task met the two operations and the two frozen prompt conditions: 600
requests. Every request ran as one non-interactive `codex exec` turn in a
read-only sandbox with an isolated home that carries no instructions and no
skills, six at a time. The model comes from the header the Codex CLI writes
on every run. The reasoning effort is `low`, the lowest the model accepts;
the CLI exposes no other decoding parameter, so the rest are `unavailable`.
Every request returned a final message on its first attempt. The shards of
this run carry the suffix `-luna-r3` through
`corpus generations --shard-suffix`.

The measurement of this run's shards needed the Markdown fix of #227, for
the same reason the Haiku run did.

## Coverage

| Operation | Prompt | Responses | Requested words | Realized words | Off length | Mean overlap | Overlap above 0.5 | Fact-sheet identifiers carried |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `generate` | `neutral` | 150 | 25.0 | 23.3 | 1 | 0.004 | 1 | 570 of 690 |
| `generate` | `plain` | 150 | 25.0 | 23.6 | 1 | 0.004 | 1 | 519 of 690 |
| `polish` | `neutral` | 150 | 25.0 | 25.0 | 2 | 0.523 | 85 | 186 of 690 |
| `polish` | `plain` | 150 | 25.0 | 24.7 | 5 | 0.437 | 74 | 178 of 690 |

All 600 responses are complete; none was refused, truncated, or failed.
Overlap is the share of response tokens inside word 8-gram matches with the
original. For `polish` a high overlap is the operation itself. Among the
300 `generate` responses two share an 8-gram with the original and 25 share
a 4-gram. This model again carries far more fact-sheet identifiers than
Haiku does, 570 of 690 against 313 under `generate`/`neutral`. Five
responses carry a heading or a fenced block and six run to more than one
paragraph. The record marks contamination by public training data as
unknown for every response.

The extraction found 1,605 units in the 600 controlled documents: 640
fragments, 561 sentences, and 404 paragraphs. The fragment share is again
higher than Haiku's, because this model answers in short labeled lines that
carry no sentence. It found 16 findings in 12,015 prose words (1.3 per
1,000 words) in 15 documents: seven passive-candidate density, three
parenthetical load, two reading grade, two long sentences, and two noun
stacks. The Haiku run found 56 findings in 14,248 words, 3.9 per 1,000.

## Paired tables

The [paired tables](paired.json) pair each original unit with its response
document. All 150 tasks pair in every arm, across 14 provenance components.
Rules whose class admits the comment role enter; 31 of 40 such rules fired
in no response and no original of any arm. Rows with any finding:

| Rule | Arm | Originals with finding | Responses with finding | Paired change (95% interval) |
| --- | --- | --- | --- | --- |
| `readability.grade-metric` | `generate`/`plain` | 0 | 1 | +0.7 points (+0.0 to +2.2) |
| `readability.grade-metric` | `polish`/`neutral` | 0 | 1 | +0.7 points (+0.0 to +2.2) |
| `repetition.exact-sentence` | `generate`/`neutral` | 6 | 0 | -4.0 points (-6.2 to -1.5) |
| `repetition.exact-sentence` | `generate`/`plain` | 6 | 0 | -4.0 points (-6.2 to -1.5) |
| `repetition.exact-sentence` | `polish`/`neutral` | 6 | 0 | -4.0 points (-6.2 to -1.5) |
| `repetition.exact-sentence` | `polish`/`plain` | 6 | 0 | -4.0 points (-6.2 to -1.5) |
| `repetition.near-sentence` | `generate`/`neutral` | 2 | 0 | -1.3 points (-3.2 to +0.0) |
| `repetition.near-sentence` | `generate`/`plain` | 2 | 0 | -1.3 points (-3.2 to +0.0) |
| `repetition.near-sentence` | `polish`/`neutral` | 2 | 0 | -1.3 points (-3.2 to +0.0) |
| `repetition.near-sentence` | `polish`/`plain` | 2 | 0 | -1.3 points (-3.2 to +0.0) |
| `repetition.ngram-density` | `generate`/`neutral` | 6 | 0 | -4.0 points (-9.6 to +0.0) |
| `repetition.ngram-density` | `generate`/`plain` | 6 | 0 | -4.0 points (-9.6 to +0.0) |
| `repetition.ngram-density` | `polish`/`neutral` | 6 | 0 | -4.0 points (-9.6 to +0.0) |
| `repetition.ngram-density` | `polish`/`plain` | 6 | 0 | -4.0 points (-9.6 to +0.0) |
| `repetition.paragraph-overlap` | `generate`/`neutral` | 5 | 0 | -3.3 points (-5.1 to -0.8) |
| `repetition.paragraph-overlap` | `generate`/`plain` | 5 | 0 | -3.3 points (-5.1 to -0.8) |
| `repetition.paragraph-overlap` | `polish`/`neutral` | 5 | 0 | -3.3 points (-5.1 to -0.8) |
| `repetition.paragraph-overlap` | `polish`/`plain` | 5 | 0 | -3.3 points (-5.1 to -0.8) |
| `syntax.long-sentence` | `generate`/`neutral` | 11 | 0 | -7.3 points (-19.1 to -0.7) |
| `syntax.long-sentence` | `generate`/`plain` | 11 | 0 | -7.3 points (-19.1 to -0.7) |
| `syntax.long-sentence` | `polish`/`neutral` | 11 | 2 | -6.0 points (-18.2 to +0.0) |
| `syntax.long-sentence` | `polish`/`plain` | 11 | 0 | -7.3 points (-19.1 to -0.7) |
| `syntax.noun-stack` | `generate`/`neutral` | 2 | 0 | -1.3 points (-5.0 to +0.0) |
| `syntax.noun-stack` | `generate`/`plain` | 2 | 2 | +0.0 points (-4.1 to +3.1) |
| `syntax.noun-stack` | `polish`/`neutral` | 2 | 0 | -1.3 points (-5.0 to +0.0) |
| `syntax.noun-stack` | `polish`/`plain` | 2 | 0 | -1.3 points (-5.0 to +0.0) |
| `syntax.parenthetical-load` | `generate`/`neutral` | 11 | 0 | -7.3 points (-14.6 to -1.0) |
| `syntax.parenthetical-load` | `generate`/`plain` | 11 | 0 | -7.3 points (-14.6 to -1.0) |
| `syntax.parenthetical-load` | `polish`/`neutral` | 11 | 3 | -5.3 points (-10.4 to -0.8) |
| `syntax.parenthetical-load` | `polish`/`plain` | 11 | 0 | -7.3 points (-14.6 to -1.0) |
| `syntax.passive-candidate-density` | `generate`/`neutral` | 1 | 1 | +0.0 points (+0.0 to +0.0) |
| `syntax.passive-candidate-density` | `generate`/`plain` | 1 | 2 | +0.7 points (+0.0 to +2.5) |
| `syntax.passive-candidate-density` | `polish`/`neutral` | 1 | 3 | +1.3 points (+0.0 to +3.2) |
| `syntax.passive-candidate-density` | `polish`/`plain` | 1 | 1 | +0.0 points (+0.0 to +0.0) |

Fifteen intervals exclude zero and every one of them is negative. No rule
fires in more responses than originals in any arm by a margin the interval
separates from zero. `readability.grade-metric`, the rule whose paired
interval clears the minimum useful difference in the Haiku run, fires in
two responses of this run across all four arms.

The tables also hold a `difference_from_h0` field. It compares a
one-paragraph response with whole H0 documents of the comment role, so
document length dominates it. It stays because the protocol names it. The
paragraph-level comparison below is the fair one.

## Paragraph-level comparison

The [run tables](tables-paragraph-luna-r3.json) count the controlled cohort
from this run's shards only: 389 documents with a paragraph unit, 404
paragraphs, 3 provenance components, 7,631 prose words. Controlled
paragraphs of all arms against all historical paragraphs:

| Rule | Historical | Controlled |
| --- | --- | --- |
| `syntax.long-sentence` | 3,019 of 81,690 (3.7%) | 1 of 404 (0.2%) |
| `syntax.passive-candidate-density` | 1,954 of 81,690 (2.4%) | 5 of 404 (1.2%) |
| `readability.grade-metric` | 1,604 of 81,690 (2.0%) | 0 of 404 (0.0%) |
| `repetition.exact-sentence` | 1,217 of 81,690 (1.5%) | 0 of 404 (0.0%) |
| `syntax.parenthetical-load` | 974 of 81,690 (1.2%) | 3 of 404 (0.7%) |
| `repetition.paragraph-overlap` | 710 of 81,690 (0.9%) | 0 of 404 (0.0%) |
| `repetition.ngram-density` | 407 of 81,690 (0.5%) | 0 of 404 (0.0%) |
| `repetition.near-sentence` | 320 of 81,690 (0.4%) | 0 of 404 (0.0%) |
| `repetition.sentence-openers` | 309 of 81,690 (0.4%) | 0 of 404 (0.0%) |
| `readability.long-paragraph` | 279 of 81,690 (0.3%) | 0 of 404 (0.0%) |
| `filler.wordy-phrase` | 248 of 81,690 (0.3%) | 0 of 404 (0.0%) |
| `repetition.paragraph-openers` | 163 of 81,690 (0.2%) | 0 of 404 (0.0%) |
| `syntax.noun-stack` | 97 of 81,690 (0.1%) | 2 of 404 (0.5%) |
| `filler.weak-intensifiers` | 32 of 81,690 (0.0%) | 0 of 404 (0.0%) |
| `repetition.syntax-template` | 19 of 81,690 (0.0%) | 0 of 404 (0.0%) |
| `format.em-dash-density` | 4 of 81,690 (0.0%) | 0 of 404 (0.0%) |
| `syntax.nominalization-chain` | 3 of 81,690 (0.0%) | 0 of 404 (0.0%) |
| `filler.announced-importance` | 1 of 81,690 (0.0%) | 0 of 404 (0.0%) |

No difference reaches three points, and no positive difference exceeds half
a point. The largest negative one is `syntax.long-sentence`, 3.5 points
below the historical rate. With 404 paragraphs a rule at 1% prevalence
fires about four times. These are 3 components; the
[screening record](../../../methods/screening/2026-09-13-development-3.json)
is what counts across the partition. The whole controlled cohort, ten runs
together, has 5,303 paragraphs in 5,038 documents across 152 components;
the [six-cohort tables](tables-paragraph.json) count it that way.

## Limits

- One family, one model, at the lowest reasoning effort the model accepts.
  A stronger effort is a different setting and would need its own run.
- Three components carry this run's paragraphs, and 131 of the 150 tasks
  come from one repository, which the ecosystem strata of the frame decide.
- The Codex CLI reports the model in a header and offers no decoding
  parameter beyond the reasoning effort, so the run cannot record a
  temperature or a seed.
- The isolated home carries no instructions and no skills, but the harness
  cannot show that the model saw nothing else.
- The overlap flag is the only contamination control. Two `generate`
  responses raise it and two share an 8-gram with their original.

## What is in this directory

- `tasks.json`, `requests.json`, `responses.json`, `records.json`: the
  saved records, with every input and output text and its hash.
- `shards/`: the controlled cohort's 3 shard manifests of this run.
- `dataset-plan.json`, `tables.json`, `tables-paragraph.json`,
  `tables-sentence.json`, `paired.json`: the six-cohort plan and tables
  after this run joined the controlled cohort.
- `tables-paragraph-luna-r3.json`: the paragraph tables with the controlled
  cohort restricted to this run's shards.
- `digests.json`: SHA-256 of every record, shard, pinned copy, finding
  artifact, and output of the measurement, which are reproducible from the
  records and the checkouts.
