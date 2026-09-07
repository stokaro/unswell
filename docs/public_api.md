# Public API policy

All public packages are experimental during `v0.1.0-alpha.*`. Pin the exact release.
An alpha suffix does not excuse unannounced incompatible changes: compare API
changes against the preceding release and document intentional differences.

| Import path | Purpose | Stability |
| --- | --- | --- |
| `github.com/stokaro/unswell` | Engine, options and versioned results | alpha |
| `github.com/stokaro/unswell/document` | Sources, byte spans and neutral document model | alpha |
| `github.com/stokaro/unswell/rule` | Rule contract, typed parameters and evidence | alpha |
| `github.com/stokaro/unswell/ruleset` | Bounded in-memory compilation of declarative rules | alpha |
| `github.com/stokaro/unswell/builtin` | Explicit builtin catalog factory | alpha |
| `github.com/stokaro/unswell/config` | Strict in-memory policy compiler | alpha |
| `github.com/stokaro/unswell/extract` | Source-preserving input extractors | alpha |
| `github.com/stokaro/unswell/nlp` | Neutral provider and capability contract | alpha |
| `github.com/stokaro/unswell/nlp/english` | Default pure-Go English backend | alpha |
| `github.com/stokaro/unswell/report` | Writers and saved-result decoding | alpha |

Packages under `internal/` are implementation details. `cmd/unswell` is an
executable, not an importable library. The tools module isolates build dependencies
from the runtime module and its minimum compiler. The consumer module is an
executable public-API contract test, not another supported library.
The separate `mcp/` module provides the `unswell-mcp` executable and its protocol
tests. It calls the public engine and keeps the MCP SDK out of the core module.
The root `e2e` directory contains only test code and fixtures. It exports no library
API; the API snapshot records its package identity because it belongs to the root
module.

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
