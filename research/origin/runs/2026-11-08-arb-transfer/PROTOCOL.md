# Frozen ARB source-matched transfer screen

This is an offline test of the unchanged lexical-richness and normalized
entropy features on a **new corpus** after the AIWritingEffect results were
visible. It asks whether a higher feature value distinguishes AI wording
from a human original across three source domains and four different
generators. It does not test editorial quality, mixed authorship, or an
individual author's use of AI. No generative model or remote detector runs.

## Input and pre-freeze contact

Use the public [ARB dataset](https://huggingface.co/datasets/giper45/ARB-Dataset)
at commit `e03387f3d53efcaec9ca068c8e902e5f55611fe9`, file
`data/arb.parquet`, exactly 27,439,482 bytes, SHA-256
`66e48240618b92ce876e2097d8df52f06661a70e31791b91a95b15688a099ab0`.
Read it with Python 3.12 and `duckdb==1.4.2`. The dataset card says the
current release focuses on H2L, but the pinned Parquet actually contains
1,800 human originals and 7,200 rows in each of the free generation, human
rewrite, and AI rewrite regimes. Each of the 1,800 `pair_id` groups has
one human original and one of each machine regime for each of four
generators. Source domains are `xsum`, `wp`, and `owt`, 600 groups each.

Before this freeze, only the schema, metadata counts, first five row hashes,
and the dataset's normalized text/source SHA-256 fields were inspected.
No lexical feature score or label-conditioned threshold was computed.
The pinned dataset builder defines a text hash as SHA-256 of UTF-8 bytes after
collapsing whitespace with `" ".join(text.split())`. All 23,400 stored
text and source hashes matched that rule in the preflight audit.

## Fixed split and feature

For each source domain separately, sort the 600 human `pair_id` values by
`(sha256((pair_id + "unswell-arb-v1").encode()), pair_id)`. The first 300
groups are human-only **calibration**; the remaining 300 are **test**.
All descendants of a group stay in the same split. This yields 900
calibration humans and 900 test humans, with 3,600 test variants per machine
regime. No AI text enters calibration. The freeze manifest records a hash of
the ordered split identities and the scorer recomputes it.

Use `research/origin/runs/2026-11-07-richness-screen/run.py` unchanged at
SHA-256 `ed17bd5b238823357a09146e18ea871abad8a1eb3ced1a6dc2b6352e1a338ad7`.
The **primary** measure is `features(text)` on the entire prose document:
mean 50-token window scores at stride 25 after the existing `prose()`
cleanup and ASCII word matching. The **secondary** measure takes the first
100 cleaned, case-folded prose tokens and passes them to `features()`.
Score directions are fixed in advance: higher richness or entropy means
more AI-like wording. Abstain if the primary has fewer than 50 prose
tokens or the secondary has fewer than 100. Never pad or shorten a document.
The score is an index, not an authorship probability.

For each feature and measure, calibrate the high-score threshold on eligible
calibration humans: sort descending, choose the `ceil(0.01*N)`th score, and
flag a score only when it is **strictly greater**. This bounds calibration
FPR at 1% even with ties. Report actual FPR on eligible test humans and TPR
separately for free generation, H2L, and LLM2L. Match each test machine row
to its human original by `pair_id`; report eligible pair count, pooled
directional AUROC with one human score per machine variant, and paired win
fraction with ties worth one half. For H2L, also report these metrics by
domain and generator. Compare LLM2L with its same-generator free generation
for a separate rewrite-effect paired direction. Report raw and eligible
denominators throughout.

The predeclared **research feasibility** rule is evaluated only on primary
full-document lexical richness in H2L. It requires: at least 95% eligible
H2L-human pairs; pooled AUROC at least 0.65; TPR at least 10% under the
calibrated threshold; observed test-human FPR no more than 2%; AUROC above
0.60 in at least two of three source domains and three of four generators.
The entropy and prefix measures are comparators; they cannot rescue a failed
primary rule. Even a passing result would justify another prospective
technical-document test, not product deployment or #349 editorial acceptance.

## Failure policy and limits

Reject changed input, protocol, script, or feature hashes; missing or
duplicated quartet members; invalid text/source hashes; mismatched source
identity; malformed provenance fields; split imbalance; or insufficient
calibration. Keep raw texts outside this repository. Results are conditional
on this benchmark's source selection and prompts. The source collections
include news, creative writing, and general web text, not software docs.
Treat the paper's human-source provenance as a dataset claim rather than an
independently audited fact. Report negative results without moving thresholds,
changing token windows, or removing a weak generator after seeing scores.
No measurement here changes the frozen editorial #349 cohorts or gates.
