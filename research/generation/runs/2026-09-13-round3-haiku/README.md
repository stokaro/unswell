# Stage C run 9, September 13, 2026: the arm reaches the cluster minimum

This directory is the record of the ninth run of stage C of the
[pattern protocol](../../../methods/llm-patterns-v1.md). It ran under
amendment 6, the third acquisition round, which adds two repositories to
the `development` partition and four to `final_test`. The family is
Anthropic Claude through session agents of the flat-rate subscription, with
no paid call, and the model is `claude-haiku-4-5-20251001`. Every number is
a rule outcome under one policy. It is not a false-positive rate, not
recall, not a share of LLM text, and not a decision. This run and its
[OpenAI counterpart](../2026-09-13-round3-luna/README.md) lift the
controlled arm of the development partition to 20 provenance components, so
the screening that follows is the first whose arms meet the protocol's
cluster minimum. It confirms no hypothesis.

## What ran

`corpus tasks` drew 150 tasks from the historical cohort of the
`development` partition with the comment role and the seed
`unswell-llm-patterns-v1`. The candidate artifacts were restricted to the
four development repositories that carry comment paragraphs and held no
controlled response: google/googletest and httpie/cli, pinned in amendment
5, and assertj/assertj-core and AutoMapper/AutoMapper, pinned in amendment
6. Five earlier task sets went to `--exclude-tasks`, and they carry all 350
task IDs the eight earlier runs used, so no task repeats. The draw is
stratified by ecosystem:

| Ecosystem | Eligible | Selected |
| --- | --- | --- |
| cpp | 18 | 3 |
| csharp | 117 | 16 |
| java | 1,055 | 131 |

httpie/cli appears in neither row. It carries 183 historical comment
paragraphs, and not one of them is an eligible task: a task needs at least
twelve words and a declaration following the paragraph. So the 150 tasks
span three repositories, assertj/assertj-core 131, AutoMapper/AutoMapper 16
and google/googletest 3, and the controlled arm gains exactly three
components. That is the minimum the round needed, with no margin.

Each task met the two operations, `generate` from its fact sheet and
`polish` of its original text, under the two frozen prompt conditions,
`neutral` and `plain`: 600 requests. Every request went to one
instruction-free agent with the request text and one line telling the agent
to use no tools. The requests ran as six workflow batches of 100. The 600
agent transcripts name the model `claude-haiku-4-5-20251001`; each agent
made exactly one call, the answer, and used no tool. Decoding parameters,
remote request identifiers, and cost are `unavailable`, as amendment 1
states. The shards of this run carry the suffix `-haiku-r3` through
`corpus generations --shard-suffix`.

The measurement of this run's shards needed the Markdown fix of #227.
Without it a single generated sentence, `Invoke assertThat(long[][]) on the
array.`, fails extraction and the whole assertj shard is skipped.

## Coverage

| Operation | Prompt | Responses | Requested words | Realized words | Off length | Mean overlap | Overlap above 0.5 | Fact-sheet identifiers carried |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `generate` | `neutral` | 150 | 25.0 | 24.6 | 0 | 0.010 | 2 | 313 of 690 |
| `generate` | `plain` | 150 | 25.0 | 24.7 | 0 | 0.010 | 2 | 303 of 690 |
| `polish` | `neutral` | 150 | 25.0 | 25.1 | 8 | 0.475 | 74 | 172 of 690 |
| `polish` | `plain` | 150 | 25.0 | 24.8 | 4 | 0.371 | 59 | 157 of 690 |

All 600 responses are complete; none was refused, truncated, or failed.
Overlap is the share of response tokens inside word 8-gram matches with the
original. For `polish` a high overlap is the operation itself. Among the
300 `generate` responses six share an 8-gram with the original and 46 share
a 4-gram, and four raise the overlap flag. None carries a heading or a
fenced block, and none runs to more than one paragraph. The record marks
contamination by public training data as unknown for every response.

The extraction found 1,549 units in the 600 controlled documents: 856
sentences, 538 paragraphs, and 155 fragments. It found 56 findings in
14,248 prose words (3.9 per 1,000 words) in 49 documents. Eighteen are
reading grade, 15 passive-candidate density, ten noun stacks, eight long
sentences, two n-gram density, two parenthetical load, and one em-dash
density.

## Paired tables

The [paired tables](paired.json) pair each original unit with its response
document. All 150 tasks pair in every arm, across 14 provenance components.
Rules whose class admits the comment role enter; 30 of 40 such rules fired
in no response and no original of any arm. Rows with any finding:

