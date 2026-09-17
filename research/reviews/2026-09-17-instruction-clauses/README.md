# Instruction clauses: precision repair without new-page recall gain

On the six fresh complete pages, full frozen-event detection remains **4/58**
in both profiles: **1/19** in Ptah and **3/39** in historical documentation.
Three false positives about concrete capabilities disappear. Previously exposed
pages gain four full detections and two partial detections, but lose two useful
full detections and one additional wording edit. The net exposed full-event gain
is two. Neither the 80% recall nor the 85% soft-precision objective is met.

## Runtime changes and limits

`filler.instruction-scaffolding` version 6 recognizes first-person tutorial setup
with a stated prerequisite and nested need/make-sure layers. It separates the
narrator from the action while retaining prerequisite, actor and ordering.
An actual condition, another actor's state, permission, uncertainty or a quoted
instruction is not enough to match.

Named method/function subjects establish an operation role for used-to or
used-for complements. Optional support adverbs, gerund complements, queried
whether/if values and nested intended-to-allow actions are handled within the
existing bounded projection. Standalone allows-user capability descriptions
now require another support layer or an explicit adjacent method announcement.
The latter restriction removes false positives, but also suppresses real
redundancy when the indirect layer occurs later in the operand or the following
method uses a different construction. Those losses remain in the measurements.

These are bounded surface relations, not full dependency analysis or semantic
entailment. There are still 53 rules and class manifest r8. Thresholds, weights
and gates did not change. Blackbox API and compiled CLI tests cover prerequisites,
actors, permission, uncertainty, named operations, guarded clauses, same- and
cross-paragraph methods, Unicode, BOM, CRLF, comments, strings, policy exemptions,
occurrence allowances, cancellation and resource abstention.

## Frozen review

The selector excluded all 118 exposed source identities and hashes and the
original historical study. Protocol, input bytes, identities and notices were
frozen before prose review. The implementing assistant read 412 extracted prose
blocks and table cells, then froze 58 defects, eight uncertainties and 34 controls
at `2026-09-17T20:04:29.886536+00:00` before runtime changes or diagnostic output.

The implementing Codex assistant is the reviewer accepted under
[ADR 0041](../../../docs/adr/0041-assistant-review-acceptance.md). These are
assistant development judgments, not human labels or independent qualification.
One page per cohort/length cell supports observed counts, not population recall.
All 124 identities are now exposed. No matcher or frozen label changed after
opening confirmation output.

| Set | Pages | Frozen defects | Before full | After full |
| --- | --- | --- | --- | --- |
| development | 26 | 57 | 14 | 14 |
| exposed_repetition | 6 | 5 | 5 | 5 |
| exposed_framing | 9 | 11 | 1 | 1 |
| exposed_local | 12 | 78 | 11 | 11 |
| exposed_context | 11 | 42 | 8 | 8 |
| exposed_instruction | 12 | 99 | 29 | 29 |
| exposed_construction | 6 | 32 | 7 | 7 |
| exposed_purpose | 6 | 28 | 4 | 4 |
| exposed_verb | 6 | 43 | 2 | 2 |
| exposed_relations | 6 | 45 | 8 | 6 |
| exposed_projection | 6 | 34 | 6 | 6 |
| exposed_scope | 6 | 72 | 7 | 7 |
| exposed_proposition | 6 | 46 | 0 | 4 |
| confirmation | 6 | 58 | 4 | 4 |

The exposed full gains are Query's indirect lookup, Load's repeated action,
the pull-image narrator setup and the task-wait narrator setup. The tracing-
purpose and repeated based-off constructions receive only partial credit.
The lost full detections are Curl's referrer wrapper and its transfer-speed
setup; its time-condition wrapper loses an additional post-diagnostic repair.
A Curl interface-lookup wrapper is a new additional repair without frozen credit.
The Prometheus proxy explanation loses a false positive. Older restricted
repetition/framing scopes remain explicit and are not pooled into all-warning
precision. [CHANGES.md](CHANGES.md) retains all changed and removed findings.

## New complete pages

| Page | Original path | Defects | Detected |
| --- | --- | --- | --- |
| c01 | docs/site/src/content/docs/concepts/desired-schema-and-sources.md | 1 | 0 |
| c02 | docs/site/src/content/docs/inference/troubleshooting.md | 10 | 1 |
| c03 | docs/site/src/content/docs/reference/extension-variables.md | 8 | 0 |
| c04 | docs/usage/models.md | 21 | 2 |
| c05 | README.md | 13 | 1 |
| c06 | CONTRIBUTING.md | 5 | 0 |

Technical findings fall from 36 to 33; strict falls from 39 to 36. Both profiles
retain four full events and one partial event. There are three additional useful
repairs found only after diagnostics, eight unresolved judgments, and 17
nonactionable findings in technical (20 in strict). The three removed warnings
are frozen controls: shared TypeVar relationships, overlay/snapshot support,
and CRIU-dependent migration. Their removal improves precision, not recall.

