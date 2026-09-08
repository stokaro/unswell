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

Each source records a relative path, SHA-256, byte count, pinned reference, source
format, declared prose language, topic, purpose, repository, and document ID.
Optional author/template/related-version/generation-task lists hold globally scoped
IDs; empty lists mean unknown. Author-language metadata needs an independent basis.
The current contract accepts declared English prose; it does not infer fluency or
a writer's language background.

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
Explicit source assignments pin the whole component. Reordering set-valued inputs
does not change the plan. Changing declarations requires a new plan.

The planner cannot promise balanced unit counts, discover unrecorded paraphrases,
or replace a curator's review for leaks. Freeze and archive the complete source
manifest before extracting units. Do not repeatedly change the seed or inputs to
select a favorable final test. Publish independent document/repository/group counts
alongside fragment counts; nested sentences and paragraphs are related observations.

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
is an error. Otherwise, code comments and strings retain generic roles.

Whole-source provenance remains in the manifest. Candidate origin is always
`unknown` until independent unit evidence is supplied. A repository-level AI-use
statement cannot become an automatic origin or quality label.

Artifacts contain `annotation.Unit` records but no actors, responses, or
adjudications. A curator selects units, creates a round with real participant
assignments and profile instructions, and uses `annotate packet` for independent
review. Keep acquisition metadata away from blinded reviewers: it contains source
and origin information. Responses must bind to the exact round packet hash.
Candidate selection alone does not complete a pilot or the human-labeled corpus.

## Reproduce and audit

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

Root [e2e tests](../../../e2e/corpus_test.go) build the command with cgo disabled and
compare every Ptah candidate's text, context, role, kind, and source segments with
an inspectable [golden](../../../e2e/corpusdata/ptah-units.golden.json). Module tests
check grouping, transformations, tampering, policy exclusions, bounds, missing
files, cancellation, and annotation compatibility. Normal CI includes both tools.
