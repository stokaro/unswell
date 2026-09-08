# Accepting existing editorial debt

A baseline records an explicit decision to accept existing findings and scored
units. Findings, source permissions, and raw and effective scores remain in the
result. Acceptance does not make the prose clean or provide an authorship label.

Create a baseline after reviewing a complete scan:

```sh
unswell baseline create . --output .unswell-baseline.json
unswell baseline check . --baseline .unswell-baseline.json
```

Creation requires a new output file. `baseline check` defaults to `new` mode and
never changes the baseline. It returns 0 for a complete pass, 1 for new policy
failures, and 2 for an operational error or incomplete check. Cancellation returns
130. Missing, invalid, incompatible, or ambiguous baseline data cannot pass through
`--no-gate`.

The ordinary check command accepts the same artifact:

```sh
unswell check . --baseline .unswell-baseline.json --gate-mode new
unswell check . --baseline .unswell-baseline.json --gate-mode all
```

`all` evaluates the entire editorial gate while displaying baseline state. `new`
accepts exact existing findings and scored units. New findings, changes to a scored
paragraph, and changed repetition evidence remain subject to the gate. An accepted
forbidden phrase cannot hide a new threshold failure in its paragraph.

The configuration default is `gate.mode: all`. Set `gate.mode: new` globally or in
a file override when the invoking application supplies a baseline. An explicit
`--gate-mode` overrides that selection for the run. The baseline path belongs to
the invoking application, so YAML cannot silently select a different debt file.

```yaml
version: 1
gate:
  mode: new
overrides:
  - files: [docs/new/**]
    gate:
      mode: all
```

## Updating and coverage

An explicit update accepts current debt for each completely scanned document and
removes its stale entries:

```sh
unswell baseline update . --baseline .unswell-baseline.json
```

Create and update scan complete source files using the normal discovery policy.
They allow findings, but reject incomplete analysis. Their output is written
atomically after validation; source, configuration, and ruleset inputs cannot be
overwritten. They do not accept stdin, report destinations, or gate overrides.
A configuration with `gate.mode: new` can still create or update a baseline.

For a partial path selection, debt in unselected files remains **unobserved**.
It is retained during updates. A missing entry in an observed file is **stale**;
checking reports it, and only an explicit update removes it. Deleted files are
also unobserved: this implementation does not infer deletion from an omitted path.
To retire deleted-file debt, create and review a replacement artifact from the
intended complete project selection.

Changes to global analysis compatibility require scanning every previously
recorded path before updating. A per-file policy or source-permission change
requires an explicit update for that file. Partial coverage cannot silently accept
a new global policy for files that were not checked.

## Identity and compatibility

`unswell-baseline-v1` stores `unswell-structural-v1` fingerprints. Identity includes
the project-relative path, rule behavior version, grammar-derived context,
canonical block content, and evidence. Markdown headings and named declarations
or keys distinguish otherwise identical prose. Line numbers and report messages
are not identity inputs. Whitespace and Markdown emphasis can change without
invalidating debt; negation, numbers, identifier case, and protected code remain
significant. Changes inside a paragraph recheck the whole paragraph.

Document-level repetition includes its complete occurrence cluster. Adding a
repeat can invalidate acceptance of older paragraphs even when their wording has
not changed. Indistinguishable duplicate candidates are an error; Unswell does not
assign arbitrary line ordinals to make them appear distinct. Anonymous structures
and repeated template sections can require separate named contexts or a narrower
initial baseline selection.

Compatibility includes effective analysis settings, rule definitions and behavior
versions, NLP, feature and scoring contracts, vocabulary, and calibration identity.
Display severity and gate mode are excluded. Ordinary result manifests retain the
full configuration provenance. Custom Go rule authors must advance behavior
versions when implementations change; Unswell cannot hash arbitrary Go behavior.

Source-permission identity includes resolved structural targets and their block
content. Moving unrelated lines is allowed. Moving a permission to another target,
changing its reason or rules, or editing its target requires explicit review and
update. Baseline acceptance never changes source-suppressed findings into accepted
baseline entries.

The artifact contains hashes and metadata, without source prose or timestamps.
Limits are 16 MiB, 20,000 accepted entries, and 10,000 documents. Nonzero effective
sentence and paragraph scores each count as an entry, as does each active finding.
Baseline collection requires normalized project-relative names even when an
ordinary scan can use an arbitrary logical name.

## Library, saved reports, and MCP

Set `Options.CollectBaseline` to obtain `RunResult.BaselineSnapshot`, then pass that
complete snapshot to `baseline.Create` or `baseline.Update`. Encode and load with
`baseline.Encode` and `baseline.Load`. The caller owns all I/O. Supply stored bytes
through `Options.Baseline` and choose `Options.GateMode` when overriding config.
Collection is optional; ordinary scans do not construct structural context.

Findings and assessments carry a separate `BaselineFingerprint` and baseline state.
`Gate.Accepted` retains reasons admitted by new mode. `RunResult.Baseline` contains
existing/new matches, stale entries, and unobserved debt. All five reporters consume
this stored result. JSON can be converted later without source files or a baseline;
SARIF maps existing debt to `baselineState: unchanged`.

```sh
unswell-mcp --config .unswell.yaml --baseline .unswell-baseline.json --gate-mode new
```

MCP loads the artifact once at startup. Tools can inspect the selection and check
drafts under it, but cannot alter the policy or accept debt. Restart after an
explicitly reviewed update. Source names sent by the client must match the
project-relative names in the artifact.

Git changed-unit selection and trusted base-branch policy remain separate roadmap
work. A baseline alone does not protect a CI configuration from changes in a PR.
