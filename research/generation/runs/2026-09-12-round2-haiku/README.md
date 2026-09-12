# Stage C run 7, September 12, 2026: new development sources

This directory is the record of the seventh run of stage C of the
[pattern protocol](../../../methods/llm-patterns-v1.md). It ran under
amendment 5, the second acquisition round, which adds ten repositories to
the `development` partition and fourteen to `final_test`. The family is
Anthropic Claude through session agents of the flat-rate subscription, with
no paid call, and the model is `claude-haiku-4-5-20251001`. Every number is
a rule outcome under one policy. It is not a false-positive rate, not
recall, not a share of LLM text, and not a decision. This run answers new
tasks from repositories no earlier run had seen. It confirms no hypothesis.

## What ran

`corpus tasks` drew 150 tasks from the historical cohort of the
`development` partition with the comment role and the seed
`unswell-llm-patterns-v1`. The candidate artifacts were restricted to the
ten repositories the second acquisition round pinned to `development`, so
the eligible frame below counts those ten and no earlier source. Four
earlier task sets went to `--exclude-tasks`, and they carry all 200 task IDs
the six earlier runs used, so no task repeats. The draw is stratified by
ecosystem:

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
sharkdp/fd 3. Requested lengths average 48 words, more than twice the 22
of the earlier runs, because these repositories document their APIs in
longer comment paragraphs.

Each task met the two operations, `generate` from its fact sheet and
`polish` of its original text, under the two frozen prompt conditions,
`neutral` and `plain`: 600 requests. Every request went to one
instruction-free agent with the request text and one line telling the agent
to use no tools. The requests ran as six workflow batches of 100. The 600
agent transcripts name the model `claude-haiku-4-5-20251001`; each agent
made exactly one call, the answer, and used no tool. Decoding parameters,
remote request identifiers, and cost are `unavailable`, as amendment 1
states. The shards of this run carry the suffix `-haiku-r2` through
`corpus generations --shard-suffix`.

## Coverage

| Operation | Prompt | Responses | Requested words | Realized words | Off length | Mean overlap | Overlap above 0.5 | Fact-sheet identifiers carried |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `generate` | `neutral` | 150 | 48.0 | 46.1 | 0 | 0.000 | 0 | 522 of 775 |
| `generate` | `plain` | 150 | 48.0 | 46.1 | 2 | 0.000 | 0 | 543 of 775 |
| `polish` | `neutral` | 150 | 48.0 | 47.6 | 5 | 0.479 | 77 | 334 of 775 |
| `polish` | `plain` | 150 | 48.0 | 47.1 | 3 | 0.343 | 51 | 342 of 775 |

All 600 responses are complete; none was refused, truncated, or failed.
Overlap is the share of response tokens inside word 8-gram matches with the
original. For `polish` a high overlap is the operation itself. Among the
300 `generate` responses one shares an 8-gram with its original and 30
share a 4-gram. Fifteen carry a heading or a fenced block and 33 run to
more than one paragraph. The record marks contamination by public training
data as unknown for every response.

The extraction found 2,475 units in the 600 controlled documents: 1,144
sentences, 673 paragraphs, and 658 fragments. It found 140 findings in
23,823 prose words (5.9 per 1,000 words) in 104 documents. Fifty-four are
reading grade, 45 passive-candidate density, 18 long sentences, 14 noun
stacks, five parenthetical load, three sentence openers, and one n-gram
density.

## Paired tables

The [paired tables](paired.json) pair each original unit with its response
document. All 150 tasks pair in every arm, across 21 provenance components.
Rules whose class admits the comment role enter; 29 of 40 such rules fired
in no response and no original of any arm. Rows with any finding:

