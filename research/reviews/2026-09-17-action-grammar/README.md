# Action grammar: bounded gains, broad recall still low

Full frozen-event detection on the six selected pages rises from **2/57 to 6/57**
in both profiles. All four gains are on Libevent 2.0, a page with a known shared
paragraph in previously reviewed Libevent 2.1 notes. On the other five pages,
recall remains **2/24**. Neither result meets the 80% recall objective.

The 16 added findings on previously exposed pages provide six full detections,
one partial detection, seven additional wording repairs, one unresolved judgment
and one false positive. The new engine retains every prior finding and full detection.
The prior Curl referrer and speed-condition losses are recovered.
These development gains do not establish performance on unseen prose.

## Runtime change

`filler.instruction-scaffolding` version 7 uses explicit infinitive and imperative
roles for reader goals, including coordinated actions with a shared object.
The narrower verb vocabulary for impersonal possibility statements is unchanged.
An adjacent nominal antecedent can carry a relative support chain; bare reader
capabilities still need another support layer or a related method.

Copular capable-of gerunds, imperative be-sure-to instructions and purpose-to-steps
announcements also qualify. A later used-operand layer can supply the missing
support relation. An immediate by-using method needs an explicit reader action
and a repeated two-noun object to link it to the preceding capability.

The matcher uses bounded surface roles. It does not establish dependency edges
or semantic entailment.
Keep prerequisites, actors, optionality and operational conditions in revisions.
The checksum false positive below demonstrates the limit: a nominal antecedent
can be an operand for an operation, not its actor. Permission, negation,
attribution, protected vocabulary and structural-boundary controls remain tested.
No threshold, weight, gate, model, class or catalog size changed: 53 rules, r8.

## Frozen review and overlap

Protocol, selector, six sources and notices were frozen before source reading.
The implementing assistant reviewed 554 extracted blocks and table cells, then
froze 57 defects, nine uncertainties and 42 controls at `2026-09-17T20:45:50.734983+00:00`.
All seven rubric categories were considered. The review follows
[ADR 0041](../../../docs/adr/0041-assistant-review-acceptance.md); it is assistant
development evidence, not human annotation or independent qualification.

The selector excluded 124 prior source identities and hashes. Source-only review
then found that c04 reused the incompleteness notice from an earlier Libevent
file. The assistant froze the [amendment](source-exposure-amendment.md) before changing
the engine or reading diagnostics. It retains all six pages and their full denominator,
while separating c04 from the other five. No page was replaced.

The exact-block audit compared extracted blocks of at least 12 whitespace tokens
with all prior sources after whitespace normalization. Its one match is retained
in [source-overlap.json](confirmation/source-overlap.json). Shorter fragments,
partial overlap and paraphrases can escape that check. The other five pages have
no known overlap; that is not proof of independent provenance. All 130 selected
source identities are now exposed. One page per cohort/length cell supports
observed counts, not population recall or a meaningful within-cell bootstrap.
No matcher or frozen label changed after confirmation output was opened.

| Set | Pages | Frozen defects | Before full | After full |
| --- | --- | --- | --- | --- |
| development | 26 | 57 | 14 | 14 |
| exposed_repetition | 6 | 5 | 5 | 5 |
| exposed_framing | 9 | 11 | 1 | 1 |
| exposed_local | 12 | 78 | 11 | 11 |
| exposed_context | 11 | 42 | 8 | 8 |
| exposed_instruction | 12 | 99 | 29 | 30 |
| exposed_construction | 6 | 32 | 7 | 7 |
| exposed_purpose | 6 | 28 | 4 | 4 |
| exposed_verb | 6 | 43 | 2 | 2 |
| exposed_relations | 6 | 45 | 6 | 8 |
| exposed_projection | 6 | 34 | 6 | 6 |
| exposed_scope | 6 | 72 | 7 | 7 |
| exposed_proposition | 6 | 46 | 4 | 4 |
| exposed_clauses | 6 | 58 | 4 | 7 |
| confirmation | 6 | 57 | 2 | 6 |

Two older studies limit review to their original categories. Their counts do
not support precision claims about all warnings. [CHANGES.md](CHANGES.md) reviews
every changed finding, including new warnings outside those older scopes.
The exposed annotation-information defect receives only partial credit: detecting
the support chain does not also identify its repeated information wording.

## Complete selected pages

| Page | Original path | Defects | Detected |
| --- | --- | --- | --- |
| c01 | docs/site/src/content/docs/reference/mcp-tools.md | 0 | 0 |
| c02 | docs/site/src/content/docs/inference/concepts/verification-and-cutover.md | 13 | 2 |
| c03 | docs/site/src/content/docs/operate/oci-registry.md | 9 | 0 |
| c04 | whatsnew-2.0.txt | 33 | 4 |
| c05 | docs/querying/functions.md | 2 | 0 |
| c06 | contrib/raftexample/README.md | 0 | 0 |

Ptah remains 2/22 and the historical cohort rises from 0/35 to 4/35. With c04
excluded, the remaining historical pages stay 0/2, and the five-page total stays
2/24. Two pages have no definite defects: the MCP tool reference and raft example.
Their repeated API and procedure structure is useful, not automatically defective.

