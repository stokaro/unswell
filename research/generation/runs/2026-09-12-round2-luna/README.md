# Stage C run 8, September 12, 2026: the OpenAI family on new sources

This directory is the record of the eighth run of stage C of the
[pattern protocol](../../../methods/llm-patterns-v1.md). It ran under
amendment 5, the second acquisition round, on the same 150 tasks as the
[Haiku round-two run](../2026-09-12-round2-haiku/README.md), so the two
families are paired by task on repositories no earlier run had seen. The
family is OpenAI through the Codex CLI of the flat-rate subscription, with
no API call, and the model is `gpt-5.6-luna`. Every number is a rule
outcome under one policy. It is not a false-positive rate, not recall, not
a share of LLM text, and not a decision. It confirms no hypothesis.

## What ran

The task set is the Haiku round-two run's task set: the same 150 task IDs
and texts, drawn once by `corpus tasks` from the historical cohort of the
`development` partition with the comment role and the seed
`unswell-llm-patterns-v1`. The candidate artifacts were restricted to the
ten repositories the second acquisition round pinned to `development`, so
the eligible frame below counts those ten and no earlier source. The draw
is stratified by ecosystem:

| Ecosystem | Eligible | Selected |
| --- | --- | --- |
| cpp | 56 | 7 |
| csharp | 721 | 74 |
| java | 75 | 8 |
| javascript | 294 | 31 |
| rust | 283 | 30 |

The 150 tasks span the seven newly acquired development repositories that
carry an eligible comment paragraph: App-vNext/Polly 66, Leaflet/Leaflet
31, ogham/exa 27, winsw/winsw 8, skylot/jadx 8, google/leveldb 7, and
sharkdp/fd 3.

Each task met the two operations and the two frozen prompt conditions: 600
requests. Every request ran as one non-interactive `codex exec` turn in a
read-only sandbox with an isolated home that carries no instructions and no
skills, six at a time. The model comes from the header the Codex CLI writes
on every run. The reasoning effort is `low`, the lowest the model accepts;
the CLI exposes no other decoding parameter, so the rest are `unavailable`.
The shards of this run carry the suffix `-luna-r2` through
`corpus generations --shard-suffix`.

One request returned no final message on its first attempt. It ran a second
time unchanged and returned text, and its response row records the retry.
Its text asks for the material rather than writing the document, which the
record keeps as the model's answer.

## Coverage

| Operation | Prompt | Responses | Requested words | Realized words | Off length | Mean overlap | Overlap above 0.5 | Fact-sheet identifiers carried |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `generate` | `neutral` | 150 | 48.0 | 41.4 | 3 | 0.000 | 0 | 711 of 775 |
| `generate` | `plain` | 150 | 48.0 | 40.5 | 5 | 0.000 | 0 | 686 of 775 |
| `polish` | `neutral` | 150 | 48.0 | 47.5 | 0 | 0.541 | 95 | 354 of 775 |
| `polish` | `plain` | 150 | 48.0 | 47.8 | 0 | 0.410 | 70 | 352 of 775 |

All 600 responses are complete; none was refused or truncated. Overlap is
the share of response tokens inside word 8-gram matches with the original.
For `polish` a high overlap is the operation itself. No `generate` response
shares an 8-gram or even a 4-gram with its original: this model writes from
the fact sheet without reproducing the comment. It carries far more
fact-sheet identifiers than Haiku does, 711 of 775 against 522 under
`generate`/`neutral`, and 38 of the 300 `generate` responses carry a
heading or a fenced block. The record marks contamination by public
training data as unknown for every response.

The extraction found 2,176 units in the 600 controlled documents: 476
sentences, 367 paragraphs, and 1,333 fragments. The fragment share is far
higher than Haiku's, because this model answers in short labeled lines that
carry no sentence. It found 12 findings in 12,870 prose words (0.9 per
1,000 words) in 10 documents: seven passive-candidate density, two n-gram
density, two reading grade, and one sentence openers. The Haiku run found
140 findings in 23,823 words, 5.9 per 1,000.

## Paired tables

The [paired tables](paired.json) pair each original unit with its response
document. All 150 tasks pair in every arm, across 21 provenance components.
Rules whose class admits the comment role enter; 30 of 40 such rules fired
in no response and no original of any arm. Rows with any finding:

