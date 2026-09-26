# Testing evidence and unrestricted guarantees

This candidate extends `filler.unscoped-assurance` from version 7 to 8. On the
unchanged 26-page development audit it recovers two previously missed events:
`p024-d01` (etcd's testing-to-reliability guarantee) and `p025-d07` (Ptah's
unrestricted trust assurance). All prior diagnostics remain intact. This is a
bounded improvement, not completion of the broader recall goal.

## Construction and controls

The matcher recognizes two written relations:

- A testing or review activity *ensures/guarantees* abstract reliability,
  correctness, security, safety, or quality, including the passive form.
- A bare *nothing is taken/accepted on trust/faith* assertion, including
  *we take/accept nothing on trust/faith* and a clause introduced by *so*.

The first construction requires both the testing activity and the abstract
quality noun. Concrete invariants, exhaustive proofs, qualified test domains,
modals, questions, quotations, attribution, and protected code do not establish
it. The second requires the unrestricted quantifier and complete predicate;
explicit request or certificate boundaries remain controls. This is a request
to state scope, not an assertion that the underlying implementation is wrong.

The construction reuses the existing extracted prose, clauses, NLP tokens,
source maps, allowance, activation, and scoring. It adds no network dependency,
model, new rule ID, gate, weight, or threshold. Clause length and candidate
budgets remain unchanged. Its bounded vocabulary cannot assess arbitrary
proofs, infer unstated conditions, or determine authorship.

`cases.json` freezes 14 positive and 34 negative constructed examples before
implementation. `freeze.json` retains their original hash. The first test run
exposed a missing *so* boundary and a fully protected input with no applicable
prose. The matcher now accepts that boundary; the test harness adds a neutral
separate paragraph. Case labels and bytes did not change. The execution freeze
records these corrections. Later function extraction only reduces complexity.

## Complete-page comparison

Both profiles use the unchanged sources and 57 annotated events from the
[September 17 audit](../2026-09-17-full-page-recall/README.md). Every event has a
before/after disposition in `review.json`. Overlapping length or contrast
warnings receive no framing credit. Three partial detections stay partial.

| Stratum | Before | After |
| --- | ---: | ---: |
| Sampled Ptah pages | 8/27 | 8/27 |
| Historical technical documents | 2/21 | 3/21 |
| Previously exposed Ptah anchors | 3/9 | 4/9 |
| All retained events, including anchors | 13/57 | 15/57 |

The aggregate is descriptive, not a population estimate or a replacement for
stratified results. Technical findings rise from 134 to 136; strict findings
rise from 153 to 155. Each profile adds the same two accepted findings and
removes none. Two accepted additions out of two are development judgments,
not evidence of 100% general precision. These pages and labels were exposed
before this change.

The editor should retain etcd's functional-test link and describe what the tests
exercise. Ptah should retain its concrete error-propagation behavior and remove
or scope the trust assurance. The proposed changes concern the wording; the
study does not verify those implementations or apply automatic rewrites.

## Separate confirmation

Four complete current Ptah Markdown pages were selected by a fixed SHA-256 rank
before implementation, excluding inference pages and limiting source size to
2,000–6,500 bytes. The snapshot commit, selection salt, source bytes, and hashes
are retained in `confirmation-inputs.json`. This narrow frame is not a random
sample of all Ptah documentation. These source versions were newly read for
this run; no universal lack of earlier conversational or training exposure is
claimed.

The same Codex assistant read all four pages and froze two definite defects,
uncertain alternatives, and full-file review coverage before seeing their
outputs. Annotation followed the initial candidate implementation; this is not
blinded or independent review. Only the separate constructed tests prompted
changes before execution. No tuning followed confirmation output.

Both versions have identical confirmation findings: four technical and five
strict. There are no new warnings and no examples of the new target family.
Full-event recall remains **0/2**; the import-page announcement is only a partial
match because its preceding imagined-reader setup remains. Two diagnostics are
accepted, two are uncertain, and strict adds one rejected punctuation diagnostic.
An accepted post-freeze wordiness suggestion is not added to the recall labels.
This confirmation tests false additions on these pages but provides no positive
transfer evidence. It cannot qualify the 80% recall or 85% precision goals.

## Reproduction and validation

Baseline: `5c0445655638dac4ec646f1564936412d5a3b5a9`. Reports preserve complete
sources, policies, findings, and exclusions. The original archive and notices
remain in the September 17 record. Current confirmation sources are from the
maintainer-authorized Ptah repository, under its retained `LICENSE.ptah`.
The reviewer is the Codex assistant under ADR0041; no human review is claimed.

Run the offline artifact checks and their negative probes:

```sh
python3 -B research/reviews/2026-09-26-evidence-claims/verify.py
python3 -B -m unittest discover -s research/reviews/2026-09-26-evidence-claims -p test_verify.py -v
```

Build either revision's CLI, then replay all complete pages into a new directory:

```sh
CGO_ENABLED=0 go build -o /tmp/unswell-evidence ./cmd/unswell
python3 -B research/reviews/2026-09-26-evidence-claims/replay.py --binary /tmp/unswell-evidence --output /tmp/unswell-evidence-replay
```

The verifier validates source identities, complete reports, exact snippets,
unchanged old findings, the entire event denominator, and all added and
confirmation diagnostic dispositions. Its negative probes reject source drift,
omitted events, invented credit, and missing judgments. It checks recorded
editorial judgments; it does not independently judge their correctness.

Public-API tests cover all 48 cases and source/allowance behavior. A CLI fixture
checks Markdown emphasis, Unicode, BOM, CRLF, protected code, JSON, and SARIF.
Existing feature goldens change only their ruleset hash. Full project acceptance
is recorded in the PR; this research report is not a CI or release claim.
