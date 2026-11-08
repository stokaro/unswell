# #349 repository-disjoint rubric transfer: negative result

The paired pilot tested whether prior editorial decisions from other
repositories would make source-conditioned admission more selective on Ptah
pages. Twelve candidates from twelve distinct Ptah pages were selected
offline, outside the pages in the previous real-source transfer sample.
Six had prior accepted and six had prior uncertain Codex assistant grades.
The control used the merged admission prompt; the treatment added 22
non-Ptah examples and a conservative necessity rubric. The test-case grades
were absent from both requests. These are exposed development cases, not
independent human labels or prospective confirmation.

The packet, both request sets, schema, scorer, and stop policy were frozen in
commit `cf6b4851fde21ac40581fc47d51ae7735f523d3e` before execution.
Packet SHA-256: `4aff664bf6879b3b8ac830007efc30296ea26fd6c00f55f9f70453119297ec34`.
All 24 sequential GPT-6-Astra medium calls returned valid, tool-free,
source-bound responses with no retries or operational errors.

| Measure | Control | Rubric | Registered rubric requirement |
| --- | ---: | ---: | ---: |
| Agreement with prior delivery grade | 6/12 | **8/12** | At least 10/12 and gain of at least 3 |
| Prior accepted retained | 6/6 | **3/6** | At least 5/6 |
| Prior uncertain withheld | 0/6 | **5/6** | At least 5/6 |
| Actual admitted edits | 12 | 4 | Separately reviewed |
| Edit safety | 12 safe | 4 safe | No unsafe or unresolved treatment edit |

The rubric corrected over-admission on five uncertain cases but lost three
accepted cases. Its net agreement gain was only two, so the preregistered
gate failed on overall agreement, paired improvement, and accepted retention.
The separate edit-safety review was performed by a Codex assistant under
ADR 0041; it was not a human review. All 16 admitted edits in the two arms
were assessed safe in their complete source pages. An old uncertain grade
is a nonaccepted delivery judgment, not proof that its diagnosis is false.
The sample is balanced and small, so 8/12 is not a precision estimate.

The calls used 448,691 input tokens, including 129,024 cached input tokens,
4,438 output tokens, and 869 reasoning output tokens through the existing
subscription. Monetary cost is unknown. The run journal SHA-256 is
`13a9a5ffb2451b6cfccb65dddf9b4c1bb44b4eabedecc3863d194ed7893453b4`;
the scored result SHA-256 is
`3112f7c1b722638309f76f6efbf090932fe429a6bd6172109ed425540c9c30ae`.
The independent [`verify_result.py`](verify_result.py) replays frozen hashes,
all 36 original source identities, twelve target spans in both arms, complete
raw traces, tool absence, exact response bindings, edit reviews, and the
registered score. Its negative tests reject changed requests, responses,
trace tool events, review records, and a false pass claim.

This prompt is retired. No runtime selector, old label, original 123-event
denominator, or #349 threshold changed. A next method needs to separate
optional style from material source-specific defects without deleting
technical distinctions; it must be tested on new pages before a full
36-page development stream and separate prospective confirmation. The
required full-event recall and accepted-delivery gates remain 80% and 85%
per complete cohort.

Reproduce the offline result and negative tests with:

```sh
make check-rubric-transfer
```
