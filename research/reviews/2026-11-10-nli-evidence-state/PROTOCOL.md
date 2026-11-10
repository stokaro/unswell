# Offline evidence-state screen for #349

This is a small, exposed synthetic feasibility screen. It asks whether a local
natural-language-inference (NLI) classifier can distinguish support,
contradiction, and absent evidence around an exact source target. It is not an
AI-authorship detector, an editorial candidate generator, an edit-safety
checker, or a product rule. No hosted model call is allowed.

The 16 cases are the already frozen support-flip packet in
`research/reviews/2026-11-08-evidence-flip-pilot/packet`. The source and
target bytes are verified against that packet's SHA-256 manifest. Each NLI
premise is the complete synthetic source with the exact target byte interval
removed; the hypothesis is the target sentence. Removing the target prevents
it from providing evidence for itself. No Markdown parser or normalization is
used. The packet includes six support cases, six no-evidence cases, and four
contradictions; the original editorial labels remain unchanged.

Use `cross-encoder/nli-MiniLM2-L6-H768` at revision
`b95119ce93d3e065de6214e38cd4a97b0f2f2c6d`, with its published ARM64
quantized ONNX model. Require exact SHA-256 for model, tokenizer, and config
as recorded in `freeze.json`. Run once locally with ONNX Runtime 1.30.0,
tokenizers 0.23.2, and NumPy 2.5.3 on CPU. Use the model's argmax label in
config order: contradiction, entailment, neutral. Reject any input over 512
token IDs instead of truncating it. No fit, threshold search, retry, or
post-result label change is allowed. The committed freeze must precede first
inference. Stop on the first input, asset, or runtime error.

Treat entailment as `support`, neutral as `no_evidence`, and contradiction as
`contradiction`. The predeclared editorial admission mapping is:

| Diagnosis family | Admit when evidence state is |
| --- | --- |
| `table-announcement`, `repeated-release-status` | support |
| `unsupported-comparison`, `unsupported-reader-need` | no evidence |

All other states withhold the supplied diagnosis, including contradictions
that may indicate a different defect. Report 16-way relation accuracy, the
three-class confusion matrix, editorial-decision agreement, each of the four
original support-flip pairs, and contradiction false admissions. The
directional feasibility gate requires at least 14/16 correct relations, at
least 14/16 correct editorial decisions, all four support-flip pairs correct,
and zero contradiction false admissions. A pass would justify a separately
frozen real-source test with complete technical pages and controls. A failure
retires this exact model, premise/hypothesis construction, and admission map
on these cases; do not tune them on the exposed outputs.

This synthetic screen cannot change the 36-page/123-event development counts,
the 80% FULL / 85% accepted-delivery requirements, source-local diagnosis
coverage, suggestion safety, or the separate prospective confirmation role.
