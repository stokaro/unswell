# Stage C run 2, September 11, 2026

This directory is the record of the second run of stage C of the
[pattern protocol](../../../methods/llm-patterns-v1.md), after the
[pilot](../2026-09-11-pilot/README.md). It ran under amendment 1: one
generator family, Anthropic Claude, through session agents of the
flat-rate subscription, with no paid call. Every number is a rule outcome
under one policy. It is not a false-positive rate, not recall, not a share
of LLM text, and not a decision. The run adds endpoints to the controlled
cohort. It confirms no hypothesis.

## What ran

Tasks came from the historical cohort of the
[period run](../../../acquisition/runs/2026-09-11-periods/README.md):
comment paragraphs of at least twelve words in the training and development
partitions, each followed by a declaration. The draw used the pilot's seed
and left out the pilot's task set through `corpus tasks --exclude-tasks`,
so no task repeats. It is stratified by ecosystem:

| Ecosystem | Eligible | Selected |
| --- | --- | --- |
| c | 4 | 1 |
| cpp | 2 | 1 |
| go | 604 | 70 |
| javascript | 184 | 22 |
| python | 32 | 5 |
| rust | 356 | 41 |

The eligible frame is smaller than the pilot's. The candidate artifacts
under `artifacts/measurement/candidates/` were rebuilt after the pilot, and
the pilot's digests pin shard manifests, not candidates. Of the pilot's 60
tasks, 27 sit in the current frame and were left out; the other 33 are
outside it. The 140 tasks span 14 repositories: 131 in the training
partition and 9 in development.

Each task met two operations, `generate` from its fact sheet and `polish`
of its original text, under the two frozen prompt conditions, `neutral`
and `plain`: 560 requests. Every request went to one instruction-free
agent with the request text and one line telling the agent to use no
tools. The first 150 requests went through session agents launched one by
one, the other 410 through two workflow runs of the same agent type. The
560 agent transcripts name the model `claude-opus-5`; each agent made
exactly one call, the answer, and used no tool. Decoding parameters,
remote request identifiers, and cost are `unavailable`, as the amendment
states. The shards of this run carry the suffix `-run2` through
`corpus generations --shard-suffix`, so they sit beside the pilot's shards
of the same repositories.

## Coverage

| Operation | Prompt | Responses | Requested words | Realized words | Off length | Mean overlap | Overlap above 0.5 | Fact-sheet identifiers carried |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `generate` | `neutral` | 140 | 22.3 | 22.1 | 0 | 0.091 | 13 | 349 of 587 |
| `generate` | `plain` | 140 | 22.3 | 20.5 | 10 | 0.053 | 8 | 304 of 587 |
| `polish` | `neutral` | 140 | 22.3 | 22.7 | 3 | 0.493 | 75 | 170 of 587 |
| `polish` | `plain` | 140 | 22.3 | 22.0 | 3 | 0.380 | 58 | 169 of 587 |

All 560 responses are complete; none was refused, truncated, or failed.
Overlap is the share of response tokens inside word 8-gram matches with
the original. For `polish` a high overlap is the operation itself. For
`generate` it is a contamination signal: 21 responses repeat the original
wording from the fact sheet alone. Most of them are testify assertion
comments of the form "X asserts that ...", reproduced from the signature.
Those responses stay in the primary analysis and are listed for the
sensitivity analysis, as the protocol requires; the record marks
contamination by public training data as unknown for every response.

Unlike the pilot, almost every `generate` response came back as one prose
paragraph: 6 of 280 carry a heading or a fenced code block. The extraction
found 1,445 units in the 560 controlled documents, 384 of them paragraphs,
and 33 findings in 11,397 prose words (2.9 per 1,000 words). Twelve of the
findings are long sentences.

## Paired tables

The [paired tables](paired.json) pair each original unit with its response
document. All 140 tasks pair in every arm, across 25 provenance
components. Rules whose class admits the comment role enter; 31 of 40 such
rules fired in no response of any arm. Rows with any finding, original or
response:

