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

The engine collects these measurements when both `Options.PreparedFeatures` and
`Options.PreparedKinds` are supplied. Each is a set; unknown or duplicate entries
fail construction. `Engine.PreparedFeatureIDs` and `PreparedUnitKinds` return the
canonical selections. Supported kinds are `sentence`, `paragraph`, and `fragment`.
Rule activations remain in the block collection and cannot be requested here.

For CLI checks, repeat the flags to choose measurements and kinds:

```sh
unswell check README.md \
  --prepared-feature prose-words --prepared-feature noun-token-ratio \
  --prepared-kind sentence --prepared-kind paragraph \
  --report json:result.json
```

The MCP server accepts the same flags at startup. `unswell_describe` exposes the
fixed selections, and `unswell_check` returns `RunResult.PreparedFeatures` through
the common engine. All five report formats consume the collection. Saved reports
can be rendered without loading the source files, NLP provider, or a model.

The collection has its own version and is omitted by default. Source records
separate the full policy hash, extraction policy hash, and preparation hash. The
last hash also binds quote/structure switches and `eligible-piece-v1`. Feature
input hashes use it as the preprocessing identity. Complete target and context
segments remain distinct from counted token segments. Grammar scope hashes are
not hashes of the prose shown to annotators. Target counts record the selected
units; excluded contexts and punctuation-only pieces do not become targets.

Existing block NLP and scoring inputs stay unchanged. This optional collection
performs a separate preparation pass over the already extracted blocks; sentence
and paragraph targets share the analysis of their enclosing eligible piece.
Per-source `analysis.max_candidates` charges measurements, mapped segments, and
measured token visits. Preparation also caps each block at 10,000 targets, each
context at 65,536 bytes, and target/context maps at 4,096 segments. Existing input
limits still apply. A required collection error makes the run incomplete even
with `--no-gate`; partial source evidence is not a complete training input.

Reader validation checks the recorded contracts, identities, ranges, counts, and
values. It cannot reconstruct feature input hashes without the original inputs.
An annotation join must verify those inputs and the extraction settings rather
than trust matching field names or a parent block ID. This collection supplies
descriptive values, not human labels, qualified models, or probabilities. See
[ADR 0023](adr/0023-prepared-feature-collection.md).

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
