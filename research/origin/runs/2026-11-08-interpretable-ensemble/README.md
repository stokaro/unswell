# Interpretable origin ensemble: within-corpus gain, transfer failure

The [registered offline protocol](PROTOCOL.md) fitted five interpretable,
non-LLM features with a logistic classifier on source-disjoint ARB training
groups and calibrated both
the fitted score and a richness-only baseline on separate human sources. It
then tested 900 other ARB human sources and their machine variants, and reused
those **unchanged** thresholds on AIWritingEffect research-paper documents.
The protocol, scorer, feature sources, exact dataset hashes, and split digest
were committed as `894351f` before scoring. No generative model was called.
The frozen protocol's phrase "model-free" means no language-model inference;
the logistic classifier is a fitted statistical model.
The public corpora were already exposed to earlier Unswell studies; this is
development evidence, not independent qualification.

| Test and score | Eligible variants | AUROC | TPR at ARB human threshold | Human FPR |
| --- | ---: | ---: | ---: | ---: |
| ARB human-to-LLM rewrite, ensemble | 3,559/3,600 | 0.736 | 279/3,559 (7.84%) | 5/900 (0.56%) |
| ARB human-to-LLM rewrite, richness | 3,559/3,600 | 0.616 | 44/3,559 (1.24%) | 4/900 (0.44%) |
| ARB direct generation, ensemble | 3,233/3,600 | 0.864 | 780/3,233 (24.1%) | same ARB humans |
| ARB direct generation, richness | 3,233/3,600 | 0.502 | 14/3,233 (0.43%) | same ARB humans |
| AIWritingEffect rewrites, ensemble | 143/143 | 0.702 | 102/143 (71.3%) | 34/89 (38.2%) test; 120/400 (30.0%) human-only |
| AIWritingEffect rewrites, richness | 143/143 | 0.798 | 11/143 (7.69%) | 0/89 test; 1/400 (0.25%) human-only |

The ensemble's ARB gain is real **within this source-disjoint corpus split**:
it exceeds richness on both rank separation and low-FPR H2L recall. But its
registered H2L TPR floor was 10%, and it reached only 7.84%. The unchanged
threshold also mislabels many human scientific papers. The preregistered
advance screen therefore **fails**. Do not deploy this score, retune it on
these exposed papers, or claim a general origin detector.

The separately replayable [post-result feature audit](posthoc-result.json),
computed by [posthoc.py](posthoc.py), suggests a concrete distribution shift. The
first 100 tokens of ARB test humans averaged 4.57 ASCII letters per word;
AIWritingEffect test humans averaged 5.61. Mean token length received the
largest positive fitted weight (0.802 after ARB standardization), so this
shift is consistent with the high academic false-positive rate. That is an
explanation to test on new sources, not a correction to the registered run.
The ARB free-generation and LLM-to-LLM rewrite figures are secondary
contrasts; AIWritingEffect tests rewriting, not free generation.

The [result](result.json) retains all regime, source-domain, generator,
coverage, threshold, and fit details, with no raw text. `freeze.json` pins the
inputs and code; `verify.py` checks these hashes, the saved result, the
decision arithmetic, and denominator bounds. Recompute with the public pinned
datasets and research-only `duckdb==1.4.2`:

```sh
python3 -B research/origin/runs/2026-11-08-interpretable-ensemble/run.py \
  --arb /path/to/arb.parquet --academic /path/to/docs.jsonl
```

This is document-level origin research. It contributes no source-local
editorial diagnosis, meaning-preserving revision, or FULL-event recall to
[#349](https://github.com/stokaro/unswell/issues/349), whose gates remain open.