| Rule | Arm | Originals with finding | Responses with finding | Paired change (95% interval) |
| --- | --- | --- | --- | --- |
| `filler.weak-intensifiers` | `generate`/`neutral` | 1 | 0 | -0.7 points (-2.6 to +0.0) |
| `filler.weak-intensifiers` | `generate`/`plain` | 1 | 0 | -0.7 points (-2.6 to +0.0) |
| `filler.weak-intensifiers` | `polish`/`neutral` | 1 | 1 | +0.0 points (+0.0 to +0.0) |
| `filler.weak-intensifiers` | `polish`/`plain` | 1 | 0 | -0.7 points (-2.6 to +0.0) |
| `filler.wordy-phrase` | `generate`/`neutral` | 2 | 0 | -1.4 points (-3.3 to +0.0) |
| `filler.wordy-phrase` | `generate`/`plain` | 2 | 0 | -1.4 points (-3.3 to +0.0) |
| `filler.wordy-phrase` | `polish`/`neutral` | 2 | 2 | +0.0 points (+0.0 to +0.0) |
| `filler.wordy-phrase` | `polish`/`plain` | 2 | 1 | -0.7 points (-2.4 to +0.0) |
| `readability.grade-metric` | `generate`/`neutral` | 2 | 0 | -1.4 points (-4.3 to +0.0) |
| `readability.grade-metric` | `generate`/`plain` | 2 | 0 | -1.4 points (-4.3 to +0.0) |
| `readability.grade-metric` | `polish`/`neutral` | 2 | 1 | -0.7 points (-2.8 to +0.0) |
| `readability.grade-metric` | `polish`/`plain` | 2 | 2 | +0.0 points (-2.5 to +1.8) |
| `repetition.exact-sentence` | `generate`/`neutral` | 6 | 0 | -4.3 points (-7.8 to +0.0) |
| `repetition.exact-sentence` | `generate`/`plain` | 6 | 0 | -4.3 points (-7.8 to +0.0) |
| `repetition.exact-sentence` | `polish`/`neutral` | 6 | 0 | -4.3 points (-7.8 to +0.0) |
| `repetition.exact-sentence` | `polish`/`plain` | 6 | 0 | -4.3 points (-7.8 to +0.0) |
| `repetition.near-sentence` | `generate`/`neutral` | 1 | 0 | -0.7 points (-1.9 to +0.0) |
| `repetition.near-sentence` | `generate`/`plain` | 1 | 0 | -0.7 points (-1.9 to +0.0) |
| `repetition.near-sentence` | `polish`/`neutral` | 1 | 0 | -0.7 points (-1.9 to +0.0) |
| `repetition.near-sentence` | `polish`/`plain` | 1 | 0 | -0.7 points (-1.9 to +0.0) |
| `repetition.ngram-density` | `generate`/`neutral` | 2 | 0 | -1.4 points (-5.5 to +0.0) |
| `repetition.ngram-density` | `generate`/`plain` | 2 | 0 | -1.4 points (-5.5 to +0.0) |
| `repetition.ngram-density` | `polish`/`neutral` | 2 | 1 | -0.7 points (-2.7 to +0.0) |
| `repetition.ngram-density` | `polish`/`plain` | 2 | 2 | +0.0 points (+0.0 to +0.0) |
| `repetition.paragraph-overlap` | `generate`/`neutral` | 2 | 0 | -1.4 points (-3.8 to +0.0) |
| `repetition.paragraph-overlap` | `generate`/`plain` | 2 | 0 | -1.4 points (-3.8 to +0.0) |
| `repetition.paragraph-overlap` | `polish`/`neutral` | 2 | 0 | -1.4 points (-3.8 to +0.0) |
| `repetition.paragraph-overlap` | `polish`/`plain` | 2 | 0 | -1.4 points (-3.8 to +0.0) |
| `repetition.sentence-openers` | `polish`/`neutral` | 0 | 1 | +0.7 points (+0.0 to +2.8) |
| `repetition.sentence-openers` | `polish`/`plain` | 0 | 1 | +0.7 points (+0.0 to +2.8) |
| `syntax.long-sentence` | `generate`/`neutral` | 8 | 0 | -5.7 points (-13.0 to -1.2) |
| `syntax.long-sentence` | `generate`/`plain` | 8 | 0 | -5.7 points (-13.0 to -1.2) |
| `syntax.long-sentence` | `polish`/`neutral` | 8 | 6 | -1.4 points (-4.3 to +0.0) |
| `syntax.long-sentence` | `polish`/`plain` | 8 | 6 | -1.4 points (-4.7 to +1.4) |
| `syntax.noun-stack` | `generate`/`neutral` | 0 | 1 | +0.7 points (+0.0 to +2.2) |
| `syntax.parenthetical-load` | `generate`/`neutral` | 5 | 0 | -3.6 points (-10.0 to +0.0) |
| `syntax.parenthetical-load` | `generate`/`plain` | 5 | 0 | -3.6 points (-10.0 to +0.0) |
| `syntax.parenthetical-load` | `polish`/`neutral` | 5 | 2 | -2.1 points (-5.3 to +0.0) |
| `syntax.parenthetical-load` | `polish`/`plain` | 5 | 2 | -2.1 points (-5.3 to +0.0) |
| `syntax.passive-candidate-density` | `generate`/`neutral` | 5 | 1 | -2.9 points (-6.7 to +0.0) |
| `syntax.passive-candidate-density` | `generate`/`plain` | 5 | 0 | -3.6 points (-6.8 to -1.1) |
| `syntax.passive-candidate-density` | `polish`/`neutral` | 5 | 2 | -2.1 points (-5.6 to +0.0) |
| `syntax.passive-candidate-density` | `polish`/`plain` | 5 | 1 | -2.9 points (-6.5 to -0.6) |