| Rule | Arm | Originals with finding | Responses with finding | Paired change (95% interval) |
| --- | --- | --- | --- | --- |
| `readability.grade-metric` | `generate`/`neutral` | 27 | 0 | -18.0 points (-37.7 to -2.6) |
| `readability.grade-metric` | `generate`/`plain` | 27 | 0 | -18.0 points (-37.7 to -2.6) |
| `readability.grade-metric` | `polish`/`neutral` | 27 | 1 | -17.3 points (-37.1 to -1.9) |
| `readability.grade-metric` | `polish`/`plain` | 27 | 1 | -17.3 points (-37.1 to -1.9) |
| `readability.long-paragraph` | `generate`/`neutral` | 16 | 0 | -10.7 points (-30.5 to +0.0) |
| `readability.long-paragraph` | `generate`/`plain` | 16 | 0 | -10.7 points (-30.5 to +0.0) |
| `readability.long-paragraph` | `polish`/`neutral` | 16 | 0 | -10.7 points (-30.5 to +0.0) |
| `readability.long-paragraph` | `polish`/`plain` | 16 | 0 | -10.7 points (-30.5 to +0.0) |
| `repetition.exact-sentence` | `generate`/`neutral` | 17 | 0 | -11.3 points (-22.4 to -0.9) |
| `repetition.exact-sentence` | `generate`/`plain` | 17 | 0 | -11.3 points (-22.4 to -0.9) |
| `repetition.exact-sentence` | `polish`/`neutral` | 17 | 0 | -11.3 points (-22.4 to -0.9) |
| `repetition.exact-sentence` | `polish`/`plain` | 17 | 0 | -11.3 points (-22.4 to -0.9) |
| `repetition.near-sentence` | `generate`/`neutral` | 4 | 0 | -2.7 points (-6.2 to +0.0) |
| `repetition.near-sentence` | `generate`/`plain` | 4 | 0 | -2.7 points (-6.2 to +0.0) |
| `repetition.near-sentence` | `polish`/`neutral` | 4 | 0 | -2.7 points (-6.2 to +0.0) |
| `repetition.near-sentence` | `polish`/`plain` | 4 | 0 | -2.7 points (-6.2 to +0.0) |
| `repetition.ngram-density` | `generate`/`neutral` | 11 | 0 | -7.3 points (-15.5 to -0.8) |
| `repetition.ngram-density` | `generate`/`plain` | 11 | 0 | -7.3 points (-15.5 to -0.8) |
| `repetition.ngram-density` | `polish`/`neutral` | 11 | 1 | -6.7 points (-15.1 to +0.7) |
| `repetition.ngram-density` | `polish`/`plain` | 11 | 0 | -7.3 points (-15.5 to -0.8) |
| `repetition.paragraph-overlap` | `generate`/`neutral` | 10 | 0 | -6.7 points (-15.0 to +0.0) |
| `repetition.paragraph-overlap` | `generate`/`plain` | 10 | 0 | -6.7 points (-15.0 to +0.0) |
| `repetition.paragraph-overlap` | `polish`/`neutral` | 10 | 0 | -6.7 points (-15.0 to +0.0) |
| `repetition.paragraph-overlap` | `polish`/`plain` | 10 | 0 | -6.7 points (-15.0 to +0.0) |
| `repetition.sentence-openers` | `generate`/`neutral` | 0 | 1 | +0.7 points (+0.0 to +1.9) |
| `syntax.long-sentence` | `generate`/`neutral` | 24 | 0 | -16.0 points (-36.2 to -1.0) |
| `syntax.long-sentence` | `generate`/`plain` | 24 | 0 | -16.0 points (-36.2 to -1.0) |
| `syntax.long-sentence` | `polish`/`neutral` | 24 | 0 | -16.0 points (-36.2 to -1.0) |
| `syntax.long-sentence` | `polish`/`plain` | 24 | 0 | -16.0 points (-36.2 to -1.0) |
| `syntax.parenthetical-load` | `generate`/`neutral` | 1 | 0 | -0.7 points (-2.5 to +0.0) |
| `syntax.parenthetical-load` | `generate`/`plain` | 1 | 0 | -0.7 points (-2.5 to +0.0) |
| `syntax.parenthetical-load` | `polish`/`neutral` | 1 | 0 | -0.7 points (-2.5 to +0.0) |
| `syntax.parenthetical-load` | `polish`/`plain` | 1 | 0 | -0.7 points (-2.5 to +0.0) |
| `syntax.passive-candidate-density` | `generate`/`neutral` | 12 | 4 | -5.3 points (-14.3 to +1.9) |
| `syntax.passive-candidate-density` | `generate`/`plain` | 12 | 1 | -7.3 points (-14.3 to -1.8) |
| `syntax.passive-candidate-density` | `polish`/`neutral` | 12 | 1 | -7.3 points (-15.5 to -1.2) |
| `syntax.passive-candidate-density` | `polish`/`plain` | 12 | 1 | -7.3 points (-15.5 to -1.2) |

