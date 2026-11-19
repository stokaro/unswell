# Go analysis adapter

The public `github.com/stokaro/unswell/goanalysis` module adapts the shared engine
to `golang.org/x/tools/go/analysis`. It supplies an analyzer and a compile-tested
`singlechecker` driver; the core module does not depend on Go analysis types or
the tools module. Extraction policy, source suppressions, rules, scores, and
baseline decisions are the same in the adapter and the CLI.

Build the prototype from a checkout:

```sh
cd goanalysis
go build -o ../bin/unswell-vet ./cmd/unswell-vet
```

Then run it in a Go module:

```sh
go vet -vettool=/absolute/path/to/unswell/bin/unswell-vet ./...
```

The driver uses the default technical policy. As input, it receives Go's selected
compilation files (which depend on build tags and platform). It does not scan
Markdown, YAML, ignored Go files, or the repository directory tree. The
`files.include` and `files.exclude` settings govern CLI discovery; the adapter
receives its source selection from Go. Exit behavior belongs to `go vet` or the
chosen Go analysis driver; Unswell CLI exit codes do not apply.
This module is a prototype built from source and has no separate release tag yet.
The local replacement in its `go.mod` uses this checkout; its pinned core dependency
also resolves to a public commit when that replacement is removed.

## Configure an analyzer

Integrations can supply an engine built with configuration bytes or a complete
configuration bundle. The adapter never discovers `.unswell.yaml` or loads policy
paths; a custom driver owns that operation. Comments and strings use the existing
context set, including per-language overrides and extraction exceptions.

```go
engine, err := unswell.New(unswell.Options{
    AllowEmpty: true,
    Config: []byte("version: 1\nextraction:\n  contexts: [comment]\n"),
})
if err != nil {
    return err
}
analyzer, err := goanalysis.New(goanalysis.Options{
    Engine: engine,
    Root: projectRoot,
    Context: ctx,
})
```

Use `[comment, string]` to check both contexts. The default adapter creates its
engine once with empty prose allowed (because valid Go files may contain only code).
A supplied engine keeps its own empty-input policy. No global mutable engine is
created; one analyzer can share its immutable engine across concurrent passes.
Custom rule and NLP implementations must honor the core concurrency contract.

`Root` is optional: an absolute directory used only to derive logical source paths.
It does not restrict package loading by the driver. With Root, every supplied file
must be inside that directory; paths relative to it select file overrides and
baseline identities. Without Root, source names are per-pass file basenames.
Duplicate logical names are errors. Generated files supplied outside that root also
require an explicit integration decision; the adapter does not silently drop them.
Set a stable project Root when sharing policy or baseline paths with a CLI scan.

## Gate and diagnostics

By default, diagnostics come from one source: the reasons in a failing completed
engine gate. Advisory findings, accepted baseline debt, and source-suppressed
findings remain in the result without failing the driver. `NoGate` also remains an
engine decision; severity does not become a second gate policy in the adapter.
Setting `Options.ReportAll` reports all unsuppressed findings (including advisory
findings and accepted debt). A Go driver may then fail even if the engine gate
passes.

Every diagnostic includes its rule ID twice: as Category and in its message. No
suggested fixes or automatic rewrites are generated. Primary and related ranges
are translated directly from original byte spans into `token.Pos`; no line/column
round trip is used. The primary range may cover several source segments. Detailed
segments, evidence, and raw/effective scores remain in the complete
`unswell.RunResult` (returned as the analyzer result for dependent analyzers).

The adapter requires the driver's `Pass.ReadFile` callback, supporting editor
overlays without opening source files itself. It reads only files represented in
`Pass.Files` and uses their FileSet names (even when `//line` changes displayed
names). It checks FileSet sizes and all reported ranges. The driver remains
responsible for providing bytes corresponding to its AST; equal byte length alone
does not prove content identity. Pass callbacks are called sequentially.

On packages with parse or type errors, the analyzer does not opt into running.
The following return Go errors: read failures, invalid input metadata,
resource-limit violations, malformed permissions, and incomplete analysis. They
cannot become a passing policy result through display mode or `NoGate`.
`Options.Context` supplies caller cancellation; `analysis.Pass` has no context
field. A blocked ReadFile callback must be canceled by its driver.

## Evidence and scope

[analysistest fixtures](../goanalysis/testdata/src) contain expected detections for
these cases: comments, interpreted and raw strings, both contexts, suppressions,
and file and language overrides. The tests compare the complete adapter result with
direct engine output. Separate position tests cover Unicode, BOM, CRLF, escaped
strings, `//line`, and related occurrences; these supplement analysistest's line
matching with exact byte-range assertions.

The [driver test](../goanalysis/driver_test.go) builds the executable and runs real
`go vet -vettool` checks: comment and string violations fail; the edited source
passes; source bytes stay unchanged. An external consumer compiles the public API;
module inventory drives native-platform, minimum-compiler, lint, and vulnerability
checks. Library-boundary checks include a negative test for importing Unswell
internals. Alpha API compatibility with previous releases is not required, and race
detection and active fuzzing are deferred until the final validation task,
[#123](https://github.com/stokaro/unswell/issues/123).

The adapter does not register itself with stock golangci-lint or gopls; a future
golangci-lint Module Plugin System wrapper can use this API while owning its own
driver and version compatibility. Git revision discovery and trusted-base policy
selection remain separate application responsibilities.
