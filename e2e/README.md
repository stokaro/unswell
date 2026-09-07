# End-to-end CLI cases

Run `go test ./e2e -count=1` from the repository root. The suite builds the real CLI
with `CGO_ENABLED=0`, then launches it in a fresh directory for each scenario.
The native CI matrix and `make check` run this package through `go test ./...`.
No installed Unswell binary, external service, or separate test framework is needed.

## Read a case

Each directory in `testdata` contains:

- `case.json`: source filenames, expected process exit code, and optional CLI modes.
- `sample.go.txt` or another language fixture: source with inline `want` annotations.
- `policy.yaml`, when the case overrides the shared two-rule policy.
- `diagnostics.golden.json`: exact diagnostics, source spans, completion, and gate state.
- `stdout.golden` and `stderr.golden`: complete user-facing process output.

For example, the [Go fixture](testdata/go/sample.go.txt) places expectations beside
the comment and string that should trigger them:

```go
// It is important to note that the client retries. // want "filler.announced-importance"
const message = "Certainly! The client opens connections." // want "scaffold.chat-preamble"
```

The annotation lists exact rule IDs expected on that source line. Multiple quoted
IDs represent multiple findings. Every finding must consume one expectation; missing,
extra, duplicate, and wrong-line findings fail the test. Unannotated prose is a
negative expectation. The harness rejects malformed annotations.

Before scanning, the harness replaces annotation bytes with spaces to preserve
offsets without feeding expected diagnostics into prose analysis. Fixture suffixes
keep source samples out of Go package discovery; the CLI receives the original
language filename. CRLF and BOM cases materialize those bytes explicitly on every OS.

## Coverage

The cases cover all 18 supported input formats, comments and strings, clean rewrites,
Markdown code protection and entities, escaped Unicode, CRLF/BOM coordinates,
interpolation boundaries, global/per-language context replacement, reasoned symbol
exceptions, directory exclusions, stdin, malformed input, invalid configuration,
empty input, and `--no-gate` behavior. Exit codes 0, 1, and 2 are explicit expectations.

The shared policy limits language-extraction cases to two phrase rules and sets
chat-preamble matching to sentence starts. The `document_position` case checks its
document-start setting separately. This suite does not claim every catalog rule or
cancellation behavior is covered; rule examples and failure-path tests also live
in the root, `extract`, `report`, and `mcp` packages.

JSON goldens retain diagnostic messages and exact byte/Unicode coordinates while
omitting unrelated catalog and build metadata. The suite checks SARIF results against
those same validated diagnostics, confirms incomplete scans are unsuccessful in
SARIF, checks source privacy, and verifies input files remain unchanged.

## Review golden changes

After an intentional behavior change, run:

```sh
go test ./e2e -run TestCLI -update -count=1
git diff -- e2e/testdata
go test ./e2e -count=1
```

Review expected rule IDs, source ranges, messages, and gate behavior before accepting
the diff. `-update` still checks annotations, exit codes, source coordinates, and
SARIF parity first. Ordinary runs never create or rewrite missing goldens.

The MCP process tests use the official client and remain in the separate MCP module.
Repository dogfooding also compares CLI and MCP results through the actual protocol.
