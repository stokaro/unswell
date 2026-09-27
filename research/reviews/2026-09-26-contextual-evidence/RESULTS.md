# GPT-6-Astra whole-page contextual review

The September 27, 2026 experiment failed its development screen. The candidate
produced useful diagnostics, but the two fully processed cohorts missed the
required 80% full-event recall. The third cohort was incomplete after a response
violated the exact-unit quotation contract. The model is not integrated into
Unswell, and the broader recall goal remains open.

## Execution

The maintainer authorized up to 36 sequential requests to GPT-6-Astra at medium
effort through the existing ChatGPT subscription. The frozen packet contained
36 complete public pages, 639,627 source bytes, and 4,549 existing engine units.
Requests did not contain the 123 reference labels, baseline findings, or repairs.
The protocol allowed one attempt per page, 600 seconds per request, and two hours
total. The preparation commit was `bf74314a8a1c91b66a1da493c00f72fccc80b3cd`.

Codex CLI 0.157.1 emitted an unstable-feature startup notice as an error item.
The original runner stopped after the first completed response. A narrowly tested
[compatibility amendment](EXECUTION-AMENDMENT.md) accepted that exact pre-turn
notice, retained the original ledger and trace, and continued only unattempted
pages under the original deadline. Model, effort, prompt, labels, source checks,
and thresholds did not change. This amendment is part of the execution record,
not an unchanged preregistered execution.

The run made 19 requests, with 18 valid responses. Request 19 completed but failed
source binding; the remaining 17 requests were not sent. There were no retries or
model fallbacks. Recorded request durations sum to 1,644.2 seconds. CLI usage totals
are 612,662 input tokens, including 15,360 cached input tokens, and 52,080 output
tokens; reported reasoning output is 1,873 tokens. Dollar cost is unavailable and
remains null. The validated traces contain no tool calls.

### Why request 19 failed

The capabilities page response contained three findings. One quoted the Markdown
opening `**Deployment reports**` inside unit `u0038`, whose source range starts
at `Deployment reports**`. The quote exists in the full page but extends outside
the specified unit by the two opening asterisks. This is a source-binding failure,
not an invented passage or a demonstrated semantic misdiagnosis.

The frozen contract rejects the complete response. No quotation was repaired,
none of its three findings received diagnostic credit, and the page was not
retried. The raw response and `invalid-response-analysis.json` remain in the
archive. A future experiment may test a different localization contract, but
this result cannot be rescued by changing that contract after observing it.

## Reviewed results

The actual reviewer was the Codex assistant under ADR 0041, not an independent
human annotator. Each finding has a disposition and separate suggestion-safety
judgment. Each frozen defect has an explicit semantic match or miss; overlapping
a target, or incidentally removing it in a suggested rewrite, does not earn credit.

| Frozen cohort | Valid pages | Full events / planned events | Full recall | Accepted findings | Accepted fraction |
| --- | ---: | ---: | ---: | ---: | ---: |
| Whole-page audit | 8/26 | 16/57 | 28.1% | 49/51 | 96.1% |
| Contextual reference | 6/6 | 14/46 | 30.4% | 34/38 | 89.5% |
| Ptah follow-up | 4/4 | 9/20 | 45.0% | 18/19 | 94.7% |

The whole-page cohort has 24 labeled events on valid pages and 33 on unavailable
pages. Its supplementary valid-page recall is 16/24 (66.7%). Those 33 unavailable
events stay in the planned denominator but are not observed semantic misses.
Reporting only the valid-page subset cannot satisfy the completeness requirement.
The other two cohorts are complete, so their low recall does not depend on the
execution failure. The original six-page cohort uses all 46 frozen events, not
the smaller 28-event subset from an earlier experiment.

Across valid responses, 101 of 108 findings were accepted and seven remained
uncertain. Uncertain findings stay in the denominator. All 108 suggestions were
judged safe as advice; no edits were applied or qualified for automatic use.
Safety is not proof that the suggestion completely repairs its own diagnosis:
for example, a BuildKit suggestion retains part of the generic scheduler-aim
sentence it criticizes. The review records that distinction.

The technical baseline plus accepted model detections covers 26/57, 18/46, and
10/20 events respectively. These unions are descriptive recall counts, not a
qualified combined detector: the complete combined finding stream has not been
reviewed for precision here.

## What the model added and missed

Accepted findings include internal scope contradictions, ambiguous references
across paragraphs, an unsupported universal claim about untested DDL parsers,
and framing such as “the scoping is the whole point.” On the command inventory,
the model detected both the layout justification and “the column earns its
place.” One page had no emitted findings and no frozen defects.

Useful additional findings do not retroactively enlarge the reference set or
increase recall. For example, a correct warning about an overbroad conversion
promise does not count as detecting redundant navigation in the same sentence.
The review also leaves several locally plausible findings uncertain because
nearby qualifications already explain the behavior in the complete page.

Misses include “what differs is reach, not standing,” “applying that plan is what
closes the loop,” the measured-rather-than-assumed aside, and many contextual
reference defects. The candidate improved over the rules on these exposed pages,
but that improvement does not establish the required completeness or transfer.

## Decision and retained evidence

Reject this model/prompt/response-contract candidate for product qualification.
Do not change the gate, lower thresholds, substitute authorship for quality, or
present its scores as probabilities. The separately frozen three-page prospective
Ptah confirmation set was not sent to the model and remains untouched.

The next research design must address both semantic omissions and brittle unit
localization. It must be fixed before any new execution and must retain the
complete-page evaluation and false-positive controls. Repeating this run with
only repaired quotes would not resolve the already demonstrated recall failures.
Follow-up work is tracked in [#349](https://github.com/stokaro/unswell/issues/349).

`evidence.tar.gz` contains the packet, raw outputs, judgments, original and amended
ledgers, invalid-response analysis, summary, and source notices. `summary.json`
adds page-level results and category breakdowns. `verify_run.py` reconstructs the
result offline, verifies the original evaluation freeze, and checks every valid
trace, command, source hash, quote, and review mapping. The preparation and
evaluation files originally frozen before execution have not been changed.

These are exposed development pages judged by the same assistant. The findings
are not population accuracy, independent human qualification, causal model-only
comparisons, or evidence of machine authorship.
