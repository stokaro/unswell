# ADR 0030: Train lexical baselines on shared prepared targets

Status: accepted for numerical research; corpus qualification remains open.

## Decision

Add the lexical variant in #50 and the A–G protocol to the existing Go training,
prediction, and evaluation path. Reuse `corpus.Build`'s source verification,
extraction, and `nlp.PrepareUnits` calls. `corpus.Prepare` retains those units and
returns them only after all frozen candidates have been reproduced. This exposes
the original token data without a second tokenizer or a new analysis engine.

`feature.CountLexical` counts word orders 1–3 and character orders 1–6 on one
prepared target. Each family can be disabled. Word sequences retain normalized
prose tokens, including negations, modality, numbers, and identifiers, and stop at
punctuation or sentence boundaries. Character sequences use Unicode code points,
`document.Normalize`, and collapsed whitespace within each sentence. Protected
boundaries remain boundaries. Neither counter applies the repetition subsystem's
content-word filter. Keys explicitly contain extracted text.

`training.RunLexical` first uses the existing row selector to identify resolved,
permitted training and calibration targets. It learns vocabulary only from the
selected training targets, ranking keys by the number of those targets containing
them, with a lexical tie break. This frequency counts targets, not independent
documents. The frozen corpus groups remain the unit of split independence.
Calibration, development, final-test, unannotated, and unresolved targets do not
contribute vocabulary entries or selection frequencies. Verification still checks
all frozen source bytes and target mappings. Each selected target is counted once;
vocabulary fitting and vector construction reuse those counts. Prediction counts
only the requested partition and unit kind after full source verification.

The representation is `log1p(count)` with no IDF or length normalization. The
existing logistic trainer fits means, scales, weights, and intercept on training
rows. Isotonic calibration remains separate. Inference ignores unknown keys and
retains observed zeros for known absent terms. An out-of-vocabulary input can
therefore have an all-zero raw vector; that is a numerical baseline response,
not evidence of acceptable quality or human authorship.

Column IDs hash explicit keys to bind them, not to merge different terms into
hashing buckets. The vocabulary preserves keys and target frequencies. The
initial implementation uses the existing dense optimizer's 128-column budget;
this is a bounded experimental baseline, not evidence that 128 terms suffice.
Larger sparse vocabularies require a separately tested numerical implementation
and comparison. Keep this limitation visible in #50's eventual decision report.

## Artifacts

All training variants use `unswell-editorial-training-v2`. Lexical fits require a
hashed vocabulary; other variants reject lexical metadata. Previous alpha training
formats are rejected explicitly. The feature contract is
`unswell-lexical-counts-v1`; source, extraction policy, actual NLP, prepared unit,
target kind, dictionary, columns, and preprocessing remain explicit identities. The learned
vocabulary hash is separate from the existing engine policy/vocabulary hashes.
`nlp.PreparationHash` shares the existing preparation formula. Cross-model
comparison permits known feature representations to differ while retaining
source-policy, preparation, target, rubric, profile, and NLP equality.
Saved predictions embed the fitted model and continue through the same evaluation
and comparison commands after the external model file has been removed.

The vocabulary is an explicit research artifact containing source-derived terms.
It is not a normal product report. Review source permissions before distributing
it. No vocabulary, model, source snippet, or new score is added to normal scans,
MCP, or reporters. Numerical fits retain `unavailable_unqualified_model`.

Preparation additionally caps retained NLP accounting at 256 MiB. Counting bounds
input size, token visits, occurrences, unique keys, and retained key bytes.
Vocabulary fitting caps candidates at 100,000 keys and 16 MiB of key bytes;
selected target histograms have a 64 MiB accounting limit, and serialized models
retain the existing 16 MiB limit. Exceeding a limit fails
without a partial model. These accounting limits are not measured RSS claims.

## Verification and remaining work

Blackbox tests cover counts, Unicode, protected boundaries, vocabulary isolation,
source reproduction, permissions, corrupted models, deterministic restoration,
and the compiled train → predict → evaluate → compare commands. Test annotations
are explicitly simulated; they do not satisfy the independent human corpus.

Issue #50 remains open for real grouped data, model selection, resource measurements,
larger-vocabulary comparisons if needed, the stylometric/tree variants, ablation,
and a measured adoption decision. Successful numerical fitting does not close
corpus qualification, applicability, calibration quality, or product acceptance.
