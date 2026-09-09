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

The APIs below share repetition computations and collect block measurements and
raw rule activations through the existing `RunResult`. Reporters and MCP consume
that collection. Model input contracts, Go training/inference integration, and
optional LLMDet features remain work under #56.

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

This shared preprocessing remains separate from persisted numeric measurements.
Collected rule activations record their identities and applicability as described
below. Model-specific overlap vectors must also identify their source, NLP,
normalization, and selection policies; normalized word sets alone do not establish
that compatibility.

## N-grams and surface templates

`ScanNgrams` streams the existing consecutive 3-8-word candidates in deterministic
start/end token order. A candidate needs two distinct words accepted by the
versioned lexical filter. Punctuation and protected tokens stop a candidate.
The returned indices refer to the original token sequence; the rule uses those
tokens' source spans and applies complete-term exemptions before grouping.

`NgramOptions.MaxVisits` counts each attempted token, including a boundary that
terminates a candidate. The returned visit count is valid even on error. A zero
budget permits an empty input only. Producing the last candidate does not finish
the scan if later boundary checks exhaust the budget. Callback failure,
cancellation, or a limit error invalidates partial output. Streaming avoids an
extra candidate collection; consumers must bound any groups they retain.

`PreparePOSPattern` folds the established NN/VB/JJ/RB prefixes and preserves other
tags. Punctuation and word tokens selected by the literal mask retain their
provider-normalized strings. A nil mask selects no word literals; a supplied mask
must cover the whole sequence. No additional normalization or POS inference runs.
Unavailable templates report `empty`, `imperative`, `question`, `protected_token`,
or `missing_pos`; the zero result reports `not_computed`. Invalid input and resource
errors return errors. Existing engine capability checks still reject a missing
required POS backend before rule evaluation.

Both APIs honor cancellation and explicit token/byte limits. Input tokens must
remain unchanged during a call; returned keys own their bytes. They consume
already extracted and validated tokens, without revalidating or replacing source
maps. The caller must preserve source selection, NLP, term, and preprocessing
identities. Rules still add their protected signatures and decide windows,
occurrence thresholds, and grouping. A common template alone does not imply that
two sentences say the same thing.

`PatternCatalog` defines these string preprocessing outputs under
`unswell-pattern-features-v1`. They are separate from numeric `Value` measurements.
Keys explicitly expose source-derived text and are not added to saved reports or
MCP responses. Repetition rule activations use the numeric collection below;
exporting model-specific lexical or template representations remains part of #56.

## Collect block measurements

Pass a set of block feature IDs in `unswell.Options.Features`, or repeat the CLI
`--feature` flag:

```sh
unswell check README.md --feature prose-words --feature type-token-ratio --report json:features.json
```

Unknown and duplicate IDs are errors. The engine requests and checks the selected
NLP capabilities before analysis; a requested POS feature cannot silently use an
untagged representation. Collection reads the same immutable set as enabled rules.
It does not repeat extraction, tokenization, or measurement. A nil or empty request
preserves ordinary reports without a `features` field.

The optional `RunResult.Features` uses `unswell-feature-collection-v1` and records
the block contract, canonical requested IDs, source/NLP/policy/preprocessing
identities, and each block's measured input hash, original span and counted source
segments. Values retain their own IDs, versions, units, and absence reasons.
Context is stored as a SHA256 hash of its ordered JSON representation because
grammar context can contain heading prose. The collection never saves that text.
`feature.SupportsBlock` exposes the current block-kind boundary. Unsupported
kinds carry only absent values with `unsupported_unit`; empty ratios remain absent
with `no_prose_words`. Neither is an observed zero or a clean training label.

The enclosing analysis completion state applies to the entire collection. Do not
train or infer from an incomplete run as if it were complete. Later rule or policy
processing errors can leave valid partial measurements; they do not make the run
successful. Complete policy failures may still have complete measurements.

