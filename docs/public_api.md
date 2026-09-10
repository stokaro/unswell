# Public API policy

All public packages are experimental during `v0.1.0-alpha.*`. Pin the exact release.
An alpha suffix does not excuse unannounced incompatible changes: compare API
changes against the preceding release and document intentional differences.

| Import path | Purpose | Stability |
| --- | --- | --- |
| `github.com/stokaro/unswell` | Engine, options and versioned results | alpha |
| `github.com/stokaro/unswell/baseline` | In-memory debt artifacts, fingerprints, and comparison | alpha |
| `github.com/stokaro/unswell/document` | Sources, byte spans and neutral document model | alpha |
| `github.com/stokaro/unswell/rule` | Rule contract, typed parameters and evidence | alpha |
| `github.com/stokaro/unswell/ruleset` | Bounded in-memory compilation of declarative rules | alpha |
| `github.com/stokaro/unswell/builtin` | Explicit builtin catalog factory | alpha |
| `github.com/stokaro/unswell/config` | Strict in-memory policy compiler | alpha |
| `github.com/stokaro/unswell/extract` | Source-preserving input extractors | alpha |
| `github.com/stokaro/unswell/feature` | Versioned prose measurements and repetition preprocessing | alpha |
| `github.com/stokaro/unswell/model` | Immutable numerical classifiers, separate calibration, and deterministic Go fitting | alpha |
| `github.com/stokaro/unswell/probability` | Explicit revision-probability packs, compatibility, and per-unit applicability | alpha |
| `github.com/stokaro/unswell/nlp` | Neutral providers, capabilities, and shared target preparation | alpha |
| `github.com/stokaro/unswell/nlp/english` | Default pure-Go English backend | alpha |
| `github.com/stokaro/unswell/report` | Writers and saved-result decoding | alpha |
| `github.com/stokaro/unswell/goanalysis` | Go analysis adapter in a separate module | alpha |

`scripts/verify-published-module.sh` builds and tests the consumer module
against a published version with no local replacement, using the consumer
sources from the same tag. It also names the public packages the working tree's
consumer imports that the published module does not contain, so a passing check
never implies that the current API shipped. The
[recorded check](release/v0.1.0-alpha.1-module-check.json) resolved
`v0.1.0-alpha.1` with module sum `h1:ZA6hvY4AxndacjiwGGnN5hiFSjn5iKi6QultXJjuY50=`
and passed, while listing seven packages added since that release.

