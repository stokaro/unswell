# Changelog

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