JSON and SARIF preserve the collection. Text, Markdown, and HTML display requested
values and absence reasons from that same result. `unswell report` can render saved
values without source files or an installed model. Readers reject incompatible
collection versions, inconsistent identities, invalid ranges, and ambiguous or
nonfinite values. This validates the record's structure; it does not authenticate
an edited report or prove the underlying source content.

Start MCP with the same repeated `--feature` flags to fix its requested set.
`unswell_describe` includes those IDs, and `unswell_check` returns the shared engine
collection. A client cannot select a different set or policy in a check request.
No source words or template keys are added to reports by collection.

See [ADR 0017](adr/0017-feature-collection.md) and the rule activation contract
below. Training and inference still need compatible model input contracts;
collection does not qualify a model or provide probabilities.

## Collect rule activations

Request `activation/<rule-id>` through the same `--feature` flag or
`Options.Features` set. The ID must belong to the registered catalog, including
configured rule packs. Selecting a feature does not enable its rule. The collection
records `unswell-rule-activations-v1` and binds each source to the rule catalog hash;
the saved manifest supplies the participating rule versions and capability contracts.

For each block, the value is the maximum raw fixed-point activation emitted by
that rule for the block, divided by 1000. Collection reads validated emissions
before diagnostic deduplication, so a stronger repeated emission sets the maximum.
It excludes weights, caps, derived gate findings, suppressions and accepted debt.
The value is a policy-dependent signal, not a probability or an editorial label.

Rules can call `View.Observe(feature.BlockObservation{...})` during their normal
computation. Status `evaluated` has no reason; `inapplicable` requires a bounded
lowercase machine identifier. `Descriptor.BlockObservations` promises one explicit
observation for every extracted block when an observer was supplied and evaluation
succeeded. Missing promised observations, duplicate records, invalid IDs and
evidence for an inapplicable block are errors even when a rule ignores the returned
error. Calls are sequential within an evaluation. No observer means collection
was not requested, so the helper does nothing.

An explicitly evaluated block with no evidence has zero activation. A positive
finding can establish an activation for an existing uninstrumented rule. A silent
uninstrumented rule leaves `applicability_unknown`. Other absence reasons are
`disabled`, `not_evaluated`, `evaluation_failed`, and `inapplicable/<reason>`.
Failed evaluations discard their numeric values, including values associated with
partial findings. The enclosing run remains incomplete.

The two readability rules account for unsupported blocks, empty prose, and the
grade metric's word/sentence minimums using their existing calculations.

The eight phrase rules record each comparison opportunity: `policy.banned-phrases`,
`scaffold.chat-preamble`, `scaffold.ai-self-reference`, `scaffold.follow-up-offer`,
`scaffold.dive-in`, `filler.announced-importance`, `filler.modern-world-opening`, and
`filler.wordy-phrase`. An eligible window has the configured
phrase's token length, starts at an allowed position, and contains no protected
tokens or exempted term. At least one comparison with no match yields zero.
`inapplicable/no_patterns`, `inapplicable/no_sentences`, and
`inapplicable/no_eligible_window` distinguish blocks without a comparison.
Document-start/end use the first/last extracted sentence, including headings and
list items. Term exemptions can remove a comparison opportunity; suppressions
preserve a matched rule's raw activation. Declaring full observations changes the
catalog hash used by collected values. Finding versions and fingerprints stay
unchanged because phrase matching has not changed.

Seven additional local rules record observations in their existing calculations:

| Rules | Applicability |
| --- | --- |
| `syntax.long-sentence` | At least one sentence with prose words; otherwise `no_prose_words`. |
| `hype.modifier-cluster` | Configured phrases and an unprotected, nonexempt word to inspect. |
| `density.connective-overuse` | Configured word minimum, phrases, and an eligible sentence opening. |
| `syntax.nominalization-chain` | Prose block, both dictionaries, and an unprotected, nonexempt prose word. |
| `syntax.noun-stack` | Prose block and an eligible word in an inspected NP chunk; terms and identifiers retain their exclusions. |
| `syntax.parenthetical-load` | Prose block, configured word minimum, and nonexempt prose words. |
| `format.em-dash-density` | Prose block with a nonzero denominator and the configured word minimum. |

