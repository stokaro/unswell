# LLMDet numerical component experiment

This increment of [#51](https://github.com/stokaro/unswell/issues/51) checks two
bounded components: the published proxy calculation on authored probability rows,
and Go execution of the published numeric classifier on finite vectors. It does
not provide a text detector or establish editorial usefulness.

The [paper](https://aclanthology.org/2023.findings-emnlp.139/) describes proxy
perplexity from stored n-gram probabilities. The source revision is
[`5d038354`](https://github.com/TrustedLLM/LLMDet/tree/5d038354006ca0c8e6aa0dadb75e8840accb51a8).
[ADR 0029](../../docs/adr/0029-llmdet-numerical-parity.md) places the numerical
implementation in the existing research module. It shares bounded JSON loading
with other research tools. The product engine and default gate do not consume it.

## Measured compatibility

The reference ran with Python 3.12.13, LightGBM 4.6.0, NumPy 2.2.6, and SciPy
1.15.3 on macOS ARM64. [`reference-v1.json`](reference-v1.json) records exact
source, classifier, script, and output hashes. [`parity-v1.json`](parity-v1.json)
records the Go comparison and coverage for each authored proxy input.
The Go calculation sources are fixed at
[`655526bd`](https://github.com/stokaro/unswell/commit/655526bd99cac65824894b118e100732f6ee2228).
The verifier checks their committed bytes and requires a clean, cgo-free probe
built from the same revision. The receipt records compiler and build metadata.

| Comparison | Controls | Maximum absolute error |
| --- | ---: | ---: |
| Pinned proxy function against Go | 15 | 0 |
| Native classifier margins against Go | 328 vectors, 9 classes | 0 |
| Native classifier responses against Go | 328 vectors, 9 classes | 1.11e-16 |

Absolute and relative tolerances were fixed at `1e-12` before comparison. The
classifier controls comprise eight constant vectors, 128 vectors with seed 5101,
and values below, at, and above the first 64 tree-root thresholds. They do not
cover every branch or show accuracy on prose. The actual classifier has 11 input
features, 9 output classes, 1,800 trees, and 11,594 splits. All observed splits use
numeric `<=` with no missing-value mode; leaf values already include shrinkage.
Go rejects missing/nonfinite features instead of claiming compatibility there.

## Reference semantics

The proxy starts at token index 2 and examines `max(0, token_count - 3)` positions,
even when only bigram contexts exist. It selects a four-gram context, then a
trigram, then a bigram. Once a context exists, an unlisted continuation uses its
residual probability; it does not trigger a shorter-context lookup.

Matched contexts increment the denominator even when a nonpositive likelihood
is skipped. The result is the negative sum of base-2 log probabilities divided
by `matched + 1`. This port retains that reference behavior without exponentiating
or silently correcting it. The numerical vocabulary size is an explicit argument;
it is not inferred from a tokenizer.

`reference_score` preserves the raw calculation. `value` is null when input is
too short, no context matches, or no likelihood contributes. Each case has a
machine-readable reason. `orders` counts matched bigram, trigram, and four-gram
contexts; `residual` includes unlisted continuations whose likelihood was skipped.
Context coverage and evaluated-likelihood coverage use possible positions as
their denominator. Present values are numerical features, not qualified scores.

These controls use float64 rows. The tables below carry the archive's own
16-bit values and rounding. Probability validation rejects nonfinite
values, values outside `[0, 1]`, duplicate IDs/contexts, and mass above `1 + 1e-12`.
Mass within that tolerance is not renormalized. Fully enumerated rows use their
retained values; an empty row uses uniform residual probability.

## Local reproduction

Keep external inputs and generated real-model exports outside tracked files.
Acquire the exact source and classifier identified in
[`resources-v1.json`](resources-v1.json). The source hash is checked before its
numerical function executes. The reference does not fetch or unpickle table data.
The classifier archive contains `nine_LightGBM_model.txt`; verify its recorded
digest after extraction. The experiment does not grant redistribution rights.

From a clean committed checkout, with a separately created Python 3.12 environment:

```sh
python -m pip install -r research/llmdet/reference-requirements.txt
python research/llmdet/reference.py \
  --source /absolute/research/detector.py \
  --classifier /absolute/research/nine_LightGBM_model.txt \
  --output /absolute/research/reference
(cd research/annotation && CGO_ENABLED=0 go build -o /tmp/llmdetprobe ./cmd/llmdetprobe)
python research/llmdet/verify.py \
  --probe /tmp/llmdetprobe --reference /absolute/research/reference \
  --output /absolute/research/parity.json
```

Ordinary `go test` uses frozen authored proxy controls and authored tree cases.
It also builds and executes the Go probe to check its output and exit codes.
No Python package, tokenizer, model service, GPU, or downloaded classifier is
required. The probe accepts explicit local JSON packs and JSON batches on stdin.
Invalid packs or incomplete numeric vectors return exit 2; cancellation returns
130. No policy decision or source scanning happens in this command.

## Tables and tokenizers

The second increment runs the proxy on prose. The published archive was
acquired and hashed. Its digest matches the advertised one, and
[`resources-v1.json`](resources-v1.json) records every member and file.

The archive holds one dictionary per model. Each is a pickled object array
of six Python dictionaries: the continuations and the probabilities of the
unigram, bigram, and trigram contexts. The `gpt2` dictionary retains 2,000
continuations for each of 24,935 unigram contexts, 1,000 for each of 70,000
bigram contexts, and 100 for each of 320,000 trigram contexts. That is 152
million values, far outside the bounds of the row proxy.

[`tables.py`](tables.py) converts one dictionary into a binary table in the
`unswell-llmdet-table-v1` form. The table keeps every context and every
retained value. Contexts are sorted for binary search, tokens take 16 bits,
and probabilities keep their original 16-bit form. Nothing is sampled or
rounded. `llmdet.LoadTable` reads a table into memory, about 600 MB per
model, and checks its sizes and probabilities. It measures token sequences
with the same arithmetic as the row proxy, and a test holds the two forms to
identical results on the parity controls.

Two details of the reference carry over on purpose. First, the vocabulary
size the reference passes for a model enters the residual mass only, and the
archive's tokens run past that number for OPT, so a table keeps any 16-bit
token. Second, the residual mass sums the retained probabilities one at a
time in 16-bit floats, as the reference's Python `sum` over NumPy scalars
does. The Go table rounds the same way. A 64-bit sum differs in the third
decimal on 2,000 values.

The reference tokenizes with each model's published tokenizer. Six of the
eleven models share one algorithm, the byte-level byte-pair encoding of
GPT-2: `gpt2`, `gpt2_large`, `neo`, `opt`, `opt_3b`, and `bart`. They use
two vocabularies over the same merge list, GPT-2's own and the RoBERTa-style
one of OPT and BART. `llmdet.LoadBPE` reads a vocabulary and a merge list.
`Encode` reproduces the reference pre-tokenization without a backtracking
regular expression. The tests hold it to encodings that `tiktoken` produced
from the same files: whitespace runs, contractions, digits, punctuation, and
non-Latin text. The GPT-2 files are compressed test data under OpenAI's
license. Every other tokenizer file stays outside the repository.

Five models stay untested. UniLM uses WordPiece, and LLaMA, Vicuna, and T5
use SentencePiece; none of those is ported. Bloom's vocabulary has 250,880
tokens, which does not fit the 16-bit table.

A pack manifest in the `unswell-llmdet-pack-v1` form names the tables and
tokenizer files of a run with their digests. `corpus train --llmdet-pack`
measures every prepared paragraph against each model. Each model yields a
proxy perplexity and a context coverage, with the reference's abstention
reasons. `scripts/llmdet-experiment.sh` runs the comparison of
[#177](https://github.com/stokaro/unswell/issues/177) on the corpus of the
baseline pilot, and [research/baselines](../baselines/README.md) records
the runs.

## Limits and next decision

The dataset card names `openrail` without complete component terms. The
OPT tokenizer files carry the OPT model license. Nothing from the archive is
committed, and neither is any weight file or exported real ensemble. The
classifier stage of the reference needs all eleven proxies, so it has not
run on prose.

The remaining work is the five unported tokenizers, the classifier stage,
and a product resource budget. This experiment establishes no human labels,
no held-out accuracy, no model qualification, and no product resource
budget.
