# Prospective transfer screen for bounded hope framing

Status: frozen before external source search or candidate execution on November
7, 2026. The candidate is commit `35e3350100fa0af260845d9230b14810c9c5fdeb`;
the comparator is main `0b5433f8289d27f97c905590b44acfaa221792c6`.
Do not modify either implementation after observing these sources. This screen
uses no model calls and does not qualify #349's complete-page cohorts.

## Selection

Use GitHub's code search API, with the first page of 30 hits in API order for
each query, run once in the order shown:

1. `"we hope that" "document" extension:md`
2. `"it is hoped that if" extension:md`
3. `"hopefully" "unnecessary" extension:md`

For each query, select the first three distinct eligible repository/files.
Exclude `stokaro/unswell`, any exact repository previously used among the 36
original #349 pages or any earlier confirmation/reserve cohort, duplicate
files, forks, non-English pages, and files outside 1–60 KiB or invalid UTF-8.
An exact Markdown source must pass the current public source-preparation path
with original byte spans. Record every excluded search hit and the reason;
never replace a source for apparent editorial quality or candidate output.
If a query yields fewer than three eligible files, retain the shortfall. Pin
each chosen repository's default-branch commit and file blob, raw SHA-256,
full original bytes in a local ignored archive, and a source URL.

This is a trigger-enriched transfer screen. It estimates neither natural
prevalence nor population precision. Search snippets reveal some source text
to the reviewer, but the candidate was committed and published before these
queries. Record that limitation; do not call the review blind or independent.

## Reference and execution

Before running either binary, inspect each complete chosen page and mark every
`hope` or `hopefully` passage as definite actionable, acceptable unchanged,
or uncertain. Preserve the complete source paragraph, exact UTF-8 byte span,
technical constraints, and rationale. The reviewer is the implementing Codex
assistant under ADR 0041, not a human annotator. Freeze judgments and source
hashes before replay. No uncertain passage may become definite after seeing a
detector output. Record missing sources and parsing failures as incomplete.

Run the pinned comparator and candidate under identical technical and strict
policies on every chosen complete page. Archive full reports and compare every
added, removed, or changed finding by original source span and content. Review
every new finding and its conditional suggestion in complete source context,
including protected identifiers, negation, conditions, and current behavior.
No accepted prior finding or FULL-event witness may disappear.

An operator transfer pass requires at least six definite actionable target
events in at least three repository-disjoint pages, at least four newly FULL
events, no new rejected or uncertain deliveries, no lost accepted witnesses,
and no unsafe or uncertain advice. If density is lower, report insufficient
evidence; if a criterion fails, record a negative result. Any outcome here is
only rule-specific transfer evidence. #349's original 80% FULL-event recall and
85% accepted-delivery requirements on complete whole, context, and historical
cohorts remain unchanged and need separate prospective confirmation.
