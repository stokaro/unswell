# MELD technical-document feasibility screen

Status: pre-inference freeze, November 11, 2026. No MELD score was observed when
this protocol, input manifest, scripts, and `freeze.json` were committed. This
is an exposed-data origin screen, separate from #349 editorial qualification.

## Question and fixed inputs

Can the pinned MELD detector retain a low observed alert rate on historical
technical documents while detecting complete technical generations and AI
polish? The historical-control frame is the existing 761 selected documents
from 2021 or earlier: 404 calibration and 357 test, selected by repository
before this experiment. The matched technical frame has 38 sources and 456
saved responses: 228 generations and 228 polish responses, across 20
repositories. These controls and responses were exposed to prior method work.
Historical dates do not establish independent human authorship.

`prepare.py` verifies the 761 exact source blobs against the existing document
manifest, verifies both saved response archives, removes fenced code, inline
code, URLs, and HTML tags with the pinned `richness.prose` helper, and writes
the exact local packet. This procedure applies to all three document groups.
A row is eligible with at least 100 extracted words. Preserve ineligible rows
and coverage denominators. `input-manifest.json` records identities, raw and
prose digests, eligibility, and word counts without publishing source text.
The exact local packet SHA-256 is in `freeze.json`. All 1,255 packet rows are
fixed before inference, including eight ineligible historical controls.

Use the public `odeveloper/meld` mirror at revision
`71d48f8e8be43885f86553565537ed015bff480a`, which describes itself as an
unmodified copy of original revision
`8990324abd92e1fa17072f6887ea1e5c1cef5abc`. Pin every model-file hash,
including the 4.1 GB weights. Run the upstream `Scorer` in canonical mode on
Apple MPS with FP32, batch size one, its default nonoverlapping 2,046-content-
token windows, 16,384-token cap, and top-quarter pooled score. Record
truncation. Do not use the upstream sigmoid or its academic threshold as a
calibrated probability or a technical-document decision.

## Frozen decision rule

Score only the eligible historical calibration documents first. Sort their
scores ascending and set the strict `>` threshold to the value leaving at
most `floor(0.01 * n)` calibration scores above it. Persist the complete
calibration scores and threshold before any test or response inference. Then
score eligible historical test documents, followed by matched sources and
responses, without changing the model, packet, or threshold.

The feasibility screen passes only if all three conditions hold:

1. At most 1% of eligible historical test documents are flagged.
2. At least 50% of eligible complete generations are flagged.
3. At least 20% of eligible complete AI-polished responses are flagged.

Report exact numerator and denominator, coverage against all 357 test
controls and all 228 responses in each operation, capped inputs, repository
splits, and paired response-versus-source direction. `verify.py` independently
recomputes the threshold, identities, counts, and gates from the saved scores.
The test documents and all matched responses become exposed development data
after this run. A failed screen retires this checkpoint/configuration for this
technical-document setting. A pass warrants a newly selected, source-disjoint
confirmation frame with checked human provenance and held-out generators; it
does not qualify a deployed origin detector or any #349 editorial event.

## Failure and resource policy

One local inference run, no retries or partial resume. Any missing source,
hash mismatch, environment mismatch, model-load error, nonfinite score, or
inference error stops the run. Preserve partial scores and `error.json`; never
count a partial packet as a negative or positive result. The runner stops at
four hours elapsed including model load. Temporary local storage is capped at
12 GB for the model, environment, input, and scores; remove the task-specific
model and environment after recording the verified result. No hosted model
calls are used.
