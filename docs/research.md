# Research plan

Unswell aims to reduce formulaic AI-style wording in code and documents. Good
AI-assisted prose should pass; poor human-written prose should receive actionable
findings. Quality and provenance need separate labels and evaluation targets.

## Purpose and non-goals

The product finds concrete constructions in technical prose and explains
what to rewrite: stock phrases, rhetorical templates, repeated sentence
shapes, and their density in a passage. The reader usually knows already
that a text came from an AI agent. [ADR 0037](adr/0037-diagnostics-not-authorship.md)
records the goal, the audit behind it, and the next stage.

It is not an authorship detector, not a generic generation detector, and
not a judge of literary quality. A clean scan of a generated text is a
correct result. A finding on a human-written text is not a false positive
by that fact. An error is a diagnostic the stated rule does not justify.
No index, score, or estimate stands for a probability that a tool wrote
the text.

The historical corpus is the comparison base and a source of candidate
constructions, not a norm. Rarity in old text is not a defect, and a new
term earns no finding for being new. Contemporary generations show which
constructions today's models use more often, more uniformly, or in less
fitting places than the historical texts of the same kind. "Unusual
against the historical corpus" and "frequent in generations" are two
observations; neither proves the other.

## What the recorded experiments show

The E3 runs of [#176](https://github.com/stokaro/unswell/issues/176),
[#177](https://github.com/stokaro/unswell/issues/177), and
[#178](https://github.com/stokaro/unswell/issues/178) fit classifiers of
cohort membership, and the origin runs of
[#179](https://github.com/stokaro/unswell/issues/179) fit a classifier of
generation endpoints. Every one is negative: no feature family, no
compression reference, and no probability table separates the cohorts.
Those are answers to classification questions. They say nothing about
whether a rule names a construction worth rewriting, and they retire no
rule. The records stay in [research/baselines](../research/baselines/README.md)
and [research/origin](../research/origin/README.md) as negative results.

The historical corpus measures the existing rules by cohort and period,
with placebo boundaries and paired changes. It proposes nothing yet. Six
papers and the catalog fixed every studied construction in advance, the
prevalence tables carry no role stratum, and the corpus holds GitHub
repositories only. The next stage in the ADR fills those gaps.

The following tasks are experiments. No model training, comparative benchmark, or
production accuracy result is claimed by this plan.

[ADR 0015](adr/0015-research-methodology.md) and the
[versioned protocol](../research/methods/protocol-v1.md) now fix the targets,
A–G comparison, split/tuning procedure, and selection criteria before fitting.
The [method registry](../research/methods/README.md) records pinned source review,
component terms, measured source-file hashes, and scoped decisions. None of its
methods has been reproduced, ported, or qualified by this research work.

The [annotation rubric](editorial-annotation.md) now defines independent quality
and origin records, reviewer instructions, adjudication, and data partitions.
Its [Go tools](../research/annotation/README.md) validate rounds, prepare blinded
packets, and measure agreement. The included tutorial has scripted responses and
does not count as the human-labeled corpus or a completed pilot.

The [corpus preparation tool](../research/annotation/corpus/README.md) freezes
source groups before extraction and verifies candidates against exact originals.
Its eight pinned Ptah files produce 378 unlabeled development candidates; they
provide workflow tests, with no human judgments or independent final-test evidence.

## Two research branches

Research now runs on two branches. The active branch collects pattern
association evidence without human labels. The deferred branch keeps every
requirement that needs human judgment and waits for annotation resources.
[ADR 0036](adr/0036-llm-pattern-evidence.md) fixes the boundary between them.

### Active: LLM-associated patterns

[#154](https://github.com/stokaro/unswell/issues/154) counts how often
defined constructions occur in four cohorts. The cohorts are old technical
English, saved LLM outputs under known conditions, AI-heavy projects, and
current text. It also measures how much an LLM operation changes them. The
[pattern protocol](../research/methods/llm-patterns-v1.md) fixes the cohorts,
the date boundary, the experiment, the partitions, the measures, the
statistics, and the budget caps for the pilot. The
[sources record](../research/methods/llm-patterns-sources-v1.json) pins the
six reviewed studies with their data terms and decisions.

| Stage | Deliverable |
| --- | --- |
| A. Scope | ADR 0036, protocol version 1 with budget caps, sources record, pilot prompts, #22 on hold |
| B. Historical corpus | Dated sources with confidence levels, a global plan, verified shards, and baseline rule measurements |
| C. Comparable experiment | Saved permitted generation records, pilot tables, sample size, and protocol version 2 |
| D. Confirmatory study | Executed comparisons, holdouts, intervals, ablations, and negative results |
| E. Evidence release | Versioned data set, data card, report, table reproduction, and evidence cards |

This branch produces prevalence, absolute differences, prevalence ratios,
paired changes, and warning load per cohort. It produces no editorial label,
no origin estimate, and no probability, and its artifacts keep
`human_corpus: not_qualified`. No paid generation is authorized until a
maintainer records a budget in the protocol.

### Deferred: human validation

[#22](https://github.com/stokaro/unswell/issues/22) is on hold with its text
and acceptance criteria unchanged. The table below lists the branch. #55 and
#21 are complete, and #22 is the labeled corpus itself. The remaining issues
need that corpus and stay open until it exists; the
[acceptance audit](acceptance.md) records their implementation.

| Step | Issue | Deliverable |
| --- | --- | --- |
| Fix research design | [#55](https://github.com/stokaro/unswell/issues/55) | Methodology ADR, versioned source registry, and comparison/selection protocol |
| Share features | [#56](https://github.com/stokaro/unswell/issues/56) | One capability-aware feature contract for rules, training, inference, and explanations |
| Execute comparisons | [#57](https://github.com/stokaro/unswell/issues/57) | Locked run manifests, grouped metrics, and saved predictions for A–G |
| Define labels | [#21](https://github.com/stokaro/unswell/issues/21) | Independent quality and provenance annotations, reviewer rubric, and adjudication |
| Collect data | [#22](https://github.com/stokaro/unswell/issues/22) | Licensed technical corpus, real human judgments, and grouped splits before fragment extraction |
| Compare small models | [#50](https://github.com/stokaro/unswell/issues/50) | Rules, lexical n-grams, stylometry, logistic regression, and small tree ensembles |
| Export and validate | [#23](https://github.com/stokaro/unswell/issues/23) | Reproducible feature contracts, model artifacts, and pure-Go inference with reference parity |
| Define applicability | [#24](https://github.com/stokaro/unswell/issues/24) | Null estimates with explicit reasons, empirically chosen limits, and consistent reports |
| Measure held-out results | [#25](https://github.com/stokaro/unswell/issues/25) | False-positive rates, recall, calibration, uncertainty, subgroup results, and ablations |
| Evaluate LLMDet | [#51](https://github.com/stokaro/unswell/issues/51) | Tokenizer/table parity, coverage, resource measurements, and an adoption decision |
| Evaluate compression | [#52](https://github.com/stokaro/unswell/issues/52) | Added value over lexical features and sensitivity to reference corpora |
| Compare model-based methods | [#53](https://github.com/stokaro/unswell/issues/53) | Optional research harness, costs, results, or justified exclusions |

The corpus includes documentation, doc comments, ordinary comments, release notes,
and selected strings, with results reported separately by role and length. Include
good technical English from non-native writers, mixed editing workflows, unseen
generators, and later data. Keep related documents, authors, prompts, templates, and
rewrites in the same partition. Human annotation requirements remain those in the
[roadmap](roadmap.md); generated labels cannot fulfill them, and neither can the
cohort measurements of the active branch.

The first comparison uses inexpensive features and measures whether they add value
beyond the phrase catalog. LLMDet then tests the value of stored probability tables.
Its published scan path avoids generative-model inference, but data size, tokenizer
compatibility, and transfer to short technical prose still require measurement.
See the [LLMDet paper](https://aclanthology.org/2023.findings-emnlp.139/) and
[implementation](https://github.com/TrustedLLM/LLMDet).

## Runtime and reporting boundaries

Primary editorial training and calibration run on Go. Python may support isolated
reference experiments and port comparison data. The shipping Go library, CLI,
and MCP server remain offline and pure Go. Their feature ordering,
normalization, tokenization, and model outputs need golden parity with the reference
implementation before an exported model can be accepted.

Exact editorial findings, heuristic indexes, calibrated revision probabilities, and
optional provenance estimates have different meanings. An origin experiment is
advisory and disabled by default; it does not determine the initial CI result.
A block score cannot supply sentence-level finding locations. Short, unsupported,
or poorly covered inputs receive an explicit reason such as `insufficient_evidence`;
applicable editorial rules still run.

Fast-DetectGPT and Binoculars require model-produced probabilities and belong in
the optional research harness. DivEye's published code has a separate license
constraint, and Ghostbuster's published workflow sends text to an API. Their issue
records these limits before any execution or source reuse. Watermark verification
requires participation during generation and is outside this linter's scope.

## Vale and Prose

Unswell already uses Prose 3.2.1 for English NLP. Vale 3.20.0 uses the same version.
[Issue #10](https://github.com/stokaro/unswell/issues/10) evaluates targeted reuse of
Vale's MIT-licensed sequence matcher for bounded token/POS rules. Keep the original
copyright after `package`, the complete notice in distributed licenses, and the
upstream commit and file provenance. Adaptation must preserve Unswell's source
mapping, RE2 limits, and public-library boundaries.

General readability, formality, or repeated API names alone do not establish an
AI-style defect. New features must earn their place through contextual evidence and
measured false positives. An optional Vale style package can distribute compatible
rules later, with one policy catalog as the source.
