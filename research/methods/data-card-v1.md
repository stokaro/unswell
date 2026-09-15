# Data card: the pattern protocol corpus, version 1

What the corpus is, where it came from, what may be done with it, and what it
cannot answer. Written for a reader who wants to reuse the measurements
rather than repeat the collection.

## What it holds

28,740 measured documents and 5,186,062 prose words across six cohorts,
drawn from 76 pinned repository snapshots and grouped into 73 independence
groups.

| Cohort | Documents | Prose words | What it is |
| --- | ---: | ---: | --- |
| historical | 7,717 | 2,119,351 | Snapshots that substantially predate mass LLM use |
| historical-2018 | 2,968 | 812,491 | Placebo boundary |
| historical-2016 | 2,269 | 528,390 | Placebo boundary |
| historical-2012 | 478 | 142,903 | Placebo boundary |
| contemporary | 3,452 | 1,156,467 | Present-day snapshots of unknown origin |
| controlled | 11,856 | 426,460 | Model output under recorded prompts |

Partitions: 7,421 training documents, 13,134 development, 1,655 calibration,
6,530 final test. A repository and everything derived from it stay in one
partition.

Roles: comment, documentation, readme, release_note, specification. A role
that could not be determined stays undetermined and is not merged into
another.

## Where the controlled arm came from

Twenty runs, five models, three families.

| Family | Models | Records |
| --- | --- | ---: |
| Anthropic | claude-opus-5, claude-haiku-4-5-20251001 | 6,328 |
| OpenAI | gpt-5.6-luna | 3,928 |
| Qwen | Qwen3.6-35B-A3B, Qwen3.8-27B | 1,600 |

Each record keeps the task, the exact request text, the raw response, the
model identity, the date, the status and the hashes, under
`research/generation/runs/`. Two operations appear: generation from a
deterministic fact sheet, and polishing of the original human text. No
response entered on its score, and none was asked for again to obtain a
style or a length.

## Provenance and its limits

Cohort membership states where a text came from. It does not label any
passage as good or bad prose, and it does not label any passage as
AI-written at the fragment level.

- `historical` is a snapshot with a stated confidence level in its origin,
  not a guarantee of human authorship.
- `contemporary` is `mixed/unknown` origin. It measures prevalence and
  distribution shift, nothing about the share of model-written text.
- `controlled` is model output whose generation conditions are recorded.
  That is the only arm where origin is known rather than inferred.

`human_corpus` stays `not_qualified` in every artifact. The human judgments
that would change it are issue 22, which is on hold.

## Licensing

Every source snapshot carries the license of the repository it came from,
recorded per source in the acquisition records. The corpus is a set of
references and measurements over those snapshots, not a redistribution of
them. One repository's license is not permission for another's data.

## What the corpus supports

Group comparisons of construction frequency, warning load per document and
per thousand words, paired comparison of an original against its edit, and
temporal comparison across the placebo boundaries.

## What it does not support

- An editorial false-positive rate. A firing rate on declared provenance is
  not a rate of unwanted findings.
- A recall of unwanted patterns. Without per-passage labels of pattern
  presence, precision and recall are unknown.
- Authorship of any single document.
- A claim about English in general. The sampling frame is open-source
  technical prose in one language, chosen by license and history
  availability.

## Size and storage

The measured corpus is about 2.4 GB and is not committed. The pinned shard
manifests, the dataset plan and the generation records are committed;
`bash scripts/measure-corpus.sh` rebuilds the measurement from them and from
the repository checkouts named in the acquisition records.

## Known defects and corrections

The extraction read example code inside documentation markup as prose until
September 15, 2026. Every measurement before that date carries it. Issue 279
records the defect, and `research/methods/llm-patterns-v2.md` carries the
before-and-after of the recount.

## Version

Version 1, September 15, 2026. The dataset plan hash is in
`artifacts/measurement/dataset-plan.json`; the rule catalog and policy
identity are recorded per shard in the findings artifacts.
