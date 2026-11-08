# Heterogeneous AI spans: prospective lexical transfer screen

Status: frozen before feature scoring. This is an offline, model-free origin
screen, separate from the editorial acceptance gates in issue #349. A positive
result would justify further testing, not an authorship claim or a product
finding.

## Input and identity

Use only the `all_layout_neutral` Parquet train, validation, and test files from
`open-text-detector/heterogeneous-ai-spans` tag `v1.4.0`, commit
`c8a59b129c34cfd9c67f739ae049abfff26fc2d3`. The SHA-256 of each file is
sealed in `freeze.json`. The dataset supplies one human control and one mixed
document per source, with human and generated spans. No raw text is committed.
The source's `human_verified=false` field means historical human provenance is
a documented candidate, not an independent certification. The corpus is
synthetic and its generated prose need not preserve the original facts.

Before scoring, require the exact file hashes, 7,018 train rows, 594 validation
rows, and 1,188 test rows; exactly two kinds per source; stable source text and
source metadata within each pair; unique IDs; group-disjoint splits; matching
raw-text hashes; complete, nonoverlapping character-span partitions; retained
human spans equal to their original source slices; and nonempty AI replacement
spans. Abort on a failed identity or reconstruction check. The validation split
is an audit only; it does not choose a feature, threshold, or direction.

## Frozen features and direction

Import the unchanged `features` function from
`research/origin/runs/2026-11-07-richness-screen/run.py`, pinned by SHA-256 in
`freeze.json`. Test lexical richness and normalized entropy separately. Higher
values mean more AI-like in both cases, following the prior technical and
academic screens. Neither feature is a probability. Do not flip direction or
fit a score combination after looking at the test results.

The primary **span** unit is one AI replacement and its exact original human
source block. Slice using the dataset's Unicode-code-point offsets, then use
the first 100 prose word tokens from each side. Both sides must have at least
100 such tokens; otherwise the pair abstains and remains in the denominator.
The 100-token fixed window controls length. These are **oracle-located** spans:
the score does not locate AI text in an unknown document.

The secondary **document** unit is the complete mixed document against its
matched complete human control. Keep the layout-neutral text intact. A document
abstains if the feature implementation returns no score. Report all eligible
and ineligible denominators.

## Split, calibration, and outcomes

Only train human text sets thresholds. For the primary span endpoint, take the
first eligible original block per train source, ordered by source start, source
end, and replacement ID. This prevents one source with two replacements from
contributing two calibration observations. For the document endpoint, take all
eligible train human controls. Set the threshold to the value at descending
rank `ceil(0.01 * N)`; flag only scores **strictly greater** than it. This
guarantees at most 1% empirical calibration-human FPR including ties. Record
the number of distinct calibration groups. No generated train row is scored.

On the untouched test split, report for each endpoint and feature: eligible
pairs, source groups, pooled AUROC, matched-pair win fraction, TPR at the fixed
threshold, calibration-human FPR, and observed test-human FPR. Report primary
span AUROC and TPR by source dataset and requested generator; report the same
document breakdowns where denominators allow. Matched observations share
sources, so these are descriptive scores, not independent confidence bounds.
Do not use validation or test results to alter the rule.

The registered primary richness feasibility rule requires all of:

1. At least 75% of all test AI replacement spans have both 100-token windows.
2. Pooled span AUROC at least 0.65 and TPR at least 10%.
3. Observed test-human FPR no greater than 2% on eligible matched source blocks.
4. Span AUROC greater than 0.60 in at least three of the four test source
   datasets with at least 40 eligible pairs each.

The secondary document richness rule uses at least 95% pair coverage, AUROC at
least 0.65, TPR at least 10%, and observed human FPR no greater than 2%.
Entropy has no admission rule and cannot rescue a failed richness rule.
Results from this dataset are an independent transfer screen for the unchanged
feature, but cannot alone qualify real-world or editorial detection.

The runner and this protocol are pinned by `freeze.json` before any feature
score is computed. One planned run writes aggregate JSON. Failures stop the
run; any correction requires a separately documented amendment and a new
freeze. No model calls, generated text, or private annotations are used.
