# End-to-end CLI cases

Run `go test ./e2e -count=1` from the repository root. The suite builds the real CLI
with `CGO_ENABLED=0`, then launches it in a fresh directory for each scenario.
The native CI matrix and `make check` run this package through `go test ./...`.
No installed Unswell binary, external service, or separate test framework is needed.

## Read a case

Each directory in `testdata` contains:

- `case.json`: source filenames, expected process exit code, and optional CLI modes.
- Optional `resources` in `case.json`: local config and dictionary files copied without becoming source arguments.
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

Use `want-suppressed "rule.id"` for a raw finding that must have a valid source
permission. A plain `want` requires an active finding. A mismatch in either
direction fails, even when the rule ID and source line are correct.

Before scanning, the harness replaces annotation bytes with spaces to preserve
offsets without feeding expected diagnostics into prose analysis. Fixture suffixes
keep source samples out of Go package discovery; the CLI receives the original
language filename. CRLF and BOM cases materialize those bytes explicitly on every OS.

## Coverage

The cases cover all 18 supported input formats, comments and strings, and clean
rewrites. Source-mapping cases include Markdown code protection and entities,
escaped Unicode, CRLF/BOM coordinates, and interpolation boundaries. Policy cases
check context replacement, reasoned symbol exceptions, and directory exclusions.
Process cases cover stdin, malformed and empty input, invalid configuration, and
`--no-gate`. Exit codes 0, 1, and 2 are explicit expectations.

The [empty table cells](testdata/markdown_empty_cells/sample.md.txt) case came from
the Ptah evaluation. It preserves detections beside empty cells and protected code,
including Unicode and Markdown emphasis in a file with CRLF and a BOM.

The [empty table rows](testdata/markdown_empty_rows/sample.md.txt) case requires
a detection on line 5 after a blank data row. Its
[CRLF variant](testdata/markdown_empty_rows_crlf/sample.md.txt) covers repeated rows,
different column counts, a BOM, Unicode, emphasis and protected code boundaries.
The [orphaned delimiter](testdata/markdown_orphaned_delimiter/sample.md.txt) case
requires exit code 2 for a known partial grammar tree, preserving the difference
between an unsupported boundary and a completed clean scan.

The shared policy limits language-extraction cases to two phrase rules and sets
chat-preamble matching to sentence starts. The `document_position` case checks its
document-start setting separately. This suite does not claim every catalog rule or
cancellation behavior is covered; rule examples and failure-path tests also live
in the root, `extract`, `report`, and `mcp` packages.

JSON goldens retain diagnostic messages and exact byte/Unicode coordinates while
omitting unrelated catalog and build metadata. The suite checks SARIF results against
those same validated diagnostics, confirms incomplete scans are unsuccessful in
SARIF, checks source privacy, and verifies input files remain unchanged.

The custom-rule cases load inline packs through the real CLI. They cover matching
Go comments and escaped strings with CRLF, and paragraph conditions with exact
regex locations, Markdown emphasis, a BOM, protected code, and a local exception.
Their annotated findings and JSON/SARIF output use the same checks as builtin rules.

The `config_inheritance` case loads a parent policy and shared term dictionary,
applies two overlapping file overrides, and retains the unexempted occurrence beside
a valid technical term. Its nested Markdown sources include emphasis, protected code,
CRLF, and a BOM. `config_cycle` requires a startup error before any report is written.
The harness verifies that source files, policy, and dependency resources stay unchanged.

The `suppression_*` cases cover sentence and Go region permissions, unused and
misspelled directives, complete multi-location evidence, and explicit file-wide
permission. Goldens retain directive locations, reasons, targets, used IDs, and
the suppressed state of raw findings. SARIF must mark exactly those findings as
accepted source suppressions with reasons. CRLF, a BOM, and Markdown emphasis
exercise the same source mapping as ordinary detections.
The [empty-row suppression case](testdata/suppression_after_empty_rows/guide.md.txt)
checks block and inline permissions after a table whose empty rows require parser
normalization. Their locations still refer to original bytes, and the final
unpermitted finding must fail the gate.

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
