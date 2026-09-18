# Clause grammar and attention-frame repairs

The previously exposed boundary sample improves from **2/56 to 7/56** fully
diagnosed defects in both profiles. The changes restore two lost simplicity
judgments and diagnose three other frozen defects, including two Ptah attention
frames. No existing finding or full detection disappears across the 18 sets.
One additional capability rewrite was identified after diagnostics in the older
repetition sample; it earns no frozen-event credit.

The six new pages remain at **2/34**, or 5.9%, in both profiles. Ptah contributes
1/17 and the historical sources 1/17. One more Ptah framing event is partially
diagnosed. These results do not meet the 80% recall or 85% soft-precision targets.
This is a bounded repair, with no measured improvement on the new pages.

## What changed

Unscoped-assurance version 7 accepts a gerund with a particle and object and a
way clause with its own finite predicate. The incomplete-goal check applies at
the boundary before a main subject, rather than rejecting every later article.
Relative restrictions inside instructions and subjectless conjunctions remain
controls. The source span excludes the introductory goal.

Instruction-scaffolding version 10 resolves supports/provides/offers/gives plus
an ability/capability and an optional generic reader, including bounded nested
support chains. Specific actors, permission restrictions, negative clauses and
protected operands remain controls. A capability still does not imply a duty.

Evaluative-closure version 9 recognizes impersonal quality/cognitive notices,
reader-attention selections and whole-point judgments about reading an
explanation. Operational verification, ordering requirements and concrete
component purposes remain controls. The existing announced-importance phrase
rule keeps ownership of its exact sentence openings.

The catalog remains 53 rules, manifest r8. Gate thresholds, weights, models,
public APIs and runtime dependencies are unchanged. Focused public-engine tests
and CLI checks preserve source mapping, revisions and technical controls.

## Frozen evidence and limits

Before runtime changes or confirmation diagnostics, the implementing Codex
assistant reviewed all seven editorial categories and froze 34 defects,
12 uncertain cases and 28 technical controls. This is maintainer-accepted
assistant review under [ADR 0041](../../../docs/adr/0041-assistant-review-acceptance.md),
not independent human annotation or a population estimate.

The selector excluded 142 prior source identities. A source-only audit found no
whole normalized block of at least 12 whitespace tokens in those sources;
shorter or paraphrased overlap is not ruled out. All 148 identities are now
exposed. The historical long selection is a CMake build file classified as text
by source metadata. Its size includes code; this set has no ordinary historical
long-prose page. It remains in the denominator rather than being replaced after
inspection. [Source notes](source-review-notes.md) retain this limitation and
the change from a draft-branch baseline to merged main.

Both phases were remeasured for all 18 sets. The merged baseline produces the
same findings as the prior retained results, but its long-reference candidate
budget abstention is gone after the separate budget repair. Every current run
is complete, with no errors, skipped rules or abstentions. There are no within-cell
bootstrap intervals: each new cohort/length cell contains only one source, and
one of those sources is mixed code and prose.

| Set | Pages | Frozen defects | Before full | After full |
| --- | --- | --- | --- | --- |
| development | 26 | 57 | 15 | 15 |
| exposed_repetition | 6 | 5 | 5 | 5 |
| exposed_framing | 9 | 11 | 3 | 3 |
| exposed_local | 12 | 78 | 11 | 11 |
| exposed_context | 11 | 42 | 9 | 9 |
| exposed_instruction | 12 | 99 | 30 | 30 |
| exposed_construction | 6 | 32 | 8 | 8 |
| exposed_purpose | 6 | 28 | 4 | 4 |
| exposed_verb | 6 | 43 | 3 | 3 |
| exposed_relations | 6 | 45 | 10 | 10 |
| exposed_projection | 6 | 34 | 6 | 6 |
| exposed_scope | 6 | 72 | 7 | 7 |
| exposed_proposition | 6 | 46 | 4 | 4 |
| exposed_clauses | 6 | 58 | 7 | 7 |
| exposed_action | 6 | 57 | 9 | 9 |
| exposed_roles | 6 | 44 | 3 | 3 |
| exposed_boundaries | 6 | 56 | 2 | 7 |
| confirmation | 6 | 34 | 2 | 2 |

The old repetition and framing sets retain narrower targeted review scope;
they do not establish precision for all their diagnostics.
[CHANGES.md](CHANGES.md) records every changed finding, and
[MISSES.md](MISSES.md) records all 32 remaining new-page defects.

## New pages and review burden

| Page | Stored source | Defects | Detected |
| --- | --- | --- | --- |
| c01 | docs/site/src/content/docs/atlas/license-boundary.md | 0 | 0 |
| c02 | docs/site/src/content/docs/testing/migrations-and-schema.mdx | 1 | 0 |
| c03 | docs/site/src/content/docs/versioned/integrity-and-safety.md | 16 | 1 |
| c04 | CMakeLists.txt | 2 | 0 |
| c05 | docs/development.md | 11 | 1 |
| c06 | docs/getting_started.md | 4 | 0 |