Missing dictionaries, sentences, or eligible tokens produce `no_patterns`,
`no_sentences`, or `no_eligible_tokens`. Unsupported prose kinds and word limits
produce `unsupported_unit`, `no_prose_words`, or `insufficient_words` as applicable.
Reasons retain the `inapplicable/` prefix in saved values. The same rule can inspect
headings for sentence length while another rule excludes them from its prose
scope. A measured zero uses the supplied POS/chunks and configured pattern
definition; it does not establish that the parse or text is correct.

Three windowed phrase rules also report applicability: `filler.section-announcement`,
`filler.empty-transition`, and `hype.metaphor-cluster`. Their existing prose runs
exclude protected sentences and unsupported block kinds. A configured phrase
length fitting at an inspected start outside an approved term establishes an
evaluated block. The first two rules inspect sentence openings; metaphors may
occur anywhere. No match, or one occurrence below the default cluster threshold,
yields zero. A cluster sets activations only on its occurrence blocks. Headings
can bridge section-announcement runs but remain `unsupported_unit` themselves.
Empty dictionaries, no sentences, or no eligible window retain their absence
reasons. Collection preserves matching and source coordinates.

Four qualifier rules also distinguish inspected candidates from omitted scopes:

| Rules | Applicability |
| --- | --- |
| `hype.vague-praise`, `hype.absolute-claim` | An eligible phrase window after the existing question/qualification screen. Code protects the candidate window, not every other phrase in its sentence. |
| `filler.weak-intensifiers` | Prose-kind and word limits, a dictionary, and an adjacent unprotected pair outside the first-token term exemption and excluded successor words. |
| `filler.stacked-hedging` | A dictionary and a nonexempt token within an inspected clause; existing punctuation, conjunction, and code boundaries remain. |

The dictionary/POS tests and counts remain unchanged. An eligible comparison below
the configured threshold yields zero. Missing eligible phrase/pair windows use
`no_eligible_window`; missing clause tokens use `no_eligible_tokens`. Empty
dictionaries, unsupported kinds, no sentences, and intensifier word limits retain
their existing reasons. Text omitted by extraction supplies no feature unit.

Six more window rules record applicability within the same prose runs:

| Rule | Eligible comparison |
| --- | --- |
| `syntax.not-only-density` | Three inspected tokens within one semicolon-delimited fragment can contain the complete marker sequence. This rule does not support term exemptions. |
| `syntax.paired-contrast-density` | Adjacent nonquestion sentences in the same block with nonexempt opening windows large enough for both fixed prefixes. |
| `syntax.whether-preface-density` | An opening window with room for the fixed prefix and an alternative token before the first inspected comma, within 32 tokens; the complete preface must be nonexempt. |
| `syntax.rhetorical-question-density` | A question with exactly a configured pattern's token length and an adjacent admissible answer in the same block. Neither whole candidate may be exempt. |
| `syntax.triad-density` | A configured dictionary and a nonexempt word inspected by the existing dictionary/POS loop. |
| `syntax.passive-candidate-density` | A sentence meeting the configured minimum and a nonexempt inspected word or matched auxiliary-participle span. |

Pairs require `window_sentences` of at least two. A smaller window is
`inapplicable/no_eligible_window`, even when the source contains the pair.
An allowed event below the density threshold is zero. Positive clusters still
activate only their occurrence blocks; independent prose does not dilute them.
Protected sentences break every one of these runs. Headings, lists, excluded
source, and independent comments retain their existing boundaries.

A passive candidate uses the sentence minimum; several short sentences cannot
satisfy it together. A visited block with no sentence reaching that minimum is
`inapplicable/insufficient_words`. Other missing windows, dictionaries, sentences,
and unsupported units retain explicit reasons. POS candidates still do not
establish grammatical voice or editorial quality.

Three repetition rules report the eligibility of their existing grouping keys:

| Rule | Eligible input |
| --- | --- |
| `repetition.exact-sentence` | A sentence meeting `min_words` with a nonempty exact key. Any protected token invalidates that sentence's key. All extracted block kinds retain their existing scope. |
| `repetition.sentence-openers` | A sentence in a paragraph meeting `min_words` with at least `opener_words` normalized words. |
| `repetition.paragraph-openers` | The same requirements, applied only to the paragraph's first sentence. A later sentence cannot make that paragraph eligible. |

A grouping key below the occurrence threshold yields zero, including a singleton.
Clusters activate their occurrence blocks. Adding an unrelated eligible paragraph
does not reduce existing activations. Suppressions keep the raw cluster and its
activations. No sentences, no inspected sentence meeting the word minimum, and no
eligible key produce `no_sentences`, `insufficient_words`, and `no_eligible_tokens`.
Opener rules report `unsupported_unit` outside paragraphs. Their word-prefix
calculation retains the provider's `Word` flags; protected code words are omitted
while other words in the same sentence remain eligible. This differs from exact
repetition's rejection of the whole protected sentence.

The remaining seven builtin rules record their actual candidate comparisons:

| Rule | Eligible input |
| --- | --- |
| `repetition.near-sentence` | Two prepared sentences reached by the existing shared-bigram index. All extracted block kinds retain their scope. |
| `repetition.paragraph-overlap` | Prepared prose blocks of the same kind reached by the content-key index within `window_blocks`. |
| `repetition.summary-echo` | An indexed comparison from an earlier nonsummary block to a later block under a selected summary heading. |
| `repetition.heading-echo` | A selected heading and its adjacent paragraph, both meeting the lexical requirements without an excluded boundary between them. |
| `repetition.ngram-density` | A nonexempt n-gram accepted by the existing scanner in a supported sentence meeting `min_words`. |
| `repetition.syntax-template` | A nonempty key from the existing POS-template computation in a supported sentence meeting `min_words`. |
| `format.list-fragmentation` | An item in a complete, eligible short unordered list that fits an inspected `window_blocks` window. |

For pair rules, preparing one unit is insufficient to supply a numeric zero. No
indexed partner, an expired comparison window, an excluded heading, or a missing
summary scope leaves `inapplicable/no_eligible_pair`. An actual candidate rejected
by the technical-contrast signature is evaluated with zero; different obligations
such as `may retry` and `must retry` retain their protection. Near repetition also
evaluates exact-copy candidates without emitting a near-match finding.

N-gram and template keys supply evaluated zeros below the occurrence threshold,
including singleton keys. Template groups still require distinct lexical
realizations to produce evidence. These grouping rules differ from pair rules:
counting an accepted key is itself their comparison opportunity.

List observations apply only to extracted list items. Ordered, task, procedural,
reference, protected, or incomplete lists yield `inapplicable/no_eligible_list`.
A complete eligible list that cannot fit its configured window yields
`inapplicable/no_eligible_window`. Allowed list counts yield zero. A code-only item
omitted by extraction has no feature unit; any surviving items still have to meet
the existing complete-list requirement.

Each rule retains its own sentence/block minimum, term exemptions, token checks,
and source boundaries. Unsupported units, empty sentences, insufficient words,
and rejected token/key candidates retain explicit absence reasons. A later
ineligible sentence cannot erase an earlier evaluated comparison in the same
block. Observations add no new candidate comparisons and do not change evidence,
scoring, or finding versions.

All 40 currently implemented builtins declare complete block observations.
External rules without observations still use `applicability_unknown`. This
completes applicability accounting for the current catalog; the broader shared
feature work in #56 and model qualification remain open. See
[ADR 0018](adr/0018-rule-activation-features.md).

The number of blocks times requested values must fit `analysis.max_candidates`
for each source. Exceeding this output bound is an operational error; values are
not sampled or silently dropped. Repository CLI/MCP self-checks collect readability,
phrase, section-announcement, stacked-hedging, noun-stack, and all six window
activations, exact-repetition and opener activations, and these seven candidate
activations, alongside word counts and lexical diversity.
