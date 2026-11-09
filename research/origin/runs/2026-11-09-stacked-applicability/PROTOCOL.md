# Stacked span filtering: calibration applicability screen

Date: November 9, 2026. This is an offline, non-generative applicability
screen inspired by [Wu et al. (2026)](https://arxiv.org/abs/2605.23190),
whose public [MGTD implementation](https://github.com/Daftstone/MGTD) uses a
stronger neural base detector. It is **not** a reproduction of their reported
detector or a qualification of an Unswell origin feature.

## Frozen inputs and roles

Use the five-feature logistic model and source-disjoint ARB role assignment
from `2026-11-08-interpretable-ensemble`, unchanged. Verify the exact hashes
of that model result, feature code, lexicon, ARB loader, ARB parquet, and
AIWritingEffect JSONL in `freeze.json`. This screen reads only the 300 ARB
calibration human sources, their four `h2l` rewrites each, and the 83 academic
validation humans plus 131 AI rewrites. It must not score ARB test sources,
academic test or human-only papers, or any #349 editorial cohort.

The source corpora were exposed before this protocol. The role restriction
avoids using test scores to decide whether a stacked test is worth running;
it does not turn the screen into independent confirmation.

## Fixed method and stop rule

Extract prose and tokenize exactly as the frozen feature code does. From each
document take up to four **non-overlapping 100-token windows**, ignoring any
short trailing window. Use the frozen logistic model to score each full
window. A document with fewer than one full window abstains. No model is
fitted, fine-tuned, or called remotely.

Apply the paper's conservative inference mask: a window is eligible for
filtering only if its logistic machine probability is strictly below `0.01`.
Among eligible windows filter at most `ceil(0.25 * window_count)` lowest
scores, while retaining at least one window. Count the documents with at
least one filtered window and the total filtered windows. Do not adjust the
threshold, window size, or cap after seeing results.

Proceed to a separately frozen source-disjoint test only if at least **5% of
eligible ARB calibration AI rewrites** and **5% of eligible academic validation
AI rewrites** have a filtered window. The dual gate requires the method to
actually modify inputs in both domains. Otherwise stop: with the weak frozen
base classifier, this particular paper-style mask cannot plausibly improve
document decisions often enough to justify a full test. Report human-control
filter rates and eligibility regardless of outcome. This is an activation
gate, not a recall or false-positive gate; passing it would prove no quality
gain by itself.

The paper combines a detector's second pass over retained text with optional
hard-EM training. This screen measures **only** whether its `0.01` mask can
activate with our existing non-LLM classifier. It does not approximate the
second pass, use a different confidence calibration, or claim the paper's
reported gains transfer to Unswell. Preserve the frozen result and a negative
outcome rather than searching thresholds on these exposed sources.

No origin classification is shipped. #349's editorial FULL-recall, accepted
finding, and suggestion-safety gates are unrelated and unchanged.
