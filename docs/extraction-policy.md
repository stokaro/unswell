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

## Literal data

Selecting strings includes embedded SQL, JSON, scripts, identifiers, and protocol
values. Unswell does not infer that a string is prose from its variable name.
In the shell formats only, a literal whose lines are mostly code is excluded as
an [embedded program](inputs.md#programs-in-shell-literals); every other grammar
keeps such data until an exception selects it. Select known data with existing
path, format, kind, and symbol exceptions. For example:

```yaml
version: 1
extraction:
  exceptions:
    - id: sql-query
      paths: ["src/database.go"]
      formats: [go]
      kinds: [string]
      symbols: [query]
      reason: "The query value is SQL consumed by the database driver."
    - id: workflow-shell
      paths: [".github/workflows/*.yaml"]
      formats: [yaml]
      kinds: [string]
      symbols: [run]
      reason: "Run fields contain shell programs rather than English prose."
```

These examples retain other strings and comments in the same files. A symbol
names an enclosing declaration or keyed field; selecting a function selects all
its string literals. Review that scope before using a function-wide exception.
Use global or per-language context sets when the intended policy applies to every
string in that scope. Disabling strings is an explicit choice, not the default.

The following cases have distinct outcomes:

| Input | Outcome |
| --- | --- |
| Source bytes are invalid UTF-8 or contain NUL | Operational error before selection |
| Python byte literal or YAML non-string scalar | Existing typed-data exclusion |
| A selected literal decodes to non-UTF-8 bytes | Whole literal excluded with `non-utf8-literal` and its original byte range |
| Byte escapes form valid UTF-8 together | Text remains selected, with escape source mapping |
| A selected block is predominantly non-Latin prose | Block excluded with `non-latin-prose`; other blocks remain checked |
| Valid text contains embedded code or fixed values | Checked unless an explicit exception selects it |
| A shell string, heredoc or here-string whose lines are mostly code | Whole literal excluded with `embedded-program` and its original byte range |
| A tag or example section of a JSDoc or Javadoc comment | Lines excluded with `doc-tag` or `doc-example`; the description keeps its block |
| A comment that starts with a listed tool prefix | Whole comment excluded with `directive` |
| Source has invalid grammar | Operational error even when an exception would select its strings |
| A selected literal has an unsupported escape | Operational error during decoding; an excluded literal is not decoded |

A matching explicit exception takes precedence over the decoded-byte and
embedded-program exclusions and retains the configured ID and reason. A
disabled string context does the same with `config:context-disabled:string`.
Valid control escapes retain protected
boundaries; they do not turn neighboring words into one phrase. UTF-8 validity
does not establish that text is English. A selected block with at least 20
letters, more than half of them outside the Latin script, becomes a
`non-latin-prose` exclusion, and the engine checks the rest of the document;
see [input formats](inputs.md). When exclusions leave no applicable prose, the
normal empty-scan gate applies, including CLI exit code 2 by default.

The [literal regression cases](../e2e/testdata/non_utf8_literals/) preserve checked
neighbors and exclusion ranges in CLI goldens. The
[data-only case](../e2e/testdata/non_utf8_only/) requires an incomplete result.
The [non-Latin block case](../e2e/testdata/non_latin_block/) keeps the English
paragraphs checked and records the excluded block.

These exceptions select source regions, not individual rule findings. They apply
to library calls and explicit CLI files as well as recursive scans. Source must
still parse successfully. To omit a file from recursive discovery entirely, use
`files.exclude`; explicit file arguments bypass discovery patterns.

Markdown code, quotations and other protected syntax follow the normal extraction
policy described in [input formats](inputs.md). Reasoned [source suppressions](suppressions.md)
permit specific rule findings while retaining raw evidence. Baseline debt handling
remains a separate item in the [roadmap](roadmap.md).
