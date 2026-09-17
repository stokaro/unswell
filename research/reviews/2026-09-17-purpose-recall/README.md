# Purpose recall: ten exposed gains, no new-page gain

Three existing rhetorical rules identify ten additional frozen defects on the
82 exposed pages. On six separately selected complete pages, detection remains
1/28 before and after. This change does not establish broad useful recall.

## Runtime change

The rules match clauses that announce a purpose or declare that a fact is worth
knowing. They also match unsupported claims about what most users want, claims
that output proves a tool understood its input, and document announcements. Local evaluation and reader candidates may follow a longer technical
premise: the candidate stays within 48 tokens and the sentence within 96.
Numbers or reasons in a separate premise no longer erase a local evaluation.
A colon followed by advice does not prove a majority or nearly-always claim.
Concrete component purposes, measurements, quotations, reported claims and
technical conditions remain controls. Some proposed purpose/action constructions
remain unsupported; the protocol lists hypotheses, not guaranteed coverage.

Versions are 5 for `filler.evaluative-closure` and
`filler.document-justification`, and 3 for `filler.unscoped-assurance`.
The catalog retains 53 rules and class manifest r8. These are experimental
warnings; weights, thresholds and gate policy are unchanged.

## Source-only evidence

The implementing assistant froze the [protocol](protocol.md), selection code
and input archive before reading the new pages. The review covered all six
complete extracted sources across the seven rubric categories. The assistant
froze labels at `2026-09-17T16:42:38.317558+00:00` before
runtime edits or inspection of diagnostic output: 28 defects, 6 uncertain events
and 15 acceptable controls. Selection excludes all 82 previously exposed source
references and hashes and retains source licenses. The two Ptah MDX pages use
native MDX extraction. Protected code and frontmatter are context only.

The reviewer is the implementing Codex assistant, accepted by the maintainer
under [ADR 0041](../../../docs/adr/0041-assistant-review-acceptance.md).
These are assistant editorial judgments. They are not human labels, independent
agreement, authorship evidence or population estimates. No extra reviewer is
required for this diagnostic work. One page per cohort/length cell does not
support an informative within-cell confidence interval.

| Set | Pages | Frozen defects | Before full | After full |
| --- | --- | --- | --- | --- |
| development | 26 | 57 | 13 | 14 |
| exposed_repetition | 6 | 5 | 3 | 3 |
| exposed_framing | 9 | 11 | 1 | 1 |
| exposed_local | 12 | 78 | 9 | 10 |
| exposed_context | 11 | 42 | 7 | 8 |
| exposed_instruction | 12 | 99 | 27 | 28 |
| exposed_construction | 6 | 32 | 1 | 7 |
| confirmation | 6 | 28 | 1 | 1 |

Both profiles detect the same frozen defects. One document-dedication warning
partly covers an exposed framing event; its future-description sentence remains
undetected, so it receives no full credit. Existing partial credits remain
partial. Two changed locations include a conjunction without changing their
editorial meaning; one warning window now covers two separately located events.
Every added and removed finding has a source-bound disposition.

## New-page result and remaining work

| Page | Original path | Defects | Detected |
| --- | --- | --- | --- |
| c01 | docs/site/src/content/docs/operate/overview.mdx | 1 | 0 |
| c02 | docs/site/src/content/docs/start/install-options.mdx | 3 | 0 |
| c03 | docs/site/src/content/docs/schema/protobuf.md | 9 | 1 |
| c04 | README.md | 6 | 0 |
| c05 | docs/GOVERNANCE.md | 6 | 0 |
| c06 | docs/configuration/alerting_rules.md | 3 | 0 |

The one detection is the existing worth-stating announcement on Ptah's Protobuf
page. Ptah remains at 1/13; the historical pages remain at 0/15. No confirmation
finding changes. The [miss ledger](CONFIRMATION-MISSES.md) retains all 27 misses.
They include purpose restatements, document self-reference, indirect instructions,
vague rankings and repeated explanations. All six pages pass both profile gates;
passing is not evidence that their wording is clean.

