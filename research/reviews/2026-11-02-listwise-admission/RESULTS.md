# Source-bound listwise admission pilot for #349

This exposed-development pilot tested whether reviewing all existing
criticisms together on each complete page could retain distinct useful
diagnoses, reject weak or duplicate claims, and avoid unsafe supplied edits.
It is a terminal negative result. The method does not qualify a detector or
change runtime behavior.

The frozen five-page packet contained 522 aliases selected by the earlier
grade-independent six-stream admission. Requests included complete original
public source, UTF-8 byte spans, protected regions, and opaque candidate
IDs. They excluded frozen grades, reference events, earlier verdicts, stream
names, and maintainer answers. The reviewer could retain existing candidates
only; it could not rewrite their diagnoses or suggestions.

All five sequential GPT-6-Astra high calls returned structurally valid,
source-bound, tool-free responses within the registered 900-second per-call
and two-hour total limits. There were no retries. The reviewer retained 93
aliases. An independent offline recount agreed with the frozen evaluator.

| Pilot cohort | Full original events | Accepted delivered criticisms | Unresolved supplied edits | Decision |
| --- | ---: | ---: | ---: | --- |
| Context, 3 pages | 10/14; required at least 13 | 26/30 (86.7%); required at least 85% | 0 | Failed recall |
| Whole, 2 pages | 3/6; required at least 5 | 48/63 (76.2%); required at least 85% | 1 | Failed recall, precision, and safety |

All five pages were available. The complete input pool had supported 13 of
the 14 context events and five of the six whole-page events on these pages.
The selector lost three context and two whole-page full-event witnesses. The
missed criticisms concerned page narration, a repeated metaphor, a slogan
that prefaced its own explanation, first-release warning wording, and
repeated ease claims around a backport procedure. The retained unsafe edit
could narrow stability constraints to a component's stabilization version
despite separately described requirements for unstable components. A cited
source unit did not prove the supplied edit safe.

The five calls used 185,228 input tokens, including 70,016 cached input
tokens, 7,738 output tokens, and 721.17 elapsed seconds. Subscription
monetary cost is unknown. The reviewer was the maintainer-accepted OpenAI
assistant under [ADR 0041](../../../docs/adr/0041-assistant-review-acceptance.md),
not an independent human annotator.

`summary.json` records the packet, raw run ledger, evaluation, independent
audit, source, and upstream freeze hashes with the aggregate measurements.
Raw prompts, traces, responses, individual grades, reference witnesses, and
the separate maintainer pair answers remain in the local archive. The
unchanged full development inventory has 36 pages, 4,549 units, and 123
reference events. This pilot measured five exposed pages; it did not run the
complete cohorts or inspect the untouched prospective confirmation pages.
The full generation and admission pipeline remains unqualified. Do not
rerun or retune this sealed configuration after observing its labels. #349
remains open.
