# Editorial claim-accounting contract and legacy replay

The existing research consumer now implements an explicit claim-preservation
contract. A grammar edit cannot resolve a separately declared rhetorical criticism
without a claim-specific review of that same edit. Retained comparison prose stays
separate from revision targets. This is representation evidence, not an improvement
in semantic detection recall. [#367](https://github.com/stokaro/unswell/issues/367)
remains open for the complete reviewed migration and downstream use.

The [developer command](../../annotation/internal/claimreview/README.md) uses the
existing extractor and original UTF-8 ranges. [ADR 0042](../../../docs/adr/0042-editorial-claim-accounting.md)
defines the trust boundary, versions, mandatory accounting, and qualification
limits. The normal engine, model-free rules, CLI/MCP, product reports, and gate
policy are unchanged. No external model ran.

## Blackbox evidence

The Go tests exercise the implemented command and typed contract without reading
private internals. Constructed examples cover independent grammar and rhetoric
claims sharing a paragraph, retained antecedent context, repeated quotations,
Unicode with CRLF, protected inline code and fences, explicit same-defect review,
immutable originals, incomplete mandatory accounting, writer errors, and cancellation.

Negative cases reject missing or invented claims, duplicate claims, unknown source
owners, mismatched quotations, split UTF-8 ranges, context relabeled as a target,
unsupported or changed-edit resolution, reused or unused approvals, unreviewed
overlap deduplication, ambiguous JSON keys, unsupported versions, and incomplete
inventories. Complete accounting never sets `editorial_qualified` true.

## Replay of the preserved pool

[replay.py](replay.py) ran all 36 exposed pages and all 804 original candidates
through the Go command locally. It checked the original packet manifest and every
request hash, unchanged source hashes, exact targets, stable original claim IDs,
and retention output. It inferred no independent claim splits, support roles,
semantic equivalence, or new quality labels. In this migration, `retained` means
the original record survived, not that its diagnosis is accepted.

| Outcome | Original candidates |
| --- | --- |
| Bound and retained without changing their original declaration | 414 |
| Invalid: original target intersects protected bytes | 225 |
| Invalid: original rationale or suggestion is absent | 165 |
| Unaccounted or silently discarded | 0 |
| Original denominator | 804 |

The strict contract is therefore **not a completed rollout over the old pool**.
The 225 broad targets need reviewed localization that separates the actual
wording target from protected or retained context. The 165 records need explicit
source-supported explanations; copying an empty field or inserting a generic
rationale would not satisfy that review. They remain invalid in this replay.
The contract was not weakened and no candidate was silently narrowed to pass.

The old source inventory uses a previous extraction result. The replay derives
current block ownership only from the existing Go extractor; it does not add a
Python Markdown parser or substitute random units. All candidate IDs and failures
stay in the original accounting. No duplicate was collapsed automatically.

The [aggregate record](summary.json) retains resource time, command hash, packet
and request identities, and explicit incompleteness. Original sources, candidate
text, judgments, detailed invalid reasons, and per-candidate records stay local.
The preserved denominators are 57 whole-page events, 46 contextual events, and
20 exposed historical events, totaling 123. This replay does not recompute those
semantic metrics or change their labels. Prospective confirmation pages were unused.

## Reproduce the mechanical check

From `research/annotation`, build the command without cgo:

```sh
CGO_ENABLED=0 go build -o /tmp/reviewclaims ./cmd/reviewclaims
go test ./internal/claimreview ./cmd/reviewclaims
```

Then, with the preserved private packet and a new output directory:

```sh
python3 -B research/reviews/2026-09-30-claim-accounting/replay.py \
  --packet /path/to/unswell-contextual-atomic-review-packet-v1 \
  --binary /tmp/reviewclaims --output /path/to/new-local-evidence
```

The Python harness is optional standard-library research tooling. The product
and ordinary Go tests require no Python model environment or network access.
Without the intentionally excluded private packet, public readers can run the
constructed blackbox tests and inspect the aggregate arithmetic, but cannot
independently reproduce the original task-level migration. No stronger
reproducibility or qualification claim is made.