The new rules did not catch any more defects on the six new pages.
Another matching fixture or a lower gate cannot establish the working 80%
recall / 85% soft-diagnostic precision objectives. Those objectives remain open.
The six confirmation pages are now exposed. Further semantic tuning requires
fresh source-only confirmation; do not relabel these misses or overwrite this
negative result. #299, #305 and unsupported #309 constructions remain open.

## Review burden and applicability

| Set | Profile | Findings | Actionable | Uncertain | Nonactionable |
| --- | --- | --- | --- | --- | --- |
| development | technical | 128 | 25 | 38 | 65 |
| development | strict | 147 | 25 | 42 | 80 |
| exposed_local | technical | 75 | 10 | 27 | 38 |
| exposed_local | strict | 81 | 10 | 27 | 44 |
| exposed_context | technical | 40 | 11 | 8 | 21 |
| exposed_context | strict | 49 | 11 | 8 | 30 |
| exposed_instruction | technical | 154 | 30 | 18 | 106 |
| exposed_instruction | strict | 169 | 30 | 18 | 121 |
| exposed_construction | technical | 62 | 7 | 3 | 52 |
| exposed_construction | strict | 66 | 7 | 3 | 56 |
| confirmation | technical | 12 | 1 | 0 | 11 |
| confirmation | strict | 15 | 1 | 0 | 14 |

Only one warning identifies a frozen defect on the new pages. There are 11
other findings in technical and 14 in strict. Length and punctuation warnings
do not identify the frozen wording edits. The old CMake text-mode limitation and
the single repeated-claim budget abstention on exposed instruction page c05
remain visible; this change does not include the separate #312 optimization.
There are no other abstentions or operational errors. All technical gates pass;
strict still fails only on the exposed instruction set's existing forbidden
worth-noting phrase. No new gate failure is manufactured.

## Reproduction

Before runtime: `fcb6cdd8dc69569c726cd6df4dda8ef32345534e`. After runtime: `f035dc42613c5c4a78716a1e6a067fa70e2ddd4f`.
The old before reports retain the prior measurement and original hashes. The new
confirmation before run uses the same preceding runtime binary. All after runs
use the new binary. Reports contain sources, exact byte ranges, model and policy
identities, commands, CPU/RSS, host and artifact hashes.

| Set | Profile | Wall seconds before → after | Peak MiB before → after |
| --- | --- | --- | --- |
| development | technical | 1.671 → 2.621 | 189.6 → 205.5 |
| development | strict | 1.682 → 1.598 | 186.5 → 204.4 |
| exposed_repetition | technical | 0.602 → 0.621 | 116.9 → 117.0 |
| exposed_repetition | strict | 0.644 → 0.603 | 114.6 → 115.4 |
| exposed_framing | technical | 0.847 → 0.808 | 137.4 → 124.7 |
| exposed_framing | strict | 0.827 → 0.796 | 129.7 → 128.6 |
| exposed_local | technical | 0.711 → 0.718 | 118.1 → 119.3 |
| exposed_local | strict | 0.708 → 0.695 | 118.5 → 118.6 |
| exposed_context | technical | 0.727 → 0.753 | 116.5 → 121.8 |
| exposed_context | strict | 0.776 → 0.741 | 114.1 → 119.9 |
| exposed_instruction | technical | 1.069 → 0.989 | 163.0 → 146.6 |
| exposed_instruction | strict | 0.994 → 0.968 | 146.9 → 163.1 |
| exposed_construction | technical | 1.106 → 0.515 | 105.1 → 102.6 |
| exposed_construction | strict | 0.469 → 0.529 | 101.5 → 102.9 |
| confirmation | technical | 0.586 → 0.515 | 97.5 → 103.8 |
| confirmation | strict | 0.561 → 0.488 | 104.5 → 102.1 |

These local timings are not a controlled performance comparison or 2-vCPU
qualification; runs occurred at different times and the new scans overlapped
ordinary test work. They do not justify a speed claim.

Run `python3 tools/render.py` and `python3 tools/test_evidence.py` here to validate
and regenerate the evidence without model calls or network access. CLI replay
uses `tools/measure.py` with an explicit binary, set and new output directory.
The validator rejects frozen-input drift, missing review rows, unsupported
cross-page credit, incidental overlap and hidden applicability changes.
