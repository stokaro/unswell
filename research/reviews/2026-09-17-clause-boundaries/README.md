# Clause-boundary repairs and new confirmation losses

The six new pages fall from **4/56 to 2/56** full frozen-event detections in
both profiles. The stricter noun-phrase check loses two useful simplicity
judgments. This candidate is **not ready for acceptance** and does not meet
the 80% recall or 85% soft-precision targets.

The previously exposed sample gains three full detections: a named parsing
function, configure options, and a process-is-simple judgment after a goal.
It loses no exposed full detection and removes two established false positives:
timeout advice and the clock comparison fragment beginning with `and`.
These repairs do not compensate for the fresh confirmation regression.

## Changes and limitations

Instruction-scaffolding version 9 resolves a protected identifier after an
explicit function/method head and recognizes configuration-option antecedents.
The checksum, token, filename and label operand exclusions remain covered.
Protected text stays opaque and never supplies a role keyword.

Unscoped-assurance version 6 requires a main subject and rejects relative
restrictions inside instructions. A goal prefix can precede a separate main
quality assertion without making that assertion scoped. Source spans isolate
the assertion and preserve the goal. The catalog remains 53 rules, manifest r8;
no gate, weight, score threshold, model, dependency or public API changed.

The clock false positive was a simpler structural error than the initial
paragraph-scope hypothesis: `and` was accepted as a subject. Excluding that
fragment fixes this instance without introducing broad paragraph suppression.
This change does not claim general cross-sentence mechanism resolution.

The new losses expose an overbroad determiner-sequence restriction:
`The way the --type flag functions is simple` has a nested nominal construction,
and `Setting up a configuration file is simple` has a particle before its
object. Both are valid subjects for a quality judgment. No runtime tuning was
made after these confirmation outputs were opened. Preserve this negative result
and add both losses to the next development regression set.

## Frozen review

The source selector excluded all 136 previous references and hashes. One short,
medium and long page per cohort were selected before prose review. The same
implementing Codex assistant read all 477 extracted blocks/table cells and
froze 56 defects, six uncertainties and 26 technical controls across all seven
categories before runtime edits. This is maintainer-accepted assistant review
under [ADR 0041](../../../docs/adr/0041-assistant-review-acceptance.md), not
human or independent annotation.

A source-only audit found no complete extracted block of at least 12 whitespace
tokens in previous sources after whitespace normalization. Shorter, partial and
paraphrased reuse can escape that audit. All 142 selected source identities are
now exposed. One page per cohort/length cell does not support a within-cell
bootstrap or a population recall estimate.

| Set | Pages | Frozen defects | Before full | After full |
| --- | --- | --- | --- | --- |
| development | 26 | 57 | 15 | 15 |
| exposed_repetition | 6 | 5 | 5 | 5 |
| exposed_framing | 9 | 11 | 3 | 3 |
| exposed_local | 12 | 78 | 11 | 11 |
| exposed_context | 11 | 42 | 9 | 9 |
| exposed_instruction | 12 | 99 | 30 | 30 |
| exposed_construction | 6 | 32 | 8 | 8 |
| exposed_purpose | 6 | 28 | 4 | 4 |
| exposed_verb | 6 | 43 | 3 | 3 |
| exposed_relations | 6 | 45 | 10 | 10 |
| exposed_projection | 6 | 34 | 6 | 6 |
| exposed_scope | 6 | 72 | 7 | 7 |
| exposed_proposition | 6 | 46 | 4 | 4 |
| exposed_clauses | 6 | 58 | 6 | 7 |
| exposed_action | 6 | 57 | 9 | 9 |
| exposed_roles | 6 | 44 | 1 | 3 |
| confirmation | 6 | 56 | 4 | 2 |

The older repetition and framing reviews retain their narrower scope; they do
not establish precision over every warning. [CHANGES.md](CHANGES.md) contains
every added and removed diagnostic, including lost detections.

## New complete pages

| Page | Original path | Defects | Detected |
| --- | --- | --- | --- |
| c01 | docs/site/src/content/docs/operate/ai-assist.md | 0 | 0 |
| c02 | docs/site/src/content/docs/schema/lineage.mdx | 11 | 0 |
| c03 | docs/site/src/content/docs/versioned/lint.md | 9 | 0 |
| c04 | GUIDE.md | 22 | 2 |
| c05 | docs/CODE_STYLE.md | 7 | 0 |
| c06 | docs/newDiagram.md | 7 | 0 |

Ptah stays at **0/20**. Historical pages fall from **4/36 to 2/36**. The short
Ptah Assist page has no required edits and no findings, so passing it is
appropriate. The other missed events remain visible in
[CONFIRMATION-MISSES.md](CONFIRMATION-MISSES.md): 54 lack full coverage.

