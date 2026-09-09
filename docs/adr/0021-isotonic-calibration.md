# ADR 0021: Add a separate Go isotonic calibration candidate

Status: accepted for implementation; product method selection remains experimental.

## Decision

Extend the shared `model` package with an immutable isotonic mapping and a Go
fitting function for classifier scores and binary labels. This supplies one
separate calibration candidate for #23. It does not select a production method,
qualify probabilities, join corpus records, or enable an engine probability gate.
Choose the method on the development protocol and fit it only on the separate
calibration partition after freezing the classifier.

Isotonic regression minimizes squared error subject to a nondecreasing fitted
response. The [pool-adjacent-violators algorithm][pav] supplies the ordered fit.
Implement the algorithm in Go without importing reference code or adding a
runtime dependency. A logistic calibration candidate can use the existing
one-feature logistic objective in a separate experiment.

## Numerical contract

Accept 1 through 100,000 finite scalar scores with labels 0 or 1. This is a
computational bound, not a claim about sufficient calibration data. Constant
scores or a single observed class have a defined numerical fit; downstream
qualification must evaluate sample adequacy and coverage separately.

Copy and sort samples by score. Aggregate equal scores before fitting, including
positive and negative zero as the same score. Counts and positive-label sums are
integers. Merge adjacent pools while their means fail to increase strictly;
cross-products use 64-bit integers so comparison is exact at the sample bound.
Merging equal means gives maximal constant pools. Store each distinct score and
its pool's mean response as a knot. Training parameters are independent of input
ordering; the recorded input hash still identifies the ordered original rows.

Prediction returns a knot's fitted response or linear interpolation between its
adjacent knots. Handle finite extreme scores without overflowing the interval
width. Scores outside the observed knot range return `ErrCalibrationRange`;
there is no implicit extrapolation or clipping. This numerical range is not a
validated domain or prose-length applicability rule.

The fitting result records algorithm version, input digest, sample/knot/pool
counts, and mean squared error on the calibration inputs. That fit error is not
a held-out Brier score or evidence that the method improves probabilities.
The digest uses SHA-256 over `unswell-isotonic-input-v1` and a NUL, followed by
little-endian uint64 sample count and, for each row, the original float64 score
bits and its uint64 label. Scores, labels, and row order are covered; classifier,
feature, rubric, group, and partition identities belong to the experiment pack.

Sorting costs O(n log n); pooling, storage, and expansion cost O(n). Validate
bounds before allocation. Check cancellation during input validation, grouping,
pooling, and expansion, and before and after the bounded sort. Failure returns
no partial fit. Restored parameters must be finite, dimensionally consistent,
strictly ordered in score, and nondecreasing within [0,1] in response.
Models own their parameters and permit concurrent inference.

## Acceptance boundaries

Use analytic pooling and tie cases, an independent small-input min-max oracle,
ordering/ownership checks, extreme interpolation, cancellation, invalid snapshots,
and an external consumer. Fixtures contain scripted numeric labels and establish
numerical behavior only. Compare calibration methods and their coverage on the
same frozen development protocol before choosing one; retain separate calibration
and final-test partitions. Integration still must bind the mapping to its frozen
classifier and compatible features, units, NLP, policy, rubric, and data rights.
No fitted response becomes a qualified revision probability merely because it
is between zero and one.

[pav]: https://docs.scipy.org/doc/scipy/reference/generated/scipy.optimize.isotonic_regression.html
