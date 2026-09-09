# ADR 0020: Share logistic training and inference in Go

Status: accepted for implementation.

## Decision

Add the pure-Go `model` package for immutable numerical logistic models and
deterministic training. Both research tools and later engine integration will use
this implementation. It accepts ordered numeric vectors and binary labels; it
does not extract prose, select editorial labels, assign data partitions, perform
I/O, or alter current engine/configuration/report contracts.

This implements the numerical baseline required by #23. The corpus/feature join
must supply only permitted, resolved editorial targets, verify identical units
and context, select compatible features, and isolate training from development,
calibration, and final test. Feature order, missing-value treatment, NLP/policy
identity, domain, unit kind, rubric, and source grouping remain requirements for
the model pack and integration in #23, #24, #56, and #57. A numerical parameter
snapshot cannot establish those contracts or qualify a product model.

## Objective and optimizer

Training uses binary labels 0/1 with both classes present. Each feature is centered
and scaled using its training population mean and standard deviation. Welford's
updates run in input order; a constant feature uses scale 1. Missing values are
rejected, including NaN. Imputation, if adopted, needs a separate explicit feature
contract and training protocol; it cannot silently turn absent evidence into zero.

For normalized vector `x`, `z = intercept + weights*x`. Minimize the mean binary
logistic loss plus `L2/2 * sum(weights^2)`. The intercept is not penalized. Evaluate
loss with a stable softplus formulation and response with a sign-stable sigmoid.
The response is the classifier's uncalibrated output, not a qualified probability
of editorial revision or machine origin.

Use deterministic full-batch damped Newton updates. Start at zero weights and
the observed training class log-odds. Build the analytic gradient and Hessian in
fixed row/column order, solve the positive-definite system with Cholesky, and use
Armijo backtracking. If the predicted decrease is within 16 upward ULPs of the
current loss, accept a step within that loss band also when it reduces the
gradient infinity norm by more than half; loss differences
alone cannot resolve progress near a stationary point. This exception does not
relax convergence. Return a model only when the gradient infinity norm meets
the requested tolerance. Iteration exhaustion, numerical failure, cancellation,
or work-budget exhaustion returns an error and no partial model.

This choice avoids a learning-rate search for the initial small dense baseline.
It costs quadratic work in feature count. Limit training to 128 features,
100,000 rows, and four million cells, plus explicit iteration and operation
budgets. These are computational limits, not evidence of scientific sample
adequacy. Dense high-dimensional n-gram training needs a separately evaluated
optimizer; do not expand this algorithm's limits to accommodate it silently.

No randomness or worker-dependent reduction is used. The input order is part of
the numerical protocol, and an input hash binds values, labels, and dimensions.
Record optimizer version and complete options with the result. Identical ordered
input on a pinned environment must reproduce. Cross-platform floating-point
identity requires separate evidence; tests use explicit tolerances for known
numeric answers.

## Inference and remaining acceptance

Models own their parameters and expose copies. Concurrent inference uses only
invocation-local state. Return normalized feature contributions to the linear
score, the intercept, total linear score, and uncalibrated response. These
contributions explain the computation, not causality or percentage-point changes.
Reject mismatched dimensions and nonfinite input/intermediate results.

Numerical fixtures test the objective derivatives, intercept solution, learning,
regularization, constant features, normalization, deterministic reproduction,
ownership, cancellation, convergence/resource failure, and extreme logits.
They do not create human labels or measure model quality. Separate calibration,
artifact loading/compatibility, feature joining, held-out evaluation, and product
gating remain open; `calibration.model: none` retains its existing behavior.
