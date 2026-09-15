# Longer technical prose, version 3

Identifier: `unswell-llm-patterns-v3`. State: frozen before generation on September 15, 2026.
The freeze record contains the task, brief, request, policy, source, and grouping
digests. No new model output or confirmatory rule measurement preceded it.

## Question and prior exposure

This study asks whether named constructions recur more often in generated or
edited technical pages than in their dated originals, and whether the resulting
diagnostics give a reviewer a useful location to inspect. It does not classify
authorship or assign human quality labels.

Version 2 is a completed, underpowered look: its final arm had 16 global
components, below its frozen minimum of 20. Its sources, hypotheses, outcomes,
and stopping rule remain historical evidence. Version 3 uses fresh repositories;
it does not enlarge or reopen that test set.

Development evidence includes the earlier short-comment experiments, the exposed
Ptah documentation, and the 27-page historical development comparison recorded in
`docs/research/contextual-prose.md`. The engine discarded sentences containing
inline code; the repair restores surrounding prose without reading protected
tokens. The contrast matcher also recognizes repeated `X, not Y` constructions.
These decisions precede version 3 outputs. Qwen outputs from earlier studies were
seen in aggregate; Qwen does not participate in version 3 feature selection. This
is transfer to new tasks with a previously studied family, not a never-seen-family
claim.

The old comment-only fact-sheet extractor cannot represent whole pages. Version
3 adds source-bound factual briefs and explicit document tasks to the existing
generation package. The same acquisition, global plan, extraction, measurement,
paired comparison, and report contracts remain in use.

## Sources and selection

The dated source frame has 31 named repositories absent from the previous corpus.
The original frame, failed acquisitions, corrected Mermaid tag spelling, and
package-registry date corroboration are retained. Inventory and extraction may
run before the freeze; rule outcomes may not. Historical inclusion requires a
snapshot dated no later than December 31, 2020 and corroboration by an independent
publication record. A date is evidence about bytes, not a per-passage human label.

Use complete permitted English Markdown or plain-text files with 400 through
6,000 extracted prose words and at least three paragraphs. Roles are README,
documentation, and release notes. Keep source headings, code, tables, and adjacent
paragraphs intact. Do not concatenate comments. Exclude translation filenames,
non-English pages, generated text, empty prose, parse failures, and unsupported
formats by recorded source criteria. Exclusions must not depend on findings.

Within each repository and role, prefer files of at most 1,600 prose words, then
order by SHA-256 of `unswell-long-prose-v3`, NUL, and source ID. Select at most one
file per repository and role, with a total cap of 66 tasks. Keep the first
eligible file under this order; record any factual-curation exclusion before
generation. Include every eligible repository in the fixed frame, including
reserves. Do not change the seed or replenish the sample after any output.

The global plan keeps shared repositories, originals, derivatives, exact copies,
and known template relationships together. Check against the old source frame.
Pin the new groups to `final_test` before generation. The final task set and each
measured arm must contain at least 20 independent components. A shortfall leaves
that comparison incomplete; it does not lower the minimum or trigger more draws.

An agent curates factual notes with byte ranges and source hashes. Notes retain
conditions, negations, identifiers, numbers, version boundaries, and technical
obligations. They do not reproduce the original's paragraph wording or insert
target constructions. Agent curation is declared; it is not independent human
annotation. The complete original is supplied only to the editing operation. Curated
summary notes use the whole document as their evidence range; they do not claim
word-level factual alignment. Exact code examples, option tables, API signatures,
and attribution rosters accompany the summaries with their original byte ranges.
Those excerpts may include code comments; they are retained technical material,
not an independently rewritten part of the brief. Ordinary indented prose is not
substituted for factual curation.

## Operations and resources

Cross every task with `generate` and `polish`, under the existing frozen
`neutral-v1` and `plain-v1` prompts. Requested words equal the original's extracted
prose word count. Neutral is primary; plain is a prespecified sensitivity condition.
Both operations must receive enough factual material to preserve the document's
purpose. Requests and the final factual briefs are frozen before delivery.

Use the already authorized resources: the weakest available Codex model at its
lowest supported effort, Claude Haiku at its lowest supported effort, and the
smaller configured Qwen endpoint with thinking disabled. Record the actual served
model, visible settings, tool restrictions, full messages, raw output, status,
request ID, duration, and token usage where available. Subscription harnesses and
the authorized router are the only resources; this protocol authorizes no new
paid service or credential permission. Unavailable resources stay untested.

The admitted set is 38 tasks in 20 repositories, totaling 58,137 prose words.
Caps: 38 tasks, three families, four arms, 456 primary requests, two concurrent
requests per family, 16,384 output tokens per router request, 12 hours of elapsed
generation, and two million reported output tokens in total. A transport failure
may be retried once with the identical request; retain both attempts. A refusal,
truncation, length deviation, or unwanted style is never retried. Stop on a usage
limit or exhausted cap. Record partial coverage; do not substitute a stronger
model or silently mix models in one run.

