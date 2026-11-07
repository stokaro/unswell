# ARB source-matched origin-feature transfer

This offline screen tested the **unchanged** Unswell lexical-richness and
normalized-entropy features on the public
[ARB benchmark](https://huggingface.co/datasets/giper45/ARB-Dataset).
ARB provides one human original and three machine regimes for each of four
generators per source: free generation, rewriting the human original (H2L),
and rewriting a generated text (LLM2L). The current dataset card describes
only H2L as released, but the pinned Parquet contains all four regimes and
complete source groups. The human-source provenance is a dataset claim,
not independently audited here.

The [protocol](PROTOCOL.md), [script](run.py), public dataset SHA-256, feature
source SHA-256, and source-disjoint split hash were sealed in
[freeze.json](freeze.json) before any lexical feature was scored. The input
is `giper45/ARB-Dataset` commit
`e03387f3d53efcaec9ca068c8e902e5f55611fe9`,
`data/arb.parquet`, SHA-256
`66e48240618b92ce876e2097d8df52f06661a70e31791b91a95b15688a099ab0`.
All 23,400 normalized text and source hashes, 1,800 source groups, and 13
members per group passed pre-score validation. The frozen split uses 300
calibration and 300 test sources in each of three domains. No raw corpus
text is stored in this repository.

## Registered primary result

Full-document lexical richness treats **higher** values as more AI-like.
The 1% calibration-human threshold was fixed without machine scores.
The [complete result](result.json) has SHA-256
`7f942c09a44e93f3c83f8acdedf92cf1283ec64b8efe2e16e945fa858316156f`.

| Test contrast | Eligible machine / 3,600 | AUROC | Matched-pair wins | TPR at calibrated threshold |
| --- | ---: | ---: | ---: | ---: |
| Free generation vs. human source | 3,577 | 0.459 | 44.6% | 3.9% |
| H2L rewrite vs. human original | 3,596 | 0.632 | 70.3% | 4.5% |
| LLM2L rewrite vs. human source | 3,559 | 0.527 | 52.9% | 5.7% |

The threshold flagged 0.89% of 900 calibration humans and 1.78% of 900
source-disjoint test humans. H2L AUROC by source domain was 0.636 for
OpenWebText, 0.648 for WritingPrompts, and 0.622 for XSum. By generator it
ranged from 0.585 for Llama 3.2 3B to 0.724 for Qwen2.5 7B. The registered
H2L richness feasibility rule **failed** on AUROC (required 0.65) and TPR
(required 10%), despite 99.9% pair coverage.

The separately registered first-100-prose-token comparator was weaker:
H2L richness AUROC 0.602, TPR 1.8%, 3,555/3,600 eligible pairs, and
0.78% observed test-human FPR. Entropy did not rescue the primary rule.
The complete result keeps each regime, feature, document measure, domain,
generator, calibration denominator, and abstention separate.

## Interpretation

The previous paired academic-paper screen in
[PR #446](https://github.com/stokaro/unswell/pull/446) found a strong
document-level contrast on 89 papers. ARB is a different corpus with
different generators and writing domains. The positive H2L rank signal here
is weaker, and the fixed low-FPR threshold misses most rewrites. Direct
generation even reverses the preregistered high-score direction. The
feature is therefore unsuitable as a general authorship detector on this
evidence. Neither a post-result direction flip nor a threshold adjustment
is an independent confirmation.

ARB is not software documentation, its source texts differ in length from
some machine variants, and the 900 test humans give a limited low-FPR
estimate. It cannot settle the separate #349 editorial full-event recall,
finding acceptance, source localization, or suggestion safety requirements.
The next method needs stronger source-aware features and a **new**
prospective holdout before any deployment claim.

## Replay

Download the pinned Parquet and verify its SHA-256 above. With Python 3.12
and research-only `duckdb==1.4.2`:

```sh
python3 -B research/origin/runs/2026-11-08-arb-transfer/run.py \
  --input /path/to/arb.parquet
```

The runner rejects changed input, source code, protocol, feature source, or
split identities before scoring. It writes only aggregate results.
