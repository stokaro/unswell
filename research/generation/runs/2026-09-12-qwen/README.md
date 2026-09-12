# Stage C run 5, September 12, 2026: a third family

This directory is the record of the fifth run of stage C of the
[pattern protocol](../../../methods/llm-patterns-v1.md), after the
[pilot](../2026-09-11-pilot/README.md), [run 2](../2026-09-11-run2/README.md),
the [Haiku run](../2026-09-12-haiku/README.md), and the
[luna run](../2026-09-12-codex-luna/README.md). It ran under amendment 3: a
third generator family, Qwen, through an OpenAI-compatible chat endpoint
that the maintainer's organization operates, with no vendor call. The model
is `HivenetQuant/Qwen3.6-35B-A3B`, the smaller of the two the endpoint
serves, with thinking disabled. Every number is a rule outcome under one
policy. It is not a false-positive rate, not recall, not a share of LLM
text, and not a decision. The run brings the controlled cohort to three
families. It confirms no hypothesis.

## What ran

The task set is the one the Haiku run used: the union of the pilot's 60
tasks and run 2's 140 tasks. The 200 task IDs and texts are the same, so
every task now has responses from all three families. The header of `tasks.json` repeats
the pilot's eligible frame with the selected counts of the two draws summed.
The 200 tasks span 25 repositories: 158 in the training partition and 42 in
development.

Each task met the two operations, `generate` from its fact sheet and
`polish` of its original text, under the two frozen prompt conditions,
`neutral` and `plain`: 800 requests.

Every request went as the one user message of a chat completion request:
the request text and the one harness line, with no system message. The
decoding parameters were temperature 0.7, top-p 0.8, seed 1, and at most
512 completion tokens, with thinking disabled through the chat template.
The endpoint named `HivenetQuant/Qwen3.6-35B-A3B` in every response. Every
response stopped on its own, and no response carried reasoning text. Six
requests ran at a time, and a request took half a second on average with
29 completion tokens. Each response row keeps the identifier the endpoint
returned, its duration, and its token count; the records name them, and a
cost is `unavailable`, because the endpoint bills no vendor call. The
bearer key stayed in the environment. The shards of this run carry the
suffix `-qwen` through `corpus generations --shard-suffix`.

The engine fixes merged earlier in the day changed the rule versions and
one rule parameter. Every earlier finding artifact therefore carried a
policy identity that this run's shards would not share. The whole corpus
was measured again under the current engine in one run before any table
below was built. The earlier records keep the tables of their own
measurements.

## Coverage

| Operation | Prompt | Responses | Requested words | Realized words | Off length | Mean overlap | Overlap above 0.5 | Fact-sheet identifiers carried |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `generate` | `neutral` | 200 | 22.3 | 21.9 | 28 | 0.005 | 1 | 534 of 807 |
| `generate` | `plain` | 200 | 22.3 | 21.0 | 36 | 0.004 | 1 | 535 of 807 |
| `polish` | `neutral` | 200 | 22.3 | 20.8 | 20 | 0.444 | 95 | 234 of 807 |
| `polish` | `plain` | 200 | 22.3 | 19.9 | 33 | 0.285 | 62 | 229 of 807 |

All 800 responses are complete; none was refused, truncated, or failed.
Overlap is the share of response tokens inside word 8-gram matches with
the original. For `polish` a high overlap is the operation itself. For
`generate` two responses raised the flag, and 3 share an 8-gram with the
original by a whitespace tokenization; 19 share a 4-gram. The model carries
534 of 807 fact-sheet identifiers under `generate`/`neutral`, between the
Haiku run's 428 and the luna run's 672. It formats little: 0 `generate`
responses carry a heading or a fenced code block, 36 hold a code span, and
0 run to more than one paragraph. It runs short more often than the other
generators: 117 responses fall outside the 30% length tolerance, 100 of
them below it. The record marks contamination by public training data as
unknown for every response.

The extraction found 2,167 units in the 800 controlled documents: 700
paragraphs, 1,116 sentences, and 351 fragments. It found 34 findings in 16,268
prose words (2.1 per 1,000 words) in 24 documents; the most frequent are
8 `syntax.parenthetical-load`, 7 `readability.grade-metric`, 7 `syntax.long-sentence`, 6 `syntax.passive-candidate-density`.

## Paired tables

The [paired tables](paired.json) pair each original unit with its response
document. 178 of the 200 tasks pair in every arm, across 35 provenance
components. The other 22 originals have no unit to pair under the current
engine. All but one are documentation tag sections in JavaScript that it
now excludes. The last is a Go comment whose unit ID moved when an earlier
unit of its shard was excluded. Rules whose class admits the
comment role enter; 36 of 40 such rules fired in no response of any arm.
Rows with any finding, original or response:

