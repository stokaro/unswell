# Standalone source-function pilot on exposed fresh pages

This is an offline, non-generative research candidate. Its design follows the
source-only review of nine pages, so those pages are **development data**, not
transfer or prospective confirmation. The source-only v2 labels and complete
current-main CLI reports are pinned in `baseline-audit.json`. The candidate
code and this protocol must be committed before reading candidate output.
No model service call, training fit, threshold sweep, or runtime rule change is
permitted in this pilot.

The candidate consumes each complete original page and the unchanged CLI's
sentence assessments and excluded byte spans. Those assessments provide the
existing Markdown extraction and source map; the candidate does not parse
Markdown a second time. It considers every eligible sentence on each page.
It first tests for an empty table announcement, an empty ease or limitation
announcement, command narration that duplicates a preceding command, and a
repeated proposition from a previous sentence. It then tests for unsupported
quality ranking or evaluative wording. A contrast with an adjacent concrete
example and an explicit safety, operational, or conditional instruction can
withhold an otherwise matching quality cue. Each output must name the source
function it found, quote original bytes, and state which content must survive
an edit. The code is the exact frozen rule specification; no labels or page
identities are inputs to it.

Emit at most one diagnosis per sentence. Never target an excluded byte span.
Do not emit a concrete replacement unless a deterministic deletion or
substitution preserves the sentence's protected tokens and syntax; otherwise
record `edit_status: unresolved`. Count unresolved edits separately. Any
source mismatch, incomplete report, invalid span, or unexpected page count
stops the run. Outputs are append-only evidence, not product findings.

Review **every** diagnosis on the complete nine pages against the sealed
positives and controls, including diagnoses outside labeled spans. Report
FULL event recall, location-only hits, accepted/delivered, optional,
nonactionable, unresolved edits, abstentions, and protected-region violations
separately. Directional advance requires at least 14/17 FULL positives,
at least 85% accepted delivered diagnoses, no `retain` control hit, and no
unsafe supplied edit. Passing this exposed screen is only permission to
acquire another repository-disjoint source-only transfer packet. A failure
retires this exact candidate without adjusting its rules on these outputs.
In either case, the unchanged 123-event development gates and reserved
prospective confirmation still control product qualification.
