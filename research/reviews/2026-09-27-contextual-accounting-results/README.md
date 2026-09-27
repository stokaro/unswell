# Contextual unit accounting: completed, rejected

The September 27, 2026 GPT-6-Astra medium candidate completed all 88 requests
and all 36 pages. It **failed the frozen development screen**. It is not a
product model and does not change rules, profiles, scores, gates, or the playground.
This is evidence for [#349](https://github.com/stokaro/unswell/issues/349), not its
closure.

| Exposed cohort | Full events | Recall | Accepted findings | Accepted fraction |
| --- | ---: | ---: | ---: | ---: |
| Whole-page audit | 44 / 57 | 77.19% | 319 / 405 | 78.77% |
| Contextual reference | 24 / 46 | 52.17% | 54 / 59 | 91.53% |
| Earlier confirmation pages, now exposed | 14 / 20 | 70.00% | 26 / 30 | 86.67% |

The unchanged requirement is at least 80% full-event recall and 85% accepted
findings **in each cohort**. Across all pages, 82 events received full credit,
32 were missed, and nine were partial. Of 494 emitted findings, 399 were accepted,
77 rejected, and 18 uncertain. Uncertain findings remain in the denominator.
The pooled totals cannot override a failed cohort.

Adding the existing rule baseline raises full-event counts to 49/57, 28/46, and
14/20. This still misses the contextual and follow-up thresholds and does not
establish the precision of a combined diagnostic stream.

## What the result establishes

Every request had full-page context and at most 64 focus units. All 4,549 units
received structural accounting; all responses passed the frozen output and source
checks. This removed the earlier run's incomplete-page problem. It did not make
the requested contribution assessments evidence of semantic coverage.

The contextual cohort remains the largest gap: 11 of 19 wordiness events and
five of eight unjustified-intensifier events were missed. Repetition also needs
better localization: on the whole-page cohort, three of six repetition events
were full and three were partial. A diagnostic that explains repetition but
locates only one required occurrence remains partial under the frozen contract.
The detailed event and finding review is retained locally with the original run.

Many accepted outputs concern grammar or contradictions inside a document. Those
are useful editorial findings, but their volume is not contextual-pattern recall.
Rejected outputs include duplicate diagnostics, optional heading polish,
overliteral readings of conventional metaphors, and requests to expand details
that an overview can reasonably leave to its linked reference.

The judgments were made by the OpenAI assistant and accepted under
[ADR 0041](../../../docs/adr/0041-assistant-review-acceptance.md). They are not
independent human annotation. All 36 pages were already exposed development data;
these results do not estimate population accuracy or demonstrate transfer.

Suggestion safety was reviewed separately: 493 safe, zero unsafe, one uncertain.
No generated edit was applied. A safe suggestion is not proof of a complete or
technically verified repair.

## Execution and preserved evidence

The frozen implementation is commit
`1793b2548a03469b09ff92c134473d68e9ff6e14`. The packet freeze is
`d78326e9297be21a362a3b8821310d4546288de8f3fe0bffb586122dde145a56`.
The original [design, prompt, protocol, evaluator, and freeze](../2026-09-27-contextual-accounting/)
remain byte-for-byte unchanged. Their preparation-time status is historical;
this page records the subsequent authorized execution and decision.

The run used Codex CLI 0.157.1, GPT-6-Astra medium, the existing ChatGPT subscription,
no tools, sequential requests, no retries, and no fallback model. All 88 attempts
were valid. Summed request time was 9,000.66 seconds, within the four-hour limit.
Reported usage was 5,613,815 input tokens, including 276,480 cached input tokens,
and 283,788 output tokens; reasoning-output telemetry was 6,948 tokens.
Monetary cost is unavailable, represented as `null`, not zero.

The complete packet, original responses, CLI traces, execution ledger, editorial
judgments, and reproducible evaluation remain local. They are not included in this
public summary. This limits independent reproduction of the aggregate results;
this page must not be treated as a public qualification artifact.

Locally, a bounded archive replay reproduced the report through the unchanged
frozen evaluator. Four replay checks passed, including rejection of altered tool
flags, removed events, and a fabricated passing summary with refreshed hashes.
The aggregate results above are a status report of that local evaluation.

## Decision and follow-up

Reject this candidate. Do not run it on the separately reserved prospective pages,
do not reduce the thresholds, and do not promote partial findings to full credit.
The earlier single-request Astra trial stopped early, so these runs are not a
controlled causal ablation of the new prompt.

The next candidate must address source-specific editorial omission and restraint:
compare a statement's contribution against its surrounding explanation, preserve
legitimate technical repetition, and localize both sides of a relational finding.
The 77 rejected and 18 uncertain findings also require an explicit precision check;
requesting more findings alone is not an acceptable improvement. Any revised
contract needs a new freeze and must retain all cohorts and all failure cases.
The current experiment remains terminal and its original evidence is not rewritten.
