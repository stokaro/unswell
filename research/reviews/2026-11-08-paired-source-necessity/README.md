# #349 paired source-necessity pilot: directional gain, failed gate

This exposed development pilot tested whether showing two proposed diagnoses
against the same complete technical page improves admission relative to
judging each independently. It uses eight same-page pairs from the previously
frozen real-source transfer sample. Each pair contains one previously accepted
and one previously uncertain diagnosis by a Codex assistant under ADR 0041.
The old grades are comparison labels, not independent human ground truth or
proof that every uncertain diagnosis is false.

The [protocol](PROTOCOL.md), all eight requests, source-binding schema, old
answer key, runner, and scorer were frozen in pre-result commit `bb3266a`
before the first model response. Packet SHA-256:
`9752424969fcfd16f4c4466ac347ad50f60ba2cf17af42d5598c60f6710efb20`.
The model saw complete original pages and two target byte spans per call; it
did not receive old grades, earlier model answers, or reviewer rationales.

All eight sequential GPT-6-Astra medium calls completed through the existing
Codex subscription without retries, tool use, schema errors, or invalid
source bindings. All 16 target bindings were valid. The calls used 144,871
input tokens, 1,821 output tokens, and 220 reasoning output tokens; no cached
input tokens were reported. Monetary cost is unknown.

| Frozen measure | Independent decisions | Paired decisions | Advance requirement |
| --- | ---: | ---: | ---: |
| Accepted-only pairs | 2/8 | **4/8** | At least 6/8 |
| Previously accepted diagnoses retained | — | **5/8** | At least 7/8 |
| Previously uncertain diagnoses withheld | — | **7/8** | At least 6/8 |
| Valid pair responses and target bindings | — | 8/8 and 16/16 | 8/8 and 16/16 |

The pair comparison withheld most previously uncertain diagnoses and doubled
the count of accepted-only pairs, but it also withheld three previously
accepted diagnoses. It **failed** the registered pair and retention gates.
Examples include curl's direct security-page instruction and Pydantic's
settings-priority entry: the paired reviewer treated their proposed edits as
optional despite earlier acceptance. The old labels remain unchanged; these
disagreements are evidence of editorial boundary uncertainty, not a reason to
retroactively adjust the denominator. The Tokio pair admitted both diagnoses,
including the previously uncertain one.

No replacement edits were requested. Suggestion safety was therefore not
measured. This pilot did not generate a complete page finding stream, measure
the 123 original #349 events, or use reserved prospective pages. The candidate
is stopped and no runtime behavior changes. A further fixed cutoff or prompt
variant on these seen pairs would not establish transfer. The next meaningful
step requires a source-grounded mechanism tested on new page-disjoint material,
then the unchanged complete development and separate confirmation gates.

The immutable [run journal](run/run.json) SHA-256 is
`53bc281355b5c18208a3a02a86a7dbe043b56694e9cd40fadc6b1c99e162200a`;
the [result](result.json) SHA-256 is
`6ce8e5705fb8adea00128621a1bb53d013cd1b500cb2473493de70a1914eff6b`.
The [verification record](verification.json) pins all 33 run files. The
verifier rechecks complete original source hashes, all request bytes, raw
model traces, tool absence, response bindings, frozen grade comparison, and
result arithmetic. Negative tests reject changed requests, a refrozen false
protocol, changed target offsets, tool events, altered responses, a false
advance claim, and a changed verification record.

Reproduce public accounting and negative checks:

```sh
make check-paired-source-necessity
```
