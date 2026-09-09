# Train an experimental editorial model

`corpus train` connects reproduced annotation targets to the existing Go logistic
optimizer and optional isotonic calibration. It uses only resolved `acceptable`
and `needs_revision` labels for the selected kind. The normal Unswell scan still
reports revision probability as unavailable until a model passes qualification.

Run the developer commands from `research/annotation`:

```sh
go run ./cmd/corpus plan < training/testdata/manifest.json > plan.json
go run ./cmd/corpus extract --root training/testdata/sources \
  < plan.json > candidates.json
go run ./cmd/corpus train --root training/testdata/sources \
  --round training/testdata/round.json --kind paragraph \
  --feature prose-words --calibration isotonic --allow-simulation \
  < candidates.json > training.json
```

This example uses six scripted documents and two simulated raters. It checks the
mechanics of fitting and partition isolation. It supplies no human labels or
statistical quality evidence. The compiled-command e2e test runs these steps.

## Selection and partitions

The command reopens bounded local files, verifies their pinned bytes and notices,
reproduces corpus targets, and matches the annotation round before fitting. It
measures selected features through the public engine. An edited saved feature
vector cannot replace source reproduction.

Each run selects exactly one `--kind`: `sentence`, `paragraph`, or `fragment`.
Repeat `--feature` to choose a set of existing prepared feature IDs. The artifact
records them in canonical order. Parent labels and block rule activations are
not inherited by smaller targets. Source origin, author, generator, and group
metadata do not enter the numerical vector.

| Partition | Use |
| --- | --- |
| `training` | Fit normalization, coefficients, and intercept |
| `development` | Reserved; no fitting, predictions, or class statistics |
| `calibration` | Fit a separate mapping of frozen linear scores when explicitly selected |
| `final_test` | Reserved; no fitting, predictions, or class statistics |

Reproduction verifies source integrity for every partition. Reserved targets
remain outside the numerical algorithms. When calibration is `none`, that
partition is also reserved. Every resolved target considered for either fit must
declare a `training` use; declarations do not establish permission independently.
Missing, uncertain, and unresolved labels remain excluded, with recorded reasons.

## Explicit fitting options

| Option | Default | Meaning |
| --- | --- | --- |
| `--kind` | required | One prepared target kind |
| `--feature` | required | Repeatable set of existing prepared IDs |
| `--missing-features` | `reject` | Fail on an unavailable feature; `exclude` records and omits the row |
| `--calibration` | `none` | `isotonic` fits separate linear-score knots |
| `--allow-simulation` | false | Permit a round that explicitly declares simulated raters |
| `--l2` | 1 | Weight regularization; the intercept is unpenalized |
| `--tolerance` | 1e-8 | Convergence bound on the gradient infinity norm |
| `--max-iterations` | 100 | Newton iteration limit |
| `--max-operations` | 1000000000 | Numerical fitting work limit |

Defaults are developer settings, not measured quality thresholds. The
[numerical model contract](../../../model/README.md) defines loss, normalization,
optimization, calibration, and hard resource limits. A missing feature is never
replaced with zero. Nonfinite values, incompatible representations, failed
convergence, and resource errors fail the operation without a partial artifact.
Exit codes are 0 for a completed experiment, 2 for an error, and 130 for cancellation.

## Result and qualification

`unswell-editorial-training-v1` records input hashes, fitting options, feature
definitions, representation identity, selected unit/group references, exclusion
counts, and fitted parameters. It contains no source prose, per-unit labels, or
raw vectors. Compact output is limited to 16 MiB; formatted CLI output must also
fit this limit. Preserve the corpus manifest and source producer records with
the training artifact, along with the actual code revision and clean-tree proof.
The digest binds the serialized result; it does not attest the claimed provenance.

Training loss and `calibration_fit_mean_squared_error` describe fitted inputs.
Neither is held-out quality. The result always records `human_corpus: not_qualified`
and `probability_status: unavailable_unqualified_model`, even when the round
declares human participants. It cannot be loaded as a qualified product model.

Reproducibility tests change reserved text and labels without changing fitted
parameters. Separate tests change calibration text without changing classifier
weights. Fixed-order fitting requires no random optimizer seed; the frozen corpus
manifest retains its split seed. Numerical tolerance applies across architectures.

The rule-activation baseline, controlled comparisons, real independent annotation,
held-out evaluation, and model qualification remain required under #21–#25 and #57.
See [ADR 0025](../../../docs/adr/0025-corpus-training.md) for integration boundaries.
