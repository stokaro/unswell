# Changelog

## 0.1.0-alpha.12

This release makes the report of `unswell check` say what to do at each mark
of the AI-sounding index, and it carries the measurement of a model that
fixes a page by that report.

The report of `unswell check` now says at each mark of the AI-sounding index
what to do with it: the replacement that `unswell fix` would make, such as
"write a comma" or "remove the bold", or, for a stock word, that only the
writer can say what the thing does. Before, a line gave the place of a mark
and general advice stood below the list, which left a person or a model
that follows the report free to pick any punctuation. One line now says that
`unswell fix --write` makes the replacements.

With this report a model's fix was measured once more, on 131 pages that no
earlier run had read and with the bars written first. Blind model judges
find its page the less AI-sounding one against the original on 111 of 113
decisive pages, and against a page that the same model fixed with the
findings of write-good on 73 of 108, 67.6% (58.3% to 75.7%). With the
earlier report that comparison held on one sample and failed on the next.

## 0.1.0-alpha.11

This release makes `unswell fix` leave alone the pages that the index calls
low, and it carries two more measurements of the command on pages that no
earlier one had read. One claim of alpha.10 did not hold up there and is
withdrawn.

`unswell fix` now leaves a page at the level `low` as it is. `unswell check`
lists no marks on such a page, and the command was measured only on pages with
an index of 40 or more, yet it replaced every em dash and arrow of any page
that had an index. `--include-low` restores that. On pages with an index of 40
or more the command writes what it wrote before. The JSON report gains
`left_at_low` on a document and in the summary.

Two more measurements on pages that no earlier run had read. On 89 pages
with the published alpha.10 binary and on 250 with the build of this
release, the index is lower after `unswell fix` on 85 and 247 pages, and
blind model judges find its page the less AI-sounding one against the
original on 67 of 68 and 231 of 233 decisive pages. Against a page that a model
fixed with the findings of write-good, the best third-party linter in these
runs, the page after `unswell fix` is the less AI-sounding one on
146 of 209, 69.9% (63.3% to 75.7%), with the bar written
before the run. One earlier claim does not hold up: that a model following
the report of `unswell check` beats the same model following write-good held
on the first sample (48 of 70) and not on the second (39 of 72). The
documentation no longer makes it.

## 0.1.0-alpha.10

This release turns the tool toward its purpose. `unswell check` gives each
Markdown page an index of the marks that people remove when they clean
AI-sounding wording out of documentation, `unswell fix` removes those marks
without a model, and six rules of lean wording that report on pages people
wrote leave the defaults. The review asks the reviewer for a verdict on every
item.

`unswell check` now gives every Markdown page of 150 prose words or more an
AI-sounding index from 0 to 100, with the levels `low`, `elevated`, and
`high`. It counts em dashes, emoji in headings and at the start of list
items, arrows, stock words, and bold labels in lists, per 1,000 prose words,
and the report lists the place of each mark and how to remove it. On held-out
pages the index separates pages before a de-slop commit from README files
written before November 2022 with an AUC of 0.867, and 3.2% of those README
files are `high`. The index has no gate and is not a probability
([ADR 0059](docs/adr/0059-page-index-and-fix.md),
[documentation](docs/ai-sounding-index.md)). The JSON report gains
`documents[].ai_sounding`.

The new command `unswell fix` removes the marks that have a mechanical
replacement: an em dash becomes a comma, a period between two full clauses,
or a colon after a short list label; an emoji leaves a heading or the start
of a list item; an arrow becomes `->`; a bold label at the start of a list
item loses its asterisks. It calls no model and writes only with `--write`.
On 83 held-out pages the index fell on 83, and a panel of two blind model
judges found the fixed page the less AI-sounding one against the original on
72 of 73, 98.6% (92.6% to 99.8%) of the decisive pages. The judges
often name these marks as their reason, and nobody judged the grammar of the
fixed text: read the change before you keep it. In Go,
`unswell.RemoveAIMarks` applies the fixes that each `AIMark` carries.

