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
its supported contexts: text has `paragraph`; Markdown and MDX have `paragraph`, `heading`,
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
an [embedded program](inputs.md#programs-in-shell-literals). GitHub Actions run
values use the shell handling described below; other grammars
keep such data until an exception selects it. Select known data with existing
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

## GitHub Actions scripts

Unswell uses a shell grammar for run steps in GitHub Actions workflows. This
applies to `jobs.*.steps[*].run` in direct `.yml` and `.yaml` children of
`.github/workflows`. YAML folding, chomping, and escapes are decoded first;
findings still point to original YAML bytes.
Comments and quoted messages remain checked, while shell operators and command
words stay outside prose metrics. Ordinary YAML values, including action inputs
named `run`, retain their existing behavior.

The shell comes from the step, job defaults, or workflow defaults. Without an
explicit shell, a container selects sh; recognized literal hosted-runner labels
select Bash on Linux/macOS and PowerShell on Windows. Supported shell names are
`bash`, `sh`, `zsh`, `fish`, `pwsh`, and `powershell`. Simple templates such as
`bash -e {0}` also work. Unknown shells or runner labels produce
`actions-shell-unknown`; scripts containing unresolved `${{ ... }}` produce
`actions-expression`. The whole script is excluded in those cases, with its
reason recorded. This does not claim complete embedded analysis. Other prose in
the file remains checked and can pass the normal gate.

YAML string contexts and exceptions apply before shell parsing. The nested
shell's context override then selects comments and strings inside the script.
Returned blocks remain YAML `string` contexts. An exception targeting the YAML
`run` field omits the entire script; an exception targeting the nested shell
format selects its comments or strings using the same workflow path.

To select ordinary YAML scalar analysis explicitly:

```yaml
version: 1
extraction:
  github_actions: strings
```

Set the mode to `shell` for the default. Unknown modes are config errors. A change
to this mode changes the policy and model input hashes. Check older measurements
before reusing them with the new contract. A known script with invalid syntax
causes an error; it never falls back to scalar analysis. A reasoned YAML exception
still skips nested parsing. See [ADR 0039](adr/0039-actions-shell-extraction.md) for
supported runner labels, template limits, and coverage details.

## Extraction outcomes

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
| A recognized Actions script with a known shell | Comments and strings checked; program syntax excluded with `actions-shell-syntax` |
| An Actions script with an unknown shell or unresolved expression | Whole scalar excluded with `actions-shell-unknown` or `actions-expression` |
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

## Structured source prose

Go comment prose uses Go doc-comment block grammar by default. Paragraphs,
headings, and list-item paragraphs are separate analysis blocks. Indented code
examples are excluded. This also applies to ordinary Go comments; it does not
interpret them as Markdown. Go doc comments support flat lists. Use
`extraction.go_comments: plain` to retain the previous comment behavior.

Select static string literals that contain Markdown explicitly:

```yaml
extraction:
  go_comments: godoc
  markdown_strings:
    - id: cli-help
      paths: [internal/cli/**/*.go]
      formats: [go]
      symbols: [Long]
```

Each selector requires a unique ID and paths. Formats and enclosing symbols
are optional and use the same matching rules as extraction exceptions. At most
100 selectors are allowed. Unselected strings keep their current behavior.
The language decoder runs before the existing Markdown grammar. Paragraphs,
nested lists, inline protection, and fenced or indented examples retain their
original source coordinates, including escapes and CRLF input.

The source `string` context and extraction exceptions apply first. The Markdown
context set then selects inner blocks. Returned block kinds remain `string`;
structure records the inner Markdown kind. GitHub Actions `run` routing takes
precedence; select `github_actions: strings` to disable that routing explicitly.
Protected interpolation and control characters make an explicitly selected
Markdown string an operational error. The parser cannot establish the structure
of a complete Markdown document from a dynamic fragment.

These settings change policy and preparation identities. Paragraph and list
boundaries change the units sent to rules and features. Remeasure older corpus
artifacts and model bindings explicitly; replacing a stored identity is not a
measurement. See the [decision](adr/0040-source-prose-boundaries.md) and the
[source prose regression](../e2e/testdata/source_prose/).

These exceptions select source regions, not individual rule findings. They apply
to library calls and explicit CLI files as well as recursive scans. Source must
still parse successfully. To omit a file from recursive discovery entirely, use
`files.exclude`; explicit file arguments bypass discovery patterns.

Markdown code, quotations and other protected syntax follow the normal extraction
policy described in [input formats](inputs.md). Reasoned [source suppressions](suppressions.md)
permit specific rule findings while retaining raw evidence. Baseline debt handling
remains a separate item in the [roadmap](roadmap.md).
