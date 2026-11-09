# EditLens ARB transfer: positive origin screen, no editorial qualification

The [protocol](PROTOCOL.md) and [runner](run.py) were frozen in
[`freeze.json`](freeze.json), SHA-256
`3df65ef3f2534d7d5b2b6a765208420d31c49af077d79e09c5710f7cbc2cef8c`,
before the first model score. The pinned ARB input and FP32 ONNX model matched
their published SHA-256 values. The selection contained 300 human-only
calibration sources and 300 disjoint test sources, each with one H2L rewrite
and one free generation. The model ran locally on CPU; no generative model or
remote detector was called. Scores are an index, not a probability of AI
authorship or a percentage of AI-written words.

| Registered measure | Result |
| --- | ---: |
| Calibration human eligibility | 300/300 |
| Calibration threshold (strictly greater) | 0.2994394 |
| Test human false positives | 3/300 = 1.00% |
| H2L rewrite eligible pairs | 299/300 |
| H2L AUROC / TPR at threshold | 0.9771 / 214/299 = 71.57% |
| Free-generation eligible pairs | 299/300 |
| Free-generation AUROC / TPR at threshold | 0.9997 / 298/299 = 99.67% |

All predeclared feasibility conditions passed. The complete result is
[`result.json`](result.json), SHA-256
`bd00d9b4aa886fc1f69c0188e78cc67115c476a06d6242124bc88ad758261a8e`.
The row-level, no-text [`scores.json`](scores.json), SHA-256
`6a424308e49f8971f48d2d38b91b8051e3a6277496884199fbaa4e60864b7430`,
supports independent recalculation. [`verify.py`](verify.py) passed; changing
one stored score made it reject the result.

By domain, H2L TPR was 70/100 OpenWebText, 63/99 WritingPrompts, and 81/100
XSum. All three human false positives came from the 100 OpenWebText test
sources, giving 3% FPR in that domain. Across four generators, H2L TPR ranged
from 39/65 (Llama 3.2 3B) to 64/71 (Qwen2.5 7B). The result is not confined
to one generator, but per-domain low-FPR stability is unproven.

## Exposed sensitivity, not part of the registered gate

Human test texts had a median 307.5 words, H2L rewrites 230.5, and full
generations 262.5. A shorter-text-only ranking reached AUROC 0.692 for H2L
and 0.656 for generation, far below EditLens on the same pairs. Restricting
after the result to pairs where neither side exceeded 512 model tokens retained
225 H2L pairs (AUROC 0.975, TPR 72.9%) and 222 generation pairs (AUROC 0.9999,
TPR 100%). This sensitivity does not replace the frozen all-pair result or
remove all length/format confounding.

The [EditLens paper](https://arxiv.org/html/2510.03154) reports training on
XSum and WritingPrompts source collections, among others. ARB includes both
collections, and the EditLens training source IDs are not available for an
overlap audit here. Thus the calibration/test split is source-disjoint **within
ARB**, but independence from the model's training sources is unverified.
ARB also contains news, creative writing, and general web text rather than
software documentation. The [public model kit](https://huggingface.co/CoderBak/editlens_roberta_modelkit)
provides an attributed, ungated FP32 conversion of the authors' checkpoint;
its numerical smoke test is not an accuracy study. This screen cannot support
deployment, an authorship claim about an individual, or closure of #349.

The next test should apply this unchanged checkpoint to technical documents
with verified source provenance and a generator/source holdout, using a fresh
calibration set. Editorial issue #349 remains separate: even a reliable origin
score would not identify a necessary, source-local, meaning-preserving edit.
