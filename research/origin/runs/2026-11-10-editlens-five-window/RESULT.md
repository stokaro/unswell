# EditLens five-window academic-paper screen: directional pass

The [protocol](PROTOCOL.md), [runner](run.py), independent [verifier](verify.py),
and [`freeze.json`](freeze.json) were committed as `6aa4a9d` before the new
human-only control scores. The freeze manifest has SHA-256
`243e1a0c6a01d72965b81f8a2393f64eeb6174ae7c53ca5caf8fbf872bd16113`.
The public source dataset and local FP32 model matched their pinned SHA-256s.
Inference used ONNX Runtime CPU. No generative or remote model call was made.

Each paper score is the mean of five selected paragraph scores. The 300
human-only calibration papers and 100 separate human-only holdout papers have
no paper overlap with the 58 matched human/AI test papers. Every selected
paragraph passed the 50-word cleaned-text minimum and the complete 512-token
window: **2,000/2,000** new human-only windows and the previously frozen
**580/580** matched test windows were eligible. No paper was replaced after
selection. The test generator/prompt choice and paragraph scores are exactly
those from the [earlier single-paragraph screen](../2026-11-10-editlens-academic-transfer/RESULT.md).

| Frozen paper-level measure | Result |
| --- | ---: |
| Mean-score threshold, strict `>` | 0.27447814408224075 |
| Calibration-human false positives | 2/300 |
| Holdout human-only false positives | **0/100** |
| Matched-original false positives | **1/58 = 1.72%** |
| Matched-paper win fraction | **55/58 = 94.83%** |
| Pooled paper-level AUROC | **0.9185** |
| AI rewrite papers detected | **26/58 = 44.83%** |
| Registered feasibility gate | **Passed** |

The [independent verifier](verify.py) reconstructed source selection and all
paper-level measures from the pinned dataset and no-text score files. It
passed, and [`verify_negative.py`](verify_negative.py) rejected four corrupted
evidence variants. [`result.json`](result.json) has SHA-256
`638cd495b1873963424fbabbd97718a429e97f638c70a5631bc27cd276cea9f0`;
[`scores.json`](scores.json) has SHA-256
`b415d9ef61763473805292d0e47e0fa17f6c4943178461dfa35bdcf6b348af24`.
Only source identities, hashes, token counts, and scores are committed; corpus
text and model weights remain outside the repository.

This is a useful aggregation signal on **exposed academic data**, not an
independent detector validation. The 58 test papers and their paragraph
scores were inspected before this follow-up was designed. The earlier
paragraph-level operating point detected 21/290 rewrites (7.24%) at 2/290
test-human false positives; the new paper-level 26/58 figure uses a separate
five-window calibration and a different unit of analysis, so the rates are not
interchangeable. Generator groups contain only 5–12 papers each. Zero false
positives among 100 holdout controls is insufficient to establish a precise
1% field false-positive rate.

The dataset's paper-year metadata do not prove that every human original was
exclusively human authored, and exact overlap with EditLens training sources
was not audited. Academic paragraphs differ from technical documentation;
the five selected windows do not constitute every paragraph of a document.
The 1.42 GB discriminative model also needs a product runtime and packaging
decision. This screen supplies neither source-local editorial diagnoses nor
meaning-preserving revisions, and changes no #349 metric or product behavior.

The registered pass warrants a **new, source-disjoint software-documentation
test** with independently checked human provenance, separate low-FPR
calibration, matched AI rewrites, and a whole-document window policy. No
origin classifier should be shipped from the exposed academic result.

## Replay

Download the source dataset and modelkit at the exact revisions and hashes in
the protocol. The inference runner checks the frozen packet before scoring.
The no-model evidence checks use Python's standard library:

```sh
python3 research/origin/runs/2026-11-10-editlens-five-window/verify.py \
  --data /path/to/paragraphs.jsonl
python3 research/origin/runs/2026-11-10-editlens-five-window/verify_negative.py \
  --data /path/to/paragraphs.jsonl
```
