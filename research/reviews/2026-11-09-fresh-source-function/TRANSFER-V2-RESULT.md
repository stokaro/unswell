# Source-function v2 transfer: failed

The source-only labels were sealed in `51b2907` before any v2 output: 12 clear
editorial opportunities across six repositories and 35 close controls across
all 12 pinned README pages. Every label is bound to original UTF-8 byte offsets
and an exact source SHA-256. The selected pages and content were frozen in
`07cc47b` and `265a805`; the original files and complete unchanged-CLI
reports are preserved under the ignored local `artifacts/research/2026-11-09-
fresh-source-function/` directory. The reviewer is the root Codex assistant,
not a human annotator. The packet is cue-enriched, not a prevalence sample.

The main CLI completed 11 of 12 pages and produced 26 findings. Page 3 had a
valid pinned source but an incomplete CLI report: `parse markdown: markdown
grammar returned an incomplete or invalid syntax tree`. The report has zero
documents and two recorded errors. An amendment in `0ee2469` counted this as
an operational error before any v2 candidate output was viewed; it changed
only report-error handling, not the frozen diagnosis core or edit planner.
Because that amendment followed a baseline observation, this is exploratory
evidence, not a clean preregistered transfer pass. The page remains in the
12-page denominator.

The candidate was then run once on the selected pages and emitted 22 diagnoses
on the 11 processable pages. `transfer-v2-candidate-output.json` preserves
the exact output. `transfer-v2-audit.json` records a disposition for **every**
candidate diagnosis and **every** unchanged-CLI finding, with source binding
and reason. All 22 diagnosis spans map to exact original source bytes, and no
diagnosis or proposed edit intersects a span excluded by the current CLI.

| Measure | Result |
| --- | ---: |
| Selected pinned pages | 12/12 |
| Complete CLI pages | 11/12 |
| Candidate FULL positives | 2/12 (16.7%) |
| Candidate partial positives | 2/12 |
| Candidate missed positives | 8/12 |
| Candidate accepted delivered | 2/22 (9.1%) |
| Candidate optional / partial / rejected | 7 / 2 / 11 |
| `retain` control hits | 3 |
| Candidate edits safe as supplied | 0/2 |
| Candidate edits unsafe as supplied | 2/2 |
| Candidate diagnoses without supplied edit | 20/22 |
| Unchanged-CLI accepted / optional / rejected | 21 / 2 / 3 |
| Combined accepted delivered | 23/48 (47.9%) |

The two FULL hits are unsupported `next generation` labels in the WPF Toolkit
and Singularity introductions. The candidate partially identifies the
repeated CoAI paragraph and the `most advanced` SkyReels rank but does not
resolve either complete burden. It misses the unsupported `number one` rank,
Oban's duplicated database instruction, Singularity's empty answer, both Vue
Formulate superlatives, SkyReels' `groundbreaking` label, and both empty
conclusion lead-ins. The unchanged CLI adds no FULL hits on these 12 labels.

False positives are decisive: `next generation` names the actual cMeta
successor on the MLCommons page, twice; `Rich Model Support` is followed by a
concrete provider list. The candidate also calls a SkyReels checklist item
redundant when the later occurrence serves a different document function.
Its two proposed deletions are unsafe: one erases a failed A800 benchmark
table cell, and the other erases the sentence defining that table's hardware,
resolution, duration, and compared implementations. The table cell is not
protected by the current CLI's `excluded` spans, so a no-excluded-overlap
check alone did not prevent this error.

**Retire this exact v2 candidate.** It fails all directional thresholds:
80% FULL recall, 85% accepted delivery, no retained-control hit, no unsafe
edit, and complete operational processing. Do not tune it against these now
exposed pages and call the same pages confirmation. The next hypothesis must
represent document function and semantic scope, distinguish a named successor
from promotional generation language, and treat table cells and table-scoping
sentences as protected evidence. The Markdown parse failure needs separate
reproduction and repair before a new page-disjoint transfer. The frozen
123-event development cohort, complete combined stream, and reserved
prospective confirmation remain unqualified. No product rule or runtime API
changed, and #349 remains open.