Technical emits 39 findings: three actionable diagnoses (including one partial
event), one additional post-diagnostic wording repair, seven uncertain findings
and 28 nonactionable findings. Strict emits 47: the same three actionable, one
additional, seven uncertain and 36 nonactionable findings. The eight extra
punctuation warnings do not diagnose additional frozen defects. The seven
code-spanning CMake length warnings are retained as false alarms, relevant to
[#307](https://github.com/stokaro/unswell/issues/307).

The preexisting instruction warning on the local-editor workflow is a false
positive: its condition selects a real alternative to browser editing. The new
rules do not repair that case. Most remaining defects require fuller rhetorical
or discourse relations; a matching word, length warning or contrast marker
receives no credit for a different diagnosis. No semantic tuning followed
confirmation output review.

## Reproduction

Before runtime: `ad06b689e0da23e6f9e31746553a3d041d02585a`.
After runtime: `f501635ae4c864ad874068c7f0c2dab9caacce97`.
The after binary was built from the clean semantic-freeze commit.
`code-freeze.json` pins runtime and focused test files; `runtime-inputs.tar.gz`
retains those bytes from the recorded revision so later repository changes do
not invalidate historical report verification. Saved reports include
source bytes, identities, finding spans and per-run resource records.

Run `python3 tools/test_evidence.py` and `python3 tools/render.py` from this
study directory. The tests reject missing findings, source drift, invented
credit, unexpected abstention, runtime changes and credit for unreviewed repairs.
To replay a set, build the recorded revision and run:

```sh
python3 tools/measure.py --binary /path/to/unswell --set confirmation --output /new/output
```

The following figures cover the complete CLI scan of each fixed set, with a
fresh process for each profile. They are local measurements, not evidence for
the separate 100,000-word production target. Host identities and binary hashes
are recorded in `reports/*/*/runs.json`.

| Set | Profile | Seconds before → after | Peak MiB before → after |
| --- | --- | --- | --- |
| development | technical | 2.635 → 2.571 | 196.3 → 187.2 |
| development | strict | 1.523 → 1.500 | 177.9 → 198.0 |
| exposed_repetition | technical | 0.532 → 0.543 | 125.3 → 119.6 |
| exposed_repetition | strict | 0.530 → 0.533 | 111.3 → 116.9 |
| exposed_framing | technical | 0.726 → 0.729 | 134.7 → 141.6 |
| exposed_framing | strict | 0.731 → 0.727 | 131.3 → 132.4 |
| exposed_local | technical | 0.617 → 0.618 | 118.8 → 122.6 |
| exposed_local | strict | 0.619 → 0.643 | 118.7 → 122.5 |
| exposed_context | technical | 0.646 → 0.643 | 121.7 → 122.2 |
| exposed_context | strict | 0.650 → 0.649 | 121.5 → 123.9 |
| exposed_instruction | technical | 0.860 → 0.856 | 148.5 → 159.8 |
| exposed_instruction | strict | 0.858 → 0.863 | 159.9 → 169.2 |
| exposed_construction | technical | 0.437 → 0.435 | 98.0 → 105.2 |
| exposed_construction | strict | 0.442 → 0.434 | 100.0 → 104.2 |
| exposed_purpose | technical | 0.438 → 0.439 | 102.4 → 104.3 |
| exposed_purpose | strict | 0.439 → 0.441 | 102.5 → 102.2 |
| exposed_verb | technical | 0.483 → 0.493 | 109.2 → 96.8 |
| exposed_verb | strict | 0.489 → 0.483 | 110.9 → 110.7 |
| exposed_relations | technical | 0.529 → 0.519 | 109.0 → 108.5 |
| exposed_relations | strict | 0.520 → 0.517 | 118.4 → 113.9 |
| exposed_projection | technical | 0.435 → 0.438 | 101.4 → 103.3 |
| exposed_projection | strict | 0.438 → 0.436 | 101.4 → 99.6 |
| exposed_scope | technical | 0.666 → 0.667 | 162.5 → 150.7 |
| exposed_scope | strict | 0.660 → 0.664 | 149.2 → 152.2 |
| exposed_proposition | technical | 0.386 → 0.389 | 82.5 → 89.4 |
| exposed_proposition | strict | 0.386 → 0.387 | 84.7 → 88.0 |
| exposed_clauses | technical | 0.415 → 0.414 | 86.6 → 86.8 |
| exposed_clauses | strict | 0.418 → 0.427 | 86.5 → 84.1 |
| exposed_action | technical | 0.403 → 0.396 | 88.5 → 96.1 |
| exposed_action | strict | 0.389 → 0.394 | 86.3 → 86.5 |
| exposed_roles | technical | 0.449 → 0.445 | 91.5 → 99.5 |
| exposed_roles | strict | 0.442 → 0.439 | 92.9 → 94.0 |
| exposed_boundaries | technical | 0.458 → 0.458 | 107.3 → 106.0 |
| exposed_boundaries | strict | 0.459 → 0.457 | 99.5 → 119.9 |
| confirmation | technical | 0.433 → 0.424 | 102.5 → 108.7 |
| confirmation | strict | 0.432 → 0.423 | 105.1 → 103.4 |
