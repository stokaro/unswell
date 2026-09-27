# Standalone numerical choice openings

Issue [#337](https://github.com/stokaro/unswell/issues/337) records an exposed
miss: “Four answers, and the right one depends on whether you control the writes.”
The existing numbered-section rule requires a repeated bare-count heading. The
following heading here names a decision, so that earlier rule correctly abstains.

`filler.unnamed-numerical-choice` addresses a different, bounded construction:
a generic count of answers, options, or choices, followed by an unnamed right or
best choice that depends on a condition. It suggests naming the actual decision
while preserving that condition. It does not determine authorship, count list
items, or claim that a stated number is wrong.

## Scope and limits

Only the first sentence of a paragraph, comment, or string is eligible. It must
contain one clause, at most 48 tokens, with a count from two through twenty in
words or 2 through 99 in digits. The construction is followed by
“, and/but the right/best one depends on” and a condition of at least two tokens.
A named topic between the count and choice, a question, quotation, requirement,
heading, list, or colon/semicolon/dash introducing alternatives is outside scope.
Negation inside the selection condition stays in the reported source context.

The new rule reuses the existing matcher budget, source map, term exemptions,
activation observations, reporting, and local score behavior. Candidate exhaustion
is an explicit abstention. It adds a warning with weight 12 and `gate: none`;
there is no changed threshold, gate shortcut, or remote model dependency.

This is experimental editorial policy. An unnamed choice can be appropriate in
other genres or contexts; this small study does not qualify default precision.
Other standalone numerical hooks remain unsupported. Useful counts such as two
credential schemes, protocol sizes, simultaneous data-race conditions, and named
migration paths remain negative controls.

## Source-first review

[inputs.json](inputs.json), [source-review.json](source-review.json), and the
permitted full sources and original notices in `inputs.tar.gz` are bound by
[input-freeze.json](input-freeze.json). They were frozen before implementation or
candidate execution. The actual reviewer is the OpenAI assistant, accepted under
[ADR 0041](../../../docs/adr/0041-assistant-review-acceptance.md), not an independent
human annotator. The review covers this family, not every defect on each page.

The Ptah consistency-mode page is exposed development data at
`654eae5591392278e6c8bce8e54737f780766f19`. The new confirmation pages are the complete
gRPC keepalive guide and Mermaid beginners overview at their recorded historical
revisions. Selection uses metadata hash order with distinct historical repositories.

The selector now reads inherited repository identities, source paths, and newer
`selection.json` and `confirmation-inputs.json` records. Sampling frames are not
mistaken for prior selections. The conservative inventory leaves **no unused,
unreserved Ptah pages** in the fixed frame; the three reserved model-confirmation
pages were not inspected or used. This shortage is disclosed rather than filled
with already reviewed Ptah pages mislabeled as fresh confirmation.

Neither new page contains a target-family positive. Six source-bound technical
or unrelated-construction controls were frozen. They can test new false positives;
they cannot measure sensitivity or justify a broad recall claim.

## Measured outcome

The baseline is merged commit `9e5acaeaceea56496178fcf830c81ed4f50b8620`.
The candidate is `f1a926b978f18bc970716160ec990f8e0cae0b94`; later commits add
catalog test expectations, equivalent line wrapping, and this evidence, not a
different matcher.

| Both profiles, three complete pages | Before | After |
| --- | ---: | ---: |
| Total findings | 3 | 4 |
| Exposed development event detected | 0/1 | 1/1 |
| New confirmation findings | 0 | 0 |
| Confirmation target-family positives | 0 | 0 |

The sole added finding is accepted: it locates original UTF-8 bytes 477–550,
including the source-write condition. A direct revision is “Choose a consistency
mode based on whether you control source writes.” All three previous findings
remain identical, with no removed finding. All scans complete without errors or
abstentions, all six controls remain clear for this rule, and both gates still
pass. The development paragraph index changes from 0 to 12, as expected for the
new advisory signal. Source bytes, extraction, and unrelated document fields stay
unchanged; catalog/configuration identities change explicitly.

The single-run resource observations were 0.40/0.66 seconds before and 0.45/0.65
seconds after for technical/strict. Peak resident memory was approximately
61.8/61.1 MiB before and 59.1/66.5 MiB after. These are observed process costs on
the recorded macOS host, not statistically established speed or memory differences.
[results/summary.json](results/summary.json) records exact identities, hashes,
complete finding dispositions, and costs. The compressed full reports are retained.

## Runtime and consumer verification

Public-API blackbox tests cover the original miss, meaningful revisions, numeric
bounds, technical requirements, named alternatives, source context, term exemptions,
BOM/CRLF/Unicode mapping, cancellation, concurrent ownership, and budget abstention.
Annotated CLI fixtures verify diagnostics in both plain and marked-up prose;
existing E2E changes are catalog/configuration identity hashes only.

The actual playground Go host at sandbox commit
`0066b27418b43d3cd27acbc49f187c2fdaecab45` was compiled against the candidate as
`js/wasm`. Both profiles locate exactly bytes 477–550 on the complete page; the
revision and six other controls remain clear, gates pass, and no panic occurs.
[wasm-result.json](wasm-result.json) and [wasm-identity.json](wasm-identity.json)
record that local consumer check. This does not claim a public playground deployment.

## Replay

Build the two recorded CLI revisions with `CGO_ENABLED=0` and their explicit
`github.com/stokaro/unswell.BuildCommit` stamps, then run:

```sh
python3 -B research/reviews/2026-09-27-standalone-numerical-hooks/tools/measure.py \
  --root research/reviews/2026-09-27-standalone-numerical-hooks \
  --before /absolute/path/to/before-unswell \
  --after /absolute/path/to/after-unswell \
  --output /new/output/directory
```

The runner verifies frozen source identities, complete scans, exact event location,
unchanged prior findings, unchanged extraction, explicit score effects, and gate
behavior. It records per-process wall time and peak memory through `/usr/bin/time`.
The initial verifier incorrectly expected configuration identity and paragraph
maximum to remain unchanged after a new rule; correcting those two explicit
comparisons changed no source, label, matcher, or event-credit requirement.

To repeat the consumer check, build the recorded sandbox host against the recorded
candidate with the same Go toolchain and stamped commit; keep its matching
`wasm_exec.js` and manifest beside the WASM. Supply that sandbox's harness and the
extracted development page to `tools/wasm-probe.mjs`, followed by an output path.
It checks the actual host interface and UTF-8 offsets, not a reimplementation.
