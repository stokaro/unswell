# EditLens full-window technical-document screen: directional pass

The frozen [protocol](PROTOCOL.md), [runner](run.py), [independent verifier](verify.py),
and [input seal](freeze.json) were committed as `ed4035a` before the first
inference. One local CPU run scored every disjoint 100-word window in the 38
historical technical sources and 456 saved generation/polish responses. The
6,484 windows covered the complete **cleaned** prose; none exceeded 512 model
tokens. No new text was generated, no hosted model was called, and no runtime
detector changed.

| Frozen direction measure | Generate | Polish |
| --- | ---: | ---: |
| Eligible pairs | 228/228 | 228/228 |
| Median response-minus-source document score | +0.3095 | +0.0276 |
| Responses scoring strictly above their own source | 221/228 | 186/228 |
| Equal source-response scores | 0/228 | 6/228 |
| Paired win fraction, ties worth one half | 96.9% | 82.9% |
| Repositories with positive mean delta | 20/20 | 19/20 |
| Positive median delta by all three model families | Yes | Yes |
| Registered direction gate | **Passed** | **Passed** |

The complete-document score is the arithmetic mean of all window score
indexes, not an authorship probability. The narrower 0.75–1.33 cleaned-word
ratio sensitivity retained a positive median delta and paired direction in
100/100 generation pairs and 162/195 polish pairs, with six more polish ties.
Generated responses were
shorter than their sources (median word ratio 0.735); polish responses were
similar in length (0.984). Length and formatting remain possible confounders.

The source pages' human authorship is unknown, and these 38 sources and 456
responses were already exposed in earlier studies. There is **no** independent
human calibration, false-positive rate, held-out generator, or new-source
true-positive rate here. The result advances the method from a prefix-only
technical screen to complete cleaned-prose scoring; it does not qualify an
AI-origin detector for product use. It provides no editorial diagnosis,
source-local edit, or FULL-event credit for issue #349.

The [decision seal](decision.json) records all four output-file SHA-256 hashes
and the single completed attempt. [Window scores](window_scores.json),
[document scores](documents.json), and [paired deltas](pairs.json) contain no
raw text or token IDs. `python3 -B verify.py` independently recalculates all
document means, pair identities, category aggregates, and gate decisions;
it passed. Nine tests in `test_run.py` and `test_verify.py` passed, including
deliberate window-score, source-identity, pair-coverage, and gate corruption.

To replay inference, use Python 3.12 with `onnxruntime==1.30.0`,
`tokenizers==0.23.2`, `numpy==2.5.3`, `emoji==2.15.0`, and `duckdb==1.4.1`.
Obtain the public modelkit at the revision and SHA-256 in `freeze.json`, then
run from the repository root:

```sh
python3 research/origin/runs/2026-11-10-editlens-technical-fullwindows/run.py \
  --model-dir /path/to/editlens-modelkit \
  --output-dir /tmp/unswell-editlens-fullwindows-replay
python3 research/origin/runs/2026-11-10-editlens-technical-fullwindows/verify.py \
  --output-dir /tmp/unswell-editlens-fullwindows-replay
```

The next separate test should use independently sourced technical documents
for a low-false-positive screen, with source and generator holdouts and checked
provenance. The frozen direction pass cannot be used to select or alter those
controls after looking at model scores.
