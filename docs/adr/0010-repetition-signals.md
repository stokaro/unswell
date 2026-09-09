# ADR 0010: bounded repetition signals

Status: accepted for experimental implementation

Issue #18 extends the existing engine with five opt-in signals: repeated n-grams,
POS templates, paragraph overlap, heading echoes, and configured summary echoes.
They use the shared tokens, source maps, terminology, evidence, and scoring groups.
They do not infer authorship or semantic equivalence. Corpus qualification remains
in #26, and the shared model-feature registry and comparative protocol remain in
#56 and #57.

N-grams contain 3 through 8 consecutive prose words by default. Punctuation, code,
and sentence boundaries cannot create new phrases. A bounded sentence window groups
nonoverlapping occurrences; nested phrases with the same occurrence cluster retain
the longest evidence. Activation depends on excess occurrences, with the actual
prose-word denominator and phrase coverage reported separately. POS templates are
a weaker signal and require multiple lexical realizations of a repeated tag shape.
List items and imperative openings are excluded from that signal.

Paragraph comparisons use a bounded inverted index and exact set Jaccard. Related
locations form connected clusters of qualifying pairs; a cluster does not assert
that every possible pair passes the threshold. Heading echoes compare a selected
heading with its first paragraph. Summary echoes require a selected heading whose
complete normalized text matches the configured heading list, and compare its
paragraphs with preceding nonsummary prose in the configured block window.

The protected signature retains negation, modal and condition cues, numeric units,
versions, identifier candidates, and approved technical terms. Different signatures
cannot produce an overlap or template match. This is a conservative lexical check,
not a proof that equal signatures preserve meaning. Terms do not supply repetition
evidence. Units containing protected code are ineligible for overlap comparisons.

The additive `Descriptor.RequiresStructure` flag requests grammar-derived block
context through the existing extractor. It is separate from NLP capabilities. The
engine honors it only for enabled rules, including file overrides. It never restores
excluded prose blocks. Heading/summary rules require an actual selected heading,
so an excluded heading's context label cannot activate a rule.

The additive parameters `min_ngram_words`, `max_ngram_words`, and `window_blocks`
are optional in saved JSON. Existing strict readers may reject a report containing
the new fields; API snapshots and the schema document this alpha compatibility
boundary. Rule dictionaries and thresholds remain explicit configuration.

Index entries, candidate traversal, and comparison work consume `max_candidates`.
Exhaustion is an error, with no hidden sampling or partial successful gate. Window
expiration bounds active postings. Evaluation observes cancellation and uses only
call-local buffers. Tests cover source coordinates, technical counterexamples,
term exemptions, structural policy, clusters, window boundaries, and scaling.

Issue #102 extends the shared signature to the existing `repetition.near-sentence`
rule and advances that rule to version `2`. This corrects lexical matches across
modal, condition, state, and adjacent numeric-unit contrasts without changing the
candidate index, similarity threshold, cluster construction, or source mapping.
Existing negation, number, and identifier switches keep their meanings; number
protection now includes the following token. The fixed contrast cues always apply.
Term exemptions and public configuration fields do not change. Rule identity and
catalog changes intentionally invalidate old baseline compatibility, while saved
reports preserve their recorded versions. This remains a lexical safeguard with
documented limits, not a claim of semantic equivalence.
