# Numerical proxy controls

`proxy-controls.json` contains fifteen authored token/row inputs and outputs from
the pinned LLMDet `perplexity` function. These are numerical regression cases,
with no prose, authorship labels, editorial judgments, or downloaded table rows.

[`reference.py`](../../../llmdet/reference.py) generates the fixture using NumPy
float64 probabilities and int64 token IDs. It extracts only the hash-checked
function from revision `5d038354006ca0c8e6aa0dadb75e8840accb51a8`; model setup,
tokenization, downloads, and other source imports do not execute. The fixture
includes context precedence, residual probability, zero likelihoods, the pinned
denominator, partial coverage, and insufficient input. Its SHA-256 is recorded in
the [reference receipt](../../../llmdet/reference-v1.json).

Blackbox tests compare the Go result with these outputs at absolute tolerance
`1e-12`, without Python or external model data. The separate authored tree tests
exercise numeric boundaries, accumulation order, snapshot ownership, cancellation,
overflow, and malformed graphs. They do not substitute for the recorded comparison
with the published classifier, which requires locally acquired external weights.
