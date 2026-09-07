# Configuration inheritance and vocabulary

Unswell compiles local YAML into one policy plan. The CLI, public engine, and MCP
server resolve file overrides through that same plan. Rules and NLP do not open
configuration files while checking source.

## Loading and precedence

The CLI searches for the nearest `.unswell.yaml` from its working directory up to
the project root. The default root is the nearest ancestor with a `.git` entry;
outside Git it is the working directory. `--project-root` selects the boundary
explicitly. A directory argument still selects that directory's sources when the
project root is higher. Include patterns and file overrides use root-relative paths.

`--config` selects an exact file. Local `extends` and dictionary paths are relative
to the file containing them. They must remain inside the project root, including
after resolving symlinks. An external config or dependency requires an appropriate
`--project-root` or explicit `--allow-config-outside-root`. This is a startup
permission, not a YAML setting. URLs, absolute dependency paths, drive names, and
backslash separators are invalid. No home config, network download, environment
expansion, or template execution is added.

The merge order is:

1. Builtin defaults.
2. Each `extends` entry from left to right, resolving its parents first.
3. The current file's own settings.
4. Matching `overrides` in list order.
5. Explicit application controls such as `--allow-empty` and `--no-gate`.

Maps merge recursively. Lists replace the inherited list, including `overrides`,
`rule_sets`, dictionary selections, phrases, and inline terms. An empty list clears
that field; an absent field preserves it. YAML null, aliases, duplicate keys,
unknown fields, invalid IDs, and incompatible parameter ranges are errors.
Extraction `contexts` retain their set semantics: order does not affect identity.

An explicit builtin profile layer replaces rule defaults and profile gate thresholds.
It preserves unrelated discovery, extraction, analysis, vocabulary, and suppression settings.
For example, put `builtin:strict-v1` before local policy files whose rules should
override that profile. `--profile` selects a builtin policy without discovery and
remains mutually exclusive with `--config`.

Cycles and duplicate direct references are rejected, including `base.yaml` versus
`./base.yaml`. A shared ancestor may occur in separate inheritance branches and
is applied in each branch's order. The loader rejects different paths to the same
physical file. Limits are 64 resources, 1 MiB per resource, 8 MiB total, inheritance
depth 16, and 256 applied layers. A config accepts at most 64 direct parents and
64 file overrides; the expanded graph allows at most 256 declared overrides.

## Shared terms and file overrides

For this layout:

```text
.unswell.yaml
policies/team.yaml
policies/terms.yaml
docs/guide.md
docs/reference/api.md
```

Use `.unswell.yaml`:

```yaml
version: 1
extends: [builtin:technical-v1, policies/team.yaml]
overrides:
  - files: [docs/reference/**]
    rules:
      hype.modifier-cluster:
        enabled: false
```

Use `policies/team.yaml`:

```yaml
version: 1
vocabulary:
  terms: [dependency injection]
  dictionaries: [terms.yaml]
  term_exemptions: [hype.modifier-cluster]
  case_sensitive: false
```

Use `policies/terms.yaml`:

```yaml
version: 1
terms: [robust estimator, control plane, schema migration]
```

Terms match complete tokenizer tokens inside one sentence and block. They do not
cross protected code. Case-insensitive matching uses the engine's token normalization;
case-sensitive matching preserves token spelling. Terms are not substrings, and
an identifier is never rewritten to make it match. Duplicate normalized terms across
inline and selected dictionary entries are errors. The effective vocabulary accepts
at most 10000 terms, each at most 1000 UTF-8 bytes and 32 tokens.

Only the listed rule IDs use term matches. A candidate fully contained in one term
is removed before counts, density, and scoring are calculated. For example, allowing
`robust estimator` does not also allow another `robust` elsewhere in the sentence.
Overlapping terms do not create a larger combined exemption. Other rules continue
to inspect the complete prose.

The phrase rules, `hype.modifier-cluster`, and `density.connective-overuse` support
term exemptions. Unsupported rules are rejected, including aggregate sentence
checks and current declarative DSL rules. Custom Go rules can declare
`Descriptor.TermExemptions` and consult `View.Exempts` before aggregation.

File overrides accept `rules`, `gate`, `analysis`, `extraction`, `vocabulary`, and `suppressions`.
They cannot add rulesets, change discovery, nest overrides, or change run-wide
`max_total_bytes`, `fail_on_empty`, or `fail_on_incomplete`. Every declared override
is validated against the merged base policy. Selected combinations are validated
again before source parsing; two individually valid layers can still conflict.

Source [suppression settings](suppressions.md) require reasons, reject unused
permissions, and disallow file-wide permissions by default. These fields merge
independently, including explicit `false` values in file overrides. They participate
in policy hashes and field origins; builtin profile layers preserve them.

## Inspecting and pinning policy

```sh
unswell config validate
unswell config explain --file docs/reference/api.md
unswell doctor
```

The explanation contains effective values, source identities, applied override IDs,
and field origins. Granular origins use JSON Pointer paths such as
`/rules/hype.modifier-cluster/enabled`; the existing `rules.<id>` origin remains
available for callers using it.

`unswell-config-bundle-v1` identifies the new hashing contract. Hashes include the
effective settings, origins, canonical identities of all referenced configs and
dictionaries, and ordered file overrides. YAML comments, formatting, and context-set
order do not change the hash. Changing a shared dictionary does. A file's hash
includes its applied override IDs; it does not depend on another file's contents.
Equivalent settings loaded through different config paths can have different hashes
because their provenance differs. Inline and external rule packs still share their
ruleset identity when their definitions match.

Pin the exact Unswell release and explicit profile version. Record the configuration
hash, ruleset hash, and NLP identities from a reviewed run; check those identities
when updating the binary or shared policy. For example, a trusted CI policy can
compare `.policy.hash` from `config explain` with its reviewed value. A hash records
identity, not approval; source PRs must not control their own expected policy hash.

Saved JSON records `manifest.config_identity`, config resource identities, ordered
overrides, and each document's `config_hash` and `applied_overrides`. SARIF and HTML
retain file policy identities. Text and Markdown name the bundle hash when local
inheritance or file overrides are present. Rendering a saved result never reloads
those files. No vocabulary entries or source prose are added to the manifest.

Library callers provide `Options.ConfigBundle` with portable names and bytes, then
use `Engine.PolicyForFile` for explanations. `Options.Config` still accepts one
buffer; it cannot be combined with a bundle. Neither API discovers local files.
`config.CompileBundle` exposes the same compiled plan independently of the engine.

The MCP startup adapter uses this loader with discovery disabled. Its explicit
config establishes the starting directory for root selection unless `--project-root`
is set. `unswell_describe` accepts an optional logical `file`; it resolves the fixed
policy without opening that source or permitting policy changes.
