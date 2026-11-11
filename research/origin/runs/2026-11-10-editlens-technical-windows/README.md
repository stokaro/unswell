# EditLens technical-document window availability

This is an **exploratory, offline input audit**, not a model or detector result.
It uses the already published long-prose-v3 corpus: 38 pre-2021 technical
source documents from 20 repositories and 456 saved model responses. The
source and response archives, helper scripts, and tokenizer match the SHA-256
identities in [`result.json`](result.json). No model weight, inference,
generative call, editorial judgment, threshold, or product rule was used.

The prior [technical-prefix screen](../2026-11-09-editlens-technical-prefix/RESULT.md)
truncated every source at 512 tokens. The later [academic-paper screen](../2026-11-10-editlens-five-window/RESULT.md)
used five complete paragraphs per paper. This audit checks whether that
paragraph policy transfers to technical Markdown and whether disjoint word
windows can cover the cleaned prose within the same token limit.

| Window policy | Source documents eligible | Generate pairs eligible | Polish pairs eligible | Total windows |
| --- | ---: | ---: | ---: | ---: |
| At least five isolated 50–300-word paragraphs | 18/38 | 71/228 | 86/228 | Varies |
| At least three isolated 50–300-word paragraphs | 30/38 | 123/228 | 136/228 | Varies |
| All cleaned words, disjoint 75-word windows | 38/38 | 228/228 | 228/228 | 8,567 |
| All cleaned words, disjoint 100-word windows | **38/38** | **228/228** | **228/228** | **6,484** |
| All cleaned words, disjoint 150-word windows | 36/38 | 216/228 | 216/228 | 4,412 |

A final fragment shorter than 50 words joins the preceding window. At 100
words, no window exceeded 512 model tokens; the maximum was 430. At 150 words,
two source windows and five response windows exceeded the cap; the largest had
645 tokens. These counts concern **all characters of the cleaned text**, not complete
original bytes: fenced code, inline code, URLs, HTML tags, and any model header
removed by the earlier EditLens preprocessing are outside the measurement.
Window boundaries can cut sentences, but no cleaned character is omitted or
repeated. A product implementation would need an explicit source map and
protected-region policy.

The 100-word policy removes the prefix-only availability problem on this
fixed corpus, making it a candidate for a separately frozen full-document
score experiment. It does **not** establish accuracy. The number of windows
varies with document length, and the saved outputs may differ from originals
in length and formatting. Original source authorship remains unknown even
though snapshots predate 2021; the corpus is exposed, and its 38 originals
cannot provide precise low-FPR calibration. A qualifying next study needs
source-disjoint technical documents with independently checked human
provenance, separate calibration and test controls, and matched AI rewrites.
Its model, window aggregation, threshold, and acceptance gate must be fixed
before inference. Issue #349's source-local editorial requirements remain
unmet.

## Replay

Use the modelkit tokenizer at revision
`f7cb4b06e5067ecdb66c566c7982a413f86f569f`, SHA-256
`2bb1a22cfbe25b8e5a232b7fc4d7fc5073923b45724a5f813b00811bb6620f66`.
The model weight is unnecessary. With Python 3.12 and the dependencies from
the earlier EditLens runner:

```sh
python3 research/origin/runs/2026-11-10-editlens-technical-windows/audit.py \
  --tokenizer /path/to/tokenizer.json --output /tmp/technical-window-result.json
cmp research/origin/runs/2026-11-10-editlens-technical-windows/result.json \
  /tmp/technical-window-result.json
python3 research/origin/runs/2026-11-10-editlens-technical-windows/verify.py \
  --tokenizer /path/to/tokenizer.json
python3 -m unittest discover \
  -s research/origin/runs/2026-11-10-editlens-technical-windows -p 'test_*.py'
```

[`audit.py`](audit.py) has SHA-256
`198b5c06cd79700fc6bec6a177146538116391b0fdf02ded096d375558cf89ae`;
`result.json` has SHA-256
`52a33f84ce4953a13b5c28078ce0e7a87e373ab7bb73426a299ff9b25f55bf11`.
The independent verifier checks input hashes, response identities, complete
pairing, internal accounting, and absence of raw text or token IDs. It does
not independently reconstruct window boundaries; the replay above does.
