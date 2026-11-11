# Changelog

## 0.1.0-alpha.6

This release publishes the content of alpha.5. The `v0.1.0-alpha.5` tag exists,
but its release did not complete and its Go module cannot be downloaded. Use
alpha.6.

The root Go module no longer contains the research records. They are now a
module of their own, `github.com/stokaro/unswell/research`, which holds no Go
packages. The source tree of the root module shrinks from 674 MB to 7 MB. The
Go module proxy refuses a source tree above 500 MB, so `go get` and
`go install` failed for alpha.5. `make check` now fails when the root module
exceeds 64 MB.

The release workflow resolves the pinned mirror tool before it publishes
anything, and a weekly workflow does the same. The tool's registry had removed
the pinned image, which stopped the alpha.5 release after its container images
were published. The job that checks this repository through the published
action now checks `README.md` and `docs`: the action uses the default analysis
timeout of 30 seconds, and the whole repository needs longer on a hosted
runner.

## 0.1.0-alpha.5

This release bases rule defaults on findings that two assistant reviewers
judged on real documentation pages, and adds review by the calling assistant
through the MCP server.

Rule defaults now rest on reviewed findings. `hype.vague-praise` ships enabled
and reports qualified words such as `powerful` and `seamless`.
`filler.wordy-phrase` reports stock verbs such as `leverage` and `utilize`.
`hype.modifier-cluster` counts more evaluative modifiers.
`repetition.sentence-openers` and `repetition.paragraph-openers` are opt-in;
enable them in configuration to keep their findings. These changes alter
configuration hashes and reported findings. The
[qualification record](docs/research/rule-qualification.md) reports precision by
rule, a final test on 772 real pages, and the limits of the measurement. Its
labels, reviews, and manifests are in
`research/corpus/2026-11-11-assistant-panel`, without page text.

One Markdown block that the grammar rejects no longer hides its file. The block
is excluded with the reason `unparsed`, the rest of the file is analyzed, and the
run stays incomplete with an error that names the file and line. This covers
HTML inside list items, code spans with angle brackets in tables, and tables the
grammar builds wrongly.

Default severity now follows reviewed precision. `syntax.long-sentence`,
`filler.instruction-scaffolding`, `filler.unscoped-assurance`, and
`hype.vague-praise` report notes. Rules whose findings two reviewers accepted at
least 85% of the time on real pages keep warnings. Use `--min-severity warning`
to see only those. Scores and the gate are unchanged.

The defaults report fewer findings that reviewers rejected.
`syntax.long-sentence` starts above 45 words, up from 35, and no longer
measures table cells; set `onset: 35` to restore the earlier threshold.
`filler.instruction-scaffolding` no longer reports a conditional reader goal or
a reminder that opens with `Be sure to`. `repetition.adjacent-word` requires
both copies in the same case with no markup between them, and reports a note.
`filler.document-metadiscourse` and `filler.unscoped-assurance` are opt-in;
enable them in configuration to keep their findings. On 578 new real pages the
default profile reports 517 findings in place of 1,174, and two reviewers
accept 77% of them, up from 64%.
[ADR 0049](docs/adr/0049-defaults-follow-reviewed-forms.md) states the rule.

MDX files now participate in directory scans and MCP checks. Markdown prose
between component tags is checked; imports, exports, expressions and tag
attributes retain explicit exclusions. Original byte coordinates survive
masking, indentation, CRLF and Unicode. Use `languages.mdx.contexts` to select
headings, paragraphs, list items or table cells. The added default file pattern
changes configuration hashes; existing diagnostic and feature values are unchanged.

The MCP server adds `unswell_review_request` and `unswell_review_submit`. The
calling assistant receives the editorial rubric and the document's sentences,
judges them in context, and submits flags. The server binds each quote to
original byte coordinates and rejects a quote that crosses protected code. These
findings are advisory, are labeled as the calling assistant's judgments, and
change neither scores nor the gate.
On 120 real pages an assistant of another model family accepted 70.1% of the
flags that two calling assistants raised this way; the
[qualification record](docs/research/rule-qualification.md) has the measurement.

Building from source now requires Go 1.26. The Go analysis adapter and its
consumer use `golang.org/x/tools` v0.50.0, which requires this compiler version.
CI tests all runtime and consumer modules with automatic upgrades disabled.

## 0.1.0-alpha.4

This release repairs extraction errors found while checking real repositories
with alpha.3. TypeScript type-only wildcard re-exports and deeply nested Markdown
lists now parse. Markdown indentation after a byte-order mark is preserved.
The Markdown repair uses a pinned `stokaro/gotreesitter` fork while its upstream
pull request is pending.

Go comments now preserve Go doc-comment paragraphs and list items. Indented
examples are excluded from prose. Set `extraction.go_comments: plain` to retain
plain extraction. Explicit `extraction.markdown_strings` selectors apply the
Markdown grammar to decoded string literals, preserving source coordinates and
protected code. A selected string with unresolved interpolation fails analysis.

GitHub Actions `run` steps now use the selected shell grammar after YAML decoding.
This keeps shell syntax out of prose findings. Shell selection follows workflow,
job, and step settings; unresolved expressions and unknown shells have explicit
exclusion reasons. Set `extraction.github_actions: strings` to check those values
as generic YAML strings.

Shared features and NLP preparation respect non-Latin exclusions. Extraction
retains a Unicode character whose bytes map to several source spans. A Markdown
link without a visible label keeps a boundary between neighboring words. The
paragraph-opener rule applies its minimum word count to the whole paragraph.

These fixes change extraction, preparation, and rule identities. Saved corpus
results keep the policy that produced them. Rerun affected inputs before joining
them with new results or model packs.

This alpha makes no precision or calibration claim. Issue #219 tracks repeated
scans of the published binary against the pinned performance corpora.

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
