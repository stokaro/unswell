# Baseline pilot, 2026-09-11

The first run of experiment E3 on the historical and contemporary shards of
`unswell-llm-patterns-v1`, produced by `scripts/baseline-experiment.sh` with
its defaults. `digests.json` names the tool commit and the digest and size of
every artifact, including the candidate artifact and the predictions that
stay outside the repository. `union.json` and `union-plan.json` reproduce
them from the acquisition work directory. `costs.json` records the wall time
and peak resident set of every stage.

## Corpus

`union.json` keeps the documentation, readme, and release-note sources of the
historical and contemporary shards, at most four per checkout in ID order. It
drops files above 100 KiB and the fourteen sources the pattern measurement
could not analyze, listed in `excluded-sources.json`. Only paragraphs are
extracted. Comment sources stay out, because a rule arm binds activations to
original blocks of the paragraph kind, and a comment block is not one. The
cap is four because the arm of all rules records every block of every
source, and a larger corpus exceeds the bound of that collection.

| Count | Value |
| --- | --- |
| Sources | 253: 124 historical, 129 contemporary, from 78 checkouts of 43 repositories |
| Roles | 165 documentation, 66 readme, 22 release notes |
| Paragraph units | 3,398: 1,631 historical, 1,767 contemporary; the median unit has 11 words |
| Groups | 43 connected components, pinned to the dataset's partitions |
| Training | 1,282 rows in 19 groups: 579 `contemporary_snapshot`, 703 `historical_snapshot` |
| Calibration | 950 rows in 6 groups: 635 contemporary, 315 historical |
| Development | 18 candidates in one group, all historical |
| Confirmation | 1,148 candidates in 17 groups: 553 contemporary, 595 historical |

One contemporary source, a generated API reference, holds 416 of the units.

## Fit

Every arm is a fit on the training partition with isotonic calibration on
the calibration partition. The threshold was frozen at 0.5 in both plans
before any partition was scored.

| Arm | Columns | Training loss | Calibration pools |
| --- | --- | --- | --- |
| Phrase list (`rules-phrases`) | 15 activations | 0.6885 | 2 |
| Rule activations (`rules-all`) | 39 activations | 0.6867 | 3 |
| N-grams (`lexical`) | 128 learned terms | 0.6756 | 3 |
| Structural features (`prepared`) | 14 features | 0.6851 | 5 |
| Forest (`forest`) | 14 features, 2,284 nodes | none | 7 |

The training prevalence is 0.452, and the entropy of that prior is 0.6885.
Every logistic loss sits at that value, so the fits learned the intercept
and little else. The zero policy filled 5,213 of the 19,230 activation cells
in the phrase arm's training rows, and 21,867 of the 49,998 cells in the rule
arm's. The reasons are structural. `format.list-fragmentation` has no
paragraph to judge on any row. `syntax.rhetorical-question-density`,
`scaffold.chat-preamble`, `scaffold.follow-up-offer`, and
`repetition.heading-echo` find no eligible window or pair on about 1,250
rows each. `readability.grade-metric` needs more words than most paragraphs
have, on 1,214 rows.

## Result

The confirmation partition has 17 provenance components. The intervals
resample them 10,000 times with the fixed seed.

| Arm | Covered | Recall | False-positive rate | Brier | Constant Brier |
| --- | --- | --- | --- | --- | --- |
| Phrase list | 1,148 (100%) | 1 [1, 1] | 1 [1, 1] | 0.2845 [0.2389, 0.3431] | 0.2506 |
| Rule activations | 1,148 (100%) | 1 [1, 1] | 1 [1, 1] | 0.2842 [0.2387, 0.3426] | 0.2506 |
| N-grams | 1,143 (99.6%) | 1 [1, 1] | 1 [1, 1] | 0.2867 [0.2407, 0.3459] | 0.2505 |
| Structural features | 1,148 (100%) | 1 [1, 1] | 1 [1, 1] | 0.2848 [0.2390, 0.3436] | 0.2506 |
| Forest | 1,144 (99.7%) | 0.9927 [0.9821, 1] | 0.9983 [0.9954, 1] | 0.2927 [0.2458, 0.3551] | 0.2505 |

No arm separates the cohorts. Every arm calls nearly every unit
contemporary. The isotonic map of each arm is flat at 0.668, the positive
rate of the calibration partition, and that is above the threshold. Before
calibration, the logistic response of the two cohorts differs by less than
0.01, and for the n-gram, structural, and forest arms the historical units
score slightly higher for the contemporary class. The Brier score is worse
than the training constant because the calibration partition holds 67%
contemporary units, against 45% in training and 48% in confirmation. The
development partition is 18 historical units in one group: every arm flags
them all, and no interval exists there.

The strata by role and by length say the same. In every role (documentation,
readme, release note) and every length band, the logistic arms cover every
unit and flag every unit; the forest misses four units and abstains on four.
`evaluation-final_test.json` of each arm carries those tables.

The paired comparisons resample the same 17 groups. A positive Brier
difference means the candidate is worse.

| Comparison | Candidates | Brier difference | Recall difference |
| --- | --- | --- | --- |
| Rule activations against the phrase list | 1,148 | -0.0003 [-0.0007, 0.0000] | 0 [0, 0] |
| N-grams against structural features | 1,148 | +0.0018 [+0.0001, +0.0038] | 0 [0, 0] |
| Forest against structural features | 1,148 | +0.0079 [+0.0033, +0.0139] | -0.0073 [-0.0179, 0] |

The 24 rules outside the phrase families add nothing over the phrase list:
three ten-thousandths of Brier score. The n-gram arm and the forest are
slightly worse than the fourteen structural features. This is the negative
result the issue allows. No feature family carries a separation, because at
this corpus size there is none to carry.

## Cost

`costs.json` records every stage as one fresh process, so each wall time
includes the cold start of the tool. The host is a macOS arm64 workstation,
not the container of the performance record.

| Stage | Wall time | Peak resident set |
| --- | --- | --- |
| Plan, extract, and verify the corpus | 7.3 s | 309 MB |
| Fit the phrase arm | 8.8 s | 572 MB |
| Fit the rule arm | 9.0 s | 809 MB |
| Fit the n-gram arm | 4.4 s | 395 MB |
| Fit the structural arm | 8.0 s | 345 MB |
| Fit the forest | 8.2 s | 349 MB |
| Predict one partition | 4.1 s to 8.9 s | up to 466 MB |
| Evaluate one partition | 0.2 s to 0.3 s | up to 100 MB |
| Compare one pair | 0.3 s to 0.4 s | up to 82 MB |
| Whole run | 124.5 s | 809 MB |

The rule arm is the expensive one: its collection holds 17,234 blocks with
39 activations each. The candidate artifact is 9.4 MB, the fitted
models are 0.7 MB to 1.1 MB each, and a prediction file is 1.3 MB to 1.8 MB.

## What this run does not show

The corpus is four sources per checkout of three documentation roles, 3,398
paragraphs. A separation that needs more text, or comment sources, is
untested. Cohort membership is a snapshot date, so a feature that separated
the cohorts would also carry the topics, repositories, and formats of each
period. None appeared here. The protocol's phrase list adds the Reinhart and
Kobak word lists to the pinned phrase catalog; those lists are not in the
repository, so the phrase arm covers the catalog's fifteen rules only. A rule
arm and a prepared arm measure different contexts, the source document and
the prepared piece, so they are never compared with each other. The logistic
arms stop at a gradient norm of 1e-6 rather than the default 1e-8, because
the flat optimum defeats the line search in rounding. Nothing here is a
judgment of any text, and no number is an authorship claim.
