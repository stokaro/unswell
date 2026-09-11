# Specification stratum, September 12, 2026

This directory records the first full measurement under Amendment 2 of
the [pattern protocol](../../../methods/llm-patterns-v1.md). The
amendment admits a `specification` role, and the three RFC document sets
of the [document run](../2026-09-11-documents/README.md) now carry it. It
also raises the candidate budget of `policy-e1.yaml` to 2,000,000, so the
four texts that failed whole in that run are measured. Every finding
artifact carries the policy identity, so the tables come from one run of
every shard, never from a mix of policies. Every number is a rule outcome
under one policy on one cohort and role. It is not a false-positive rate,
not recall, and not a decision.

## What ran

```sh
bash scripts/acquire-documents.sh
bash scripts/measure-corpus.sh
```

The document sets were acquired again so their records declare the role;
`records/` holds the three records. The measurement planned and measured
all 494 shards. Nine documents fail as before, each on a block of
non-Latin prose; no document fails on a budget. The four texts of 26,000
words or more entered the tables, and so did the five repository files
that the old budget had failed on the near-sentence, parenthetical, and
em-dash rules.

## The specification stratum

The historical cohort now holds a `specification` stratum of eight texts,
168,503 words, and 5,780 paragraphs. The
documentation stratum returns to its 3,767 paragraphs of
repository documentation. From [tables-paragraph.json](tables-paragraph.json),
the share of paragraphs with a finding in each, with the 95% cluster
interval:

| Rule | Documentation | Specification |
| --- | --- | --- |
| `readability.grade-metric` | 65 of 3,767 (1.7%, [0.8, 3.3]) | 936 of 5,780 (16.2%, [13.4, 18.1]) |
| `syntax.long-sentence` | 115 of 3,767 (3.1%, [1.9, 4.8]) | 829 of 5,780 (14.3%, [11.6, 18.1]) |
| `syntax.passive-candidate-density` | 133 of 3,767 (3.5%, [2.4, 5.2]) | 765 of 5,780 (13.2%, [10.8, 14.4]) |
| `syntax.parenthetical-load` | 118 of 3,767 (3.1%, [1.2, 6.6]) | 485 of 5,780 (8.4%, [6.2, 11.3]) |
| `repetition.sentence-openers` | 34 of 3,767 (0.9%, [0.5, 1.4]) | 224 of 5,780 (3.9%, [3.1, 4.2]) |
| `repetition.paragraph-openers` | 21 of 3,767 (0.6%, [0.3, 0.9]) | 106 of 5,780 (1.8%, [1.6, 2.0]) |
| `filler.wordy-phrase` | 9 of 3,767 (0.2%, [0.1, 0.5]) | 89 of 5,780 (1.5%, [0.9, 1.8]) |
| `repetition.exact-sentence` | 7 of 3,767 (0.2%, [0.0, 0.4]) | 54 of 5,780 (0.9%, [0.5, 1.3]) |
| `readability.long-paragraph` | 5 of 3,767 (0.1%, [0.0, 0.3]) | 47 of 5,780 (0.8%, [0.4, 1.1]) |

The scaffold, hype, transition, and hedging rules stay at zero in both
strata. The other rules sit below one percent in both.

## What the stratum shows

Specifications written between 1999 and 2018 carry three to nine times
the rate of long sentences, high reading grade, passive candidates, and
parenthetical load of repository documentation. The intervals of the two
strata do not touch. The genre carries that difference, not the era,
because every text is historical. That is what the comparison base is
for.

A finding of these four rules on a specification says that the text
reads like a specification. A rule that fires on a tenth of the
paragraphs of RFC 7231 names a construction the genre uses on purpose.
The stratum has eight components, one per text, because the sets share
no provenance key. A set-level component would need the publisher as an
author key, which the record does not carry.

## What is in this directory

- `records/`: the three acquisition records with the declared role.
- `tables.json`, `tables-paragraph.json`, `tables-sentence.json`: the
  document, paragraph, and sentence tables of the full measurement.
- `digests.json`: SHA-256 of the records, of every finding artifact, and
  of the three tables.
