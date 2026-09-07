# ADR 0003: compile configuration bundles and exact term exemptions

Status: accepted; implementation and release acceptance are tracked by #11.

A single-buffer configuration cannot express shared local policy. Inheritance and
file overrides need explicit dependency loading, deterministic merge order, and
the same effective policy in the CLI, library, and MCP adapter.

## Configuration ownership

Compile a caller-supplied bundle of named YAML bytes in `config`. The bundle names
are portable slash-separated paths relative to an agreed project root. The engine
receives this bundle through `Options`; it never resolves filesystem paths itself.
The CLI and MCP startup adapter load local resources through one application helper.
Normal analysis does not reload configuration, expand environment variables, or
fetch network resources.

Resolve `extends` depth first, in list order, relative to the containing file.
Reject cycles, duplicate direct references, unavailable resources, invalid names,
and resource limits before analysis. A caller must explicitly allow dependencies
outside the project root. The application loader checks real filesystem boundaries,
including symlinks. Configuration discovery stops at the project root and does not
read a hidden home configuration.

Apply common defaults, ordered inherited layers, the project layer, and matching
file overrides in order. Maps merge by key; lists replace their previous value.
Reject null and aliases rather than treating them as absence. Explicit false and
empty lists remain meaningful. Builtin profile layers replace rule defaults and
profile gate defaults; they preserve unrelated application settings and vocabulary.

The compiler returns an immutable plan. Its base policy describes discovery and
run-wide limits. Per-file resolution returns an owned policy snapshot, with field
origins and a canonical hash. Overrides can change rules, local analysis limits,
extraction, vocabulary, and local gate thresholds. They cannot introduce new
rulesets, discover more files, nest overrides, or change the run-wide byte budget.
Validate selected combinations before parsing prose, including combinations whose
individual layers are valid but whose merged parameter ranges are incompatible.

Record canonical configuration and dictionary identities, the ordered override
identities, and each document's effective policy hash. Whitespace and YAML comments
do not alter identities. Saved reports retain the resulting diagnostics, scores,
and gate decisions without needing the original configuration files.

## Vocabulary

Support inline terms and shared local YAML dictionaries. Dictionary files contain
`version: 1` and `terms`. Lists select the shared files explicitly; dictionaries do
not execute code or expand more files. Validate duplicates across inline and shared
terms after the selected case and token normalization. Keep the original document,
technical identifiers, negations, and quantities intact.

Find terms as complete token sequences inside a single sentence and source block.
Do not cross protected tokens. Explicit rule IDs choose which diagnostics may use
these matches. Rules must declare support and exclude matching candidates before
computing counts or density. Reject unsupported term exemptions instead of silently
discarding them or removing an aggregate finding whose other evidence remains valid.
A term occurrence does not exempt its sentence or its other rules.

## Interfaces and acceptance

`config explain --file` and MCP description for a logical filename use the same
per-file resolver as analysis. The CLI checks selected per-file size limits before
reading source bytes. Explicit CLI controls remain the final policy layer.

Keep the public single-buffer config API working. Additive public types, report
identities, and the identity algorithm require ledger, schema, and API evidence.
Test merge precedence, cycles and duplicate references, root boundaries, unsupported
fields, dictionary changes, exact term spans, overlapping file overrides, resource
limits, and concurrent reuse. Add root CLI goldens and CLI/MCP self-check evidence.
