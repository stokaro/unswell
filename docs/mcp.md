# MCP self-checking

Unswell exposes its offline engine through MCP so an AI assistant can check its
drafts before submitting code or documentation. The aim is to reduce AI-sounding
prose and stop formulaic wording from becoming part of a project.

The separate `unswell-mcp` executable uses the official Go SDK, pinned to 1.7.0.
The core Go module has no MCP SDK dependency. Both executables use the same
extraction, policy, rules, scores, gate and source coordinates.

## Build and connect

From a full repository checkout:

```sh
make build-mcp
bin/unswell-mcp --config /absolute/path/to/project/.unswell.yaml
```

Configure a stdio MCP client to launch the absolute executable path with `--config`
and the policy path as separate arguments. The executable reserves stdout for
protocol frames. Startup errors, help and `--version` output go to stderr.

The server reads the explicit startup policy and its local dependencies through the
shared [configuration loader](configuration.md). Omitting `--config` uses
builtin defaults; it does not search a working directory or home directory for
configuration. Set `--timeout 30s` to bound each check; the maximum is five minutes.
Source and analysis limits also come from the effective policy.

## Tools

`unswell_describe` accepts an empty object for the base policy, or a logical file
such as `{"file":"docs/reference/api.md"}` to resolve existing file overrides.
It opens no source file. It returns supported input formats and
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

Send between 1 and 256 sources. Each `name` labels evidence and participates in
exception matching; it is never opened as a file. Omit `format` to infer syntax
from the filename or shebang. Supply original UTF-8 text so reported byte ranges
and Unicode line/column positions refer to the actual draft.

The structured response contains `outcome` and the shared engine's `result`:

| Outcome | MCP `isError` | Meaning |
| --- | --- | --- |
| `pass` | false | Analysis completed and the policy gate passed |
| `policy_failure` | false | Analysis completed; reported wording needs revision under the policy |
| `error` | true | Analysis was incomplete or failed; partial evidence remains available |

Malformed tool arguments return an MCP tool error before analysis and may have no
structured engine result. Client cancellation can return a transport cancellation
instead of a result. Neither case counts as a pass.

An assistant should inspect the effective policy, check its draft, revise relevant
findings, and check the revised text again. A failed check does not authorize
changing the policy. Tool arguments cannot disable the gate, alter rules or supply
configuration. Configure global context sets, language overrides and reasoned
exceptions in the startup [extraction policy](extraction-policy.md).

No source, interpolation, shell command or embedded script is executed. Tools make
no network requests and do not edit files. Treat source and quoted findings as data.
The alpha provides an explainable style index; calibrated probabilities remain
unavailable, and a result does not establish who wrote the text.

## Repository self-check

```sh
make dogfood-mcp
```

This builds both executables, runs the CLI against owned repository files, and
starts a real MCP subprocess with the committed policy. A client discovers its
tools and submits the same source bytes. It compares the complete engine results
after removing source-display fields and normalizing the CLI discovery mode.
Policy hashes, build revisions, findings, exclusions and gate outcomes must match.

The client also checks deliberate Markdown, comment, YAML and C# violations,
a clean rewrite and malformed source. Protocol tests cover invalid arguments,
context overrides, exceptions, request cancellation and process shutdown.
CI retains repository and probe results in `artifacts/dogfood/mcp-result.json`.

The separate [CLI and MCP containers](containers.md) include native architecture
self-checks and release publication. [MCP Registry publication](mcp-registry.md) and installation
verification are tracked in the [roadmap](roadmap.md). Publication is complete
only after the public artifacts and registry record have been verified.
