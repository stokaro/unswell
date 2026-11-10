# EditLens academic transfer: ranking signal, low-FPR gate failed

The [protocol](PROTOCOL.md) and [runner](run.py) were frozen before any new
EditLens score. [`freeze.json`](freeze.json) has SHA-256
`0adcf9887ba86c8b7abb8171cf217963325b389a26f95117fd882f3fcdb7233e`.
The public AIWritingEffect paragraph file matched its pinned SHA-256; the
EditLens FP32 ONNX model and all nine small modelkit files matched the
previously frozen hashes. Inference used local CPU with no generative or
remote model call. No corpus text is included in this repository artifact.

The deterministic sample used 300 human-only calibration papers from
2017–2018 and five matched human/AI paragraph pairs from each of 58 separate
test papers marked 2017–2021. One generator/prompt variant per test paper was
selected by hash before scoring. All **880** selected paragraphs retained at
least 50 words after cleaning and fit the model's full 512-token window; none
was excluded or truncated. The 290 test pairs share 58 papers, so they are
not 290 independent source examples.

| Frozen measure | Result |
| --- | ---: |
| Calibration threshold, strict `>` | 0.47820500036080676 |
| Calibration human false positives | 2/300 |
| Test human false positives | **2/290 = 0.69%** |
| Matched-pair win fraction | **262/290 = 90.34%** |
| Paired pooled AUROC | **0.8350** |
| AI rewrites detected at threshold | **21/290 = 7.24%** |
| Test papers with positive mean score difference | 55/58 |
| Registered feasibility gate | **Failed**: required at least 20% TPR |

The independent [`verify.py`](verify.py) reconstructed the selected source
identities and all aggregate measures from the pinned public data and
[`scores.json`](scores.json). It passed; [`verify_negative.py`](verify_negative.py)
rejected four deliberately corrupted score/result variants. The complete
[`result.json`](result.json) has SHA-256
`a3e8bd0d05a15a1557fcac77d4cef4483b8d3d97f8da9c2733e8afa5965de5c8`;
the no-text scores have SHA-256
`6779dff96cbd320c940276cde5629f62ac7a9d7fcf9600961b3386629d91a299`.

The score shift is real in this selected set but too small for the frozen
low-false-positive operating point. Median score was 0.0409 for test humans
and 0.2278 for paired AI rewrites; the calibrated threshold was 0.4782.
Descriptively, the 55 Claude 5.1 Fable rewrite paragraphs had zero detections
at that threshold, while GPT-6 Astra had 2/60. These rows come from only 11
and 12 papers, respectively; they do not estimate model-wide failure rates.

## Exposed sensitivity and limits

Test rewrites had median 136.5 tokenizer tokens versus 143 for their human
sources; 201/290 rewrites were shorter. A shorter-text-only ranking gave
AUROC 0.5193 on these same pairs, far below EditLens's 0.8350. This post-result
check weakens a simple length-only explanation but does not rule out PDF
extraction, prompt, formatting, or paper-style artifacts.

The previous Unswell lexical-feature screen had already exposed this corpus,
though not to EditLens. The dataset's venue/year fields and the authors'
pre-ChatGPT PDF-date claim were not independently verified for every selected
paper; publication before 2022 is not proof of exclusively human authorship.
EditLens's published training domains are reviews, creative writing, web
articles, and news, rather than conference papers, but exact training-source
overlap was not audited. The model is a 1.42 GB discriminative classifier and
has no source-local editorial diagnosis or safe replacement text.

This particular transfer fails its registered low-FPR recall gate. Do not
raise the false-positive budget, retune the threshold on these test labels,
or ship an origin score from this result. A genuinely new software-
documentation corpus with verified human provenance, source/generator
holdouts, and a whole-document window policy is needed for product
qualification. Issue #349's frozen editorial FULL-event recall, accepted-
finding, localization, and edit-safety targets remain unchanged.

## Replay

Download the pinned paragraph file and modelkit revision named in the
protocol. Verify their hashes, then run:

```sh
python3 research/origin/runs/2026-11-10-editlens-academic-transfer/verify.py \
  --data /path/to/paragraphs.jsonl
python3 research/origin/runs/2026-11-10-editlens-academic-transfer/verify_negative.py \
  --data /path/to/paragraphs.jsonl
```

The inference runner checks the frozen protocol, code, selection, data, and
model before any score. The verifier requires only Python's standard library;
the inference environment is specified in the protocol.
