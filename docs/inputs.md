# Input formats

Unswell checks English prose for formulaic AI-style wording before it reaches
code or documentation. Syntax grammars locate the prose; the editorial rules
and English NLP backend analyze that text locally.

| Format | Filename examples | Prose regions |
| --- | --- | --- |
| `text` | `notes.txt` | Paragraphs |
| `markdown` | `README.md`, `guide.markdown` | Headings, paragraphs, list items and GFM table cells |
| `go` | `client.go` | Comments, interpreted strings and raw strings |
| `javascript` | `app.js`, `view.jsx`, `module.mjs`, `module.cjs` | Comments, quoted strings and static template fragments; JSDoc tag and example lines are excluded |
| `typescript`, `tsx` | `client.ts`, `module.mts`, `view.tsx` | Comments, quoted strings and static template fragments; tag and example lines of `/**` comments are excluded |
| `python` | `app.py`, `types.pyi` | Comments, strings, docstrings and static f-string fragments |
| `rust` | `lib.rs` | Line and block comments, quoted and raw strings |
| `java` | `Client.java` | Comments, strings and text blocks; Javadoc tag lines are excluded |
| `c`, `cpp` | `client.c`, `client.cpp`, `client.hpp` | Comments, strings and C++ raw strings |
| `csharp` | `Client.cs`, `script.csx` | Comments, quoted, verbatim and raw strings, and static interpolation fragments |
| `yaml` | `config.yaml`, `workflow.yml` | Comments, plain and quoted string values, literal and folded block scalars |
| `bash`, `sh` | `check.bash`, `check.sh`, `.bashrc`, `.profile` | Comments, quoted strings and heredoc text; program-shaped literals are excluded |
| `zsh` | `check.zsh`, `.zshrc`, `.zprofile` | Common shell comments and quoted strings through the Bash grammar; program-shaped literals are excluded |
| `fish` | `check.fish` | Comments, single-quoted strings and static double-quoted fragments; program-shaped literals are excluded |
| `powershell` | `check.ps1`, `module.psm1`, `data.psd1` | Comments, quoted strings and here-strings; program-shaped literals are excluded |

Use `--format` to select a grammar explicitly. For example, a C++ header named
`client.h` needs `--format cpp`; `.h` defaults to C. Extensionless inputs can use
a recognized shebang. Pass them explicitly or add their paths to `files.include`.
Shebangs are inspected as text. No source, interpreter, substitution or generator
is executed. Recursive discovery uses the extension and shell startup-file patterns
in the effective configuration.

Markdown uses gotreesitter's block and inline grammars, including GFM tables and
task lists. Inline code, fenced and indented code, HTML, image syntax, link
destinations and front matter are protected. Link labels remain prose. Quoted
blocks are excluded unless `analysis.include_quotes` is enabled.

Empty GFM data rows retain the table-cell context of following rows. Blank lines
and other block boundaries still end the table. The adapter uses a mapped block
parser input to work around the pinned scanner's empty-row handling; inline prose
and reported ranges always come from the original source.

Markdown inputs may end without a final newline. The block parser receives
mapped framing to retain the first block's grammar wrapper and finish the last
block. It never becomes source text or part of a reported range. Headings retain
their structure at EOF, and leading protected blocks retain their exclusions.

Original UTF-8 byte coordinates survive CRLF, escaped Unicode, Markdown entities,
emphasis and removed delimiters. A decoded character points to its full original
escape. Interpolation expressions and command substitutions insert protected
boundaries so words on either side cannot form a false phrase. Strings nested
inside an interpolation are checked as separate literals.

C and C++ accept a function-like macro used as a statement with a block and
no semicolon, such as `list_for_each(pos, head) { ... }`. The call is blanked
in a copy, so the block keeps its offsets. That holds only when the call
contains no comment or string; any other incomplete tree stays an error.

Go retains its standard parser's comment grouping and generated-file, directive
and cgo metadata. The syntax tree validates source structure and supplies strings.
Import paths and Go struct tags are excluded as technical metadata.

Every region that extraction leaves out is recorded on the document with its
byte span and a reason. Saved reports carry that list. The reasons include the
following.