Six rules of `unswell check` now ship off: `filler.wordy-phrase`,
`syntax.long-sentence`, `filler.instruction-scaffolding`, `hype.vague-praise`,
`repetition.adjacent-word`, and `format.decorative-emoji`. Each reports on more
than one in twenty of 121 pages that people wrote before November 2022, and
none of them follows likely-manual de-slop edits
([ADR 0058](docs/adr/0058-default-rules-quiet-on-human-pages.md)). With the
rules that ship on, 9.1% of those pages have a finding where 50.4% had one.
To keep the six, use `extends: [builtin:strict-v1]` or enable them by name.

`unswell review` now asks the reviewer for a verdict on every item, not only
for the items to revise. A reviewer that lists only what it wants changed
picks differently on each run: two runs on the same 40 files repeated 77.4%
of each other's flagged sentences, and 83.9% with the verdict form, which
also flags more of the words that a de-slop commit then changed
([ADR 0057](docs/adr/0057-reviewer-judges-every-item.md)). An answer that
judges fewer than four in five items fails and is asked again. The answer is
longer, about a line per sentence. In Go, `Document.ParseFlags` replaces
`review.ParseFlags`.

The MCP tool `unswell_review_submit` takes the same answer: `items` with one
verdict per item replaces `flags`, and a submission that judges fewer than
four in five items is refused. The measurement behind the change was made
through `unswell review`, not through the MCP tools.

`unswell review` and `unswell trim` now say when an endpoint cut an answer off
at its output limit, and do not ask again: set a larger `--max-tokens`. Before,
such an answer was reported as malformed JSON after every attempt.

`unswell check` has two new default rules. `format.decorative-emoji` reports
emoji in Markdown headings and list items, and leaves alone paragraphs,
tables, code, check marks, and crosses. `format.em-dash-frequency` reports a
document once when it holds at least three em dashes and at least three per
1,000 words; `onset` and `allowed_occurrences` change the limits. Neither has
a gate. Both were tested on 597 files before and after commits whose message
says that they remove AI-sounding wording
([ADR 0056](docs/adr/0056-rules-tested-on-de-slop-commits.md)). The emoji
findings are lower after the commit in 77.7% (69.8% to 84.0%) of the pairs
that differ, and the listed em dashes in 70.4% (64.1% to 76.0%). On the
same pairs the default rules of alpha.9 give 64.0% (56.1% to 71.2%)
and fire in 158 pairs; with the new rules they give
72.7% (67.8% to 77.1%) and fire in 365.

`scaffold.chat-preamble` now also reports a chat opener in the first sentence
after the leading headings of a document, as in a README whose title is
followed by `Sure, here is a summary`. Before, a title hid the opener. The
rule reports nothing on 3,302 pages of public repositories. It does not
look after later headings, where a person may answer a question with `Sure,`.
The new phrase position is `prose-start`; `document-start` keeps its meaning.

The limits that `unswell rules show scaffold.chat-preamble` prints now say
that in source code the first comment or string of a file counts as the start
of a document, and that a file of chat replies belongs in the extraction
policy. On 265,025 source files of open-source packages the rule reports
nothing.

## 0.1.0-alpha.9

This release makes the two model commands read more answers, adds one chat
opener to the default rules, and records the first measurement of a part of
the tool against human editors.

`unswell review` and `unswell trim` now read an answer in which the model
corrects itself. A model sometimes writes its JSON object, notices a mistake,
and writes the object again. Such an answer failed the document; the last
object is now the answer. A correction that is cut off is still an error.

