# AITDNA mixed-authorship origin-feature screen

This offline pilot tests two unchanged lexical features against the
[AITDNA user-study dataset](https://huggingface.co/datasets/UKPLab/AITDNA).
Its question is whether the features distinguish the study's **human-only**
setting from its **AI-assisted** setting. These are document-level study labels,
not verified token-level provenance or editorial-quality judgments. No model
calls, Unswell runtime changes, or participant texts are included in this run.

The [v1 protocol](PROTOCOL.md), [frozen script](run.py), and feature-source
hash were sealed in [freeze.json](freeze.json) before the complete dataset was
retrieved. One row was inspected for schema before the freeze, without scoring.
The four downloaded dataset-server pages were hash-frozen in
[acquisition.json](acquisition.json). They have no commit binding in the server
response. After scoring, all 362 ordered rows were compared with the exact
repository-commit parquet; every data and metadata value matched. The parquet
hash and comparison are in [provenance.json](provenance.json). The result is
[result.json](result.json), SHA-256
`eef027dfe2d9d6e488dafe42d1e87ac2e42d80327a3ff463e7d4b95406caa20d`.

| Fixed 100-prose-token prefix | Human-only | AI-assisted | Directional AUROC |
| --- | ---: | ---: | ---: |
| Lexical richness | 95 | 267 | 0.686 |
| Normalized token entropy | 95 | 267 | 0.665 |

All 362 documents were scored, with no abstentions. Both features exceeded
the registered 0.60 directional AUROC screen overall and in all three tasks
containing both classes. Within authors who have both classes, the assisted
mean was higher for 64/83 authors on richness and 62/83 on entropy. The
predeclared feasibility screen passed. These are rank comparisons; no
classifier, threshold, false-positive rate, or calibrated probability exists.

After the v1 result was visible, a [separate protocol](PROTOCOL-v2.md) and
[frozen script](sensitivity.py) were recorded in [freeze-v2.json](freeze-v2.json).
The [v2 result](result-v2.json) is an exposed-data sensitivity check. Matching
both author and task leaves 78 pairs: richness rises in 51, falls in 25, and
ties in 2; entropy rises in 49, falls in 28, and ties in 1. The positive
direction is less consistent under this tighter comparison.

AITDNA's document notion labels a document `Bot` when at least half of its
tokens are attributed to AI; otherwise it is `User`. The positive study
setting includes 214 `Bot` and 53 `User` documents. Against the 95 human-only
documents, richness/entropy AUROC is 0.720/0.696 for the `Bot` group and
0.546/0.542 for the `User` group. The latter can still contain AI-authored
content. This split suggests that the aggregate signal mostly reflects
documents with substantial AI authorship; it is weak for lighter assistance.
The split was specified after v1 and is not independent confirmation.

## Reproduction

Download `document/test-00000-of-00001.parquet` from the dataset repository at
commit `21d77f7d44833507750c0da309295c4147c1bda5` and verify its SHA-256 is
`ff98ea823d2760b4f46f3ad189414a51cfc763591eccd3eb29be088a60d9d300`.
With Python 3.12 and research-only `duckdb==1.4.2`, run from the repository
root:

```sh
python3 -B research/origin/runs/2026-11-07-aitdna-screen/replay.py \
  --parquet /path/to/test-00000-of-00001.parquet
```

The [portable replay](replay.py) verifies input and frozen-file hashes and
recomputes both saved aggregates using the unchanged feature source. The
original v1/v2 scripts retain their historical absolute feature paths and are
included as frozen acquisition records; replay is the portable entry point.
Raw dataset pages and participant text are intentionally omitted.

## Limits and next decision

Participants, tasks, and writing conditions are related, so document AUROC
does not establish generalization to new authors or technical documentation.
The task-matched and document-notion splits narrow the apparent effect.
This supports a provenance-aware origin-feature track only. A separate
source-matched technical cohort with independently known provenance, held-out
authors/repositories, and low-false-positive evaluation is needed before any
detector claim. It does not add a source-local editorial finding or change
[#349](https://github.com/stokaro/unswell/issues/349)'s frozen event gates.
