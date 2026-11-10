# Historical technical-document control transfer

Date: November 10, 2026. This is a prospective, local, non-generative
AI-origin **control** screen. It measures whether the fixed EditLens score and
an already exposed academic-prose threshold flag historical technical
documentation. It does not supply matched AI rewrites, verified human-only
authorship, or issue 349 editorial evidence.

## Source frame, frozen before any control score

Use the [DocMine v2 2021 repository list](https://zenodo.org/records/5195084),
`Repositories.csv` MD5 `035a49eded3e1cbe18f71ecc766f59ba`, only as a
repository discovery frame. The exact document text comes from pinned GitHub
Git blobs, not DocMine's concatenated text. Select candidate repositories
with at least five DocMine textual files, at least 100 recorded stars, one of
four recorded languages (C#, C++, Java, Python), and no overlap with the 20
repositories in `long-prose-v3`. Rank by SHA-256 of lowercase `owner/repo`, a
NUL byte, and `unswell-techdoc-source-pilot-v1`; inspect the first 50 in each
language (Python has 46). For each, use its latest commit whose committer date
is before 2021-08-15 UTC. Select up to eight first-party Markdown/RST blobs
per repository, ranked by SHA-256 of lowercase repository, NUL, path, NUL,
and `unswell-techdoc-document-pilot-v1`.

The frozen frame file SHA-256 is
`de5483d357981ba5f129518788afb104ad06f26d0567fd515c7c25794feca02c`;
its selected-identity SHA-256 is
`7c56001b0f64fae5ce452e52e0afece2ef4335dafaea5481efdb6c899d381299`.
The no-text `inventory.json`, `blob_audit.json`, and `eligibility.json` retain
the 196 inspected repositories, 506 attempted blobs, and every selection or
exclusion decision. The blob audit's ephemeral local source paths were removed
before freezing these metadata files; source identities and hashes did not
change. `verify_selection.py` independently checks the audit chain and
selection. The final frame records 462 selected documents from 153 repositories, 3,735 disjoint
100-word windows after the same text cleaning as the prior full-window screen,
and every exclusion reason. All selected blobs passed exact Git blob SHA-1,
raw-content SHA-256, UTF-8, source-disjointness, and 512-token checks before
any historical-control score was viewed. No source text is committed. Seven visibly
non-English pages, one table without prose, one generated test fixture, and
vendored dependencies were excluded before inference. Do not replace any
selected page or alter exclusions based on scores.

These pre-2021 source dates are strong evidence that the pages predate current
LLM writing, but they do not identify each author or rule out older automated
text generation. Report a **historical-control flag rate**, not a verified
human false-positive rate. Training-source overlap with EditLens remains
unknown.

## Frozen model and decision

Use the same public EditLens ONNX FP32 model, tokenizer, code/URL stripping,
cleaning, disjoint 100-word windows, arithmetic document mean, CPU provider,
four threads, batch size eight, and exact file hashes as the technical
full-window screen. Model SHA-256:
`ddd1173f2ef517ad499965e5029fae8099a8054a2bc76d8134e5889cc4ed1b3e`.
The threshold is fixed at `0.27447814408224075`, the strict `>` threshold
from the already published five-window academic paper calibration. This
deliberately tests transfer across both genre and window geometry. Do not
calibrate or tune on any historical technical-document score.

Score all 462 selected documents once. Report complete eligibility, flagged
count/rate, median and p10/p25/p75/p90 quantiles of the score index (linear
interpolation at rank `p * (n - 1)`), and flagged counts by language,
repository, and document-length band. Preserve every individual
no-text score, source identity, and error. Stop on the first input, checksum,
provider, shape, or inference error; do not silently drop or replace a page.

The directional **control transfer** gate requires all 462 documents to
score and **zero** flags at the fixed threshold. Zero of 462 gives a
conservative one-sided 95% upper bound below 1% for this frame under an
independent-sampling assumption; repository clustering and uncertain
authorship limit that interpretation. Any flag is a gate failure; retain
the result without post-hoc threshold changes. Passing only warrants a new
matched-rewrite screen on source-disjoint technical documents, with a
separate provenance review and generator holdout. It does not qualify a
product detector or solve editorial issue 349.
