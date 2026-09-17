# Rhetorical relations: three exposed gains, no confirmation gain

The two updated rules add three findings on the 94 exposed pages, all identifying
frozen Ptah defects. On six new complete pages, full detection remains **3/45**
in both profiles. Another defect is only partly identified. These results do not
establish good complete-page recall.

## Implementation and rejected shortcut

`filler.evaluative-closure` version 6 recognizes cognitive-worth modifiers on
information nouns, gerund-anaphor purpose clefts, relative feature-purpose
predicates and discourse subjects declaring importance or editorial integrity.
`filler.unscoped-assurance` version 4 recognizes result tails that predict a
generic reader's preferences or understanding. The implementation uses bounded
source-token relations and the existing POS contract. It does not infer semantic
equality, dependencies or the author's identity.

The new tests also exposed an existing negation bug: a negative nominal-worth
complement could match. It now remains a control. Conditions, real component
purposes, named requirements, measured results, quoted claims and protected
operands have close counterexamples. Source mapping, exemptions and occurrence
policy use the existing engine. Bounds remain 48 candidate tokens and 96 sentence
tokens; defaults retain their weights, thresholds and gates. There are still 53
rules and class manifest r8.

The [scope ablation](scope-ablation.json) disabled two guards only in a temporary
development binary. It added two findings across 94 exposed pages, with no other
changes. That experiment does not justify removing the guards. Their original
behavior is retained. The protocol's broader measured-versus-asserted and
coherence hypotheses are not implemented or demonstrated by this change.

## Frozen review

The selector excluded all 94 exposed source references and hashes and the
original historical study. The assistant froze sources, selection and notices before
reading, then reviewed all six complete extracted pages across the seven rubric
categories. The assistant froze labels at `2026-09-17T17:43:05.769056+00:00` before runtime changes
or diagnostic output: 45 defects, five uncertain events and 23 controls.

The implementing Codex assistant is the maintainer-accepted reviewer under
[ADR 0041](../../../docs/adr/0041-assistant-review-acceptance.md). This is assistant
development evidence. Human annotation and independent qualification remain
unperformed. One page per cohort/length cell supports counts; it cannot establish
population recall or a reliable within-cell confidence interval. The selector
retained the short connection guide, which has no definite wording defect. The
review records the unexpanded historical include and excludes its unseen content.

| Set | Pages | Frozen defects | Before full | After full |
| --- | --- | --- | --- | --- |
| development | 26 | 57 | 14 | 14 |
| exposed_repetition | 6 | 5 | 3 | 3 |
| exposed_framing | 9 | 11 | 1 | 1 |
| exposed_local | 12 | 78 | 10 | 10 |
| exposed_context | 11 | 42 | 8 | 8 |
| exposed_instruction | 12 | 99 | 28 | 29 |
| exposed_construction | 6 | 32 | 7 | 7 |
| exposed_purpose | 6 | 28 | 4 | 4 |
| exposed_verb | 6 | 43 | 0 | 2 |
| confirmation | 6 | 45 | 3 | 3 |

Both profiles have the same full-event counts. Three new exposed findings cover
an import-purpose restatement, a decision-worth-stating announcement and an
entry-honesty judgment. [Every changed finding](CHANGES.md) has a source-bound
review. No old finding is removed or changed beyond its rule identity.

## Separate confirmation

| Page | Original path | Defects | Detected |
| --- | --- | --- | --- |
| c01 | docs/site/src/content/docs/operate/ai-agent-connect.md | 0 | 0 |
| c02 | docs/site/src/content/docs/direct/plan-and-approve.md | 3 | 0 |
| c03 | docs/site/src/content/docs/atlas/feature-matrix.md | 9 | 0 |
| c04 | docs/MANUAL.md | 23 | 0 |
| c05 | docs/usage/schema.md | 4 | 0 |
| c06 | docs/gantt.md | 6 | 3 |

