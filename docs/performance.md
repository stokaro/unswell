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

`--corpus DIR` measures a real project tree instead. The script copies the tree
without its version-control metadata and writes a policy that keeps the default
file selection, so the tree's own format mix is what gets measured. Each
`--exclude` pattern is added to the policy and listed in the observation. A real
tree is measured as it is; nothing sizes it to the target. `--timeout` bounds
the analysis and defaults to ten minutes, so a slow tree is measured rather than
cut off at the tool's default deadline. A scan the kernel kills leaves no
report; the observation then records the exit code with zero counts.

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

## Measured on real project trees

The same 2-vCPU Linux host, under the same limits, scanned three trees from the
research corpus with a build of the recorded commit. The
[rerun of the synthetic corpus](performance/linux-amd64-2vcpu-512mib-rerun.json)
on that build took 2.451 s cold with a 226 MB peak, close to the first record.
Each tree kept the default file selection. The policy excludes directories
whose content is not English prose by design, and each record lists them. For
FastAPI they are the Japanese and Chinese translation trees; for date-fns the
locale data; for pytest one test file of non-Latin test strings. Each record
also names the snapshot commit of its tree.

| Tree | Prose words | Documents | Bytes | Cold | Warm | Peak resident | Outcome |
| --- | --- | --- | --- | --- | --- | --- | --- |
| [pytest](performance/linux-amd64-2vcpu-512mib-pytest.json) | 137,373 | 304 | 6.8 MB | 8.016 s | 7.649 s | 398,409,728 bytes | complete; the gate failed on findings (exit 1) |
| [FastAPI](performance/linux-amd64-2vcpu-512mib-fastapi.json) | 121,028 | 700 | 7.5 MB | 13.213 s | 15.006 s | 401,010,688 bytes | incomplete; 3 operational errors (exit 2) |
| [date-fns](performance/linux-amd64-2vcpu-512mib-date-fns.json) | not reported | not reported | 11.2 MB | killed after 10.618 s | killed after 10.789 s | 547,254,272 bytes at the kill | killed by the memory limit (exit 137) |

The target holds for pytest. It does not hold for FastAPI, which needs 13 s for
121 thousand words. It does not hold for date-fns: the scan crossed the 512 MiB
limit after ten seconds and the kernel killed it, so it reported nothing. On the
arm64 host below, which enforces no limit, the same date-fns scan completed in
6.7 s with a 739 MB peak and counted 106,594 prose words in 1,151 documents.

