# ADR 0018: Collect rule activations with explicit applicability

Status: accepted for implementation.

## Decision

Extend the existing feature request set with `activation/<rule-id>`. The engine
validates the rule against its registered catalog. Requesting an activation never
enables a disabled rule or bypasses its policy. Existing block requests retain
their definitions, and collection remains absent unless explicitly requested.

Each selected activation is a block feature: the maximum raw fixed-point
activation among this rule's evidence occurrences in that block, divided by 1000.
Maximum aggregation counts a cluster once and cannot multiply identical emitted
evidence. It does not apply severity, score weights, rule/group caps, suppressions,
accepted debt, changed-unit selection, or derived gate diagnostics. It is not a
probability, a document score copied to sentences, or a quality label.

The input is validated emitted evidence before diagnostic deduplication. A
stronger repeated emission sets the maximum; adding the same emission again
cannot increase it.

An absent finding does not establish a measured zero. Add an optional observer to
the existing rule View. A rule reports each block it evaluated or found
inapplicable while doing its normal computation. The observer validates original
block IDs, states, reasons, duplicates, contradictions, and bounds; errors latch
like emitter errors and cannot be ignored to obtain a passing run. The rule API
retains its existing Evaluate and Emit signatures.

A descriptor can declare complete block observations. A successful evaluation
with that declaration must account for every extracted block. An evaluated block
with no evidence has an observed zero. An inapplicable block has no value and an
explicit reason. Existing rules without that declaration remain compatible:
positive evidence establishes an observed activation; missing observations and
missing evidence remain `applicability_unknown`, never an invented zero.

Disabled rules have absent values with `disabled`. Rules not reached after an
earlier failure remain `not_evaluated`. A failed evaluation or observer/emitter
error makes its values absent with `evaluation_failed`, including blocks with
partial evidence. Previously completed values may remain in an incomplete run;
the enclosing completion state still prevents treating it as complete model data.

Bind values to the existing source, context, policy, vocabulary, NLP and rule
identities. Preserve original extraction boundaries. Activation values do not
restore excluded prose, expose raw source, or require another extraction/NLP pass.
Limit stored observations and collected values before allocation, preserve
deterministic ordering, support cancellation, and own all returned data.

## Integration and evidence

Use the same feature collection and repeated CLI/MCP startup flags. Reporters
only render completed engine data. Save the activation formula contract and
participating rule versions explicitly; readers reject incompatible definitions
and contradictory applicability/value states. Old block-only reports remain
readable with their original contract.

Instrument the existing readability computations first: unsupported block kinds,
empty prose, and the ARI minimum word/sentence requirements must produce explicit
absence reasons; a measured value below the configured onset produces zero.
Other rule families require their own applicability accounting before their
negative observations can serve as dense training inputs. This remaining work is
part of #56 and cannot be replaced by assuming every silent rule was applicable.

The phrase catalog records applicability at its existing matching loop. A block
is evaluated when at least one configured phrase has a candidate window of the
required token length at an allowed position, without protected tokens or a
configured term exemption. A candidate need not match the phrase: a completed
comparison with no evidence is a measured zero. No patterns, no sentences, or no
eligible window instead produce explicit absence reasons. Document-start/end
positions retain their existing meaning across all extracted sentences,
including headings and list items. Collection does not change matching or merge
independent blocks to create candidates. These descriptors declare complete block
observations, which changes the catalog hash. The finding algorithm version stays
unchanged because the phrase lists, findings, and matching semantics stay intact;
collecting observations must not change finding or baseline fingerprints.

Only requested observations need a separate protected-token eligibility scan.
For each phrase and sentence, stop this optional scan after an eligible window
has established applicability. The matcher still checks every possible match.
The Go benchmark isolates both modes on 100,000 prepared tokens and 128 phrases;
it does not stand in for the full extraction/NLP performance budget.

Local syntax, modifier, connective, insertion, and punctuation rules also report
their applicability while computing existing counts. Length requires a sentence
with prose words. Modifier and nominalization dictionaries must be nonempty;
token checks require an unprotected, nonexempt token in the rule's inspected
scope. Noun-stack observations cover eligible words in inspected NP chunks.
Connective checks require the configured word minimum and an eligible sentence
opening. Parenthetical load requires the configured word minimum and a nonexempt
prose word; em-dash density retains its prose-kind, nonzero-denominator, and
minimum-word requirements. An applicable count below its configured activation
threshold is zero. Inapplicability records its cause instead. Nominalization
checks perform the additional eligibility comparison only when observations are
requested, stopping after the first eligible token in each sentence.

Keep each rule's existing traversal, resource accounting, evidence, and finding
version. The new observation declaration changes the catalog hash. Empty
candidate scopes cannot become negative training examples merely because an
emitter was silent. Windowed rhetoric, repetition, and list rules require their
own accounting and remain outside this local increment.

The three windowed phrase rules also declare full observations: section
announcements, empty transitions, and metaphor clusters. They reuse the existing
prose runs and phrase traversal. A configured token length fitting at an inspected
start, outside an approved term, establishes applicability even when the indexed
dictionary lookup finds no match. Opening-only rules retain that restriction;
metaphors retain matching anywhere in a sentence. Protected sentences remain
excluded in their entirety, including a matching phrase beside protected code.
Unsupported block kinds, no patterns, no sentences, and no eligible window retain
distinct absence reasons. A single allowed occurrence is an evaluated zero;
cluster activations apply only to the blocks carrying occurrences. Headings may
bridge section-announcement runs but never acquire a numeric value themselves.

Optional eligibility checks use sorted unique pattern lengths and stop after a
block has supplied a candidate. Observation state belongs to one evaluation and
is bounded by the existing block and phrase limits. Matching, candidate charges,
window grouping, source boundaries, and finding versions remain unchanged. No
observer means no eligibility scans or observation map. Paired rhetoric, other
editorial rules, repetition, and lists still require their own observations.

Vague-praise and absolute-claim observations use the same phrase eligibility
accounting after the existing question/qualification screen. A block without an
unqualified eligible window remains inapplicable. Their matcher protects each
candidate window rather than excluding an entire sentence containing code;
collection must preserve that distinction from windowed rules. A phrase beside
code may still match, while protected tokens cannot join a candidate. Prose that
extraction omitted never becomes a synthetic zero-valued unit.

Weak-intensifier observations require the existing prose-kind and word minimum,
a nonempty dictionary, and an inspected adjacent pair. Both tokens must be
unprotected, the first must be outside an approved term, and the second cannot
be one of the excluded successor words. The unchanged dictionary/POS test then
decides whether that candidate contributes to the density. With no eligible pair,
the block is inapplicable; an eligible count below the onset is zero.

Stacked-hedging observations require a nonempty dictionary and a nonexempt token
inside a clause inspected by the existing matcher. Protected tokens, punctuation,
conjunctions, and configured clause markers retain their boundaries. The same
distinct-cue count and threshold determine the activation. Empty candidate scopes
retain no-pattern/no-sentence/no-eligible-token reasons. These four descriptors
change the catalog hash but retain their finding versions, evidence, and scoring.

Acceptance includes positive and zero activations, each absence state, disabled
rules, uninstrumented external rules, ignored observer errors, partial failures,
cluster occurrences and duplicates, policy overrides, source protection,
concurrent ownership, saved-report roundtrips, all five formats, root CLI golden
cases, and CLI/MCP self-checking. Existing diagnostic and scoring behavior must
remain unchanged. Model training and qualification remain separate requirements.
