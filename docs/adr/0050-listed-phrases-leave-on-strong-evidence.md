# ADR 0050: A listed phrase leaves its list on strong evidence

Status: proposed on November 12, 2026. It takes effect when the maintainer
merges it.

## Context

[ADR 0047](0047-qualified-defaults-and-assistant-review.md) lists a phrase when
both reviewers accept at least 80% of five or more reviewed findings of its
word family. A test set becomes development data once it is opened, so the
lists were read again after the third test, and three families left.

The fourth test showed what this does over time. Read with that set included,
five more families fall below 80%: `easy to use` at 16 of 21, `sophisticated`
at 15 of 20, `comprehensive documentation` at 13 of 17, `exceptional` at 9 of
12, and `the bottom line` at 4 of 6. Each is one finding away from passing.
Between the third and the fourth set, whole rules moved by ten to twenty points
in both directions, so a family of twenty findings moves more.

Removal on the share alone works one way only. A family that leaves produces no
more findings, so it cannot return when the next set would have lifted it. With
every new set the lists would lose the families that happened to fall, and
never regain one.

## Decision

A listed family leaves when both conditions hold over all reviewed findings:

- both reviewers accept fewer than 80% of five or more findings, and
- the upper end of the 95% Wilson interval for that share is below 80%.

A family that fails the first condition and passes the second stays listed and
is named in the qualification record.

Inclusion does not change. A candidate still needs 80% of five or more
development findings, and its findings on later sets decide whether it stays.

Applied to all reviewed findings, the fourth set included:

| Family | Reviewed | Both accept | Upper end | Decision |
| --- | ---: | ---: | ---: | --- |
| `state-of-the-art` | 26 | 57.7% | 74% | leaves |
| `foster` | 10 | 50.0% | 76% | leaves |
| `easy to use` | 21 | 76.2% | 89% | stays |
| `sophisticated` | 20 | 75.0% | 89% | stays |
| `comprehensive documentation` | 17 | 76.5% | 90% | stays |
| `exceptional` | 12 | 75.0% | 91% | stays |
| `comprehensive guide` | 6 | 66.7% | 90% | stays |
| `the bottom line` | 6 | 66.7% | 90% | stays |

## Consequences

`comprehensive guide` left its list after the third test on four accepted
findings of six. That is not strong evidence, and it returns.

`hype.vague-praise` keeps words that reviewers accept three times in four. It
reports notes, which is the severity for evidence below 85%
([ADR 0048](0048-severity-follows-review.md)). On the fourth set both reviewers
accepted 67.9% of its 56 findings.

The lists change more slowly than the share suggests, in both directions. A
family needs about twenty findings near 60%, or ten near 50%, before it leaves.
