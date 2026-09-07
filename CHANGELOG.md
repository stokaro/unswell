# Changelog

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

This alpha does not provide calibrated probabilities, a rule DSL, suppressions,
baselines, changed-unit analysis, dependency parsing, or stable-rule precision
claims. See `docs/roadmap.md` for the remaining specification requirements.
