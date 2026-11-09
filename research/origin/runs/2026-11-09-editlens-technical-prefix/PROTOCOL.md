# Frozen EditLens technical-document prefix screen

Date: 2026-11-09. This screen was designed after the EditLens ARB transfer
result was visible. It reuses the unchanged, SHA-pinned ONNX FP32 checkpoint
and the 38-source/456-response long-prose-v3 technical corpus previously used
for lexical features. This is exposed, retrospective paired evidence, not an
independent detector qualification or issue #349 editorial evaluation.

## Inputs

Use all 38 original technical Markdown documents from 20 repositories and all
456 saved responses: three generator families (`haiku`, `luna`, `qwen`), two
operations (`generate`, `polish`), and two prompts (`neutral`, `plain`). Verify
the archives against their wrappers and source hashes before scoring. The
input archive SHA-256 is
`c47cced06c21e5fb8222aff06a8d220c8ec1623ac2638e59bd04d0223ebc7ec2`;
the output archive SHA-256 is
`b2eaed6ebfcb6dfa01fc31a48dea6020a0b18449f823ec481eef885e8ceeca48`.
The historical authorship of source pages is **unknown**; they are paired
source baselines, not verified human negatives. Keep raw texts out of the
result artifact.

Use EditLens RoBERTa-large ONNX FP32 from
`CoderBak/editlens_roberta_modelkit` revision
`f7cb4b06e5067ecdb66c566c7982a413f86f569f`, SHA-256
`ddd1173f2ef517ad499965e5029fae8099a8054a2bc76d8134e5889cc4ed1b3e`.
Use exactly the frozen cleaning, tokenizer, 512-token cap, 50-word minimum,
four CPU threads, and expected-bucket score from the ARB screen. Before that
cleaning, remove fenced code, inline code, URLs, and HTML tags with the
unchanged prose extractor from the November 7 richness screen. Apply the same
steps to every original and response. Record code-stripping retention,
eligibility, word count, and truncation separately. The score is a prefix
index, not a whole-document authorship probability.

## Registered analysis

Score the 38 originals once and each of the 456 responses once. Pair every
response only with its own original. Report, separately for `generate` and
`polish`: raw and eligible pairs, median response-minus-original score,
paired win fraction (ties count one half), mean delta per repository, number
of repositories with positive/negative/zero mean, and median delta by family.
Also report per-operation median response/source prose-word ratio and the
count of pairs inside the 0.75–1.33 ratio band. Count originals and responses
that exceed 512 tokens, and report the positive-repository count on pairs
where neither side was truncated as a sensitivity. Do not change the primary
all-eligible denominator based on this sensitivity.

The **generation direction gate** requires at least 95% eligible generate
pairs, positive repository-mean delta in at least 15 of 20 repositories, and
positive median delta in all three generator families. The **editing direction
gate**, evaluated independently, requires at least 95% eligible polish pairs,
positive repository-mean delta in at least 15 of 20 repositories, and positive
median delta in all three generator families. These directional gates do not
measure AUROC, true- or false-positive rates, calibrated thresholds, or
editorial value. A pass only motivates a new technical-document test with
verified human origins and repository/generator holdouts. Do not retune the
model, preprocessing, or gates after seeing scores.

Reject changed input/model/tokenizer/prose-extractor hashes, missing source
lineage, duplicate request identity, invalid model outputs, incomplete 38/456
coverage, or mismatched frozen runner. No hosted or generative model calls.
