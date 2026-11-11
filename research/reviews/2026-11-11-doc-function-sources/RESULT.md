# #349 document-function acquisition: negative feasibility screen

Status: terminal offline acquisition result. The predeclared [protocol](PROTOCOL.md)
was applied without model calls, changing #349 labels, or opening calibration or
transfer source content. This is an assistant source-only review under ADR 0041,
not a human annotation, an AI-authorship judgment, or a detector-quality result.

## Pinned selection and source integrity

The four archived GitHub search responses and the [selection](selection.json)
record 32 distinct repositories in eight topic/function cells: database,
security, observability, and testing, each split into guide and
concept/reference cells with four sources apiece. In each cell the first two documents are development,
the third calibration, and the fourth transfer. The complete selection was
sealed before development content was fetched. Sixteen development blobs were
then fetched by pinned Git blob SHA-1, checked for exact size and UTF-8, and
recorded with SHA-256 in [development-sources.json](development-sources.json).
Their full bytes remain local under the Git-ignored
`artifacts/research/2026-11-11-doc-function-sources/sources/` directory.

One selected development document, `database-guide-01.source` from
`oceanbase/miniob`, is a Chinese debugging guide. It is a valid pinned UTF-8
source, but it is outside this English editorial review's language scope. It
remains in the 16-source denominator with no English labels. The metadata-only
selection did not include a language check, and its GitHub repository topic
does not always describe the selected document's subject: for example, the
security cell includes a functional-testing architecture page, and an
observability guide is mainly a Kubernetes deployment walkthrough. Those are
properties of this frozen acquisition, not reasons to replace sources after
reading them.

## Source-only review

The root Codex assistant reviewed the complete English development pages and
recorded exact original UTF-8 byte spans in
[source-only-review.json](source-only-review.json). The review contains **8
clear editorial opportunities across 6 repositories** and **32 close retain
controls**. Four initially plausible candidates were examined against their
full-page context and kept out of the positive count; the review records each
reason. In particular, the full ContextForge performance section also includes
a tool request and metrics inspection, and the CPAMP FAQ explicitly directs
readers to Full Mode for its full feature set. Counting either as a contradiction
would inflate the result.

The positively labeled cases include repeated section announcements, delayed
operational caveats, an internally inconsistent precedence example, a broad
completion claim despite documented next steps, and an unsupported opening
superlative. The controls preserve numbers, negation, versions, conditions,
limits, and technical statements that should remain. The precedence example
requires implementation verification before editing; a diagnosis is not a
license to guess its intended value.

| Requirement | Minimum | Observed |
| --- | ---: | --- |
| Clear opportunities | At least 16 | 8: fail |
| Distinct positive repositories | At least 8 | 6: fail |
| Close controls | At least 32 | 32: pass |
| `retain` controls | At least 8 | 32: pass |

The acquisition gate fails. No algorithm, model prompt, threshold, or
calibration/transfer evaluation was run on this packet. The earlier README
packet is not pooled into these counts. The frozen #349 cohorts and the 80%
FULL-event recall / 85% accepted-finding requirements remain unchanged and
unmet; this packet has no measured product gain.

## Consequence for the next hypothesis

Two independently frozen public-source acquisitions have now failed to supply
the minimum positive/control balance for a clean evidence-binder experiment.
Selecting by README or by repository topic plus documentation path does not
reliably find the contextual editorial cases that #349 is missing. A third
generic scrape would repeat that sampling problem. The next work should use
the already frozen #349 missed-event and archived candidate evidence to test
automatic source-grounded admission and delivery, with a protocol that keeps
development replay distinct from page-disjoint transfer and prospective
confirmation. That is a research direction, not a claim that such a system has
qualified.

## Offline verification

`python3 -B research/reviews/2026-11-11-doc-function-sources/verify_packet.py`
checks the search snapshot hashes, selected ranks and blob identities, source
manifest linkage, exact review provenance, UTF-8 span lengths, and frozen
counts without downloading or opening source content. The local
`materialize_review.py` additionally verifies each development source byte
hash and each unique quote before producing the review. The seven focused
offline tests include negative cases for incomplete search results, invalid
source blobs, changed original spans, and missing or repeated quotes.
