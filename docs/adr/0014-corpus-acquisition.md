# ADR 0014: frozen source groups before corpus extraction

Status: accepted for acquisition tooling; human corpus acceptance remains open.

## Decision

Extend the existing research annotation module with corpus preparation. A curator
records original source hashes, permissions, provenance, grouping metadata, and
the extraction policy. The planner connects related sources before any prose is
extracted. A component receives one training, development, calibration, or final
test assignment. Explicit assignments take precedence and conflicting assignments
within a connected component are errors.

Repository, document, author, template, generation-task, and related-version keys
connect sources. Identical source hashes also connect them. Unknown author or
template metadata stays unknown; it does not establish independence. The plan
records connected components and the source manifest digest. A changed manifest
requires a new plan, even when source names are unchanged.

The deterministic assignment uses a versioned hash algorithm, a recorded seed,
and explicit partition weights. It does not promise balanced fragment counts or
independent repositories in every partition. Publish the achieved support and
missing partitions; do not split a large component to make a target count pass.

Extraction consumes that frozen plan and exact original bytes. It calls the
existing public `extract` and `nlp/english` packages. Corpus preparation owns
unit selection, context, and acquisition metadata; it implements no second parser,
tagger, scoring system, or editorial classifier. Protected boundaries separate
units and cannot be removed to join unrelated prose.

Candidate artifacts contain unlabeled units and source mappings. They can exist
before human participants are assigned. They carry no invented actors or judgments
and cannot satisfy the annotation or corpus-size requirements of #22. Importing
real responses later must bind them to the exact annotation packet under #21.

The developer command reads local files beneath an explicit source root and checks
hashes and byte counts. It never runs the source or retrieves URLs. The library
receives bytes from its caller and performs no filesystem or network access.
Artifact verification repeats extraction with the declared compatible pipeline
and compares the complete candidate artifact, including source segments.

## Consequences and boundaries

All related Ptah versions remain in development because the project has already
been inspected while developing Unswell. The maintainer's AI-use statement stays
repository-level provenance, with unknown unit origin unless direct evidence exists.
It does not establish editorial quality or independent final-test performance.

Grouping checks establish consistency with recorded relationships and identical
sources. They cannot discover every unrecorded paraphrase, shared template, author,
or generation prompt. Curator review and later near-duplicate checks remain
required; a clean manifest is not proof that no semantic leakage exists.

Source permissions, provenance, English-language declarations, and participant
identity remain assertions requiring an audit. Source bytes and mappings can be
verified mechanically; the tool must not present that verification as an audit of
all other assertions. Full #22 acceptance still requires the real human corpus,
adequate independent groups, and the frozen evaluation protocol.
