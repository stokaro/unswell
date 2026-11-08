# #349 source-bound disposition pilot

This is a directional research pilot for editorial semantic diagnoses, not a
qualification result or a product detector change. The question is whether
requiring an explicit `retain`, `revise`, or `uncertain` decision for every
eligible source unit improves complete-event recall over a free-form diagnosis
pass after both arms receive the same exhaustive inventory.

The pinned public README pages are Dagger, Prefect, and Airbyte. They provide
128 eligible units (51, 37, and 40). The Codex assistant reviewed the complete
pages **before model output** and marked four positive events and five protected
counterexamples in [labels.json](labels.json). These are source-only editorial
judgments, not external truth, independent human annotations, or AI-authorship
labels. Airbyte has no frozen positive event and tests overdiagnosis.

The [protocol](packet/protocol.json) bounds execution to nine sequential,
tool-free GPT-6-Astra medium calls through the existing Codex subscription:
one inventory and two review arms per page. Each call has one attempt, a
10-minute timeout, and no retry; the total limit is three hours. The arm order
is counterbalanced across pages. A CLI, schema, trace, coverage, source-binding,
timeout, or resource error stops the run at once. The frozen packet includes
original source bytes, Git blob and SHA-256 identities, labels, prompts,
schemas, request preparation, validation, scoring code, and negative tests.
The model requests never include frozen labels or review grades.

The treatment must disposition all 128 units. Each `revise` decision must
produce exactly one source-bound diagnosis, up to 20 per page. The control
receives the same source and inventory and can return up to 20 free-form
diagnoses. Both arms can return zero findings. No replacement edits are
requested, so suggestion safety is not measured here.

After nine valid calls, a separate assistant review must inspect the complete
combined finding stream and every treatment disposition in source context.
Every delivered finding counts in the acceptance denominator. Each frozen event
is graded FULL, PARTIAL, or MISSED with references to delivered findings; each
protected counterexample is checked for a flag. A directional advance needs
at least 3/4 FULL treatment events, at least two more FULL events than control,
at least 85% accepted delivered treatment findings, complete source binding,
and no protected counterexample flagged. A positive pilot would justify a
larger development test, not completion of #349. The required 36-page,
123-event development cohort and separate prospective confirmation remain.

Offline preparation and validation:

```sh
python3 -B packet.py prepare
python3 -B -m unittest discover -s . -p 'test_packet.py' -v
python3 -B run.py freeze
python3 -B run.py verify
```

The execution and scoring commands are `python3 -B run.py execute` and
`python3 -B score.py score`. Run records are immutable evidence: never retry,
repair, or discard an invalid response to improve a score.