Technical findings fall from 39 to 37; strict findings fall from 42 to 40.
After review, two findings fully diagnose frozen events, two offer additional
post-diagnostic phrase repairs, and 14 remain uncertain. The remaining 19
technical and 22 strict findings are nonactionable on this review. Full-event
diagnostic fractions are 2/37 and 2/40; including additional repairs gives 4/37
and 4/40. Those repairs cannot increase frozen recall.

Technical passes all six pages. Strict fails the ripgrep page on the existing
announced-importance phrase; all three Ptah pages still pass. No threshold was
changed to force a failure. A gate result is not a completeness measurement.

## Reproduction

Before runtime: `241c65bb77b23551064bd7e45c2ea57de3434bf3`. After runtime: `dbd11d798beabf030142aab5a11f120a8ef53e26`.
The after binary was built from the clean semantic-freeze commit. Exposed
before reports are retained from the previous iteration, and the fresh before
run uses its pinned binary. All 17 sets were replayed in both profiles with
unchanged sources and policy. The existing #312 budget abstention on exposed
instruction c05 remains explicit; the validator rejects any other abstention,
error or skipped rule.

| Set | Profile | Seconds before → after | Peak MiB before → after |
| --- | --- | --- | --- |
| development | technical | 2.561 → 1.774 | 201.9 → 182.2 |
| development | strict | 1.507 → 1.835 | 183.6 → 206.6 |
| exposed_repetition | technical | 0.533 → 0.791 | 116.2 → 114.5 |
| exposed_repetition | strict | 0.534 → 0.785 | 116.8 → 126.2 |
| exposed_framing | technical | 0.729 → 0.840 | 129.4 → 129.5 |
| exposed_framing | strict | 0.725 → 0.840 | 140.7 → 132.6 |
| exposed_local | technical | 0.607 → 2.877 | 118.9 → 133.7 |
| exposed_local | strict | 0.616 → 1.027 | 120.3 → 123.7 |
| exposed_context | technical | 0.639 → 0.780 | 117.4 → 125.2 |
| exposed_context | strict | 0.643 → 0.954 | 127.0 → 119.9 |
| exposed_instruction | technical | 0.875 → 1.144 | 163.3 → 170.3 |
| exposed_instruction | strict | 0.866 → 1.018 | 165.8 → 159.1 |
| exposed_construction | technical | 0.439 → 0.482 | 98.0 → 103.2 |
| exposed_construction | strict | 0.434 → 0.500 | 103.3 → 103.9 |
| exposed_purpose | technical | 0.441 → 4.091 | 102.6 → 115.6 |
| exposed_purpose | strict | 0.448 → 1.355 | 103.1 → 97.6 |
| exposed_verb | technical | 0.480 → 0.580 | 97.0 → 96.3 |
| exposed_verb | strict | 0.483 → 0.744 | 104.8 → 103.9 |
| exposed_relations | technical | 0.520 → 0.775 | 121.2 → 111.7 |
| exposed_relations | strict | 0.516 → 0.689 | 114.6 → 118.0 |
| exposed_projection | technical | 0.438 → 3.553 | 103.8 → 104.5 |
| exposed_projection | strict | 0.437 → 2.052 | 107.4 → 106.5 |
| exposed_scope | technical | 0.659 → 0.738 | 169.4 → 148.6 |
| exposed_scope | strict | 0.657 → 0.763 | 148.1 → 162.7 |
| exposed_proposition | technical | 0.386 → 0.455 | 90.9 → 82.9 |
| exposed_proposition | strict | 0.383 → 0.492 | 82.0 → 90.8 |
| exposed_clauses | technical | 0.417 → 0.570 | 85.8 → 88.7 |
| exposed_clauses | strict | 0.417 → 0.632 | 86.0 → 84.2 |
| exposed_action | technical | 0.401 → 0.502 | 91.2 → 88.2 |
| exposed_action | strict | 0.387 → 0.510 | 90.1 → 86.1 |
| exposed_roles | technical | 0.438 → 0.549 | 91.9 → 90.4 |
| exposed_roles | strict | 0.433 → 0.500 | 89.8 → 94.2 |
| confirmation | technical | 0.575 → 0.520 | 111.0 → 99.3 |
| confirmation | strict | 0.529 → 0.556 | 101.3 → 103.8 |

These are host observations, not a controlled comparison or 2-vCPU performance
qualification. Ordinary tests overlapped part of the measurement. Run
`python3 tools/render.py` and `python3 tools/test_evidence.py` to validate and
regenerate the report. `tools/measure.py` takes an explicit CLI binary and a
new output directory. [VALIDATION.md](VALIDATION.md) records software checks.
This candidate has not been merged or deployed to the playground.
