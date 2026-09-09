# ADR 0027: Freeze research predictions before reading evaluation labels

Status: accepted for implementation; scientific and product qualification remain open.

Issues #57 and #25 need saved predictions from the existing Go baselines. A fitted
training artifact is numerical evidence. It does not authorize a normal scan to
report an editorial probability.

Add a label-free measurement entry point to the existing corpus join pipeline.
Training and prediction share extraction, target binding, feature identities, and
numeric conversion. Prediction restores the shared Go numerical models; it never
fits weights, normalizers, calibration, or thresholds. Its input contains no
annotation round. Evaluation joins a saved prediction artifact to a separate
round only after predictions have been frozen.

A run plan pins the corpus, training artifact, protocol bytes, partition, context,
response channel, and threshold. Development and final test are explicit choices;
training and calibration cannot be selected for held-out prediction. The plan
hash records that choice, but cannot prove when an operator saw labels. A final
scientific run still needs the committed manifest and access record required by
the research protocol. A digest is not an attestation.

Prepared features use the target representation after NLP analyzes its entire
eligible coherent piece. Sentence targets can retain surrounding NLP context.
Rule activations preserve the source document and require a complete matching block. Record these as distinct
contexts. They cannot establish the protocol's controlled B-versus-D comparison
until the compared methods use identical available context. No missing feature,
partial target, or out-of-range calibration score becomes a zero response.

Saved predictions contain IDs, input hashes, numerical responses, applicability,
and producer identities, with no source text or editorial labels. All selected
targets remain in the denominator. A later evaluation may exclude unresolved
labels, while reporting their reasons. Ordinary CLI/MCP behavior and the public
runtime result schema remain unchanged.

The first executable path covers the existing prepared and rule logistic fits.
Lexical models, trees, LLMDet, heavy references, paired uncertainty, scientific
qualification, and final-test acceptance retain their existing issue criteria.
Neither a simulated tutorial nor a fitted number closes those requirements.
