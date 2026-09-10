# ADR 0036: LLM-associated pattern evidence as a separate research branch

Status: accepted for the branch boundaries and protocol version 1; no corpus,
measurement, evidence card, or rule decision exists yet.

## Context

[#154](https://github.com/stokaro/unswell/issues/154) asks which concrete
constructions go with LLM-generated and LLM-edited technical English, how
stable that link is, and which of those constructions an explainable rule can
detect at an acceptable warning load. The answers must come from real data and
executed experiments, without hiring annotators, building an annotation
platform, training a large model, or buying GPU time.

[#22](https://github.com/stokaro/unswell/issues/22) is on hold for resource
reasons. The editorial protocol `unswell-research-v1` predicts `needs_revision`
from independent human labels, so it cannot serve a branch that has no human
labels. [ADR 0015](0015-research-methodology.md) and
[#59](https://github.com/stokaro/unswell/issues/59) keep editorial quality and
origin apart, keep origin metadata out of editorial features, and forbid an
origin estimate from blocking CI. This branch has to fit inside those rules
without changing their meaning.

## Decision

Open a third kind of research evidence, pattern association evidence, beside
the two kinds the project already records.

| Evidence | Question it answers | Where it lives |
| --- | --- | --- |
| Editorial label | Does a human reader judge this unit to need revision? | Annotation rounds and `unswell-research-v1` |
| Origin estimate | How similar is this unit to a defined provenance class? | The opt-in origin channel and its pack |
| Pattern association | How often does a defined construction occur in each cohort, and how much does an LLM operation change it? | The `unswell-llm-patterns` protocol and its evidence cards |

Pattern association evidence describes constructions and cohorts. It assigns
no label to a unit, trains no classifier as a requirement, and produces no
probability. Cohort membership (historical, controlled generation, natural
AI-heavy, contemporary) is the comparison axis of this branch. Research
measurement uses it as a grouping factor and nothing else. It never enters an
editorial model, a product pack, a rule input, or a gate decision.
That is the explicit boundary that reconciles this branch with #59: origin
stays out of editorial inputs and out of CI, and this branch measures rules
against cohorts instead of estimating origin.

The branch keeps three claims apart and studies only the first two. First, a
rule recognizes a formally described construction. Second, the construction
has a stable link with the studied LLM workflows. Third, removing the
construction helps the reader. The third claim is a human judgment. It stays
with the deferred human-validation branch, and no association can stand in
for it.

A historical firing rate is not an editorial false-positive rate. The share of
LLM outputs with a finding is not recall of all unwanted patterns. Documents
and reports from this branch use the terms prevalence, absolute difference,
prevalence ratio, paired change, and warning load, and they say which cohort
each number describes.

### What this evidence cannot do

- It cannot fill an annotation response, an adjudication, or a decision export.
- It cannot move `human_corpus` above `not_qualified` in any artifact.
- It cannot mark an origin pack or a revision pack as accepted.
- It cannot count toward the 5,000 labeled units, two annotators, or 1,000
  held-out units that #22 requires.
- It cannot change a profile, a default gate, or a severity by itself. A product
  change that cites an evidence card is a separate decision with its own ADR.
- It cannot turn an association score into a share of AI authorship or a
  revision probability.

### Reuse and research adapters

The branch reuses the public `extract`, `nlp`, and `feature` packages, the
prepared-unit contract, the corpus planner and candidate artifacts of
[ADR 0014](0014-corpus-acquisition.md), the evaluation strata of #153, and the
stage cost measurement of #149. No second linter, parser, feature formula,
index, or reporter belongs to it. Each missing capability becomes an adapter
in the existing research module, with a strict schema, bounded limits, and
tests on frozen fixtures. An adapter never creates an annotation round or an
actor record to satisfy an existing input requirement.

Data acquisition and any model call live in research tooling outside the
product. The library, CLI, MCP server, ordinary tests, and ordinary CI stay
offline, pure Go, `CGO_ENABLED=0`, and free of downloads and child processes.
Ordinary CI checks the adapters on compact frozen fixtures; the full corpus is
reproduced by a separate documented run.

### Budget

No paid generation is authorized as of September 10, 2026. A paid call needs a
recorded budget amendment in the protocol that names the amount, the caps on
requests, tokens, retries, and wall-clock time, and the person who authorized
it. Without that record, any generation tooling built for this branch must
run in dry-run mode and make no paid call.
Reproduction of a published analysis never requires paying for generation
again, because the exact saved outputs are the input to every analysis.

## Dependency check against main

Checked on September 10, 2026 against commit
`f3a9d215394c11eae765c350d54f0e82d515198c`. The table records what the branch
can use today and what it still lacks. Old issue comments were not treated as
evidence of a capability.

| Capability | State on main | Gap for this branch |
| --- | --- | --- |
| Source-preserving extraction and English NLP | Public packages with exact source segments, protected regions, and capability failures | None |
| Shared feature contract (#56) | Block and prepared-unit contracts, activation features, repetition family | Candidate constructions that need dependency parses stay unavailable until a real backend exists |
| Corpus manifest and planner (ADR 0014) | Sources carry hashes, rights, notices, eleven role values with separate error, log, and UI string roles, an origin label scoped to a document or repository, and partitions. `format` covers the eighteen extractor languages. Components are frozen before extraction | No fields for the snapshot date, the cohort, the date confidence, or the dating evidence. No `unknown` role value. No near-duplicate or fork detection: until a versioned detector with a fixed threshold exists, forks and near duplicates need a curated related-version or template key |
| Candidate limits | 10,000 sources or units per artifact, 128 MiB artifacts, 64 MiB of declared bytes | Sharded artifacts under one dataset manifest, with grouping, deduplication, and partition checks across shards |
| Label-free measurement | `join` measures features but requires an annotation round; `evaluate` requires decisions | A measurement path over candidates that records rule activations and findings without a round |
| Generation records | `origin.generation_record` is one free-text field per source | A record per response with prompt, model identity, parameters, raw output, transformations, status, cost, and hashes |
| Evaluation strata (#153) | Words, role, source language, prose language, origin | Cohort, period, operation, generator family, and prompt condition strata |
| Uncertainty | Paired group bootstrap for prediction metrics (#117); a zero-count case returns `insufficient_evidence` | Cluster intervals for prevalence, absolute difference, prevalence ratio, and paired change; a one-sided exact binomial bound over components; exploratory false-discovery control |
| Stage cost (#149) | Measured per pipeline stage with peak memory | None |
| Method registry (#55) | RAID and DetectRL-X reviewed with data use deferred | Six studies recorded in the [sources record](../../research/methods/llm-patterns-sources-v1.json): RAID and DetectRL-X reconfirmed, plus Reinhart, Kobak, Liang, and Rallapalli |
| Ptah candidates | 378 unlabeled development candidates from one repository | Ptah stays a development case study; it is never the sole confirmation or a final test |

## Consequences

The active research roadmap is this branch, tracked by #154. It delivers
two results: v0, the historical corpus with baseline measurements and an
executable comparison protocol, and v1, the confirmatory study on comparable
LLM data. Unavailable LLM data completes v0 alone and blocks nothing else. The
human validation branch, tracked by #22 with #23, #24, #25, #26, #50, #51,
#52, #53, #56, #57, and #58 bound to it, stays open and unchanged. The
[research plan](../research.md) lists the two branches side by side.

The branch assigns partitions with the editorial seed and weights, so a
component that both protocols touch lands in the same partition under both
and the editorial final test is preserved by construction.

This branch records the exposure history of every discovery and evaluation
group. After a developer has seen results on a group, #22 cannot use that
group as an independent human held-out benchmark; #22 reserves new groups for
that purpose.

A negative result is a complete result. The branch may end with the conclusion
that some or all candidates are not useful enough. That conclusion authorizes
neither shipping an unconfirmed rule nor removing a useful general rule that
happens not to separate the cohorts.
