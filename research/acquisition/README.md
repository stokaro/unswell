# Historical acquisition

This directory holds the sampling frame of the historical pilot of the
[pattern protocol](../methods/llm-patterns-v1.md) and the log of what the
acquisition script did with it. It contains no source text; the script
produces shard manifests that point at exact bytes by repository, commit,
path, and digest.

[`sources-v1.json`](sources-v1.json) lists public repositories with a
permissive license, English documentation, and a release that an independent
record should date on or before the boundary, December 31, 2020. The list is
a design input chosen by license, ecosystem diversity, and history
availability; nothing in it depends on the text. Each entry names the tag to
check out and the record that dates the release: a PyPI, npm, crates.io, or
Maven Central version, or a GitHub release. Tags and dates in the list are
claims that the script verifies at run time.

[`sources-contemporary-v1.json`](sources-contemporary-v1.json) names the same
repositories at their latest published GitHub release, so each contemporary
snapshot shares a provenance component with its historical one and the E4
temporal check compares like with like. A contemporary snapshot dated before
November 30, 2022 is skipped, and its origin stays unknown.

`bash scripts/acquire-corpus.sh --cohort historical|contemporary` clones
each snapshot at its tag, reads the commit and its date, asks the named
record for the publication date, and writes an acquisition record. A tag
named `latest` resolves to the latest published GitHub release, and the log
records the resolved tag. Outputs go to `artifacts/acquisition/<cohort>` and
checkouts to a per-cohort work directory outside the repository. GitHub
release lookups use `GITHUB_TOKEN` when it is set; without it the anonymous
limit of sixty requests per hour turns later entries into `vcs_only`, which
the log shows. A date the record confirms makes the snapshot
`corroborated`; a snapshot without a readable record stays `vcs_only`, and
one dated after the boundary is skipped and logged. The offline
`corpus acquire` command then applies the fixed selection rules and writes
the shard manifest with every exclusion. A notice kept as a link into the
checkout is read; any other link is left out. The script records every outcome
in `acquisition-log.json` under `artifacts/acquisition`, including clone
failures and missing notices, and it never edits a checkout.

The period frames `sources-historical-2012-v1.json`,
`sources-historical-2016-v1.json`, and `sources-historical-2018-v1.json`
name the same repositories with the tag `before`. The script then walks the
repository's tags by commit date and takes the newest release tag committed
on or before the frame's boundary. A release tag is the entry's tag prefix,
an optional `v`, and a dotted version; a PyPI-style beta such as `18.9b0`
counts only when no other release tag precedes the boundary. The version
derived from the tag corroborates it against the entry's registry or
release record as usual. The cohorts `historical-2012` and
`historical-2016` are the protocol's placebo pseudo-boundaries inside H0;
`historical-2018` is the H1 sensitivity boundary. The walk needs
`GITHUB_TOKEN`.

`bash scripts/measure-corpus.sh` runs every cohort's shard manifests under
`artifacts/acquisition` through one `corpus dataset plan`, `pin`, and
`verify`, then per shard through `plan`, `extract`, `verify`, and `measure`
under `policy-e1.yaml`, and `corpus analyze` builds the E1 tables, with a
contrast for every cohort beside the historical baseline, from every finding
artifact. A shard that fails a step
twice is skipped and written to `failed-shards.json`. `--only SHARD` repeats
one shard, and `--resume` continues a run that stopped, skipping shards
that already have findings; a full run starts from empty outputs. An acquisition proves that exact bytes existed at a dated
snapshot under a stated license. It proves nothing about who wrote them.

After the tables, the driver counts paragraphs and sentences as units. It
orders the cohorts by period: 2012, 2016, 2018, the H0 baseline, the
contemporary snapshots, then natural and controlled text. For each
consecutive pair it writes a `corpus first-appearance` filter. The unit
tables of that pair keep only the later cohort's units that the earlier
snapshot of the same repository does not hold. Their contrast is against
the earlier cohort. Last, `corpus dedupe` keeps each unit text once across
every cohort in that order. The driver builds the unit tables restricted to
that selection.