Packages under `internal/` are implementation details. `cmd/unswell` is an
executable, not an importable library. The tools module isolates build dependencies
from the runtime module and its minimum compiler. The consumer module is an
executable public-API contract test, not another supported library.
The `model` package supplies the shared numerical logistic objective, training
normalization, optimizer, and immutable inference. It takes ordered complete
numeric vectors. A sigmoid response is uncalibrated; the parameter snapshot does
not establish feature compatibility, corpus qualification, or a usable probability
gate. Existing engine results and `calibration.model: none` retain their contracts.
See [ADR 0020](adr/0020-logistic-numerical-core.md) and [model documentation](../model/README.md).
`Forest`, `NewForest`, and `FitForest` add the bounded nonlinear candidate in #50.
Snapshots contain complete preorder trees; inference returns owned paths and leaf
class fractions. `ForestOptions` records randomization and resource limits.
Forest responses are uncalibrated and are distinct from logistic linear scores.
See [ADR 0031](adr/0031-forest-numerical-core.md) and the [forest contract](../model/forest.md).
`FitOptions.Validate` exposes the same numerical option checks used by fitting,
so callers can reject invalid limits before preparing rows. It does not inspect data.
`Isotonic` and `FitIsotonic` add a separate monotonic calibration candidate over
frozen scores and binary labels. Snapshots contain ordered knots; inference
interpolates only inside their range. `ErrCalibrationRange` is a numerical
boundary, not a validated domain classifier. A fitted mapping does not establish
independent calibration data, acceptable held-out error, or editorial probability
qualification. See [ADR 0021](adr/0021-isotonic-calibration.md).
The separate `mcp/` module provides the `unswell-mcp` executable and its protocol
tests. It calls the public engine and keeps the MCP SDK out of the core module.
The `research/annotation` consumer module defines versioned research artifacts
and developer commands for annotation and corpus preparation. Its `corpus`
subpackage consumes public extraction/NLP contracts; it does not extend the
supported engine, CLI, MCP, or saved-report APIs. See
[ADR 0013](adr/0013-annotation-protocol.md) and [ADR 0014](adr/0014-corpus-acquisition.md).
`annotation.Round.Decisions` adds a separate research export with resolved labels,
unresolved reasons, original input identities, and detached target references.
It shares primary-response selection with agreement calculations. The existing
round schema and product reports retain their contracts. See
[ADR 0019](adr/0019-editorial-decision-export.md).
`annotation.Round.MatchTargets` and `corpus.Join` bind decisions to reproduced
corpus targets and measurements from the public engine. The separate
`unswell-corpus-feature-bindings-v1` artifact retains unqualified decisions,
source-group partitions, and complete prepared mappings. It does not change
product results or qualify training data. See
[ADR 0024](adr/0024-corpus-feature-bindings.md).
`corpus.MeasureFindings` and the `corpus measure` command run one pinned
policy over a reproduced corpus and attach each finding to the candidates that
contain its primary span. The `unswell-corpus-findings-v1` artifact records
counts per document and per unit under that policy's identity; it carries no
label and no rate. See [ADR 0036](adr/0036-llm-pattern-evidence.md).
The research `training` package and `corpus train` command connect these reproduced
targets to Go fitting with separate training and calibration partitions. Their
`unswell-editorial-training-v3` artifact records unqualified numerical experiments;
it does not extend the engine's model loading or probability contract. See
[ADR 0025](adr/0025-corpus-training.md).
Training v3 requires an explicit `Options.Estimator` and one options/model pair:
`Fit`/`Logistic` or `Forest`/`Forest`. These fields are pointers, and the unused pair
must be nil. Prediction v2 distinguishes forest responses from logistic scores.
Older training and prediction versions are rejected. Both estimators share corpus
selection, frozen predictions, and evaluation; see [ADR 0031](adr/0031-forest-numerical-core.md).
`feature.NewCompression`, `Compression.Measure`, and `CompressionCatalog` add an
optional reference-dependent primitive over those same prepared targets. Results
retain compressor/reference identities, full source bindings, raw byte sizes,
and absent values at compression boundaries. It is not registered in model-free
feature collection or the product gate. Reference-pack selection, training, and
qualification remain separate work; see [ADR 0032](adr/0032-compression-measurements.md).
`annotation.Round.Origins` exports separately curated, source-bound origin claims.
`corpus.BuildCompressionBank` assembles explicit ordered training references over
the same prepared targets and records the common connected-group exclusion union.
The bank contains source prose and remains unqualified. Consumers must apply its
reservations before fitting; see [ADR 0033](adr/0033-compression-reference-bank.md).
`corpus.JoinRules`, `training.RunRules`, and `--rule-config` add the existing
engine's raw activations as an explicit research feature source. They share
target verification, row selection, Go fitting, and calibration with prepared
features. The artifact declares block measurement scope and source-document
context; unmatched targets and disabled rules remain unavailable. See
[ADR 0026](adr/0026-rule-baseline-binding.md).
Research adds label-free `corpus.Measure` and `MeasureRules`, bounded numerical
`training.Load`, frozen `training.Predict`/`LoadPredictions`, and the separate
`evaluation` package. The developer commands `corpus predict` and `corpus evaluate`
separate model execution from evaluation labels. They preserve explicit context,
missing responses, and unqualified status. Predictions embed the current training
artifact; product model loading and reports are unchanged.
See [ADR 0027](adr/0027-frozen-research-predictions.md).
The research `evaluation.Compare` and `RunComparison` APIs add paired numerical
metrics and source-group intervals over saved predictions. `corpus compare`
validates a frozen comparison plan and reports common coverage alongside the
full eligible flow. Its distinct artifact version leaves existing evaluation
and product results unchanged. See [ADR 0028](adr/0028-paired-research-evaluation.md).
The research `llmdet` package validates explicit token/probability rows and finite
numeric ensembles for component compatibility experiments. `cmd/llmdetprobe`
loads an explicit local pack at the application boundary. These research APIs
do not tokenize prose or extend product inference. Their separate formats retain
coverage, absent features, reference margins, and uncalibrated responses. See
[ADR 0029](adr/0029-llmdet-numerical-parity.md) and the
[component experiment](../research/llmdet/README.md).
The root `e2e` directory contains only test code and fixtures. It exports no library API.

Alpha APIs may change without backward compatibility. The package ledger and
consumer tests describe the current contract; CI does not compare API snapshots
against earlier releases. See [the development policy](../CONTRIBUTING.md).