| Rule | Arm | Originals with finding | Responses with finding | Paired change (95% interval) |
| --- | --- | --- | --- | --- |
| `readability.grade-metric` | `generate`/`neutral` | 27 | 21 | -4.0 points (-17.5 to +7.1) |
| `readability.grade-metric` | `generate`/`plain` | 27 | 22 | -3.3 points (-18.5 to +8.9) |
| `readability.grade-metric` | `polish`/`neutral` | 27 | 6 | -14.0 points (-27.8 to -1.9) |
| `readability.grade-metric` | `polish`/`plain` | 27 | 3 | -16.0 points (-34.7 to -1.8) |
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
| `repetition.ngram-density` | `polish`/`neutral` | 11 | 1 | -6.7 points (-13.6 to -0.8) |
| `repetition.ngram-density` | `polish`/`plain` | 11 | 0 | -7.3 points (-15.5 to -0.8) |
| `repetition.paragraph-overlap` | `generate`/`neutral` | 10 | 0 | -6.7 points (-15.0 to +0.0) |
| `repetition.paragraph-overlap` | `generate`/`plain` | 10 | 0 | -6.7 points (-15.0 to +0.0) |
| `repetition.paragraph-overlap` | `polish`/`neutral` | 10 | 0 | -6.7 points (-15.0 to +0.0) |
| `repetition.paragraph-overlap` | `polish`/`plain` | 10 | 0 | -6.7 points (-15.0 to +0.0) |
| `repetition.sentence-openers` | `generate`/`neutral` | 0 | 1 | +0.7 points (+0.0 to +1.9) |
| `repetition.sentence-openers` | `generate`/`plain` | 0 | 2 | +1.3 points (+0.0 to +3.8) |
| `syntax.long-sentence` | `generate`/`neutral` | 24 | 7 | -11.3 points (-29.9 to +1.1) |
| `syntax.long-sentence` | `generate`/`plain` | 24 | 0 | -16.0 points (-36.2 to -1.0) |
| `syntax.long-sentence` | `polish`/`neutral` | 24 | 6 | -12.0 points (-29.3 to +0.9) |
| `syntax.long-sentence` | `polish`/`plain` | 24 | 4 | -13.3 points (-29.8 to -1.0) |
| `syntax.noun-stack` | `generate`/`neutral` | 0 | 9 | +6.0 points (+2.2 to +9.7) |
| `syntax.noun-stack` | `generate`/`plain` | 0 | 5 | +3.3 points (+0.0 to +8.5) |
| `syntax.parenthetical-load` | `generate`/`neutral` | 1 | 1 | +0.0 points (-2.3 to +1.7) |
| `syntax.parenthetical-load` | `generate`/`plain` | 1 | 1 | +0.0 points (-2.3 to +1.7) |
| `syntax.parenthetical-load` | `polish`/`neutral` | 1 | 0 | -0.7 points (-2.5 to +0.0) |
| `syntax.parenthetical-load` | `polish`/`plain` | 1 | 3 | +1.3 points (-1.7 to +5.0) |
| `syntax.passive-candidate-density` | `generate`/`neutral` | 12 | 8 | -2.7 points (-6.5 to +0.0) |
| `syntax.passive-candidate-density` | `generate`/`plain` | 12 | 9 | -2.0 points (-4.7 to +0.0) |
| `syntax.passive-candidate-density` | `polish`/`neutral` | 12 | 10 | -1.3 points (-5.2 to +3.1) |
| `syntax.passive-candidate-density` | `polish`/`plain` | 12 | 16 | +2.7 points (-3.7 to +8.6) |

Thirteen intervals exclude zero. Twelve are negative, and they repeat what
the earlier runs showed. An original paragraph earns a repetition finding
when another part of its document repeats it. A one-paragraph response has
no other part. Those rows measure the context of the original, not the
model. The `polish` reading-grade rows and the `plain` long-sentence rows
are the operation writing a plainer or a shorter sentence.

One interval is positive. It is also the first in this project whose point
estimate reaches the minimum useful difference of three points. The rule is
`syntax.noun-stack` under `generate`/`neutral`, at +6.0 points, from +2.2 to
+9.7. Nine responses carry a noun stack against none of their originals.
The same rule under `generate`/`plain` reaches +3.3 points with an interval
that touches zero. Longer fact sheets give this model more identifiers to
string into one noun phrase than the 22-word tasks of the earlier runs did.
This is one arm of one run on 21 components. It is a candidate to write a
hypothesis template for, not a confirmed construction.

The tables also hold a `difference_from_h0` field. It compares a
one-paragraph response with whole H0 documents of the comment role, so
document length dominates it. It stays because the protocol names it. The
paragraph-level comparison below is the fair one.

## Paragraph-level comparison

The [run tables](tables-paragraph-haiku-r2.json) count the controlled cohort
from this run's shards only: 482 documents with a paragraph unit, 673
paragraphs, 7 provenance components, 16,224 prose words. Controlled
paragraphs of all arms against all historical paragraphs:

