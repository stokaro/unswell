# Changelog

## 0.1.0-alpha.3

Third experimental release, cut because the second one never published. The
release run of the alpha.2 tag stopped in the mirror job and published
nothing. The pinned digest of the Skopeo image had disappeared from its
registry. This release carries the alpha.2 content with the mirror tool
pinned to the current index of the same version.

The engine now records a block of mostly non-Latin prose as an exclusion with
the reason `non-latin-prose`. It checks the other blocks of the document, and
the run stays complete. Before, such a block failed the whole document.

Extraction records three more exclusions. A shell string, heredoc, or
here-string whose lines are mostly code is an `embedded-program`. A tag
section of a JSDoc or Javadoc comment is a `doc-tag` or a `doc-example`. A
comment that starts with a known tool prefix is a `directive`, the name that
replaces `go-directive` and `source-directive`.

A rule that exhausts its candidate budget on one document now abstains
there. The result lists the abstention, the other rules keep their findings,
and the run stays complete.

The English provider splits a sentence after a dotted identifier or version
such as `chi.Router.`, and its identity version is
`prose-v3.2.1/chunks-v1/boundaries-v1`. Version 4 of `syntax.noun-stack`
ends a run at a configured verb form. Version 2 of `syntax.parenthetical-load`
skips an insertion shorter than `min_insertion_words`. The grade metric, at
version 2, grades prose words only. Feature and policy hashes change with
these versions.

## 0.1.0-alpha.2

Second experimental release of the same offline engine, CLI, MCP server, and
Go library.

Extraction keeps Markdown prose after lone-pipe table boundaries, around
byte-valued literals, and at block edges at the start and end of a file. It
accepts function-like macro statements in C and C++. Noun stacks stop at a
mistagged "cannot". Near-sentence matching keeps technical contrasts. Empty
environment values in Bash commands extract correctly. The parser timeout
scales with the input, document-heavy trees get bounds on memory and time,
and the saved-result bound rises to 256 MiB.

The CLI checks committed changes with full paragraph context under a trusted
merge-base policy. It reports an opt-in origin estimate in a separate channel
that gates nothing. It estimates revision probability from a configured pack.
It collects prepared prose features and rule activations through the CLI and
MCP. SARIF output is checked against the schema and a real consumer. A public
Go analysis adapter and vet driver check comments and strings through the same
engine.

A release now dispatches update pull requests for the Homebrew tap and the
GitHub Action, mirrors the images to Docker Hub with digest checks, and
reproduces the six release binaries across hosts.

The research corpus and its records under `research/` change no product
behavior. This alpha still has no calibrated probability gate, no stable-rule
precision claim, and no human-labeled corpus.

## 0.1.0-alpha.1

First experimental release of an offline Go library and CLI for English prose.
The project was created to reduce AI-sounding text and keep formulaic AI-style
wording out of source code and documentation.
It extracts Markdown/GFM, source comments and string literals through gotreesitter
grammars, with original source ranges and reasoned configuration exceptions.
Plain text remains supported. Source formats include Go, JavaScript/TypeScript/TSX,
Python, Rust, Java, C/C++, C#, YAML, Bash/POSIX sh, common Zsh syntax, Fish and PowerShell.
Global context sets and per-language replacements select comments, strings and
document prose. Reasoned exceptions can narrow selection by path and named owner.
The engine provides 16 configurable rule IDs, explainable local scores, policy
gates and six builtin profiles. One analysis produces text, JSON, SARIF, HTML and
Markdown reports.

A separate MCP server lets AI assistants check their drafts with the same engine
and fixed policy. CI compares its results with CLI evidence and checks failure,
rewrite, cancellation and invalid-input behavior through the official MCP client.

Separate CLI and MCP images support Linux AMD64 and ARM64. Release workflows
verify public image pulls and publish the stdio MCP package to the official
MCP Registry. A dedicated Homebrew tap and GitHub Action include installation
checks and use Unswell to check their own documentation and supported source.
Public availability is established by the release verification workflows.

This alpha does not provide calibrated probabilities, a rule DSL, suppressions,
baselines, changed-unit analysis, dependency parsing, or stable-rule precision
claims. See `docs/roadmap.md` for the remaining specification requirements.