The separate `goanalysis/` module exports `New` and `Options`, builds a singlechecker
driver, and tests against the public engine. The adapter uses `analysis.Pass.ReadFile` for source
access and never imports Unswell internals. Default diagnostics reflect engine
gate failures; `ReportAll` explicitly includes advisory and accepted findings.
All raw results remain available to dependent analyzers as `unswell.RunResult`.

Shared block measurements add the public `feature` package, optional
`rule.Descriptor.SharedFeatures`, and `rule.View.Features`. Enabled consumers
receive one immutable set per enriched document. Public callers can also use
`feature.Measure` and `feature.NewSet` with explicit NLP, policy, source, vocabulary,
preprocessing, and resource identities. Returned values distinguish observed zero
from absence. See [shared features](shared-features.md) and
[ADR 0016](adr/0016-shared-features.md).

The descriptor flag is additive in the existing saved-result schema. Old reports
remain readable; strict older readers need an update for the optional field.
Existing evidence metric names, values, and order retain their definitions. The
block contract has its own identity, `unswell-block-features-v1`; it does not
reinterpret the aggregate `unswell-features-v1` scoring/baseline identity. Descriptor
hashes change for rules requesting the shared set. Statistical model collection,
raw activation vectors, and the remaining #56 feature families are not yet complete.

Lexical preprocessing adds `feature.WordSet`, `WordLimits`, `NewWordSet`,
`Overlap`, `CompareWords`, `InformativeWord`, and `LexicalCatalog`. These share the
existing repetition set-overlap formula and candidate filter. Available empty
sets differ from unavailable zero values; an empty union has no Jaccard number.
The contract is separately identified by `feature.LexicalContract`. No saved-result
fields change, and the block `Catalog` retains its existing measurement order.
Callers must retain source and preprocessing identities for model use and must
not export the explicit word/key accessors in reports by default.

Pattern preprocessing adds `SequenceLimits`, `NgramOptions`, `Ngram`, `ScanNgrams`,
`POSPattern`, `PreparePOSPattern`, and `PatternCatalog`, identified by
`feature.PatternContract`. These string keys have the common descriptor schema,
but are not numeric `Value` measurements. The streaming callback owns its n-gram
key and receives original token indices; failure stops the scan and invalidates
partial output. Surface templates keep explicit absence reasons and caller-selected
literal tokens. These additions preserve saved-result fields and existing rule
evidence. See [shared features](shared-features.md) for caller policy obligations.

`Options.Features` and `Engine.FeatureIDs` select and expose a canonical set of
implemented block measurements. `feature.SupportsBlock` declares their supported
unit kinds. Requested capabilities are validated before analysis. Collection reads
the existing shared set and adds optional `RunResult.Features` containing
`FeatureCollection`, `FeatureSource`, and `FeatureUnit` records under
`FeatureCollectionVersion`. Source/policy/NLP identities and original segments
travel with the numeric values; incomplete runs remain incomplete model inputs.
The additive field is omitted by default. Old reports remain readable; strict
older readers need an update when collection is explicitly requested. See
[ADR 0017](adr/0017-feature-collection.md) for compatibility and remaining scope.

Optional `FeatureUnit.Binding` adds `FeatureBlockBinding` and
`FeatureBlockBindingContract`. The `mapped-block-v1` binding contains full mapped
text and outer-whitespace-trimmed hashes, with separate complete source maps.
It includes punctuation and protected separators that counted token spans omit.
Readers accept legacy reports without it and validate present bindings, including
preservation of interior segments after trimming. Strict older readers need an
update for this additive field. Full and trimmed maps count against the existing
collection budget. No rule values or prepared-target semantics change.
See [ADR 0026](adr/0026-rule-baseline-binding.md).

`nlp.PrepareUnits`, `PreparedUnit`, `UnitBinding`, `UnitOptions`, and `UnitLimits`
add shared sentence/paragraph/fragment preparation under `nlp.UnitContract`.
Prepared units retain separate target, surrounding prose, and grammar identities;
accessors return owned NLP data. `feature.MeasureUnit` and `UnitCatalog` reuse
descriptive formulas under `feature.UnitContract` with an explicit target scope.
They verify the actual provider and requested capabilities. Existing block
collection, rule activations, and saved reports keep their contracts. See
[prepared units](prepared-units.md) and [ADR 0022](adr/0022-shared-prose-units.md).