| Rule | Arm | Originals with finding | Responses with finding | Paired change (95% interval) |
| --- | --- | --- | --- | --- |
| `filler.weak-intensifiers` | `generate`/`neutral` | 1 | 0 | -0.6 points (-2.1 to +0.0) |
| `filler.weak-intensifiers` | `generate`/`plain` | 1 | 0 | -0.6 points (-2.1 to +0.0) |
| `filler.weak-intensifiers` | `polish`/`neutral` | 1 | 0 | -0.6 points (-2.1 to +0.0) |
| `filler.weak-intensifiers` | `polish`/`plain` | 1 | 0 | -0.6 points (-2.1 to +0.0) |
| `filler.wordy-phrase` | `generate`/`neutral` | 1 | 0 | -0.6 points (-1.8 to +0.0) |
| `filler.wordy-phrase` | `generate`/`plain` | 1 | 0 | -0.6 points (-1.8 to +0.0) |
| `filler.wordy-phrase` | `polish`/`neutral` | 1 | 0 | -0.6 points (-1.8 to +0.0) |
| `filler.wordy-phrase` | `polish`/`plain` | 1 | 0 | -0.6 points (-1.8 to +0.0) |
| `readability.grade-metric` | `generate`/`neutral` | 1 | 2 | +0.6 points (+0.0 to +1.8) |
| `readability.grade-metric` | `generate`/`plain` | 1 | 0 | -0.6 points (-2.1 to +0.0) |
| `readability.grade-metric` | `polish`/`neutral` | 1 | 2 | +0.6 points (+0.0 to +1.8) |
| `readability.grade-metric` | `polish`/`plain` | 1 | 2 | +0.6 points (+0.0 to +1.8) |
| `repetition.exact-sentence` | `generate`/`neutral` | 7 | 0 | -3.9 points (-6.6 to -0.7) |
| `repetition.exact-sentence` | `generate`/`plain` | 7 | 0 | -3.9 points (-6.6 to -0.7) |
| `repetition.exact-sentence` | `polish`/`neutral` | 7 | 0 | -3.9 points (-6.6 to -0.7) |
| `repetition.exact-sentence` | `polish`/`plain` | 7 | 0 | -3.9 points (-6.6 to -0.7) |
| `repetition.near-sentence` | `generate`/`neutral` | 2 | 0 | -1.1 points (-2.7 to +0.0) |
| `repetition.near-sentence` | `generate`/`plain` | 2 | 0 | -1.1 points (-2.7 to +0.0) |
| `repetition.near-sentence` | `polish`/`neutral` | 2 | 0 | -1.1 points (-2.7 to +0.0) |
| `repetition.near-sentence` | `polish`/`plain` | 2 | 0 | -1.1 points (-2.7 to +0.0) |
| `repetition.paragraph-overlap` | `generate`/`neutral` | 3 | 0 | -1.7 points (-4.1 to +0.0) |
| `repetition.paragraph-overlap` | `generate`/`plain` | 3 | 0 | -1.7 points (-4.1 to +0.0) |
| `repetition.paragraph-overlap` | `polish`/`neutral` | 3 | 0 | -1.7 points (-4.1 to +0.0) |
| `repetition.paragraph-overlap` | `polish`/`plain` | 3 | 0 | -1.7 points (-4.1 to +0.0) |
| `syntax.long-sentence` | `generate`/`neutral` | 3 | 0 | -1.7 points (-3.8 to +0.0) |
| `syntax.long-sentence` | `generate`/`plain` | 3 | 0 | -1.7 points (-3.8 to +0.0) |
| `syntax.long-sentence` | `polish`/`neutral` | 3 | 0 | -1.7 points (-3.8 to +0.0) |
| `syntax.long-sentence` | `polish`/`plain` | 3 | 0 | -1.7 points (-3.8 to +0.0) |
| `syntax.noun-stack` | `generate`/`neutral` | 0 | 1 | +0.6 points (+0.0 to +2.1) |
| `syntax.noun-stack` | `generate`/`plain` | 0 | 1 | +0.6 points (+0.0 to +2.1) |
| `syntax.parenthetical-load` | `polish`/`neutral` | 0 | 1 | +0.6 points (+0.0 to +2.2) |
| `syntax.parenthetical-load` | `polish`/`plain` | 0 | 1 | +0.6 points (+0.0 to +2.2) |
| `syntax.passive-candidate-density` | `generate`/`neutral` | 4 | 0 | -2.2 points (-4.4 to -0.6) |
| `syntax.passive-candidate-density` | `generate`/`plain` | 4 | 1 | -1.7 points (-4.0 to +0.7) |
| `syntax.passive-candidate-density` | `polish`/`neutral` | 4 | 3 | -0.6 points (-3.0 to +1.9) |
| `syntax.passive-candidate-density` | `polish`/`plain` | 4 | 1 | -1.7 points (-3.4 to +0.0) |

These are counts over 178 pairs. The intervals that exclude zero:

- `repetition.exact-sentence` in every arm, at -3.9 points
- `syntax.passive-candidate-density` under `generate`/`neutral`, at -2.2 points

An original paragraph carries a repetition finding when another part of
its document repeats it. A one-paragraph response has no other part, so
those rows measure the context of the original, not the model. Rows where
more responses than originals carry a finding:

- `readability.grade-metric` under `generate`/`neutral`, 2 against 1
- `readability.grade-metric` under `polish`/`neutral`, 2 against 1
- `readability.grade-metric` under `polish`/`plain`, 2 against 1
- `syntax.noun-stack` under `generate`/`neutral`, 1 against 0
- `syntax.noun-stack` under `generate`/`plain`, 1 against 0
- `syntax.parenthetical-load` under `polish`/`neutral`, 1 against 0
- `syntax.parenthetical-load` under `polish`/`plain`, 1 against 0

The tables also hold a `difference_from_h0` field. It compares a
one-paragraph response with whole H0 documents of the comment role, so
document length dominates it. It stays because the protocol names it. The
paragraph-level comparison below is the fair one.

## Paragraph-level comparison

The [run tables](tables-paragraph-qwen.json) count the controlled cohort
from this run's shards only: 700 documents with a paragraph unit, 700
paragraphs, 25 provenance components, 13,404 prose words. Controlled
paragraphs of all arms against all historical paragraphs, both measured
under the current engine:

| Rule | Historical | Controlled |
| --- | --- | --- |
| `syntax.long-sentence` | 2,176 of 49,729 (4.4%) | 7 of 700 (1.0%) |
| `syntax.passive-candidate-density` | 1,400 of 49,729 (2.8%) | 5 of 700 (0.7%) |
| `readability.grade-metric` | 1,056 of 49,729 (2.1%) | 1 of 700 (0.1%) |
| `syntax.parenthetical-load` | 536 of 49,729 (1.1%) | 8 of 700 (1.1%) |
| `repetition.exact-sentence` | 498 of 49,729 (1.0%) | 0 of 700 (0.0%) |
| `repetition.paragraph-overlap` | 277 of 49,729 (0.6%) | 0 of 700 (0.0%) |
| `repetition.sentence-openers` | 273 of 49,729 (0.5%) | 0 of 700 (0.0%) |
| `filler.wordy-phrase` | 172 of 49,729 (0.3%) | 0 of 700 (0.0%) |
| `repetition.near-sentence` | 147 of 49,729 (0.3%) | 0 of 700 (0.0%) |
| `repetition.paragraph-openers` | 135 of 49,729 (0.3%) | 0 of 700 (0.0%) |
| `repetition.ngram-density` | 135 of 49,729 (0.3%) | 4 of 700 (0.6%) |
| `readability.long-paragraph` | 74 of 49,729 (0.1%) | 0 of 700 (0.0%) |
| `syntax.noun-stack` | 58 of 49,729 (0.1%) | 2 of 700 (0.3%) |
| `filler.weak-intensifiers` | 20 of 49,729 (0.0%) | 0 of 700 (0.0%) |
| `repetition.syntax-template` | 14 of 49,729 (0.0%) | 0 of 700 (0.0%) |
| `syntax.nominalization-chain` | 3 of 49,729 (0.0%) | 0 of 700 (0.0%) |
| `filler.announced-importance` | 1 of 49,729 (0.0%) | 0 of 700 (0.0%) |

With 700 controlled paragraphs, a rule at 1% fires about 7 times. `syntax.long-sentence` sits 3.4 points below the historical rate, more than the protocol's minimum useful difference of three points. The generators write shorter sentences than the comment paragraphs they answer, and the sentence boundary repair of the current engine lowers every cohort's long-sentence rate. The largest positive difference is `repetition.ngram-density` at +0.3 points.
The whole controlled cohort, six runs together, has 3,321 paragraphs in 3,290
documents across 132 components; the
[six-cohort tables](tables-paragraph.json) count it that way.

## Limits

- Three families have run, one model each; protocol version 2 keeps its
  other requirements.
- The endpoint is the organization's own; its serving stack, quantization,
  and system prompt handling are not part of this record beyond the model
  name it returns and the parameters the harness sent.
- The model is a quantized 35B mixture with 3B active parameters and
  thinking disabled, the smaller of the two the endpoint serves; the larger
  one did not run.
- The tasks are the earlier runs' tasks, so a comparison between families
  is paired by task; the runs are hours apart and the engine changed in
  between, so every table here comes from one re-measurement.
- Requested lengths are short (about 22 words), because the originals are
  comment paragraphs; a one-sentence response has no paragraph unit.
- The overlap flag is the only contamination control.

## What is in this directory

- `tasks.json`, `requests.json`, `responses.json`, `records.json`: the
  saved records, with every input and output text and its hash.
- `shards/`: the controlled cohort's 25 shard manifests of this run.
- `dataset-plan.json`, `tables.json`, `tables-paragraph.json`,
  `tables-sentence.json`, `paired.json`: the six-cohort plan and tables of
  the re-measurement, with every run of the day in the controlled cohort.
- `tables-paragraph-qwen.json`: the paragraph tables with the controlled
  cohort restricted to this run's shards.
- `digests.json`: SHA-256 of every record, shard, pinned copy, finding
  artifact, and output of the measurement, which are reproducible from the
  records and the checkouts.
