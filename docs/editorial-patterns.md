# Contextual editorial patterns

These rules target formulaic wording that can leak from AI drafts into code and
documents. They examine the wording and its context, without attributing authorship.
All these rules are experimental; ten are disabled in builtin profiles.
`syntax.paired-contrast-density`, `syntax.repeated-reframing`, and
`filler.document-metadiscourse` ship enabled in technical and strict as zero-score
notes. It identifies repeated contrast frames for review; it does not
assert that either alternative is unnecessary. The project enables the remaining
rules for its own self-checks. Editorial precision and recall remain unmeasured.

Enable individual rules through the existing configuration:

```yaml
version: 1
rules:
  filler.section-announcement:
    enabled: true
    parameters:
      window_sentences: 8
      allowed_occurrences: 1
      saturation_occurrences: 4
  syntax.rhetorical-question-density:
    enabled: true
    parameters:
      max_answer_words: 12
overrides:
  - files: [docs/faq.md]
    rules:
      syntax.rhetorical-question-density: {enabled: false}
```

`gate: none` remains the default for these rules. Their configured weights can
contribute to local score gates. An explicit `gate: forbid` makes a rule a project
policy prohibition; it does not establish universal editorial correctness.
`hype.absolute-claim` and the three enabled construction rules have zero weight and cap
by default and only request review.
Existing [terminology exemptions](configuration.md) and reasoned
[source suppressions](suppressions.md) apply before the corresponding counts or gate.

## Measured signals

| Rule | Candidate and scope | Default measurement |
| --- | --- | --- |
| `filler.section-announcement` | Configured section-announcement phrases at sentence starts | Phrase count in 8 sentences; allow 1, saturate at 4 |
| `filler.empty-transition` | Configured empty transitions at sentence starts | Phrase count in 8 sentences; allow 1, saturate at 4 |
| `hype.vague-praise` | Configured praise phrases in affirmative prose | One lexical candidate, full activation |
| `hype.metaphor-cluster` | Configured complete metaphor phrases, never isolated nouns | Phrase count in 8 sentences; allow 1, saturate at 4 |
| `hype.absolute-claim` | Configured guarantee phrases without explicit qualifications | One lexical candidate; zero default score |
| `filler.weak-intensifiers` | Dictionary adverbs directly before an adjective/adverb | Matches per 100 prose words; minimum 20 words, onset 4, saturation 12 |
| `filler.stacked-hedging` | Distinct dictionary modal/adverb cues within one clause | Cue count; onset 2, saturation 5 |
| `syntax.paired-contrast-density` | `rather than`, `instead of`, `X, not Y`, and adjacent negative/positive about-sentence pairs | Pattern count in 8 sentences; allow 1, saturate at 4; zero score |
| `syntax.repeated-reframing` | Adjacent nominal denial/redefinition clauses with linked subjects; both clauses are evidence | Pair count in 8 sentences; allow 1, saturate at 4; zero score |
| `filler.document-metadiscourse` | Document subject plus communicative verb, or an in-document first-person announcement | Clause count in 8 sentences; allow 0, saturate at 4; zero score |
| `syntax.triad-density` | Three comma/conjunction-linked evaluative dictionary words, at least two tagged adjectives | Triad count in 8 sentences; allow 1, saturate at 4 |
| `syntax.whether-preface-density` | A whether-you-are opening with `or` before a comma in the first 32 tokens | Preface count in 8 sentences; allow 1, saturate at 4 |
| `syntax.rhetorical-question-density` | A complete configured question followed by a short prose answer in the same block | Pair count in 8 sentences; allow 1, saturate at 4; at most 12 answer words |

Window activation is `clamp((count - allowed) / (saturation - allowed), 0, 1)`,
stored in integer thousandths. One pair contributes one to the count even though
its evidence contains two source occurrences. Hedge activation uses the same ramp
with `onset` and `saturation`. Intensifier activation applies that ramp to its rate;
the evidence also reports the raw count and actual prose-word denominator.
Windows begin at candidate sentences, contain complete patterns, and emit each
candidate in at most one cluster. A two-sentence pair needs a window of at least 2.

Matches are case-folded with the existing English tokenizer and typographic
apostrophe normalization. Configured phrase lists replace the defaults. Fixed
rhetorical templates and lexical/POS roles are documented in `rules show ID`;
these rules have no dependency-parser requirement or hidden model lookup.

## Context and limits

Windowed rules can connect adjacent prose paragraphs. Section announcements also
cross grammar-recognized headings so repeated introductions in separate sections
are counted. Other windowed rules stop at headings. Lists, table cells,
and nonwhitespace source gaps end every run. Inline code does not discard its
surrounding sentence: eligible prose on either side remains in the window.
Each candidate must stay within unprotected tokens; a phrase cannot join words
across code or consume code as a keyword. Code operands may flank `, not`.
All sentences retain their original distance, including sentences mentioning code.
Independent comments and string literals form separate runs. A paired contrast or
question/answer must fit within one block. These conservative boundaries prevent
code removal or unrelated source fragments from creating a rhetorical pattern.

New rules analyze paragraph, comment, and string contexts. They do not judge a
heading or three requirements in a list as an evaluative triad. A literal landscape,
travel journey, robust estimator, single valid contrast, and a modal expressing
real uncertainty are negative examples. Whole configured metaphor phrases can
still have legitimate uses; an appropriate local exemption remains necessary.

The triad matcher requires three dictionary entries and at least two adjective POS
tags. It tolerates one noun tag because the pinned tagger can assign one to an
ambiguous modifier such as `seamless`. This remains a surface heuristic. Four-item
lists and lists made entirely of noun tags do not match. The triad and modifier
rules share the inflation correlation group and the engine's evidence caps.

