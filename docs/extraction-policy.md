# Extraction policy

Select the prose contexts to check with a global set and per-language overrides:

```yaml
version: 1
extraction:
  contexts: [comment, string, paragraph, heading, list-item, table-cell]
  languages:
    csharp:
      contexts: [comment]
    yaml:
      contexts: [comment, string]
    markdown:
      contexts: [paragraph, heading]
```

The global default contains all six values shown above. Each format uses only
its supported contexts: text has `paragraph`; Markdown has `paragraph`, `heading`,
`list-item` and `table-cell`; source languages and YAML have `comment` and `string`.
The keys under `languages` are format names from the [input table](inputs.md).

A language override replaces the global set. An omitted language inherits it.
An explicit empty set (`contexts: []`) selects no prose; it does not restore defaults.
The empty-scan gate still applies if the run has no applicable prose.
Unknown values, duplicate entries and contexts unsupported by an overridden
language are configuration errors. Set order does not affect the policy hash.

Disabled regions retain their spans with `config:context-disabled:<context>`
exclusion reasons. Syntax validation still runs. Existing protections for code,
URLs, quotations and technical metadata remain in effect within selected contexts.

## Reasoned exceptions

Unswell's purpose is to keep formulaic AI-style wording out of ordinary code and
documentation. Tests and phrase dictionaries sometimes need that wording as data.
An extraction exception records why a selected comment or string is not checked.
It preserves the source and leaves the rule catalog and severity settings intact.

```yaml
version: 1
extends: [builtin:strict-v1]
extraction:
  exceptions:
    - id: phrase-fixtures
      paths: ["fixtures/**/*.go"]
      formats: [go]
      kinds: [string]
      symbols: [exampleText]
      reason: "These strings are deliberate positive examples for an editorial rule."
```

Every nonempty selector must match. Values within a selector are alternatives:

- `paths` is required and matches project-relative source names with `*`, `**`
  and `?`. It uses the same glob syntax as file discovery.
- `kinds` is required and contains `comment`, `string`, or both.
- `formats` optionally limits the exception to supported source grammars.
- `symbols` optionally selects enclosing named declarations or keyed fields.
  It supports strings only and matches source spelling exactly. A symbol
  restriction on comments is rejected because leading comments do not have a
  consistent declaration parent across grammars.
  YAML mapping keys select their value subtree; quoted keys match with their
  original quote delimiters. C# names select their enclosing declaration.
- `id` is required, unique, and contains lowercase letters, digits or hyphens.
  It must start with a letter and have at most 64 characters.
- `reason` is required, nonblank, and at most 500 bytes.

A matched exclusion retains its original byte span and records
`config:<id>: <reason>` in saved results. The effective policy and its hash include
the exception. Invalid selectors and unknown fields fail before analysis. At most
100 exceptions are allowed, with at most 100 paths and 100 symbols per exception.
Unmatched exceptions remain in the policy; this alpha does not reject unused entries.

For a test file containing many deliberately invalid input strings, omit `symbols`
and select `kinds: [string]`. Its comments remain checked. For a runtime file,
prefer a named fixture or catalog field so neighboring strings remain checked.
The repository's [.unswell.yaml](../.unswell.yaml) demonstrates both cases.

These exceptions select source regions, not individual rule findings. They apply
to library calls and explicit CLI files as well as recursive scans. Source must
still parse successfully. To omit a file from recursive discovery entirely, use
`files.exclude`; explicit file arguments bypass discovery patterns.

Markdown code, quotations and other protected syntax follow the normal extraction
policy described in [input formats](inputs.md). Fine-grained finding suppressions
and baseline debt handling are tracked separately in the [roadmap](roadmap.md).
