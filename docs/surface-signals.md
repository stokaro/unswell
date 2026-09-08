# Surface syntax and readability

Unswell offers eight experimental rules for reviewing formulaic prose in comments,
strings, and documents. They share the existing extraction, English NLP, terminology,
evidence, and local scoring contracts. All eight are disabled in every builtin
profile. Their thresholds are configurable starting points, without measured
precision or a claim that a matching passage needs revision.

A noun-heavy API description, passive construction, long paragraph, or punctuation
choice can serve the reader. Keep conditions, negation, limits, and technical terms
when reviewing a finding. These rules do not rewrite text or determine authorship.
A future dependency backend is tracked separately in #20; model-feature versioning
and comparative qualification remain in #56 and #57.

## Rules and defaults

| Rule | Candidate condition | Parameters and defaults | Weight/cap |
| --- | --- | --- | --- |
| `syntax.nominalization-chain` | Configured verb + optional modifiers + configured common noun + `of` + a common-noun complement | `verbs`, `nouns` | 12/24 |
| `syntax.noun-stack` | Consecutive NN/NNS tokens inside a shallow NP chunk | `onset: 3`, `saturation: 7` | 8/16 |
| `syntax.passive-candidate-density` | Multiple sentences with be + up to four adverbs + VBN | `min_words: 8`, `window_sentences: 8`, `allowed_occurrences: 1`, `saturation_occurrences: 4` | 6/12 |
| `syntax.parenthetical-load` | Balanced insertion word count, pair count, or nesting | `min_words: 20`, `onset: 12`, `saturation: 36`, `allowed_occurrences: 2`, `saturation_occurrences: 5`, `allowed_depth: 1`, `saturation_depth: 4` | 8/16 |
| `readability.long-paragraph` | Both a long block and several long sentences | `onset: 120`, `saturation: 240`, `sentence_words: 25`, `min_long_sentences: 2` | 8/16 |
| `readability.grade-metric` | ARI formula above the selected level | `min_words: 50`, `min_sentences: 2`, `onset: 12`, `saturation: 20` | 0/0 |
| `format.em-dash-density` | More than the permitted number of U+2014 characters and excess local density | `min_words: 40`, `allowed_occurrences: 1`, `onset: 2`, `saturation: 6` | 0/0 |
| `format.list-fragmentation` | Multiple short, complete unordered lists in one bounded section | `window_blocks: 32`, `max_item_words: 10`, `max_list_items: 3`, `allowed_occurrences: 2`, `saturation_occurrences: 5` | 0/0 |

Count and density boundaries activate strictly above onset. Minimum word and
sentence counts are inclusive. Linear activation saturates at the configured limit;
parenthetical load uses the largest of its three activations. Existing scoring caps
and evidence deduplication still apply. Adding an independent clean paragraph does
not reduce an already measured paragraph's score.

For a local advisory check:

```yaml
version: 1
extends: [builtin:technical]
rules:
  syntax.nominalization-chain:
    enabled: true
    gate: none
    parameters:
      verbs: [perform, performs, performed, conduct, conducts, conducted]
      nouns: [evaluation, assessment, review, verification]
  syntax.noun-stack:
    enabled: true
    gate: none
  readability.grade-metric:
    enabled: true
    gate: none
    score: {weight: 0, cap: 0}
  format.list-fragmentation:
    enabled: true
    gate: none
    score: {weight: 0, cap: 0}
```

`gate: none` removes direct prohibition; a nonzero weight can still contribute to
configured index thresholds. ARI and formatting measurements default to zero weight.
An explicit `gate: forbid` is a user's policy choice, not scientific qualification.
Dictionaries contain literal word forms; no lemmatization is implied. Unknown
parameters, invalid word lists, and invalid ranges are rejected before analysis.

## Syntax and terminology

Nominalization matching requires the complete verb/noun/complement pattern, not a
suffix such as `-tion`. Up to four modifiers can precede the nominal head and up to
six can precede its complement. Noun stacks stop at punctuation, noncommon-noun tags,
approved terms, and identifier candidates. The shallow NP boundary comes from the
NLP backend; an invalid chunk range causes an operational error.

The passive candidate counts sentences, including one occurrence per matching
construction in the evidence. Independent comments, protected code, and structural
boundaries cannot be combined to reach its sentence allowance. Negating adverbs stay
in the source range. A VBN tag can also describe a state, and a tagger can mistake a
verb for a noun. No grammatical dependency or confirmed passive voice is inferred.