Ptah remains at 0/12 and the historical pages at 3/33. Technical emits 35 findings:
29 nonactionable, five actionable and one additional actionable phrase outside
the frozen labels. Strict emits 37: 31 nonactionable, the same five actionable
findings and the same additional phrase. Two of the five actionable findings each
cover only part of one two-block margins instruction; neither gets full credit.
The other three identify existing instruction scaffolding on the Gantt page.
An In-order-to phrase receives no frozen recall credit.

The [miss ledger](CONFIRMATION-MISSES.md) retains all 42 defects without full
coverage, including the partially detected one. Length and punctuation warnings
do not get credit for overlapping a different wording defect. All six pages pass
both gates; that does not establish that they need no editing.

The 80% recall / 85% soft-precision working objectives remain unmet. Successive
bounded-construction additions have produced exposed examples without new-page
improvement. This result argues against treating more such additions as a plan
for broad coverage. General paraphrastic repetition and indirect explanations
remain unresolved in #299 and #305. These 100 pages are now exposed; another
semantic iteration needs a new confirmation set and a different justified
hypothesis. Do not revise these labels or lower gates to obtain a passing result.

## Applicability and reproduction

The existing #312 repeated-claim budget abstention remains on exposed instruction
page c05; its separate fix is outside this branch. The old CMake text-mode limit
also remains visible. No new abstention or operational error appears. Technical
passes every set; strict fails only the preexisting forbidden worth-noting phrase
in the exposed instruction set.

Before runtime: `979dcaac2b20993e138cd8909900a0b60908eaca`. After runtime: `e9cb84fcc51671dd865559707063b188569b78df`.
Exposed before reports retain the previous run and hashes. The new-page before
run and all after runs use explicit binaries. Reports retain source bytes, source
ranges, runtime/policy identities, commands, host, duration and peak RSS.

| Set | Profile | Seconds before → after | Peak MiB before → after |
| --- | --- | --- | --- |
| development | technical | 2.146 → 2.198 | 200.2 → 206.4 |
| development | strict | 1.482 → 1.526 | 205.0 → 190.5 |
| exposed_repetition | technical | 0.525 → 0.554 | 117.9 → 116.1 |
| exposed_repetition | strict | 0.517 → 0.537 | 131.2 → 119.0 |
| exposed_framing | technical | 0.714 → 0.726 | 139.5 → 129.4 |
| exposed_framing | strict | 0.714 → 0.727 | 141.4 → 135.2 |
| exposed_local | technical | 0.593 → 0.613 | 120.4 → 114.5 |
| exposed_local | strict | 0.605 → 0.617 | 117.7 → 121.9 |
| exposed_context | technical | 0.672 → 0.652 | 121.2 → 118.8 |
| exposed_context | strict | 0.694 → 0.648 | 120.9 → 118.0 |
| exposed_instruction | technical | 2.671 → 0.858 | 145.5 → 159.8 |
| exposed_instruction | strict | 0.869 → 0.865 | 160.1 → 149.9 |
| exposed_construction | technical | 0.466 → 0.433 | 105.1 → 104.5 |
| exposed_construction | strict | 0.492 → 0.436 | 103.2 → 103.9 |
| exposed_purpose | technical | 0.486 → 0.441 | 102.5 → 103.5 |
| exposed_purpose | strict | 0.492 → 0.438 | 110.5 → 97.6 |
| exposed_verb | technical | 0.540 → 0.485 | 104.2 → 98.3 |
| exposed_verb | strict | 0.537 → 0.483 | 96.0 → 95.6 |
| confirmation | technical | 0.610 → 0.514 | 111.5 → 118.1 |
| confirmation | strict | 0.536 → 0.516 | 106.9 → 116.3 |

These scans overlapped ordinary tests and are not a controlled performance
comparison or 2-vCPU qualification. Run `python3 tools/render.py` and
`python3 tools/test_evidence.py` here to validate and regenerate the evidence.
`tools/measure.py` replays a binary and set into a new directory. No model call or
resource download is required. The input and annotation freezes remain unchanged.
