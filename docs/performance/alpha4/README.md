# Alpha.4 scan measurements

The published [alpha.4 release](https://github.com/stokaro/unswell/releases/tag/v0.1.0-alpha.4) at
`7e5b93398903ce2f108c84ebf8168dd48d6696df` completed all 24 scans on September 13, 2026.
These records address [#219](https://github.com/stokaro/unswell/issues/219).
The archive and executable hashes are in [environment.json](environment.json);
all six release binaries also passed the [source reproduction audit](../../release/v0.1.0-alpha.4-audit.json).

| Corpus | Documents | Prose words | Six-sample time range | Highest RSS | Complete result |
| --- | ---: | ---: | ---: | ---: | --- |
| synthetic | 57 | 102,005 | 1.060–1.121 s | 251.84 MiB | pass, exit 0 |
| pytest | 304 | 136,739 | 3.634–4.375 s | 386.37 MiB | policy failure, exit 1 |
| fastapi | 702 | 122,659 | 4.444–5.684 s | 385.86 MiB | policy failure, exit 1 |
| date-fns | 1,155 | 74,765 | 2.966–4.939 s | 395.20 MiB | pass, exit 0 |

All processes stayed within 10 seconds and 512 MiB, without OOM kills or rule
abstentions. Synthetic, pytest, and FastAPI each exceed 100,000 prose words.
The date-fns workload is smaller and does not establish that word-count target.
Policy failures are complete analyses whose findings fail the configured gate.

## Reproduce

The wrapper needs Bash, Docker, curl, jq, tar, and a SHA-256 utility:

```sh
bash docs/performance/alpha4/reproduce.sh diabolocom /tmp/unswell-alpha4-measurement
```

The output directory must not exist. The wrapper verifies the published archive
and executable, downloads the original pinned [inputs](inputs.json), and runs
[run.sh](run.sh) in the [container](Dockerfile). Limits are two CPUs of quota,
512 MiB, no swap, no network, and `GOMAXPROCS=2`. Each corpus receives three
pairs of scans; each process writes all five report formats. The wrapper removes
its container and tagged image after collecting the results.

This measurement prepared the payload and executed the inner script with Docker
commands. The full download wrapper was not run. The measurement exercised the
same binary, input archives, inner script, container recipe, and limits. Its
container and image were removed; existing host services were left running.

## Coverage and limits

The source archives, file exclusions, policy bytes, and selected input byte totals
match alpha.3. Every previously processed document keeps its source hash, format,
and byte count. [The coverage comparison](coverage-vs-alpha3.json) records each
added document and changed prose count.

FastAPI now includes `docs/en/docs/advanced/security/oauth2-scopes.md`.
The three repaired date-fns files are `pkgs/core/src/fp/index.ts`,
`pkgs/core/src/index.ts`, and `pkgs/core/src/types.ts`. No failed file was removed.
The parser failures from [#230](https://github.com/stokaro/unswell/issues/230)
and [#231](https://github.com/stokaro/unswell/issues/231) are absent in every sample.

The extractor now skips Actions shell syntax, expressions, bodies in unsupported
shells, and some TypeScript import/export literals. This changes how much
prose it checks, even in files whose source bytes match.

Pytest has 128 fewer prose words across five workflow files. FastAPI and date-fns
also have fewer words in their workflow files, plus the new files listed above.
Each sample lists the reasons for excluding text.

Other services shared the host, with load averages from 0.17 to 1.24 before the
pairs. The run did not reserve CPUs or control filesystem caches. Each sample
started a new process, after source copying or synthetic calibration. The names
`cold` and `warm` in the raw harness paths do not establish a cold cache.
The base image is pinned, but Debian package repositories are not frozen; the
environment record lists the actual installed packages.

## Evidence

Each of the twelve `CORPUS-N.json` files records both scans in a pair. It lists
the outcomes, peak RSS for each process, exclusions, and abstentions. It also
records byte counts and hashes for JSON, SARIF, HTML, Markdown, and text.

All 120 reports exist. Report hashes and source identities match across the six
samples for each corpus. Running the committed summarizer on the saved reports
reproduced all twelve records byte for byte.

The raw reports are retained in the measurement artifacts rather than duplicated
here. Reproduction retains them under the chosen output directory. Source
snippets were not requested. These measurements establish scan cost and
completeness on the named inputs; they make no claim about editorial precision.
