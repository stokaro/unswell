# Public API policy

All public packages are experimental during `v0.1.0-alpha.*`. Pin the exact release.
An alpha suffix does not excuse unannounced incompatible changes: compare API
changes against the preceding release and document intentional differences.

| Import path | Purpose | Stability |
| --- | --- | --- |
| `github.com/stokaro/unswell` | Engine, options and versioned results | alpha |
| `github.com/stokaro/unswell/document` | Sources, byte spans and neutral document model | alpha |
| `github.com/stokaro/unswell/rule` | Rule contract, typed parameters and evidence | alpha |
| `github.com/stokaro/unswell/builtin` | Explicit builtin catalog factory | alpha |
| `github.com/stokaro/unswell/config` | Strict in-memory policy compiler | alpha |
| `github.com/stokaro/unswell/extract` | Source-preserving input extractors | alpha |
| `github.com/stokaro/unswell/nlp` | Neutral provider and capability contract | alpha |
| `github.com/stokaro/unswell/nlp/english` | Default pure-Go English backend | alpha |
| `github.com/stokaro/unswell/report` | Writers and saved-result decoding | alpha |

Packages under `internal/` are implementation details. `cmd/unswell` is an
executable, not an importable library. The tools module isolates build dependencies
from the runtime module and its minimum compiler. The consumer module is an
executable public-API contract test, not another supported library.
The separate `mcp/` module provides the `unswell-mcp` executable and its protocol
tests. It calls the public engine and keeps the MCP SDK out of the core module.

Callers own returned result slices. Source bytes must not change during analysis.
One engine supports concurrent calls. Custom rules and providers are trusted Go
code and must honor the concurrency and read-only view contracts.

The engine reads no configuration paths, environment variables, current directory,
standard streams or network endpoints. The CLI supplies source bytes and policy.
Reporters consume an existing result and return write errors to the caller.
