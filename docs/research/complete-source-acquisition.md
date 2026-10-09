# Complete-purpose acquisition and its learning effect

A source-only acquisition experiment now has 49 completed reviews. Sampling by
student uncertainty found more mandatory edits than its random control in this
packet. Adding all 49 reviews to the preceding masked student nevertheless
reduced complete target-byte coverage from 11/123 events to 0/123. The resulting
candidate fails the recall screen and is rejected.

## Acquisition and review

The protocol was registered before selection. It uses all 13 exposed development
projects, excluding 333 source units from earlier review packets, including the
old quarantines. For each project, it draws up to two random controls and then
up to two remaining units with the smallest absolute global-necessity logit.
The scoring head excludes the queried project from training and calibration.
Scores are acquisition rankings, not calibrated probabilities. Heads without a
usable delivery cutoff can still rank acquisition candidates.

Eligibility requires 24 cached target BPE tokens, not 24 prose words or a
paragraph. Linked release labels and API signatures can therefore be selected.
Etcd has only one eligible unreviewed unit. Its three-case shortfall remains in
the record; no project is dropped and no replacement source is introduced.

The mixed packet excludes selection arms, scores, model explanations, original
judgments, and reference-event labels. The root Codex assistant reviewed the
complete focus unit and neighboring frozen source, consulting wider original
source when needed. This is assistant review under ADR 0041, not independent
human annotation. All 49 responses were completed and sealed before reading the
arm ledger.

Each response explicitly assesses global edit necessity and all eight
nonexclusive purposes, with a separate reason for each decision. Unknown and
optional answers remain masked. Mandatory edits have exact original UTF-8
quotes and spans. No repaired text is automatically labeled clean, and untouched
tokens are not automatically labeled background.

| Selection arm | Cases | Mandatory edit | Retain | Optional | Unknown |
| --- | ---: | ---: | ---: | ---: | ---: |
| Random control | 25 | 3 | 18 | 4 | 0 |
| Necessity uncertainty | 24 | 9 | 12 | 3 | 0 |

The mandatory-edit yield is 12.0% for this random packet and 37.5% for this
uncertainty packet. These are acquisition yields from one small, grouped,
exposed-source comparison with one reviewer. They do not establish population
prevalence, classifier precision, or a generally superior selection strategy.

A separate read-only audit reconstructed all 13 saved ranking heads, selection
choices, mixed review order, source slices, localized edits, and observation
masks exactly. It also preserved all 333 exclusions and the three-case shortfall.
The completed review set has 14 localized edits, expanding to 16 positive
purpose-specific targets when a repair serves more than one purpose.

## What supervision improved

The 49 reviews add 14 positive, 338 negative, and 40 optional or unknown intent
dimensions. Combined with the preceding 256 cases, the 305-case packet has
102 positive, 349 negative, and 1,989 unobserved intent dimensions. Global
necessity has 91 positive, 183 negative, and 31 optional or unknown labels.

| Purpose | Positive units, before → after | Negative units, before → after |
| --- | ---: | ---: |
| Clarity | 4 → 8 | 0 → 42 |
| Empty framing | 36 → 36 | 3 → 50 |
| Fluency | 26 → 35 | 0 → 36 |
| Needless complexity | 1 → 2 | 1 → 47 |
| Needless repetition | 1 → 1 | 2 → 50 |
| Unjustified intensifiers | 4 → 4 | 2 → 44 |
| Vague claims | 5 → 5 | 1 → 44 |
| Wordiness | 11 → 11 | 2 → 36 |

All eight purposes now have both classes in the combined packet. That is not a
qualification result. Some calibration folds still lack a positive purpose:
needless complexity has both classes in only 1/13 calibration folds, repetition
in 3/13, and wordiness in 3/13. Even the presence of both classes does not prove
that ten correct complaints can be selected at the required precision.

New mandatory cases mostly concern grammar and clarity. The acquisition does
not add mandatory examples of empty framing, repetition, intensification,
vague claims, or wordiness. Useful retention controls include real version
counts, operational negations, and contrasts between different API behaviors.
They must not become violations just because a sentence contains a count,
negative construction, or contrast.

## Matched learning comparison

A separate registration was frozen before fitting. It uses the same cached
ModernBERT features, 1,536-to-64-to-9 head, project folds, initialization,
optimizer, 120 epochs, loss weights, cutoff grid, and delivery procedure as the
[preceding masked experiment](partial-label-learning.md). The control uses its
256 cases; the treatment adds the 49 complete-purpose reviews. Optional and
unknown dimensions remain masked in both arms. No repaired variants enter.

