# Construction recall: useful exposed gains, limited confirmation

Five existing rule families now recognize more grammatical scaffolding and
repeated definitions. On exposed instruction pages, full detections increase
from 12/99 to 27/99. On six separately selected complete pages, they increase
from 0/32 to 1/32. The broad useful-recall objective remains unmet.

## Changes and limits

Cognitive-worth announcements and impersonal modal notices retain attached
reasons and conditions. Indirect instructions cover gerund methods, nominalized
methods, reader-purpose clauses and additional supported actions. Document
maintenance announcements and two bounded relational repetitions are included.
The engine still uses local NLP, source maps, clause limits and existing budgets.
Quoted and protected construction words remain controls; ordinary actors,
qualified relations and operational constraints must remain in the proposed edit.

Rule versions are 4 for evaluative closure and document justification, and 2 for
instruction scaffolding, redundant predicates and definition echo. The catalog
still has 53 rules and class manifest r8 is unchanged. All five remain experimental
warnings. Weights, profile thresholds and gates are unchanged.

The implementing Codex assistant reviewed sources under
[ADR 0041](../../../docs/adr/0041-assistant-review-acceptance.md). These are the
maintainer-accepted reviewer's judgments on the selected pages. They are not
human labels, independent agreement, authorship evidence or population accuracy.
No additional reviewer is required for this diagnostic work.

## Frozen inputs and observed recall

The [protocol](protocol.md) and input identities were frozen before reading the
selected prose. Annotations were frozen at `2026-09-17T15:40:52.866214+00:00`, before runtime changes or
inspection of diagnostic output. They contain 32 defects, 7 uncertain events
and 17 acceptable controls across all seven categories. Selection excludes all
76 previously reviewed sources and hashes, uses one page per cohort/length cell,
and retains source licenses. Shared repositories do not establish independent
domain transfer. No interval is estimated from a single page in each cell.

| Set | Pages | Frozen defects | Before full | After full |
| --- | --- | --- | --- | --- |
| Original development | 26 | 57 | 13 | 13 |
| Exposed repetition targets | 6 | 5 | 3 | 3 |
| Exposed framing targets | 9 | 11 | 1 | 1 |
| Exposed local pages | 12 | 78 | 7 | 9 |
| Exposed context pages | 11 | 42 | 7 | 7 |
| Exposed instruction pages | 12 | 99 | 12 | 27 |
| New complete-page confirmation | 6 | 32 | 0 | 1 |

Both profiles give the same full-event credits. The 17 new exposed detections
comprise two local-page events and 15 instruction-page events. One exposed
notice covers only part of a wordiness event; one additional indirect method
has no frozen defect label and remains uncertain. Source-only labels stay intact.

The new confirmation gain is the worth-and-surprise announcement on the Ptah
target-layout page. Ptah moves from 0/18 to 1/18; historical sources remain 0/14.
Five pages gain no full detection. An existing wordy-phrase finding covers only
part of the containerd introduction and never counts as a full detection.
The working 80% recall / 85% soft-diagnostic precision goals remain unfulfilled.

| Page | Original path | Cohort | Length | Defects | Detected |
| --- | --- | --- | --- | --- | --- |
| c01 | docs/site/src/content/docs/schema/yaml.md | ptah | short | 1 | 0 |
| c02 | docs/site/src/content/docs/inference/strategies/choose-a-target-layout.md | ptah | medium | 7 | 1 |
| c03 | docs/site/src/content/docs/versioned/apply.md | ptah | long | 10 | 0 |
| c04 | CMakeLists.txt | historical | long | 3 | 0 |
| c05 | docs/content-flow.md | historical | medium | 6 | 0 |
| c06 | docs/Setup.md | historical | short | 5 | 0 |

| Category | Defects | Detected |
| --- | --- | --- |
| empty_framing | 11 | 1 |
| formulaic_transitions | 0 | 0 |
| needless_complexity | 2 | 0 |
| needless_repetition | 6 | 0 |
| unjustified_intensifiers | 3 | 0 |
| vague_claims | 4 | 0 |
| wordiness | 6 | 0 |

## Review burden and applicability

