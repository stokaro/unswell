# Stacked span filtering: current classifier cannot activate the mask

The [frozen calibration screen](PROTOCOL.md) tested whether the `0.01`
machine-confidence mask from [Wu et al. (2026)](https://arxiv.org/abs/2605.23190)
could operate on Unswell's existing five-feature, non-LLM origin classifier.
The exact protocol, source hashes, and script were committed as `d843723`
before any window score was computed. The author's
[MGTD project](https://github.com/Daftstone/MGTD) uses stronger neural
detectors; this is an applicability test of our existing weak classifier,
not a reproduction or disproof of the paper.

| Frozen calibration cohort | Eligible documents | Documents filtered | Confident windows / all windows |
| --- | ---: | ---: | ---: |
| ARB AI rewrites | 1,183/1,200 | 1 (0.085%) | 1/2,150 |
| ARB human controls | 300/300 | 0 | 1/724 |
| AIWritingEffect validation AI rewrites | 131/131 | 0 | 0/524 |
| AIWritingEffect validation humans | 83/83 | 0 | 0/332 |

The predeclared continuation gate required masking in at least 5% of
eligible AI rewrites **in each corpus**. Both fail. The ARB human window below
the confidence threshold belongs to a single-window document, so the
retain-one-window constraint correctly prevents its removal. No threshold
was changed, and test or human-only sources were not scored. A second-pass
detector, hard-EM retraining, and a low-FPR origin evaluation were therefore
not run. `result.json` SHA-256 is
`0546f2a7137fff50158c6ea6e005a653e7ba526a2b2a257015dd72354fdabadf`.

The result identifies a concrete prerequisite: this paper-style inference
mask needs a base detector with well-calibrated, confidently human-like
window scores. Our current classifier rarely produces them. Future work may
test a stronger non-generative base detector on new, source-disjoint corpora
with a separately frozen protocol; relaxing `0.01` on these exposed samples
would be a different method requiring fresh confirmation. This experiment
adds no shipped origin detection and no #349 editorial FULL-event credit.

Recompute with the pinned local corpora and research-only `duckdb==1.4.2`:

```sh
python3 -B -m unittest discover \
  -s research/origin/runs/2026-11-09-stacked-applicability \
  -p 'test_*.py'
python3 -B research/origin/runs/2026-11-09-stacked-applicability/screen.py
```

The runner rejects changed input, model, code, and protocol hashes and
retains only aggregate counts. It does not collect or publish raw source text.
