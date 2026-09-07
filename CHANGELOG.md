# Changelog

## 0.1.0-alpha.1

First experimental release of an offline Go library and CLI for English prose.
It extracts plain text, Markdown/GFM and Go comments with original source ranges.
The engine provides 16 configurable rule IDs, explainable local scores, policy
gates and six builtin profiles. One analysis produces text, JSON, SARIF, HTML and
Markdown reports.

This alpha does not provide calibrated probabilities, a rule DSL, suppressions,
baselines, changed-unit analysis, dependency parsing, or stable-rule precision
claims. See `docs/roadmap.md` for the remaining specification requirements.
