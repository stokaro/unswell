# Instruction projection: limited gains and false matches

This candidate adds five fully detected frozen defects on the 100 exposed pages.
On six separately frozen complete pages, full detection rises from **2/34 to 3/34**
in both profiles. Ptah remains **1/16**; the historical pages rise from **1/18 to
2/18**. Two additional Ptah defects remain only partially detected. The working
80% recall and 85% soft-precision objectives are not met. Keep this candidate in
review; this result does not justify claiming useful broad coverage or immediate
promotion of the generic-reader recognizer.

## Implementation

`filler.instruction-scaffolding` version 4 follows bounded support predicates to
an action and operand. It recognizes nominal ability, generic reader enablement,
modal used-to actions, nested intended-to-enable passive actions and nominal
actions carried out with a method. The original source tokens retain actors,
operands, optionality and conditions. An affirmative ability statement must not
become an obligation or an assertion that an action actually happened.

A capability announcement and its immediate anaphoric method can form one event
across adjacent paragraphs. Both source ranges are retained. Headings, lists,
excluded code, unrelated sentences and compound announcements break the relation.
Independent comments and strings remain separate. Other events retain their
block scope. Table cells remain outside the rule's declared contexts; the table
entry c06-d04 therefore remains a miss, despite the constructed clause regression.
The broader prefix/relative-information hypothesis c06-d05 also remains unmet.

The new projection uses POS and infinitive roles, not dependency parsing or
semantic equivalence. Its bounds are 48 candidate tokens and 96 sentence tokens.
There are still 53 rules and class manifest r8. Weights, thresholds and gates are
unchanged. The old generic users control is now an explicit positive construction;
named actor permission and concrete capability limits remain negative controls.
Tests exercise source mapping, opaque operands, optionality, conditions,
exemptions, occurrence policy, cancellation and structural boundaries.

## Frozen review and source identity

The selector excluded all 100 exposed references and hashes and the original
historical study. Source bytes, selection, notices and protocol were frozen before
reading. The assistant read every extracted prose block and table cell across the
seven editorial categories, then froze labels at `2026-09-17T18:10:13.870766+00:00` before
runtime edits or diagnostics: 34 defects, eight uncertainties and 24 controls.

The implementing Codex assistant is the maintainer-accepted reviewer under
[ADR 0041](../../../docs/adr/0041-assistant-review-acceptance.md). These are assistant
development judgments, not human labels or independent qualification. One page
per cohort/length cell supports counts rather than population estimates. No
semantic tuning followed inspection of confirmation output. These 106 identities
are now exposed and cannot serve as fresh confirmation for the next iteration.

| Set | Pages | Frozen defects | Before full | After full |
| --- | --- | --- | --- | --- |
| development | 26 | 57 | 14 | 14 |
| exposed_repetition | 6 | 5 | 3 | 3 |
| exposed_framing | 9 | 11 | 1 | 1 |
| exposed_local | 12 | 78 | 10 | 10 |
| exposed_context | 11 | 42 | 8 | 8 |
| exposed_instruction | 12 | 99 | 29 | 29 |
| exposed_construction | 6 | 32 | 7 | 7 |
| exposed_purpose | 6 | 28 | 4 | 4 |
| exposed_verb | 6 | 43 | 2 | 2 |
| exposed_relations | 6 | 45 | 3 | 8 |
| confirmation | 6 | 34 | 2 | 3 |

The five exposed gains are a referrer instruction, cookie reuse, transfer-speed
setup, optional field metadata and a two-paragraph Gantt configuration instruction.
The last previously had two partial diagnostics; relating both ranges earns full
event credit without counting the instruction twice. No other frozen event gains
or loses full credit. All five gains are on historical pages.

## Review burden and false matches

Across the exposed pages, the 14 added/replaced strict findings comprise five
frozen actionable findings, four additional actionable wording edits absent from
the frozen labels, three uncertain judgments and two nonactionable findings.
Two old partial Gantt diagnostics are replaced by one complete diagnostic.
Additional post-diagnostic edits receive no frozen recall credit. The older
repetition/framing sets retain their original restricted annotation scopes;
changed findings outside those scopes are still reviewed separately.

The candidate has two false matches. It takes issue-link residue as the subject
of an imperative Allow user changelog entry. It also treats a relative clause
defining an HTML upload form as a reader instruction. Generic feature introductions in the ImGui and alert-rule guides
remain uncertain, as does a previously frozen highlight.js example. These cases
show that the actor boundary and genre/context decision need further work. The
source-bound judgments and four additional repairs are in [CHANGES.md](CHANGES.md).
Do not silently remove these findings or revise frozen labels to improve results.

