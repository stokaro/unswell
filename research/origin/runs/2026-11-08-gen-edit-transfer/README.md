# GEN edit-operation transfer: result

The scoring code and thresholds were committed as `ad5a06e662c493c65c205415f6c8347f9d1c2aef` before `result.json` was produced. This is a model-free origin-feature study, not a release of an AI-origin detector. It tests proxies motivated by [Shan, Lee, and Hao, *AI Writers Have a Consistent Stylometric Footprint, but AI Editors Do Not*](https://arxiv.org/abs/2608.27855v2) on [GEN](https://huggingface.co/datasets/szyszy/GEN). The proxies and Unswell tokenizer differ from the paper's implementation. The test split is source-disjoint within each of five domains; only human calibration sources determine thresholds.

## Main result: first 100 prose tokens

| Task and prespecified direction | Eligible / available machine variants | AUROC | Paired wins | TPR | Test-human FPR | Decision |
| --- | ---: | ---: | ---: | ---: | ---: | --- |
| Full generation, higher lexical richness | 17,785 / 19,684 | 0.880 | 0.910 | 36.2% | 0.32% | Research signal; external transfer required |
| Full generation, higher token entropy | 17,785 / 19,684 | 0.865 | 0.898 | 38.2% | 0.48% | Research signal; external transfer required |
| Editing, higher closed-class fraction | 45,639 / 46,407 | 0.295 | 0.217 | 0.15% | 1.20% | Fails frozen feasibility gate |
| Editing, lower token entropy | 45,639 / 46,407 | 0.350 | 0.291 | 0.22% | 0.80% | Fails frozen feasibility gate |

TPR uses the frozen human-only upper-tail threshold for each feature. AUROC pools correlated variants as a descriptive statistic, not as independent examples for uncertainty estimates. Paired wins compare each machine variant with its specific human source; ties count as half. Generation results span 2,497 eligible test source groups. Editing results span only 498 test source groups (of 2,500), so the high variant eligibility must not be mistaken for broad source coverage. Full-prose metrics and editing richness are secondary in [`result.json`](result.json).

The full-generation signal appears in all five domains, but its fixed-threshold performance is uneven across models: lexical-richness TPR ranges from 1.9% for `gpt-oss-120B` to 68.1% for `gpt-oss-20B`. The headline pooled TPR therefore does not establish model-independent detection. Editing's two prespecified directions are reversed in every domain; the paper-inspired proxies do not transfer to this setup. Editing richness rises (AUROC 0.658, TPR 2.9%), but it was a secondary observation and does not satisfy the detection gate. Do not select it post hoc as a replacement primary hypothesis.

## Source and data integrity

GEN revision `a0f143c1ae9c35b684898a5f9c0ab6efb49be486` was pinned. The raw human and generated files contain 5,000 and 39,391 rows. The full edited file was streamed and hashed: 1,521,045,885 bytes, SHA-256 `2e224f472103e9dc4f569eeba91451734985421d67cebd41f79491198211cf25`, 273,418 rows. This differs by two from the paper's edited-row count; no rows were fabricated to match the paper. Of the edited rows, 1,116 `story_generation` variants have `seed_text` that does not match the human source after whitespace normalization. They were excluded from paired analysis. The fixed 30%-target selection has 90,768 source-matched rows and SHA-256 `a051dc08d31f0457a3a89b4e0e530e99872060760f56414b2269c17da3a64d45` as gzip. The generated file has 38 duplicate source/model identities and 647 missing source/model combinations; duplicates were resolved by the smallest UTF-8 text hash before any feature calculation.

The local raw corpus is not committed. [`freeze.json`](freeze.json) pins its file hashes, the extraction and scoring sources, lexicon, and source split. [`decision.json`](decision.json) pins the saved result digest. `make check-gen-edit-transfer` verifies the committed artifacts without downloading the corpus; full recomputation requires downloading the pinned GEN revision and running `acquire.py` then `run.py` with the four input paths. The acquisition and runner reject changed hashes. The complete per-domain, per-model, and per-edit-category denominators and metrics are in the saved result.

## Scope of the finding

This is a positive *screen* for full generation on GEN and an explicit negative for the two editing hypotheses. It does not identify the origin of a particular document. The next defensible step is a new, separately frozen external transfer test on technical prose with unseen source groups and prompt families. No product rule, score, or origin claim changes here.

Issue #349 concerns editorial diagnosis and safe revision, not AI-origin classification. This study contributes **zero** #349 FULL-event recall, accepted findings, or revision-safety evidence. Its status and frozen gates remain unchanged.
