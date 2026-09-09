# Frozen research predictions and evaluation

This developer tool restores the existing Go logistic baselines, saves predictions
without evaluation labels, and computes metrics from those saved records. It does
not expose probabilities in normal Unswell scans or qualify a model. Issues #57,
#25, and the research umbrella #59 remain open.

The [training guide](../training/README.md) describes creating a numerical model.
Compile the same `corpus` command from this module with `CGO_ENABLED=0`. Ordinary
builds and tests use no Python, Node.js, model service, or GPU.

## Freeze a prediction plan

Store a JSON plan with version `unswell-research-predictions-v1` and these fields:

| Field | Meaning |
| --- | --- |
| `id` | Stable trial identifier, at most 128 bytes |
| `protocol_sha256` | SHA-256 of the exact retained protocol file bytes |
| `model_sha256` | Embedded `sha256` from the frozen training artifact |
| `corpus_sha256` | Embedded `sha256` from the frozen candidate artifact |
| `partition` | `development` or `final_test`; never training or calibration |
| `context` | `prepared_piece` for NLP over the enclosing eligible piece, `source_document` for rule activations |
| `response` | `logistic` for uncalibrated sigmoid response, or `isotonic` for the separately fitted mapping |
| `threshold` | Explicit number in [0, 1]; positive means response greater than or equal to this value |

All digests are 64 lowercase hexadecimal characters. Unknown fields, missing
thresholds, unsupported contexts, and incompatible model artifacts are errors.
The plan cannot request isotonic output from a model without that mapping.
No threshold search, normalization, model selection, or fitting runs during
prediction. A protocol change requires a different digest and trial record.

Commit the complete research run manifest and access record required by
[protocol v1](../../methods/protocol-v1.md) before opening final labels. This
compact executable plan does not replace that manifest. A hash records content;
it cannot establish when an operator saw labels or whether metadata is honest.
The shared preparation analyzes an eligible coherent piece before selecting
sentence targets. Prepared features describe the target, but its NLP can retain
context from that piece. This is not a target-only NLP experiment. Prepared and
source-document predictions cannot establish a controlled comparison until their
available context agrees with the declared comparison protocol.

From the research module, after preparing the named inputs:

```sh
corpus predict --root sources --model model.json --plan trial.json \
  --protocol protocol.md < candidates.json > predictions.json

corpus evaluate --corpus candidates.json --round evaluation-round.json \
  < predictions.json > metrics.json
```

The rule baseline adds `--rule-config rules.yaml` to `predict`. Its bytes and
effective feature/policy identities must match training. Prediction accepts no
`--round`; evaluation accepts no source root or model path. Tutorial evaluation
requires `--allow-simulation` and reports `basis: simulation`. Unknown authorship
metadata never supplies an editorial or confirmed-human label.

## Saved evidence and missing results

Predictions retain the plan and its digest, the complete frozen numerical model,
the source/candidate verification with producer and verifier build identities,
and one row for every selected target. Rows contain source/group/unit IDs, the
feature-input hash, raw linear score, uncalibrated response, selected response,
threshold decision, and applicability. They contain no source text, feature
vectors, per-unit training labels, or evaluation labels.

A partial block, missing feature, or score outside the calibration range produces
an inapplicable row. Its selected response and threshold decision are `null`.
Out-of-range calibration retains the independently available raw and logistic
scores. Corrupt models, incompatible identities, source mismatches, missing
permissions, and cancellation fail the command with no partial JSON result.

Evaluation verifies the corpus and exact target identities against the independent
round, requires the model's rubric/profile, and retains unannotated or unresolved
labels as counted exclusions. It summarizes every eligible labeled row, including
abstentions. The saved prediction digest binds all numerical inputs; the decision
and round digests bind evaluation labels. Changed labels can change metrics but
cannot change already saved predictions or cause model execution.

## Metric definitions

The primary counts distinguish eligible targets, covered targets, independent
source groups, TP/FP/TN/FN on covered targets, and abstained positives/negatives.
The acquisition count also includes missing or uncertain labels. A missing class
or denominator yields `null`, including FPR on a positive-only set.

- Coverage is covered divided by eligible labeled targets.
- Precision, recall, and FPR use the covered subset and its displayed counts.
- Full-flow recall also counts abstained positives as undetected.
- Binary-event Brier is mean squared error against `needs_revision` on covered
  targets. The constant comparison uses fitted training prevalence on the same
  targets; it never estimates that constant from evaluation labels.
- ECE uses ten fixed equal-width response bins, left-inclusive and right-exclusive
  except the final bin includes 1. Each term is the bin's sample share times the
  absolute difference between mean response and observed positive rate. Empty bins
  have null rates. This is event-probability calibration, not top-class confidence.
- The same counts and metrics are retained for each source group, in sorted order.
  These group results are not confidence intervals or automatic independence proof.

The response kind remains explicit. A Brier/ECE calculation on uncalibrated
sigmoid outputs does not make them calibrated product probabilities. Both
artifacts retain `probability_status: unavailable_unqualified_model`.

The root e2e test runs plan, extraction, training, prediction, and independent
evaluation for the prepared and learned-rule baselines. It removes the external
model file before evaluation, checks deterministic output, rejects evaluation
labels on the prediction command, and changes only final labels to test separation.
Unit tests use hand-computed metric examples and numerical/absence boundary cases.
All tutorial observations and raters are explicitly simulated.

## Remaining acceptance

The executable path covers the current prepared-feature and rule-activation
logistic fits. It does not claim the full A–G comparison or a qualified D feature
set. Group-macro comparisons, paired bootstrap intervals, qualified FPR bounds,
predeclared subgroups, figures, lexical models, ablations, LLMDet/reference adapters,
and measured resource accounting retain their existing research requirements.
Unexecuted methods remain unverified in the method registry. Human acquisition,
independent annotation, final-test custody, calibration selection, and scientific
qualification require real evidence; these simulated tests cannot supply it.
