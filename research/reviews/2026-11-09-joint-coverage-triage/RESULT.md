# #349 joint coverage and delivery triage: stopped pilot

**Status: terminal negative screen, not product qualification.** The frozen
packet is commit `2c1b4cae6173` with freeze-manifest SHA-256
`041d42c594aff96e882d023419dfa7d17618c5422e52e5d01cf64bfd6cc643e2`.
The first of at most four sequential calls used GPT-6-Astra medium through the
existing Codex subscription. The CLI returned one schema-valid response in
63.629 seconds, but source binding rejected a new finding that overlapped a
protected source span. The runner stopped at the first error as required. Pages
6, 8, and 10 were **not called**. The response is preserved as raw evidence;
there is no valid four-page score, and the eight frozen positive events remain
in the denominator for any later complete evaluation.

## First-response audit

The page 2 response reviewed all nine baseline findings and proposed seven
new diagnoses. Five baseline findings were kept and four withheld. The prior
full-source audit had accepted all nine baseline findings. Therefore, even if
the source-binding error were corrected, this exact candidate would fail its
predeclared requirement to retain all previously accepted findings. Do not
reinterpret the previous acceptance labels to improve this screen.

New diagnosis 4 quoted the repetition in the `--network host` explanation on
source line 164, byte span `[13236, 13332)`. That prose sits inside the frozen
excluded span `[13215, 13912)`, so `run.py bind` rejected the entire page. The
other six new quotes did not overlap an excluded span. New diagnosis 7 targets
the source region of frozen positive event `transfer-v2-008` and describes the
repeated project objective. This is a useful lead for future design, but the
invalid page is not graded FULL. It does not offset the four accepted findings
that the same response withheld.

The model was told to preserve code and protected boundaries, but the prompt
did not enumerate excluded byte spans. The offending line is blockquoted
prose; the current baseline report excludes its containing block. A future
protocol could surface source protection explicitly or investigate whether
that exclusion is appropriate. This run may not be repaired, replayed, or
continued under its frozen packet. No model calls beyond page 2 were made.

## Evidence and check status

- `run/run.json` records `stopped`, `first_error`, and one attempted call;
  SHA-256 `f9cc8cb74773e8581e14fce8523cb1efc71600a102bf5456c6151fe5c872f502`.
- `run/page-002/response.json` preserves the unmodified model response;
  SHA-256 `500dae8f124f2d3021c6f9ef61c13420f4e0e0f91c7a2ad52734d5ca14cd0786`.
- `run/page-002/events.jsonl` preserves the one-turn trace;
  SHA-256 `8af9175996a9d27a13f6a784ea8d851a43c53a952fbee5f616abfd7533a91a2e`.
- `run/page-002/stderr.txt` is empty. No `bound.json`, `review.json`, or
  `result.json` exists because binding failed before scoring.

The sandboxed `make check` reached `check-performance` and stopped because
macOS `/usr/bin/time -l` could not read `kern.clockrate` in the sandbox. The
same command succeeded with the required permission, establishing an
environmental prerequisite rather than a repository test failure. The final
tree still requires an unrestricted `make check` before publication.
