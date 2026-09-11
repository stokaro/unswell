# Period run, September 11, 2026

This directory is the published record of the third stage B run of the
[pattern protocol](../../../methods/llm-patterns-v1.md). It adds three
dated historical periods to the two cohorts of the [second
run](../2026-09-11/README.md): the placebo pseudo-boundaries of December
31, 2012 and 2016 inside H0, and the H1 sensitivity boundary of December
31, 2018. Every cohort holds the same repositories at the newest release
tag committed on or before its boundary. All five share one global plan,
and every shard was measured again under that plan. Every number is a rule
outcome under one policy on one cohort. It is not a false-positive rate,
not recall, and not a decision.

## What ran

```sh
GITHUB_TOKEN=... bash scripts/acquire-corpus.sh --cohort historical-2012 \
  --sources research/acquisition/sources-historical-2012-v1.json
GITHUB_TOKEN=... bash scripts/acquire-corpus.sh --cohort historical-2016 \
  --sources research/acquisition/sources-historical-2016-v1.json
GITHUB_TOKEN=... bash scripts/acquire-corpus.sh --cohort historical-2018 \
  --sources research/acquisition/sources-historical-2018-v1.json
GITHUB_TOKEN=... bash scripts/acquire-corpus.sh --cohort historical
GITHUB_TOKEN=... bash scripts/acquire-corpus.sh --cohort contemporary \
  --sources research/acquisition/sources-contemporary-v1.json
bash scripts/measure-corpus.sh
```

The historical and contemporary cohorts were acquired again from the
existing checkouts, because source IDs now carry the cohort. Two periods
that resolve to the same release therefore keep distinct sources, and the
later one enters its cohort as repeated text. `FasterXML/jackson-databind`
joins the historical cohort for the first time: its notice is a link into
the checkout, which the acquisition now follows. The policy is
[`policy-e1.yaml`](policy-e1.yaml) on profile `strict-v1` with every
catalog rule enabled. The rule classes are
[`rule-classes-v1.json`](../../../methods/rule-classes-v1.json) revision 1.

## Cohorts

| Cohort | Repositories acquired | Documents | Components | Prose words | Findings | Per 1,000 words |
| --- | --- | --- | --- | --- | --- | --- |
| `historical-2012` | 11 (6 corroborated, 5 `vcs_only`) | 477 | 19 | 155,064 | 1,127 | 7.27 |
| `historical-2016` | 36 (29 corroborated, 7 `vcs_only`) | 2,266 | 78 | 538,095 | 3,851 | 7.16 |
| `historical-2018` | 42 (39 corroborated, 3 `vcs_only`) | 2,965 | 105 | 822,358 | 5,253 | 6.39 |
| `historical` | 43 (40 corroborated, 3 `vcs_only`) | 3,725 | 128 | 1,119,894 | 7,920 | 7.07 |
| `contemporary` | 37 (37 corroborated, 0 `vcs_only`) | 3,448 | 119 | 1,145,832 | 6,727 | 5.87 |

The 2012 period reaches eleven repositories, because most of the frame
did not exist or had no release tag before 2013. Five of its snapshots are
`vcs_only`: GitHub release records did not exist then, and only registries
can corroborate a 2012 tag. The dataset holds 40 global components and
12,897 sources over 453 shards, all pinned, verified, and
measured; no shard was skipped.

## Placebo comparisons

Each pair contrasts a period with the one before it, the same way the
contemporary cohort meets H0. The later cohort keeps only the paragraphs
whose text the earlier snapshot of the same repository does not hold. The
difference takes all paragraphs of the earlier cohort as its reference. A
repository without a snapshot in the earlier period stays out of the pair.

| Pair | Later units | Kept as new | Repositories without an earlier snapshot |
| --- | --- | --- | --- |
| 2016 vs 2012 | 26,840 | 5,030 | 25 |
| 2018 vs 2016 | 37,381 | 8,919 | 6 |
| 2020 (H0) vs 2018 | 48,499 | 10,116 | 1 |
| contemporary vs 2020 (H0) | 41,974 | 21,820 | 0 |

