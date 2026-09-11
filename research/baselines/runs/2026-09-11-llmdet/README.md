# LLMDet pilot, 2026-09-11

The first run of the LLMDet part of experiment E3 on the corpus of the
[baseline pilot](../2026-09-11-pilot/README.md), produced by
`scripts/llmdet-experiment.sh` with its defaults. `digests.json` names the
tool commit and the digest and size of every artifact, including the
candidate artifact and the predictions that stay outside the repository.
`pack.json` names the six dictionary tables and the tokenizer files of the
run with their digests; the tables stay outside the repository too, and the
[LLMDet experiment](../../../llmdet/README.md) says how to rebuild them.
`union.json` and `union-plan.json` reproduce the corpus from the acquisition
work directory. `costs.json` records the wall time and peak resident set of
every stage.

## Corpus and models

The union is the one of the baseline pilot, so the corpus is the same: 253
sources, 3,398 paragraphs, and the same partitions. Six of the eleven
published dictionaries ran, the ones whose tokenizer the port covers:
`gpt2`, `gpt2_large`, `neo`, `opt`, `opt_3b`, and `bart`. Each gives two
columns, a proxy perplexity and a context coverage. UniLM, LLaMA, Vicuna,
T5, and Bloom stay untested.

A paragraph enters a fit only when every column has a value, under the
`exclude` policy. Three abstentions of the reference apply. A paragraph of
fewer than four tokens has no position to examine. A paragraph whose
contexts never appear in a model's dictionary has no matching context. A
paragraph whose matched positions all carry a zero likelihood has no
evaluated probability. Every partition of `model.json` records each
abstention by column, and the exclusion counts name the first abstaining
column of a row.

| Partition | Candidates | Fitted or covered | Abstentions |
| --- | --- | --- | --- |
| Training | 1,282 | 964: 428 `contemporary_snapshot`, 536 `historical_snapshot` | 83 rows under four tokens; 1 row without an evaluated probability |
| Calibration | 950 | 833: 575 contemporary, 258 historical | 6 rows under four tokens; 2 rows without an evaluated probability |
| Confirmation | 1,148 | 645 covered (56.2%) | 302 rows under four tokens; 2 rows without an evaluated probability; 2 rows outside the calibration range |

A row also leaves a fit when one model's dictionary holds none of its
contexts. The counts below are rows per model. In the confirmation
partition only the first abstaining column of a row is recorded, so its row
gives the rows that had every token but no context in that model.

| Partition | neo | bart | gpt2 | gpt2_large | opt | opt_3b |
| --- | --- | --- | --- | --- | --- | --- |
| Training | 213 | 76 | 6 | 6 | 5 | 5 |
| Calibration | 103 | 38 | 0 | 0 | 1 | 1 |
| Confirmation, first column | 125 | 72 | 0 | 0 | 0 | 0 |

The reference examines positions from the third token on, so a paragraph of
fewer than four tokens has no score. That is the largest abstention: a fifth
of the corpus is one short line. The `neo` dictionary has the fewest unigram
contexts and abstains most often for a missing context.

## Arms

Every arm is a logistic fit with isotonic calibration, stopped at a gradient
norm of 1e-6, with the threshold frozen at 0.5 in both plans before any
partition was scored. The structural and n-gram arms are the ones of the
baseline pilot, refitted on this run.

| Arm | Columns | Training rows | Training loss | Calibration pools |
| --- | --- | --- | --- | --- |
| Structural features (`prepared`) | 14 features | 1,282 | 0.6851 | 5 |
| N-grams (`lexical`) | 128 learned terms | 1,282 | 0.6756 | 3 |
| Proxy columns (`llmdet`) | 12: perplexity and coverage of six models | 964 | 0.6832 | 1 |
| Joint (`prepared-llmdet`) | 26: the 14 features and the 12 columns | 964 | 0.6781 | 2 |

The entropy of the training prior is 0.6880 on the full rows and
0.6869 on the 964 rows the proxy arms fit. Every loss sits at that
value, so no arm learned more than the intercept.

## Result

The confirmation partition has 17 provenance components. The intervals
resample them 10,000 times with the fixed seed.

| Arm | Covered | Recall | False-positive rate | Brier | Constant Brier |
| --- | --- | --- | --- | --- | --- |
| Structural features | 1,148 (100%) | 1 [1, 1] | 1 [1, 1] | 0.2848 [0.2390, 0.3436] | 0.2506 |
| N-grams | 1,143 (99.6%) | 1 [1, 1] | 1 [1, 1] | 0.2867 [0.2407, 0.3459] | 0.2505 |
| Proxy columns | 645 (56.2%) [36.7%, 76.7%] | 1 [1, 1] | 1 [1, 1] | 0.3024 [0.2345, 0.3722] | 0.2484 |
| Joint | 647 (56.4%) [36.8%, 76.9%] | 1 [1, 1] | 1 [1, 1] | 0.3020 [0.2346, 0.3714] | 0.2484 |

No arm separates the cohorts. Every covered unit is called contemporary,
because every isotonic map sits at 0.69, the positive rate of the
calibration rows, and that is above the threshold. Before calibration, the
proxy columns give the two cohorts a logistic response of 0.450 and 0.448.
The coverage of the proxy arms is the share of paragraphs with a value in
every column, and its interval over components is wide, because short
paragraphs cluster by repository.

| Comparison | Candidates | Brier difference | Recall difference |
| --- | --- | --- | --- |
| Proxy columns against structural features | 1,148 | +0.0176 [-0.0089, +0.0513] | 0 [0, 0] |
| Joint against structural features | 1,148 | +0.0172 [-0.0089, +0.0505] | 0 [0, 0] |
| Joint against n-grams | 1,148 | +0.0154 [-0.0116, +0.0491] | 0 [0, 0] |

The proxy columns add nothing over the structural features or the n-grams.
Every Brier interval covers zero, and the point estimates lean the wrong
way, because the proxy arms abstain on the units the other arms score. This
is the negative result the issue allows.

## Cost

`costs.json` records every stage as one fresh process on a macOS arm64
workstation. Each wall time includes the cold start of the tool. A proxy
stage hashes and loads six tables of about 600 MB, one after another. That
costs about a minute and holds up to 2.5 GB.

| Stage | Wall time | Peak resident set |
| --- | --- | --- |
| Plan, extract, and verify the corpus | 7.5 s | 290 MB |
| Fit the structural arm | 8.4 s | 362 MB |
| Fit the n-gram arm | 4.6 s | 379 MB |
| Fit the proxy arm | 72.3 s | 2383 MB |
| Fit the joint arm | 76.2 s | 2280 MB |
| Predict one partition of a proxy arm | 76.0 s | 2093 MB |
| Whole run | 491.3 s | 2383 MB |

The six tables take 3.7 GB on disk, and the archive they came from 4.8 GB.
The fitted models are 0.6 MB to 0.8 MB each, and the record is 5.3 MB. A product
that shipped this feature would carry a multiple of the whole scan budget
in memory for one model alone. That cost, and the result above, are the
adoption evidence.

## What this run does not show

Six of eleven models ran, on one corpus of documentation roles. The
classifier stage of the reference, which needs all eleven proxies, did not
run. A paragraph of fewer than four tokens has no score by the reference's
own rule, so short prose is outside what these columns can say. Cohort
membership is a snapshot date; a separating feature would carry topics,
repositories, and formats too. None appeared. Nothing here is a judgment of
any text, and no number is an authorship claim.