Hedge clauses end at punctuation, conjunctions, protected tokens, or explicit
conditional boundaries. Repeated copies of the same hedge count once. The rule
does not prescribe removing `may`, `might`, or necessary conditions. Intensifiers
before ordinal/identity words such as `first` or `same` are excluded.

The guarantee and praise screens skip questions, explicit negation, selected
modals, and stated conditions conservatively. They do not determine truth or
understand every possible qualification. Question patterns come from the configured
catalog; an arbitrary FAQ question is not classified as rhetorical. Answers with
numbers or protected code are excluded. `max_answer_words` accepts 1 through 100.

The existing `syntax.not-only-density` is now version 3. It requires markers in
source order and uses the same structural/candidate boundaries, retaining its
existing thresholds. Its evidence marks the actual contrast span and the unit is
`patterns`, with at most one candidate per sentence. Baseline compatibility includes
the rule version; changed behavior is not hidden behind an unchanged identity.
The other windowed editorial rules and passive-candidate rule are version 2.
Comma-not recognition excludes incomplete alternatives and selected additive or
parenthetical idioms such as `not only`, `not surprisingly`, and `not to mention`.
Two necessary technical contrasts can still produce a note. Retain their meaning
when editing, or disable this advisory rule for the relevant path.

## Evidence and qualification

The new rules reuse the existing tokens, source maps, `rule.Metric`, evidence,
scoring, and `RunResult`. Reporters and MCP do not recompute them. Dictionary and
threshold changes participate in existing rule/config identities. Candidate checks
are limited by `max_candidates`. A rule that exhausts them abstains on that
document with the reason `budget_exhausted`: it reports no findings there, and
the other rules' findings stay. Token volume and emitted findings retain the
engine's independent limits. Those limits and cancellation produce an
incomplete/error result, never a silently successful required check.

[CLI fixtures](../e2e/testdata/editorial_patterns) retain exact expected detections,
related locations, and technical counterexamples. The [Go case](../e2e/testdata/editorial_code)
checks an escaped string, source permissions, and independent comment/string runs.
Both exercise CRLF and BOM. The [inline-context cases](../e2e/testdata/contextual_inline)
add annotated Markdown, MDX, and Go examples, clean rewrites, protected markers,
and independent code blocks. Library tests cover exact Unicode/Markdown coordinates,
window boundaries, terminology, rate denominators, correlation caps, cancellation,
resource errors, and concurrent reuse. Every rule also has executable catalog examples.

These are functional tests, not a human-labeled quality evaluation. The shared
model-feature registry (#56), comparative harness (#57), and rule qualification
(#26) remain open work. [ADR 0009](adr/0009-editorial-patterns.md) records the
integration boundary. No revision probability or quality improvement percentage
is claimed for these additions.

## Construction frames

[The construction evaluation](research/rhetorical-frames.md) records the source-bound
Ptah example and constructed probes. The shared internal frame representation
holds original sentence/token ranges, so related locations include both clauses
and every repetition. It does not add a second extractor or NLP provider.

`syntax.repeated-reframing` recognizes a nominal subject, copula and negated nominal
or about-complement, followed immediately by an affirmative copular clause. The
second subject must repeat the first or use `it`, `they`, `this`, `these`, `that`,
or `those`. Copular number and tense must agree. Gerunds can serve as nominal
subjects. Contractions and typographic apostrophes use the existing tokenizer.
Semicolons, colons, dashes and sentence-ending punctuation separate clauses.
Pairs stay in one block; groups can span adjacent paragraphs but stop at headings,
fences, lists, excluded blocks, and nonwhitespace source gaps. Questions do not
match. Subject matching is a shallow heuristic, not coreference resolution.

`filler.document-metadiscourse` recognizes a determiner, an optional spatial
modifier, a document noun, and a communicative verb with a following complement.
It also recognizes an `in`/`throughout` preface followed by `we` or `I`, optionally
with `will` or `shall`. Its small noun and verb vocabularies define grammatical
roles; it does not store complete announcement phrases. Inflected verbs use an
explicit vocabulary. Negated announcements and direct `See ...` links do not
match. A useful navigation sentence can match and remain worth keeping.

Both recognizers bound clauses at 48 tokens and subjects at eight tokens.
Inline code may occupy a nominal slot and appear in the source context, but its
contents cannot supply an operator, document noun, or communicative verb.
Protected subjects cannot be linked by lexical equality. Terminology exemptions
apply to whole clause ranges. Both rules report block activation through the
existing feature contract, with an evaluated zero for examined prose containing
no recognized frame. Unsupported or empty blocks remain unavailable. Candidate
budget exhaustion reports abstention; it does not produce a fabricated zero.

Neither rule establishes semantic redundancy or the truth of a claim. An index
and a gate are separate from construction recognition. Defaults deliberately
expose the constructions without assigning editorial risk points or forbidding
technical explanations. Projects can configure a stricter local policy.

## Layered instruction wording

`filler.instruction-scaffolding` version 4 follows bounded support predicates to
an action and its operand: nominal ability, generic reader enablement, modal
used-to clauses and nested intended-to-enable clauses. The diagnostic retains
the actor, action, operands and conditions. Its advice preserves possibility;
it does not change a capability into an obligation. Named actor permissions,
negation, reported claims, hazards and a bare can-plus-action remain controls.
These are surface relations over existing tokens and POS tags, not a dependency
parse or a claim of semantic equivalence.

An immediate anaphoric method can be related across two adjacent prose paragraphs.
The source ranges remain separate. Headings, lists, excluded code and unrelated
sentences break the relation; compound announcements are not linked across
paragraphs. Independent comments and string literals remain separate. Other
findings retain their block scope and existing occurrence policy. Bounds remain
48 candidate tokens and 96 sentence tokens for the new projection. Default
weights, thresholds and gates are unchanged.
