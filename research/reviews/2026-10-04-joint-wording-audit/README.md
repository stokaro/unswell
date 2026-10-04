# Joint wording audit: failed development screen

A full-page model audit did not improve contextual editorial detection. The
transport-compatible analysis fully recognizes 55 of the original 123 events,
compared with 97 in the previous, also unqualified configuration. It gains two
events and loses 44. Do not ship this configuration or use it to veto ordinary
engine diagnostics. Related: [#349](https://github.com/stokaro/unswell/issues/349).

The hypothesis was that one source-ordered audit could separate useful technical
content from the wording layers carrying it while checking the current engine's
suggestions. The model received every complete source, original-unit index,
source word ordinals, and all 143 current technical-profile diagnostics. It
received no private grade, reference event, earlier model proposal, or prospective
source. The [prompt](prompt.md) inspects eight wording operators and requires
separate judgments for retained, replaced, rejected, or uncertain engine notes.
Every validated new diagnosis and retained engine note is delivered. Invalid
pages retain all their ordinary notes with an explicit unavailable status.
Accepted-only postselection, diagnosis repair, and automatic edits are absent.

The [protocol](protocol.json) and private evaluation registration predate
inference. The scope remains 36 sources, 4,549 original units, 804 original
judgments, and 123 original events. All 3,844 exposed prior criticisms remain
frozen; an equivalent uncertain or rejected criticism cannot acquire a better
grade merely by receiving a new name or narrower highlight. Full-event credit
requires the actual complete problem and all its original locations, rather
than an overlapping phrase or a grammar repair inside a compound problem.

## Complete result

The [evaluation](evaluation.json) includes the complete actual stream, including
ordinary fallback findings on the two unavailable pages:

| Exposed cohort | Available pages | Full events | Full-event recall | Accepted actual findings |
| --- | ---: | ---: | ---: | ---: |
| Whole pages | 24/26 | 30/57 | 52.6% | 137/235 (58.3%) |
| Context pages | 6/6 | 19/46 | 41.3% | 49/60 (81.7%) |
| Historical confirmation | 4/4 | 6/20 | 30.0% | 19/23 (82.6%) |

Each cohort requires at least 80% full-event recall, at least 85% accepted
findings, every page available, and safe or verification-only advice. None
passes. The 318 delivered findings comprise 205 accepted, 99 uncertain, and
14 rejected findings. All advice is reviewed as verification-only; this does
not make the underlying diagnoses correct. A page can preserve technical
facts and still miss a separate narration, evaluation, or indirect relation.
Identifying one layer inside a passage does not establish complete coverage.

The operator breakdown is retained in the evaluation, including 56 accepted
of 101 evaluation findings and 13 of 28 action-support findings. Ordinary
grammar defects are easier for this configuration than establishing the
necessity of many rhetorical revisions. These are exposed assistant judgments,
not independently measured population precision or authorship probabilities.

## Execution and transport failure

GPT-5.6-Sol with high effort ran through Codex CLI 0.160.0 under the subscription,
with one sequential request per page, no tools, no retries, and no model fallback.
The [execution record](execution.json) reports 36 calls in 8,837.7 seconds,
1,894,156 input tokens, and 380,528 output tokens, including 155,525 reasoning
tokens. The total budget was three hours and each call had a 900-second limit.
Paid API calls were zero; subscription monetary cost is unavailable.

The original strict trace validator marked all 36 calls invalid because the CLI
emitted one startup feature notice as an error item before each model turn.
Those statuses are preserved and the original execution is not passed.
A [separate compatibility analysis](transport-registration.json), registered
after three responses were saved but before their semantic contents were read,
recognizes only that exact single pre-turn metadata notice. Other errors,
multiple turns, tool use, duplicate notices, or altered final messages remain
failures. Six negative transport controls passed. This uses the same cached
responses and adds zero model calls.

Under that interpretation, 34 responses satisfy the complete source contract.
Two still contain an invented diagnosis reference and remain unavailable;
their findings are not salvaged. Compatibility repairs no diagnosis, changes
no source, and does not qualify the original strict protocol.

## Evidence and limits

Review is by the maintainer-accepted root assistant under ADR 0041, not an
independent human or native-speaker annotator. All three cohorts, including
historical confirmation, are exposed development inputs. Untouched prospective
confirmation and runtime integration remain required; neither began here.
The changed model, effort, input representation, and discovery/selection design
prevent a causal model-capacity comparison with the earlier configuration.

The [manifest](manifest.json) binds every original source and baseline request.
[Evidence identities](evidence-identities.json) bind the registered packet,
actual run, delivered stream, private prior index, and complete root review.
Private judgments, reference events, complete responses, and their source
payloads are not republished. Recomputing semantic grades requires the preserved
local sealed packet; this public record alone cannot reproduce that review.

An initial final-review seal included nested registration manifests that the
shared inventory verifier excludes. Its failed inventory is retained locally.
The [sealing record](review-sealing-record.json) binds the corrected inventory
rule and every nested registration; no outcome, model response, or earlier
registration changed.

Run `python3 -B verify.py` and `python3 -B -m unittest test_verify.py` here.
These checks verify public identities, complete counts, ratios, decision gates,
and operator totals. They do not independently prove private semantic grades.
[ADR 0044](../../../docs/adr/0044-joint-wording-audit-rejected.md) rejects this
configuration. Further work must demonstrate a different source-evidence
representation and preserve already recognized contextual problems; another
name for the same criticism or a larger response is not evidence of progress.
