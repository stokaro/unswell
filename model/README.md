# Numerical logistic models

This package also supplies a [separate isotonic calibration candidate](calibration.md).
Selecting and qualifying an editorial calibration method requires corpus evidence.

This package supplies a shared Go implementation for the initial logistic baseline
in #23. It performs numerical training and inference over ordered vectors. It does
not extract prose, assign labels, select features, load files, or make CI decisions.
Research tools and later engine integration use the same objective and evaluator.

`FitLogistic` takes `[]Example` and explicit `FitOptions`. Targets must be 0 or 1,
with both classes present. All rows need the same feature count. Missing values,
NaN, and infinities are errors. The caller must establish source reproduction,
label permissions, independent partitions, feature ordering and applicability.
Only the training partition may contribute rows or normalization statistics.

Training computes population means and standard deviations in input order using
Welford accumulation. Scale is `sqrt(sum_squared_deviations)/sqrt(rows)`. Constant
features use scale 1; a nonconstant feature whose variance underflows fails.
No implicit imputation runs. A future missing-indicator policy needs an explicit
feature contract shared by training and inference.

For normalized input `x`, the linear score is `z = intercept + weights*x`.
The objective is mean binary logistic loss plus `L2/2 * sum(weights^2)`.
The intercept is unpenalized. Stable softplus and sigmoid calculations handle
large finite logits without overflowing the exponential.

`unswell-logistic-newton-v1` uses full-batch Newton steps, Cholesky solves, and
Armijo backtracking with coefficient `1e-4` and at most 30 halvings. Initialization
uses zero weights and training class log-odds for the intercept. Convergence
requires gradient infinity norm at or below `Tolerance`; exhausted iterations,
invalid numerical steps, cancellation, and work limits return no model.
When the predicted decrease `step * gradient·direction` is at most 16 upward ULPs
of the current loss, a step may also be accepted within that loss band if it cuts
the gradient infinity norm by more than half. This does not relax the final
gradient tolerance.

Limits are explicit:

| Input | Range |
| --- | --- |
| Features | 1–128 |
| Rows | 2–100,000 |
| Total cells | At most 4,000,000 |
| Training values | Finite, absolute value at most `1e12` |
| `L2` | `1e-8`–`1e3` |
| `Tolerance` | `1e-12`–`0.1` |
| `MaxIterations` | 1–1,000 |
| `MaxOperations` | 1–`1e12` |

These limits bound numerical work and allocations; they do not define sufficient
data for editorial judgments. Dense Hessians cost quadratic work in feature count.
Large n-gram models need a separately evaluated optimizer.

Operation accounting is a reproducible logical budget, not elapsed time or CPU
instructions. For `n` rows, `d` features and `p=d+1` parameters, preparation charges
`n*(4*d+1)`. An objective evaluation charges `n*2*p`; requesting the Hessian adds
`n*p*(p+1)/2`. Each linear solve charges `p^3+2*p^2`. Line-search evaluations are
charged separately. Budget checks precede each charged stage.

The training result records algorithm, complete options, input hash, iteration and
operation counts, final regularized loss, gradient norm, and immutable model.
There is no random initialization, shuffle, parallel reduction, or seed. Input
order is part of the protocol. The SHA-256 input hash covers the UTF-8 prefix
`unswell-logistic-input-v1` followed by NUL, then little-endian unsigned 64-bit row
and feature counts; each row contributes its label as uint64 and its values as
IEEE-754 float64 bits. Reordering rows changes the hash. The options remain a
separate recorded input. This hash does not authenticate editorial labels or
feature identities.

`NewLogistic` validates and copies a numerical `Parameters` snapshot. `Evaluate`
returns normalized values, individual linear contributions, intercept, total linear
score, and uncalibrated sigmoid response. Inputs must match the dimensions and
remain unchanged during a call. Parameters and returned slices are independent;
concurrent calls share no mutable evaluation buffers. Nonfinite intermediates fail.

The response is not a qualified revision probability. The snapshot is not a model
pack: feature/NLP/profile compatibility, rubric and unit applicability, separate
calibration, group-aware evaluation, and permissions remain required. Ordinary
Unswell results and gates continue their existing rules-only behavior. Numeric
fixtures with scripted labels verify the implementation; they cannot satisfy the
human corpus or held-out quality targets. See [ADR 0020](../docs/adr/0020-logistic-numerical-core.md).