| Rule | Arm | Originals with finding | Responses with finding | Paired change (95% interval) |
| --- | --- | --- | --- | --- |
| `format.em-dash-density` | `generate`/`plain` | 0 | 1 | +0.7 points (+0.0 to +2.2) |
| `readability.grade-metric` | `generate`/`neutral` | 0 | 7 | +4.7 points (+0.7 to +10.7) |
| `readability.grade-metric` | `generate`/`plain` | 0 | 6 | +4.0 points (+0.0 to +10.3) |
| `readability.grade-metric` | `polish`/`neutral` | 0 | 2 | +1.3 points (+0.0 to +3.2) |
| `readability.grade-metric` | `polish`/`plain` | 0 | 3 | +2.0 points (+0.0 to +4.8) |
| `repetition.exact-sentence` | `generate`/`neutral` | 6 | 0 | -4.0 points (-6.2 to -1.5) |
| `repetition.exact-sentence` | `generate`/`plain` | 6 | 0 | -4.0 points (-6.2 to -1.5) |
| `repetition.exact-sentence` | `polish`/`neutral` | 6 | 0 | -4.0 points (-6.2 to -1.5) |
| `repetition.exact-sentence` | `polish`/`plain` | 6 | 0 | -4.0 points (-6.2 to -1.5) |
| `repetition.near-sentence` | `generate`/`neutral` | 2 | 0 | -1.3 points (-3.2 to +0.0) |
| `repetition.near-sentence` | `generate`/`plain` | 2 | 0 | -1.3 points (-3.2 to +0.0) |
| `repetition.near-sentence` | `polish`/`neutral` | 2 | 0 | -1.3 points (-3.2 to +0.0) |
| `repetition.near-sentence` | `polish`/`plain` | 2 | 0 | -1.3 points (-3.2 to +0.0) |
| `repetition.ngram-density` | `generate`/`neutral` | 6 | 0 | -4.0 points (-9.6 to +0.0) |
| `repetition.ngram-density` | `generate`/`plain` | 6 | 1 | -3.3 points (-9.3 to +1.7) |
| `repetition.ngram-density` | `polish`/`neutral` | 6 | 1 | -3.3 points (-9.4 to +1.6) |
| `repetition.ngram-density` | `polish`/`plain` | 6 | 0 | -4.0 points (-9.6 to +0.0) |
| `repetition.paragraph-overlap` | `generate`/`neutral` | 5 | 0 | -3.3 points (-5.1 to -0.8) |
| `repetition.paragraph-overlap` | `generate`/`plain` | 5 | 0 | -3.3 points (-5.1 to -0.8) |
| `repetition.paragraph-overlap` | `polish`/`neutral` | 5 | 0 | -3.3 points (-5.1 to -0.8) |
| `repetition.paragraph-overlap` | `polish`/`plain` | 5 | 0 | -3.3 points (-5.1 to -0.8) |
| `syntax.long-sentence` | `generate`/`neutral` | 11 | 1 | -6.7 points (-17.0 to -0.7) |
| `syntax.long-sentence` | `generate`/`plain` | 11 | 4 | -4.7 points (-14.3 to +2.4) |
| `syntax.long-sentence` | `polish`/`neutral` | 11 | 2 | -6.0 points (-16.5 to +0.0) |
| `syntax.long-sentence` | `polish`/`plain` | 11 | 1 | -6.7 points (-18.5 to +0.0) |
| `syntax.noun-stack` | `generate`/`neutral` | 2 | 5 | +2.0 points (-2.8 to +8.4) |
| `syntax.noun-stack` | `generate`/`plain` | 2 | 1 | -0.7 points (-4.5 to +1.9) |
| `syntax.noun-stack` | `polish`/`neutral` | 2 | 1 | -0.7 points (-2.5 to +0.0) |
| `syntax.noun-stack` | `polish`/`plain` | 2 | 3 | +0.7 points (+0.0 to +2.1) |
| `syntax.parenthetical-load` | `generate`/`neutral` | 11 | 0 | -7.3 points (-14.6 to -1.0) |
| `syntax.parenthetical-load` | `generate`/`plain` | 11 | 0 | -7.3 points (-14.6 to -1.0) |
| `syntax.parenthetical-load` | `polish`/`neutral` | 11 | 2 | -6.0 points (-13.1 to +0.0) |
| `syntax.parenthetical-load` | `polish`/`plain` | 11 | 0 | -7.3 points (-14.6 to -1.0) |
| `syntax.passive-candidate-density` | `generate`/`neutral` | 1 | 0 | -0.7 points (-2.2 to +0.0) |
| `syntax.passive-candidate-density` | `generate`/`plain` | 1 | 0 | -0.7 points (-2.2 to +0.0) |
| `syntax.passive-candidate-density` | `polish`/`neutral` | 1 | 9 | +5.3 points (+0.0 to +15.7) |
| `syntax.passive-candidate-density` | `polish`/`plain` | 1 | 6 | +3.3 points (+0.0 to +10.9) |

Thirteen intervals exclude zero. Twelve are negative and repeat the earlier
runs: repetition rows measure the context of the original rather than the
model, and the long-sentence and parenthetical rows are the operation
writing a shorter or plainer sentence.