`unswell trim` was measured against human editors
([ADR 0054](docs/adr/0054-trim-measured-against-editors.md)). On answers of
chat models that experts edited, the words it removes are words the expert also removed or
replaced in 71.2% and 68.0% of cases with two models, against 62.7% and 59.3%
by chance in the same sentences. Per removed word the text moves 0.13 and
0.15 words nearer to the expert's version than under random removal from the
same sentences. On fiction and on documentation written by people the
registered rule is not met. See [trim](docs/trim.md#measured-behavior).

`scaffold.chat-preamble` now also reports a document that opens with `Sure,`
or `Sure!`, as in `Sure, here is a summary`. Experts who edited answers of chat
models removed that opener in 39 of 39 test answers
([ADR 0055](docs/adr/0055-phrase-listed-on-editor-reference.md)). The rule
has `gate: forbid`, so such a document now fails the check; the business
profile leaves the rule off as before.

## 0.1.0-alpha.8

This release adds two commands that work with a model you name, `unswell
review` and `unswell trim`. The rules and their defaults are those of alpha.7
with one phrase restored.

`unswell trim` asks a model for sentences with empty words removed and keeps a
proposal only if a deterministic check passes: no word added, replaced, or
moved, at least half of the sentence left, and no negation, number, modal
word, code, link, quotation mark, or emphasis marker removed. Only sentences
that lie in one piece of source are offered. Files change only with `--write`.
See [trim](docs/trim.md) and
[ADR 0053](docs/adr/0053-trim-admitted-by-model-consensus.md).

`unswell review` reads documents with a model you name: an endpoint that
implements the OpenAI chat completions API, or a local program that takes a
prompt on standard input. The model flags items under the rubric of the MCP
review tools. The command binds each flag to source bytes and drops a flag
that does not quote its item. A second model, named with the `--confirm-`
flags, can judge the bound flags; the text report then shows the confirmed
ones. Findings are advisory and the command exits 0 when the review completed.
A document that cannot be reviewed makes the run exit 2. See
[review in context](docs/review.md) and
[ADR 0051](docs/adr/0051-review-command.md).

Both commands send the selected documents to the model their user names and
to nothing else. `unswell check`, the Go library, the GitHub Action, and the
MCP server still send nothing anywhere.

`unswell review` was measured against the sentences that people edited and
against another model's annotation. `unswell trim` was judged by a consensus
of models only; no comparison with human editors existed when this version
was released, and alpha.9 records the first one. The numbers, with their
limits, are in the documentation of
[review](docs/review.md#measured-behavior) and
[trim](docs/trim.md#measured-behavior), in
[ADR 0052](docs/adr/0052-claims-follow-human-references.md), and in the
[record](research/reviews/2026-11-12-human-reference/README.md), whose check
recounts every number.

The new public package `github.com/stokaro/unswell/review` holds the items, the
rubric, the prompts, the binding, and the trim check. It makes no model call. The MCP review
tools now use it. `unswell_review_request` returns an empty item list for a
document without prose; it returned an error before.

`hype.vague-praise` reports `comprehensive guide` again. It had left the list
on four accepted findings of six, which is not strong evidence.
[ADR 0050](docs/adr/0050-listed-phrases-leave-on-strong-evidence.md) now asks
for the upper end of the 95% interval to be below 80% before a listed phrase
leaves. Without that condition the lists would lose a few phrases with every
new test set and never regain one. Configuration hashes change.

## 0.1.0-alpha.7

This release publishes what alpha.6 could not. The Go module and the container
images of `v0.1.0-alpha.6` exist and work, but its release run stopped when
Docker Hub rejected the mirror credential, so alpha.6 has no archives and no
MCP registry entry.

The Docker Hub mirror no longer gates a release. A rejected credential still
fails the mirror job and the run, and the archives, the registry entry, and the
distribution updates are published without it. Run the
`Mirror containers to Docker Hub` workflow for the version once the credential
works again. The weekly `Release prerequisites` workflow now logs in to Docker
Hub with that credential, so an expired token shows before a release.

`unswell check` now uses the available processors, at most eight, when `--jobs`
is not given. It used one worker, so a large repository could exceed the
default analysis timeout of 30 seconds and end with an operational error. On
this repository the check takes 4 seconds in place of 12. The result does not
depend on the number of workers.

`hype.vague-praise` no longer reports `state-of-the-art` and
`comprehensive guide`, and `filler.wordy-phrase` no longer reports `foster` and
its forms. Over all reviewed pages two reviewers accepted fewer than 80% of
their findings, which is the listing criterion. Configuration hashes change.

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
