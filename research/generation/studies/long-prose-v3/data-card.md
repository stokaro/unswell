# Data card: complete technical documents

## Purpose and unit

Version 3 compares recognizable constructions in complete technical documents
and model responses to the same tasks. A document is an observation; the global
provenance component is the resampling unit. Paragraphs and sentences overlap and
do not create additional independent samples. This data set carries no editorial
quality labels and does not qualify an authorship classifier.

The historical layer contains 38 files and 58,137 extracted prose words from 20
repositories: 17 documentation files, 14 READMEs, and seven release-note files.
The declared range is 400–6,000 prose words with at least three paragraphs.
Files retain headings, code, tables, and neighboring paragraphs. Source formatting
therefore affects how much contiguous prose a rule can inspect.

## Selection and provenance

The fixed source frame names 31 public repositories absent from the earlier
study. Every admitted snapshot predates January 1, 2021 and has date corroboration
from a release or package-registry record. A timestamp establishes neither human
authorship of every passage nor clean editorial quality. Historical origin stays
`unknown`; editing it preserves that label with a generation record. Generation
from facts has known experimental provenance.

The selection audit retains every repository's admission or exclusion reason and
the complete eligible pool. It favors files of at most 1,600 words, then a fixed
hash order, with one file per repository and role. It selects all 20 qualifying
repositories; neither rule findings nor model outcomes enter selection. Final
source hashes, source ranges, licenses, notices, and global groups are frozen.

Permission records are tied to each source snapshot. Included code and document
excerpts retain their original notices. Generated files inherit the recorded
permissions and attribution obligations of their factual source. The study
contains public technical material; it includes no private Ptah content or router
credential. Full source bytes and generated output are deliberately published for
reproduction, unlike ordinary Unswell reports that omit source text by default.

## Tasks and generation

An agent curated factual summaries with source ranges and exact technical
excerpts. The summaries use whole-document evidence ranges and do not claim
sentence-level factual alignment. Their seals validate bytes and identities,
not semantic completeness. They preserve version boundaries, commands, numbers,
conditions, negations, and attribution material. This is declared agent curation,
not independent human annotation.

Each task has four requests: generation from facts or editing the full source,
under neutral or plain instructions. The requested word count is the original's
extracted prose length. Model output is imported as Markdown. Differences between
original Markdown/plain text and generated Markdown, copied technical excerpts,
or incomplete adherence to the brief remain possible confounders.

The three selected models are `gpt-5.6-luna` through Codex at low effort,
`claude-haiku-4-5-20251001` through Claude at low effort, and
`HivenetQuant/Qwen3.6-35B-A3B` through the authorized router with thinking disabled.
The router uses temperature 0.7, top-p 0.9, and a 16,384-token output limit.
CLI decoding settings that are not exposed remain unavailable. Codex records the
requested model; its reply does not independently identify the served model.
Claude's `modelUsage` and the router's response identify their served models.

Visible system and user inputs, raw final outputs, request identifiers, reported
tokens, durations, attempts, and statuses are retained. CLI built-in context may
remain despite project-instruction and tool restrictions; these records are not
a claim that the three harnesses have identical hidden instructions. Two requests
per family may run concurrently. A transport failure gets at most one identical
retry. Length deviations, truncations, refusals, and unwanted wording get none.

The final four Haiku requests used a declared dispatch amendment. At 1,962,742
reported tokens, the runner's reservation for a call and a possible retry blocked
further admission. The dispatch ceiling rose from 2,000,000 to 2,150,000 tokens;
the fixed 456-request list and all experimental settings remained unchanged.
`generation-completion.json` records the time, request IDs, prior data exposure,
and wrapper hash. Historical, Qwen, and Luna construction results had been seen;
the joint primary estimates and decisions had not been computed. The original
freeze is retained, and the result report states the actual token total.

## Measurement and allowed claims

The existing Go engine performs extraction, NLP, rule evaluation, and activation
collection under one frozen policy. Python utilities only deliver authorized
research requests, marshal records, and aggregate saved results. They are not
dependencies of ordinary builds, tests, CLI, or MCP operation.

The four primary constructions and twelve family comparisons were fixed before
generation. The Go paired analysis uses global components and reports uncertainty;
the decision utility applies the frozen Holm correction, minimum effect, support,
and review-load requirements. Remaining rules are descriptive. Nothing joins the
confirmatory list after inspection.

Document-level rule abstention excludes that pair from the rule's estimate. Block
activation features separately record applicability and absence reasons. A positive
block activation means contributing evidence, not a second diagnostic. Sentence
and paragraph tables bind diagnostics from the complete-document scan; they do not
claim that the rule was rerun independently on each sentence.

Reported structural windows are whitespace-adjacent paragraph runs and potential
sentence pairs. They describe opportunities, not the matcher's rule-specific event
clusters. Findings whose evidence crosses blocks provide a separate direct measure
of context use. Headings and adjacent heading/paragraph pairs are reported apart.

## Limitations and later use

This is a fixed, small, open-source technical sample. It does not represent every
genre, language background, current model, prompt, or editing workflow. The 20-group
floor is not a guarantee of power. Role and realized-length slices can be smaller
still. Historical warning prevalence is review load, not a false-positive rate.
Unknown shared authors or templates can leave dependence outside the recorded
groups. Public source files may also have appeared in a model's training data.
The saved 8-gram overlap flag cannot separate memorization from retained technical
excerpts or deliberate preservation during editing. Fresh repositories for this
study do not imply unseen pretraining material for the models.
No independently verified author-language-background metadata is available for
this source frame, so no native/non-native subgroup error comparison is claimed.

The frozen protocol, sources, and outputs are now exposed evaluation data. They
must not be relabeled as an unused holdout for later feature selection or human
quality validation. Human assessment remains the separate deferred work in #22
and #26. Agent-written rewrite fixtures prove expected construction changes while
preserving named facts; they do not establish reader preference or accuracy.

Saved data reproduces the analysis without any model access. Regenerating the
responses may produce different text even with the same visible settings. The
archive is a reproducible record of this run, not a promise of deterministic
external generation or universal detection of AI-written prose.