One is positive and its point estimate clears the minimum useful difference
of three points: `readability.grade-metric` under `generate`/`neutral` at
+4.7 points, from +0.7 to +10.7. Seven responses carry a reading-grade
finding against none of their originals. The second round found the same
shape in `syntax.noun-stack`; here that rule reaches +2.0 points with an
interval spanning zero. Two arms of two runs are not a confirmed
construction. The [screening record](../../../methods/screening/2026-09-13-development-3.json)
is what decides a candidate.

The tables also hold a `difference_from_h0` field. It compares a
one-paragraph response with whole H0 documents of the comment role, so
document length dominates it. It stays because the protocol names it. The
paragraph-level comparison below is the fair one.

## Paragraph-level comparison

The [run tables](tables-paragraph-haiku-r3.json) count the controlled cohort
from this run's shards only: 528 documents with a paragraph unit, 538
paragraphs, 3 provenance components, 12,419 prose words. Controlled
paragraphs of all arms against all historical paragraphs:

| Rule | Historical | Controlled |
| --- | --- | --- |
| `syntax.long-sentence` | 3,019 of 81,690 (3.7%) | 7 of 538 (1.3%) |
| `syntax.passive-candidate-density` | 1,954 of 81,690 (2.4%) | 12 of 538 (2.2%) |
| `readability.grade-metric` | 1,604 of 81,690 (2.0%) | 12 of 538 (2.2%) |
| `repetition.exact-sentence` | 1,217 of 81,690 (1.5%) | 0 of 538 (0.0%) |
| `syntax.parenthetical-load` | 974 of 81,690 (1.2%) | 2 of 538 (0.4%) |
| `repetition.paragraph-overlap` | 710 of 81,690 (0.9%) | 0 of 538 (0.0%) |
| `repetition.ngram-density` | 407 of 81,690 (0.5%) | 2 of 538 (0.4%) |
| `repetition.near-sentence` | 320 of 81,690 (0.4%) | 0 of 538 (0.0%) |
| `repetition.sentence-openers` | 309 of 81,690 (0.4%) | 0 of 538 (0.0%) |
| `readability.long-paragraph` | 279 of 81,690 (0.3%) | 0 of 538 (0.0%) |
| `filler.wordy-phrase` | 248 of 81,690 (0.3%) | 0 of 538 (0.0%) |
| `repetition.paragraph-openers` | 163 of 81,690 (0.2%) | 0 of 538 (0.0%) |
| `syntax.noun-stack` | 97 of 81,690 (0.1%) | 7 of 538 (1.3%) |
| `filler.weak-intensifiers` | 32 of 81,690 (0.0%) | 0 of 538 (0.0%) |
| `repetition.syntax-template` | 19 of 81,690 (0.0%) | 0 of 538 (0.0%) |
| `format.em-dash-density` | 4 of 81,690 (0.0%) | 0 of 538 (0.0%) |
| `syntax.nominalization-chain` | 3 of 81,690 (0.0%) | 0 of 538 (0.0%) |
| `filler.announced-importance` | 1 of 81,690 (0.0%) | 0 of 538 (0.0%) |

No difference reaches three points here. The largest positive one is
`syntax.noun-stack`, 1.2 points above a historical rate of 0.1%, which is a
eleven-fold rate on seven paragraphs. The largest negative one is
`syntax.long-sentence`, 2.4 points below. These are 538 paragraphs in 3
components, well under the cluster minimum; the screening is what counts
across the whole partition. The whole controlled cohort, ten runs together,
has 5,303 paragraphs in 5,038 documents across 152 components; the
[six-cohort tables](tables-paragraph.json) count it that way.

## Limits

- One family, two models. The comparison against the OpenAI family is the
  [luna round-three run](../2026-09-13-round3-luna/README.md), which
  answered the same 150 tasks.
- Three components carry this run's paragraphs, and 131 of the 150 tasks
  come from one repository. The draw is seeded and stratified by ecosystem,
  and java carries 1,055 of the 1,190 eligible units, so the imbalance is
  the frame's, not a choice.
- The controlled arm of the partition reaches exactly 20 components. One
  repository yielding no paragraph would have left it at 19.
- The agent harness ran the search agent type with an instruction-free
  system prompt; the agents used no tools, but the harness cannot show that
  the model saw nothing else.
- The overlap flag is the only contamination control. Four `generate`
  responses raise it and six share an 8-gram with their original.

## What is in this directory

- `tasks.json`, `requests.json`, `responses.json`, `records.json`: the
  saved records, with every input and output text and its hash.
- `shards/`: the controlled cohort's 3 shard manifests of this run.
- `dataset-plan.json`, `tables.json`, `tables-paragraph.json`,
  `tables-sentence.json`, `paired.json`: the six-cohort plan and tables
  after this run joined the controlled cohort.
- `tables-paragraph-haiku-r3.json`: the paragraph tables with the
  controlled cohort restricted to this run's shards.
- `digests.json`: SHA-256 of every record, shard, pinned copy, finding
  artifact, and output of the measurement, which are reproducible from the
  records and the checkouts.
