# MELD technical-document feasibility: failed polish gate

Status: complete negative screen, November 11, 2026. The pre-inference packet
was committed as `ca1bf78df0ebf3eaaab1273e623d093d01762173`, before any
MELD inference. The exact checkpoint, input, code, environment, resource
budget, and decision rule are in [freeze.json](freeze.json) and
[PROTOCOL.md](PROTOCOL.md). All inference was local on Apple MPS. No hosted
model call, retry, threshold sweep, or partial resume occurred. The run took
2,270 seconds.

The frozen scorer completed 402 eligible calibration controls, persisted its
strict `>` threshold of **3.5719921588897705**, then scored 351 eligible test
controls, 38 matched sources, and all 456 responses. Eight historical controls
were ineligible under the predeclared 100-word minimum. The independent
verification amendment is explained in [AMENDMENT.md](AMENDMENT.md): the frozen
runner overwrote the input `family` field with MELD's predicted family in
response score rows. The original verifier stopped on that mismatch. The
separately committed `verify_v2.py` restores input-family attribution from the
sealed pre-inference manifest while checking every other identity and leaving
all numeric scores and the threshold unchanged. It passed on the original
[saved scores](scores/), and a second run on the copied files produced the
same [result.json](scores/result.json).

| Frozen gate | Observed | Decision |
| --- | ---: | --- |
| Historical test-control alerts ≤1% | 1/351 (0.28%) | Pass |
| Complete-generation detection ≥50% | 147/228 (64.47%) | Pass |
| AI-polish detection ≥20% | 31/228 (13.60%) | **Fail** |

Eligible-document coverage was 402/404 calibration, 351/357 historical test,
228/228 generation, and 228/228 polish. One historical test document hit
MELD's 16,384-token input cap; the other scored documents were not capped.
The one flagged test control came from `TheBerkin/Rant` at
`docs-src/language/functions.md`. Historical commit dates are not verified
human-authorship labels: 1/351 is an *observed historical-control alert rate*,
not a field false-positive rate. These controls and all response pairs are now
exposed development data.

The generation result is uneven across the three saved generator families:
Haiku 67/76, Luna 8/76, and Qwen 72/76. Polish detection was Haiku 8/76,
Luna 5/76, and Qwen 18/76. Each of the 20 response repositories had at least
one flagged generation; only 14 had a flagged polish response. Model scores
were above their own source for 225/228 generations and 212/228 polish
responses, a descriptive paired ranking signal that does not meet the frozen
absolute-threshold gate. All scores are indices, not authorship probabilities.

**Decision:** retire this MELD checkpoint/configuration for the declared
technical-document feasibility screen. Do not select a new threshold on these
test documents. A future origin attempt needs a new source-disjoint frame,
checked human provenance, held-out generators, and a separately frozen
decision. This result supplies no source-local editorial diagnosis, accepted
revision, or progress toward the #349 FULL-event and accepted-finding gates.

To replay the saved-score verification without model weights:

```sh
python3 research/origin/runs/2026-11-11-meld-feasibility/verify_v2.py \
  --scores research/origin/runs/2026-11-11-meld-feasibility/scores \
  --output /tmp/unswell-meld-reverified.json
```
