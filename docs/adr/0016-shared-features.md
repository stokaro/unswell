# ADR 0016: Share versioned measurements between analysis consumers

Status: accepted for implementation. Tracks #56 and the protocol in ADR 0015.

The readability rules currently count words, characters, sentence lengths, word
frequencies, and POS groups independently. Training needs these same formulas.
Copying them into a research module would allow training and inference to disagree.

## Package and ownership

Add a public `feature` package below `rule` and the engine. It consumes the neutral
`document` and `nlp` contracts. It does not import rules, configuration, extraction,
reporters, or application code. It performs no resource discovery or I/O.

Each descriptor records a stable ID and version, family, numeric type, unit,
analysis scope, formula, required capabilities, normalization, minimum evidence,
and behavior when unavailable. Unknown IDs are errors. An unavailable value has a
reason and no number. Counts can be zero; undefined ratios cannot.

The first implementation moves the existing block measurements and ARI formula
without changing their definitions. An immutable set owns the results for one
enriched document. Rules declare `Descriptor.SharedFeatures` and receive that same
set in `View.Features`. Construction occurs only when an enabled rule requests it.
The two readability rules consume it; direct rule callers can construct the same
set or use the public block measurement function. Getters return owned slices.

The engine bounds total token visits using its existing candidate limit. Each
rule still charges its own logical visits, including cache hits, so sharing cannot
turn a previous resource failure into a passing scan. Distinct-word storage is
bounded explicitly. Cancellation and invalid input return errors, not partial
measurements presented as complete. No mutable model or document buffers are
shared between engine calls.

## Identity and source policy

Feature identities include the contract, descriptor definitions, NLP identity,
effective configuration, vocabulary, and preprocessing identity. A unit identity
also covers its kind, context, token boundaries and tags, source segments, and
content. A changed normalizer, vocabulary, profile, model, source, or extraction
policy must not reuse a cached vector under the previous identity.

The package receives already extracted prose. It does not restore excluded code,
URLs, strings, or comments. Protected tokens never enter counts or denominators.
Source segments come from eligible tokens; a bounding span alone does not imply
that the intervening bytes were analyzed. Independent comments remain separate.
Vocabulary exemptions are rule-specific: they must not silently remove terms from
unrelated descriptive measurements. Suppressions and baseline acceptance do not
change raw measurements or manufacture negative training labels.

## Remaining feature families and consumers

The next implementation shares lexical set construction and overlap measurements.
`feature.WordSet` owns distinct normalized words; callers supply already selected
words and explicit size limits. It performs no additional normalization or source
selection. The existing stopword filter used for repetition candidate indexing
moves with it under a separate versioned lexical contract. Numeric pair results
use the common `Value` and `Descriptor` contracts, including absent Jaccard for an
empty union. Obtaining sorted words or content keys is an explicit library call;
these strings must not be added to saved reports by default.

Near-sentence comparisons build each eligible sentence's set once and reuse it for
candidate pairs. Paragraph/summary/heading comparisons use the same set and overlap
implementation. Their existing term exclusions, protected signatures, candidate
windows, and evidence grouping remain rule policy. Reuse must not change logical
candidate charges. A mathematical overlap alone does not establish repetition of
the same information: different negation, numbers, units, and required conditions
still prevent a qualifying pair under the selected rule's guards.

This preprocessing API does not substitute for the remaining engine-wide feature
collection. Its lexical values must be paired with the caller's source, NLP,
normalization, and policy identities before use in a persisted model vector.

N-gram and POS-template preparation also belongs in `feature`. N-grams are
streamed in start/end token order, with one logical visit for every attempted
token, including a punctuation or protected-token boundary. Streaming avoids
allocating a second unbounded candidate collection. A callback failure or exhausted
budget makes the computation incomplete; callers must not accept partial evidence.
Callbacks receive owned normalized keys and half-open indices into the original
token sequence. Callers still apply complete-term exclusions and protected
signatures before grouping occurrences, then use the original token source spans.

POS preparation returns an owned surface key or a machine-readable absence reason.
It folds only the established NN/VB/JJ/RB tag families, keeps punctuation and
explicitly selected term tokens literal, and does not infer missing tags.
Questions, imperative openings, protected tokens, and absent POS keep their
existing exclusions. The caller adds the protected signature; a shared POS key
alone must never establish semantic equivalence or an editorial violation.

These two preprocessing definitions use a separate pattern contract and the
common descriptor schema. Source-derived keys are exposed only by explicit
library calls and are not numeric feature values or default report content.
Token/byte limits, cancellation, and explicit normalization apply to both APIs.
Input tokens must remain unchanged during a call; no token or source buffers are
retained. Source-map validation remains in the existing document/engine boundary.

The complete #56 contract also needs shared repetition computations and a neutral
adapter for raw rule activations. Activation features are computed after rule
evaluation, before policy suppression; they must not depend on derived threshold
findings. A disabled, incomplete, or inapplicable rule does not establish zero
activation. Keep this stage separate from pre-rule measurements to avoid a cycle.

Dependency features require a validated dependency tree and the declared scheme.
POS ratios and shallow chunks cannot replace missing dependencies. Optional
LLMDet features will declare the exact tokenizer, tables, preprocessing, coverage,
and resource identities qualified by #58. This first change does not register
unimplemented features or invent model results.

The eventual engine collection option must attach feature vectors to the existing
`RunResult`, through an explicit additive schema change. Go training, inference,
and explanations must consume that same representation and implementation.
Reporters and MCP must not rerun extraction or feature computation. Training-only
normalization is fitted on training data and versioned separately; provenance and
generator labels are not editorial inputs.

This initial implementation leaves saved result semantics unchanged, apart from
the optional descriptor declaration. Existing evidence metric names, values,
ordering, activation thresholds, and rule behavior versions remain unchanged.
Old reports remain readable; strict readers need the additive descriptor field.
It does not qualify a statistical model or close #56's remaining integration work.

## Verification

Test formulas, zero versus missing values, capability failures, Unicode and source
segments, protected boundaries, limits, cancellation, deterministic identities,
and ownership under concurrent reads. Existing readability e2e goldens must pass
without being regenerated. Exercise the shared set through an external rule and
the public engine, and retain API, schema, module, and architecture checks.
