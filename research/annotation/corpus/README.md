# Corpus preparation

This package prepares unlabeled candidates for [#22](https://github.com/stokaro/unswell/issues/22).
It reuses Unswell's extraction and English NLP packages, then produces units for
the existing [annotation protocol](../README.md). It supplies no human judgments,
trained model, or qualified evaluation set. See [ADR 0014](../../../docs/adr/0014-corpus-acquisition.md).

## Reproduce the Ptah candidates

Run these commands from `research/annotation` with the supported Go compiler:

```sh
mkdir -p ../../artifacts/corpus
CGO_ENABLED=0 go build -o ../../artifacts/corpus/corpus ./cmd/corpus
../../artifacts/corpus/corpus plan < corpus/testdata/ptah-manifest.json > ../../artifacts/corpus/plan.json
../../artifacts/corpus/corpus extract --root corpus/testdata/ptah < ../../artifacts/corpus/plan.json > ../../artifacts/corpus/candidates.json
../../artifacts/corpus/corpus verify --root corpus/testdata/ptah < ../../artifacts/corpus/candidates.json > ../../artifacts/corpus/verification.json
```

On Windows, build `corpus.exe` and use that filename. Commands read JSON from stdin,
write JSON to stdout, and write errors to stderr. Exit codes are 0 on success,
2 on input/output failure, and 130 on cancellation. Publish redirected output only
after a successful exit. Failed extraction returns no partial artifact.
The explicit source root may contain unrelated files; only declared regular files
are read. Final symlinks are rejected, and `os.Root` confines path resolution.

The library receives source and notice bytes from its caller. Neither it nor the
command fetches references, executes imported code, or invokes another analyzer.
Module acquisition and compiler setup happen before running the built command.

## Freeze the inputs

[manifest.schema.json](manifest.schema.json) defines the strict JSON input.
The [Ptah manifest](testdata/ptah-manifest.json) is a complete executable example.
Duplicate keys, unknown fields, invalid Unicode, and inconsistent declarations
are errors. The schema and semantic checks run together.

| Field | Meaning |
| --- | --- |
| `version`, `id`, `seed` | Contract, acquisition batch, and recorded assignment seed |
| `weights` | Positive training/development/calibration/final-test shares totaling 10,000 |
| `extraction_policy` | Existing `extract.Policy`: global contexts, per-language overrides, and exceptions |
| `unit_kinds` | Nonempty set of `sentence`, `paragraph`, and `fragment` |
| `sources` | Exact original files with context, group metadata, rights, and retained notices |
| `sources[].snapshot` | Optional dating of the exact bytes: `date`, `confidence`, `evidence`, and research `cohort` |

Each source records a relative path, SHA-256, byte count, pinned reference, source
format, declared prose language, topic, purpose, repository, and document ID.
Optional author/template/related-version/generation-task lists hold globally scoped
IDs; empty lists mean unknown. Author-language metadata needs an independent basis.
The current contract accepts declared English prose; it does not infer fluency or
a writer's language background.

A `snapshot` dates the bytes for the temporal cohorts of the
[pattern protocol](../../methods/llm-patterns-v1.md). Its `date` uses the
`YYYY-MM-DD` form. `confidence` is `corroborated` when an independently
published release or archive agrees with the version-control date, `vcs_only`
when only version-control metadata exists, and `unknown` otherwise. Only an
unknown confidence permits an empty date. `cohort` is `historical`,
`controlled`, `natural`, or `contemporary`. A historical source
needs a dated snapshot, and a controlled source needs generated or edited
origin with its generation record. Candidates repeat the cohort of their
source; a source without a snapshot yields candidates with an empty cohort and
enters no temporal analysis. A snapshot dates bytes and says nothing about
authorship or quality.

Permission evidence and at least one retained notice are required. Allowed uses
are explicit; annotation permission alone does not authorize training or publication.
The tool verifies notice bytes and declarations, not the truth of a license or
provenance assertion. Review imported resources before recording those assertions.

## Assign groups before extraction

The planner connects sources sharing a repository, document, author, template,
related-version ID, generation task, or exact source hash. Connections are
transitive. A connected component belongs to one partition. Conflicting explicit
partition assignments fail; unknown relationships cannot establish independence.

`connected-sources-sha256-v1` hashes sorted typed group keys to identify a component.
For unpinned groups, it hashes the algorithm name, seed, and component ID separated
by NUL bytes. The first eight SHA-256 bytes, interpreted as an unsigned big-endian
integer modulo 10,000, select a bucket in the ordered partition weights.
Explicit source assignments pin the whole component. Reordering sources, grouping
keys, unit kinds, and context sets does not change the plan. Extraction exception
and selector order is retained. Changing declarations requires a new plan.

The planner cannot promise balanced unit counts, discover unrecorded paraphrases,
or replace a curator's review for leaks. Freeze and archive the complete source
manifest before extracting units. Do not repeatedly change the seed or inputs to
select a favorable final test. Publish independent document/repository/group counts
alongside fragment counts; nested sentences and paragraphs are related observations.

## Find near duplicates before grouping

The planner connects only the keys a manifest declares and cannot discover an
unrecorded copy. The `duplicates` command supplies that evidence for curation:

```sh
corpus duplicates --root ./sources < manifest.json > duplicates.json
corpus duplicates --root ./sources --apply < manifest.json > manifest-related.json
```

Detector `unswell-near-duplicates-v1` lowercases each source, keeps letters
and digits, hashes every run of eight words, and signs the set with 128
MinHash values. Sources that agree on one of 32 bands become candidate pairs,
and each candidate pair gets its exact Jaccard similarity from the full sets.
Pairs at or above the threshold, 0.5 unless the command says otherwise, form
clusters, and each cluster receives the key `near-duplicate-v1:` followed by
a digest of its member IDs. Sources with fewer than eight words are listed as
too short. The report records the manifest digest, the parameters, every
compared pair, and the clusters.

`--apply` writes the manifest with each cluster key added to the
related-version keys of its members, so the planner puts the cluster in one
component. The threshold and the detector version are part of the frozen
dataset record. The detector sees copied prose; it cannot see a paraphrase or
a shared template that changes most words, and curator review stays required.

## Plan a sharded dataset

One manifest holds at most 10,000 sources. A larger corpus is a dataset of
shard manifests that repeat one seed, weight set, extraction policy, and unit
kind set, with a pinned rule-class hash:

```sh
corpus dataset plan --root ./dataset < dataset.json > dataset-plan.json
corpus dataset pin --root ./dataset --output ./pinned < dataset-plan.json
corpus dataset verify --root ./dataset --pinned ./pinned < dataset-plan.json
```

`plan` reads every shard beneath the root by its declared path and digest. It
rejects a shard whose header differs from the dataset. It connects sources
across all shards with the keys a single manifest uses and assigns whole
global components with the same algorithm. A source ID or path appears in one
shard only, and a notice shared by shards must agree everywhere. The plan
lists each global group with the shards it spans, and each source with its
group and partition. A group may span shards and never spans partitions.

`pin` writes a copy of each shard with every partition set to its global
value. The copy uses a fixed encoding whose digest the plan records, and an
existing file is never overwritten. The ordinary `plan`, `extract`, and
`verify` commands then run on each pinned shard, and their local components
inherit the global partition through the pins. `verify` recomputes the
dataset plan from the original shards. With `--pinned` it also checks every
pinned copy against its recorded digest. Limits are 64 shards and 200,000
sources; each shard keeps the manifest limits above.

## Units and source ranges

Extraction uses the frozen policy, source bytes, and existing English sentence and
token provider. Protected boundaries split eligible prose; they are not removed to
join independent text. Context is the enclosing contiguous eligible piece.
Paragraphs and comments without protected boundaries may supply paragraph and
sentence units. Strings, headings, and broken pieces supply fragments. Requested
kinds without eligible units remain absent, with an empty-source reason recorded.

Each candidate retains original UTF-8 source segments, context, kind, role, word
count, component, partition, source identity, and extraction-policy hash. Escaped
strings and Markdown transformations retain their original ranges. The output also
keeps extraction exclusions. Reviewed `role_regions` can distinguish doc comments,
error messages, logs, UI text, and other strings. A region cutting through a unit
is an error. Otherwise, code comments and strings retain generic roles. A
source whose role cannot be determined declares `unknown`, and its prose units
keep that value.

Whole-source provenance remains in the manifest. Candidate origin is always
`unknown` until independent unit evidence is supplied. A repository-level AI-use
statement cannot become an automatic origin or quality label.

Artifacts contain `annotation.Unit` records but no actors, responses, or
adjudications. A curator selects units, creates a round with real participant
assignments and profile instructions, and uses `annotate packet` for independent
review. Keep acquisition metadata away from blinded reviewers: it contains source
and origin information. Responses must bind to the exact round packet hash.
Candidate selection alone does not complete a pilot or the human-labeled corpus.

## Bind decisions to measured targets

From `research/annotation`, use a curated annotation round whose units retain the
candidate IDs and acquisition metadata:

```sh
go run ./cmd/corpus join --root ./sources --round ./round.json \
  --feature prose-words --feature noun-token-ratio < candidates.json > joined.json
```

`join` verifies the original sources and notices again, reproduces the candidate
artifact, and compares every round target with its candidate. It measures all
corpus units with the public engine, using the plan's kinds and extraction policy.
Feature selection is explicit; unsupported IDs, duplicates, and block activation
IDs fail. Default engine limits apply and cannot be raised by this command.

The output contains decisions separately from prepared measurements. Bindings
retain the exact source, unit ID, group, partition, and measured input hash. A
round may select a subset; other targets are retained without invented labels.
Missing or uncertain judgments remain unresolved. Origin curation does not change
editorial targets, but changing text, context, source metadata, extraction, or
rights requires a matching new candidate artifact.

Instruction, policy, extraction, preparation, NLP, and feature hashes keep distinct
meanings. An unset global context set uses defaults. Per-language overrides
require an explicit set; empty sets disable extraction.
Matching uses every original segment and both target/context hashes. A bounding
span or equal unit ID alone is insufficient. An incomplete measurement fails the
whole operation; no partial JSON is written. Exit codes remain 0, 2, and 130.

The versioned research output stays `human_corpus: not_qualified`, including when
all judgments are resolved. It omits prose and rationales but retains paths,
ranges, group metadata, and numerical features. It is not a blinded packet or a
training authorization. Keep the original inputs: a saved artifact's digest does
not prove reproduction. See [ADR 0024](../../../docs/adr/0024-corpus-feature-bindings.md).

## Measure findings without labels

The `measure` command runs one pinned policy over every source of a verified
candidate artifact and records where each finding lands:

```sh
corpus measure --root ./sources --policy ./policy.yaml < candidates.json > findings.json
```

The command verifies sources and notices again first, and the policy's
extraction must equal the frozen corpus policy. Each finding attaches to every
candidate whose original segments contain its primary span, so a sentence
finding also counts for the paragraph around it; an analysis picks the unit
kind it reports.
Findings outside every candidate stay in the document totals as `unbound`.
Derived and suppressed findings are counted apart and enter no candidate.
Each document records its prose words, blocks, and counts by rule; each unit
records its cohort, kind, role, and words. The artifact pins the policy's
configuration hash, ruleset hash, scoring profile, and rule descriptors.

The output is `unswell-corpus-findings-v1` and keeps
`human_corpus: not_qualified`. A count is a rule outcome under that policy:
neither a quality judgment, nor a false-positive rate, nor recall, because no
record says whether the construction is present. The
[pattern tables](../patterns/README.md) read these artifacts together with the
[rule classes](../../methods/rule-classes-v1.json); see the
[protocol](../../methods/llm-patterns-v1.md).

## Reproduce and audit

### Reserve compression references

The `reference-bank` command builds an explicit developer artifact for the
[compression experiment](../../../docs/adr/0033-compression-reference-bank.md):

```sh
corpus reference-bank --root ./sources --round ./round.json \
  --selection ./selection.json < candidates.json > bank.json
```

The [selection schema](compression-bank.schema.json) requires the bank version,
target `kind`, `compression.level`, `compression.max_input_bytes`, `cohorts`, and
`allow_simulation`. Each cohort declares its `id`, endpoint policy `origin`, and
ordered `unit_ids`. Supported policies are `human`, `generated`, and `mixed`.
Mixed references contain both endpoint classes; topic and length matching are
separate experimental requirements. Unit order is preserved, with LF separators.

Sources, notices, candidates, and round targets are verified again. Each selected
unit needs a training partition, the selected kind, declared training permission,
and a unit-scoped endpoint claim from the round. Repository-level assertions,
unknown origin, edited origin, and mixed-origin units are not endpoint labels.
Editorial responses are not required and do not select reference membership.
Tutorial rounds require `allow_simulation: true`; their bank remains simulated.

The output contains reference prose, original source ranges, origin and rights
declarations, notice identities, compressor identities, and hashes. It is not a
blinded packet or a normal scan report. Training permission does not authorize
redistribution. The tool validates declarations and byte reproduction, not their
independent truth or scientific adequacy.

`reserved_groups` is the union across all cohorts. `reserved_targets` includes
every related candidate in those groups, including other unit kinds and targets
not selected as seeds. Empty related sources remain in the group records.
`remaining_training_units` counts unreserved training candidates of the selected
kind; it may be zero. Future fitting must apply the same exclusion union to every
cohort and baseline. This command does not train a model or alter the corpus plan.

Limits are eight cohorts, 128 ordered IDs per cohort, a 256 KiB selection file,
the primitive's 32 KiB prefix limit including separators, and a 4 MiB bank output.
Oversized targets or metadata fail without truncation or partial output. The
existing input limits and exit codes also apply. The root command test retains a
[golden projection](../../../e2e/corpusdata/compression-bank.golden.json) of reference
order, byte ranges, simulation status, and reserved source IDs.

### Build a product pack

The `pack` command converts one fitted artifact into the pack the engine loads:

```sh
corpus pack --id editorial-paragraph-v1 --min-words 30 < training.json > pack.json
```

It copies the numerical parameters and the measurement contract the artifact
already records, and adds only what an artifact cannot hold: an identifier, the
applicability floor chosen from validation data, the estimation target, and any
acceptance a maintainer is prepared to state. `--task origin_endpoint` builds a
pack for the separate origin channel instead.

Only a logistic artifact with separate isotonic calibration becomes a pack. The
result declares `experimental` unless the artifact records a qualified corpus and
`--accepted` names a published evaluation, so a simulated tutorial run cannot
produce a pack that claims acceptance. Every generated pack is loaded back
through the engine's own contract before it is written.

### Audit the source preparation

After matching a round, the [`corpus train` workflow](../training/README.md) can fit
the shared Go logistic model on permitted training targets and optionally calibrate
it on the separate calibration partition. It retains development and final test
as reserved and produces an explicitly unqualified research artifact.

`plan` records the canonical manifest digest, components, algorithm, and assignments.
`extract` revalidates that plan and checks every source and notice hash before
building candidates. The artifact records pipeline/NLP/dependency identity and a
digest of compact Go JSON with its `sha256` field omitted.

`verify` repeats extraction and compares all candidates, mappings, and exclusions.
A recalculated hash on invented text is insufficient. Matching pipeline contracts,
NLP identities, and dependency records are required. Compiler/VCS metadata may
differ; both producer and verifier identities remain visible. A local module
replacement can have empty dependency version/checksum fields; retain its commit
and clean-tree evidence separately. Reproduction on these inputs does not establish
universal compatibility or authenticate a claimed producer revision.

Successful verification reports `source_and_candidates_reproduced` and
`human_corpus: not_qualified`. It does not certify annotation quality, permissions,
provenance, unseen relationships, or scientific sample adequacy. Candidates include
source text; distribute only material whose recorded permissions allow it.

Preparation limits are 10,000 sources/units, 2 MiB per source or notice, 64 MiB of
total declared bytes, 16 MiB manifests, and 128 MiB input/output artifacts. Eligible
context is limited to 65,536 bytes and each unit to 1,024 source segments. Compact
candidate serialization has a conservative 64 MiB preparation budget, reserving
16 MiB for the manifest. Exceeding a limit fails; sources are never silently
truncated. These limits do not define a reliable statistical sample length.

Collection sizes are checked in a streaming pass before typed JSON allocation:
10,000 entries per array/object, with array limits of 100,000 for group keys,
1,024 for source segments, and 1,000 for role regions. Semantic checks then enforce
the more specific limits. These checks bound structure; they do not certify a
fixed peak process-memory budget for arbitrary inputs.

Root [e2e tests](../../../e2e/corpus_test.go) build the command with cgo disabled and
compare every Ptah candidate's text, context, role, kind, and source segments with
an inspectable [golden](../../../e2e/corpusdata/ptah-units.golden.json). Module tests
check grouping, transformations, tampering, policy exclusions, bounds, missing
files, cancellation, and annotation compatibility. Normal CI includes both tools.
