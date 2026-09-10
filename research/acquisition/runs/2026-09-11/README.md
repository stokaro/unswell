# Two-cohort run, September 11, 2026

This directory is the published record of the second stage B run of the
[pattern protocol](../../../methods/llm-patterns-v1.md). It holds the
historical cohort of the [first run](../2026-09-10/README.md), re-acquired
with snapshot-scoped source IDs, and a contemporary cohort of the same
repositories at their latest published GitHub release. Both cohorts share
one global plan. Every number in the tables is a rule outcome under one
policy on one cohort. It is not a false-positive rate, not recall, and not a
decision.

## What ran

```sh
GITHUB_TOKEN=... bash scripts/acquire-corpus.sh --cohort historical
GITHUB_TOKEN=... bash scripts/acquire-corpus.sh --cohort contemporary \
  --sources research/acquisition/sources-contemporary-v1.json
bash scripts/measure-corpus.sh
```

The measurement tool was built from commit
`528c919c5edc61b80f63c7ab2d728b8563e33beb` with a clean tree. The policy is
[`policy-e1.yaml`](policy-e1.yaml) on profile `strict-v1` with every catalog
rule enabled, configuration hash
`928269cf6586a4a22ed563c380783e6228589eadd2d15a377ca81a4f2369a56d`, ruleset
hash `554c01192fd28df96bb54294c764e031d3cefebee311ef2c50cde1919ad1a7c7`. The
rule classes are [`rule-classes-v1.json`](../../../methods/rule-classes-v1.json)
revision 1.

## Outcome

- Historical: 42 of 43 repositories acquired, 39 corroborated and 3
  `vcs_only`; `FasterXML/jackson-databind` again had no notice file found.
- Contemporary: 37 of 43 acquired, all corroborated by a GitHub release
  record. Three repositories publish no GitHub release object for their
  latest tag. Two, `pkg/errors` and `rust-lang/regex`, have a latest release
  before the contemporary window. One, `date-fns/date-fns`, names its notice
  file in a form the script does not know. The logs record each case.
- Dataset: 237 shards, 41 global components, 36 of them spanning both
  cohorts, 6,783 sources, all pinned and verified against the
  [plan](dataset-plan.json).
- Historical cohort: 3,328 documents (3 failed under the policy) over 116
  components, 194,353 units, 976,567 prose words, 6,020 findings, 6.16 per
  1,000 words.
- Contemporary cohort: 3,448 documents (4 failed) over 119 components,
  231,144 units, 1,145,832 prose words, 6,727 findings, 5.87 per 1,000 words.

## Contrast

The [tables](tables.json) contrast the contemporary cohort with the
historical baseline for every rule. The interval resamples the 41
components jointly, so a repository's two snapshots enter or leave a
replicate together. No difference reaches the protocol's minimum useful
difference of three percentage points.

| Rule | Historical | Contemporary | Difference | 95% interval |
| --- | --- | --- | --- | --- |
| `syntax.passive-candidate-density` | 10.0% | 12.6% | +2.6 points | -0.7 to +6.0 |
| `readability.grade-metric` | 21.1% | 23.2% | +2.1 points | -7.1 to +11.6 |
| `syntax.noun-stack` | 1.7% | 2.5% | +0.8 points | +0.0 to +1.6 |
| `format.em-dash-density` | 0.0% | 0.3% | +0.3 points | +0.1 to +0.6 |

Of the 19 `llm_associated_candidate` rules, 17 fired in neither cohort.
`filler.weak-intensifiers` stays at 0.5% in both, and
`filler.announced-importance` fires once in the contemporary cohort.

## Limits of this run

A contemporary snapshot repeats most of its repository's historical text,
because documentation changes slowly between releases. The protocol asks for
a separate analysis of newly added or changed text, counted at its first
evidenced appearance. That analysis is not in this run, so the contrast
above understates any shift in new text. The contemporary origin
is unknown or mixed: nothing here labels any text as AI-written, and a
contemporary prevalence is not a share of LLM text. No controlled cohort
exists yet; stage C needs a recorded budget before any generation.

## What is not in this directory

The 237 shard manifests, their pinned copies, and the 237 finding artifacts
are reproducible from the records and the checkouts; their SHA-256 digests
are in [`digests.json`](digests.json). The checkouts are not redistributed.
