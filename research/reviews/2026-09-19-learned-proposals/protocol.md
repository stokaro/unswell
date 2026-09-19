# Development comparison of learned unit proposals

Written before fitting this candidate or reading its predictions. This is a
development experiment on the exposed assistant judgments preserved by the
representation audit. It does not qualify a product model or supply fresh
confirmation. The ultimate task remains useful source-localized editorial
diagnostics with good whole-page recall, not merely high unit classification
accuracy.

## Input and grouping

Use all 157 complete pages from `2026-09-19-representation-audit`. Keep its
872 positive and 1,234 explicit-control negative units, and all 27,644 unlabeled
units. Labels identify units containing a complete reviewed edit target, not
the precise cause of the defect. No unmarked or uncertain text becomes clean.
No origins, reviewer names, repository IDs or labels are model features.

Keep all units from the same source in one group. Merge source groups when
their prepared units share an exact text hash and contain at least twelve
prose words. This prevents reuse of longer exact passages across folds; it
does not prove absence of short or paraphrased overlap. Publish the resulting
groups and short duplicate counts. Grouping uses all source identities before
training, but no learned vocabulary, normalization or target-dependent distance.

Sort connected groups by SHA256(`review-proposals-v1\n` + canonical source ID),
then assign them round-robin to five folds. For outer fold i, fit on the three
folds other than i and (i+1)%5, choose an operating threshold on (i+1)%5, and
evaluate on i. Every page is evaluated once. These are internal development
folds, not an independent final test. Report Ptah and historical cohorts and
the 1–9, 10–39 and 40+ word bands separately.

## Fixed models

Use the current Go `model.FitLogistic` implementation, L2=0.01, gradient
tolerance=1e-6, at most 100 Newton iterations and 1e12 logical operations.
Standardization is fitted on training rows only. Preserve missing numeric
features using a zero placeholder and a separate missingness indicator. A
placeholder alone is not a measured zero.

Compare four fixed representations:

- L: log(1 + eligible prose words).
- S: all current shared descriptive unit measurements and missingness flags.
- W: L plus up to 127 shared word n-gram columns, orders 1–3.
- SW: S plus enough word columns to stay within the existing 128-feature limit.

Word counts come from `feature.CountLexical` over the same prepared units, with
protected boundaries preserved. Use binary presence, no vocabulary rewriting,
stop-word removal or test-data IDF. Select terms present in at least three
training source groups by training-only chi-square association with the binary
label; break equal scores by the exact key. Keep feature and example order
deterministic. Do not fit on the same page that is being evaluated.

On the development fold, select the score threshold with the greatest positive
unit recall subject to explicit-control FPR <=1% and precision >=85%. Include
all tied scores together. No qualifying nonempty threshold means
`no_operating_point`, not perfect precision or a calibrated zero risk. These
constraints apply only to the explicitly labeled development units; they do not
establish FPR on the full stream. Raw linear scores are never percentages.

## Outputs and decisions

Retain every evaluated unit's raw score and proposed selection, including
unlabeled units, plus fitted coefficients, feature identities, training input
hashes, split manifests and operating-point decisions. Report labeled-unit
metrics, unlabeled review burden and the number of complete annotated event
targets whose units were selected. The latter is retrieval coverage, not
full-event diagnostic recall: a generic unit proposal does not explain the edit.

Compare against the current rules without giving overlap with a long-sentence
warning credit for a different annotated defect. Preserve the original
source-bound full-event judgments. Also compare W/S/SW against L so that length
or fragmented controls cannot masquerade as linguistic progress.

Do not ship a classifier from these measurements alone. A plausible candidate
must next show source-specific explanations and pass a new complete-page
confirmation sample frozen before tuning or inspecting its outputs. If the
representations fail, retain the result and investigate richer context rather
than lowering the product gate or relabeling misses. All fitting and feature
computation in this experiment run locally in Go; Python only assembles input
records and verifies or summarizes saved outputs.
