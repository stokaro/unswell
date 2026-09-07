# Unswell

A linter for lean English, built to reduce AI-sounding prose.

Unswell's main motivation is to keep formulaic AI-style wording from leaking into
source code and documentation. AI-assisted work often leaves behind conversational
preambles, generic claims of importance, inflated modifiers and repeated setup.
Unswell finds these patterns in prose and explains what to revise before the text
reaches a commit or a published document.

The checks apply to text regardless of its author. They enforce a chosen editorial
policy through concrete findings and explainable scores. Unswell runs locally as
a Go library and CLI, without sending source text to an AI service.

**Development alpha.** The current implementation targets stages 0–1 of the
[technical roadmap](docs/roadmap.md). Rules and defaults are experimental.
Revision probability is unavailable until a labeled corpus and compatible
calibration model meet the later acceptance criteria.

## Quick start

Build with the minimum Go version declared in [go.mod](go.mod):

```sh
go build -o bin/unswell ./cmd/unswell
bin/unswell config init --profile technical
bin/unswell check README.md docs/
```

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
preserving operational errors. Display filtering never changes the gate.

The index uses points from 0 to 100. Every local score includes its contributing
rules, activations, deduplication decisions and caps. A clean paragraph cannot
reduce another paragraph's score. The tool makes no authorship or factuality claim.

## Input and configuration

Inputs include plain text, Markdown/GFM, and comments and string literals in Go,
JavaScript, TypeScript/TSX, Python, Rust, Java, C/C++, C#, YAML, Bash/POSIX sh, Zsh's common
shell syntax, Fish and PowerShell. Markdown uses block and inline grammars.
See [formats and extraction limits](docs/inputs.md) for extensions and dialect details.
For stdin, provide `--filename` or a supported `--format`. Coordinates use original
UTF-8 byte ranges and one-based Unicode code point columns.

Recursive Git scans select tracked files. Explicit files may be untracked and
bypass recursive include/exclude patterns. Outside Git, directories use a bounded
filesystem walk. Inputs must remain within the current project root. Symlinks are
not followed during discovery. Unsupported explicit formats fail.

The CLI reads `.unswell.yaml` in its current project root, or the exact file named
by `--config`. It reads no home configuration. The alpha accepts one builtin profile
and explicit rule overrides; local inheritance and file overrides belong to stage 2.
Select checked contexts globally or per language with the
[extraction policy](docs/extraction-policy.md). Use reasoned exceptions for intentional
examples, dictionaries and other text that should remain outside an editorial check.
Exceptions can select paths, languages, comments or strings, and named string owners.

AI assistants can call the same engine through the separate [MCP server](docs/mcp.md).
Its tools check supplied drafts and expose the fixed policy. `make dogfood-mcp`
verifies repository checks through a real MCP client and subprocess.

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

The [roadmap](docs/roadmap.md) preserves the remaining requirements: the custom
rule DSL, full suppressions, baseline, committed changed-unit checks, trusted
policy comparison, Go analysis integration and calibrated revision probabilities.
The alpha does not silently claim those capabilities.

Code, URLs, front matter, directives and quoted Markdown blocks are excluded from
ordinary prose checks. POS-based chunks are surface candidates, not grammatical
dependencies. Lexical overlap is not a claim of semantic equivalence. Thresholds
are initial editorial policy, not measured precision or confidence.

Code is MIT licensed. Runtime dependencies and the embedded English model have
their own notices. No scan downloads rules, sends prose to a service or requires
Python, Node.js, a server, an API key or cgo.
