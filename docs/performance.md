# Scan cost and resource limits

The September 7, 2026 specification states one performance target for a complete
scan: 100,000 prose words within 10 seconds and 512 MiB on a 2-vCPU Linux host.
This page records how that is measured and what a measurement showed. It
describes this build on the named host. It is not a capacity promise, and it
says nothing about editorial quality.

## Protocol

`scripts/measure-performance.sh` builds a fixed corpus, scans it twice, and
prints one JSON observation. The corpus spreads prose across every scanned
format: Markdown paragraphs and headings, Go comments, YAML values, and Bash
comments. Its sentences repeat by design. A fixed workload keeps two
measurements comparable. Repetition also keeps the repetition rules busy, so the
numbers are an upper bound for that rule family, not an average document.

The script grows the corpus until the tool itself counts at least the requested
number of prose words. A generator estimate would not do: only the tool decides
what counts as prose. Both scans write all five report formats and evaluate the
gate, because a real run pays for both. The first scan is labeled cold and the
second warm. Neither one controls the file cache of the operating system, so the
pair shows repeat behavior rather than a true process cold start.

Peak memory is the maximum resident set of the whole process, read from
`/usr/bin/time`. It covers the runtime, extraction, NLP, rules, and report
writing together. It is not the memory of any single stage.

```sh
make build
bash scripts/measure-performance.sh --words 100000 --label my-host --output observation.json
```

`--self-test` runs the same harness on a small corpus and checks it end to end,
including that an exceeded `analysis.max_file_bytes` limit exits 2 and does not
report a complete analysis. `make check` runs that self-test; the full
measurement is deliberately separate, because it is slow and host-specific.

## Measured on a 2-vCPU Linux host

[The recorded observation](performance/linux-amd64-2vcpu-512mib.json) binds the
runtime commit, compiler, host, limits, and raw counts. It was taken in a
container limited to two CPUs of quota and a hard 512 MiB memory limit, with
`GOMAXPROCS=2`, on Debian 12 with Go 1.25.0 and `CGO_ENABLED=0`.

| Measurement | Value | Target |
| --- | --- | --- |
| Prose words scanned | 102,005 in 57 documents | at least 100,000 |
| Cold scan | 2.702 s | within 10 s |
| Warm scan | 2.595 s | within 10 s |
| Peak resident memory | 197,971,968 bytes | within 512 MiB |

The kernel enforced that memory limit. The scan ran to completion inside a
512 MiB cgroup with no swap, so the number is a bound the host held, not one the
run happened to stay under.

The host is shared and its CPU was not isolated, so these numbers carry ordinary
scheduling noise. A container CPU quota is not the same thing as a two-processor
machine: the container still reports eight logical CPUs, which is why
`GOMAXPROCS` is set explicitly rather than left to the runtime.

## Bounded failure

Exceeding a configured analysis limit is an operational failure, not a quiet
truncation. A scan whose input passes `analysis.max_file_bytes` exits 2, reports
an incomplete analysis, and cannot pass a gate. The self-test asserts that
outcome on every `make check`.

## Limits of this evidence

One corpus on one host does not qualify this tool's speed. Real corpora differ
in format mix, in sentence length, and in how many repetition candidates they
carry. A document with many findings also pays more to write its reports. This
run used a build from the recorded commit, not a published release binary.
Measurements on other architectures, on real project corpora, with published
release binaries, and with a configured probability model are tracked in
[#174](https://github.com/stokaro/unswell/issues/174).