The finite source frame and resource cap determine the attainable precision.
Twenty groups is an admission floor, not a power guarantee. Report interval width
and unresolved effects explicitly. The three-point MID is retained to prevent
interpreting a small positive estimate as useful; wide intervals cannot establish
equivalence to zero. Additional paragraphs do not count as independent groups.

## Frozen comparisons

The primary unit is the complete document. Compare each response with its own
original, separately by family and operation. The four named constructions are:

1. `syntax.paired-contrast-density`: repeated contrast templates in a local window.
2. `syntax.passive-candidate-density`: surface passive candidates in a window.
3. `syntax.noun-stack`: a replication of the earlier candidate in a new role.
4. `readability.grade-metric`: a replication and a length-sensitive control.

Only the first two test the newly exposed contextual opportunity. The latter two
retain the earlier hypotheses' identities but do not claim to repeat the same
short-comment estimand. Other rules receive descriptive tables and dispositions;
they cannot join the confirmatory list after results are read.

Primary family-by-construction tests use `generate/neutral`: 12 comparisons with
Holm correction at 0.05, a three-percentage-point minimum paired increase, at least
20 measured components, and at least five response-support components. Use the
existing joint component bootstrap with 10,000 replicates, PCG seed 17, percentile
95% intervals, and its documented two-sided tail convention. No interval or zero
support is a reason for an inconclusive result, not evidence of absence. A claim
of cross-family association needs support in at least two families and a positive
point estimate in the third. Editing and plain-prompt results are sensitivities,
with the same intervals and support counts, not additional primary discoveries.

The prespecified advisory-load budget is at most one finding per 1,000 historical
prose words and findings on at most 25% of historical documents per construction.
This is a review-workload choice, not a false-positive target. A construction that
exceeds either budget remains restricted even when its association test passes.
The budget does not disable existing general readability rules or qualify a
blocking default. Publish the two loads regardless of the hypothesis outcome.

Publish total and applicable pairs separately. A rule abstention on either side
removes that pair from that rule's estimand; it never counts as zero. Retain
failures, truncations, and unmatched sources in coverage. Show all-output results
at the requested budget and a separate realized-length comparison, without
discarding deviations from the primary table.

## Context, controls, and interpretation

Report findings per document, per 1,000 prose words, and per applicable sentence
and paragraph. Stratify by role, operation, family, and realized length bands:
under 400, 400–799, 800–1,599, 1,600–3,199, and at least 3,200 words. Count paragraph,
sentence, heading, and contiguous-window opportunities. Small role slices remain
descriptive; the global group minimum is not claimed for every slice.

Run the same frozen documents with all 40 rules, then derive phrase, structural,
and contextual ablations from retained rule identities. Contrast full-document
window counts with unit-local counts and the pre-repair engine on the exposed
development sample. Retain the earlier temporal placebo and earlier-boundary
evidence under its original policy; do not mix its counts with version 3 outcomes.

Inspect examples after measurement: exact constructions, necessary technical
contrasts, code boundaries, missed opportunities, and meaning-preserving rewrites.
Select examples by stable source order within each outcome category, not by the
best-looking score reduction. Preserve a compact offline fixture set. A rewrite
must retain negation, quantities, conditions, and identifiers. Agent review can
establish a regression example, but cannot qualify a precision claim or replace
the deferred human annotation of #22/#26.

Report combined-profile document prevalence and warning load alongside the
individual rules, with both document-micro and repository-macro aggregation.
No numeric historical-warning cutoff qualifies a default in this experiment:
these originals have no independent quality labels. Promotion to a blocking
default requires the deferred human qualification; even a significant association
with low historical prevalence is insufficient.

Default contrast diagnostics remain advisory and unweighted. A positive frequency
result cannot turn every contrast or passive construction into a defect. Frozen
reports, model outputs, manifests, hashes, failures, reproduction commands, and
per-candidate dispositions complete the research record even when results are
negative or inconclusive. Failure to execute an admitted comparison remains an
explicit incomplete item, not a successful research outcome.

## Execution amendment: dispatch reservations

After 148 Haiku responses, the three models had reported 1,962,742 tokens.
The delivery script still reserved 65,536 tokens for each new call, including
a possible retry. That reservation prevented the final four fixed requests.
The dispatch ceiling was raised to 2,150,000 tokens to complete those requests.
The original two-million-token plan remains in the frozen input archive.

This administrative change occurred after the historical measurements and the
Qwen and Luna construction results were inspected. Haiku usage and delivery
statuses had been inspected; the joint primary estimates and Holm decisions
had not been computed. The sample, visible inputs, models, efforts, retry
eligibility, deadline, hypotheses, exclusions, and statistical criteria stayed
unchanged. No new task or response replacement was admitted.

The [study record](../generation/studies/long-prose-v3/README.md) publishes
`generation-completion.json` and the completion wrapper with its checksum.
They record the time, exact remaining request IDs, prior exposure, and changed
dispatch setting. Treat this as a disclosed execution amendment, not an
unamended execution of the original resource plan. The results report the
actual token total separately from either reservation ceiling.
