# Preparing targets for annotation and features

Corpus preparation and public feature callers use `nlp.PrepareUnits` on an already
extracted `document.Block`. Supply the existing NLP provider, requested target
kinds and capabilities, and explicit limits. Preparation performs no extraction or
resource discovery. It supports the token, sentence, POS, chunk, and dependency
representations in the current neutral document model; unsupported requests fail.

The `eligible-piece-v1` selection contract splits protected boundaries and trims
surrounding whitespace. Whole paragraphs and comments can produce sentence and
paragraph targets. Strings, headings and split pieces produce fragments. Each
piece is analyzed once. Sentence targets reuse its analysis; token offsets in
`PreparedUnit.Block()` refer to the target's own text. Independent comments never
share prose context. An empty selection contains no fabricated zero-valued unit.

`PreparedUnit.Binding()` returns complete target/context source segments and
separate hashes for target text, surrounding prose and grammar scope. Text hashes
cover exact extracted UTF-8 bytes; the grammar hash covers the JSON label array.
Identical text at different positions remains distinguishable. Complete segments
include eligible spaces and punctuation; counted token spans serve another purpose.
`Context()` exposes surrounding prose only through an explicit library call.

`feature.MeasureUnit` uses the existing descriptive formulas with the target's
explicit scope. Its input identity must match the provider and capabilities
actually used. Supply source, policy, vocabulary and preprocessing identities too.
Missing POS remains absent. The measurement hash covers these identities and the
complete target/context binding. `UnitCatalog` describes the selected scope;
the existing block `Catalog` and `FeatureCollection` retain their meaning.

Limits bound bytes, context size, targets, tokens and source segments. Failed
preparation returns no partial set. Prepared units own their data; returned blocks,
bindings, provider metadata and capability slices can be changed by their caller
without changing the prepared unit. Concurrent accessor calls are supported.

The [external consumer](../examples/consumer/units_test.go) extracts Markdown,
prepares two sentence targets and their paragraph, then measures them through the
public APIs. Corpus and CLI goldens continue to test the existing candidate ranges.

Joining accepted annotations and collecting model units through `RunResult`
remain integration work. A block's raw rule activation cannot be assigned to its
sentences by this API. Prepared units and numerical measurements do not qualify
human labels, splits, training permissions, model applicability or probabilities.
