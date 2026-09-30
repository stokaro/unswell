# Original rule evidence and an explicit child replay

The Go adapter `research/annotation/cmd/explainrules` repairs the old packet
adapter's loss of measurable rule evidence. It does not alter the frozen
804-candidate packet or qualify a detector. Related work: [#367](https://github.com/stokaro/unswell/issues/367)
and [#349](https://github.com/stokaro/unswell/issues/349).

The original baseline was rebuilt with `CGO_ENABLED=0` at
`e146774df81ee8c106ece4abc4017d41c23a46e0` and run in technical mode on all
36 exact original documents. Every source hash, finding ID, rule category,
diagnostic, primary/related range, and quoted target matches the 165 original
rule candidates. There are no missing, additional, or changed findings.
The [reconstruction record](reconstruction.json) identifies the new report.
The missing original gzipped report was not recovered byte for byte.

All 165 original rationales were empty; 138 suggestions were empty. The new
projection preserves the actual metrics and supplies their observable basis
without asserting that a rule activation proves an editorial defect. Missing
suggestions use the engine diagnostic as review guidance, with an explicit
basis. The 27 existing suggestions remain unchanged. All observations are
`unreviewed`, and the caller must assess necessity, claim boundaries, and roles.

The [child replay](summary.json) accounts for all 165 rule candidates:

| Outcome | Count |
| --- | ---: |
| Original observations preserved | 165 |
| Bind and retain through the unchanged accounting stage | 92 |
| Original target intersects protected source | 73 |
| Unaccounted or silently discarded | 0 |

Retained records establish representation preservation, not new accepted labels.
All original targets and quotations remain unchanged. The protected-range
failures identify remaining localization work; no automatic truncation, claim
splitting, support-role inference, or semantic deduplication makes them pass.
The 639 model candidates are outside this rule-only child replay. Their earlier
record and the complete frozen pool remain intact.

No full-event metrics are recomputed, no model is called, and the original
57 whole-page, 46 contextual, and 20 historical exposed events stay fixed.
The prospective confirmation reserve is unused. Detailed sources, proposals,
and review annotations remain local; only this aggregate record is published.

## Reproduce the local migration

Build the pinned baseline, scan exact source bytes with the same technical
policy, and first verify its 165 findings against the original packet.
From `research/annotation`, build both consumers:

```sh
CGO_ENABLED=0 go build -trimpath -o /tmp/explainrules ./cmd/explainrules
CGO_ENABLED=0 go build -trimpath -o /tmp/reviewclaims ./cmd/reviewclaims
/tmp/explainrules < /path/to/reconstructed-report.json > /path/to/observations.json
```

Then run this directory's `replay.py` with explicit `--packet`, `--report`,
`--observations`, `--binary`, and a new `--output` directory. The script verifies
the frozen requests and every observation's source, locations, identity,
evidence, diagnostic, and action basis before using the existing Go accounting
consumer. It writes a separate child record and does not modify the originals.
It trusts the supplied reconstructed report's provenance; it does not recompute
rule formulas or authenticate the caller. Full replay requires the preserved
local packet, which is not part of the public artifact.

The Go blackbox tests cover observed quantities, explicit and missing guidance,
unavailable evidence, invalid numbers, complete-report handling, source snippet
redaction, cancellation, and writer failures. Run the optional research-script
negative checks with `python3 -B -m unittest test_replay.py` in this directory.
