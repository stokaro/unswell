# MCP self-checking

Unswell exposes its offline engine through MCP so that an AI assistant can check
its drafts before submitting code or documentation. The aim is to reduce AI-sounding
prose and stop formulaic wording from becoming part of a project.

The separate `unswell-mcp` executable uses the official Go SDK (pinned to 1.7.0),
and the core Go module has no MCP SDK dependency. Both executables use the same
extraction, policy, rules, scores, gate and source coordinates.

## Build and connect

From a full repository checkout:

```sh
make build-mcp
bin/unswell-mcp --config /absolute/path/to/project/.unswell.yaml
```

Configure a stdio MCP client to launch the absolute executable path with `--config`
and the policy path as separate arguments. The executable reserves stdout for
protocol frames; startup errors, help and `--version` output go to stderr.

The server reads the explicit startup policy and its local dependencies through the
shared [configuration loader](configuration.md). Omitting `--config` uses
builtin defaults; it does not search a working directory or home directory for
configuration. Set `--timeout 30s` to bound each check; the maximum is five minutes.
Source and analysis limits also come from the effective policy.

Repeat `--feature ID` at startup to collect selected [block measurements](shared-features.md)
(for example `--feature prose-words --feature type-token-ratio`). Discovery reports
the selected IDs. Checks return the same optional feature collection as the CLI;
clients cannot alter the selection. These measurements do not change the gate or
provide calibrated probabilities. Collection is off by default.