Differences in percentage points with the joint cluster interval,
paragraph level:

| Rule | 2016 vs 2012 | 2018 vs 2016 | 2020 (H0) vs 2018 | contemporary vs 2020 (H0) |
| --- | --- | --- | --- | --- |
| `syntax.long-sentence` | -0.2 (-1.1 to +0.8) | -3.8 (-6.2 to -1.9) | -1.6 (-3.6 to -0.1) | -1.6 (-3.0 to -0.3) |
| `syntax.passive-candidate-density` | +0.3 (-1.3 to +2.0) | +0.4 (-0.3 to +1.3) | +0.2 (-0.5 to +0.8) | +0.4 (-0.3 to +1.3) |
| `readability.grade-metric` | +0.2 (-0.9 to +1.3) | -1.7 (-3.1 to -0.3) | -1.0 (-2.1 to -0.0) | -0.6 (-1.6 to +0.3) |
| `syntax.parenthetical-load` | +1.0 (-0.4 to +3.7) | -0.1 (-0.8 to +0.9) | +0.8 (-0.0 to +1.8) | +0.1 (-0.4 to +0.8) |
| `syntax.noun-stack` | +0.1 (-0.1 to +0.4) | -0.0 (-0.1 to +0.1) | -0.0 (-0.1 to +0.1) | +0.1 (+0.0 to +0.2) |
| `format.em-dash-density` | +0.0 (+0.0 to +0.0) | +0.0 (+0.0 to +0.0) | +0.0 (+0.0 to +0.0) | +0.0 (+0.0 to +0.1) |

The same at sentence level:

| Rule | 2016 vs 2012 | 2018 vs 2016 | 2020 (H0) vs 2018 | contemporary vs 2020 (H0) |
| --- | --- | --- | --- | --- |
| `syntax.long-sentence` | -0.3 (-1.0 to +0.6) | -2.9 (-4.8 to -1.4) | -1.3 (-2.7 to -0.0) | -1.2 (-2.4 to -0.2) |
| `syntax.passive-candidate-density` | -0.1 (-1.1 to +0.8) | +0.2 (-0.2 to +0.8) | -0.0 (-0.5 to +0.4) | +0.2 (-0.3 to +0.8) |
| `readability.grade-metric` | -0.1 (-0.8 to +0.6) | -1.5 (-2.6 to -0.6) | -1.0 (-1.8 to -0.4) | -0.6 (-1.3 to +0.1) |
| `syntax.parenthetical-load` | +0.8 (-0.4 to +2.8) | -0.1 (-0.7 to +0.5) | +0.6 (+0.0 to +1.4) | +0.1 (-0.3 to +0.6) |
| `syntax.noun-stack` | +0.1 (-0.1 to +0.3) | -0.0 (-0.1 to +0.1) | -0.0 (-0.1 to +0.0) | +0.1 (-0.0 to +0.1) |
| `format.em-dash-density` | +0.0 (+0.0 to +0.0) | +0.0 (+0.0 to +0.0) | +0.0 (+0.0 to +0.0) | +0.0 (+0.0 to +0.1) |

The post-2022 change sits inside the distribution of the placebo changes.
New paragraphs carry fewer long sentences than the earlier snapshot in
every pair, and the 2018 pair shows the largest such drop. The passive rule
moves up by under half a point in every pair. The parenthetical rule moves
up in three pairs, most in the 2016 pair. Nothing in the contemporary pair
stands apart from the placebo pairs, except the em-dash rule: it fires in 8
new contemporary paragraphs and in no paragraph of any historical period.
The em-dash count is 0.04% of the new paragraphs.

## One count per text across five periods

The unit selection keeps the first occurrence of each paragraph text across
the five cohorts in period order, whatever repository a copy sits in.

