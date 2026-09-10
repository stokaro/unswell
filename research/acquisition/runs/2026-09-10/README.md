# Historical pilot run, September 10, 2026

This directory is the published record of the first stage B run of the
[pattern protocol](../../../methods/llm-patterns-v1.md). It holds the
acquisition log, the acquisition records, the dataset manifest and plan, the
E1 policy, the E1 tables, and the digests of the artifacts that stay out of
the repository. Every number in the tables is a rule outcome under one
policy on one cohort. It is not a false-positive rate, not recall, and not a
decision.

## What ran

```sh
GITHUB_TOKEN=... bash scripts/acquire-corpus.sh
bash scripts/measure-corpus.sh
```

The measurement tool was built from the working tree of the commit that adds
this directory. The finding artifacts record producer revision
`995c814603ff5347d5466972319e3f23e6a8960d` with `modified: true`, because
the tree changed while the pilot was tuned. The policy is
[`policy-e1.yaml`](policy-e1.yaml) on profile `strict-v1` with every catalog
rule enabled. Its configuration hash is
`928269cf6586a4a22ed563c380783e6228589eadd2d15a377ca81a4f2369a56d` and its
ruleset hash is
`554c01192fd28df96bb54294c764e031d3cefebee311ef2c50cde1919ad1a7c7`. The rule
classes are [`rule-classes-v1.json`](../../../methods/rule-classes-v1.json)
revision 1.

## Outcome

- Sampling frame: 43 repositories; 42 acquired, 39 with a corroborated
  release date and 3 `vcs_only`. `FasterXML/jackson-databind` was not
  acquired because its notice file could not be found in the checkout; the
  [log](acquisition-log.json) records every outcome.
- Dataset: 117 shards, 41 global components, 3,330 sources, all pinned and
  verified against the [plan](dataset-plan.json).
- Measurement: 117 shards, 3,330 documents, 193,925 units, 972,852 prose
  words. Three documents failed under the policy (a rule budget) and stay in
  the artifacts as coverage gaps; no shard was skipped.
- Historical cohort: 3,327 measured documents over 116 components, 5,958
  findings, 6.12 per 1,000 words, 1,206 documents with at least one finding.

Of the 19 `llm_associated_candidate` rules, 18 never fired on this cohort;
`filler.weak-intensifiers` did, 23 times. Of the 20 `general_style` rules, 15 fired,
led by `syntax.long-sentence` (24.7% of documents, interval 18.4% to 31.4%)
and `readability.grade-metric` (21.1%). A rule with zero findings carries the
one-sided 97.5% Clopper-Pearson bound over 116 components, about 3.1%. The
[tables](tables.json) hold every row with its interval.

## What is not in this directory

The 117 shard manifests, their pinned copies, and the 117 finding artifacts
(77 MB) are reproducible from the records and the checkouts. Their SHA-256
digests are in [`digests.json`](digests.json), so a rerun can be compared
file by file. The checkouts themselves are not redistributed. Each record
names the repository, tag, commit, and license, and the acquisition script
obtains them again.

## Limits of this run

No controlled cohort exists yet, so the tables carry no contrast; stage C
needs a recorded budget before any generation. A zero count on 116 components
bounds the share of components that could carry the construction; it does not
show that the construction is absent from technical English. #162 and #163
record the two extractor limits the run met (a C grammar failure and the C#
parse cost).
