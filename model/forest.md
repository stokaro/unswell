# Numerical forest baseline

`FitForest` supplies the nonlinear candidate in #50 over the same ordered
`[]Example` vectors as the logistic model. It does not select corpus rows,
extract text, establish permissions, or qualify editorial probabilities.
See [ADR 0031](../docs/adr/0031-forest-numerical-core.md).

The algorithm is `unswell-binary-forest-v1`. It fits binary CART-style trees with
Gini splits and averages the positive-class fraction in the reached leaves.
This averaging differs from assigning every tree a hard class vote. Missing
values are rejected. Raw values are not standardized; monotone feature scaling
is unnecessary for these threshold comparisons. Both classes must occur in the
supplied training data; a bootstrap sample may contain only one class.

## Fitting protocol

A fit owns a Go `math/rand/v2` PCG generator initialized with the supplied seed
and second seed `0x9e3779b97f4a7c15`. Trees are built sequentially in preorder.
With bootstrap enabled, each tree draws `n` row indices with replacement through
`IntN(n)`. Without bootstrap, each tree starts with the ordered training rows.

At each eligible node, `FeaturesPerSplit` columns are selected without replacement
by a partial Fisher–Yates shuffle. Zero means `floor(sqrt(width))`. Selecting all
columns consumes no random draws. The selected columns are sorted before split
search. There is no fallback to columns outside the selected subset.

For each column, sort rows by value and consider boundaries between distinct
values whose children both contain at least `MinLeaf` samples. Bootstrap copies
count as separate samples. The binary Gini reduction is:

```text
2 * (left_positive * right_n - right_positive * left_n)^2
  / ((left_n + right_n)^2 * left_n * right_n)
```

Choose the largest reduction; ties retain the first column, then the lowest
boundary. A zero reduction is permitted, allowing later splits to identify
feature interactions. The threshold is `a + (b-a)/2`; if rounding reaches `b`,
use `a`. The left comparison is inclusive. Children preserve sample order.

Stop at pure nodes, the depth limit, insufficient samples for two leaves, or no
eligible boundary in the selected columns. No pruning, class weighting,
out-of-bag estimates, or automatic hyperparameter tuning is performed.

## Resource and identity contract

Inputs use the logistic baseline's limits: 2–100,000 rows, 1–128 features,
4,000,000 cells, and finite values within `-1e12..1e12`. Forest limits are 64 trees,
12 split levels per path, and 65,535 total nodes. The caller supplies a smaller
`MaxNodes` if needed. Hitting that allocation budget fails the whole fit; it does
not silently truncate trees. `MinLeaf` is a complexity parameter, not a measure
of sufficient editorial evidence.

`MaxOperations` is a logical budget within 1–`1e12`, not a CPU-instruction or time
measurement. Preparation charges `n*(width+1)`. Sampling charges `n` per tree;
each node charges its sample count. An eligible node charges `width` for column
selection, and each selected column charges `n*(4*bits.Len(n)+4)` for sorting and
scanning. Partitioning charges the node's sample count. Final snapshot validation
and copying charge twice the total node count. Cancellation is checked while
preparing, sampling, counting and partitioning rows, and around each bounded sort.

The input hash covers `unswell-forest-input-v1` and NUL, little-endian uint64 row
and feature counts, then each row's uint64 label and IEEE-754 float64 value bits.
Options, including seed, remain separate recorded inputs. Reproducibility requires
the same ordered input, options, algorithm and Go toolchain. A changed seed need
not change every learned tree or prediction.

## Inference and explanation

`NewForest` accepts only complete preorder trees. It rejects invalid feature or
child indices, cycles, shared or unreachable nodes, excess depth, nonfinite
thresholds, and inconsistent parent/child counts. Leaves have no children or
split threshold. It copies the supplied snapshot; `Parameters` returns another
owned copy. No file loading, network access, or executable payload is supported.

`Evaluate` returns each path, leaf, sample counts and class fraction, followed by
their mean. These paths explain the comparisons made for that input; they do not
establish causality or locate a prose defect. Concurrent calls own their result
buffers. All errors and cancellation return no partial model or evaluation.

Separate calibration must use an explicitly named forest response, not a logistic
linear score. Numerical controls and scripted labels test the implementation;
real grouped data, comparison with simple baselines, applicability, calibration,
resource measurements and editorial qualification remain required.