The full-or-partial frozen-label actionable fraction is 5/33 in technical and
5/36 in strict. Even counting the three additional repairs gives only 8/33 and
8/36; neither is evidence for the 85% soft-precision objective. The fragment
in-order-to does not cover perform-the-following-steps, and long-sentence or
contrast-density overlap does not diagnose the frozen wordiness underneath.
All six pages pass both gates. A pass is not proof of satisfactory wording.

The [miss ledger](CONFIRMATION-MISSES.md) retains all 54 events without full
coverage, including one partial event: 14 empty framing, 23 wordiness, six
unjustified intensifier, nine vague claim and two needless repetition events.
No formulaic-transition or needless-complexity defect was frozen in this sample.
The remaining work includes narrated desire and setup, nominal action wrappers,
metaphorical paraphrases, repeated page promises and ungrounded benefit claims.
Another small construction extension cannot by itself establish broad recall.

## Reproduction and validation

Before runtime: `e18f3dcc29bdc46cd5812979238d9e1e0aaa04d3`. After runtime: `4bac934fd7085d5a53f15ad2a71a8f77bd493649`.
The after binary was built from the clean task checkout immediately after the
semantic freeze commit. Exposed before reports retain the preceding iteration;
all after sets were replayed. The new before reports use the pinned before
binary. JSON records original bytes, source mappings, policy/engine identities,
resource costs and the unchanged #312 budget abstention on exposed instruction
page c05. No other abstention or operational error is accepted by the validator.

Technical passes all sets. Strict fails only the preexisting forbidden
worth-noting phrase in the exposed instruction set. The historical CMake
text-mode limitation remains inherited. This work is not deployed to the
playground; local evidence does not establish remote CI or merged-commit status.

| Set | Profile | Seconds before → after | Peak MiB before → after |
| --- | --- | --- | --- |
| development | technical | 2.576 → 3.537 | 194.4 → 211.2 |
| development | strict | 2.015 → 1.637 | 200.8 → 190.5 |
| exposed_repetition | technical | 1.183 → 0.623 | 121.7 → 122.0 |
| exposed_repetition | strict | 0.695 → 0.603 | 113.6 → 125.6 |
| exposed_framing | technical | 1.255 → 0.829 | 141.5 → 127.1 |
| exposed_framing | strict | 1.223 → 0.835 | 131.0 → 132.5 |
| exposed_local | technical | 1.006 → 0.710 | 121.4 → 122.3 |
| exposed_local | strict | 1.078 → 0.712 | 121.0 → 117.3 |
| exposed_context | technical | 1.048 → 0.737 | 129.6 → 120.7 |
| exposed_context | strict | 1.066 → 0.739 | 124.9 → 123.0 |
| exposed_instruction | technical | 1.364 → 0.978 | 152.8 → 165.2 |
| exposed_instruction | strict | 1.329 → 0.989 | 163.8 → 160.5 |
| exposed_construction | technical | 0.879 → 0.528 | 103.9 → 107.4 |
| exposed_construction | strict | 0.764 → 0.510 | 97.9 → 100.8 |
| exposed_purpose | technical | 0.811 → 0.531 | 103.1 → 102.9 |
| exposed_purpose | strict | 0.834 → 0.513 | 106.9 → 104.3 |
| exposed_verb | technical | 0.895 → 0.563 | 104.6 → 98.0 |
| exposed_verb | strict | 0.871 → 0.558 | 99.2 → 100.0 |
| exposed_relations | technical | 0.904 → 0.602 | 115.6 → 108.5 |
| exposed_relations | strict | 0.856 → 0.609 | 117.6 → 124.2 |
| exposed_projection | technical | 0.871 → 0.519 | 110.2 → 109.0 |
| exposed_projection | strict | 0.838 → 0.490 | 111.3 → 105.4 |
| exposed_scope | technical | 1.135 → 0.741 | 164.8 → 154.8 |
| exposed_scope | strict | 0.772 → 0.744 | 166.2 → 163.3 |
| exposed_proposition | technical | 0.387 → 0.434 | 84.0 → 85.3 |
| exposed_proposition | strict | 0.386 → 0.451 | 79.7 → 80.3 |
| confirmation | technical | 0.881 → 0.476 | 85.2 → 93.6 |
| confirmation | strict | 0.588 → 0.495 | 86.7 → 85.7 |

Measurements overlapped ordinary checks and are not a controlled performance
comparison or 2-vCPU qualification. Run `python3 tools/render.py` and
`python3 tools/test_evidence.py` to validate and regenerate artifacts.
`tools/measure.py` replays an explicit binary into a new output directory.
See [VALIDATION.md](VALIDATION.md) for actual product-check results.
