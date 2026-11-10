# EditLens academic-prose transfer screen

Date: 2026-11-10. This is a local, non-generative origin-classification
feasibility screen. It does not test editorial necessity, source-local
findings, safe edits, or the acceptance gates in issue #349. The
AIWritingEffect corpus was previously exposed to a different Unswell lexical
feature screen. This study is a new model on exposed source material, not
independent product confirmation.

## Pinned inputs

Use `Jerry999/AIWritingEffect` revision
`6a45eddc021257b7c49de398c2fffa370c1d1bfa`, file
`dataset/t100/paragraphs.jsonl`, SHA-256
`6e0eb8058578afce4aaf81c822b70762afd9f18534eb5f1e89aa976cd051b793`.
The dataset pairs conference-paper paragraphs with AI rewrites and assigns
papers to `train`, `val`, `test`, and `human_only` splits. Select only the
`test` papers marked 2017–2021 and `human_only` controls marked 2017–2018;
this excludes 2022–2023 test papers whose pre-ChatGPT PDF dates would need
individual verification. Venue/year metadata and the authors' PDF-date claim
still do not prove exclusively human authorship. Keep text out of committed
artifacts. Preserve paper, pair, generator, prompt, and source hashes.

Use the same public EditLens ONNX FP32 checkpoint and tokenizer frozen in the
[ARB transfer](../2026-11-09-editlens-transfer/PROTOCOL.md): modelkit revision
`f7cb4b06e5067ecdb66c566c7982a413f86f569f`, ONNX SHA-256
`ddd1173f2ef517ad499965e5029fae8099a8054a2bc76d8134e5889cc4ed1b3e`.
Reuse that run's unchanged preprocessing, scoring, CPU provider, and output
interpretation. A score is an index, not an authorship probability. No fit,
fine-tuning, prompt change, or score calibration on machine examples is allowed.
The local environment is Python 3.12 with `onnxruntime==1.30.0`,
`tokenizers==0.23.2`, `numpy==2.5.3`, `emoji==2.15.0`, and
`duckdb==1.4.1` (the last package is imported by the frozen ARB helper).

## Selection before model scores

Eligible source rows have 50–300 Unicode word tokens before preprocessing,
`chunk_ok == true`, and `kept_original == false`. For each `test` paper with
`year <= 2021`,
require matched `human`/`ai` rows with the same `pair_id`, the human row's
`negative == paired`, and the AI row's generator and prompt present. Some
papers have several AI generator/prompt variants. For each paper, consider
variants with at least five eligible matched pairs and select one by
`(sha256(paper_id + NUL + generator + NUL + prompt + NUL +
"unswell-editlens-variant-v1"), generator, prompt)`. Within that variant,
sort eligible pair IDs by `(sha256(paper_id + NUL + pair_id + NUL +
"unswell-editlens-academic-v1"), pair_id)` and take the first five. Require
exactly 58 test papers and 290 selected pairs. Never choose a variant by its
score.

For each `human_only` paper, select its eligible paragraph with the smallest
`(sha256(paper_id + NUL + pair_id + NUL +
"unswell-editlens-control-v1"), pair_id)`. Sort papers by
`(sha256(paper_id + NUL + "unswell-editlens-calibration-v1"), paper_id)` and
take the first 300. Require 400 available control papers and no paper shared
with `test`. This produces 300 human-only calibration rows and 290 matched
test human/AI pairs. Freeze the selected identities and source hashes before
scoring. The test rows remain unseen by the threshold selection.

Clean and tokenize with the frozen EditLens method. A row is eligible for the
registered metrics only if its cleaned text still has at least 50 Unicode
word tokens and its full tokenizer output has at most 512 tokens, including
special tokens. Record all excluded and truncated rows. Do not replace
ineligible rows after scoring. The model may score truncated prefixes to
expose them, but those scores cannot enter the registered whole-paragraph
metrics.

## Measures and gate

Set the high-score threshold to the third-largest eligible calibration-human
score, with a strict `>` decision. Report eligible calibration and test-human
counts, test-human false positives, paired AUROC, matched-pair win fraction
(ties count half), AI true positives, and coverage. Report paper-level
clustering and descriptive results by generator and prompt; multiple
paragraphs from one paper are not independent sources. The 300 calibration
papers support a coarse 1% operating point, not precise individual authorship
claims.

The registered feasibility screen passes only if at least 285 calibration
controls, 275 test humans, and 275 matched pairs remain eligible; test-human
false-positive rate is at most 2%; paired AUROC is at least 0.75; AI true-
positive rate at the frozen threshold is at least 20%; and the matched-pair
win fraction is at least 65%. A pass justifies a new software-documentation
holdout with independently checked human provenance and a whole-document
window policy. A failure is retained without changing the threshold, sample,
preprocessing, model, or gate.

Reject changed hashes, duplicate identities within a generator/prompt variant,
split leakage, missing pairs, model provider drift,
invalid logits, or an incomplete run. Record raw and eligible denominators.
Do not infer any improvement for issue #349 from this origin score. No
external model service or new generative-model request is used.
