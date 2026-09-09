# ADR 0025: Fit research models from reproduced corpus targets

Status: accepted for research training; statistical qualification remains open.

## Decision

Connect the existing numerical Go model package to the source-reproducing corpus
workflow. `research/annotation/training` receives a corpus artifact, a validated
round, pinned source and notice bytes, and explicit fitting options. It calls
`corpus.Join`; a caller-supplied saved feature file is not evidence of reproduction.
The `corpus train` developer command supplies these inputs through the existing
bounded source and round loaders.
`model.FitOptions.Validate` exposes the existing optimizer validation without
preparing rows; fitting and the research command share these numerical limits.

A run selects one target kind and an ordered set of existing prepared features.
The positive label is `needs_revision`, and the negative label is `acceptable`.
Unresolved, missing, and uncertain judgments do not become numeric labels. The
source, curator, generator, role, group, and partition metadata never enter the
feature vector. Rights declarations must include `training` for every resolved
target considered for either classifier or calibration fitting. The operation
checks declarations; it does not establish legal permission or human participation.

The classifier uses only the frozen `training` partition. Normalization, labels,
and weights are supplied to `model.FitLogistic` in stable unit-ID order. Optional
isotonic calibration evaluates the frozen classifier's linear scores only on the
separate `calibration` partition and calls `model.FitIsotonic`. Classifier fitting
never consumes calibration rows. `development` and `final_test` remain reserved;
the command does not predict them, publish their labels, or calculate their errors.
Full-corpus verification and target reproduction still check their source integrity.

Feature identity and compatibility include the prepared unit contract, target
kind, descriptors, actual NLP/capabilities, engine policy, extraction/preparation,
and vocabulary hashes. The annotation instructions retain a separate profile hash.
One run does not combine different kinds, policies, NLP representations, or feature
definitions. Missing features either fail the requested run or exclude a row under
an explicit `exclude` policy. They are never replaced with zero. Exclusion counts
and selected row references are recorded so coverage remains visible.

Tutorial rounds require an explicit simulation option. Every output remains an
experimental numerical artifact, including when a pilot or corpus declares human
raters. Neither successful fitting nor a sigmoid response qualifies revision
probabilities. The artifact records the original round/corpus/join hashes, options,
split counts, selected unit/group/input identities, numerical parameters, training
input hashes, convergence evidence, and optional calibration knots. It omits prose,
rationales, and raw feature vectors. Errors return no partial model artifact.

## Scope and follow-up

This supplies the data path for the regularized Go logistic model and its separate
calibration. It does not select hyperparameters, qualify data, run final evaluation,
load a product model pack, change `calibration.model: none`, or alter normal scan
results. There is no random optimizer seed: the existing numerical algorithm is
deterministic in row order. The source split seed stays in the corpus manifest.

The rule-activation logistic baseline and controlled A–G comparisons remain
required under #23 and #57. Parent block activations must not be copied onto
sentences or fragments; this run accepts the existing prepared feature contract.
The experimental artifact cannot justify choosing a production model before those
comparisons, real independent annotation, calibration checks, and held-out evidence.

The existing corpus and numerical limits bound sources, targets, cells, features,
iterations, and work. A serialized training artifact has a separate 16 MiB limit.
The implementation stays local and pure Go; it neither runs reference models nor
downloads data. Fitting fixtures use explicitly simulated labels and prove only
the data path, source bindings, partition isolation, and numerical behavior.
