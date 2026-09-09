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
| `--rule-config` | absent | Select rule activations with an explicit local inline policy, limited to 1 MiB |
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

Controlled comparisons, real independent annotation, held-out evaluation, and
model qualification remain required under #21–#25 and #57.
See [ADR 0025](../../../docs/adr/0025-corpus-training.md) for integration boundaries.

## Baseline using existing rules

Pass `--rule-config` and activation IDs to use the ordinary engine's raw rule
values. The same `plan` and `extract` commands above produce the input:

```sh
go run ./cmd/corpus train --root training/testdata/sources \
  --round training/testdata/round.json --kind paragraph \
  --feature activation/policy.banned-phrases \
  --rule-config training/testdata/rules.yaml \
  --calibration isotonic --allow-simulation \
  < candidates.json > rule-training.json
```

The included policy supplies two phrases for the scripted fixture. The compiled
e2e test verifies observed training values of 0 and 1 and separate calibration.
These results establish mechanics only. Use a prospectively selected policy on
real labeled data for a scientific comparison.

`corpus join` also accepts `--rule-config` to inspect block values, source/policy
identities, and every matched or unmatched target before fitting. The policy
must be self-contained; external includes are not loaded. It must preserve the
frozen corpus extraction policy. Requesting an activation does not enable its
rule. Disabled rules retain `disabled`, and other unavailable values keep their
reasons. `--missing-features reject` fails on a selected, resolved missing row;
`exclude` records the reason. Neither setting replaces a missing value with zero.

A target must match a complete original block after the shared outer-whitespace
trim. Text, context, source hash, and all source segments must agree. A protected
piece or a sentence inside a longer block cannot inherit the block's activation.
This restriction also applies when the parent carries an annotation.

The training identity declares `feature_source: rule_activations`, block-scoped
columns, `context: source_document`, activation and binding contracts, actual NLP,
ruleset, and effective policy hashes. `rule_config_sha256` records the exact
supplied bytes; retain that policy with the experiment. Prepared-specific
extraction/preparation hash fields are empty for this representation; the effective
policy hash binds extraction. `include_structure` is true because block collection
requests source structure. Source-level input hashes
also bind surrounding document content. Compare coverage and common targets
before comparing this baseline with isolated prepared-target features.

Both paths use the same selection, training rights, numerical optimizer,
normalization, and calibration code. Neither path consumes reserved labels or
turns a numerical fit into a qualified probability. See
[ADR 0026](../../../docs/adr/0026-rule-baseline-binding.md).

## Frozen predictions and evaluation

`corpus predict` restores this numerical artifact without fitting weights or
calibration. It requires a frozen plan, matching protocol bytes, and the original
candidate corpus and sources. It accepts no annotation round. Rule predictions
require the exact inline configuration used for fitting. Prepared prediction
measures target features after NLP analyzes the enclosing eligible piece; rule
prediction retains source-document context.

`corpus evaluate` reads saved predictions and an independent annotation round.
It requires no source root, model file, or model execution. See the
[evaluation guide](../evaluation/README.md) for the two-stage command contract,
metric denominators, simulation handling, and remaining scientific acceptance.