| Rule | Historical | Controlled |
| --- | --- | --- |
| `syntax.long-sentence` | 2,723 of 69,579 (3.9%) | 13 of 673 (1.9%) |
| `syntax.passive-candidate-density` | 1,824 of 69,579 (2.6%) | 41 of 673 (6.1%) |
| `readability.grade-metric` | 1,513 of 69,579 (2.2%) | 35 of 673 (5.2%) |
| `repetition.exact-sentence` | 928 of 69,579 (1.3%) | 0 of 673 (0.0%) |
| `syntax.parenthetical-load` | 706 of 69,579 (1.0%) | 3 of 673 (0.4%) |
| `repetition.paragraph-overlap` | 534 of 69,579 (0.8%) | 0 of 673 (0.0%) |
| `repetition.ngram-density` | 316 of 69,579 (0.5%) | 1 of 673 (0.1%) |
| `repetition.sentence-openers` | 296 of 69,579 (0.4%) | 0 of 673 (0.0%) |
| `readability.long-paragraph` | 246 of 69,579 (0.4%) | 0 of 673 (0.0%) |
| `filler.wordy-phrase` | 218 of 69,579 (0.3%) | 0 of 673 (0.0%) |
| `repetition.near-sentence` | 204 of 69,579 (0.3%) | 0 of 673 (0.0%) |
| `repetition.paragraph-openers` | 154 of 69,579 (0.2%) | 0 of 673 (0.0%) |
| `syntax.noun-stack` | 85 of 69,579 (0.1%) | 11 of 673 (1.6%) |
| `filler.weak-intensifiers` | 31 of 69,579 (0.0%) | 0 of 673 (0.0%) |
| `repetition.syntax-template` | 15 of 69,579 (0.0%) | 0 of 673 (0.0%) |
| `syntax.nominalization-chain` | 3 of 69,579 (0.0%) | 0 of 673 (0.0%) |
| `format.em-dash-density` | 2 of 69,579 (0.0%) | 0 of 673 (0.0%) |
| `filler.announced-importance` | 1 of 69,579 (0.0%) | 0 of 673 (0.0%) |

Two differences reach the minimum useful difference of three points, and
both are positive. `syntax.passive-candidate-density` sits 3.5 points above
the historical rate. `readability.grade-metric` sits 3.0 points above it.
`syntax.noun-stack` is 1.5 points above a historical rate of 0.1%. That is
a twelvefold rate on eleven paragraphs. The largest negative difference is
`syntax.long-sentence`, two points below the historical rate.
These are 673 paragraphs in 7 components, well under the cluster minimum of
20 components, and the [screening record](../../../methods/screening/2026-09-12-development-2.json)
is what decides a candidate. The whole controlled cohort, eight runs
together, has 4,361 paragraphs in 4,121 documents across 146 components;
the [six-cohort tables](tables-paragraph.json) count it that way.

## Limits

- One family, two models. The comparison against the OpenAI family is the
  [luna round-two run](../2026-09-12-round2-luna/README.md), which answered
  the same 150 tasks.
- Seven components carry this run's paragraphs. The protocol's cluster
  minimum is 20 per arm, so no card of this run alone can leave
  `inconclusive`.
- The agent harness ran the search agent type with an instruction-free
  system prompt; the agents used no tools, but the harness cannot show that
  the model saw nothing else.
- The requested lengths are longer than every earlier run's, so a
  comparison with the earlier runs mixes the model with the task length.
- The overlap flag is the only contamination control. One `generate`
  response shares an 8-gram with its original.

## What is in this directory

- `tasks.json`, `requests.json`, `responses.json`, `records.json`: the
  saved records, with every input and output text and its hash.
- `shards/`: the controlled cohort's 7 shard manifests of this run.
- `dataset-plan.json`, `tables.json`, `tables-paragraph.json`,
  `tables-sentence.json`, `paired.json`: the six-cohort plan and tables
  after this run joined the controlled cohort.
- `tables-paragraph-haiku-r2.json`: the paragraph tables with the
  controlled cohort restricted to this run's shards.
- `digests.json`: SHA-256 of every record, shard, pinned copy, finding
  artifact, and output of the measurement, which are reproducible from the
  records and the checkouts.
