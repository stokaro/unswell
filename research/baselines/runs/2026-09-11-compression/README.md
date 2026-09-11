# Compression pilot, 2026-09-11

The first run of the compression part of experiment E3 on the corpus of the
[baseline pilot](../2026-09-11-pilot/README.md), produced by
`scripts/compression-experiment.sh` with its defaults. `digests.json` names
the tool commit and the digest and size of every artifact, including the
candidate artifact, the reference bank, and the predictions that stay
outside the repository. `union.json`, `union-plan.json`, and
`selection.json` reproduce them from the acquisition work directory.
`costs.json` records the wall time and peak resident set of every stage.

## Corpus and bank

The union is the one of the baseline pilot, so the corpus is the same: 253
sources, 3,398 paragraphs, and the same partitions. `selection.json` seeds
the reference bank from one training group, the group whose smaller cohort
holds the most paragraphs. Seeds follow unit ID order up to 8 KiB of text
per reference, and the mixed cohort alternates the two seed lists. The bank
compresses at zlib level 6. Its prose stays outside the repository; the
record keeps its digest and the selected unit IDs.

| Reference | Seeds | Bytes | Sources |
| --- | --- | --- | --- |
| Historical | 91 paragraphs | 7,040 | 4 |
| Contemporary | 73 paragraphs | 5,244 | 4 |
| Mixed | 116 paragraphs: 59 contemporary, 57 historical | 8,190 | 4 |

The bank reserves that group. Its 164 training paragraphs leave every fit,
so each arm trains on the same 1,118 rows: 506 `contemporary_snapshot`
against 612 `historical_snapshot`. The calibration and scored partitions are
unchanged.

## Arms

Every arm is a logistic fit with isotonic calibration, stopped at a gradient
norm of 1e-6, with the threshold frozen at 0.5 in both plans before any
partition was scored.

| Arm | Columns | Training loss | Calibration pools |
| --- | --- | --- | --- |
| Structural features (`prepared`) | 14 features | 0.6847 | 4 |
| N-grams (`lexical`) | 128 learned terms | 0.6724 | 2 |
| Reference columns (`compression`) | 6: incremental bytes and reference gain against each cohort | 0.6878 | 6 |
| Joint (`prepared-compression`) | 20: the 14 features and the 6 columns | 0.6841 | 3 |

The entropy of the training prior is 0.6880. Every loss sits at that value,
so no arm learned more than the intercept. The reference column with the
most weight, `compression.reference-gain/historical`, has -0.022 on
standardized inputs. In the joint arm those fourteen features carry more
than any reference column.

## Result

The confirmation partition has 17 provenance components. The intervals
resample them 10,000 times with the fixed seed.

| Arm | Covered | Recall | False-positive rate | Brier | Constant Brier |
| --- | --- | --- | --- | --- | --- |
| Structural features | 1,148 (100%) | 1 [1, 1] | 1 [1, 1] | 0.2848 [0.2390, 0.3436] | 0.2505 |
| N-grams | 1,144 (99.7%) | 1 [1, 1] | 1 [1, 1] | 0.2864 [0.2409, 0.3454] | 0.2505 |
| Reference columns | 1,121 (97.7%) | 1 [1, 1] | 1 [1, 1] | 0.2809 [0.2412, 0.3337] | 0.2514 |
| Joint | 1,134 (98.8%) | 1 [1, 1] | 1 [1, 1] | 0.2858 [0.2400, 0.3439] | 0.2503 |

No arm separates the cohorts. Every covered unit is called contemporary,
because every isotonic map sits near 0.66, the positive rate of the
calibration partition, and that is above the threshold. Before calibration,
the reference columns give the two cohorts a logistic response of 0.452 and
0.451. The reference arm abstains on 27 units and the joint arm on 14, where
the calibration score falls outside the fitted range.

| Comparison | Candidates | Brier difference | Recall difference |
| --- | --- | --- | --- |
| Reference columns against structural features | 1,148 | -0.0039 [-0.0110, +0.0035] | 0 [0, 0] |
| Joint against structural features | 1,148 | +0.0010 [-0.0001, +0.0019] | 0 [0, 0] |
| Joint against n-grams | 1,148 | -0.0006 [-0.0033, +0.0016] | 0 [0, 0] |

The reference columns add nothing over the structural features or the
n-grams, and nothing over the phrase list of the baseline pilot, which the
structural features already matched. Every Brier interval covers zero. This
is the negative result the issue allows.

## Cost

`costs.json` records every stage as one fresh process on a macOS arm64
workstation. Each wall time includes the cold start of the tool.

| Stage | Wall time | Peak resident set |
| --- | --- | --- |
| Plan, extract, and verify the corpus | 7.5 s | 299 MB |
| Build the bank | 3.8 s | 326 MB |
| Fit the structural arm | 8.2 s | 367 MB |
| Fit the n-gram arm | 4.5 s | 368 MB |
| Fit the reference arm | 4.8 s | 405 MB |
| Fit the joint arm | 13.8 s | 402 MB |
| Predict one partition of a reference arm | 4.8 s | 391 MB |
| Whole run | 122.3 s | 471 MB |

A reference measurement compresses the reference with and without the
target, twice per cohort. That is cheaper than the fourteen structural
features, and the joint arm pays for both. The bank is 0.8 MB, the fitted
models are 0.7 MB to 0.8 MB each, and the record is 5.3 MB.

## What this run does not show

One bank, one seed group, one byte budget, and one compression level ran.
A reference of another size, another group, or another level is untested,
and so is a bank of comment sources, because the corpus holds documentation
roles only. The reservation removes one group from training, so every arm
here saw 164 rows fewer than the baseline pilot. Cohort membership is a
snapshot date; a separating feature would carry topics, repositories, and
formats too. None appeared. Nothing here is a judgment of any text, and no
number is an authorship claim.
