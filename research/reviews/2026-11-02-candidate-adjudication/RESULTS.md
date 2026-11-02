# Candidate adjudication studies for #349

These two adaptive exposed-development studies are terminal negative results.
They do not qualify a detector, change runtime behavior, or close #349. Both
draw from the same 36-page development corpus of 4,549 units and 123 reference
events; the pilots tested only selected exposed examples. The original source
inventory is SHA-256
`de5f6644d8446950cbda45c96f23970bef771d1dbf74b477209086e49d87955b`;
the frozen assistant grades are SHA-256
`a4194e85c3b04e26353d1ac740faa0efd838094f896ce3306406de6b8c7774db`.
The reviewer is the maintainer-accepted OpenAI assistant under ADR 0041, not
an independent human annotator.

## Independent candidate judgment

The first 19-request packet stopped before a model answer: strict JSON Schema
rejected `source_refs.uniqueItems`, and 18 requests were unattempted. Its
packet and run hashes are in `summary.json`. The separate corrected v2 packet
tested 19 candidates on `context/c03.md` and completed all 19 sequential
GPT-6-Astra medium calls with valid tool-free answers. The frozen pilot failed:

| Measure | Observed | Required |
| --- | ---: | ---: |
| Fully supported events on `context/c03.md` | 1/2 | 2/2 |
| Accepted diagnoses retained | 4/9 | At least 8 |
| Accepted fraction of delivered diagnoses | 4/4 (100%) | At least 85% |

The only accepted witness for the missed `context/c03-d02` event was a
criticism of the redundant phrase “of any kind” after “no compatibility
guarantee.” The v2 model called the emphasis functional and rejected that
criticism. Five of nine frozen accepted diagnoses were rejected overall. This
is an editorial judgment disagreement, not a source-location failure. Two
retained diagnoses have assessed safe suggestions; two have no
suggestion-safety assessment. The complete 36-page detector quality remains
unmeasured for this configuration.

## Paired diagnosis comparison

The first nine-request pair packet stopped before a model answer because the
outer command sandbox denied local Unix-socket creation needed by Codex CLI.
Its trace was empty and eight requests were unattempted. A local socket probe
failed in that sandbox and passed outside it. The separate v2 packet tested
nine pairs from seven exposed pages, included a preflight before any model
request, and passed all seven offline tests. The v2 run completed all nine
sequential GPT-6-Astra medium calls with valid tool-free answers and no
retries. It failed every preregistered criterion:

| Measure | Observed | Required |
| --- | ---: | ---: |
| Frozen accepted diagnoses kept | 6/8 | At least 7/8 |
| Frozen nonaccepted diagnoses kept | 10/10 | At most 1/10 |
| Correct preference in accepted/uncertain pairs | 1/8 | At least 7/8 |
| Both nonaccepted diagnoses rejected in the remaining pair | No | Yes |

The model frequently treated two diagnoses of the same source span as equally
actionable. In two pairs it rejected the frozen accepted diagnosis and kept
the uncertain one. A shared span or broad category is not proof that the
diagnoses make the same claim. These results reject pairwise adjudication as
the next selector direction; favorable decisions from other runs are not
combined with them.

## Scope and evidence

The individual-candidate v2 calls consumed 287.952 summed call-seconds,
568,615 input tokens, and 6,445 output tokens. The pairwise v2 calls consumed
236.612 summed call-seconds, 228,311 input tokens, and 5,155 output tokens.
Subscription monetary cost is unknown. Both runs used original public source,
original UTF-8 source maps, immutable candidates, sequential single attempts,
disabled tools, and the frozen fail-fast policy. Frozen grades, reference
events, previous model verdicts, and prospective pages were not sent. No
prospective page was inspected or scored.

`summary.json` records packet, run, and evaluation hashes plus the measured
counts. Raw prompts, traces, responses, source copies, and per-example grades
are retained locally rather than included in this public aggregate report.
Consequently a reader without that archive cannot independently reproduce
the semantic grades or the CLI traces. The frozen evaluators replayed the
available local archive and checked source binding, raw trace identity,
tool-free completion, response validity, original grades, and event support.
Local structural audits also bound every selected candidate and reference
target to its original UTF-8 bytes. These checks establish the stated local
result, not a new full-cohort or prospective result.

The unchanged #349 requirements remain at least 80% full-event recall and
85% accepted delivered criticisms in each complete cohort, followed by
separate confirmation and suggestion-safety review. These pilot outcomes do
not justify changing labels, thresholds, product code, or the public gate.
