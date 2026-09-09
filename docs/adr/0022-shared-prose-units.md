# ADR 0022: Bind features to the same targets as annotation

Status: accepted for implementation. Tracks #56 and #23.

The block feature collection and corpus decisions currently describe different
units. A block context hash covers grammar scope labels; a corpus context hash
covers surrounding prose. Counted token segments also differ from the complete
target segments. Comparing their bounding ranges cannot establish an exact join.

## Shared preparation

Add `nlp.PrepareUnits` below the engine, with the existing neutral provider and
document types. It receives an already extracted block and explicit limits and
capabilities. It performs no extraction, file access, model loading, or policy
discovery. Corpus preparation and the public feature API use this same function.

The selection contract is `eligible-piece-v1`, preserving corpus preparation:
split at protected NUL boundaries and trim surrounding Unicode whitespace. An
unsplit paragraph or comment can produce sentences and a paragraph. Other blocks
and split pieces produce fragments. A sentence is analyzed within its enclosing
piece; independent comments are never joined. Each piece calls the supplied NLP
provider once. Requested sentence targets reuse those tokens and tags, with local
offsets rebased onto the target. Paragraphs retain the complete piece analysis.

Prepared units own their data and expose detached blocks and bindings. Each
binding records target kind, original block identity/kind, complete target and
context segments, and separate SHA-256 identities for target prose, context prose,
and grammar scope. Prose hashes cover the exact extracted UTF-8 bytes; grammar
scope hashes cover its JSON string array. No token-only segment list substitutes
for a complete target. These bindings are descriptive; source bytes, policy,
vocabulary, NLP identity and actual capabilities must also accompany model inputs.

Preparation rejects invalid mappings and provider output, unavailable requested
capabilities, invalid unit kinds, resource exhaustion, and cancellation. It
returns no partial set on error. Limits bound input/context bytes, units, tokens,
and source segments. Tokens, chunks and dependency trees are copied without
reinterpreting them; missing required representations cannot be fabricated.

## Feature and corpus compatibility

Add `feature.MeasureUnit` and a separate unit feature contract. They reuse the
existing descriptive formulas and missing-value semantics, with an explicit
sentence/paragraph/fragment scope. The measurement hash additionally binds the
unit's target/context identity and selection contract. Unit metadata is never
converted into a numeric predictor. Raw rule activations remain block-scoped;
this API does not assign them to sentences or rerun rules on a sliced document.

Corpus preparation replaces its private piece/target selection with the shared
implementation. Its existing candidate format, target/context policy, ordering,
and source ranges remain unchanged. Saved corpus verification still re-extracts
the source. Numeric fixtures and candidate reproduction do not establish human
labels, permission to train, independent partitions, or probability accuracy.

Existing `feature.Measure`, `FeatureCollection` v1, scoring, reports and CLI/MCP
behavior retain their contracts. Adding model-unit collection to the engine and
joining accepted decisions remain explicit integration work. The first consumer
proves shared target preparation and descriptive measurement through the public
API without changing block-level rule evaluation.

## Verification

Keep corpus and existing CLI goldens unchanged. Test sentences and paragraphs
from the same source, protected gaps, repeated text at different positions,
Unicode, CRLF, entities, escaped strings, mapping changes, and surrounding prose.
Check strict provider validation, cancellation, limits, ownership and concurrent
reads. Compare scoped counts with the established formulas and verify that changes
to context or representation change the measurement identity.