| Rule | 2012 | 2016 | 2018 | 2020 (H0) | contemporary |
| --- | --- | --- | --- | --- | --- |
| `syntax.long-sentence` | 98 of 6,682 (1.5%) | 192 of 12,738 (1.5%) | 133 of 13,780 (1.0%) | 302 of 13,358 (2.3%) | 290 of 16,948 (1.7%) |
| `syntax.passive-candidate-density` | 158 of 6,682 (2.4%) | 245 of 12,738 (1.9%) | 209 of 13,780 (1.5%) | 298 of 13,358 (2.2%) | 346 of 16,948 (2.0%) |
| `readability.grade-metric` | 95 of 6,682 (1.4%) | 161 of 12,738 (1.3%) | 119 of 13,780 (0.9%) | 262 of 13,358 (2.0%) | 264 of 16,948 (1.6%) |
| `syntax.parenthetical-load` | 40 of 6,682 (0.6%) | 237 of 12,738 (1.9%) | 138 of 13,780 (1.0%) | 304 of 13,358 (2.3%) | 240 of 16,948 (1.4%) |
| `syntax.noun-stack` | 14 of 6,682 (0.2%) | 30 of 12,738 (0.2%) | 12 of 13,780 (0.1%) | 19 of 13,358 (0.1%) | 36 of 16,948 (0.2%) |
| `format.em-dash-density` | 0 of 6,682 (0.0%) | 0 of 12,738 (0.0%) | 0 of 13,780 (0.0%) | 0 of 13,358 (0.0%) | 8 of 16,948 (0.0%) |
| kept units | 6,696 | 14,821 | 13,793 | 13,417 | 17,065 |

The same at sentence level:

| Rule | 2012 | 2016 | 2018 | 2020 (H0) | contemporary |
| --- | --- | --- | --- | --- | --- |
| `syntax.long-sentence` | 98 of 8,090 (1.2%) | 188 of 16,043 (1.2%) | 134 of 16,610 (0.8%) | 306 of 15,779 (1.9%) | 303 of 20,268 (1.5%) |
| `syntax.passive-candidate-density` | 153 of 8,090 (1.9%) | 226 of 16,043 (1.4%) | 189 of 16,610 (1.1%) | 267 of 15,779 (1.7%) | 301 of 20,268 (1.5%) |
| `readability.grade-metric` | 92 of 8,090 (1.1%) | 119 of 16,043 (0.7%) | 91 of 16,610 (0.5%) | 226 of 15,779 (1.4%) | 226 of 20,268 (1.1%) |
| `syntax.parenthetical-load` | 39 of 8,090 (0.5%) | 235 of 16,043 (1.5%) | 126 of 16,610 (0.8%) | 292 of 15,779 (1.9%) | 227 of 20,268 (1.1%) |
| `syntax.noun-stack` | 14 of 8,090 (0.2%) | 30 of 16,043 (0.2%) | 12 of 16,610 (0.1%) | 17 of 15,779 (0.1%) | 33 of 20,268 (0.2%) |
| `format.em-dash-density` | 0 of 8,090 (0.0%) | 0 of 16,043 (0.0%) | 0 of 16,610 (0.0%) | 0 of 15,779 (0.0%) | 8 of 20,268 (0.0%) |
| kept units | 8,108 | 18,702 | 16,624 | 15,848 | 20,402 |

The contemporary values lie within the range of the historical periods for
every rule. The 2018 period has the lowest values for the long-sentence,
grade, and passive rules, and H0 the highest; the contemporary period is
between them.

## Limits of this run

A temporal change is never a causal LLM effect, and these tables show none
to explain. The H1 baseline (2018) gives the contemporary contrast a
second reference; against it the new contemporary text moves in the same
direction as against H0. The 2012 pair rests on eleven repositories and
2,809 new paragraphs, so its intervals are wide. The contemporary origin is
unknown or mixed: nothing here labels any text as AI-written. No
controlled cohort exists yet; stage C needs a recorded budget before any
generation.

## What is in this directory

- The five acquisition logs and their records.
- The dataset manifest and plan.
- The document, paragraph, and sentence tables.
- The eight restricted tables of the pairs and the two unique-unit tables.

The commands above reproduce the rest from the records and the checkouts:
453 shard manifests, their pinned copies, 453 finding artifacts, eight
first-appearance filters, and two unit selections. Their SHA-256 digests
are in [`digests.json`](digests.json). The checkouts are not redistributed.