| Reason | Region |
| --- | --- |
| `directive` | A shebang or a [tool directive comment](#directive-comments) |
| `comment-code` | An indented example in a comment |
| `doc-tag`, `doc-example` | A line in a tag or example section of a [documentation comment](#documentation-comments) |
| `embedded-program` | A shell literal that [holds a program](#programs-in-shell-literals) |
| `byte-literal` | A Python byte literal |
| `non-utf8-literal` | A literal whose decoded bytes are not UTF-8 |
| `interpolation` | An expression inside a string |
| `config:<id>: <reason>`, `config:context-disabled:<context>` | A region that a configured exception or context set leaves out |

For a `non-utf8-literal` the full literal range is omitted from prose;
neighboring comments and text strings remain checked. Byte escapes that together
form UTF-8, such as Go `"\xc3\xa9"`, remain text. Invalid source encoding still
fails before extraction, and an input containing only excluded data still reaches
the empty-scan gate.

YAML keys, aliases and resolved non-string values are recorded as exclusions.
Anchored strings are checked where they are defined; aliases are not expanded.
The grammar supplies source spans. The existing YAML decoder supplies scalar
values, including folding, chomping and explicit tags, without constructing
application objects. A disagreement between the two parsers fails analysis.
YAML embedded scripts remain string values; they are not executed or recursively parsed.
SQL, JSON, identifiers, and fixed protocol strings also remain selected text unless
the project supplies a [reasoned exception](extraction-policy.md#literal-data).

Choose checked contexts globally or per format with the
[extraction policy](extraction-policy.md). Omitted settings check every supported
prose context; language overrides replace that set.

## Documentation comments

In JavaScript, TypeScript, TSX and Java, a `/**` comment follows the JSDoc and
Javadoc tag grammar. The description before the first tag is a prose block. A
line that starts with `@name` opens a tag section that runs to the next tag or
the end of the comment. Tag sections are excluded line by line: the lines of an
`@example` carry the reason `doc-example`, the lines of every other tag carry
`doc-tag`. A `@` that does not start a line, a `/*` comment and a line comment
do not open a tag section. Prose written after a tag, such as under
`@description` or `@remarks`, is not checked.

## Directive comments

A comment that starts with a prefix from the following list is an instruction
for a tool. It is excluded with the reason `directive` and ends the comment
block around it. The match is exact and case-sensitive; a prefix that ends in a
letter or digit must be followed by a character that is not one, so `nolint`
does not match `nolinting`. A sentence that mentions a tool is prose. The list
is fixed; there is no pattern matching on other words.

| Format | Prefixes |
| --- | --- |
| every format | a shebang; `-*-`, `@generated`, `Code generated`, `SPDX-`, `coding:`, `coding=`, `unswell-`, `vi:`, `vim:` |
| `go` | `#nosec`, `+build`, `go:`, `line `, `lint:`, `nolint` |
| `javascript`, `typescript`, `tsx` | `@flow`, `@jsx`, `@jsxFrag`, `@jsxImportSource`, `@jsxRuntime`, `@ts-check`, `@ts-expect-error`, `@ts-ignore`, `@ts-nocheck`, `biome-ignore`, `c8 ignore`, `deno-fmt-ignore`, `deno-lint-ignore`, `eslint-disable`, `eslint-enable`, `eslint-env`, `istanbul ignore`, `jscs:`, `jshint `, `jslint `, `oxlint-`, `prettier-ignore`, `tslint:`, `v8 ignore` |
| `python` | `flake8:`, `fmt:`, `isort:`, `mypy:`, `noqa`, `nosec`, `pragma:`, `pylint:`, `pyright:`, `pytype:`, `ruff:`, `type: ignore`, `yapf:` |
| `java` | `$NON-NLS`, `@formatter:`, `CHECKSTYLE`, `NOPMD`, `NOSONAR`, `checkstyle:`, `noinspection`, `spotless:` |
| `c`, `cpp` | `GCOVR_EXCL_`, `IWYU pragma:`, `LCOV_EXCL_`, `NOLINT`, `NOLINTBEGIN`, `NOLINTEND`, `NOLINTNEXTLINE`, `clang-format`, `clang-tidy`, `cppcheck-suppress` |
| `csharp` | `<auto-generated`, `ReSharper ` |
| `yaml` | `checkov:`, `prettier-ignore`, `renovate:`, `yaml-language-server:`, `yamllint ` |
| `bash`, `sh`, `zsh`, `fish`, `powershell` | `shellcheck `, `shfmt:` |

Unswell's own `unswell-` comments are [source suppressions](suppressions.md).
They are recorded with the reason `suppression-directive`.

## Programs in shell literals

In Bash, POSIX shell, Zsh, Fish and PowerShell, a quoted string, heredoc or
here-string often holds a program for another interpreter, such as a jq filter
or a Python heredoc. Such a literal is excluded as a whole with the reason
`embedded-program`. Content that starts with a shebang is a program. Other
content is a program when it has at least three nonblank lines and at least
three in five of them look like code.

A line looks like code in these cases:

- The line ends with a brace or a semicolon, or it starts with a pipe.
- An operator such as `==`, `&&`, `->` or `$(` appears in it, or it is an assignment.
- A program keyword such as `import`, `def` or `SELECT` starts it.
- Two or more brackets, pipes, semicolons or equals signs appear in it.

A line that ends with a period, question mark or exclamation mark never counts
as code.

The test reads lines; it does not parse the program. It can keep a short or
sentence-shaped program, and it can drop a message whose lines look like code.
A configured exception or a disabled string context decides before the test
and keeps its own reason. Other grammars do not run the test; select their
embedded data with an [exception](extraction-policy.md#literal-data). An
excluded literal is not decoded, so an unsupported escape inside it is not an
error.

## Alpha limits

Each grammar parse has a time bound of five seconds plus half a second per
KiB of source. A large file under load therefore does not fail as a parse
timeout, and a runaway parse still stops. The C# grammar is the slowest:
about two seconds for a 19 KB file on an idle machine.

The pinned backend is gotreesitter 0.52.0. An owned Bash grammar instance preserves
the grammar's empty-value token before whitespace, allowing commands such as
`! A=x B= command` to parse. This also applies to the shared Bash subset used for
POSIX shell and Zsh. The adapter retains substitutions, source ranges, and parser
error checks; it does not execute shell code or modify gotreesitter's global
grammar. See the [parser comparison](../research/parsers/bash-empty-values.md).

Grammar errors, partial parses and
unsupported string escapes produce incomplete analysis and CLI exit code 2.
This is not a substitute for a compiler or a language-specific linter.

A lone `|` after a Markdown table ends the table. Following prose retains its
paragraph or list-item context, including inside quoted blocks. The scanner
adapter handles this boundary without changing source bytes. Rows with two or
more pipes still support empty cells. Incomplete trees with orphaned delimiters
remain errors; see the [boundary investigation](../research/parsers/markdown-lone-pipe.md).

There is no dedicated Zsh grammar in this backend. Zsh uses the Bash grammar;
extended Zsh syntax can fail explicitly. Bash also has upstream parsing limits,
including some inline regular expressions and mixed quoted patterns. Our shell
scripts use named regex variables and explicit version comparisons to avoid those
forms. Fish receives a synthetic final newline when one is absent; this contributes
no prose and does not change reported coordinates.

C# raw strings support delimiter widths and closing-delimiter indentation.
The pinned grammar rejects some valid escaped-brace forms in ordinary interpolated
strings, such as `$"Read {{literal}}."`. Such input fails explicitly. Regular
interpolations and raw interpolations use protected expression boundaries.

Extraction reads individual literal syntax. It does not evaluate concatenation,
formatting calls, runtime values or unquoted shell arguments. JSX text nodes are
not yet a prose context; quoted JSX attributes are supported. Standard escapes,
Unicode scalar escapes and supported surrogate pairs are decoded. Python named
Unicode escapes and some uncommon language extensions fail explicitly.

The public format list is intentionally smaller than the backend's grammar catalog.
The default Go dependency embeds that catalog; no grammar is downloaded at runtime.
Upstream `GOT_*` parser debugging variables are outside Unswell's configuration
contract and should be unset for reproducible runs.
