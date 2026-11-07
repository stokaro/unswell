# Frozen paired technical-prose origin screen

This offline screen tests whether the unchanged Unswell lexical-richness and
normalized-entropy features transfer from the exposed AITDNA study to paired
conference-paper prose. The target is **AI rewriting of human prose**, not
full generation, editorial quality, or proof of an individual author's use of
AI. No generative model or remote detector is called.

## Provenance and pre-freeze contact

Use `Jerry999/AIWritingEffect` at Hugging Face commit
`6a45eddc021257b7c49de398c2fffa370c1d1bfa`, file
`dataset/t100/paragraphs.jsonl`, exactly 118,584,471 bytes and SHA-256
`6e0eb8058578afce4aaf81c822b70762afd9f18534eb5f1e89aa976cd051b793`.
The pinned `manifest.json` lists 69,821 AI, 39,459 paired-human, and 21,013
unpaired-human paragraphs; the downloaded file has 130,293 rows. The dataset
card says the source papers predate ChatGPT and that split membership is by
paper. That provenance claim is accepted as a dataset statement, not
independently audited here.

Before this protocol was sealed, the first five rows were inspected for schema
and all rows were counted by metadata only. No feature value, label-conditioned
score, or threshold was computed. The `test` split has 89 papers, 4,300
paired-human rows, and 7,015 AI rows. The `val` split has 83 papers and 4,139
paired-human rows. The `human_only` set has 400 additional papers and 21,013
unpaired-human rows. The larger `train` split is ignored by the scorer.

## Fixed scoring and comparisons

Use `research/origin/runs/2026-11-07-richness-screen/run.py` unchanged at
SHA-256 `ed17bd5b238823357a09146e18ea871abad8a1eb3ced1a6dc2b6352e1a338ad7`.
Clean each paragraph with its `prose()` function, take its first 100 ASCII
prose tokens, case-fold them, join them with spaces, and pass the result to
`features()`. This yields the mean of three 50-token windows at stride 25.
The preregistered positive direction for both features is **higher means AI
rewritten**. Paragraphs under 100 tokens abstain; do not pad, shorten, or
rescore them. Preserve `chunk_ok=false` AI rows in the primary test and report
a separate `chunk_ok=true` sensitivity result. Do not use AI rows for
calibration.

Calibrate one high-score threshold per feature solely on eligible `val`
paired-human rows: sort scores descending, let `k=ceil(0.01*N)`, and use the
`k`th score as the threshold. Flag scores **strictly above** the threshold,
which gives no more than 1% false positives on calibration rows even with
ties. Report the observed false-positive rate on eligible `test` paired-human
rows and on the separate `human_only` set. The latter is an out-of-split
negative stress test, not a calibration pool.

Match each `test` AI paragraph to its `test` human source by `pair_id`. Report
raw coverage and pair eligibility, then: (1) the fraction of eligible pairs
with AI score above the matched human score, ties worth one half; (2) pooled
directional AUROC with one copy of the matching human score per eligible AI
row; (3) AI true-positive rate at the calibrated threshold; and (4) these
statistics by generator and prompt. Grouped cells without both classes or at
least 20 eligible pairs are reported as unavailable. The same source paragraph
may occur in several pairs; the pooled AUROC therefore weights it by the
number of eligible AI variants. No score is a probability.

The registered **research feasibility** rule for lexical richness requires
all of: at least 80% of all `test` AI rows eligible with their matched human;
pooled AUROC at least 0.65; calibrated-threshold TPR at least 10%; observed FPR
no more than 2% on both `test` paired humans and `human_only` humans; AUROC
above 0.60 for at least five of seven generators and three of four prompts.
Entropy is a separately reported comparator and cannot rescue a failed
richness rule. Meeting this exploratory rule would justify another frozen
technical-document screen, not shipping an origin detector. Failure is a
negative result on this corpus; do not tune a threshold or token window on
the exposed test split.

## Failure policy and limits

Fail closed on a changed input or feature hash, malformed labels or split
metadata, duplicated human `pair_id`, an AI row without its source, overlapping
paper splits, or missing calibration. Record all abstentions and keep raw text
outside the repository. Report `chunk_ok` sensitivity and paper-level coverage
without selecting or dropping a generator based on its score. This corpus is
academic writing, not software documentation. It contains AI rewrites but no
independent token-level authorship labels for mixed prose. None of these
measurements changes the frozen 123 editorial events or gates in #349.
