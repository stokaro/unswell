# TextSight transfer: origin signal does not meet both gates

This [offline protocol](PROTOCOL.md) transferred two unchanged scores and
their ARB-calibrated thresholds from the [interpretable ensemble
run](../2026-11-08-interpretable-ensemble/README.md) to the public
[TextSight AI text detection benchmark](https://huggingface.co/datasets/textsightai/ai-text-detection-benchmark-2026).
The protocol and scorer were written before downloading and scoring these
files, then frozen locally with SHA-256 hashes. This is an exposed research
screen, not independently audited preregistration. No generative model was
called and no threshold was fitted on TextSight.

The dataset revision is
`423b1768f0827675295fd6dc3cd255ad2f411183`. Its publisher describes
155 pre-2022 Wikipedia passages and 183 2015–2021 PMC passages as human,
plus 56 September 2026 generated passages from five models. The corpus is
not source paired; 306 pre-1940 literature passages serve only as a distant
human control. We did not independently establish the publisher's provenance.

| Frozen score | AI flags | Wikipedia false flags | PMC false flags | Primary human FPR | Unpaired AUROC | Advance screen |
| --- | ---: | ---: | ---: | ---: | ---: | --- |
| Five-feature ensemble | 30/56 (53.6%) | 17/155 | 24/183 | 41/338 (12.1%) | 0.807 | Fail: false positives |
| Richness alone | 2/56 (3.6%) | 0/155 | 2/183 | 2/338 (0.59%) | 0.772 | Fail: recall |

All primary human and AI passages met the fixed 100-word eligibility rule.
Ten of 306 literature controls abstained; neither score flagged an eligible
literature passage. The ensemble flagged only 2/12 Claude Opus 5 passages,
versus 9/12 Claude Haiku 4.5 passages. These small generator strata cannot
support per-model performance claims. The full [result](result.json) preserves
the other strata and denominators.

The ensemble transfers a rank signal, but its fixed operating point violates
the registered 1% human false-positive ceiling. Richness meets that ceiling
but violates the 10% AI recall floor. Neither score advances. Do not deploy
either threshold or adjust it against this exposed corpus. A future test needs
fresh source-matched passages, prompt-held-out generation, and separately
frozen calibration and confirmation sources. This run adds no editorial
findings to [#349](https://github.com/stokaro/unswell/issues/349).

The [freeze](freeze.json) records the dataset revision, file hashes, scorer,
protocol, prior fit, and feature-source hashes. The raw dataset stays at its
publisher. To reproduce, download its four `data/*.jsonl` files at the pinned
revision into a local directory, then run:

```sh
python3 -B research/origin/runs/2026-11-08-textsight-transfer/verify.py \
  --repo . --data /path/to/textsight/data
```