| Set | Profile | Findings | Actionable | Uncertain | Nonactionable |
| --- | --- | --- | --- | --- | --- |
| Original development | technical | 127 | 24 | 38 | 65 |
| Original development | strict | 146 | 24 | 42 | 80 |
| Exposed local pages | technical | 73 | 8 | 27 | 38 |
| Exposed local pages | strict | 79 | 8 | 27 | 44 |
| Exposed context pages | technical | 39 | 10 | 8 | 21 |
| Exposed context pages | strict | 48 | 10 | 8 | 30 |
| Exposed instruction pages | technical | 153 | 29 | 18 | 106 |
| Exposed instruction pages | strict | 168 | 29 | 18 | 121 |
| New complete-page confirmation | technical | 57 | 2 | 3 | 52 |
| New complete-page confirmation | strict | 61 | 2 | 3 | 56 |

Actionable includes partial matches. Full recall counts each frozen event once.
Unlabeled plausible edits stay uncertain. The one added confirmation warning
is actionable, but a single positive cannot qualify the rule's default precision.
The large nonactionable count is an existing product limitation, not evidence
that every long technical sentence or contrast needs revision.

The historical long cell selected curl's CMakeLists.txt in its inherited text
format. It produces 24 warnings over program structure, including repeated
feature/protocol/backend inventories. This is a source-format limitation:
comments and option descriptions were reviewed, but code and copyright text are
not editorial defects. The page remains in all measurements and denominators.
It reinforces the embedded-code follow-up in #307; no sample was replaced.

The existing repeated-claim budget abstention remains on exposed instruction
page c05 in both phases and profiles (#312). There are no other abstentions or
operational errors. All technical gates pass. Strict fails only on the exposed
instruction set, as it did before this iteration, because of the preexisting
forbidden worth-noting phrase. All new confirmation pages pass both gates.

## Evidence and replay

Before: `296ebe84ee30c94ec0df46c04a1a8773e7db68cc`. After: `fcb6cdd8dc69569c726cd6df4dda8ef32345534e`.
The initial implementation is recorded in refactor-equivalence.json; replay after
splitting helper functions preserved every document, finding, gate, abstention,
error and status. Only the tool commit changed in report manifests.

Every added and removed finding has a disposition. Changed editing guidance is
reviewed as a replacement even with an unchanged location. Cognitive-worth and
notice diagnostics may identify either wordiness or empty framing in the rubric;
the same semantic eligibility applies to both engines, and source-bound full or
partial judgments are still required. This changes no frozen label or baseline
credit. [CHANGES.md](CHANGES.md) lists all reviewed changes. The
[miss ledger](CONFIRMATION-MISSES.md) retains all 31 missed confirmation defects.

| Set | Profile | Wall seconds before → after | Peak MiB before → after |
| --- | --- | --- | --- |
| Original development | technical | 2.194 → 1.671 | 183.3 → 189.6 |
| Original development | strict | 1.512 → 1.682 | 187.7 → 186.5 |
| Exposed repetition targets | technical | 0.540 → 0.602 | 112.7 → 116.9 |
| Exposed repetition targets | strict | 0.539 → 0.644 | 118.3 → 114.6 |
| Exposed framing targets | technical | 0.734 → 0.847 | 137.6 → 137.4 |
| Exposed framing targets | strict | 0.741 → 0.827 | 142.9 → 129.7 |
| Exposed local pages | technical | 0.618 → 0.711 | 121.6 → 118.1 |
| Exposed local pages | strict | 0.615 → 0.708 | 119.6 → 118.5 |
| Exposed context pages | technical | 0.654 → 0.727 | 117.9 → 116.5 |
| Exposed context pages | strict | 0.739 → 0.776 | 126.8 → 114.1 |
| Exposed instruction pages | technical | 0.886 → 1.069 | 145.1 → 163.0 |
| Exposed instruction pages | strict | 0.873 → 0.994 | 160.3 → 146.9 |
| New complete-page confirmation | technical | 0.742 → 1.106 | 96.2 → 105.1 |
| New complete-page confirmation | strict | 0.685 → 0.469 | 98.4 → 101.5 |

These are local observations, not a controlled speed comparison or the roadmap
2-vCPU qualification. The original reference runs and initial implementation
measurement ran at different times; final after runs used GOMAXPROCS=4 after the
full test run. Reports retain per-process CPU, RSS, commands, host and hashes.

Run `python3 tools/render.py` and `python3 tools/test_evidence.py` in this review
directory to validate and regenerate the retained evidence without network or
model calls. CLI replay uses tools/measure.py with an explicit binary and a new
output directory. Frozen inputs and labels are checked before execution.

These six pages are now exposed. Further semantic tuning needs a separately
selected, source-only confirmation. #299 remains open for broad recall; #305 and
unsupported #309 constructions remain open. More matching fixtures cannot by
themselves establish the requested product quality.
