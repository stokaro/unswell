# Role strata, September 11, 2026

This directory records the pattern tables of the current measurement with
a stratum per document role in every row. It is the role item of
[ADR 0037](../../../../docs/adr/0037-diagnostics-not-authorship.md): a
comment, a README, and a reference page are not compared against one
pooled rate. The inputs are the 486 finding artifacts of the measurement
after the [second generation run](../../../generation/runs/2026-09-11-run2/README.md)
joined the controlled cohort; `digests.json` names each one. Every number
is a rule outcome under one policy on one cohort and role. It is not a
false-positive rate, not recall, and not a decision.

## What ran

```sh
corpus analyze --classes research/methods/rule-classes-v1.json --findings ... > tables.json
corpus analyze --classes research/methods/rule-classes-v1.json --findings ... \
  --unit-kind paragraph > tables-paragraph.json
corpus analyze --classes research/methods/rule-classes-v1.json --findings ... \
  --unit-kind sentence > tables-sentence.json
```

The tables keep their version and every earlier field. Each cohort row
gains `roles`, one stratum per role the rule admits. A stratum holds the
documents, components, words, findings, and counted units of that role,
and the share with a finding. Its interval is the same cluster bootstrap
the row carries, over the components that hold documents of that role.
The strata add up to the row.

## Roles by cohort

The document tables give the size of each stratum.

| Cohort | Comment | Documentation | README | Release note |
| --- | --- | --- | --- | --- |
| Historical | 3,335 documents, 809,650 words | 294 documents, 182,289 words | 71 documents, 44,422 words | 27 documents, 84,803 words |
| Contemporary | 2,745 documents, 748,813 words | 606 documents, 228,108 words | 70 documents, 27,255 words | 27 documents, 141,656 words |
| Controlled | 800 documents, 15,920 words | none | none | none |

The controlled cohort holds comments only, because every task came from a
comment paragraph. Its like-for-like baseline is the comment stratum of the
historical cohort, not the pooled row.

## Paragraph prevalence by role

From [tables-paragraph.json](tables-paragraph.json), the share of
paragraphs with a finding in the two largest roles, with the 95% cluster
interval.

| Rule | Cohort | Role | Paragraphs with finding |
| --- | --- | --- | --- |
| `syntax.long-sentence` | historical | `comment` | 1,290 of 41,161 (3.1%, [2.0, 4.7]) |
| `syntax.long-sentence` | historical | `documentation` | 115 of 3,767 (3.1%, [2.0, 4.8]) |
| `syntax.long-sentence` | contemporary | `comment` | 1,231 of 33,557 (3.7%, [2.2, 5.7]) |
| `syntax.long-sentence` | contemporary | `documentation` | 109 of 4,740 (2.3%, [1.4, 3.3]) |
| `syntax.long-sentence` | controlled | `comment` | 9 of 529 (1.7%, [0.0, 5.3]) |
| `syntax.passive-candidate-density` | historical | `comment` | 422 of 41,161 (1.0%, [0.8, 1.3]) |
| `syntax.passive-candidate-density` | historical | `documentation` | 133 of 3,767 (3.5%, [2.4, 5.2]) |
| `syntax.passive-candidate-density` | contemporary | `comment` | 337 of 33,557 (1.0%, [0.6, 1.5]) |
| `syntax.passive-candidate-density` | contemporary | `documentation` | 171 of 4,740 (3.6%, [2.6, 4.6]) |
| `syntax.passive-candidate-density` | controlled | `comment` | 7 of 529 (1.3%, [0.3, 2.4]) |
| `syntax.parenthetical-load` | historical | `comment` | 333 of 41,161 (0.8%, [0.6, 1.1]) |
| `syntax.parenthetical-load` | historical | `documentation` | 118 of 3,767 (3.1%, [1.3, 6.8]) |
| `syntax.parenthetical-load` | contemporary | `comment` | 171 of 33,557 (0.5%, [0.3, 0.8]) |
| `syntax.parenthetical-load` | contemporary | `documentation` | 143 of 4,740 (3.0%, [1.3, 5.2]) |
| `syntax.parenthetical-load` | controlled | `comment` | 6 of 529 (1.1%, [0.0, 3.5]) |
| `readability.grade-metric` | historical | `comment` | 820 of 41,161 (2.0%, [1.3, 2.9]) |
| `readability.grade-metric` | historical | `documentation` | 65 of 3,767 (1.7%, [0.8, 3.3]) |
| `readability.grade-metric` | contemporary | `comment` | 763 of 33,557 (2.3%, [1.4, 3.4]) |
| `readability.grade-metric` | contemporary | `documentation` | 89 of 4,740 (1.9%, [1.1, 2.8]) |
| `readability.grade-metric` | controlled | `comment` | 0 of 529 (0.0%, [0.0, 0.0]) |
| `repetition.ngram-density` | historical | `comment` | 87 of 41,161 (0.2%, [0.1, 0.3]) |
| `repetition.ngram-density` | historical | `documentation` | 9 of 3,767 (0.2%, [0.1, 0.5]) |
| `repetition.ngram-density` | contemporary | `comment` | 52 of 33,557 (0.2%, [0.1, 0.2]) |
| `repetition.ngram-density` | contemporary | `documentation` | 11 of 4,740 (0.2%, [0.1, 0.4]) |
| `repetition.ngram-density` | controlled | `comment` | 3 of 529 (0.6%, [0.0, 2.1]) |

## What the strata show

Genre carries more of the variation than era for two rules. Passive
candidates sit near 1.0% of comment paragraphs and near 3.5% of
documentation paragraphs in both eras. Parenthetical load sits near 0.5%
to 0.8% in comments and near 3% in documentation in both eras. A pooled
row, where comments are nine paragraphs in ten, hides the documentation
rate. Long sentences move in opposite directions by role between the
eras: up in comments, down in documentation, each inside its interval.

The earlier generation records compared the controlled comments against
the pooled historical row. Against the historical comment stratum the
picture is the same: every difference sits inside the interval, and none
reaches the protocol's minimum useful difference of three points.

## What is in this directory

- `tables.json`, `tables-paragraph.json`, `tables-sentence.json`: the
  document, paragraph, and sentence tables with role strata.
- `digests.json`: SHA-256 of every input finding artifact and of the three
  tables.
