# Verb scaffolding: three exposed gains, no new-page gain

The updated rule adds four warnings on the 88 exposed pages. Three identify
frozen defects; the fourth identifies an additional wording issue absent from
that review. On six new complete pages, full detection stays at **0/43**.
This result does not establish broad useful recall.

## Runtime

`filler.instruction-scaffolding` version 3 recognizes two bounded constructions:
a nominal action or gerund nested between enables/allows and a passive infinitive,
and a nominal action performed through a gerund method. It distinguishes an
action from a concrete object: configuration files being uploaded is a control.
The guidance retains actors, method, conditions and optionality. Ordinary
permissions, simple passives, quotes, questions, protected vocabulary, negation
and conditional actions remain excluded. Candidates stay within 48 tokens.

The [protocol](protocol.md) includes hypotheses that remain unimplemented,
including broader capability chains and role definitions. The catalog still has
53 rules and class manifest r8. Default weights, thresholds and gates are unchanged.
The three gained defects occur in old SQLite and Prometheus documentation; this
step supplies no measured increase on Ptah.

## Frozen source review

The selector excluded all 88 exposed references and hashes and the original
historical study. The assistant froze sources, license notices and selection before
reading prose, then reviewed all six complete extracted pages across all seven
rubric categories. The assistant froze labels at `2026-09-17T17:16:09.790162+00:00` before
runtime edits or diagnostic output: 43 defects, 8 uncertain events, 18 controls.
The reviewer is the implementing Codex assistant, accepted under
[ADR 0041](../../../docs/adr/0041-assistant-review-acceptance.md). These are
assistant development judgments, not human labels, independent qualification,
authorship evidence or population recall. One page per cohort/length cell does
not support an informative within-cell confidence interval.

| Set | Pages | Frozen defects | Before full | After full |
| --- | --- | --- | --- | --- |
| development | 26 | 57 | 14 | 14 |
| exposed_repetition | 6 | 5 | 3 | 3 |
| exposed_framing | 9 | 11 | 1 | 1 |
| exposed_local | 12 | 78 | 10 | 10 |
| exposed_context | 11 | 42 | 8 | 8 |
| exposed_instruction | 12 | 99 | 28 | 28 |
| exposed_construction | 6 | 32 | 7 | 7 |
| exposed_purpose | 6 | 28 | 1 | 4 |
| confirmation | 6 | 43 | 0 | 0 |

Profiles have identical full-event counts. Existing partial credits remain
partial. The CSS styling method warning and an existing In-order-to warning on
the new Raft page receive `additional_actionable` dispositions, with explicit
post-diagnostic edits. Neither increases frozen recall or its actionable fraction.
The validator rejects attaching frozen credit to these rows. Updating the advice
to preserve actors changes 24 existing findings without changing their locations
or coverage; each disposition is retained and explicitly reviewed.

## New pages

| Page | Original path | Defects | Detected |
| --- | --- | --- | --- |
| c01 | docs/site/src/content/docs/inference/concepts/lifecycle.md | 3 | 0 |
| c02 | docs/site/src/content/docs/extend/query-builder.md | 6 | 0 |
| c03 | docs/site/src/content/docs/atlas/retained-divergences.md | 16 | 0 |
| c04 | docs/BUGS.md | 13 | 0 |
| c05 | raft/README.md | 4 | 0 |
| c06 | SUPPORTED_LANGUAGES.md | 1 | 0 |

Ptah remains at 0/25 and the historical pages at 0/18. Technical emits 41 findings:
39 nonactionable, one uncertain and one additional actionable phrase. Strict emits
48: 46 nonactionable, one uncertain and the same phrase. Length or punctuation
that overlaps a wording defect is not credit for detecting it. The
[miss ledger](CONFIRMATION-MISSES.md) preserves all 43 defects. All six pages pass
both gates. That result does not establish that the pages need no editing.

The new-page result remains negative. More isolated grammatical constructions
have not generalized to the main Ptah problems: repeated justifications, evaluative
tails and indirect explanations. The working 80% recall / 85% soft-precision goals
remain unfulfilled. #299 and the broader repetition work in #305 remain open.
These pages are now exposed; another semantic iteration needs fresh confirmation,
not rewritten labels or lowered gates.

## Applicability and reproduction

The existing #312 repeated-claim budget abstention remains on exposed instruction
page c05; its separate PR is outside this branch. The old CMake text-mode limit
also remains visible. No other abstention or operational error appears. Technical
passes every set; strict fails only the preexisting forbidden worth-noting phrase
in the exposed instruction set. No new gate failure is introduced.

Before runtime: `f035dc42613c5c4a78716a1e6a067fa70e2ddd4f`. After runtime: `979dcaac2b20993e138cd8909900a0b60908eaca`.
Before reports for exposed pages retain the prior run and hashes. New-page before
and all after scans use explicit binaries. Reports retain source bytes, source
ranges, policy and runtime identities, commands, host, time and peak RSS.

| Set | Profile | Seconds before → after | Peak MiB before → after |
| --- | --- | --- | --- |
| development | technical | 2.621 → 2.146 | 205.5 → 200.2 |
| development | strict | 1.598 → 1.482 | 204.4 → 205.0 |
| exposed_repetition | technical | 0.621 → 0.525 | 117.0 → 117.9 |
| exposed_repetition | strict | 0.603 → 0.517 | 115.4 → 131.2 |
| exposed_framing | technical | 0.808 → 0.714 | 124.7 → 139.5 |
| exposed_framing | strict | 0.796 → 0.714 | 128.6 → 141.4 |
| exposed_local | technical | 0.718 → 0.593 | 119.3 → 120.4 |
| exposed_local | strict | 0.695 → 0.605 | 118.6 → 117.7 |
| exposed_context | technical | 0.753 → 0.672 | 121.8 → 121.2 |
| exposed_context | strict | 0.741 → 0.694 | 119.9 → 120.9 |
| exposed_instruction | technical | 0.989 → 2.671 | 146.6 → 145.5 |
| exposed_instruction | strict | 0.968 → 0.869 | 163.1 → 160.1 |
| exposed_construction | technical | 0.515 → 0.466 | 102.6 → 105.1 |
| exposed_construction | strict | 0.529 → 0.492 | 102.9 → 103.2 |
| exposed_purpose | technical | 0.515 → 0.486 | 103.8 → 102.5 |
| exposed_purpose | strict | 0.488 → 0.492 | 102.1 → 110.5 |
| confirmation | technical | 0.544 → 0.540 | 104.3 → 104.2 |
| confirmation | strict | 0.475 → 0.537 | 109.4 → 96.0 |

These local scans overlapped ordinary test work and are not a controlled speed
comparison or 2-vCPU performance qualification. Run `python3 tools/render.py` and
`python3 tools/test_evidence.py` here to validate and regenerate the evidence.
`tools/measure.py` replays an explicit binary and set into a new output directory.
No model calls or downloads are needed. Inputs, annotations and unsupported
hypotheses stay frozen.
