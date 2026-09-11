# Baseline comparisons

This directory records experiment E3 of the pattern protocol of
[ADR 0036](../../docs/adr/0036-llm-pattern-evidence.md). The question is
whether small offline baselines separate the cohorts of the protocol better
than the rules and the phrase list, and which feature families carry that
separation. [#176](https://github.com/stokaro/unswell/issues/176) is the
mirrored issue; its original, #50, stays on hold with the human-labeled
corpus. The target is cohort membership: a historical snapshot against a
contemporary one. It is never a label of quality and never an origin claim.

## How a run works

`scripts/baseline-experiment.sh` orders the research commands and records
digests and stage costs. `corpus dataset union` joins the historical and
contemporary shards of every repository. `plan`, `extract`, and `verify`
build the corpus from the acquisition work directory. Every arm runs `corpus
train --labels cohort`, then `predict` on the selection and confirmation
partitions, then `corpus evaluate --labels cohort` against the same labels.
`compare` pairs the arms that share an input context.

```sh
bash scripts/baseline-experiment.sh --record research/baselines/runs/<run>
```

The arms are the ones the issue names, on identical inputs.

| Arm | Directory | Features | Estimator |
| --- | --- | --- | --- |
| Phrase list | `rules-phrases` | The activations of the fifteen rules of the phrase families `filler`, `hype`, and `scaffold` | Logistic regression |
| Rule activations | `rules-all` | The activations of all 39 classified rules except the phrase prohibition | Logistic regression |
| N-grams | `lexical` | Word and character n-gram counts, with a vocabulary learned on the training partition | Logistic regression |
| Structural features | `prepared` | Fourteen structure, lexical, part-of-speech, and readability features | Logistic regression |
| Forest | `forest` | The same fourteen features | Random forest |

Every arm fits on the training partition and calibrates on the calibration
partition. Its threshold is frozen at 0.5 before either scored partition is
read. A rule that cannot fire on a unit counts as zero there, and the
artifact records how often. A comparison stays inside one input context. The
rule arms compare with each other, and the prepared arms compare with each
other, as the evaluation contract requires.

`scripts/compression-experiment.sh` runs the compression part of the same
experiment for [#178](https://github.com/stokaro/unswell/issues/178) on the
same union. `corpus reference-bank --labels cohort` seeds a historical, a
contemporary, and a mixed reference from one training group. `train
--compression-bank` measures every paragraph against each reference, alone
or joined with the structural features, and `train --reserve-bank` keeps
the compared baselines on the same rows, without the group the bank
reserved. The bank holds source prose, so a record keeps its digest and
selection, not its bytes.

```sh
bash scripts/compression-experiment.sh --record research/baselines/runs/<run>
```

`scripts/llmdet-experiment.sh` runs the LLMDet part for
[#177](https://github.com/stokaro/unswell/issues/177) on the same union.
`train --llmdet-pack` measures every paragraph against the dictionary
tables of a local pack, a proxy perplexity and a context coverage per
model, alone or joined with the structural features. The
[LLMDet experiment](../llmdet/README.md) describes the tables, the
tokenizer port, and the pack; a record keeps the pack manifest with its
digests, not the tables.

```sh
bash scripts/llmdet-experiment.sh --record research/baselines/runs/<run>
```

## What a run does not establish

A cohort difference is a difference between two sets of snapshots. It carries
the composition of repositories, topics, and roles, and a run does not
separate those from the writing. A feature that separates the cohorts marks a
period, not a text that needs revision; that question stays with #50 and
#22. The rule arms bind activations to original blocks of the paragraph
kind, so comment sources stay out of every run. No arm judges quality, and no
number here is an authorship verdict or a share of text written by a tool.

## Runs

| Run | What it covers | Record |
| --- | --- | --- |
| 2026-09-11 pilot | First fit of the five arms on the documentation, readme, and release-note paragraphs of both cohorts, four sources per checkout, threshold frozen at 0.5 | [runs/2026-09-11-pilot](runs/2026-09-11-pilot/README.md) |
| 2026-09-11 compression | The reference columns of one bank, alone and joined with the structural features, against the structural features and the n-grams on the pilot's corpus | [runs/2026-09-11-compression](runs/2026-09-11-compression/README.md) |
| 2026-09-11 LLMDet | The proxy perplexity and context coverage of six ported models, alone and joined with the structural features, against the structural features and the n-grams on the pilot's corpus | [runs/2026-09-11-llmdet](runs/2026-09-11-llmdet/README.md) |