Both failing trees carry far more documents and bytes than the synthetic corpus:
7.5 MB in 700 documents and 11.2 MB in 1,151 against 0.7 MB in 57. Peak memory
follows the documents and bytes scanned, not the prose words. The synthetic
corpus, with few large documents, is the cheap case. FastAPI's three operational
errors and date-fns's five stay in the records. They come from a YAML file of
non-Latin link titles, one Markdown and three TypeScript files the grammars
could not parse, and one JavaScript file of locale names.
[#181](https://github.com/stokaro/unswell/issues/181) bounded memory and time
on such trees; the section after the arm64 host records the same scans again.

## Measured on an arm64 host

The same script ran on an Apple M3 Pro under macOS 26.5 with `GOMAXPROCS=2` and
no memory limit. The host enforces nothing, so its peak figures show what each
scan took. Resident figures on macOS run higher than on Linux for the same
scan, so the two hosts are not compared with each other.

| Corpus | Prose words | Documents | Cold | Warm | Peak resident | Outcome |
| --- | --- | --- | --- | --- | --- | --- |
| [synthetic](performance/darwin-arm64-2threads.json) | 102,005 | 57 | 1.164 s | 1.123 s | 278,921,216 bytes | pass |
| [pytest](performance/darwin-arm64-2threads-pytest.json) | 137,373 | 304 | 4.303 s | 3.668 s | 473,956,352 bytes | complete; gate failed (exit 1) |
| [FastAPI](performance/darwin-arm64-2threads-fastapi.json) | 121,028 | 700 | 6.257 s | 6.232 s | 502,431,744 bytes | incomplete; 3 errors |
| [date-fns](performance/darwin-arm64-2threads-date-fns.json) | 106,594 | 1,151 | 6.651 s | 5.534 s | 738,721,792 bytes | incomplete; 5 errors |

## Measured after the bounds of #181

Profiles on the 2-vCPU host put the cost of the failing trees in four places,
none of them in analysis. Every parse built a fresh parser, and a Markdown
document parses every inline block on its own. Resolving the policy of a file
copied it through JSON twice per file. Every per-source partial result carried
its own copy of the rule catalog until the merge. Writing a file report
serialized it into memory as a whole first.

The extractor now keeps parsers per grammar. A plan resolves each override
combination once. Only the batch result carries the manifest, and file reports
stream to disk. The runtime also keeps its total memory under a soft limit of
384 MiB unless `GOMEMLIMIT` is set, so the heap no longer grows to twice its
live size. Results, hashes, and reports of an unchanged tree are the same bytes
as before.

The same four scans ran again on the same 2-vCPU host with a build of that
change. Each record names the commit and the tree hash of the build; the tree
hash survives the squash merge that renames the commit.

| Corpus | Prose words | Documents | Cold | Warm | Peak resident | Outcome |
| --- | --- | --- | --- | --- | --- | --- |
| [synthetic](performance/linux-amd64-2vcpu-512mib-synthetic-bounded.json) | 102,005 | 57 | 1.939 s | 1.932 s | 233,537,536 bytes | pass |
| [pytest](performance/linux-amd64-2vcpu-512mib-pytest-bounded.json) | 137,373 | 304 | 6.405 s | 6.343 s | 347,471,872 bytes | complete; gate failed (exit 1) |
| [FastAPI](performance/linux-amd64-2vcpu-512mib-fastapi-bounded.json) | 121,028 | 700 | 7.689 s | 7.756 s | 402,526,208 bytes | incomplete; 3 errors |
| [date-fns](performance/linux-amd64-2vcpu-512mib-date-fns-bounded.json) | 106,594 | 1,151 | 5.235 s | 7.348 s | 408,363,008 bytes | incomplete; 5 errors |

Every measured tree now completes within the target on the recorded host.
FastAPI fell from 13.2 s to 7.7 s. date-fns, killed before, completes in five
to seven seconds with a 408 MB peak. The peak of a document-heavy tree still
sits near 400 MB: the soft limit holds the heap, and the mapped binary and the
runtime's own bookkeeping add the rest.

The arm64 host ran the same four scans with the same build
([synthetic](performance/darwin-arm64-2threads-synthetic-bounded.json),
[pytest](performance/darwin-arm64-2threads-pytest-bounded.json),
[FastAPI](performance/darwin-arm64-2threads-fastapi-bounded.json),
[date-fns](performance/darwin-arm64-2threads-date-fns-bounded.json)): 0.890 s,
2.935 s, 3.415 s, and 2.484 s cold. Its resident figures stay above the Linux
ones for the same scans and are not compared with them.

## Measured with the published release binaries

Every record above used a build from the recorded commit. The
`v0.1.0-alpha.1` release assets are a separate build, so the
`linux_amd64` and `darwin_arm64` archives were downloaded again, verified
against the release's `SHA256SUMS`, and measured with
`--binary`. Each record carries the digest of the binary it measured. Both
binaries embed Go 1.27.1 and the release source `247fe6f6`. That build
predates the parser-timeout scaling of
[#171](https://github.com/stokaro/unswell/issues/171) and the bounds of
[#181](https://github.com/stokaro/unswell/issues/181), and it has no
`--project-root` flag. The script runs such a binary from the corpus
directory, so the policy's patterns resolve against the tree, and the record
names the flag it left out.

On the 2-vCPU Linux host:

| Corpus | Prose words | Documents | Cold | Warm | Peak resident | Outcome |
| --- | --- | --- | --- | --- | --- | --- |
| [synthetic](performance/linux-amd64-2vcpu-512mib-synthetic-release.json) | 102,005 | 57 | 1.826 s | 1.821 s | 198,979,584 bytes | pass |
| [pytest](performance/linux-amd64-2vcpu-512mib-pytest-release.json) | 137,373 | 304 | 6.086 s | 6.232 s | 384,761,856 bytes | complete; gate failed (exit 1) |
| [FastAPI](performance/linux-amd64-2vcpu-512mib-fastapi-release.json) | 121,028 | 700 | 10.391 s | 10.184 s | 403,501,056 bytes | incomplete; 3 errors |
| [date-fns](performance/linux-amd64-2vcpu-512mib-date-fns-release.json) | 98,620 | 1,150 | 6.389 s | 6.225 s | 527,056,896 bytes | incomplete; 6 errors |

The published build behaves like the build of its time. It meets the target
on the synthetic corpus and on pytest. FastAPI runs a few tenths of a second
over the ten seconds, as the pre-#181 record of that tree did. date-fns
completes, with a peak nine megabytes under the 512 MiB limit; the earlier
kill of that tree came from a scan without the locale exclusion. A release
that carries #171 and #181 will measure as the bounded records above do.
Until one exists, the published binary is the one users run, and these are
its figures.

On the arm64 host, with the same binaries of that release:

| Corpus | Prose words | Documents | Cold | Warm | Peak resident | Outcome |
| --- | --- | --- | --- | --- | --- | --- |
| [synthetic](performance/darwin-arm64-2threads-synthetic-release.json) | 102,005 | 57 | 0.871 s | 0.864 s | 243,597,312 bytes | pass |
| [pytest](performance/darwin-arm64-2threads-pytest-release.json) | 137,373 | 304 | 2.912 s | 2.903 s | 411,533,312 bytes | complete; gate failed (exit 1) |
| [FastAPI](performance/darwin-arm64-2threads-fastapi-release.json) | 121,028 | 700 | 4.813 s | 4.688 s | 458,145,792 bytes | incomplete; 3 errors |
| [date-fns](performance/darwin-arm64-2threads-date-fns-release.json) | 98,620 | 1,150 | 3.013 s | 2.897 s | 648,052,736 bytes | incomplete; 6 errors |

## Measured with a configured origin model

No accepted pack exists. None can exist until the human-labeled corpus of
[#22](https://github.com/stokaro/unswell/issues/22) qualifies one. The cost
of a configured model does not wait for that. An experimental pack of the
same shape loads through the same path. `corpus pack --task origin_endpoint`
built one from the fitted artifact of the
[documentation run](../research/origin/runs/2026-09-11-documentation/README.md)
with a twelve-word floor. The script passes it with `--origin-model` under
a policy that enables the origin channel and accepts an experimental pack.
Each record carries the pack's digest. The estimate is unqualified and
never gates. Only its cost is measured here.

| Corpus | Host | Cold | Warm | Peak resident | Without the model |
| --- | --- | --- | --- | --- | --- |
| [synthetic](performance/linux-amd64-2vcpu-512mib-synthetic-origin.json) | Linux | 1.977 s | 2.022 s | 228,999,168 bytes | 1.939 s, 233,537,536 bytes |
| [pytest](performance/linux-amd64-2vcpu-512mib-pytest-origin.json) | Linux | 6.396 s | 6.258 s | 340,975,616 bytes | 6.405 s, 347,471,872 bytes |
| [FastAPI](performance/linux-amd64-2vcpu-512mib-fastapi-origin.json) | Linux | 7.756 s | 7.557 s | 399,089,664 bytes | 7.689 s, 402,526,208 bytes |
| [date-fns](performance/linux-amd64-2vcpu-512mib-date-fns-origin.json) | Linux | 5.400 s | 6.160 s | 410,337,280 bytes | 5.235 s, 408,363,008 bytes |
| [synthetic](performance/darwin-arm64-2threads-synthetic-origin.json) | arm64 | 0.873 s | 0.878 s | 234,143,744 bytes | 0.890 s |
| [pytest](performance/darwin-arm64-2threads-pytest-origin.json) | arm64 | 2.852 s | 2.818 s | 472,924,160 bytes | 2.935 s |
| [FastAPI](performance/darwin-arm64-2threads-fastapi-origin.json) | arm64 | 3.314 s | 3.203 s | 468,418,560 bytes | 3.415 s |
| [date-fns](performance/darwin-arm64-2threads-date-fns-origin.json) | arm64 | 2.398 s | 2.305 s | 672,055,296 bytes | 2.484 s |

The model costs nothing the measurement can separate from noise. Every scan
stays within a few tenths of a second and a few megabytes of the same scan
without it. Every Linux scan stays within the target. The pack is fourteen
logistic weights and an isotonic map. Its evaluation per paragraph is a few
multiplications beside the extraction and the rules the scan pays for
anyway.

## Limits of this evidence

Three real trees on two hosts do not qualify this tool's speed. Real trees
differ in format mix, in sentence length, and in how many repetition candidates
they carry, and the three measured ones already span the target on one side and
the other. The exclusions are a policy choice a maintainer would make for an
English scan; each record lists them. The published binaries are the first
release, which predates the bounds the current build carries. The origin
model is an experimental pack, not an accepted one. A later release and an
accepted pack can be measured the same way.
