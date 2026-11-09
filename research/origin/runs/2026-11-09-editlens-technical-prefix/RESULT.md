# EditLens technical-prefix screen: positive paired signal

The [protocol](PROTOCOL.md) and [runner](run.py) were frozen in
[`freeze.json`](freeze.json), SHA-256
`a6ce8835d8998647e35f11d05aac181d86eab12fa530d9ff97ced84e0dd4a01b`,
before the first technical-document score. The 38 original documents and all
456 saved responses matched the pinned long-prose-v3 archives; the checkpoint
and tokenizer matched the earlier [ARB screen](../2026-11-09-editlens-transfer/RESULT.md).
No model was called to create new text, and no raw text was written into this
artifact. The 228 responses per operation share 38 sources in 20 repositories,
so they are not 228 independent source examples.

| Frozen direction measure | Full generation | AI polish |
| --- | ---: | ---: |
| Eligible source-response pairs | 228/228 | 228/228 |
| Median response-minus-source score | +0.7085 | +0.0366 |
| Responses scoring higher than own source | 220/228 | 178/228 |
| Repositories with positive mean delta | 20/20 | 20/20 |
| Positive median delta in all three families | Yes | Yes |
| Registered direction gate | Passed | Passed |

The independent [`verify.py`](verify.py) recalculation passed. The complete
[`result.json`](result.json) has SHA-256
`e676daae5b490db7801e29b0200bd11b7001a455960dbda63f5ffea57dcd695f`.
No-text [`scores.json`](scores.json) and [`pairs.json`](pairs.json) have SHA-256
`5aa52d8ff0ae64cf5a190e7770cd55e1d37322fb20f6dea5422a8fdc8d703893`
and `31a299cd7a7ae77ef2b051262ea355643538cc3d64f707af7d71d09654c3de7c`.
The median polish delta was positive for Haiku (+0.0261), Luna (+0.0581), and
Qwen (+0.0232). Full-generation family medians ranged from +0.4089 to +0.8175.

The model accepts at most 512 tokens. **All 38 originals and 429 of 456
responses exceeded that limit**, so this measures only beginning-of-document
prose after code and URL removal. No untruncated source-response pair exists
for a whole-document sensitivity. Source documents retained a median 71.9%
of raw words after code and URL removal; responses retained 79.2%. Generated
responses were shorter than their sources (median cleaned-word ratio 0.735),
while polished responses were similar in length (0.984). These differences
could affect the score, especially for generation.

The original pages' human authorship is unknown, and no human calibration or
source-disjoint test threshold exists for this corpus. Therefore these deltas
are **not** true-positive rates, false-positive rates, or evidence that the
model can identify AI authorship on a whole technical document. The preceding
ARB test had strong labeled results, but its source collections may overlap
the EditLens training collections. Together the screens identify a promising
local, non-generative **origin** method that deserves a new technical-doc
benchmark with verified human origins, source and generator holdouts, an
independent low-FPR threshold, and an explicit whole-document window policy.

Issue #349 remains open. EditLens gives no source-local editorial diagnosis
or safe replacement text, and the frozen #349 recall and accepted-finding
targets have not changed.