`Options.PreparedFeatures` and `PreparedKinds` select prepared target collection;
`Engine.PreparedFeatureIDs` and `PreparedUnitKinds` expose owned canonical sets.
`RunResult.PreparedFeatures` contains `PreparedFeatureCollection`,
`PreparedFeatureSource`, and `PreparedFeatureUnit` under
`PreparedFeatureCollectionVersion`. Target/context bindings, counted segments,
actual capabilities, and separate policy/extraction identities accompany the
numeric values. This optional additive field leaves default saved output and
block collection semantics unchanged. See [ADR 0023](adr/0023-prepared-feature-collection.md).

The additive `rule.Parameters.MaxAnswerWords` field configures the experimental
question/answer pattern. Its zero value is omitted from saved parameter objects;
enabled use requires a value from 1 through 100. The report schema includes the
optional field, and existing saved settings remain valid.
Older readers with a strict schema may reject new reports containing this field;
use a reader that supports the producing tool's schema additions.

Repetition signals add optional `rule.Parameters.MinNgramWords`, `MaxNgramWords`,
and `WindowBlocks` fields. Their supported ranges are 3 through 8 words with an
ordered minimum/maximum, and 1 through 128 prose blocks. Zero values are omitted
from saved JSON. `rule.Descriptor.RequiresStructure` requests the existing
grammar-derived block context for an enabled rule, including per-file overrides;
it does not restore excluded prose. The saved-report schema includes these
additions. Strict older readers need an update to accept reports using them.
Existing rule IDs and behavior versions retain their meaning. See
[repetition signals](repetition-signals.md) for formulas and limits.

Callers own returned result slices. Source bytes must not change during analysis.
One engine supports concurrent calls. Custom rules and providers are trusted Go
code and must honor the concurrency and read-only view contracts.

The engine reads no configuration paths, environment variables, current directory,
standard streams or network endpoints. The CLI supplies source bytes and policy.
Reporters consume an existing result and return write errors to the caller.

The custom-rule DSL adds `ruleset.Load`, immutable `Set` accessors,
`Options.RuleSets`, and `config.Compile`. The latter returns both the effective
policy and the additional implementations from inline `rule_sets`. `config.Load`
still returns only the policy; the engine registers inline rules from the compiled plan.
Declarative packs extend the selected Go registry. Duplicate IDs are errors.

Rule descriptors gain optional `Origin` metadata, examples gain an optional
`Format`, and effective policies gain `RuleSets`. These are additive fields in
the version-1 report schema. Old reports remain readable; readers pinned to an
older release that reject unknown fields need an update to read these new fields.
Ruleset identities are included in policy and ruleset hashes. Changing matcher
definitions invalidates their identity even when the human-readable release stays
the same. See [custom rules](custom-rules.md) for the versioned YAML contract.

Configuration inheritance adds `config.Bundle`, `Reference`, `SourceIdentity`,
`OverrideIdentity`, `Vocabulary`, and immutable `Plan`, plus `CompileBundle`,
`References`, and `ResolveReference`. Plans return owned base and per-file policies
and a conservative list of possibly enabled rule IDs. `Options.ConfigBundle`,
`Engine.PolicyForFile`, and `Engine.Catalog` expose those capabilities through the
engine. The caller grants any outside-root permission explicitly; the library still
performs no filesystem access.

Rules gain opt-in `Descriptor.TermExemptions`, `View.TermExemptions`, `View.Exempts`,
`TokenRange`, and immutable `TermMatches` constructed with `NewTermMatches`.
Supporting rules must exclude candidates before computing aggregate metrics.

Policies add identity, vocabulary, sources, overrides, and applied override metadata.
`unswell-config-bundle-v1` intentionally changes config hashes to include resource
provenance. Equivalent inline/external configurations may now have different config
hashes while retaining the same ruleset hash and findings. Context-set ordering is
still irrelevant. Existing config hashes cannot be compared across these algorithms.
See [configuration](configuration.md) for merge, root, origin, and pinning semantics.

The saved-result schema gains optional manifest config identity/source/override
fields and per-document config hashes and applied overrides. These additions retain
the existing schema version and old-report readability. Strict older readers need
an update to accept the added fields. MCP description gains an optional logical
filename and returns the same selected policy used by analysis.

Source suppressions add `document.Directive`, `Document.Directives`,
`config.Suppressions`, and `Policy.Suppressions`. The engine validates and resolves
them after extraction and NLP. Results add `Suppression`, `SuppressionTarget`,
`RunResult.Suppressions`, `Finding.SuppressionIDs`, and
`Assessment.EffectiveContributions`. Existing findings, source locations, raw
contributions, and raw document statistics retain their meaning. Effective unit
scores and gate decisions now account for valid source permissions.

