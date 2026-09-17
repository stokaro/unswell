# Repeated-claim budget: complete analysis with the same findings

The pinned Ptah migration reference now completes repeated-claim analysis at
the unchanged default budget. The previous runtime abstained on this page.
The optimization combines key construction with source-map validation, removing
a separate quoted serialization pass over each prose token. Private keys use
length prefixes; exact token identity and protected-code distinctions remain.

## Evidence

- All 82 retained pages were replayed in technical and strict modes. Findings,
  source evidence, assessments, scores and gates are identical. Only one old
  repeated-claim abstention per profile disappears, on exposed instruction c05.
- A separate generous-budget control compares every block activation and its
  ownership under both runtimes on all 82 pages in both profiles. Complete
  reports are identical except for the runtime revision. Candidate eligibility
  and zero-versus-unavailable observations are preserved.
- The private accounting probe preserves all 327 eligible claims across 562
  considered sentences and 340 blocks. Sentence-stage work decreases from
  105,599 to 73,372 units. Block traversal and final group emission are additional;
  these units are algorithm accounting, not CPU instruction counts.
- The public-API regression verifies the frozen source SHA-256, no abstention at
  100,000 units, actual measured activation values, and parity with a 1,000,000-unit
  control. Existing tiny-budget and cancellation tests still require failure.
  Mapping tests cover UTF-8, CRLF, entities, emphasis, whitespace, links, case
  and opaque operands.

The optimization adds **zero** editorial detections. The preceding
[construction review](../2026-09-17-construction-recall/README.md) still has
1/32 full detections on its new confirmation pages. This resolves an applicability
failure in #312; it does not satisfy the broad recall objective or justify a
higher score. Profiles, thresholds, rule descriptors and default limits are unchanged.

## Sources and reproducibility

Before: `fcb6cdd8dc69569c726cd6df4dda8ef32345534e`. After: `d0e29d2475833a8198c47f67898761c42dcac083`.
All inputs and their licenses remain in the earlier frozen archives. This is a
behavior-preserving optimization, so no source-only labels were added or changed.
The [protocol](protocol.md) defines the intended applicability difference.

The affected source is the pinned Ptah migration command reference:
`docs/site/src/content/docs/atlas/migrate-commands.md`, snapshot
`654eae5591392278e6c8bce8e54737f780766f19`, 97,753 bytes, SHA-256
`330fb59a5fe90f3659d64d53cff20c403bdfe5fb29325c53c7a6f4d67905ed1a`.
Its licensed bytes are `sources/c05.md` in the instruction review's archive.

Run `python3 tools/budget_parity.py`, `python3 tools/test_evidence.py` and
`python3 tools/render.py` from this directory. They validate all 28 report pairs,
the retained baseline, source hashes, configuration, runtime identities and
exact applicability changes. Negative tests reject changed sources, findings,
gates, candidate observations and hidden or inconsistent abstention records.

Use `tools/measure.py --binary PATH --set NAME --output NEW_DIRECTORY` for a fresh
CLI replay. Add `--activations` for the separately recorded generous-budget
control; that option does not change the product configuration. Each worker
records one CLI child's CPU, wall time and peak RSS. Logs remain beside new outputs.

The optional private accounting probe retains the old implementation as its
reference. Copy `tools/capture.go.txt` to a temporary Go file and run it from the
repository root with the extracted c05 source path; save stdout as a gob file.
Copy `tools/work-probe.go.txt` to a new `builtin/claim_fusion_probe_internal_test.go`,
then run `go test -count=1 ./builtin -run TestClaimFusionProbe -v -args -claim-document /absolute/path/document.gob`.
Remove that temporary test afterward. Its whitebox justification covers private
key equality classes and budget counters, which the public result does not expose.
The gob is local measurement data and is not a product model-pack format.

## Recorded resource observations

| Set | Profile | Pages | Findings | Wall seconds before → after | Peak MiB before → after |
| --- | --- | --- | --- | --- | --- |
| development | technical | 26 | 127 | 1.671 → 2.337 | 189.6 → 185.8 |
| development | strict | 26 | 146 | 1.682 → 1.619 | 186.5 → 202.2 |
| exposed_repetition | technical | 6 | 41 | 0.602 → 0.592 | 116.9 → 115.5 |
| exposed_repetition | strict | 6 | 47 | 0.644 → 0.583 | 114.6 → 114.9 |
| exposed_framing | technical | 9 | 88 | 0.847 → 0.792 | 137.4 → 129.4 |
| exposed_framing | strict | 9 | 114 | 0.827 → 0.800 | 129.7 → 137.5 |
| exposed_local | technical | 12 | 73 | 0.711 → 0.743 | 118.1 → 118.3 |
| exposed_local | strict | 12 | 79 | 0.708 → 0.718 | 118.5 → 122.0 |
| exposed_context | technical | 11 | 39 | 0.727 → 0.690 | 116.5 → 124.2 |
| exposed_context | strict | 11 | 48 | 0.776 → 0.708 | 114.1 → 120.8 |
| exposed_instruction | technical | 12 | 153 | 1.069 → 0.952 | 163.0 → 152.2 |
| exposed_instruction | strict | 12 | 168 | 0.994 → 0.941 | 146.9 → 163.8 |
| confirmation | technical | 6 | 57 | 1.106 → 0.465 | 105.1 → 102.5 |
| confirmation | strict | 6 | 61 | 0.469 → 0.453 | 101.5 → 104.0 |

The before records come from the preceding review; the after records were taken
during local development on the same host. These are reproducibility records,
not a controlled speed comparison or the roadmap's 2-vCPU performance qualification.
Full commands, binary hashes and observations are retained in each runs.json.
