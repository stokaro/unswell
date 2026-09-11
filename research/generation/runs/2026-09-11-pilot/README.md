# Stage C pilot, September 11, 2026

This directory is the record of the development pilot of stage C of the
[pattern protocol](../../../methods/llm-patterns-v1.md). It ran under
amendment 1: one generator family, Anthropic Claude, through session agents
of the flat-rate subscription, with no paid call. Every number is a rule
outcome under one policy. It is not a false-positive rate, not recall, not
a share of LLM text, and not a decision. The pilot checks the tooling and
sets the sample size of the confirmatory run. It confirms no hypothesis.

## What ran

Tasks came from the historical cohort of the [period run](../../../acquisition/runs/2026-09-11-periods/README.md):
comment paragraphs of at least twelve words in the training and development
partitions, each followed by a declaration. The seeded draw is stratified by
ecosystem:

| Ecosystem | Eligible | Selected |
| --- | --- | --- |
| c | 43 | 2 |
| cpp | 11 | 1 |
| csharp | 392 | 9 |
| go | 774 | 17 |
| java | 12 | 1 |
| javascript | 906 | 19 |
| python | 100 | 3 |
| rust | 364 | 8 |

Each task met two operations, `generate` from its fact sheet and `polish`
of its original text, under the two frozen prompt conditions, `neutral`
and `plain`: 240 requests. Every request went to one
instruction-free agent of the harness, with the request text and one line
telling the agent to use no tools. The agent transcripts name the model
`claude-opus-5`; each agent made exactly one call, the structured
answer. Decoding parameters, remote request identifiers, and cost are
`unavailable`, as the amendment states.

## Coverage

| Operation | Prompt | Responses | Requested words | Realized words | Off length | Mean overlap | Overlap above 0.5 | Fact-sheet identifiers carried |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `generate` | `neutral` | 60 | 22.3 | 22.9 | 1 | 0.041 | 2 | 186 of 220 |
| `generate` | `plain` | 60 | 22.3 | 22.3 | 4 | 0.052 | 3 | 165 of 220 |
| `polish` | `neutral` | 60 | 22.3 | 22.8 | 0 | 0.621 | 41 | 72 of 220 |
| `polish` | `plain` | 60 | 22.3 | 22.6 | 0 | 0.646 | 42 | 72 of 220 |

All 240 responses are complete; none was refused, truncated,
or failed. Overlap is the share of response tokens inside word 8-gram
matches with the original. For `polish` a high overlap is the operation
itself. For `generate` it is a contamination signal: five responses repeat
the original wording from the fact sheet alone, and two of them reproduce a
testify doc comment word for word from its signature. Those responses stay
in the primary analysis and are listed for the sensitivity analysis, as the
protocol requires; the record marks contamination by public training data
as unknown for every response.

Most `generate` responses came back as structured Markdown: a heading, a
fenced code block with the signature, then prose, sometimes a bullet list.
The originals are single comment paragraphs. The extraction therefore found
661 units in the 240 controlled documents, 145 of them paragraphs,
and 6 findings in 4,523 prose words (1.33 per 1,000 words).

## Paired tables

The [paired tables](paired.json) pair each original unit with its response
document. Of the 60 tasks, 47 pair in every arm; 13 originals sit in
documents whose policy run failed, so they are coverage, not pairs. The
arms span 25 provenance components. Rules whose class admits the
comment role enter; 38 of 40 such rules fired in no response of any arm.
Rows with any finding, original or response:

| Rule | Arm | Originals with finding | Responses with finding | Paired change (95% interval) |
| --- | --- | --- | --- | --- |
| `repetition.exact-sentence` | `generate`/`neutral` | 2 | 0 | -4.3 points (-11.8 to +0.0) |
| `repetition.exact-sentence` | `generate`/`plain` | 2 | 0 | -4.3 points (-11.8 to +0.0) |
| `repetition.exact-sentence` | `polish`/`neutral` | 2 | 0 | -4.3 points (-11.8 to +0.0) |
| `repetition.exact-sentence` | `polish`/`plain` | 2 | 0 | -4.3 points (-11.8 to +0.0) |
| `repetition.near-sentence` | `generate`/`neutral` | 1 | 0 | -2.1 points (-8.0 to +0.0) |
| `repetition.near-sentence` | `generate`/`plain` | 1 | 0 | -2.1 points (-8.0 to +0.0) |
| `repetition.near-sentence` | `polish`/`neutral` | 1 | 0 | -2.1 points (-8.0 to +0.0) |
| `repetition.near-sentence` | `polish`/`plain` | 1 | 0 | -2.1 points (-8.0 to +0.0) |
| `repetition.paragraph-overlap` | `generate`/`neutral` | 1 | 0 | -2.1 points (-4.6 to +0.0) |
| `repetition.paragraph-overlap` | `generate`/`plain` | 1 | 0 | -2.1 points (-4.6 to +0.0) |
| `repetition.paragraph-overlap` | `polish`/`neutral` | 1 | 0 | -2.1 points (-4.6 to +0.0) |
| `repetition.paragraph-overlap` | `polish`/`plain` | 1 | 0 | -2.1 points (-4.6 to +0.0) |
| `syntax.parenthetical-load` | `generate`/`neutral` | 1 | 0 | -2.1 points (-8.6 to +0.0) |
| `syntax.parenthetical-load` | `generate`/`plain` | 1 | 0 | -2.1 points (-8.6 to +0.0) |
| `syntax.parenthetical-load` | `polish`/`neutral` | 1 | 1 | +0.0 points (+0.0 to +0.0) |
| `syntax.parenthetical-load` | `polish`/`plain` | 1 | 1 | +0.0 points (+0.0 to +0.0) |
| `syntax.passive-candidate-density` | `generate`/`neutral` | 0 | 1 | +2.1 points (+0.0 to +7.4) |

These are small counts over 47 pairs. No response carried a finding that
its original also carried. The one response finding of the
`generate`/`neutral` arm is a single passive candidate. The tables also hold
a `difference_from_h0` field. It compares a one-paragraph response with
whole H0 documents of the comment role, so document length dominates it.
It stays because the protocol names it. The paragraph-level comparison
below is the fair one.

## Paragraph-level comparison

From the [paragraph tables](tables-paragraph.json) of the six-cohort
measurement, controlled paragraphs of all arms against all historical
paragraphs:

| Rule | Historical | Controlled |
| --- | --- | --- |
| `syntax.passive-candidate-density` | 586 of 46,200 (1.3%) | 3 of 145 (2.1%) |
| `syntax.parenthetical-load` | 455 of 46,200 (1.0%) | 2 of 145 (1.4%) |
| `syntax.long-sentence` | 1,415 of 46,200 (3.1%) | 1 of 145 (0.7%) |
| `readability.grade-metric` | 902 of 46,200 (2.0%) | 0 of 145 (0.0%) |
| `syntax.noun-stack` | 43 of 46,200 (0.1%) | 0 of 145 (0.0%) |
| `format.em-dash-density` | 0 of 46,200 (0.0%) | 0 of 145 (0.0%) |
| `filler.weak-intensifiers` | 19 of 46,200 (0.0%) | 0 of 145 (0.0%) |

With 145 controlled paragraphs, a rule at 1% prevalence fires about once.
Nothing here separates a cohort. The pilot sizes the confirmatory run. At
the protocol's minimum useful difference of three points and the historical
prevalences above, an arm needs hundreds of pairs per rule of interest, not
47. The sample-size record of stage C follows from that.

## Limits

- One family. Every other family is `untested`; protocol version 2 needs
  three.
- The agent harness ran the search agent type with an instruction-free
  system prompt; the agents used no tools, but the harness cannot show that
  the model saw nothing else.
- Requested lengths are short (about 22 words), because the originals are
  comment paragraphs; the structured Markdown of many responses splits them
  into fragments, and the paragraph analysis sees only part of the text.
- The generator reproduced public text in a few cases; the overlap flag is
  the only contamination control.

## What is in this directory

- `tasks.json`, `requests.json`, `responses.json`, `records.json`: the
  saved records, with every input and output text and its hash.
- `shards/`: the controlled cohort's 19 shard manifests.
- `dataset-plan.json`, `tables.json`, `tables-paragraph.json`,
  `tables-sentence.json`, `paired.json`: the six-cohort plan and tables.
- `digests.json`: SHA-256 of every shard, pinned copy, finding artifact, and
  output of the measurement, which are reproducible from the records and the
  checkouts.
