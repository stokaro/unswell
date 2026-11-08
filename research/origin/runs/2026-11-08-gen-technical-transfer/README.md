# GEN threshold transfer to technical prose

This offline study applies the fixed GEN first-100-prose-token thresholds to two
previously exposed corpora. The [protocol](PROTOCOL.md), scorer, and input hashes
were frozen in commit `5ccf936223afd62ad714b801489719fd1cf953e7` before
scoring. The [freeze](freeze.json), [result](result.json), and
[decision](decision.json) preserve the exact inputs and outcome. No model calls
were made for this study.

## Result

| Corpus and operation | Eligible | Richness flags | Entropy flags | Paired richness wins |
| --- | ---: | ---: | ---: | ---: |
| Technical free generation | 228/228 | 6 (2.6%) | 11 (4.8%) | 70.8% |
| Technical polish | 228/228 | 0 | 1 (0.4%) | 50.7% |
| Academic test AI rewrites | 143/143 | 21 (14.7%) | 26 (18.2%) | 93.7% |

The technical generation richness delta is positive in 17 of 20 repositories.
Yet all three technical generator families fall below the prespecified 5% flag
floor: haiku 2/76, luna 1/76, and qwen 3/76. The primary threshold also misses
the 10% overall flag floor. The **advance screen fails**. A higher paired score
is a useful research signal, but this unchanged GEN threshold does not provide
enough technical generation recall for direct reuse.

The 38 original technical documents produced zero richness flags. Their
authorship is unknown, so this is a source-control flag rate, not a measured
human false-positive rate. The academic dataset's `human_only` partition produced
5/400 richness flags (1.25%); the test and validation human partitions produced
1/89 and 0/83, respectively. Academic human provenance is asserted by the
dataset and was not independently verified. The AI rows are rewrites of papers,
not free generations, and their 21/143 flag rate cannot rescue the primary
technical-generation gate. One rewrite generator contributed 11 of those 21
flags, so the aggregate also hides substantial generator variation.

This is a transport stress test on sources used in earlier Unswell research.
It is not prospective confirmation or a qualified origin detector. It gives no
credit toward the separate editorial diagnostics gate in issue #349.

## Reproduction

Fetch `Jerry999/AIWritingEffect` commit
`6a45eddc021257b7c49de398c2fffa370c1d1bfa`, file
`dataset/t100/docs.jsonl`, and verify its SHA-256 against `freeze.json`.
Then run:

```sh
python3 -B research/origin/runs/2026-11-08-gen-technical-transfer/run.py \
  --dataset /path/to/docs.jsonl --output /tmp/gen-technical-transfer-result.json
cmp /tmp/gen-technical-transfer-result.json \
  research/origin/runs/2026-11-08-gen-technical-transfer/result.json
make check-gen-technical-transfer
```

The repository archives and all other source hashes are checked by the scorer.
The offline verifier checks frozen repository inputs and the saved result without
requiring the external dataset file.