These are small counts over 140 pairs. The two intervals that exclude
zero belong to `syntax.long-sentence` under `generate` and to
`syntax.passive-candidate-density` under `generate`/`plain`: the
generated text drops a finding the original carried. That is the
operation writing a shorter or a more direct sentence. It is not a
property of the cohort, and a `polish` response keeps most of the same
findings. Two rules fire in a response and not in its original:
`repetition.sentence-openers` once under each `polish` prompt, and
`syntax.noun-stack` once under `generate`/`neutral`. The tables also
hold a `difference_from_h0` field. It compares a one-paragraph response
with whole H0 documents of the comment role, so document length dominates
it. It stays because the protocol names it. The paragraph-level comparison
below is the fair one.

## Paragraph-level comparison

The [run tables](tables-paragraph-run2.json) count the controlled cohort
from this run's shards only: 376 documents with a paragraph unit, 384
paragraphs, 13 provenance components, 7,387 prose words. Controlled
paragraphs of all arms against all historical paragraphs:

| Rule | Historical | Controlled |
| --- | --- | --- |
| `syntax.long-sentence` | 1,415 of 46,200 (3.1%) | 8 of 384 (2.1%) |
| `syntax.parenthetical-load` | 455 of 46,200 (1.0%) | 4 of 384 (1.0%) |
| `syntax.passive-candidate-density` | 586 of 46,200 (1.3%) | 4 of 384 (1.0%) |
| `repetition.ngram-density` | 96 of 46,200 (0.2%) | 3 of 384 (0.8%) |
| `filler.wordy-phrase` | 78 of 46,200 (0.2%) | 1 of 384 (0.3%) |
| `readability.grade-metric` | 902 of 46,200 (2.0%) | 0 of 384 (0.0%) |
| `syntax.noun-stack` | 43 of 46,200 (0.1%) | 0 of 384 (0.0%) |
| `filler.weak-intensifiers` | 19 of 46,200 (0.0%) | 0 of 384 (0.0%) |
| `format.em-dash-density` | 0 of 46,200 (0.0%) | 0 of 384 (0.0%) |

With 384 controlled paragraphs, a rule at 1% prevalence fires about four
times. No difference reaches the protocol's minimum useful difference of
three points. The whole controlled cohort, pilot and this run together, has
529 paragraphs in 511 documents across 32 components; the
[six-cohort tables](tables-paragraph.json) count it that way.

## Limits

- One family. Every other family is `untested`; protocol version 2 needs
  three.
- The agent harness ran the search agent type with an instruction-free
  system prompt; the agents used no tools, but the harness cannot show that
  the model saw nothing else. Two launch paths ran the same agent type; the
  record does not separate them.
- The eligible frame differs from the pilot's, so the two runs are two
  draws from two frames of the same cohort, not one draw of 200 tasks.
- Requested lengths are short (about 22 words), because the originals are
  comment paragraphs; a one-sentence response has no paragraph unit.
- The generator reproduced public text in 21 `generate` responses; the
  overlap flag is the only contamination control.

## What is in this directory

- `tasks.json`, `requests.json`, `responses.json`, `records.json`: the
  saved records, with every input and output text and its hash.
- `shards/`: the controlled cohort's 14 shard manifests of this run.
- `dataset-plan.json`, `tables.json`, `tables-paragraph.json`,
  `tables-sentence.json`, `paired.json`: the six-cohort plan and tables
  after this run joined the controlled cohort.
- `tables-paragraph-run2.json`: the paragraph tables with the controlled
  cohort restricted to this run's shards.
- `digests.json`: SHA-256 of every record, shard, pinned copy, finding
  artifact, and output of the measurement, which are reproducible from the
  records and the checkouts.
