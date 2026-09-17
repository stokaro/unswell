#!/usr/bin/env python3
"""Render the validated budget and parity results."""
import sys
sys.dont_write_bytecode=True
import budget_parity as m


def main():
    result=m.evaluate();m.audit.write(m.ROOT/'summary.json',result)
    rows=['| Set | Profile | Pages | Findings | Wall seconds before → after | Peak MiB before → after |',
          '| --- | --- | --- | --- | --- | --- |']
    for row in result['rows']:
        if row['kind']!='reports':continue
        b,a=row['costs']['before'],row['costs']['after']
        rows.append(f"| {row['set']} | {row['profile']} | {row['pages']} | {row['findings']} | {b['wall_seconds']:.3f} → {a['wall_seconds']:.3f} | {b['max_rss_bytes']/2**20:.1f} → {a['max_rss_bytes']/2**20:.1f} |")
    table='\n'.join(rows)
    engines=result['engines']
    text=f'''# Repeated-claim budget: complete analysis with the same findings

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

Before: `{engines['before']}`. After: `{engines['after']}`.
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

{table}

The before records come from the preceding review; the after records were taken
during local development on the same host. These are reproducibility records,
not a controlled speed comparison or the roadmap's 2-vCPU performance qualification.
Full commands, binary hashes and observations are retained in each runs.json.
'''
    (m.ROOT/'README.md').write_text(text)


if __name__=='__main__':main()
