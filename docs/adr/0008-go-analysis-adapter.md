# ADR 0008: Go analysis adapter

Status: accepted.

The public `goanalysis` module provides an analyzer over the public Unswell engine.
It owns no rules, extraction, scoring, or gate policy. The separate module keeps
`golang.org/x/tools` out of the core dependency graph. It participates in the same
module tests, strict lint, API snapshot checks, race tests, and vulnerability checks.

`New` constructs an `analysis.Analyzer`. Callers can supply an immutable engine,
an explicit logical project root, and cancellation context. The default engine is
created once with empty prose allowed, because a valid Go package need not contain
human-readable comments or strings. Supplied engine options retain their semantics.
An optional all-findings display mode is distinct from the default gate mode.

The adapter reads only the files provided in `analysis.Pass.Files`, through the
driver's `ReadFile`. It never opens source files, discovers config, or reads the
working directory itself. AST/FileSet metadata locates each complete Go source;
the engine's Go extractor handles configured comments, strings, and exceptions.
Original byte spans become token positions without converting through line/column
numbers, including for Unicode, escaped strings, CRLF, and `//line` directives.
Related locations remain related diagnostics. Unsupported safe fixes are not
invented. Invalid positions, read failures, or incomplete analysis return errors.

Default diagnostics come from the completed engine gate. Suppressions, baseline
acceptance, thresholds, and `NoGate` remain engine decisions. All-findings mode may
cause a Go driver to fail on advisory findings; its name and documentation must
state that distinction. The complete `RunResult` is also returned to dependent
analyzers, so a diagnostic-only driver does not erase raw evidence or scores.

Calls to Pass callbacks are sequential. The immutable engine can be reused across
packages and concurrent passes without a global mutable singleton. An optional
caller context supports cancellation; the Go analysis API has no context field.

A compile-tested singlechecker driver and `analysistest` fixtures demonstrate the
public boundary. Tests compare adapter results with direct engine results and
verify exact offsets beyond analysistest's line-based expectation matching. A
standard `go vet -vettool` invocation supplies executable integration evidence.
Configuration loading and a golangci-lint Module Plugin System wrapper can be
provided by a caller; this prototype does not claim to register with stock
golangci-lint or gopls.
