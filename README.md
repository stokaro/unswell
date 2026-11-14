# Unswell

A linter for lean English, built to reduce AI-sounding prose.

Unswell's main motivation is to keep formulaic AI-style wording from leaking into
source code and documentation. AI-assisted work often leaves behind conversational
preambles, generic claims of importance, inflated modifiers and repeated setup.
Unswell finds these patterns in prose and explains what to revise before the text
reaches a commit or a published document.

The checks apply to text regardless of its author. They enforce a chosen editorial
policy through concrete findings and explainable scores. The checker runs locally as
a Go library and CLI and sends source text nowhere. Two further commands,
`unswell review` and `unswell trim`, read a document with a model you name; see
[Review and trim with a model](#review-and-trim-with-a-model).

Unswell is not an authorship detector. A clean scan of a generated text is a
correct result, and a finding on a human-written text is a finding about the
wording. No score is a probability that a tool wrote the text; see
[ADR 0037](docs/adr/0037-diagnostics-not-authorship.md) for the goal and its
non-goals.

**Development alpha.** The published release is `v0.1.0-alpha.11`. Stage 2 features
include custom rules, baselines, and checks of changed code. Go tools can train,
calibrate, and load models.

Rules and defaults remain experimental. No editorial model has met the criteria
for use in the product. See the [acceptance audit](docs/acceptance.md) for features, release
evidence, and work still open.
Revision probability is unavailable until a labeled corpus and compatible
calibration model meet the later acceptance criteria.

## What it does

`unswell check` applies fixed rules offline: announcements, stock phrases,
repetition, a page that leans on em dashes. The strict profile adds wordy
constructions, long sentences, and emoji in headings. It explains each finding,
gives the same result on every run, and can fail a build on wording your
project has decided to refuse.

`unswell check` also gives each Markdown page an
[AI-sounding index](docs/ai-sounding-index.md) from 0 to 100 and lists the
marks behind it: em dashes, emoji in headings and list items, arrows, stock
words, bold labels in lists. On pages from before a commit that removes
AI-sounding wording, set against README files written before November 2022,
the index separates the two with an AUC of 0.87, and 3.2% of those
README files reach its high level. `unswell fix` removes the marks that have
a mechanical replacement, without a model. On 250 pages that no earlier
measurement had read, blind model judges found the page after `unswell fix`
the less AI-sounding one against the original on 231 of 233 decisive pages,
and against a page fixed with the findings of a third-party linter on
146 of 209. A model that follows the report of `unswell check` gave
111 of 113 and 73 of 108 on 131 more such pages.

`unswell review` gives a document to a model you name. The model reads it in
context and flags sentences to tighten. Unswell binds each flag to the exact
place in the source and drops a flag that does not quote its sentence.

`unswell trim` asks a model for sentences with empty words removed and keeps a
proposal only if it deletes words and nothing else. It changes a file only when
you ask.

On answers that chat models gave to instructions and that experts then
edited, the words that `unswell trim` removes are words the expert also
removed or replaced in 71.2% and 68.0% of cases with two trimming models,
against 62.7% and 59.3% for random words of the same sentences. Removing
them brings the text nearer to the expert's version than removing random
words of the same sentences does, by 0.13 and 0.15 words per removed word.

How each part was measured is in the documentation of
[review](docs/review.md#measured-behavior) and
[trim](docs/trim.md#measured-behavior).

## Quick start

Build with the minimum Go version declared in [go.mod](go.mod):

```sh
go build -o bin/unswell ./cmd/unswell
bin/unswell config init --profile technical
bin/unswell check README.md docs/
bin/unswell fix README.md docs/
```

Install the published alpha without Go through the
[Homebrew tap](https://github.com/stokaro/homebrew-unswell):

```sh
brew install stokaro/unswell/unswell
```

A separate [GitHub Action](https://github.com/stokaro/unswell-action) runs pinned
release archives and writes JSON/SARIF reports. See [installation and release
verification](docs/installation.md) for availability and platform checks.

For example, this introduction:

```text
It is important to note that the client retries a request only after a connection failure.
```

produces `filler.announced-importance`. A direct version preserves the condition:

```text
The client retries a request only after a connection failure.
```

Use the strict profile to make the introductory phrase a hard policy violation:

```sh
printf '%s\n' 'It is important to note that the client may retry.' |
  bin/unswell check --stdin --filename draft.md --profile strict
```

## Review and trim with a model

Fixed rules name constructions. `unswell review` asks a model you name to read
each document in context, binds its flags to source bytes, and drops a flag
that does not quote its item:

```sh
bin/unswell review README.md docs \
  --endpoint https://api.example.com/v1/chat/completions \
  --model example-model --api-key-env REVIEW_API_KEY
```

A local program that reads a prompt on standard input works as well, and a
second model can shorten the list.

`unswell trim` proposes sentences with words removed:

```text
docs/guide.md:3:1: trim
    - These features significantly improve the ability to type-check your code.
    + These features improve the ability to type-check your code.
```

A proposal is shown only if a deterministic check passes: no word is added,
replaced, or moved, at least half of the sentence remains, and no negation,
number, modal word, code, or link is removed. `--write` applies the trims to the
files.

Both commands send the selected documents to the model you name and to nothing
else. Their output is advisory and never fails a build. See
[review in context](docs/review.md) and [trim](docs/trim.md).

## Reports and gate

One analysis can write all five reports:

```sh
bin/unswell check README.md docs/ \
  --report text:- \
  --report json:artifacts/unswell.json \
  --report sarif:artifacts/unswell.sarif \
  --report html:artifacts/unswell.html \
  --report markdown:artifacts/unswell.md
```

Add `--include-source` to retain snippets and highlighted paragraphs in saved
reports. Source text is omitted by default. Reports can contain internal prose;
choose artifact access accordingly.

```sh
bin/unswell report artifacts/unswell.json --format html --output artifacts/review.html
bin/unswell explain README.md --at 3:1
bin/unswell rules list
bin/unswell rules show repetition.near-sentence
bin/unswell rules test
bin/unswell config validate
bin/unswell config explain --file README.md
bin/unswell doctor
```

Exit codes: `0` for a complete pass, `1` for a complete policy failure, `2` for
operational errors or incomplete analysis, and `130` for cancellation. An empty
scan fails unless explicitly allowed. `--no-gate` permits advisory checks while
preserving operational errors. Display filtering never changes the gate. A rule
that exhausts its candidate budget abstains on that document with the reason
`budget_exhausted`; the report lists the abstention and the other rules' findings
stay.

The index uses points from 0 to 100. Every local score includes its contributing
rules, activations, deduplication decisions and caps. A clean paragraph cannot
reduce another paragraph's score. The tool makes no authorship or factuality claim.

## Input and configuration

Inputs include plain text, Markdown/GFM, MDX, and comments and string literals in Go,
JavaScript, TypeScript/TSX, Python, Rust, Java, C/C++, C#, YAML, Bash/POSIX sh, Zsh's common
shell syntax, Fish and PowerShell. Markdown uses block and inline grammars.
See [formats and extraction limits](docs/inputs.md) for extensions and dialect details.
For stdin, provide `--filename` or a supported `--format`. Coordinates use original
UTF-8 byte ranges and one-based Unicode code point columns.

Recursive Git scans select tracked files. Explicit files may be untracked and
bypass recursive include/exclude patterns. Outside Git, directories use a bounded
filesystem walk. Inputs must remain within the current project root. Symlinks are
not followed during discovery. Unsupported explicit formats fail.

The CLI reads the nearest `.unswell.yaml` up to the project root, or the exact file
named by `--config`. It reads no home configuration. Local inheritance, ordered file
overrides, and shared term dictionaries use one
[configuration plan](docs/configuration.md) in the CLI, library, and MCP server.
Add team policies with [custom rule packs](docs/custom-rules.md), load them through
`--ruleset` or inline `rule_sets`, and execute their examples with `unswell rules test`.
Select checked contexts globally or per language with the
[extraction policy](docs/extraction-policy.md). Use reasoned exceptions for intentional
examples, dictionaries and other text that should remain outside an editorial check.
Exceptions can select paths, languages, comments or strings, and named string owners.

AI assistants can call the same engine through the separate [MCP server](docs/mcp.md).
Its tools check supplied drafts and expose the fixed policy. `make dogfood-mcp`
verifies repository checks through a real MCP client and subprocess.
Separate [CLI and MCP containers](docs/containers.md) support Linux AMD64 and ARM64,
including repository checks through both interfaces in CI. Both images are public
in GHCR, and Docker Hub mirrors the release with matching digests.

```yaml
version: 1
extends: [builtin:technical-v1]
language: en
calibration:
  model: none
rules:
  filler.announced-importance:
    enabled: true
    severity: error
    gate: forbid
  policy.banned-phrases:
    parameters:
      phrases: ["leverage synergies"]
      positions: [any]
```

A phrase position is `any`, `sentence-start`, `document-start`, `prose-start`
(the first sentence after the leading headings), or `document-end`.

Available profiles are `technical`, `strict`, `minimal`, `business`, `reference`
and `custom`; each also accepts a `-v1` suffix. Unknown settings, duplicate keys,
unsupported rule parameters and unavailable model names fail before analysis.

## Go library

```go
engine, err := unswell.New(unswell.Options{})
if err != nil {
    return err
}
result, err := engine.Analyze(ctx, document.Source{
    Name: "draft.md",
    Format: document.Markdown,
    Bytes: draft,
})
if err != nil {
    return err
}
return report.Write(writer, "json", result, report.Options{})
```

The [external consumer](examples/consumer) implements a custom rule using only
public packages. See the [API policy](docs/public_api.md) for supported imports and
ownership contracts. Go rules are trusted code and must support concurrent calls.

## Development and limits

Unswell checks its own README, documentation, Go code and Bash scripts with the committed
[strict repository policy](.unswell.yaml). Run `make dogfood`, or `make check` for
all checks. CI runs the same gate and uploads all five report formats. A negative
CLI probe must fail with exit code 1, proving the policy is active. Third-party
licenses and test-data documents are outside this editorial policy; inline code
and fenced examples use the normal extractor's protected boundaries.
The policy excludes deliberate fixture strings and catalog data with recorded reasons.
Comments and runtime strings remain checked. Bash also passes syntax, ShellCheck
and shfmt checks, including negative probes for each gate.

Reasoned [source suppressions](docs/suppressions.md) permit individual rule findings
while retaining raw evidence, scores, and an audit record across CLI and MCP.
Explicit [baselines](docs/baseline.md) accept existing debt without removing findings
or changing scores. The `new` gate checks new findings and changed scored units.
Committed [changed-unit checks](docs/changes.md) compare merge-base to HEAD with
clean-source verification and retain whole-document repetition context.
Use [trusted policy](docs/trusted-policy.md) to apply merge-base rules and baseline
while auditing candidate policy changes and new source permissions.
The public [Go analysis adapter](docs/go-analysis.md) and its prototype vet driver
check comments and strings through the same engine.
The [contextual editorial catalog](docs/editorial-patterns.md) adds opt-in filler,
hype, and rhetorical-pattern checks with explicit evidence and technical counterexamples.
The [repetition catalog](docs/repetition-signals.md) adds bounded phrase, POS-template,
paragraph, heading, and summary comparisons with measured overlap evidence.
The [surface catalog](docs/surface-signals.md) adds opt-in syntax, readability, and
formatting measurements with explicit counting protocols and protected boundaries.
The research pipeline that fits and evaluates those models is a separate
developer workflow; see [training and evaluation](docs/training.md).
The active [roadmap](docs/roadmap.md) improves contextual diagnostics and checks
them on real technical prose. Human qualification and calibrated revision
probabilities remain deferred under ADR 0037, with their original requirements
preserved. Neither capability is claimed by this alpha.

Code, URLs, front matter, directives and quoted Markdown blocks are excluded from
ordinary prose checks. POS-based chunks are surface candidates, not grammatical
dependencies. Lexical overlap is not a claim of semantic equivalence. Thresholds
are initial editorial policy, not measured precision or confidence.

Code is MIT licensed. Runtime dependencies and the embedded English model have
their own notices. No scan downloads rules, sends prose to a service or requires
Python, Node.js, a server, an API key or cgo.