Eighteen intervals exclude zero and every one of them is negative. No rule
fires in more responses than originals in any arm except
`repetition.sentence-openers`, once, with an interval that touches zero.
`syntax.noun-stack`, the one rule whose paired interval is positive and
above the minimum useful difference in the Haiku run, fires in no response
of this run at all.

The tables also hold a `difference_from_h0` field. It compares a
one-paragraph response with whole H0 documents of the comment role, so
document length dominates it. It stays because the protocol names it. The
paragraph-level comparison below is the fair one.

## Paragraph-level comparison

The [run tables](tables-paragraph-luna-r2.json) count the controlled cohort
from this run's shards only: 349 documents with a paragraph unit, 367
paragraphs, 7 provenance components, 6,259 prose words. Controlled
paragraphs of all arms against all historical paragraphs:

| Rule | Historical | Controlled |
| --- | --- | --- |
| `syntax.long-sentence` | 2,723 of 69,579 (3.9%) | 0 of 367 (0.0%) |
| `syntax.passive-candidate-density` | 1,824 of 69,579 (2.6%) | 5 of 367 (1.4%) |
| `readability.grade-metric` | 1,513 of 69,579 (2.2%) | 2 of 367 (0.5%) |
| `repetition.exact-sentence` | 928 of 69,579 (1.3%) | 0 of 367 (0.0%) |
| `syntax.parenthetical-load` | 706 of 69,579 (1.0%) | 0 of 367 (0.0%) |
| `repetition.paragraph-overlap` | 534 of 69,579 (0.8%) | 0 of 367 (0.0%) |
| `repetition.ngram-density` | 316 of 69,579 (0.5%) | 1 of 367 (0.3%) |
| `repetition.sentence-openers` | 296 of 69,579 (0.4%) | 0 of 367 (0.0%) |
| `readability.long-paragraph` | 246 of 69,579 (0.4%) | 0 of 367 (0.0%) |
| `filler.wordy-phrase` | 218 of 69,579 (0.3%) | 0 of 367 (0.0%) |
| `repetition.near-sentence` | 204 of 69,579 (0.3%) | 0 of 367 (0.0%) |
| `repetition.paragraph-openers` | 154 of 69,579 (0.2%) | 0 of 367 (0.0%) |
| `syntax.noun-stack` | 85 of 69,579 (0.1%) | 0 of 367 (0.0%) |
| `filler.weak-intensifiers` | 31 of 69,579 (0.0%) | 0 of 367 (0.0%) |
| `repetition.syntax-template` | 15 of 69,579 (0.0%) | 0 of 367 (0.0%) |
| `syntax.nominalization-chain` | 3 of 69,579 (0.0%) | 0 of 367 (0.0%) |
| `format.em-dash-density` | 2 of 69,579 (0.0%) | 0 of 367 (0.0%) |
| `filler.announced-importance` | 1 of 69,579 (0.0%) | 0 of 367 (0.0%) |

No difference is positive, and the largest negative one is
`syntax.long-sentence` at 3.9 points below the historical rate. With 367
paragraphs a rule at 1% prevalence fires about four times. These are 7
components, well under the cluster minimum of 20, and the
[screening record](../../../methods/screening/2026-09-12-development-2.json)
is what decides a candidate. The whole controlled cohort, eight runs
together, has 4,361 paragraphs in 4,121 documents across 146 components;
the [six-cohort tables](tables-paragraph.json) count it that way.

## Limits

- One family, one model, at the lowest reasoning effort the model accepts.
  A stronger effort is a different setting and would need its own run.
- Seven components carry this run's paragraphs. The protocol's cluster
  minimum is 20 per arm, so no card of this run alone can leave
  `inconclusive`.
- The Codex CLI reports the model in a header and offers no decoding
  parameter beyond the reasoning effort, so the run cannot record a
  temperature or a seed.
- The isolated home carries no instructions and no skills, but the harness
  cannot show that the model saw nothing else.
- One request needed a second attempt, and its recorded answer asks for the
  material instead of writing a document. It stays in every table.
- The overlap flag is the only contamination control, and no `generate`
  response raised it.

## What is in this directory

- `tasks.json`, `requests.json`, `responses.json`, `records.json`: the
  saved records, with every input and output text and its hash.
- `shards/`: the controlled cohort's 7 shard manifests of this run.
- `dataset-plan.json`, `tables.json`, `tables-paragraph.json`,
  `tables-sentence.json`, `paired.json`: the six-cohort plan and tables
  after this run joined the controlled cohort.
- `tables-paragraph-luna-r2.json`: the paragraph tables with the controlled
  cohort restricted to this run's shards.
- `digests.json`: SHA-256 of every record, shard, pinned copy, finding
  artifact, and output of the measurement, which are reproducible from the
  records and the checkouts.