## Separate confirmation

| Page | Original path | Defects | Detected |
| --- | --- | --- | --- |
| c01 | docs/site/src/content/docs/operate/kubernetes-operator.md | 4 | 1 |
| c02 | docs/site/src/content/docs/inference/concepts/consistency.md | 10 | 0 |
| c03 | docs/site/src/content/docs/extend/components.md | 2 | 0 |
| c04 | etcdctl/README.md | 4 | 1 |
| c05 | docs/CHECKSRC.md | 9 | 0 |
| c06 | docs/configuration/template_reference.md | 5 | 1 |

Technical emits 17 findings: five actionable (three complete and two partial)
and 12 nonactionable. Strict emits 18 with the same five actionable and 13
nonactionable. The only added finding identifies the nominal ability wrapper in
the Prometheus template reference. All six pages pass both gates.

The [miss ledger](CONFIRMATION-MISSES.md) preserves 31 events without full coverage,
including two partial detections. A length warning overlapping a wording defect
does not earn credit. The adjacent committed tokens have different grammatical
roles; deleting one would break the sentence. That existing repeated-word warning
is recorded as nonactionable, not as detection of the frozen circular explanation.

## Applicability and reproduction

The existing #312 repeated-claim budget abstention remains on exposed instruction
page c05. The old CMake text-mode limit remains explicit. No new abstention or
operational error appears. Technical passes every set; strict fails only the
preexisting forbidden worth-noting phrase in the exposed instruction set.

Before runtime: `e9cb84fcc51671dd865559707063b188569b78df`. After runtime: `d2989e83389751b0fbf2f9d55e0cff8546002fe2`.
Exposed before reports retain the previous run. New-page before reports and all
after reports use explicit binaries. Reports retain source bytes and ranges,
engine and policy identities, commands, host, wall time and peak RSS.

| Set | Profile | Seconds before → after | Peak MiB before → after |
| --- | --- | --- | --- |
| development | technical | 2.198 → 2.216 | 206.4 → 188.8 |
| development | strict | 1.526 → 1.639 | 190.5 → 189.9 |
| exposed_repetition | technical | 0.554 → 0.573 | 116.1 → 126.8 |
| exposed_repetition | strict | 0.537 → 0.592 | 119.0 → 121.0 |
| exposed_framing | technical | 0.726 → 0.979 | 129.4 → 128.5 |
| exposed_framing | strict | 0.727 → 0.827 | 135.2 → 135.3 |
| exposed_local | technical | 0.613 → 0.705 | 114.5 → 122.6 |
| exposed_local | strict | 0.617 → 0.699 | 121.9 → 121.4 |
| exposed_context | technical | 0.652 → 0.729 | 118.8 → 124.2 |
| exposed_context | strict | 0.648 → 0.717 | 118.0 → 120.1 |
| exposed_instruction | technical | 0.858 → 0.941 | 159.8 → 146.1 |
| exposed_instruction | strict | 0.865 → 0.940 | 149.9 → 166.6 |
| exposed_construction | technical | 0.433 → 0.496 | 104.5 → 103.8 |
| exposed_construction | strict | 0.436 → 0.510 | 103.9 → 103.5 |
| exposed_purpose | technical | 0.441 → 0.526 | 103.5 → 103.7 |
| exposed_purpose | strict | 0.438 → 0.491 | 97.6 → 108.9 |
| exposed_verb | technical | 0.485 → 0.544 | 98.3 → 96.9 |
| exposed_verb | strict | 0.483 → 0.552 | 95.6 → 114.2 |
| exposed_relations | technical | 0.514 → 0.577 | 118.1 → 119.8 |
| exposed_relations | strict | 0.516 → 0.579 | 116.3 → 116.0 |
| confirmation | technical | 0.550 → 0.487 | 100.7 → 108.8 |
| confirmation | strict | 0.498 → 0.558 | 111.5 → 110.8 |

These measurements overlapped ordinary checks and are not a controlled performance
comparison or 2-vCPU qualification. Run `python3 tools/render.py` and
`python3 tools/test_evidence.py` here to validate and regenerate the report and
negative evidence probes. `tools/measure.py` replays an explicit binary and set
into a new output directory. Runtime analysis needs no model call or download.
See [VALIDATION.md](VALIDATION.md) for implementation checks and their limits.
