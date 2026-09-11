# Document sets, September 11, 2026

This directory records the first sources outside a repository that
entered the research corpus. It is the source item of
[ADR 0037](../../../../docs/adr/0037-diagnostics-not-authorship.md).
Three sets of RFC texts from the RFC Editor joined the historical cohort
under the corpus contract. The contract asks for a publication month
that two records state, a copyright statement that permits copying, a
commit identity that digests the raw texts, and the selection rules of a
repository. An
RFC text does not change after publication, so its date is a property
of the text and not of a tag. A date proves nothing about who wrote the
text, and nothing here is an authorship claim.

## What ran

```sh
bash scripts/acquire-documents.sh
bash scripts/measure-corpus.sh --resume
```

The script fetched the RFC index and each text. It read the publication
month from the index entry and from the header of the text, and it took
the last day of that month as the snapshot date. It removed page headers,
footers, form feeds, and the dotted table of contents, wrote the texts at
the top level of a set directory, and compiled a `LICENSE` from the
copyright sections. `records/` holds the three acquisition records and
the log entries. The resumed measurement planned every shard again and
measured the eight new ones, one text each.

| Set | Texts | Words | Date | License |
| --- | --- | --- | --- | --- |
| `ietf/http-1.1-1999` | RFC 2616 | 54,683 | 1999-06-30 | Full Copyright Statement of the Internet Society, 1999 |
| `ietf/http-1.1-2014` | RFC 7230 to 7235 | 84,160 | 2014-06-30 | IETF Trust Legal Provisions, BCP 78 |
| `ietf/tls-1.3-2018` | RFC 8446 | 38,218 | 2018-08-31 | IETF Trust Legal Provisions, BCP 78 |

Words are counted in the stripped texts. Every set is dated by the index
entry and the header of each text, so every snapshot is `corroborated`.

## What the measurement did

| Text | Words | Outcome |
| --- | --- | --- |
| RFC 2616 | 54,683 | failed: the candidate budget of the frozen policy ran out on `format.em-dash-density` |
| RFC 7230 | 26,054 | failed, the same budget |
| RFC 7231 | 29,972 | failed, the same budget |
| RFC 7232 | 7,015 | measured: 248 paragraphs, 171 findings |
| RFC 7233 | 5,376 | measured: 260 paragraphs, 111 findings |
| RFC 7234 | 11,353 | measured: 466 paragraphs, 255 findings |
| RFC 7235 | 4,390 | measured: 169 paragraphs, 82 findings |
| RFC 8446 | 38,218 | failed, the same budget |

The four longest texts fail whole. The frozen E1 policy keeps the
default candidate budget of 100,000. A text of 26,000 words or more
exhausts it on the em-dash rule, and the exhaustion fails the document by
design. The four texts of 11,400 words or fewer completed. The [Ptah review](../../../reviews/2026-09-11-ptah/README.md)
recorded the same limit on a source file. A larger budget changes the
policy identity of every finding artifact, so it waits for a full
measurement. The four texts stay in the corpus with their shards and
enter no table.

## What the four texts add

The four measured texts are 1,143 paragraphs of the historical
documentation stratum, which held 3,767 paragraphs of repository
documentation before. From [tables-paragraph.json](tables-paragraph.json)
and the [role run](../2026-09-11-roles/README.md) before it, the share of
paragraphs with a finding in each part:

| Rule | Repository documentation | RFC 7232 to 7235 |
| --- | --- | --- |
| `syntax.long-sentence` | 115 of 3,767 (3.1%) | 159 of 1,143 (13.9%) |
| `readability.grade-metric` | 65 of 3,767 (1.7%) | 133 of 1,143 (11.6%) |
| `syntax.passive-candidate-density` | 133 of 3,767 (3.5%) | 112 of 1,143 (9.8%) |
| `syntax.parenthetical-load` | 118 of 3,767 (3.1%) | 96 of 1,143 (8.4%) |
| `repetition.sentence-openers` | 34 of 3,767 (0.9%) | 32 of 1,143 (2.8%) |
| `repetition.paragraph-openers` | 21 of 3,767 (0.6%) | 20 of 1,143 (1.7%) |
| `repetition.exact-sentence` | 7 of 3,767 (0.2%) | 14 of 1,143 (1.2%) |
| `filler.wordy-phrase` | 9 of 3,767 (0.2%) | 5 of 1,143 (0.4%) |

A specification written in 2014 carries three to seven times the rate of
long sentences, high reading grade, passive candidates, and parenthetical
load of repository documentation. The scaffold, hype, and filler rules
stay at zero in both. This is the comparison base at work: a construction
this frequent in a human specification marks nothing by itself, and a
finding on a specification has to be read against its genre. The
documentation role now pools two genres. A `specification` role is the
follow-up, and the stratum will need it before the rates above enter any
comparison.

## What is in this directory

- `records/`: the three acquisition records and their log entries.
- `tables.json`, `tables-paragraph.json`, `tables-sentence.json`: the
  document, paragraph, and sentence tables of the measurement after the
  sets joined.
- `digests.json`: SHA-256 of the eight shards, the three records, every
  finding artifact, and the three tables.
