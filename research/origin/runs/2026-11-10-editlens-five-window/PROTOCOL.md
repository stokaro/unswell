# EditLens five-window academic-paper aggregation

Date: 2026-11-10. This is an exploratory, local origin-classification screen.
The 58 test papers and their paragraph scores were already exposed in the
[single-paragraph screen](../2026-11-10-editlens-academic-transfer/RESULT.md).
The five-window aggregation and gate below are fixed before scoring the new
human-only paper controls, but this is not independent confirmation. It does
not test editorial necessity, localization, or safe edits for issue #349.

## Pinned inputs and selection

Use `Jerry999/AIWritingEffect` revision
`6a45eddc021257b7c49de398c2fffa370c1d1bfa`, file
`dataset/t100/paragraphs.jsonl`, SHA-256
`6e0eb8058578afce4aaf81c822b70762afd9f18534eb5f1e89aa976cd051b793`.
Use the same EditLens ONNX FP32 checkpoint, tokenizer, cleaning, CPU provider,
and score definition as the earlier screen: modelkit revision
`f7cb4b06e5067ecdb66c566c7982a413f86f569f`, model SHA-256
`ddd1173f2ef517ad499965e5029fae8099a8054a2bc76d8134e5889cc4ed1b3e`.
Pin all modelkit file hashes and the exact prior test-score file. No fit,
fine-tuning, external model call, or new generated text is allowed. Keep raw
source text and model weights out of committed artifacts.

Eligible source paragraphs have `chunk_ok == true`, `kept_original == false`,
and 50–300 Unicode word tokens. For each of the 400 `human_only` papers dated
2017–2018, require at least five eligible `human` / `negative == unpaired`
paragraphs. Sort its paragraphs by
`(sha256(paper_id + NUL + pair_id + NUL +
"unswell-editlens-control-v1"), pair_id)` and take five. Sort paper IDs by
`(sha256(paper_id + NUL + "unswell-editlens-calibration-v1"), paper_id)`.
The first 300 papers calibrate the threshold; the remaining 100 are untouched
human-only test controls. The first paragraph per calibration paper matches
the earlier one-window selection. Freeze all five selected source identities
and hashes before inference.

Reuse exactly the 290 matched test human/AI pairs from the earlier frozen
selection: five pairs per each of 58 papers marked 2017–2021, one generator
and prompt variant per paper. Verify each prior no-text score against its
selected source identity and the earlier score-file SHA-256. Use no test label
to select a window, model, score direction, or threshold. After the frozen
EditLens cleaning and tokenization, a paper is eligible only if all five
selected paragraphs retain at least 50 words and fit in 512 tokens including
special tokens. Record every ineligible paper without replacing its windows.

## Aggregation, measures, and stop decision

The paper score is the arithmetic mean of its five paragraph scores. This
fixed mean avoids a maximum's multiple-window false-positive inflation and
does not assume the paragraphs are independent. Set the threshold to the
third-largest eligible calibration-paper mean, with a strict `>` decision.
Report raw and eligible paper counts, calibration false positives, separate
human-only and matched-original test false-positive rates, AI rewrite true
positive rate, pooled paper-level AUROC, and matched-paper win fraction (ties
count half). Report results by generator descriptively; generator groups are
small and papers, not paragraphs, are the independent units. The 300-paper
calibration offers only a coarse low-FPR operating point.

The directional feasibility gate requires at least 285 eligible calibration
papers, 95 eligible human-only test papers, and 55 eligible matched pairs;
both test-human false-positive rates at most 2%; paper-level AUROC at least
0.75; rewrite true-positive rate at least 20%; and matched-paper win fraction
at least 65%. A pass warrants a new software-documentation corpus with
verified human provenance, new generator/source holdouts, and a defined
whole-document window policy. It does not qualify this method for product
use. A failure is retained without changing aggregation, threshold, sample,
or gate.

Reject changed source/model/prior-score hashes, duplicate identities,
calibration/test paper overlap, missing pairs, invalid logits, unexpected
runtime provider, or partial output. The source paper-year metadata do not
prove exclusively human authorship, and the dataset has already been seen in
other Unswell feature experiments. These limits apply even if the gate passes.
