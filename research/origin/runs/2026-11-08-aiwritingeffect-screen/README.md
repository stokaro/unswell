# Paired academic-prose origin-feature screens

The [AIWritingEffect dataset](https://huggingface.co/datasets/Jerry999/AIWritingEffect)
links human conference-paper paragraphs and full paper variants to AI rewrites
of the same content. Its published metadata separates papers across train,
validation, test, and human-only groups. This is an external technical-prose
test of the **unchanged** lexical features used in the earlier AITDNA screen.
It is not a test of Unswell's editorial findings or a claim about authorship
of any individual paper.

The [primary paragraph protocol](PROTOCOL.md), [script](run.py), and feature
source were sealed in [freeze.json](freeze.json) before any feature was scored.
The input is the public Hugging Face file at commit
`6a45eddc021257b7c49de398c2fffa370c1d1bfa`, SHA-256
`6e0eb8058578afce4aaf81c822b70762afd9f18534eb5f1e89aa976cd051b793`.
The complete [result](result.json) has SHA-256
`61da1fc4b8cc74ef025782bc9d9c281c052eed179a85dd067e817e69da857730`.
No raw dataset text is committed.

| Fixed 100-prose-token paragraph prefix | Result |
| --- | ---: |
| Test AI rows / eligible matched pairs | 7,015 / 2,131 (30.4%) |
| Lexical-richness pooled AUROC / matched-pair win fraction | 0.686 / 85.8% |
| Entropy pooled AUROC / matched-pair win fraction | 0.668 / 83.5% |
| Richness TPR at a validation-human 1% FPR threshold | 6.0% |
| Observed richness FPR on eligible test human / separate human-only paragraphs | 0.61% / 0.51% |

The registered richness feasibility rule **failed**: only 30.4% of AI rows
had an eligible human match, and TPR was below its 10% minimum. The reported
false-positive rates condition on eligible paragraphs; 2,830/4,300 test
human and 14,395/21,013 human-only paragraphs were too short for this
100-token feature. The low rate on eligible rows does not establish useful
whole-corpus detection.

After seeing that result, a separate [document protocol](PROTOCOL-v2.md),
[script](run_docs.py), and [freeze](freeze-v2.json) were recorded before
document scores were computed. The pinned `docs.jsonl` input has SHA-256
`428de968a71d64ddaf9f0d5afe7e4e295b2adeace77f57fd29c6252ae082d29c`.
The complete [document result](result-v2.json) is a **post-result sensitivity
check on the same underlying papers**, not independent confirmation.

| Document measure on 89 test papers / 143 AI variants | Eligible pairs | Richness AUROC | Matched-pair wins |
| --- | ---: | ---: | ---: |
| All available prose | 143 | 0.902 | 143/143 |
| First 100 prose tokens | 143 | 0.798 | 134/143, counting ties as half |

The median AI-to-human prose-token ratio was 0.966. A post-result artifact
inspection found median differences of zero in the feature's raw/prose-token
fraction and in backslash frequency; the first-100-token comparison also
retained a positive signal. These checks make a simple line-wrapping or
code-stripping explanation less likely, but do not rule out PDF extraction,
paper structure, prompt, or generator artifacts. The source papers' asserted
pre-ChatGPT provenance was not independently verified. The 143 variants share
only 89 human papers, so the pooled AUROC is correlated and must not be read
as 143 independent source samples. The validation split has only 83 human
papers, too few for a stable document-level 1% FPR operating point.

The practical next step is a **new paper- and prompt-held-out corpus** with
independently checked human provenance, document-level negative controls, and
a low-FPR endpoint. Do not retune this exposed cohort or ship an origin score
from these screens. No #349 event, accepted editorial finding, or suggestion
safety result changed.

## Replay

Download the two files from the pinned dataset commit and verify the SHA-256
values above. From the repository root, using Python 3.12 or later:

```sh
python3 -B research/origin/runs/2026-11-08-aiwritingeffect-screen/run.py \
  --input /path/to/paragraphs.jsonl
python3 -B research/origin/runs/2026-11-08-aiwritingeffect-screen/run_docs.py \
  --input /path/to/docs.jsonl
```

Each script checks its own hash, protocol hash, feature-source hash, and input
hash against its frozen manifest before scoring. No model, network, or private
data is needed for replay.
