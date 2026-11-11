# ADR 0049: Thresholds, matcher forms, and weak rules follow review

Status: proposed on November 11, 2026. It takes effect when the maintainer
merges it.

## Context

[ADR 0048](0048-severity-follows-review.md) set severity by rule. Its test on
589 new pages reported that both reviewers accepted 59.9% of the notes. A rule
is the wrong unit for that result. Inside one rule, review separates findings
that hold from findings that do not.

- `syntax.long-sentence` reported every sentence above 35 words. On 2,630
  reviewed findings from real pages, both reviewers accepted 47.6% between 36
  and 39 words, 64.1% between 40 and 44, and 77.9% from 45 words. In table
  rows they accepted 45.0% of 109 findings: a long cell is an inventory more
  often than a sentence.
- `filler.instruction-scaffolding` reported a conditional reader goal, such as
  `If you want to change the port, you can set the option`, and a reminder that
  opens with `Be sure to`. Both reviewers accepted 18.1% of 138 such findings.
  They accepted 78.1% of the 96 findings from the rule's other forms.
- `repetition.adjacent-word` compared words without case and read them from the
  extracted text. Both reviewers accepted 15 of 32 findings. The rejected ones
  were a name beside a common noun, a heading before its sentence, or two words
  that Markdown markup separates.
- Two rules have no form that separates accepted findings from rejected ones.
  Both reviewers accepted 45.1% of 133 `filler.document-metadiscourse` findings
  and 29.2% of 48 `filler.unscoped-assurance` findings. They rejected most
  assurance findings because a neighboring sentence gave the mechanism, and a
  local rule does not read it.

## Decision

Reviewed findings are read by rule, matcher form, block kind, and parameter
band.

- **Removal.** A rule, a matcher form, or a block kind leaves the default when
  both reviewers accept fewer than half of 30 or more of its reviewed findings
  on real pages. A removed rule becomes opt-in and records the reason.
- **Threshold.** A length threshold is the lowest five-word step from which
  both reviewers accept at least 75% of the findings.
- **Severity.** ADR 0048 still decides severity. A rule whose matcher changes
  keeps the severity that its earlier version earned until the new version has
  evidence.
- **Test.** These choices are made on pages that are already open. A set of new
  pages reports the result.

Applied to the current evidence:

| Rule | Change | Version |
| --- | --- | --- |
| `syntax.long-sentence` | The onset is 45 words and the saturation 75. Table cells are not measured. | 2 |
| `filler.instruction-scaffolding` | Conditional reader goals and `Be sure to` reminders are not reported. | 19 |
| `repetition.adjacent-word` | Both copies have the same case, and no markup separates them. A sentence may open with the capitalized copy. The rule reports a note. | 4 |
| `filler.document-metadiscourse` | Opt-in. | 3 |
| `filler.unscoped-assurance` | Opt-in. | 12 |

## Consequences

On the 2,118 real pages reviewed before this decision, the default profile
reports 1,733 findings in place of 4,219. Both reviewers accepted 81.5% of the
reviewed findings that remain, against 63.0% before. These pages chose the
changes, so that share overstates the effect.

A third set of 578 new pages reports the result. Both reviewers judged the
findings of the earlier defaults and of the changed ones on the same pages.

| | Earlier defaults | Changed defaults |
| --- | ---: | ---: |
| Findings | 1,174 | 517 |
| Both accept, all reviewed findings | 64.2% | 77.3% |
| Both accept, notes | 60.6% | 72.0% |
| Both accept, warnings | 90.1% | 92.3% |

Both reviewers accepted 54.2% of the 646 dropped findings they reviewed. A
project that wants the dropped long sentences sets `onset: 35` for the rule,
and a project that wants the two rules enables them in its configuration.

`syntax.long-sentence` stays the weakest default: both reviewers accepted 67.0%
of its 276 findings on the new pages, and half of the 38 findings above 70
words, where the sentence is often a list. The
[qualification record](../research/rule-qualification.md) has the tables.

Configuration hashes, rule versions, and reported findings change.
