# Input formats

Unswell checks English prose for formulaic AI-style wording before it reaches
code or documentation. Syntax grammars locate the prose; the editorial rules
and English NLP backend analyze that text locally.

| Format | Filename examples | Prose regions |
| --- | --- | --- |
| `text` | `notes.txt` | Paragraphs |
| `markdown` | `README.md`, `guide.markdown` | Headings, paragraphs, list items and GFM table cells |
| `go` | `client.go` | Comments, interpreted strings and raw strings |
| `javascript` | `app.js`, `view.jsx`, `module.mjs`, `module.cjs` | Comments, quoted strings and static template fragments |
| `typescript`, `tsx` | `client.ts`, `module.mts`, `view.tsx` | Comments, quoted strings and static template fragments |
| `python` | `app.py`, `types.pyi` | Comments, strings, docstrings and static f-string fragments |
| `rust` | `lib.rs` | Line and block comments, quoted and raw strings |
| `java` | `Client.java` | Comments, strings and text blocks |
| `c`, `cpp` | `client.c`, `client.cpp`, `client.hpp` | Comments, strings and C++ raw strings |
| `csharp` | `Client.cs`, `script.csx` | Comments, quoted, verbatim and raw strings, and static interpolation fragments |
| `yaml` | `config.yaml`, `workflow.yml` | Comments, plain and quoted string values, literal and folded block scalars |
| `bash`, `sh` | `check.bash`, `check.sh`, `.bashrc`, `.profile` | Comments, quoted strings and heredoc text |
| `zsh` | `check.zsh`, `.zshrc`, `.zprofile` | Common shell comments and quoted strings through the Bash grammar |
| `fish` | `check.fish` | Comments, single-quoted strings and static double-quoted fragments |
| `powershell` | `check.ps1`, `module.psm1`, `data.psd1` | Comments, quoted strings and here-strings |

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

Original UTF-8 byte coordinates survive CRLF, escaped Unicode, Markdown entities,
emphasis and removed delimiters. A decoded character points to its full original
escape. Interpolation expressions and command substitutions insert protected
boundaries so words on either side cannot form a false phrase. Strings nested
inside an interpolation are checked as separate literals.

Go retains its standard parser's comment grouping and generated-file, directive
and cgo metadata. The syntax tree validates source structure and supplies strings.
Import paths and Go struct tags are excluded as technical metadata. Recognized
source directives and Python byte literals are also recorded as exclusions.

YAML keys, aliases and resolved non-string values are recorded as exclusions.
Anchored strings are checked where they are defined; aliases are not expanded.
The grammar supplies source spans. The existing YAML decoder supplies scalar
values, including folding, chomping and explicit tags, without constructing
application objects. A disagreement between the two parsers fails analysis.
YAML embedded scripts remain string values; they are not executed or recursively parsed.

Choose checked contexts globally or per format with the
[extraction policy](extraction-policy.md). Omitted settings check every supported
prose context; language overrides replace that set.

## Alpha limits

The pinned backend is gotreesitter 0.52.0. Grammar errors, partial parses and
unsupported string escapes produce incomplete analysis and CLI exit code 2.
This is not a substitute for a compiler or a language-specific linter.

A lone `|` after a Markdown table can leave orphaned delimiter tokens in the
pinned grammar's tree ([#69](https://github.com/stokaro/unswell/issues/69)). Unswell
rejects that incomplete tree instead of skipping the following text. Rows with
two or more pipes support empty cells normally.

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
