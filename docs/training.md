# Training and evaluation

This page describes the research pipeline that turns a pinned corpus into a
fitted model, a frozen prediction into an evaluation, and a fitted artifact into
a pack the engine can load. It is a developer workflow, not a product feature.
Ordinary analysis never runs it.

Nothing here qualifies a model. Qualification needs the human-labeled corpus of
[#22](https://github.com/stokaro/unswell/issues/22) and the published evaluation
of [#25](https://github.com/stokaro/unswell/issues/25). The commands refuse to
pretend otherwise: a run on simulated labels requires `--allow-simulation`, and
every artifact it produces stays experimental.

## The command shape

`corpus` lives in the `research/annotation` module. Each stage reads one JSON
artifact on standard input and writes one on standard output, so a pipeline is
an ordinary shell pipeline and every intermediate result can be saved and
compared.

```sh
cd research/annotation
go run ./cmd/corpus <stage> [options] < input.json > output.json
```

| Stage | Reads | Writes |
| --- | --- | --- |
| `plan` | A manifest of sources | A resolution plan |
| `extract` | A plan | A corpus artifact with prepared targets |
| `verify` | A corpus artifact | The same artifact after source verification |
| `join` | A corpus artifact | Feature bindings for the selected annotation round |
| `train` | A corpus artifact | A fitted numerical artifact |
| `predict` | A corpus artifact | Frozen predictions for one declared plan |
| `evaluate` | Frozen predictions | Metrics against an independent round |
| `compare` | Frozen predictions | A paired comparison against a comparator |
| `reference-bank` | A corpus artifact | A compression reference bank |
| `pack` | A fitted artifact | A product pack the engine loads |
| `figures` | A saved evaluation | One SVG chart |

## Preparing a corpus

`plan` resolves a manifest into an explicit list of sources. `extract` reads
those sources from `--root` and produces prepared targets through the same
extraction and NLP the product uses. `verify` re-reads the pinned sources and
fails when any of them changed. `join` binds an annotation round to the prepared
targets and reports the features each target carries.

`extract`, `verify`, `join` and `train` all require `--root`; `plan` refuses it.
`join` and `train` also require `--round` and at least one `--feature`.

## Fitting a model

```sh
go run ./cmd/corpus train --root sources --round round.json \
  --kind paragraph --feature prose-words --calibration isotonic < corpus.json > model.json
```

`--kind` selects one target kind: `sentence`, `paragraph` or `fragment`. Repeat
`--feature` to select the columns. `--estimator` chooses `logistic` (the default)
or `forest`. `--calibration` chooses `none` or `isotonic`; calibration is fitted
separately from the estimator, on its own partition. `--missing-features`
decides whether an absent column rejects the run or excludes the target.
`--l2`, `--tolerance` and `--max-iterations` bound the logistic fit, and the
`--forest-*` options bound the forest. `--rule-config` supplies the exact inline
policy when a feature is a rule activation.

The artifact records the feature contract, the model and NLP identities, the
seeds and the split manifest, so the same inputs produce the same bytes.

## Provenance labels for the origin task

The origin channel of [ADR 0035](adr/0035-origin-channel.md) estimates a
different target: whether a unit is the endpoint of a generation record. Its
labels come from provenance the corpus declares, never from a blinded round.
`--labels provenance` replaces `--round` on `join`, `train`, and `evaluate`:

```sh
go run ./cmd/corpus train --root sources --labels provenance \
  --kind paragraph --feature prose-words --calibration isotonic < corpus.json > origin-model.json
```

The same flag serves the lexical baseline (`--lexical`), the rule-activation
baseline (`--rule-config`), and `compare`, so every arm of a comparison can
fit and score on the same provenance labels. `--labels cohort` labels the
cohort task instead: `contemporary_snapshot` against `historical_snapshot`.
Controlled responses and natural snapshots stay unresolved. Its rule is
`annotation.CohortProfile`, and such an artifact records task
`cohort_membership` and never becomes a pack. It serves the comparison of
feature families on the pattern cohorts.

The rule is frozen as `annotation.OriginProfile`, and its digest fills the
profile field of the decision set. A unit of a controlled source whose origin
is `generated` with document scope is `endpoint_generated`. A unit of a
historical cohort with origin `human` or `unknown` is `human_snapshot`, a dated
snapshot from before the boundary. Every other unit stays unresolved with a
reason: `polished_response` for `human_ai_edited`, `contemporary_snapshot`,
`natural_cohort`, a declared mixed or edited origin, or a controlled source
without a generation record. Those units are counted as exclusions, and the
rule reads nothing from the text.

A corpus that spans cohorts comes from `corpus dataset union`. It joins the
pinned sources of selected shards into one manifest and keeps every partition
the dataset assigned. It narrows the unit kinds and prefixes each path with the
checkout directory the acquisition driver uses. `plan` and `extract --root`
then read every cohort from the work directory:

```sh
go run ./cmd/corpus dataset union --root artifacts/acquisition --id origin-pilot \
  --cohort controlled --cohort historical --role comment --unit-kind paragraph \
  --max-per-checkout 40 --uncapped-cohort controlled --max-source-bytes 102400 \
  < artifacts/measurement/dataset-plan.json > union.json
```

`evaluate --generation records.json` reads the generation records. Each
controlled source then gets three strata: operation, prompt, and family. Every
evaluation also carries bootstrap intervals over its source groups.
`scripts/origin-experiment.sh` runs the whole sequence, from the union to the
scored partitions, and [research/origin](../research/origin/README.md) records
each run with its digests and what it does not establish.

The artifact records task `origin_endpoint` and rubric
`unswell-origin-endpoint-v1`, its class counts use the two labels above, and
`pack --task origin_endpoint` accepts only such an artifact. An artifact
fitted for one task cannot become a pack for the other. Nothing here
qualifies an origin model: the labels are declared provenance, the generator
may have seen the historical text, and the generation records flag verbatim
overlap.

## Frozen prediction and evaluation

Prediction is frozen: a plan declares the protocol, the model, the corpus, the
partition, the context and the threshold, and the command refuses to run when
any of those digests disagree.

```sh
go run ./cmd/corpus predict --root sources --model model.json \
  --plan plan.json --protocol protocol.md < corpus.json > predictions.json
```

`evaluate` scores those predictions against an independent annotation round.

```sh
go run ./cmd/corpus evaluate --corpus corpus.json --round evaluation-round.json \
  < predictions.json > evaluation.json
```

`predict` refuses `--round` outright, so labels cannot reach the prediction
step. Supplying an independent evaluation round is a protocol obligation the
command does not verify: it checks the digests, not who annotated what. A
simulated round needs `--allow-simulation` and marks the result
`experimental_metrics` with basis `simulation`.

## Comparing two methods

`compare` takes a frozen comparison plan, both trials' predictions, the shared
corpus and the evaluation round, and reports a paired comparison. It keeps the
input policy, preparation, target, rubric, profile and NLP guards fixed, so two
trials differ only in the method under test.

## Feature-family ablations

An ablation is the same pipeline run twice with different `--feature` sets and
compared as a pair. Nothing about it is special-cased: `train` selects the
columns, `predict` freezes each trial against the same corpus and protocol, and
`compare` reports the paired difference with its group support.

Keep everything else fixed. Both trials must share the protocol, corpus,
partition, kind, context, rubric, profile, preparation and extraction policies,
vocabulary and NLP; `compare` refuses a pair that differs anywhere else. That
restriction is the point of an ablation: a difference that could come from two
sources measures nothing.

Declare the families and their order before opening evaluation labels. Choosing
which ablation to report after seeing the results is model selection, and it
invalidates the held-out partition for every trial that follows.

## Converting an artifact into a pack

```sh
go run ./cmd/corpus pack --id my-pack --min-words 12 < model.json > pack.json
```

`pack` copies the numerical parameters and the measurement contract the artifact
already records. It adds only what an artifact cannot hold: an identifier, the
applicability floor chosen from validation data, the estimation target, and any
acceptance a maintainer states.

`--task origin_endpoint` builds a pack for the separate
[origin channel](adr/0035-origin-channel.md). `--accepted` needs a qualified
corpus in the artifact and a named `--evaluation`, so a tutorial run cannot
produce a pack that claims acceptance.

The engine loads a pack through `check --model` and `check --origin-model` with
`calibration.model: pack` or `origin.model: pack` in the policy. An experimental
pack needs `accept_experimental: true` and can never gate a build; see
[scoring](scoring.md) for the applicability statuses and the gate rules.

## Figures and cost

`corpus figures --plot reliability` and `--plot risk-coverage` render one saved
evaluation record as SVG. They compute nothing, so a published chart cannot
disagree with the metrics beside it.

`make research-cost` measures each stage separately and prints wall time, peak
resident set and exit code per stage. `make check` runs its self-test.

## What a tutorial run does not establish

The fixtures under `research/annotation/training/testdata` are simulated. They
exercise the pipeline end to end and nothing else. A pack built from them
declares `experimental` and `not_qualified`, an ordinary scan reports no
qualified probability, and no published number describes rule precision or model
calibration. The [research plan](research.md) records what each method still
needs.
