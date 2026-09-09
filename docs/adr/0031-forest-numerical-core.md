# ADR 0031: Add a bounded forest for the nonlinear baseline

Status: accepted for numerical experiments; qualification remains open.

## Decision

Implement #50's small tree ensemble in the existing `model` package. It consumes
the same ordered numeric vectors as the logistic baseline. Corpus selection,
source preparation, feature identities, splits, and permissions remain with the
existing research pipeline. No second prose analyzer is introduced.

[Breiman's random forests paper](https://www.stat.berkeley.edu/~breiman/randomforest2001.pdf)
motivates combining randomized trees. This implementation is a specified binary
classification variant, not a port of a particular library or a reproduction of
the paper's experiments. It uses bootstrap samples, random feature subsets,
binary Gini splits, and the mean of leaf class fractions. Leaf fractions are
uncalibrated numerical responses. They are not qualified editorial probabilities.

The algorithm records its seed and limits. Trees grow in preorder, split ties use
feature order then threshold order, and inference returns each traversed path and
leaf response. Equal-impurity splits are allowed within the depth/leaf limits so
that a feature interaction such as XOR can be represented. No corpus-level
importance is presented as an explanation of one prediction.

Training rejects missing/nonfinite vectors and bounds rows, features, cells,
depth, trees, nodes, and logical operations. Exhausted resource limits return no
partial model. Bootstrap draws and feature selection use a private PCG generator;
concurrent fits do not share mutable state. Disabling bootstrap is an explicit
control for mathematical fixtures and comparisons.

The numerical snapshot is immutable and contains only features, nodes, thresholds,
and class counts. Loading validates its complete tree structure before inference.
It cannot execute code, fetch resources, or select a gate policy. Calibration and
the research artifact must distinguish forest responses from logistic linear
scores; neither kind may be silently reinterpreted as the other.

## Acceptance

Blackbox controls must verify hand-computed splits, feature interactions,
bootstrap determinism, missing values, corruption, cancellation, resource limits,
owned snapshots, and individual tree paths. Corpus integration must retain split
isolation, use the existing train/predict/evaluate commands, and compare variants
through #57. Numeric labels in tests do not meet #22's human annotation requirement.
The model remains experimental until grouped evaluation, calibration, applicability,
and the algorithm comparison satisfy the research acceptance protocol.

## Research artifact integration

The estimator is explicit in training v3. Exactly one options/model pair is
allowed. Restoring a forest validates tree structure, input width, training-row
counts, algorithm identity, and recorded work/depth/leaf/node limits. Logistic
remains the default. The command rejects flags belonging to the other estimator.
No earlier alpha format is accepted through a compatibility path.

Prediction v2 distinguishes logistic linear scores and forest mean responses.
Calibration declares which channel its separate knots consume. Missing target
features and calibration support retain explicit absence; no zero is substituted.
The existing comparison protocol already permits different classifier responses
while requiring identical target, corpus, context, and preprocessing identities.
Blackbox and compiled-command fixtures exercise all three feature representations.
