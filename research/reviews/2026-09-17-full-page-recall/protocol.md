# Whole-page editorial recall audit

Protocol: `unswell-full-page-recall-v1`, fixed September 17, 2026.
Reference engine: `cc79256188310f9d63a86f259044df4c3a27af1d`.

## Question

Which actionable wording defects does Unswell miss when a reviewer reads an
entire technical document? Measure defect recall and the usefulness of emitted
diagnostics separately. A length warning that overlaps empty framing does not
detect that framing. A passing gate is not an editorial label.

Use the seven categories and technical exceptions in
[the existing rubric](../../../docs/editorial-annotation.md). Preserve technical
meaning when proposing an edit. Required conditions, repeated identifiers,
necessary navigation, release obligations, and concrete alternatives are not
defects merely because their form recurs. Record uncertain cases separately.

## Fixed sampling

Use the complete 136-page Ptah snapshot from September 16, pinned at
`654eae5591392278e6c8bce8e54737f780766f19`, and the 38 historical documents
in the frozen long-prose-v3 input archive. Preserve complete source bytes,
original formats, rights records, notices, and date evidence.

Select two Ptah pages per format (Markdown/MDX) and length stratum: fewer than
800 prose words, 800–1,999, and at least 2,000. Select four historical documents
per length stratum, processing long, medium, then short and retaining at most
one document per repository. Use SHA-256 of `full-page-recall-v1\n` followed by
the cohort, repository, and original path as the fixed rank. Do not resample
after reading the selected text or detector results. Report any undersupplied
cell explicitly. Lengths are the previous extraction measurements, not outcomes.

Add PostgreSQL and database-URLs as separate, previously exposed anchors if
not already selected. Do not pool their purposively selected results into the
sample's headline estimate. Preserve their full pages, including long sections.

These are samples from two available frames, not population estimates for all
technical writing. Historical documents have dated provenance, not automatic
acceptable labels. Ptah has maintainer-reported repository-level AI provenance;
individual passages retain unknown authorship. The two cohorts differ in topic,
format, length, and selection history; an editorial difference cannot establish
an AI-origin effect.

## Review and freeze

The reviewer is the Codex assistant in this session. No human annotation,
independent second rater, or blinded provenance is claimed. Prior study exposure
and the two known anchors are disclosed. This auxiliary development audit cannot
qualify rules under #26 or replace #22's independent human corpus.

Read every selected source in order, including surrounding code as context.
Record a coverage ledger whose consecutive byte ranges cover the complete file,
with a content-specific rationale for each reviewed section. Code, frontmatter,
and link destinations remain context rather than prose defects. Do not silently
omit a section because it has no findings or is difficult to evaluate.

For each defect record a stable ID, category, exact source-bound target and any
supporting passages, rationale, and a proposed edit or concrete repair action.
Record optional/uncertain alternatives without converting them into positives.
One editorial problem is one event even when it spans several paragraphs.
Save explicit acceptable counterexamples and why their similar wording matters.
Seal the source manifest and completed annotations before opening the new
diagnostic outputs or changing any rule. Preexisting knowledge cannot be undone;
the freeze prevents retrospective edits from following new results.

## Match findings to problems

Run the public CLI on whole sources with explicit technical and strict profiles,
an empty local configuration, and source inclusion. Use the original format;
MDX must not be analyzed as Markdown. Record binary/commit/config identities,
complete status, abstentions, extraction exclusions, and all diagnostics.

Review every emitted finding. Assign `actionable` with the IDs of the defects it
actually describes, `nonactionable` with a technical reason, or `uncertain` with
the unresolved distinction. Matching requires relevant diagnostic semantics and
source overlap with a target/supporting passage. Incidental overlap, a matching
keyword, or warning count alone earns no credit. Preserve duplicate diagnostics
but count a detected defect only once. Record newly discovered defects in a
separate post-freeze addendum; do not silently enlarge the primary denominator.

Report event recall (detected / annotated defects), missed events by category,
actionable/uncertain/nonactionable diagnostic counts, and clean-paragraph alarm
rates where assessment units and annotations allow a denominator. Also report
page-level results, review coverage, extraction gaps, and findings per 1,000
prose words. Keep uncertain and operationally incomplete results visible; never
turn missing data into a clean result. Strict and technical are separate runs.

Provide paired page-bootstrap intervals with a fixed seed for sample estimates,
and a leave-one-page-out sensitivity check. They describe variability across
these reviewed pages, not annotator agreement or population certainty. Ptah is
one repository, so no repository-generalization interval is available.

## Deliverables and acceptance

- Complete pinned sources with notices and a reproducible selection manifest.
- Source-bound annotations, coverage ledger, explicit reviewer identity, and a
  freeze binding the original annotations before detector-output review.
- Whole-page public-engine outputs for both profiles and a complete disposition
  ledger, including false alarms and undetected defects.
- An executable offline evaluator with negative tests for source drift, invalid
  spans, incomplete coverage, missing finding dispositions, duplicate credit,
  semantic mismatch, and missing/failed documents.
- Automatically generated counts, page/category tables, uncertainty/sensitivity,
  and a readable list of missed passages with proposed repairs.
- A diagnosis of the actual detection gaps and prioritized engineering actions,
  tied to measured misses rather than a desired gate outcome.

Do not tune gate thresholds. Any subsequent rule changes must report before/after
on this frozen development audit and obtain a separately sampled confirmation
set. This audit alone is not that confirmation set.

## Review mechanics and scope

The assistant read source-bound prose packets in source order, with byte/line
locations and surrounding source consulted for references and protected regions.
Packets used public `extract.Parse` with structure included, without running NLP
or rules. Repeated table headers were grouped for reading; every original cell
remains in the source. HTML, MDX tags/expressions and the excluded quotation were
inspected separately so extraction exclusions were not automatically labeled clean.
Link destinations were abbreviated in reading packets, then checked in the
original source when they affected a claim's evidence (including etcd's test link).
Code examples, administrative frontmatter and destination URLs are contextual,
not editorial targets. Spelling errors and factual/version drift are outside this
seven-category wording review.

This is a complete-file source audit. It does not expand imported components,
execute examples, read text embedded in images or fetch linked documentation.
The glossary page invokes `GlossaryList`, whose external definitions are not in
its source. The document-export page imports preview images and a generated HTML
sample. Those rendered resources remain outside the denominator; this is not a
claim that the complete rendered website has been annotated.

The historical long stratum contains three changelogs and one release-policy
page. Retain this outcome of fixed sampling and report it; do not substitute
other pages to make the cohort comparison look more balanced.
