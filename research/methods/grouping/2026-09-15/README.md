# Global grouping repair for paired tables

The sampler copied `Candidate.GroupID` from an individual extraction shard into
each task. The paired analysis then resampled that archived ID, while its H0
documents used finding-shard groups. Splitting one repository among shards could
therefore increase the apparent number of independent components and separate an
original from related baseline material during bootstrap draws.

The corrected sampler resolves groups through the global dataset plan. Its
`--min-groups` option checks the selected draw before request generation, without
changing the draw or seed. The paired command now requires `--plan`; it resolves
every original, response, and H0 source through that same plan. Missing bindings,
duplicate source identities, and derivatives assigned to another global component
are errors. Output version 2 records the dataset identity.

## Recalculation

All 20 committed generation runs were replayed against the same saved findings
and responses under both implementations. No generator, extraction, rule scan,
selection, or editorial annotation ran. The baseline is commit
`f3d09fd3aeaeb355d6519b9cacb98598f149ef6d`; the corrected implementation is the
change accompanying this record. Both binaries used the same product engine.

The [table](tables.md) lists measured pairs and independent groups per run.
For `2026-09-13-confirm-haiku`, 184 pairs span 16 global groups, while version 1
counted 56. For `2026-09-13-long-haiku`, 159 pairs span 11 global groups; version 1
counted 20. Task-file counts can be larger still because some originals have no
measurement. Coverage gaps remain in each arm's record.

Pair counts, finding counts, and point estimates are identical before and after
for every run and rule. The correction changes resampling groups, intervals, and
zero-count upper bounds. Earlier version 1 paired intervals are superseded for
statistical use. Their saved artifacts remain unchanged for audit.

The screening command already uses global dataset-plan groups. Its frozen
confirmation artifact still has SHA-256
`dc56db3b3ae5d63c1a6d1bea67c885d3790f1272224af2b711cf535b0f70e0ce`.
This repair neither changes that result nor creates another permitted
confirmatory look. The minimum of 20 components remains unmet in those arms.
Issues #218 and #154 retain their outstanding experimental requirements.

## Artifacts and reproduction

`paired-replay.json.gz` contains compressed JSON with all before/after arm counts,
rule estimates, intervals, and report hashes for all 20 runs. It also records
task/record hashes, both binary hashes, and the global plan hash. The [table](tables.md)
records compressed and uncompressed SHA-256. It contains no source prose or keys.

Inspect a run with:

```sh
gzip -dc research/methods/grouping/2026-09-15/paired-replay.json.gz | \
  jq '.runs[] | select(.run == "2026-09-13-long-haiku")'
```

Build the research CLI from the baseline commit and this change with cgo disabled.
The existing acquisition and measurement pipeline supplies the pinned plan and
findings. Pass a new, nonexistent output directory; the replay refuses to replace
earlier evidence:

```sh
python3 scripts/replay-paired-groups.py \
  --before /path/to/baseline-corpus --after /path/to/corrected-corpus \
  --runs research/generation/runs --findings artifacts/measurement/findings \
  --classes research/methods/rule-classes-v1.json \
  --plan artifacts/measurement/dataset-plan.json --output /path/to/new-replay
```

The script selects the complete historical cohort and each run's own controlled
shards. It preserves full command outputs beside the summary. A partial replay
has `status: in_progress`; only all completed runs produce `status: complete`.
Recomputing tables on already observed outcomes corrects arithmetic and provenance;
it is not independent confirmation or a quality-labeling exercise.
