# Source-bound sectionwise admission pilot for #349

This exposed-development pilot tested whether reviewing small groups of
source-adjacent criticisms against the **complete original page** would
improve on the preceding one-call-per-page listwise admission study. The
frozen packet partitioned the same 522 grade-independently selected aliases
on the same five public pages into 23 requests of at most 24 candidates.
Every alias appeared once. The model saw original UTF-8 source, exact units,
protected regions, opaque candidate IDs, diagnoses, and unchanged supplied
advice. It did not see prior grades, reference events, earlier decisions,
stream origins, or prospective pages.

All 23 sequential GPT-6-Astra high calls returned valid, source-bound,
tool-free responses. There were no retries, response repairs, or fallbacks.
An independent offline audit reconstructed the packet partitions, checked
raw trace and response hashes, and recounted every frozen event witness,
grade, and supplied-edit safety judgment. It agreed with the registered
evaluator. The method **failed** its unchanged pilot screen:

| Pilot cohort | Full original events | Accepted delivered criticisms | Unresolved supplied edits | Decision |
| --- | ---: | ---: | ---: | --- |
| Context, 3 pages | 11/14; required at least 13 | 31/34 (91.2%); required at least 85% | 0 | Failed recall |
| Whole, 2 pages | 4/6; required at least 5 | 60/85 (70.6%); required at least 85% | 1 | Failed recall, precision, and safety |

All five pages were available; the model retained 119 of 522 aliases. The
complete input pool had supported 13/14 context and 5/6 whole-page events.
Compared with the earlier listwise run, sectionwise review recovered two
context events and one whole-page event, but lost another context event. It
still rejected all accepted witnesses for a path-versus-switch slogan and a
wordy prerequisite. On the backport page it retained a repetition criticism
while omitting the distinct unsupported-degree component of the same frozen
event. It also retained the earlier unsafe stability edit, whose scope may
omit constraints on unstable components. Smaller batches alone did not
resolve editorial-boundary decisions, compound event support, or edit safety.

The 23 calls used 700,251 input tokens, including 31,744 cached input
tokens, 11,421 output tokens, and 844.69 elapsed seconds under the registered
three-hour limit. Subscription monetary cost is unknown. The reviewer was
the maintainer-accepted OpenAI assistant under
[ADR 0041](../../../docs/adr/0041-assistant-review-acceptance.md), not an
independent human annotator.

`summary.json` preserves the packet, run, evaluation, independent-audit,
source, and upstream freeze hashes with aggregate counts. Raw prompts,
traces, responses, individual grades, and reference witnesses remain in the
local archive. The unchanged development inventory has 36 pages, 4,549
original units, and 123 reference events. This pilot covers five exposed
pages, not the complete cohorts or untouched prospective confirmation.
The full generation and admission pipeline remains unqualified. Do not
rerun or retune this sealed configuration after observing the labels.
Issue #349 remains open.