Parenthetical load measures balanced `()` and `[]` in extracted prose. Nested words
are counted once, while each balanced nonempty pair contributes to pair count and
nesting. An unfinished or mismatched frame is discarded; protected boundaries reset
open frames. In Markdown, link delimiters are handled by extraction and do not become
parenthetical evidence. Escaped literal brackets and brackets in source strings are
eligible. Nesting beyond 256 levels returns an error instead of truncating a result.

Configured term exemptions apply to nominalization, noun stacks, passive candidates,
parenthetical load, and list fragmentation. The readability formulas and dash density
use their documented prose denominator without term exemptions. Necessary technical
vocabulary is a reason to review these measurements, not replace terminology merely
to lower a number.

## Measurement protocol

The ARI formula is:

```text
4.71 * characters / words + 0.5 * words / sentences - 21.43
```

The constants are documented in [NIST's readability discussion, section 6.2.2](https://tsapps.nist.gov/publication/get_pdf.cfm?pub_id=914701).
Unswell implements the formula over its extracted tokens. This implementation does
not reproduce a published comprehension study or qualify the result on source code.
It uses no syllable dictionary, downloaded resources, or external process.

A word is a token marked as a word by the pinned English backend and not protected
by extraction. Characters are Unicode letters and digits within those tokens,
excluding punctuation. Sentences are NLP segments containing at least one such word.
The unrounded result uses `ARI-formula-units`; it is not an age prediction,
probability, or percentage of AI-written text. A block below its minimum length has
no ARI finding. This does not mean zero editorial risk or confirmed human authorship.

Both readability rules use shared counting functions and report:

- Word, character, and sentence counts; minimum, maximum, and mean sentence length.
- Population standard deviation of sentence word counts, computed with Welford's method.
- Distinct normalized words / words and once-occurring normalized words / words.
- Noun, verb, adjective, and adverb token counts divided by prose word count.

These are descriptive metrics attached to a matching block, not a general feature
export API or calibrated model input contract. Type-token ratio depends on length.
Lexical variety and POS distributions do not establish quality or authorship.
Readability rules require POS support to keep their reported metric set consistent.
Missing support is rejected, rather than replaced with zeros.

Dash density is the U+2014 count per 100 local prose words. Ordinary hyphens and en
dashes do not count. Source locations retain Markdown entity bytes and Go escape
sequences. A single em dash is allowed by default, including in a short passage.

## Markdown lists

List identity, item membership, nesting, markers, and task state come from grammar
nodes. The pinned grammar sometimes places adjacent `list_item` nodes directly in
a section; those siblings are grouped by their grammar marker. No line-based Markdown
parser is introduced. Metadata is requested only when structural analysis is enabled.

Fragmentation requires complete selected lists. Ordered lists, task lists, nested
lists, lists containing multiple paragraphs or protected blocks, and lists exceeding
the configured size are excluded. Items with sentence punctuation, numeric or
identifier tokens, approved terms, or recognized instruction openings are excluded.
Selected reference/procedure headings also exclude their sections; the catalog names
these headings in the implementation. These guards are limited heuristics, not a
complete classifier for API documentation or instructions.

Headings and excluded blocks separate windows. A protected or unselected item cannot
silently disappear and make a larger list eligible. One long list does not become
multiple lists. Each emitted cluster retains the original ranges of all its items.

## Validation and remaining work

The root e2e cases [surface_patterns](../e2e/testdata/surface_patterns) and
[surface_code](../e2e/testdata/surface_code) exercise the compiled CLI with annotated
expected detections and exact JSON, text, and SARIF checks. Unit and engine tests
cover formulas, counterexamples, term exemptions, boundaries, grammar membership,
source maps, invalid capabilities/chunks, limits, cancellation, and concurrent reuse.
The MCP protocol test compares passive-candidate evidence with the public engine.

Work consumes the existing `max_candidates` budget and observes cancellation.
All buffers are local to an analysis. Bounded list windows and indexed mapped ranges
avoid comparing every punctuation mark with every sentence. Resource exhaustion makes
a scan incomplete. Regular rules and index scoring remain available without a model.

Corpus precision, calibration, feature ablation, and default qualification remain
open. Repository self-checks and Ptah smoke scans verify execution and expose findings;
they are not independently annotated evaluation data.