The run completed 26 heads in 179.71 seconds, with a 629,702,656-byte peak child
resident set. All 13 control weight files reproduced the preceding masked
heads exactly. A separately registered saved-weight audit reproduced every
calibration grid and delivered span and independently checked target-byte unions.
No provider request, new encoder pass, or encoder fit occurred.

| Measurement | Historical masks | Added complete-purpose reviews |
| --- | ---: | ---: |
| Source cases | 256 | 305 |
| Folds with an admissible global-necessity proxy cutoff | 3/13 | 3/13 |
| Available original units | 2,411/4,549 | 788/4,549 |
| Proposed criticisms | 1,262 | 144 |
| Events with complete target-byte union, any purpose | 11/123 | 0/123 |
| Events with complete union and matching purpose | 5/123 | 0/123 |
| Semantic FULL recall and actual usefulness | Unmeasured | Unmeasured |
| Product acceptance credit | 0 | 0 |

All 123 original events and all 4,549 units remain accounted for in each arm.
The treatment has no complete byte-union witness in any cohort: 0/46 context,
0/57 whole-document, and 0/20 historical exposed confirmation. It cannot meet
80% FULL recall under the frozen component-localization contract. The stream
therefore fails the recall screen before semantic usefulness review.

Fewer proposals do not demonstrate better precision. The experiment changes
both training and calibration supervision, so it does not separate their
individual effects or the effects of random and uncertainty acquisitions.
Acquisition is also not nested validation: ranking heads for other projects may
have seen old labels from the subsequently tested project. Direct fitting and
calibration still exclude that project, but this study supplies no independent
generalization claim.

## Research and Python directions

The [prior literature review](partial-label-learning.md#research-and-python-implementations)
covers LLMaAA, partial-token active learning, and small-text. This experiment
uses a controlled uncertainty-versus-random idea; it does not replicate those
papers or import their results as Unswell accuracy.

The failed local token head also motivates examining relations between text
parts. Primary sources and author implementations were inspected on October 9,
2026:

| Source | Relevant capability | Boundary |
| --- | --- | --- |
| [DisCoDisCo](https://aclanthology.org/2021.disrpt-1.6/), [author Python code](https://github.com/gucorpling/DisCoDisCo) | Segmentation, connective detection, and relation classification, including a sentence-pair classifier with features beyond contextual embeddings. | A discourse relation is not an unnecessary editorial move. No model or package was run here. |
| [DMRST](https://aclanthology.org/2021.codi-main.15/), [paper](https://aclanthology.org/2021.codi-main.15.pdf) | Joint segmentation and document-level rhetorical trees, including the relative roles of connected spans. | Its paper identifies domain-generalization problems. Tree accuracy does not establish criticism usefulness on technical documentation. |
| [Neural RST-based Evaluation of Discourse Coherence](https://aclanthology.org/2020.aacl-main.67/), [paper](https://aclanthology.org/2020.aacl-main.67.pdf) | Tests automatically parsed rhetorical-tree features for coherence classification on GCDC. | Coherence labels differ from the need for a specific technical edit; this is a feature hypothesis, not a transferred quality score. |
| [IUDEX Python parsers](https://github.com/larc-iu/iudex) | Author-maintained implementations with raw-text DMRST inference and tree output. | Source was pinned and inspected; package and model inference were not executed. Model licensing and a source-map adapter still require qualification. |

IUDEX source at commit `2bb85a7a0190d64d7d7152c920d9f13f3a331512` declares
Python 3.10 or newer and an MIT software license. Context7 returned no matching
package, so the official source was used. Its DMRST path retains character maps
internally but returns tree segments with rewritten whitespace. An adapter must
preserve original character and UTF-8 byte spans before that rewrite; searching
repeated output strings in the source is insufficient.

The next representation comparison should test relation features and
purpose-specific acquisition against the existing local head. It must retain
legitimate contrasts and literal counts, keep existing reference labels frozen,
and assess every actual complaint and complete event. Detecting a relation or
acquiring more labels alone supplies no product improvement credit.

## Evidence and product status

The [measurement record](complete-source-acquisition-result.json) binds both
registrations, all reviews by digest, all 26 saved heads, terminal native runs,
calibration grids, complete unit accounting, source-selection audit, and the
independent component-union audit. Raw reviews and model artifacts remain in the
local research archive. Publishing these aggregates does not claim public
reproduction of private editorial judgments.

The published analyzer remains at 37/123 FULL events (30.1%) and 94/154 useful
actual criticisms (61.0%). Targets remain at least 80% FULL and 85% usefulness
in each cohort, zero unsafe or unresolved supplied advice, and separate untouched
confirmation. No runtime, model distribution, gate, or playground change is
included. [Issue #349](https://github.com/stokaro/unswell/issues/349) remains open.
