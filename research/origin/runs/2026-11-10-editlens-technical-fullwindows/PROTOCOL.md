# EditLens full-window technical-document screen

Date: November 10, 2026. This is an exposed, exploratory AI-origin direction
screen on the existing `long-prose-v3` corpus. It does not measure issue #349
editorial quality or qualify an origin detector for product use.

## Frozen inputs

Use the same 38 historical technical documents and 456 saved responses as the
[technical-prefix screen](../2026-11-09-editlens-technical-prefix/PROTOCOL.md).
The document identity SHA-256 is
`de1ae6afa3304c34bad98ed41d7588bef2e094b5b69fb732cad0e57dd7797d3b`.
Retain the original source-response pairing, three generator families, two
operations, and two prompts. Do not change the saved outputs or source labels.
The source documents predate 2021, but their authorship is unknown.

Use the public EditLens RoBERTa-large ONNX FP32 checkpoint at modelkit revision
`f7cb4b06e5067ecdb66c566c7982a413f86f569f`, SHA-256
`ddd1173f2ef517ad499965e5029fae8099a8054a2bc76d8134e5889cc4ed1b3e`.
Use the existing pinned code/URL stripping, EditLens cleaning, tokenizer, score
index, CPU provider, four CPU threads, and batch size eight. No fine-tuning,
new text generation, hosted model call, or score calibration occurs.

## Window and document score

Use the complete cleaned prose, divided into disjoint 100-word windows by
`word_windows` from the published [eligibility audit](../2026-11-10-editlens-technical-windows/README.md).
Fold a tail shorter than 50 words into the preceding window. The windows must
concatenate exactly to the cleaned text and each must fit in 512 model tokens.
Reject a document with fewer than 50 cleaned words or any over-cap window;
do not silently truncate or replace it. A document score is the arithmetic
mean of all its window score indexes. A response delta is its mean minus its
own source's mean. Do not select windows by model score or label.

Freeze the identity and ordered token-length hash before inference. Record
every window score, document score, and pair delta without raw text or token
IDs in the committed results. The first-window prefix result is prior exposed
evidence; it must not tune this aggregation.

## Registered measures and decision

Report availability, window count, median delta, paired win fraction (ties
count one half), repository-mean sign counts, and median delta by generator
family separately for `generate` and `polish`. Report the same direction
statistics for pairs whose cleaned word ratio is 0.75–1.33 as a sensitivity;
do not replace the primary denominator. Count source documents, not responses,
as independent origins. Report changes in window count and word ratio so a
positive delta is not presented as an unconfounded authorship effect.

The exploratory direction gate for **each** operation requires at least
217/228 eligible source-response pairs, positive mean delta in at least 17/20
repositories, a paired win fraction of at least 0.70, and positive median
delta in all three generator families. A pass only warrants a new source-
disjoint technical-document benchmark with checked human provenance, a
separate human calibration/test split, matched rewrites by held-out generators,
and a low-false-positive operating point. A failure retires this exact full-
window mean configuration on this exposed corpus. Neither outcome supplies
true-positive or false-positive rates here.

Reject changed archive, document identity, tokenizer, model, code, ordered
window inventory, provider, output shape, or incomplete scoring. Preserve
errors and the original denominators. Do not retune this protocol after seeing
scores.
