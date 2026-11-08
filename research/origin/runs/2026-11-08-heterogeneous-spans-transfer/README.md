# Heterogeneous mixed-authorship transfer screen

This offline screen applies the **unchanged** Unswell lexical-richness and
normalized-entropy features to the public
[Heterogeneous AI Spans dataset](https://huggingface.co/datasets/open-text-detector/heterogeneous-ai-spans).
The corpus pairs 4,400 mixed documents with 4,400 unchanged human controls and
marks each generated replacement with original-source coordinates. Its
historical human-origin evidence is documented but not independently certified.
The mixed documents are synthetic, and generated replacements need not preserve
the source's facts. This screen uses only the pinned layout-neutral view.

The [protocol](PROTOCOL.md), [runner](run.py), unchanged feature source hash,
and three input hashes were committed as `d8af694` **before scoring**. The
dataset is pinned to tag `v1.4.0`, commit
`c8a59b129c34cfd9c67f739ae049abfff26fc2d3`. The [freeze](freeze.json)
requires exact Parquet bytes: train 7,018 rows, validation 594, and test 1,188.
The pre-score audit checked all 8,800 text hashes, complete source pairs,
character-span reconstruction, original human slices, 5,283 AI spans, and
source-group isolation between splits. No corpus text is stored here.

## Registered primary result: known replacement span

Each test AI replacement was paired with its exact original human block.
The first 100 prose words on each side formed an equal-length feature window.
This is an **oracle-located** comparison: it does not find a generated span in
an unknown document. Higher lexical richness was fixed as the AI direction.
The threshold came from one eligible human block per train source and allowed
at most 1% calibration-human FPR.

| Measure | Result |
| --- | ---: |
| Eligible test AI/human span pairs | 566 / 713 (79.4%) |
| Distinct test source groups | 91 |
| Lexical-richness AUROC | 0.652 |
| Matched-pair wins, ties half-weighted | 70.0% |
| TPR at the frozen threshold | 5.65% |
| Calibration-human FPR | 0.94% (28 / 2,992) |
| Eligible test-human FPR | 0.71% (4 / 566) |

Span AUROC varied substantially by source: Beige Book 0.852 (62 eligible
pairs), Project Gutenberg 0.543 (292), JMLR papers 0.797 (156), and Standard
Ebooks 0.672 (56). The matched generator subsets also vary; small subsets
cannot establish model-specific reliability. The registered richness rule
**failed** on TPR (required 10%), even though coverage, pooled AUROC, observed
FPR, and three of four domain conditions met their thresholds. Normalized
entropy gave AUROC 0.620 and TPR 4.59%; it had no rescue rule.

## Secondary complete-document result

All 594 test mixed documents and matched human controls were eligible.
Full-document richness AUROC was **0.608**, paired wins 72.1%, and TPR 3.20%
at 1.01% observed test-human FPR. Its separate registered feasibility rule
also **failed** on AUROC and TPR. The AI spans represent only part of each
document, so the document result must remain distinct from the oracle-located
span result. The [complete aggregate result](result.json) preserves source and
generator breakdowns, both features, all denominators, thresholds, and pass
decisions. Its SHA-256 is
`819e6295f291586cd1befd088b7755c7a70cdf95212c467907e650c1a50b5d1f`.

This is a useful source-matched signal for some non-fiction snippets, but the
low-FPR miss rate rules out this unchanged feature as a general detector. The
dataset's construction may leave prompt, extraction, and genre cues; only 91
independent source groups appear in the test split. Neither a source-specific
threshold chosen now nor a post-result score-direction change would be an
independent confirmation. These origin results do not satisfy the separate
editorial event recall, finding acceptance, byte localization, and revision
safety requirements in #349.

## Replay

Download `data/all_layout_neutral/{train,validation,test}.parquet` from the
pinned commit above and verify the three hashes in `freeze.json`. With Python
3.12 and research-only `duckdb==1.4.2`:

```sh
python3 -B research/origin/runs/2026-11-08-heterogeneous-spans-transfer/run.py \
  --input-directory /path/to/pinned-parquet-directory
```

The runner rejects changed code, protocol, input files, source lineage,
offsets, spans, or split identities before scoring. It writes only aggregates.