These are additive fields in the existing result schema. Old saved reports remain
readable; strict older readers need an update. The new suppression policy defaults
intentionally change effective policy hashes. No probability field is reinterpreted.
Malformed or unused permissions return an operational error and an incomplete result;
unused-permission errors preserve the raw findings and audit records already computed.

Baseline support adds `Options.Baseline`, `CollectBaseline`, and `GateMode`, the
public `baseline` package, `extract.Options.IncludeStructure`, and optional
`document.Block.Context`. Config gates add `Mode`, defaulting to `all`; this changes
effective configuration hashes. Result additions include snapshots, comparisons,
finding and assessment baseline identities, selected gate modes, and
`GateDecision.Accepted`. Old saved results remain readable under the existing schema
version; strict older readers need an update for these additive fields.

`Finding.ID` now distinguishes separate source occurrences of the same legacy
fingerprint. It is a run-local reference used by scoring and suppression traces,
and can change with line movement. The existing `Finding.Fingerprint` field retains
its algorithm. Durable debt uses the separately versioned `BaselineFingerprint`;
do not persist finding IDs as baseline identities. See [baseline behavior](baseline.md).

Committed comparison adds `Engine.AnalyzeChanged`, `ChangeSelection`,
`ChangedDocument`, and application-owned `GitSelection` provenance. Results gain
optional change identities/states on findings and assessments, `RunResult.Changes`,
`Manifest.Git`, and `GateDecision.Unchanged`. These additive fields retain the
existing schema version and old-report readability; strict older readers need an
update. Raw scores, baseline acceptance, and ordinary analysis keep their semantics.
See [changed-unit analysis](changes.md) for selection and clean-source requirements.

Trusted comparison adds `Engine.AnalyzeChangedWithOptions`, `ChangeOptions`,
`PolicyChange`, and `PolicyComparison`. Results gain optional `PolicyComparison`,
and suppressions gain optional `TrustState` and `TrustFingerprint`. These additive
fields retain the existing schema version; strict older readers need an update.
The caller asserts which engine policy, baseline, and before-sources are trusted.
The engine filters current permissions and performs full selection for reported
resource changes without loading policy resources. Ordinary analysis, baseline
identities, and directive syntax retain their contracts. See
[trusted policy](trusted-policy.md) for scope, hashes, and CI trust boundaries.

Dependency evaluation adds `document.DependencyTree`, `DependencyArc`, optional
`Sentence.Dependencies`, `nlp.Identity.DependencyScheme`, and
`rule.Descriptor.DependencyScheme`. `nlp.ValidateDependencies` checks block
coverage, and `ValidateDependencyTree` validates basic sentence-local trees
against extracted byte ranges. Scheme IDs are compared
exactly; an empty rule scheme permits structural analysis without interpreting
labels. These additions do not advertise dependency support in the builtin NLP
provider. See [ADR 0012](adr/0012-dependency-contract.md).

Providers must supply the base token/sentence capabilities and every requested
representation. A requested dependency tree that is absent, malformed, or mapped
across protected content now fails before rule execution, including with
`NoGate`. Previously the engine checked the advertised dependency name without
having a tree contract. Custom providers relying on that incomplete behavior
must implement the representation. Existing rules-only providers are unchanged.

The optional scheme fields extend saved provider/rule identities under the
existing schema version. Old reports remain readable; strict older readers need
an update for reports containing the new fields. Trees remain in the in-memory
document model and are not added to normal saved reports or MCP results. The
separate dependency research command explicitly emits source-bearing traces for
reference experiments and is not part of the supported product API.

Rule activation collection adds `feature.ActivationContract`, `ActivationDescriptor`,
`ActivationBuilder`, `NewActivations`, `BlockObservation`, `ValidateActivation`,
and `ValidApplicabilityReason`. The mutable builder is local to one evaluation;
its returned values are owned. `rule.Observer`, `View.Observer`, `View.Observe`,
and `Descriptor.BlockObservations` expose optional applicability reporting without
changing the Rule or Emitter interfaces. Missing observations never imply zero.

