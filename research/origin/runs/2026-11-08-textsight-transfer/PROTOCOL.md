# Fixed-threshold origin transfer screen

Date: November 8, 2026. This is an exposed research screen for non-generative
origin features, separate from the editorial gates in Unswell issue #349.
The comparison uses the unchanged five-feature ARB logistic fit and the unchanged
richness-only score from `2026-11-08-interpretable-ensemble`. No model calls,
fitting, threshold selection, or product changes are permitted in this screen.

The external input is TextSight's `ai-text-detection-benchmark-2026` dataset at
Hugging Face revision `423b1768f0827675295fd6dc3cd255ad2f411183`.
Use exactly its four data files: 155 pre-2022 Wikipedia human passages, 183
2015–2021 PMC human passages, 306 pre-1940 literature control passages, and
56 September 2026 generated passages from five named models. Wikipedia and
PMC form the primary human comparison. Literature is reported separately as
an intentionally distant negative control. The dataset is not source paired,
so no paired win rate or causal authorship claim is permitted.

Score the first 100 extracted prose words with the existing feature code.
Fewer than 100 words means abstention and remains in the coverage denominator.
Use the fixed ARB human-calibrated strict `>` thresholds: ensemble
`1.3251545797561963`; richness `0.9199999999999999`. Report eligible and
flagged counts for every file and each generator, primary human FPR, AI TPR,
and unpaired AUROC among eligible primary human and AI texts. Do not calculate
an alternative threshold on TextSight. Preserve source SHA-256 and file counts.

The feasibility screen passes a score only if at least 80% of the 338 primary
human and 56 AI passages are eligible, primary human FPR is at most 1%, each
primary human source FPR is at most 2%, and overall AI TPR is at least 10%.
Report every generator separately; the small generator strata cannot establish
robust per-model detection. A pass permits a new source-matched, prompt-held-out
test, not deployment or an authorship claim. A failure retires this exact
threshold on this corpus; no cutoff fitting on the exposed test is allowed.

Check data identities and required counts before scoring. On changed input,
missing field, nonfinite score, or source mismatch, stop without a result and
do not repair individual rows. The dataset publisher sells detection software;
its provenance is reported as provided, not independently established here.
