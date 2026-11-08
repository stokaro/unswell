# Interpretable feature ensemble: frozen offline transfer protocol

Date: November 8, 2026. This is an offline, model-free AI-origin feature
experiment, separate from the editorial diagnosis requirements in #349. The
feature-family and cross-domain design are motivated by [El Attar et al.,
Findings of AACL 2026](https://aclanthology.org/2026.findings-aacl.10/), but
this is a five-feature proxy, not a reproduction of their 284-feature system.
No generative model, remote detector, or package-level network call is used.

## Frozen inputs and provenance

- ARB: `giper45/ARB-Dataset` commit
  `e03387f3d53efcaec9ca068c8e902e5f55611fe9`, `data/arb.parquet`,
  27,439,482 bytes, SHA-256
  `66e48240618b92ce876e2097d8df52f06661a70e31791b91a95b15688a099ab0`.
  Reuse the existing verified ARB loader to check all 23,400 rows, source
  hashes, 1,800 complete source groups, and regime metadata. The primary
  machine class is `h2l`, a machine rewrite of a human source. The `free_llm`
  and `llm2l` regimes are separately reported transfer contrasts.
- AIWritingEffect: `Jerry999/AIWritingEffect` commit
  `6a45eddc021257b7c49de398c2fffa370c1d1bfa`, `dataset/t100/docs.jsonl`,
  75,432,568 bytes, SHA-256
  `428de968a71d64ddaf9f0d5afe7e4e295b2adeace77f57fd29c6252ae082d29c`.
  Verify the 2,598 row identities, split counts, and paper-disjoint roles
  already defined by that dataset. Only the 89 test originals, 143 test AI
  rewrites, and 400 `human_only` papers are scored. Keep repeated variants
  clustered by paper in interpretation.

These public corpora have been inspected and scored by earlier Unswell studies.
This is a development/transfer screen, not an untouched prospective test.
The human provenance is the datasets' assertion; do not imply independent
authorship verification. No #349 page or label is read.

## Fixed feature and fit

Use the unchanged `prose`, `WORD`, and 50-token/25-stride richness and entropy
definitions from `2026-11-07-richness-screen/run.py`. Use the fixed closed-class
lexicon from `2026-11-08-gen-edit-transfer/closed_class.txt`. Every document
is represented by its first 100 cleaned prose tokens; fewer than 100 means
abstention. The five features, in this order, are:

1. Mean 50-token-window lexical richness.
2. Mean normalized 50-token-window token entropy.
3. Closed-class-token fraction among the 100 tokens.
4. Mean ASCII token length divided by ten.
5. Repeated adjacent-token bigram coverage: the number of positions belonging
   to a bigram appearing at least twice, divided by 99. Count every occurrence.

No character, layout, punctuation, or source-rewrite-delta feature is used.
The score is a document-level index, never an authorship probability or an
editorial finding. The primary baseline is unchanged richness alone, with
the high-score direction. Both scores use exactly the same eligible rows.

For each ARB source domain, order its 600 human `pair_id` values by
`(SHA-256("unswell-ensemble-v1|" + pair_id), pair_id)`. Assign first 200 to
training, next 100 to human-only calibration, last 300 to test. Descendants
inherit the source role. Keep a training source only when its human and all
four H2L variants have at least 100 prose tokens; require at least 570 of
600 complete training sources. Standardize each feature with mean and
population standard deviation from those retained rows; zero deviation
becomes one. Fit a five-weight logistic index with
intercept using 500 full-batch gradient steps, learning rate 0.5, L2 penalty
0.02 on weights, and zero initialization. For `N` retained training sources,
each human has weight `0.5/N`; each H2L variant has weight `0.5/(4*N)`.
No test or academic row
influences standardization or fit.

For each score, sort the eligible calibration-human scores descending,
choose the `ceil(0.01*N)`th score as threshold, and flag only strictly greater
scores. Require at least 285 of 300 calibration humans. This bounds observed
calibration FPR at 1% even with ties. Do not
recalibrate in either test corpus, change feature direction, select a cutoff
from machine rows, or tune after seeing results.

## Fixed reports and decision

Report source and variant eligibility, abstentions, source-disjoint ARB test
human FPR, H2L/free/LLM2L TPR, pooled directional AUROC and matched-pair win
fraction for the ensemble and richness baseline. Break out H2L by ARB domain
and generator. For AIWritingEffect, report test-human and `human_only` FPR,
AI-rewrite TPR, pooled AUROC, and matched-pair wins, including per-generator
counts. Report exact positive/negative denominators and threshold scores.

The **research advance screen** passes only if the ensemble, on the fixed ARB
test, has at least 95% eligible H2L pairs, AUROC at least 0.70, TPR at least
10% at the human-calibrated threshold, observed test-human FPR at most 2%,
and no lower AUROC or TPR than richness on that same test. It must also have
at least 95% eligible academic test pairs, AIWritingEffect AUROC at least
0.70, TPR at least 10%, test-human FPR at most 2%, and `human_only` FPR at
most 2%. Failure of any primary condition retires this exact ensemble; a
positive subset cannot substitute for it. Even a pass would justify a new
prospective software-documentation cohort, not product deployment or #349
editorial qualification.

Verify input, feature-source, lexicon, protocol, script, and source-role
digests before scoring. Reject changed identities, missing variants, source
hashes, or calibration counts. Preserve only aggregate metrics and hashes in
the tracked result; raw texts stay in local untracked files. A separately
written verifier must reject a changed result or frozen input identity.
