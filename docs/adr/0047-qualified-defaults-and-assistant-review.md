# ADR 0047: Qualified rule defaults and review by the calling assistant

Status: proposed on November 11, 2026. It takes effect when the maintainer
merges it.

## Context

[ADR 0046](0046-assistant-panel-corpus.md) asked for a corpus annotated by two
assistants and a local model fitted on it. The corpus now exists. The
[qualification record](../research/rule-qualification.md) reports its
measurements. Three of them decide what follows.

- The engine was close to silent. On 772 real documentation pages it reported
  1,489 findings, and 1,224 of them were long sentences. Six textbook sentences
  of machine-style praise produced no finding. Phrase rules shipped with three
  to seven example phrases, and 24 of 58 rules were opt-in because no corpus
  precision had been measured.
- A fitted word model does not carry to real pages. With repositories held out,
  it reached 18.7% recall at 85% precision on generated documents. Trained on
  those documents, it reached no recall at that precision on 164 real pages.
  Sentence embeddings did worse than word n-grams, so the learnable signal is
  lexical.
- Two strong assistants agree on which sentences need revision with a Cohen
  kappa of 0.43. Most sentences they both mark are contextual: a claim without
  scope, a repeated statement, a frame that adds nothing.

## Decision

**Stop the fitted model.** The stop rule of ADR 0046 applies on real pages. The
negative result is recorded, and no model pack ships.

**Qualify rules by review of their findings.** Two assistants from different
model families judge each finding under the statement of its rule. This is the
evidence that [#26](https://github.com/stokaro/unswell/issues/26) asks for under
[ADR 0041](0041-assistant-review-acceptance.md).

- A phrase joins a rule's list when both reviewers accept at least 80% of at
  least five development findings of its word family.
- A rule that ships enabled becomes opt-in when both reviewers reject at least
  60% of 15 or more development findings.
- Lists and defaults are chosen on development pages. A separate set of real
  pages is opened once and reports the result.
- A third assistant audits a sample by blind paired comparison with placebo
  edits. The audit is reported beside the review and does not replace it.

**Let the calling assistant review in context.** The MCP server gives the
calling assistant the editorial rubric and the document's items, and binds the
assistant's flags to exact source bytes. These findings are the assistant's
judgments. They are advisory, carry the label `calling assistant`, and never
reach scores or the gate. The library still reads no network and starts no
process.

## Consequences

`hype.vague-praise` ships enabled with a qualified list.
`filler.wordy-phrase` and `hype.modifier-cluster` gain qualified entries.
`repetition.sentence-openers` and `repetition.paragraph-openers` become opt-in.
Configuration hashes change.

Fixed rules cover a small part of what the annotators mark. The record states
that share. Contextual coverage depends on the assistant that calls the server,
and no claim about it is made for a run without one.

The 85% audit threshold of ADR 0046 was written for a fitted model. Rule
defaults follow the review criterion above, and the audit result is published
with its limits.

The corpus stays the development base. A rule or phrase without reviewed
findings stays opt-in or unlisted until it has them.
