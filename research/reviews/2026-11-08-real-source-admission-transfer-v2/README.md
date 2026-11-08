# #349 real-source admission transfer: negative result

This exposed development screen tested whether the exact independent,
source-conditioned admission prompt that passed the synthetic evidence-flip
pilot transfers to real technical pages. The offline sample contained 24
source-bound diagnoses on 16 complete pages across 12 repository groups:
12 previously accepted and 12 previously uncertain Codex assistant grades.
Those grades are frozen comparison labels, not human ground truth. The sample
was deliberately balanced, so its accepted fraction is not an estimate of
real-world delivery precision.

V1 froze the sample in pre-result commit `71a3f83` and stopped on its first
CLI invocation before any model response because the restricted host process
could not initialize the in-process app-server client. It did not retry.
V2 changed only the host execution permission. Its 24 request files are
byte-identical to v1; the v2 protocol and scorer were committed in `4b3891d`
before the first model request. The v2 packet freeze SHA-256 is
`6deb4b7b2b2e82ffdc3acea95d44a09f5aa6b66b57162b9eca58c129e23bf994`.

All 24 sequential GPT-6-Astra medium calls returned valid source-bound,
tool-free responses without retries or operational errors. The model admitted
20 diagnoses, rejected two, and abstained on two. The separate source review
was by a Codex assistant under ADR 0041, not an independent human annotator.

| Registered transfer measure | Observed | Advance requirement |
| --- | ---: | ---: |
| Archived-grade delivery agreements | **14/24** | At least 21/24 |
| Prior accepted diagnoses retained | **11/12** | At least 10/12 |
| Prior uncertain diagnoses withheld | **3/12** | At least 10/12 |
| Valid source bindings | **24/24** | 24/24 |
| Admitted edit safety | 16 safe, **1 unsafe, 3 unresolved** | Zero unsafe or unresolved |

The registered gate failed on selectivity and edit safety. Nine of the 12
prior uncertain diagnoses were admitted. Several admissions treated optional
shortening or a different phrasing of already understandable prose as a
necessary edit. The unsafe Ptah edit deleted the explicit distinction between
PgBouncer rejecting a startup parameter and a schema being invalid. The
unresolved edits altered a font-coverage claim, removed a performance
assurance, or weakened observed membership variation to mere possibility.
The archived uncertain grades are not automatically proof those diagnoses
are false; the frozen gate still counts them as nonaccepted deliveries.

The 24 calls used 407,294 input tokens, including 120,960 cached input tokens,
4,415 output tokens, and 725 reasoning output tokens through the existing
subscription. Monetary cost is unknown. The independent
[`verify_result.py`](verify_result.py) checks v1's terminal stop, identical
v1/v2 request bytes, all 36 original source hashes, all 24 original target
byte spans, raw model traces, no tool use, exact response bindings, every
edit review, and the scored result. Its six positive and negative tests reject
request, response, review, result, and v1-stop tampering. The v2 run journal
SHA-256 is `cd28653cf4d096637f1d146029019a87ca04c17e052f0c14a47e495294d76118`;
the scored result SHA-256 is
`0dec00c52d16ec79097cb2c429a9cebd76b10e1b442caddc12b7edda05f353d0`.

This result does not change runtime rules, the old grades, the 123 frozen
reference events, or the #349 thresholds. It does not measure original-event
recall or actual full-stream delivery precision. The candidate is stopped:
there is no complete 36-page development run or prospective confirmation.
The next hypothesis needs a stronger source-grounded necessity test and edit
safety check, then a new frozen evaluation without fitting a cutoff to these
24 seen cases.

Reproduce the offline accounting and negative tests with:

```sh
make check-real-source-transfer
```
