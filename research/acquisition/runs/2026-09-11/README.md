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
because documentation changes slowly between releases. The contrast above
therefore understates any shift in new text. The first appearance section
counts that text apart. The contemporary origin
is unknown or mixed: nothing here labels any text as AI-written, and a
contemporary prevalence is not a share of LLM text. No controlled cohort
exists yet; stage C needs a recorded budget before any generation.

## First appearance

The paragraph and sentence tables count units instead of documents. A
paragraph unit is one paragraph or comment block of a measured document.
List items, headings, table cells, and string literals are fragments. They
stay outside these tables.

| Cohort | Paragraphs | Prose words in them |
| --- | --- | --- |
| Historical | 38,328 | 481,307 |
| Contemporary | 39,536 | 505,326 |

At this level no contrast between the full cohorts exceeds one point.

A filter then lists the contemporary units whose exact text the historical
snapshot of the same repository does not hold. Every contemporary
repository has a historical snapshot. The five repositories with a
historical snapshot only enter the historical side alone.

| Kind | Contemporary units | Repeated | New |
| --- | --- | --- | --- |
| Paragraph | 41,974 | 20,154 | 21,820 |
| Sentence | 53,164 | 27,245 | 25,919 |

The restricted tables count only the new units on the contemporary side.
The historical side is unchanged. Four contemporary documents failed under
the policy. The filter lists their units, but they enter no table.

Files in this directory:

- `tables-paragraph.json` and `tables-sentence.json` hold all units of
  both cohorts.
- `first-appearance-contemporary-paragraph.json` and
  `first-appearance-contemporary-sentence.json` are the filters.
- `tables-first-appearance-contemporary-paragraph.json` and
  `tables-first-appearance-contemporary-sentence.json` are the restricted
  tables.

New contemporary paragraphs against all historical paragraphs. Each cell
gives units with a finding of all counted units. The interval is the joint
cluster bootstrap over the 41 components.

| Rule | Historical | Contemporary, new text | Difference | 95% interval | Ratio |
| --- | --- | --- | --- | --- | --- |
| `syntax.long-sentence` | 1,282 of 38,328 (3.3%) | 318 of 21,699 (1.5%) | -1.9 points | -3.6 to -0.4 | 0.44 |
| `syntax.passive-candidate-density` | 416 of 38,328 (1.1%) | 366 of 21,699 (1.7%) | +0.6 points | -0.1 to +1.4 | 1.55 |
| `readability.grade-metric` | 742 of 38,328 (1.9%) | 288 of 21,699 (1.3%) | -0.6 points | -1.7 to +0.4 | 0.69 |
| `syntax.parenthetical-load` | 296 of 38,328 (0.8%) | 243 of 21,699 (1.1%) | +0.3 points | -0.2 to +1.0 | 1.45 |
| `syntax.noun-stack` | 35 of 38,328 (0.1%) | 37 of 21,699 (0.2%) | +0.1 points | +0.0 to +0.2 | 1.87 |
| `format.em-dash-density` | 0 of 38,328 (0.0%) | 8 of 21,699 (0.0%) | +0.0 points | +0.0 to +0.1 | undefined |

The same rules at sentence level:

| Rule | Historical | Contemporary, new text | Difference | 95% interval | Ratio |
| --- | --- | --- | --- | --- | --- |
| `syntax.long-sentence` | 1,288 of 47,860 (2.7%) | 332 of 25,779 (1.3%) | -1.4 points | -2.8 to -0.2 | 0.48 |
| `syntax.passive-candidate-density` | 418 of 47,860 (0.9%) | 328 of 25,779 (1.3%) | +0.4 points | -0.1 to +1.0 | 1.46 |
| `readability.grade-metric` | 742 of 47,860 (1.6%) | 258 of 25,779 (1.0%) | -0.5 points | -1.4 to +0.2 | 0.65 |
| `syntax.parenthetical-load` | 296 of 47,860 (0.6%) | 235 of 25,779 (0.9%) | +0.3 points | -0.1 to +0.8 | 1.47 |
| `syntax.noun-stack` | 35 of 47,860 (0.1%) | 34 of 25,779 (0.1%) | +0.1 points | -0.0 to +0.1 | 1.80 |
| `format.em-dash-density` | 0 of 47,860 (0.0%) | 8 of 25,779 (0.0%) | +0.0 points | +0.0 to +0.1 | undefined |

One interval excludes zero. It points the other way from the hypothesis:
new contemporary paragraphs carry fewer long sentences than the historical
text. The passive and parenthetical rules move up by under one point. Their
intervals include zero. The em-dash rule fires in 8 new paragraphs and in
no historical one. Of the 19 candidate rules, 17 fire in neither cohort.
The weak-intensifier rule fires in 4 new paragraphs and 15 historical ones.

The protocol set its minimum useful difference of three points for
documents. Unit prevalences run near one percent. No unit-level difference
can reach three points while the ratio moves a great deal. The ratio column
is there for that reason. The protocol does not yet say what ratio would
count.

The filter compares each repository with its own earlier snapshot only.
Text copied across repositories still counts once per copy. So does text
repeated inside one snapshot.

The tool of the change that adds these tables built them from the measured
artifacts of this run. It also regenerated the document tables. Their
numbers are unchanged. They now name their unit and carry the counted
totals. A rerun of `scripts/measure-corpus.sh --resume` on the merged
commit reproduces the seven files byte for byte; their digests are in
[`digests.json`](digests.json).

## What is not in this directory

The 237 shard manifests, their pinned copies, and the 237 finding artifacts
are reproducible from the records and the checkouts; their SHA-256 digests,
and those of the tables and filters here, are in
[`digests.json`](digests.json). The checkouts are not redistributed.
