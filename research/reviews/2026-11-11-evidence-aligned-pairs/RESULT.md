# #349 evidence-aligned source-pair acquisition result

Status: **underpowered offline acquisition; no candidate run or product lift**.

The source search and selection were frozen before any selected README content
was opened. All 24 repositories were selected from 41 metadata candidates. The
12 development README blobs were fetched by pinned Git blob identity; each
matched its recorded size and SHA-1, decoded as UTF-8, and received a SHA-256.
The six calibration and six transfer sources remain unopened.

The root Codex assistant read all 12 development pages and recorded nine clear
source-only editorial opportunities across seven repositories, with 13 paired
`retain` controls. Exact original UTF-8 spans and the source evidence spans are
in [source-only-review.json](source-only-review.json). The review is an exposed
assistant judgment under ADR 0041, not an independent human annotation.
Unlabeled passages are unknown, not clean negatives.

| Frozen screen | Required | Observed | Result |
| --- | ---: | ---: | --- |
| Clear positive opportunities | 12 | 9 | Fail |
| Repositories with a clear positive | 6 | 7 | Pass |
| Close controls | 24 | 13 | Fail |
| `retain` controls | 6 | 13 | Pass |

The packet therefore stops at the predeclared acquisition gate. It does not
test an evidence-aligned detector, justify any cutoff, change the original
#349 labels or 57/46/20 denominators, or meet the 80% FULL-event recall and
85% accepted-finding gates. No model calls were made.

## Source and review integrity

- Selection SHA-256: `9795837fd6eab4cb27a66ed2bbd730cad278a728c033b395d0cf65a15f05beda`.
- Development source manifest SHA-256: `f10def35de225cd5cf18992f33a34a03522d98032b26547929823f3b1d9fb3f2`.
- Source-only review specification SHA-256: `04c389dd7b517d8eb540939bc427d8daf48d249b6968662a816d2338e9fdc19b`.
- Materialized review SHA-256: `47352ec1205f7e35d4c8b500e22d5f5e38d2b57a6aee73970be04b3a4dadc777`.
- The source bytes are local under
  `artifacts/research/2026-11-11-evidence-aligned-pairs/sources/`, excluded
  from Git. The public packet includes pinned source identities, hashes, and
  short review excerpts. `materialize_review.py` verifies the local hashes and
  resolves every excerpt to one exact byte span.

## What this changes about the next attempt

The rank-based GitHub `developer-tools` search heavily favored agent-tool
READMEs: eight of the 12 development pages describe agents or their workspaces.
Several are polished capability summaries with important scope limitations.
The source-only review found too few clear editorial defects to evaluate an
automatic evidence binder. A future acquisition should pre-register a broader
mix of document functions and topics using metadata alone, then freeze source
identities before reading content. It must keep calibration and transfer pages
separate, include close controls, and still test automatic candidate discovery
and the full delivered stream against the original #349 gates.
