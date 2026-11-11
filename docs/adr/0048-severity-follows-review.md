# ADR 0048: Default severity follows reviewed precision

Status: proposed on November 11, 2026. It takes effect when the maintainer
merges it.

## Context

[ADR 0047](0047-qualified-defaults-and-assistant-review.md) qualified rules by
review of their findings. Its final test, once opened, became development data.
Read by rule, it shows that one severity covers very different evidence. On
1,915 reviewed long-sentence findings from real pages, both reviewers accepted
62%. On 199 wordy-phrase findings they accepted 90%. Both rules reported
warnings, and long sentences were three of four findings.

Three rules built on generated documents also hold up less well on real pages.
Both reviewers accepted 39% of 164 instruction-scaffolding findings there,
against 69% on generated documents, and 31% of 39 unscoped-assurance findings,
against 78%.

## Decision

A rule's default severity states how far its findings have been confirmed.

- **Warning.** Both reviewers accepted at least 85% of 50 or more reviewed
  findings on real pages.
- **Note.** The rule has 50 or more reviewed findings on real pages and falls
  below that share, or it has fewer than 50 and at least 30 with a share below
  85%.
- **Unchanged.** A rule with fewer than 30 reviewed findings on real pages keeps
  its severity until it has evidence.
- **Opt-in.** The rule of ADR 0047 still applies: both reviewers reject at
  least 60% of 15 or more development findings.

Applied to the current evidence:

| Rule | Reviewed on real pages | Both accept | Severity |
| --- | ---: | ---: | --- |
| `filler.wordy-phrase` | 199 | 90.5% | warning |
| `hype.vague-praise` | 64 | 82.8% | note |
| `syntax.long-sentence` | 1,915 | 62.3% | note |
| `filler.instruction-scaffolding` | 164 | 39.0% | note |
| `filler.unscoped-assurance` | 39 | 30.8% | note |

`filler.document-metadiscourse` already reports notes. Severity changes neither
score weights nor the gate.

## Consequences

`--min-severity warning` shows the findings that review has confirmed. Notes
remain visible by default and say that the evidence is weaker.

A rule earns a warning by evidence and loses it the same way. The
[qualification record](../research/rule-qualification.md) reports each round on
pages that did not choose the severities. On 589 new pages, both reviewers
accepted 82.6% of the 115 warnings and 59.9% of the reviewed notes. The one
warning rule with enough evidence, `filler.wordy-phrase`, held at 90.6% of 85
findings. Two rules that kept warnings for lack of evidence pulled the class
down: `repetition.adjacent-word` at three of nine and
`repetition.exact-sentence` at four of six. With these findings
`repetition.adjacent-word` reaches 32 reviewed findings on real pages and
becomes a note in the next round.

Configuration hashes and reported severities change. Findings, spans, and
scores do not.