`Options.Features` also accepts registered `activation/<rule-id>` requests.
`FeatureCollection.ActivationContract` and `FeatureSource.RulesetHash` bind the
new values to their formula and rules. These fields are additive and omitted for
ordinary block collection. Existing reports remain readable; strict older readers
need the new fields. Rule identities now include their observation declaration.
See [ADR 0018](adr/0018-rule-activation-features.md) for failure states and the
remaining applicability coverage required before training.

Research lexical training adds `corpus.Prepare`, `training.RunLexical`, and an
explicit source-derived vocabulary under `unswell-editorial-lexical-training-v1`.
`feature.CountLexical` and `ValidLexicalKey` share bounded word/character formulas
on existing `nlp.PreparedUnit` data; `nlp.PreparationHash` shares the existing source
preparation identity between the engine and research consumers. Existing v1 model
bytes remain unchanged. Learned lexical dictionaries have their own identity,
separate from the engine's policy/vocabulary identity. Predictions, evaluation,
and comparison use the existing commands and retain unqualified status. See
[ADR 0030](adr/0030-lexical-research-baseline.md).

The `probability` package decodes one explicitly supplied revision-probability
pack and decides applicability per unit. `Load` validates the contract, columns,
capabilities, digest, and author declarations; `Compatible` compares only the run
inputs that change measured values; `Estimate` returns a calibrated value for an
applicable unit and an explicit abstention status otherwise. A missing measurement
is never imputed, and a mismatched vector is a caller error. `StatusUnavailable`
keeps the existing model-free status of results. Loading a pack establishes
neither corpus qualification nor a usable probability gate; engine integration,
calibrated gating, and acceptance remain separate work. See
[ADR 0034](adr/0034-probability-pack.md).

`Options.Model`, `check --model`, and `calibration.model: pack` connect that pack
to the engine. `Assessment.ProbabilityStatus` now carries the decided status,
`ProbabilityDetail` names an unavailable column, and `Manifest.Probability`
records the pack's declarations. `config.Policy.Calibration` is absent unless a
policy requests a model, so model-free identities, accepted debt, and existing
reports are unchanged. Estimates do not affect findings, the index, or the gate.
Every writer and `unswell-mcp --model` report the same decisions.
`config.Gate.Probability` adds the calibrated gate. It is absent unless a policy
configures it, requires an accepted pack, and produces the derived
`gate.<kind>-probability` and `gate.<kind>-probability-unavailable` diagnostics.
Estimates still do not change findings or the index.
`report.ValidateSARIF` and `report.SARIFSchemaVersion` check one SARIF report
against the published 2.1.0 schema, embedded so validation reads no external
resource. A valid document is well formed; it is not evidence that a particular
consumer accepts it. See [SARIF output and consumers](sarif.md).
`probability.TaskOrigin` and `Pack.Task` add the separate origin target, and
`Options.OriginModel`, `config.Origin`, `Assessment.OriginEstimate`,
`OriginStatus`, `OriginDetail` and `Manifest.Origin` carry it through the engine
and the reports. The channel is absent unless a policy configures it, and no
gate consults it. See [ADR 0035](adr/0035-origin-channel.md).
The research `training.BuildPack` and `corpus pack` convert one fitted artifact
into a pack the engine loads. They copy its numerical parameters and measurement
contract and add only an identifier, an applicability floor, the estimation
target, and an acceptance a maintainer states. A simulated artifact cannot
produce an accepted pack.
`report.Formats` names the five report writers in a stable order. Callers that
offer every format no longer repeat the list, and a writer that is added without
being listed fails its own contract test.

The research `evaluation.RiskPoint` adds a risk-coverage curve to the full-flow
summary. It sorts covered decisions by confidence. At each reachable threshold it
gives the error rate among the accepted prefix. A reader can then see where a
selective model stops being right. Saved records now carry
`unswell-research-evaluation-v4`, which also carries recall at fixed
false-positive limits, a prevalence-sensitivity table, and a per-stratum
breakdown by recorded corpus attributes (`evaluation.RecallPoint`,
`evaluation.PrevalencePoint`, `evaluation.Stratify`). A new `figures` package and `corpus figures`
render that record as SVG. Both read the saved values and compute nothing, so a
chart cannot disagree with the metrics beside it.

The `feature` catalog adds a repetition family: `peak-word-frequency-ratio`,
`repeated-bigram-ratio`, `sentence-opener-repeat-ratio` and
`duplicate-sentence-ratio`. They join the existing contract, so earlier
descriptors keep their definitions and positions, while measurement identities
change because they cover the descriptor set. See
[ADR 0016](adr/0016-shared-features.md).
