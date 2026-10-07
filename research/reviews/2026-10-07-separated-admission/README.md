# Separated editorial judgments: recall improves, acceptance fails

The completed study recognizes 104 of the original 123 editorial events,
compared with 96 in the earlier unqualified admission configuration. It gains
14 complete events and loses six. The long-document cohort still fails:
1,627 of 2,101 delivered criticisms are useful (77.4%), and five actual
suggestions have unresolved effects on technical meaning. Do not ship this
configuration. [#349](https://github.com/stokaro/unswell/issues/349) remains open.

## Complete development result

The source scope remains 36 pages, 4,549 original units, 804 original judgments,
and 123 original events. Every one of the 3,588 actual proposals from six frozen
streams receives a judgment. Duplicate proposals, uncertain criticisms, and
rejected criticisms remain in the usefulness denominator when delivered.

| Exposed cohort | Available pages | Complete events | Recall | Useful actual criticisms | Unresolved supplied advice | Accepted |
| --- | ---: | ---: | ---: | ---: | ---: | --- |
| Whole pages | 26/26 | 48/57 | 84.2% | 1,627/2,101 (77.4%) | 5 | No |
| Context pages | 6/6 | 39/46 | 84.8% | 515/548 (94.0%) | 0 | Yes |
| Previously used confirmation | 4/4 | 17/20 | 85.0% | 190/198 (96.0%) | 0 | Yes |

Each cohort must have at least 80% complete-event recall, at least 85% useful
actual criticisms, every contributing page available, and no unresolved supplied
advice. A partial diagnosis earns no complete credit. Components from different
answers cannot be combined into a complete diagnosis.

The [evaluation](evaluation.json) records all 2,847 delivered criticisms:
2,332 accepted, 490 uncertain, and 25 rejected. The uncertainty count is part
of the failure, not evidence supporting acceptance.

## What improved and what became more expensive

The [comparison](comparison.json) uses the same original sources, grades, and
events as the earlier completed `source_surface` admission configuration.
The proposal pool and admission system change together; this comparison does
not establish an isolated effect of the prompt or model.

| Cohort | Complete events: before → after | Useful criticisms: before → after | Newly complete / lost | Nonaccepted deliveries: before → after |
| --- | --- | --- | --- | --- |
| Whole pages | 44/57 → 48/57 | 329/494 (66.6%) → 1,627/2,101 (77.4%) | 7 / 3 | 165 → 474 |
| Context pages | 36/46 → 39/46 | 113/127 (89.0%) → 515/548 (94.0%) | 6 / 3 | 14 → 33 |
| Previously used confirmation | 16/20 → 17/20 | 40/44 (90.9%) → 190/198 (96.0%) | 1 / 0 | 4 → 8 |

Recall and the useful fraction improve, while the absolute number of
nonaccepted remarks rises. Six previously complete events are also lost.
More findings alone would be a misleading progress metric.

## Hypothesis and execution

The [registered protocol](protocol.json) separates editorial necessity from
the safety of the actual supplied suggestion. The [prompt](prompt.md) includes
32 invented policy illustrations. They contain no benchmark reference events
or private benchmark grades. Every request supplies a complete source and every
unchanged public criticism field. No finding is generated, rewritten, combined,
or deduplicated during admission.

The fixed delivery rule requires an independently useful diagnostic and either
no supplied advice, a safe supplied edit, or actual verification instructions
that preserve the source's constraints. Unsafe advice cannot be repaired by
adding a caveat after the model answers.

GPT-6-Astra with medium effort ran through pinned Codex CLI 0.160.1 under the
existing subscription. Four pilot calls completed first; their complete answers
were reused byte for byte after a positive calibration screen. The remaining
32 pages received one sequential call each. All 36 pages are available; no
retries, tools, fallback models, or response repairs were used. The
[execution record](execution.json) retains actual token usage and elapsed time.
The two-hour limit includes pilot inference and preparation between stages.
Paid API calls were zero; subscription monetary cost is unavailable.

## Remaining errors

The model often treats the availability of a shorter sentence as sufficient
reason for a diagnostic. It sometimes overlooks configuration authority,
release-impact summaries, local command context, and bounded API orientation.
A compound criticism can also survive because one grammar complaint is valid
even when a second complaint misreads the workflow.

The five unresolved suggestions concern a compatibility assurance, waiting-time
semantics, a pronoun's antecedent, when stability obligations begin, and a support
commitment. A verification clause does not protect an edit when verification
governs retaining a promise but deletion remains unconditional.

The [component analysis](component-ablation.json) reports all six subsets
registered before the full result was inspected. No individual subset meets
every cohort's targets. The original requests co-judged all six streams;
these subsets do not measure isolated inference or standalone generation.
No component is selected using its cohort's private outcomes.

## Evidence and limits

The [manifest](manifest.json) binds each complete source, original unit count,
proposal count, and actual request. [Evidence identities](evidence-identities.json)
bind the sealed local packet, run, private evaluation, and completed artifact
audit. The audit verifies transport, exact source projection, all indices,
byte-identical pilot reuse, frozen grades, and the arithmetic. It is not an
independent human assessment of editorial quality.

The maintainer accepts root-assistant editorial review under ADR 0041. All three
cohorts are exposed development data, including the earlier confirmation set.
Private judgments, complete responses, and source payloads are retained locally
and are not republished. Public checks reproduce the accounting; they cannot
independently reproduce private semantic adjudication.

Run `python3 -B verify.py` and `python3 -B -m unittest test_verify.py` here.
These checks are also part of `make check`. They reject changed scope, missing
aliases, inconsistent ratios, unsafe advice being passed, and false product
qualification. [ADR 0045](../../../docs/adr/0045-separated-admission-rejected.md)
records the rejection. Standalone generation, new prospective confirmation,
and current-main integration remain unqualified.
