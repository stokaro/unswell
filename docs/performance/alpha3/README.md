# Alpha.3 scan measurements

These September 13, 2026 observations address
[#219](https://github.com/stokaro/unswell/issues/219). The [performance page](../../performance.md#published-alpha3-on-diabolocom-september-13-2026)
records the result and its limits. FastAPI and date-fns remain incomplete;
passing the resource budgets does not close the issue.

## Reproduce

The host-side script needs Bash, Docker, curl, jq, tar, and a SHA-256 utility.
Choose a remote Docker context explicitly and an output directory that does
not exist:

```sh
bash docs/performance/alpha3/reproduce.sh diabolocom /tmp/unswell-alpha3-measurement
```

The script verifies the published archive and executable against the hashes
recorded here, downloads the three pinned source archives in
[inputs.json](inputs.json), and runs [run.sh](run.sh) in the
[container](Dockerfile). The container has a two-CPU quota, 512 MiB of memory,
no swap, no network, and `GOMAXPROCS=2`. Each corpus receives three pairs of
scans, all using the existing measurement harness. The script copies the
evidence out and removes its container and tagged image on exit. It leaves
shared images and services alone.

For this record, we prepared inputs and copied results with Docker commands.
The wrapper above repeats those steps. The measurements exercised its inner
script, container recipe, payload, and limits. We have not run the full wrapper.

## Evidence

[environment.json](environment.json) records the release identity, CPU,
container image and package versions, cgroup limits, source archive hashes,
measurement script hashes, and host load readings. The base image is pinned;
Debian package repositories are not frozen, so a later image build may have
different package versions. Compare the recorded environment when reproducing.

The twelve `CORPUS-N.json` files each hold a first and a repeat sample. Both
start fresh processes; neither establishes a cold filesystem cache. Corpus
copying and synthetic size calibration happen before measurement. Shared-host
load averages were between 1.20 and 2.01 immediately before the pairs; CPUs
were not reserved exclusively.

Each sample records peak RSS, elapsed time, and the exit code. It also records
the analysis status, errors, excluded spans, and rules that abstained. RSS
covers the whole process, including report output. Each report has a byte
count and a SHA-256 hash.

Byte totals sum the excluded spans for each reason. They do not count prose
words. The document hash uses a JSON list with these fields from each document:
`name`, `format`, `source_hash`, `bytes`, `prose_words`, and `excluded`. It keeps
report order, sorts keys, and omits JSON whitespace before hashing the bytes.
Missing or malformed JSON leaves the analysis absent. It cannot supply counts
or prove that a scan passed.

The raw reports occupy about 1 GiB and are retained by `--artifacts` in the
reproduction output, rather than duplicated in this directory. No source
snippets were requested. All six JSON reports for a given corpus have the
same hash; counts, document coverage, and outcomes are stable across samples.

## Parser follow-ups

The three date-fns failures share the TypeScript `export type *` form,
tracked in [#230](https://github.com/stokaro/unswell/issues/230). Ordinary
wildcard exports and named type exports pass. Removing only the two wildcard
type exports from a disposable copy of `types.ts` makes that file complete.
The original benchmark inputs were unchanged.

FastAPI's `docs/en/docs/advanced/security/oauth2-scopes.md` fails on a nested
list. Four levels pass and five through eight fail with the same simple
sentence at each level, tracked in
[#231](https://github.com/stokaro/unswell/issues/231).

These failures reproduce on alpha.3 and on the pinned main source build
`1f37aaa3aa40e6a47e2af777856ac7bd9bcff536` (binary SHA-256
`b34c925d73098efa8fe31b0d8a43cbe79603fe8ad93754f1fa8ab4b37c2fc74e`).
That comparison is parser triage, not a measurement of another release.
The original full files were checked on both binaries; compact controls and
the modified `types.ts` copy were also checked on main. Both builds use
gotreesitter v0.52.0.
