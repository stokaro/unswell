# Shared block measurements

The `feature` package computes descriptive measurements for enriched paragraphs,
comments, and human-readable strings. The engine constructs one immutable set when
an enabled rule declares `SharedFeatures`. Readability rules reuse that set. Code
that calls a rule directly can provide the set or let the readability rule measure
its supplied block using the same implementation.

The contract is `unswell-block-features-v1`. `Catalog` returns each formula's ID,
version, family, numeric type, unit, scope, required capabilities, normalization,
minimum evidence, limitations, and missing-value behavior. The first fourteen
features cover counts, sentence lengths, lexical diversity, POS ratios, and ARI.
They are descriptions of the input; they do not estimate revision probability or
machine authorship. Technical terminology and necessary repetition remain valid.

`Measure` accepts an enriched `document.Block`, explicit `Identity`, and `Limits`.
It does not invoke extraction or NLP. `NewSet` applies the same computation to
separate blocks with a total token and byte budget. Getters return owned values,
lengths, and source segments; changing them cannot change another consumer's result.
The input must remain unchanged until construction returns. Constructed results
retain no input text buffers and permit concurrent reads.

Only tokens marked `Word` and not `Protected` enter measurements. Character counts
include Unicode letters and digits. Frequencies use the provider's `Normal` value
without additional normalization. Sentence lengths omit sentences with no eligible
words. The standard deviation uses the population denominator and Welford's update.
POS ratios count the NN, VB, JJ, and RB tag prefixes. ARI retains its existing
formula and can be negative; no grade or probability is inferred from that value.

`Value.Number` is nil when unavailable, with a machine-readable `Reason`.
An empty block has observed zero counts and unavailable ratios. An unrequested POS
representation yields `capability_missing`, even when the provider advertises POS.
An eligible word with a missing tag after POS was requested is an input error.
Requested tokens must cover the extracted text; only whitespace and protected
NUL boundaries may fall outside their ranges. Missing token output is an error
instead of an observed zero word count. Token text and original segments must
match the supplied source map. POS ratios use the documented Penn-style prefixes;
models using these ratios must pin a compatible provider and tag convention.
Unknown IDs return errors; requesting a dependency feature cannot produce a POS
substitute. `Counts.Available` also distinguishes absent counts from observed zero.

The identity records the source, effective policy, vocabulary, preprocessing, NLP
resources, and capabilities actually requested. The engine uses the effective
policy hash for both policy and vocabulary compatibility, conservatively invalidating
the unit identity on any policy change. Direct rule evaluation identifies its
caller-supplied, unversioned NLP input explicitly. Use an engine-supplied set or
explicit versioned inputs for a persisted training or inference artifact.

Each unit hash includes these identities, the contract definitions, source segments,
kind, context, text, normalized words, tags, token boundaries, and protection flags.
Capability ordering is irrelevant. A hash describes compatibility, not a trained
model's applicability or permission to use a source. Author and generator metadata
are not feature values. Returned numeric values and source segments omit input
text; rare words and token traces are not exported by this interface.

Limits are explicit positive integers, at most `1 << 30`, with unique words bounded
by token visits. `MaxTokens` counts every input token, including punctuation and
protected tokens. `MaxUniqueWords` bounds each block's frequency map. `MaxBytes`
bounds extracted text, with token text/normalization/tag storage limited to sixteen
times that budget. Context and identity metadata have separate bounds. Set
construction also bounds the total tokens, bytes, and blocks. The engine uses its
existing candidate, file-byte, and block limits. Cache reuse still consumes each
rule's logical candidate budget. Errors and cancellation return no partial set.

Vocabulary exemptions remain specific to the rule supporting them. They do not
remove terms from descriptive word counts. Extraction exclusions and protected
tokens remain excluded. Suppressions and baseline acceptance affect later policy;
they do not change these raw measurements.

The remaining #56 work covers shared repetition computations, raw rule activation
features with applicability, optional LLMDet contracts, and engine collection for
Go training, inference, and explanations. It must use this package and the existing
`RunResult`; reporters and MCP must not compute independent feature vectors.

## Lexical overlap

`NewWordSet` builds an immutable representation of already selected normalized
words under explicit word, distinct-word, and byte limits. It performs no extra
normalization and owns its keys. The caller must select the correct source unit
and apply protected-content and term policies before constructing the set.
Empty, invalid UTF-8, and NUL-containing normalized words are errors.

`CompareWords` computes an observed intersection and union. `LexicalCatalog`
defines those two counts and Jaccard, using the same `Descriptor` and `Value`
contracts as block measurements. The lexical contract is
`unswell-lexical-features-v1`. An unavailable set cannot be compared. Two available
empty sets have observed zero counts and unavailable Jaccard with `empty_union`.
The zero value of `Overlap` has no observed numbers.

`ContentKeys` uses the existing repetition candidate filter. It removes short
keys and listed stopwords from candidate retrieval, while those words remain in
the overlap denominator. The length check uses bytes, preserving the original
formula; it is not a language or semantic-content classifier. `Words` and
`ContentKeys` explicitly expose normalized strings to library callers. They are
not added to saved reports or MCP responses.

Near-sentence rules construct one word set per eligible sentence and reuse it
for candidate comparisons. Paragraph, summary, and heading comparisons consume
the same set and overlap implementation. Existing rule policy still determines
term exemptions, windows, protected signatures, candidate charges, and clusters.
Numerical overlap does not override a difference in negation, a numerical limit,
or any other condition protected by the rule. Existing evidence values and
coordinates retain their definitions.

This shared preprocessing remains separate from the pending collection API.
Persisted training and inference vectors must identify their source, NLP,
normalization, and selection policies; normalized word sets alone do not establish
that compatibility. N-gram extraction, syntax-template preparation, raw rule
activations, and the remaining #56 collection work still need integration.
