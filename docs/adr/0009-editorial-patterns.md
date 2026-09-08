# ADR 0009: Contextual editorial patterns

Status: accepted.

Issue #17 completes the filler, hype, and rhetorical portion of the specification
through the existing rule engine. Eleven new rules remain experimental and disabled
in every builtin profile until explicitly enabled. Their thresholds are proposed
policy, not corpus-qualified defaults or probabilities. Existing directives keep
their universal syntax.

The rules consume the shared `document` tokens, POS tags, and original source spans.
Common matchers produce bounded occurrence clusters and `rule.Metric` measurements;
the existing engine owns scoring, correlation caps, suppressions, baseline, and all
reporting. No independent NLP or research implementation is introduced. Rule IDs,
versions, parameter dictionaries, capability requirements, and metric definitions
identify these activation features. The broader registry and model compatibility
contract remain tracked by #56; comparative qualification remains #57/#26.

The existing `syntax.not-only-density` moves to version 2 and shares the bounded
window implementation. It requires the markers in source order and stops at
protected or structural boundaries. Its default thresholds stay unchanged; the
version change makes baseline and result compatibility explicit.

Phrase-pattern windows count measured occurrences, allowing one isolated instance.
They span adjacent prose paragraphs but stop at headings, lists, tables, protected
gaps, and independent source comments or strings. Paired constructions must occur
within one block. A rolling window emits each occurrence in at most one cluster.
There is no document-wide all-pairs comparison. Candidate checks and stored matches
are bounded by `max_candidates`, and evaluation observes caller cancellation.

Dictionary matches cannot cross protected tokens. Curated evaluative adjectives
and POS roles distinguish modifier triads from requirement lists. A triad requires
at least two adjective tags and allows one noun tag for an ambiguous dictionary
modifier; this limitation is explicit. Hedge clusters
stay within a clause and allow two cues; suggestions retain meaningful uncertainty.
Intensifiers require an adverb role before an adjective or adverb, a minimum prose
length, and a declared rate per 100 prose words. Absolute-claim candidates exclude
questions, explicit negation, modals, and stated conditions conservatively; the
message asks the author to verify scope without declaring the claim false.

The rhetorical-question rule matches a configured question catalog followed by a
short answer. It does not infer whether arbitrary questions are rhetorical. FAQ
and reference material can leave it disabled or use existing file overrides.
The additive `max_answer_words` parameter has explicit bounds and is omitted from
serialized existing settings when zero. API and report schemas record the addition.

Rule descriptors and the catalog documentation specify formulas, units, lexical
scope, and limitations. Positive, negative, boundary, source-mapping, resource,
and correlation tests provide functional evidence. They do not substitute for
human corpus annotation, measured precision, or held-out qualification.
