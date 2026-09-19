# Frozen contextual encoder comparison

Use the exact input, labels, source groups and five folds of
`2026-09-19-learned-proposals`: 157 exposed assistant-reviewed pages.
This is development evidence, not fresh confirmation, human annotation,
diagnostic credit, calibrated probability or product qualification.

Before real-data encoding or prediction, select the small frozen
`google/bert_uncased_L-2_H-128_A-2` architecture from the pinned ONNX export
and pure-Go Hugot runtime in `feasibility/`. Selection follows synthetic
resource and numerical probes, not editorial results. Redistribution provenance
for the export remains unresolved: do not ship weights. No fine-tuning.
Historical source overlap with pretraining is unknown; folds do not rule it out.

Go exports exact eligible-piece-v1 paragraph/fragment targets, verifying the
frozen source bindings and text hashes. Do not reconstruct from raw spans,
include labels or metadata, restore protected code, or concatenate across gaps.
Here target and prepared context coincide: attention is within one piece,
not cross-block or whole-page reasoning.

Pinned tokenizer, no truncation, at most 256 token IDs including special tokens.
Long inputs retain unavailable reason sequence_limit, never a zero vector.
Empty/corrupt inputs fail. Attention-mask mean pooling and L2 normalization
produce 128 dimensions. Pure-Go CPU inference, CGO disabled, batches of four,
sequence buckets 32/64/128/256. Cache only identical target hashes under the
same encoder manifest. Retain every labeled, unlabeled and unavailable row.

Compare E (128 encoder columns) and ES (E plus the existing 36 numerical and
missingness columns). Use the unchanged PR340 Go logistic optimizer: L2=0.01,
training-only means/scales, unpenalized intercept, gradient tolerance 1e-8,
500 iterations and three billion accounted operations. No hyperparameter search.
For evaluation fold i, threshold on (i+1)%5 and train on the other three folds.
Use only explicitly labeled available rows for fitting.

Threshold constraints stay explicit-control FPR <=1% and labeled precision >=85%,
maximizing true selections and keeping ties together. Apply constraints on
available development rows; report covered and full-stream denominators.
No qualifying nonempty point means abstention. Unavailable outputs have null
scores, are never selected, and remain in full-stream recall denominators.

Compare both arms with retained L/S/W/SW and SW128/1024/8192. Report fold, cohort,
length-band, coverage, event retrieval, operation and resource measurements.
Preserve prior rule credit separately. Do not declare the best exposed result
confirmed. Repeat Go fitting and require identical outputs. Independently
compare embeddings with ONNX Runtime on the first 32 available unique hashes
in ascending SHA256 order, tolerance 1e-4, before interpreting editorial scores.

A selection has no automatic explanation. Product adoption requires reviewed
source-specific diagnoses and new complete-page confirmation frozen before
outputs. Gates stay unchanged. No product model is installed by this study.