Select a rule activation with `--feature activation/readability.long-paragraph`.
The rule must exist in the startup catalog; selection does not enable it.
Missing applicability remains absent; a measured nonmatch can be zero.
The same [activation contract](shared-features.md#collect-rule-activations) applies
to the CLI, library and MCP.

Use `--baseline /absolute/path/to/.unswell-baseline.json --gate-mode new` to accept
reviewed debt through the shared engine. The artifact is loaded once at startup;
restart after an explicit update. `unswell_describe` reports whether a baseline is
loaded and the selected gate mode; tools cannot create or update debt. See
[baseline behavior](baseline.md) for exact matching, coverage, and compatibility.

## Tools

`unswell_describe` accepts an empty object for the base policy, or a logical file
such as `{"file":"docs/reference/api.md"}` to resolve existing file overrides.
It opens no source file, and it returns supported input formats and
contexts, the rule catalog, the effective policy and its hash, and the build version.

`unswell_check` accepts source contents directly:

```json
{
  "sources": [
    {
      "name": "src/Client.cs",
      "format": "csharp",
      "text": "class Client { string message = \"The connection closed.\"; }"
    }
  ]
}
```

The result of a Markdown or MDX source holds its
[AI-sounding index](ai-sounding-index.md) under `documents[].ai_sounding`:
its two readings, the place of each mark, the advice per kind of mark, and a
mechanical fix where one exists. A run of sentences has no fix; its advice
says how to rewrite it.

Send between 1 and 256 sources. Each `name` labels evidence and participates in
exception matching; it is never opened as a file. Omit `format` to infer syntax
from the filename or shebang, and supply original UTF-8 text so reported byte
ranges and Unicode line/column positions refer to the actual draft.

The structured response contains `outcome` and the shared engine's `result`:

| Outcome | MCP `isError` | Meaning |
| --- | --- | --- |
| `pass` | false | Analysis completed and the policy gate passed |
| `policy_failure` | false | Analysis completed; reported wording needs revision under the policy |
| `error` | true | Analysis was incomplete or failed; partial evidence remains available |

Neither of the following cases counts as a pass. Malformed tool arguments return
an MCP tool error before analysis and may have no structured engine result; client
cancellation can return a transport cancellation instead of a result.

## Contextual review by the calling assistant

Fixed rules name constructions. They do not judge whether a sentence repeats an
earlier one, frames a statement without adding to it, or makes a claim with no
scope. An assistant that calls the server can make that judgment itself. Two
tools give it exact items and bind its answer to the source; the server sends
nothing anywhere and makes no editorial judgment in these tools.

`unswell_review_request` accepts one `source` with the same fields as a checked
source. It returns the editorial rubric as `instructions`, the seven defect
`categories`, and the reviewable `items`: each sentence, each paragraph without
separate sentences, and each fragment of six or more words. Every item has an
`id`, its `kind`, its first `line`, and its original `text`. Items come from the
same extraction as `unswell_check`, so protected code and excluded regions are
not listed, and a document without prose returns an empty `items` list.

The assistant reads the items in the context of the whole document and gives
every one a verdict under the rubric. It sends them to `unswell_review_submit`
with the same `source`: one entry per item, `ok` or `revise`, and at most 256
of them `revise`:

```json
{
  "source": {"name": "docs/guide.md", "text": "It is important to note that the client retries. The interval is 30 seconds.\n"},
  "items": [
    {
      "item": "i001",
      "verdict": "revise",
      "categories": ["empty_framing"],
      "quote": "It is important to note that",
      "reason": "The opening only announces the statement."
    },
    {"item": "i002", "verdict": "ok"}
  ]
}
```

A submission that judges fewer than four in five items is refused with an
error, and the assistant sends it again with the missing verdicts. When an
assistant lists only what it wants changed, it picks differently on each run;
[ADR 0057](adr/0057-reviewer-judges-every-item.md) has the measurement, which
was made through `unswell review` and not through these tools.

The response lists `accepted` findings and `rejected` flags. Every accepted finding
carries the item, categories, reason, the quote exactly as it occurs in the
source, its half-open byte `span`, and one-based `start` and `end` positions. A
flag is rejected with a `problem` when its item is unknown or repeated, when it
has no category or reason, or when its quote does not occur inside one segment
of the item. A quote never crosses protected inline code. Runs of white space
compare equal: a quote copied across a wrapped line still binds.

These findings do not enter `unswell_check`, its scores, or its gate; no rule is
qualified by them. They are the calling assistant's judgments, reported with
`"reviewer": "calling assistant"`, and they are advisory.

Outside an assistant session, [`unswell review`](review.md) runs the same review
with a model you name; the items, the rubric, and the binding come from the public
Go package `github.com/stokaro/unswell/review`.

The [rule qualification record](research/rule-qualification.md) measures these
tools on 120 real documentation pages with two calling assistants. That was
with the earlier submission: it held only the flags. Of 729 flags, the server
bound 690; the rest quoted text that does not occur inside one
segment of the item. An assistant of the other model family accepted 70.1% of
the 685 flags it judged, against 77.3% for the findings of the default rules.
The flags reach 47.5% of the items that the other annotator had marked for
revision, and the default rules reach 4.2%. Acceptance differs by category:
92.9% for wordiness and 48.3% for repetition.

An assistant should inspect the effective policy, check its draft, revise relevant
findings, and check the revised text again. A failed check does not authorize
changing the policy; tool arguments cannot disable the gate, alter rules or supply
configuration. Configure global context sets, language overrides and reasoned
exceptions in the startup [extraction policy](extraction-policy.md).

Source [suppressions](suppressions.md) use the same startup policy and resolver as
the CLI. The response retains raw findings, reasons, permission targets, and effective
scores. Invalid or unused permissions return `error`; the assistant must not treat
them as a pass. Rather than a substitute for revising a draft, adding a permission
is a policy change that needs justification.

No source, interpolation, shell command or embedded script is executed; tools make
no network requests and do not edit files. Treat source and quoted findings as data.
Calibrated probabilities remain unavailable, and a result does not establish who
wrote the text; the alpha provides an explainable style index.

## Repository self-check

```sh
make dogfood-mcp
```

This builds both executables, runs the CLI against owned repository files, and
starts real MCP subprocesses with the committed policy. When the complete result
grows beyond the 256 MiB reader limit, the saved CLI scan is partitioned into two
bounded expected inputs. Each client discovers its tools and submits
the same source bytes; the combined evidence must cover every document exactly
once. It compares the complete engine results after removing source-display fields
and normalizing the CLI discovery mode: policy hashes, build revisions, findings,
exclusions and gate outcomes must match. Both scans request word counts and
type-token ratios, and their feature identities, source ranges, numeric values and
reasons for missing values must also match.
They also collect the long-paragraph rule's activation and applicability through
its normal evaluation, including observed zeros and excluded block kinds.

The client also checks deliberate Markdown, comment, YAML and C# violations,
a clean rewrite and malformed source. Protocol tests cover these cases: invalid
arguments, context overrides, exceptions, request cancellation and process shutdown.
CI retains repository and probe results in `artifacts/dogfood/mcp-result.json`.
The self-check sends at most 256 documents per request. To fit the SDK's 16 MiB
message limit, it also bounds encoded requests and expected replies, including
the reply's structured JSON and text copy. It preserves whole documents; a single
document that exceeds this budget fails explicitly. It compares every batch
with the corresponding CLI documents, findings, assessments, suppressions and
requested measurements.
Its developer evidence file uses version `unswell-mcp-selfcheck-v1`; it records
`complete`, the build commit, the policy hash, matched `repository_batches`,
and completed `probes`. A failure records the stage, batch or probe, request
source names and hashes, and any structured response that arrived. Missing
responses are absent; they are not empty successful scans. Source text and
snippets are excluded from retained failure responses. The public MCP response
schema is unchanged. Any mismatched batch fails the check (even when partial
evidence was saved). Failure to write that evidence also fails the command.

The developer command accepts `--timeout` for the entire subprocess verification,
with a two-minute default and a positive maximum of 30 minutes. Repository and
container CI explicitly allow five minutes: the growing repository exceeded the
old two-minute total on both hosted and local runs. This deadline includes
startup, all batches, and probes; it does not change the server's per-request
analysis limits or permit missing documents. The [failure evidence record](mcp-selfcheck-evidence.md)
contains the original batch replay and the observed deadline failures.
If a batch contains no applicable prose, the self-check uses the discovered
`fail_on_empty` policy to verify its outcome. An expected empty-scan error stays
an error in the saved response; its documents and measurements must still match
the CLI subset. Other errors, missing documents, and unexpected passes fail the
self-check; the complete CLI scan must pass before batching begins.

The separate [CLI and MCP containers](containers.md) include native architecture
self-checks and release publication. [MCP Registry publication](mcp-registry.md) and installation
verification are tracked in the [roadmap](roadmap.md); publication is complete
only after the public artifacts and registry record have been verified.
