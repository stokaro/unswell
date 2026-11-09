# Frozen EditLens source-matched transfer screen

Date: 2026-11-09. This is an origin-classification feasibility screen. It
does not test editorial necessity, source-local findings, safe rewrites, or
issue #349 acceptance. The ARB corpus and its earlier lexical-feature results
were already visible; this is a new model on exposed data, not an independent
corpus confirmation. No generative model call is made.

## Inputs and preprocessing

Use the ARB Parquet at `giper45/ARB-Dataset` revision
`e03387f3d53efcaec9ca068c8e902e5f55611fe9`, file `data/arb.parquet`,
SHA-256 `66e48240618b92ce876e2097d8df52f06661a70e31791b91a95b15688a099ab0`.
Require the 1,800 human source groups and 13 rows per group as in the existing
ARB screen. Preserve its source, regime, and generator identities. Do not
publish raw corpus text.

Use EditLens RoBERTa-large ONNX FP32 from the public, ungated model kit
`CoderBak/editlens_roberta_modelkit` revision
`f7cb4b06e5067ecdb66c566c7982a413f86f569f`, file `onnx/model.onnx`,
SHA-256 `ddd1173f2ef517ad499965e5029fae8099a8054a2bc76d8134e5889cc4ed1b3e`.
The kit claims byte-identical source weights from the gated Pangram checkpoint
at revision `f93e1ace74528cfb48f337ab2fe946fb71a728cb`; this is the kit
maintainer's provenance claim, not an independent upstream comparison. The
FP32 conversion passed its 24 synthetic numerical checks, but has no reported
independent accuracy test. Pin and verify the tokenizer and configuration
files from the same kit revision. Run ONNX Runtime CPU with four intra-op
threads. No model fitting or fine-tuning is allowed.

Follow the official EditLens inference preprocessing: convert emojis to names,
discard text before `</think>`, drop a first paragraph beginning with one of
`Sure`, `Here`, `Abstract`, `Title`, `I'm happy to help`, `Certainly`, then
lowercase and collapse whitespace. The first-paragraph removal is an upstream
method choice and may alter corpus content; report how many texts it changes.
Tokenize with the pinned kit tokenizer, cap at 512 tokens including special
tokens, and record truncation. Score a text only when the cleaned version has
at least 50 Unicode word tokens. Score is the softmax expected bucket index
divided by three, as in the official inference script. It is an index, not an
authorship probability or a percentage of AI-authored words.

## Frozen sample and measures

For each domain (`owt`, `wp`, `xsum`), sort human `pair_id`s by
`(sha256(pair_id + "unswell-arb-v1"), pair_id)`. Take positions 0:100 as
human-only calibration and 300:400 as test. No calibration machine text is
scored. For each test `pair_id`, choose exactly one generator by
`int(sha256(pair_id + "unswell-editlens-v1")[:8], 16) % 4`, indexing the
ordered generator list `gemma2_9b`, `llama32_3b`, `mistral7b`, `qwen25_7b`.
Score the test human original plus its `h2l` and `free_llm` variants from that
generator. This gives 300 calibration humans, 300 test humans, 300 H2L and
300 free-generation variants before eligibility filtering. Test sources are
disjoint from calibration sources. The generator selection is fixed before
scoring.

For each machine regime separately, report eligible source-matched pairs,
directional pooled AUROC (higher score means more AI-like), matched-pair win
fraction (ties count one half), and TPR. Calibrate one high-score threshold on
eligible calibration humans: the third-largest score, with a strict `>`
decision. Report calibration and test-human FPR, and all raw and eligible
denominators. Also report AUROC and TPR by domain and generator where at
least 20 eligible pairs exist, plus lengths and truncation counts by regime.
The threshold may not be moved after inspecting test labels.

The registered feasibility gate requires at least 95% eligible matched pairs
in both machine regimes and at least 95% eligible calibration/test humans;
test-human FPR no more than 2%; H2L AUROC at least 0.75 and TPR at least 20%;
free-generation AUROC at least 0.85 and TPR at least 50%. All conditions must
hold. A pass would justify a new software-documentation holdout test, not
product integration or closure of #349. A failure is retained without a new
threshold, score direction, subset, or model variant.

Reject changed data/model/tokenizer hashes, missing or duplicate source
members, invalid source lineage, split overlap, wrong sample counts, invalid
model outputs, or insufficient calibration. Keep model files and raw data
outside version control. The screen's short, non-document domains and
512-token cap limit transfer to Unswell's technical documentation.