Technical findings rise from 31 to 34; strict rises from 33 to 36. Five findings
cover the six full events because one diagnostic relates both SSL alternatives.
There is one additional post-diagnostic repair and nine unresolved judgments.
The other 19 technical and 21 strict findings are nonactionable in this review.
The frozen-label actionable fraction is **5/34** for technical and **5/36** for
strict; including the additional repair gives **6/34** and **6/36**. The 85%
soft-precision objective is not met. All six pages pass both gates.

The [miss ledger](CONFIRMATION-MISSES.md) retains all 51 remaining defects.
They include unsupported benefit claims, indirect paraphrases, reader judgments,
page narration and repetition that a local action grammar does not resolve.
The shared Libevent notice itself remains missed. Detections elsewhere in that
file do not make it an unexposed page.

Two false positives need fixes. The existing evaluative-closure rule reads
`row counts` as a claim of value, although it names a measured field. The new
relative-support path mistakes a checksum's role in caching for an indirect
instruction. Review found both after the semantic freeze. They remain in the
measured result, with no later repair counted as a confirmation gain.

## Reproduction and limits

Before runtime: `4bac934fd7085d5a53f15ad2a71a8f77bd493649`. After runtime: `0e0941ae26dcfa89a4b5d92fcb3080cf889fc596`.
The build used the clean semantic-freeze checkout. The before reports for exposed
pages come from the prior iteration. The new engine replayed all 15 sets. The six-page before run uses the pinned preceding binary. Source
bytes, mappings, policy identities, resource costs and actual engine commits
remain in the archived JSON reports.

The known #312 repeated-claim budget abstention on exposed instruction page c05
is unchanged. No other abstention or operational error is accepted. Technical
passes all sets. Strict retains the earlier forbidden worth-noting failure in
the exposed instruction set. Text-mode Libevent extraction includes C examples;
they were not treated as prose defects. The inherited historical CMake text-mode
limitation also remains explicit.

| Set | Profile | Seconds before → after | Peak MiB before → after |
| --- | --- | --- | --- |
| development | technical | 3.537 → 2.525 | 211.2 → 191.1 |
| development | strict | 1.637 → 1.492 | 190.5 → 188.5 |
| exposed_repetition | technical | 0.623 → 0.546 | 122.0 → 111.6 |
| exposed_repetition | strict | 0.603 → 0.540 | 125.6 → 129.0 |
| exposed_framing | technical | 0.829 → 0.736 | 127.1 → 129.9 |
| exposed_framing | strict | 0.835 → 0.723 | 132.5 → 136.1 |
| exposed_local | technical | 0.710 → 0.633 | 122.3 → 122.1 |
| exposed_local | strict | 0.712 → 0.656 | 117.3 → 127.0 |
| exposed_context | technical | 0.737 → 0.693 | 120.7 → 121.8 |
| exposed_context | strict | 0.739 → 0.671 | 123.0 → 121.1 |
| exposed_instruction | technical | 0.978 → 0.921 | 165.2 → 160.6 |
| exposed_instruction | strict | 0.989 → 0.961 | 160.5 → 148.4 |
| exposed_construction | technical | 0.528 → 0.485 | 107.4 → 102.7 |
| exposed_construction | strict | 0.510 → 0.491 | 100.8 → 105.9 |
| exposed_purpose | technical | 0.531 → 0.503 | 102.9 → 102.7 |
| exposed_purpose | strict | 0.513 → 0.504 | 104.3 → 102.8 |
| exposed_verb | technical | 0.563 → 0.551 | 98.0 → 105.2 |
| exposed_verb | strict | 0.558 → 0.548 | 100.0 → 96.6 |
| exposed_relations | technical | 0.602 → 0.578 | 108.5 → 118.0 |
| exposed_relations | strict | 0.609 → 0.573 | 124.2 → 112.5 |
| exposed_projection | technical | 0.519 → 0.499 | 109.0 → 111.8 |
| exposed_projection | strict | 0.490 → 0.493 | 105.4 → 103.2 |
| exposed_scope | technical | 0.741 → 0.733 | 154.8 → 155.9 |
| exposed_scope | strict | 0.744 → 0.735 | 163.3 → 155.6 |
| exposed_proposition | technical | 0.434 → 0.445 | 85.3 → 79.8 |
| exposed_proposition | strict | 0.451 → 0.468 | 80.3 → 87.2 |
| exposed_clauses | technical | 0.476 → 0.477 | 93.6 → 90.6 |
| exposed_clauses | strict | 0.495 → 0.483 | 85.7 → 90.0 |
| confirmation | technical | 0.466 → 0.448 | 96.2 → 89.5 |
| confirmation | strict | 0.390 → 0.444 | 89.9 → 87.7 |

Measurements overlapped ordinary checks and are not a controlled performance
comparison or 2-vCPU qualification. Run `python3 tools/render.py` and
`python3 tools/test_evidence.py` to validate and regenerate the report.
`tools/measure.py` replays an explicit binary into a new directory. See
[VALIDATION.md](VALIDATION.md) for actual product-check results. This iteration
has not been merged or deployed to the playground.
