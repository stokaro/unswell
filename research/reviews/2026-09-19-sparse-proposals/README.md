# A larger lexical vocabulary still leaves most defects unselected

Increasing the allowed feature count from 128 to 8,192 raises positive-unit
selection from 74/872 to 119/872 (8.5% to 13.6%). It also raises explicit-control
selections from 12/1,234 to 21/1,234. On Ptah specifically, positive selections
rise from 39/347 to 46/347. The capacity limit explains some misses, but removing
it does not produce a useful general editing diagnostic.

This study extends the [compact comparison](../2026-09-19-learned-proposals/README.md)
using exactly the same exposed assistant labels and grouped splits. The
[protocol](protocol.md) and [plan](split-plan.json) were fixed before the first
real-data fit. It introduces no new confirmation, human labels, product model,
profile changes or gate changes. The broad recall objective remains unmet.

## Fixed comparison

All three variants combine the existing 36 numeric and missingness columns with
training-only binary word n-grams. A term needs three training source groups.
The 8,192-column allowance therefore selects only 2,099–2,646 eligible columns,
depending on the fold. It does not actually fit 8,192 terms on this small corpus.

A separate development fold chooses a threshold maximizing positive selections
at explicit-control FPR at most 1% and labeled precision at least 85%. All five
folds find an operating point. These constraints do not guarantee the same rates
in the excluded evaluation groups. The comparison retains every evaluation
unit, including units without a quality label.

| Allowance | Actual columns | Positive units selected / 872 | Clean controls selected / 1,234 | Labeled precision | Unlabeled selections / 27,644 | Event targets retrieved / 994 |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| SW128 | 128 | 74 (8.5%) | 12 (0.97%) | 86.0% | 100 | 81 |
| SW1024 | 1,024 | 112 (12.8%) | 25 (2.03%) | 81.8% | 154 | 129 |
| SW8192 | 2,099–2,646 | 119 (13.6%) | 21 (1.70%) | 85.0% | 148 | 134 |

The wider models exceed the 1% control-FPR target on evaluation groups. For
SW8192, historical pages contribute 73/525 positive selections and 17/533 control
selections (3.19%); Ptah contributes 46/347 and 4/701 (0.57%). The
[summary](summary.json) retains fold, cohort and length-band denominators.
These finite, deliberately reviewed controls do not estimate FPR across all
production prose. Unknown units are not silently counted as clean examples.

Event retrieval requires selection of every prepared piece containing an event's
eligible target. It is not a supported diagnosis of that event. Of SW8192's 134
retrieved targets, 24 were already diagnosed by the retained rule evaluation
and 110 had been missed. The [rule comparison](rule-reference.json) preserves
144/994 prior full-event diagnostic credits. The generic model adds zero
supported diagnoses to that total; it neither identifies the defect nor explains
an edit that preserves the surrounding technical meaning.

## Numerical compatibility and resources

The research trainer uses deterministic sparse L-BFGS with the same normalized
logistic objective and L2=0.01 as the dense trainer. It leaves the product's dense
model-pack limit unchanged. Both training and scoring run in Go without cgo.

The [optimizer comparison](optimizer-comparison.json) compares SW128 with the
retained dense SW fits before attributing changes to capacity. Features and
evaluation units match in every fold. Maximum raw-score difference is
0.00004172; maximum weight difference is 0.000004462. No selected unit changes.
Normalization differences are at floating-point rounding scale. Both the old
and new operating points are retained, including their numerical thresholds.

All 15 fits converge within 49–61 iterations and below the fixed 1e-8 gradient
limit. Each accounts for 11.5–56.4 million optimizer operations, below its fixed
three-billion limit. The complete run took 9.86 seconds on the recorded
macOS/arm64 host. This observation includes preparation, fitting and output;
it does not measure peak RSS or qualify product performance. The implementation
also bounds rows and nonzero entries. Oversized, nonfinite, unconverged or
canceled fits return errors without a partial model.

A second execution reproduced every fitted parameter and all 89,250 evaluation
rows byte for byte. Blackbox tests compare against the existing dense numerical
reference, cover constant large-valued columns and variance underflow, and show
that evaluation-label changes cannot affect training or threshold selection.
Evidence tests reject capacity violations, unconverged models, missing fits,
changed feature identities and altered optimizer settings.

## Decision and next boundary

Keep these fits as research baselines. The larger lexical vector still misses
most reviewed targets and cannot support a source-specific diagnosis. Do not
ship it, lower the gate or select another threshold after observing evaluation
errors. The same 157 exposed pages, short-template overlap, deliberately selected
controls and prepared-piece context limits remain from the preceding study.

The next candidate must connect text to surrounding propositions and distinguish
redundant framing from technical constraints. Before adding more capacity or
another optimizer, it needs an explicit explanatory task and source-bound
positive/control examples. Useful unit retrieval alone cannot satisfy that task.
Any product candidate then needs fresh complete-page confirmation with labels
frozen before its outputs are inspected, under ADR 0041.

## Reproduction

The [measurement record](measurement-record.json) binds the executable, runtime
base, frozen input, protocol, splits and trainer source archive. The exact Go
sources are in [trainer-source.tar.gz](trainer-source.tar.gz). Rebuild on that
runtime base with those sources; later feature changes must not overwrite this
measurement. The Python scripts only assemble and audit existing artifacts.

```sh
cd research/annotation
CGO_ENABLED=0 go build -o /tmp/unswell-sparsereview ./cmd/sparsereview
cd ../..
PYTHONDONTWRITEBYTECODE=1 python3 research/reviews/2026-09-19-sparse-proposals/tools/check_sparse.py \
  --rerun /tmp/unswell-sparsereview
PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover \
  -s research/reviews/2026-09-19-sparse-proposals/tools -p 'test_*.py'
```

The verifier reconstructs the same source-bound input, validates all models,
recalculates the summaries and compares an optional rerun byte for byte.
[validation.json](validation.json) records completed checks separately from
pending acceptance. Retained raw scores are not calibrated probabilities.
