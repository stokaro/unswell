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

`bash scripts/acquire-corpus.sh` clones each snapshot at its tag, reads the
commit and its date, asks the named record for the publication date, and
writes an acquisition record. GitHub release lookups use `GITHUB_TOKEN` when
it is set; without it the anonymous limit of sixty requests per hour turns
later entries into `vcs_only`, which the log shows. A date the record confirms makes the snapshot
`corroborated`; a snapshot without a readable record stays `vcs_only`, and
one dated after the boundary is skipped and logged. The offline
`corpus acquire` command then applies the fixed selection rules and writes
the shard manifest with every exclusion. The script records every outcome
in `acquisition-log.json` under `artifacts/acquisition`, including clone
failures and missing notices, and it never edits a checkout.

`bash scripts/measure-corpus.sh` runs the shard manifests through
`corpus dataset plan`, `pin`, and `verify`, then per shard through `plan`,
`extract`, `verify`, and `measure` under `policy-e1.yaml`, and `corpus analyze`
builds the E1 tables from every finding artifact. A shard that fails a step
twice is skipped and written to `failed-shards.json`. `--only SHARD` repeats
one shard, and `--resume` continues a run that stopped, skipping shards
that already have findings; a full run starts from empty outputs. An acquisition proves that exact bytes existed at a dated
snapshot under a stated license. It proves nothing about who wrote them.
