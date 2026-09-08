# Research plan

Unswell aims to reduce formulaic AI-style wording in code and documents. Good
AI-assisted prose should pass; poor human-written prose should receive actionable
findings. Quality and provenance need separate labels and evaluation targets.

The following tasks are experiments. No model training, comparative benchmark, or
production accuracy result is claimed by this plan.

The [annotation rubric](editorial-annotation.md) now defines independent quality
and origin records, reviewer instructions, adjudication, and data partitions.
Its [Go tools](../research/annotation/README.md) validate rounds, prepare blinded
packets, and measure agreement. The included tutorial has scripted responses and
does not count as the human-labeled corpus or a completed pilot.

The [corpus preparation tool](../research/annotation/corpus/README.md) freezes
source groups before extraction and verifies candidates against exact originals.
Its eight pinned Ptah files produce 378 unlabeled development candidates; they
provide workflow tests, with no human judgments or independent final-test evidence.

## Execution order

| Step | Issue | Deliverable |
| --- | --- | --- |
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
[roadmap](roadmap.md); generated labels cannot fulfill them.

The first comparison uses inexpensive features and measures whether they add value
beyond the phrase catalog. LLMDet then tests the value of stored probability tables.
Its published scan path avoids generative-model inference, but data size, tokenizer
compatibility, and transfer to short technical prose still require measurement.
See the [LLMDet paper](https://aclanthology.org/2023.findings-emnlp.139/) and
[implementation](https://github.com/TrustedLLM/LLMDet).

## Runtime and reporting boundaries

Python may support isolated experiments and model conversion. The shipping Go
library, CLI, and MCP server remain offline and pure Go. Their feature ordering,
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
