# Sparse capacity comparison on exposed review units

Freeze this protocol before fitting or inspecting the expanded models. The
preceding experiment's SW model uses at most 128 columns and retrieves 74/872
positive units. ADR 0030 explicitly states that this dense-optimizer limit is
not evidence that 128 lexical terms suffice. Test that constraint before
concluding that lexical evidence itself is exhausted.

Use exactly the source-bound input and five grouped splits retained in
`2026-09-19-learned-proposals`. No new labels, sources, controls, reviewed edits,
group definitions or thresholds enter this experiment. All pages are exposed
assistant development data, not fresh confirmation or human annotation.

Compare SW128, SW1024 and SW8192. Each uses the same shared numeric features,
missingness flags and binary word n-grams as SW. Terms must appear in at least
three training source groups; rank by the same training-only chi-square score
with the same deterministic tie break. Capacities include all numeric columns.
Report the actual selected vocabulary size if fewer terms are eligible.

The sparse Go optimizer minimizes the same average logistic loss and L2=0.01
penalty on standardized slopes. Training rows alone set means and population
standard deviations; constant columns have scale one. The intercept remains
unpenalized. Sparse centering accounts for omitted zeros algebraically.
Use deterministic L-BFGS with ten retained updates, Armijo coefficient 1e-4,
unit initial step and up to forty halvings per line search. Stop at maximum
absolute gradient <=1e-8 or fail after 500 iterations. Never return a partial
model on convergence, numerical, cancellation or resource errors.

The tolerance was tightened from 1e-6 before any real-data fit: a synthetic
correlated-column comparison showed a 0.000163 score difference at that older
tolerance. Keep the preceding dense fit unchanged and report remaining drift.

Limits: 8,192 columns, 100,000 training rows, four million nonzero entries,
and three billion accounted optimizer operations per fit. Bounds do not claim
measured peak RSS. The wider research artifact does not relax the public dense
model pack contract or add a product runtime dependency.
Numeric entries must be finite and within -1e12..1e12, as in the dense trainer;
variance underflow on a varying column is an error, not a constant feature.

Before evaluating the real data, blackbox tests must compare sparse scores with
the existing dense Newton fit on small numerical fixtures. The actual SW128
fit must also be compared with the preceding SW outputs: report coefficient
and score differences and any changed selections near a threshold. Do not
silently attribute optimizer drift to increased vocabulary capacity.

Keep the operating-point protocol unchanged: select on the separate development
fold with explicit-control FPR <=1% and labeled precision >=85%, maximizing
positive-unit recall and keeping score ties together. No qualifying nonempty
point is explicit abstention. Evaluate each held-out development group once,
retaining all unlabeled outputs. Do not tune capacity or thresholds on those
results and call the resulting selection fresh confirmation.

Report the preceding metrics and source-bound event retrieval for every model,
with fold, cohort and length-band results, actual feature counts, runtime and
operation counts. Preserve the earlier measured full-event rule credit.
Selection remains a proposal, not an explained editorial diagnosis or a
calibrated probability. Source-specific defect explanations and new frozen
complete-page confirmation are still required before product adoption.
