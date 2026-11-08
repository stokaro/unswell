# Paired source-necessity pilot for #349

Status: frozen preparation; no model result is implied by this document.

## Hypothesis

The independent source-conditioned admission screen retained 11/12 previously
accepted diagnoses but withheld only 3/12 previously uncertain diagnoses. A
comparison of two diagnoses on the **same complete page** may help distinguish
a material, source-supported edit from optional shortening. This is a different
decision task from independently accepting each candidate. It does not create
new candidates and cannot alone qualify the whole-page system.

Eight complete pages have one previously accepted and one previously uncertain
Codex assistant diagnosis in the existing real-source transfer sample. Pair
selection uses only this preexisting same-page condition. Groups are sorted by
name; item IDs are sorted within a group and the order is reversed for every
second group. This puts the accepted item in position A four times and B four
times. The old grades are in a separate answer key and absent from requests.
Neither prior model decisions nor old editorial rationales enter requests.

Each request includes one complete original public source with SHA-256, two
distinct target quotes with original UTF-8 byte offsets, and the two proposed
diagnoses. The model must decide `revise`, `retain`, or `uncertain` separately
for A and B, with concise source-specific reasons. `revise` means that the
specified diagnosis identifies an independently useful, material edit in this
source. Optional brevity, generic style preference, and a different defect at
the target are insufficient. It may revise both or neither. No replacement
text is requested, so edit safety is not measured. Raw source and diagnostics
are data; no tools, outside sources, or authorship inference are allowed.

## Execution and frozen screen

The packet freezes all request bytes, schema, source bindings, model
`gpt-6-astra` at medium effort, Codex subscription route, one sequential
attempt per pair, 10-minute per-call timeout, and two-hour total limit. Eight
calls maximum. Stop on the first CLI, schema, trace, binding, timeout, or
resource error; never retry or repair a response. The run must be committed
only after packet and scorer are committed. Source pages and old review grades
are exposed development material, not new independent human labels or
prospective confirmation.

Baseline independent decisions on these 16 items produce exactly two of eight
correct accepted-only pairs. The paired task advances only if at least six of
eight pairs choose `revise` for the previously accepted diagnosis and `retain`
or `uncertain` for the previously uncertain one, at least seven of eight
previously accepted diagnoses are retained, at least six of eight previously
uncertain diagnoses are withheld, and all sixteen target bindings and eight
responses are valid. All other outcomes stop this candidate. A pass would
justify a new page-disjoint screen, not a product claim. The unchanged #349
36-page/123-event development gates, full combined-stream review, advice
safety, and separate prospective confirmation remain required.

Prepare and verify offline:

```sh
python3 -B run.py prepare
python3 -B run.py verify
python3 -B -m unittest discover -s . -p 'test_*.py'
```

After a frozen commit, execution is `python3 -B run.py execute`. The run
journal and raw traces are immutable evidence, including a failed first call.
