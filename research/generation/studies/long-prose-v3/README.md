# Complete technical documents: version 3

This study follows [protocol version 3](../../../methods/llm-patterns-v3.md)
for issues #218 and #154. Its independent source frame contains 38 documents
from 20 repositories, with 58,137 extracted prose words. Three model families
receive the same documents under generation/editing and neutral/plain conditions:
456 requests. Historical authorship remains unknown; dated snapshots are not
independent human quality labels.

`inputs.tar.gz` preserves the complete freeze made at 11:28:24 UTC on September 15,
2026, before delivery or rule measurement. `inputs.json` records its checksum and
the checksum of `freeze.json`. The archive contains source files and their notices,
manifests, global groups, candidate units, curated factual briefs, tasks, exact
requests, the protocol, policy, runner, and the research-code patch used to build
the frozen corpus binary. The engine is commit
`26b30e8fb2c5ba46cc0255a6d538000ae9a9d0d4`.
This is a locally recorded advance commitment, not external preregistration.
The final four requests used the disclosed dispatch-reservation amendment in
the [protocol](../../../methods/llm-patterns-v3.md#execution-amendment-dispatch-reservations).
The original freeze remains unchanged; `outputs.tar.gz` also contains the
completion record and wrapper. The results distinguish the changed reservation
ceiling from actual reported token usage.

`selection-audit.tar.gz` retains supplemental acquisition logs, rights records,
the inventory manifests, and all 118 eligible-document records before the date
check. Two of the 31 repositories fall outside the date boundary, three lack
independent date corroboration, and six have no document meeting the format,
length, and role constraints. The remaining 20 supply the 38 selected documents.
Replaying the fixed rank and preferred-length selection reproduces all 38 source
IDs. These supplemental files were copied after generation started; they were not
members of the main freeze. The final admitted sources, selection, and two source
frames are in that freeze.

The generation runner is an isolated research utility. Ordinary builds and CI
do not run it or require model access. It uses the authorized Codex and Claude
subscriptions and the configured Qwen router. It never executes generated text.
The archive records visible inputs; subscription CLIs may supply additional
built-in context that their output does not expose.

Document briefs are agent-curated, with source-bound summaries and exact technical
excerpts. They preserve identifiers, numbers, conditions, and attribution material;
their seal verifies byte bindings, not independent semantic review. Generation
from facts and editing the full source are therefore separate experimental arms.

## Saved evidence

The [results](results.md) report every primary comparison, historical review load,
length sensitivity, context opportunities, ablations, and all 40 rule decisions.
The [data card](data-card.md) records the sample and interpretation limits.
The [construction review](inspection.md) distinguishes removable prefaces from
necessary technical contrasts, repeated terms, and release obligations.

Five archives retain the inputs and results. Each adjacent JSON file names the
archive's byte size and SHA-256; result archives also list every member's hash.

| Archive | Contents |
| --- | --- |
| `inputs.tar.gz` | The original freeze, sources, briefs, tasks, requests, policy, protocol, runner, and research-code patch |
| `selection-audit.tar.gz` | Supplemental acquisition and selection records |
| `outputs.tar.gz` | Raw delivery records, outcomes, imported responses, generation records, and coverage |
| `measurements.tar.gz` | Source copies, global plan, verified candidates, findings, paired estimates, aggregate tables, decisions, and source-bound cases |
| `features.tar.gz` | Batched public-CLI activation reports, with source mappings and applicability |

The corpus command hashes raw configuration bytes. The CLI hashes its configuration
bundle. Their config identities therefore differ even though both read the same
frozen policy. The analysis checks its file hash, matching rule identities and
catalogs, source hashes, and every document's diagnostic counts. It does not equate
the two config hashes.

## Reproduce without model access

Use Python 3.11 or later and the Go version in `go.mod`. These are research
utilities; ordinary product tests neither unpack the full study nor call a model.
The replay creates a new directory and leaves published archives unchanged.
The commands below rebuild the corpus adapter and public CLI from the study's
merged source commit, `966ffd841bcfa3dbbed2bc706ae742c18c2b569d`.
The replay verifies archive members, reimports saved responses, and remeasures
source text. It then recomputes activation features and reproduces the tables
and report.
Later rule changes, including noun-stack version 5, require this pinned checkout
to reproduce the frozen version 4 measurements.

```sh
study_scratch="$(mktemp -d)"
study_source="$study_scratch/source"
git worktree add --detach "$study_source" 966ffd841bcfa3dbbed2bc706ae742c18c2b569d
study_archives="$study_source/research/generation/studies/long-prose-v3"
study_tools="$study_archives/tools"
(
  cd "$study_source/research/annotation"
  CGO_ENABLED=0 GOMAXPROCS=2 go build -p=2 -o "$study_scratch/corpus" ./cmd/corpus
)
(
  cd "$study_source"
  CGO_ENABLED=0 GOMAXPROCS=2 go build -p=2 -o "$study_scratch/unswell" ./cmd/unswell
)
GOMAXPROCS=2 python3 "$study_tools/reproduce-long-study.py" \
  --archives "$study_archives" --study "$study_scratch/replay" \
  --corpus "$study_scratch/corpus" --binary "$study_scratch/unswell"
git worktree remove "$study_source"
```

Module dependencies must already be available for an offline build. The replay
sends zero generation requests. Omitting `--binary` reuses verified saved activation
reports while still remeasuring all rule findings. Including it also reconstructs
the shared features through the CLI. The final check requires byte-identical
imported records, findings, statistical tables, decisions, cases, and report.
`reproduction.json` records the elapsed time, platform, peak child-process memory,
binary identity, and comparison count. A different binary hash is allowed only
through the replay's explicit independent-rebuild path; matching outputs supply
the compatibility evidence.

The archived `run-models.py` is the delivery record. Reproduction never invokes
it. New generation would consume authorized resources and produce a different
sample; model names and settings alone cannot reproduce the saved text.

## Recorded replay

[The replay record](reproduction.json) reports a complete run on macOS with
`GOMAXPROCS=2`: 493 documents were remeasured, activation features were rebuilt,
and 281 derived files plus the report matched byte for byte. It took 168.73
seconds and the largest child-process peak was 617,889,792 bytes (about 589 MiB).
The repository checks were also running on the host; this is an execution record,
not an isolated performance benchmark. It measures the full research replay,
including JSON analysis, rather than the product's separate 100,000-word target.

The first replay reproduced the same JSON values but exposed dictionary-key
ordering in an aggregate. The reporting utilities now serialize aggregate keys
in sorted order; a second complete replay established the byte comparison above.
The Go measurements, estimates, and scientific decisions did not change.

The final generation total was 1,982,385 reported tokens, below the original
2,000,000-token ceiling. The disclosed amendment changed admission reservations;
it admitted only the last four requests from the original fixed list.
