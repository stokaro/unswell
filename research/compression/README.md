# Compression cost observations

The public [compression benchmark](../../feature/compression_benchmark_test.go)
measures #52's numerical primitive through the shared prepared-unit API. It uses
repeated ASCII fixtures without origin or quality labels. These inputs establish
neither performance on a representative corpus nor predictive value.

The matrix fixes target sizes at 48, 1,024, and 65,536 UTF-8 bytes; reference
prefixes, including the appended LF, at 32 and 32,768 bytes; and zlib levels at
0 and 9. Level 0 is a stored-block control. Every measurement retains validation,
detached source data, result construction, and hashing inside the timed operation.

`reuse-reference` constructs the reference once before timing. Each timed call
still creates its own compressor workspace. `new-reference` constructs and
compresses the reference before every timed measurement. Neither mode measures
whole-process cold start or a complete scan. Extraction, NLP initialization, and
prepared-target construction happen outside each timed loop.

## Reproduce

From the repository root on the declared Go version, build a test executable
without cgo, then run only the benchmark:

```sh
CGO_ENABLED=0 go test -c -o /tmp/unswell-compression-costs.test ./feature
GOMAXPROCS=2 /tmp/unswell-compression-costs.test -test.run '^$' \
  -test.bench '^BenchmarkCompression$' -test.benchtime=100ms \
  -test.benchmem -test.count=5
```

The command skips ordinary tests, coverage, race detection, and active fuzzing.
Use `-test.benchtime=1x -test.count=1` to check every case quickly. Published
timings retain all repetitions; do not select the fastest result as a capacity
claim. To inspect allocations, read `B/op` and `allocs/op`. Cumulative allocations
per operation are not peak resident memory.

The [recorded observation](darwin-arm64-go1.27.1.json) binds the runtime commit,
benchmark source hash, compiler, hardware, command, and raw files. On macOS the
resource wrapper was `/usr/bin/time -l -o resources.txt`, placed immediately
before the compiled executable. Its maximum RSS covers the whole benchmark
process, including the harness, Go runtime, extraction, and NLP preparation.
The compiler runs separately. The hardware was shared; CPU load was not isolated.

These measurements do not qualify a default, establish the 2-vCPU Linux scan
budget, or complete #52. Representative corpora, reference-pack acquisition and
loading, true process cold start, Linux resource measurements, and predictive
comparisons remain open. See [ADR 0032](../../docs/adr/0032-compression-measurements.md).
