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
| `github.com/stokaro/unswell/nlp` | Neutral provider and capability contract | alpha |
| `github.com/stokaro/unswell/nlp/english` | Default pure-Go English backend | alpha |
| `github.com/stokaro/unswell/report` | Writers and saved-result decoding | alpha |
| `github.com/stokaro/unswell/goanalysis` | Go analysis adapter in a separate module | alpha |

Packages under `internal/` are implementation details. `cmd/unswell` is an
executable, not an importable library. The tools module isolates build dependencies
from the runtime module and its minimum compiler. The consumer module is an
executable public-API contract test, not another supported library.
The separate `mcp/` module provides the `unswell-mcp` executable and its protocol
tests. It calls the public engine and keeps the MCP SDK out of the core module.
The root `e2e` directory contains only test code and fixtures. It exports no library
API; the API snapshot records its package identity because it belongs to the root
module.

The separate `goanalysis/` module exports `New` and `Options`, builds a singlechecker
driver, and tests against the public engine. Its API snapshot is
`docs/api/goanalysis.api`. The adapter uses `analysis.Pass.ReadFile` for source
access and never imports Unswell internals. Default diagnostics reflect engine
gate failures; `ReportAll` explicitly includes advisory and accepted findings.
All raw results remain available to dependent analyzers as `unswell.RunResult`.

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
