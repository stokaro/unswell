# Unswell Go analysis adapter

This module exposes `goanalysis.New` over the public Unswell engine. It supports
Go comments and strings, preserves original source ranges, and returns the full
engine result alongside Go diagnostics.

See [configuration, driver usage, and limits](../docs/go-analysis.md).
Run `go test ./...` here for analysistest fixtures, byte-range checks, and the
real `go vet -vettool` workflow. The repository module inventory also runs these
tests in CI.
