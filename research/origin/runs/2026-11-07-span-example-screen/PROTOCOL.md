# Author-disjoint span-example origin screen

Status: frozen before encoder inference on November 7, 2026. This is an
offline, exposed-data feasibility screen, not a product detector or an
editorial finding for [#349](https://github.com/stokaro/unswell/issues/349).

## Hypothesis and fixed input

[ExaGPT](https://aclanthology.org/2026.findings-acl.380/) compares spans with
human and generated examples, then presents the neighbors as origin evidence.
Its published scores depend on a domain- and generator-matched datastore. Test
one smaller, fixed-window variant on the already exposed AITDNA user study:
does an author-disjoint span-neighbor margin distinguish the study's human-only
condition from its AI-assisted condition better than the existing lexical
richness feature? This is **not** a reproduction of ExaGPT: it uses a different
encoder, four fixed spans, and no learned segmentation.

The only dataset input is `UKPLab/AITDNA` commit
`21d77f7d44833507750c0da309295c4147c1bda5`, file
`document/test-00000-of-00001.parquet`, SHA-256
`ff98ea823d2760b4f46f3ad189414a51cfc763591eccd3eb29be088a60d9d300`.
Expect all 362 ordered rows. Use the same final snippet concatenation and
case-folded ASCII word extraction as the existing
`research/origin/runs/2026-11-07-aitdna-screen` analysis. Verify the source
hash and script identity before reading any labels for scoring. Retain all
rows in the coverage report. The ten peer-review rows have no human-only
study control in the same task and therefore abstain from this fixed method.
No participant text is committed.

The local discriminative encoder is the published APE code at commit
`a6723d56e70a5fc92577858c9c2a2f17a6cbfe2a`, checkpoint revision
`e6dc09deb8c9e18323f17bbfd35b9025d89d8044` with weights SHA-256
`ad1992f37a1327f9e6cd73dcbf7652660bf83d47694258e1b06bda8bba6155eb`,
and BGE-small architecture/tokenizer revision
`5c38ec7c405ec4b44b94cc5a9bb96e735b38267a`. The locally rebound APE
config and tokenizer asset hashes are checked by the runner and recorded in
the frozen input manifest. The model runs offline on CPU with four threads;
no training, service call, download, fallback, or truncation is allowed.

## Fixed representation and comparison

Take the first 100 eligible prose words of each document, as in the earlier
AITDNA feature screen. Split them into four consecutive 25-word spans. Encode
each span with the APE checkpoint's BGE encoder and mean-pool non-padding token
states, then L2-normalize. Tokenized spans over 128 tokens abstain rather than
truncate; the row remains in all coverage denominators.

For each target span, compare cosine similarity only with spans from **other
authors in the same study task**. Within each class, retain the maximum
similarity per source document, then average the five nearest distinct source
documents. The span score is AI-assisted mean minus human-only mean. The row
score is the mean of its four span scores. A row abstains if either class has
fewer than five eligible other-author source documents. This balanced margin
uses no fitted threshold, test label, or candidate-specific tuning. Retrieve
five nearest examples per class to preserve inspectable span evidence; publish
only row indices and scores, not participant text or neighbor excerpts.

## Evaluation and stop rule

Report scored and abstaining rows by task and condition, directional AUROC for
the three tasks with both classes, their overall pooled AUROC, and within-author
and within-author-and-task paired directions. Also report the Bot-majority and
User-minority AI-assisted subsets separately under the dataset's document
notion. Recompute lexical richness on the exact same scored rows using the
unchanged earlier feature source. The registered feasibility screen requires
the span score's pooled AUROC to exceed lexical richness by at least 0.05,
reach at least 0.70 overall, reach at least 0.60 on each of the three tasks,
and have a positive paired author-and-task direction in at least 60% of
non-tied pairs. This is an admission threshold for further origin research,
not a detector qualification threshold. Failure retires this fixed variant.

Study-setting labels are not token-level provenance; AI-assisted documents can
contain mostly human writing. Authors and tasks are related, and this cohort is
not technical documentation. No false-positive rate, calibration, editorial
recall, or new #349 event is inferred. A positive screen would still require
a source-matched technical cohort with independently known provenance and
held-out authors, repositories, and generator families before product use.
