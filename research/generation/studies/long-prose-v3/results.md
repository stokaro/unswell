# Complete-document results, September 15, 2026

This record executes protocol version 3 on 38 dated technical documents from
20 repositories and three model families. It measures constructions and review
load. It supplies no human quality labels, authorship probabilities, or default
blocking-rule qualification.

The earlier version 2 confirmation remains an underpowered result with 16 global
components. Version 3 uses fresh repositories and a separate frozen comparison.

## Coverage

Every request has one saved outcome. Truncations and other incomplete responses
remain in coverage and do not become zero-finding documents. Off-length outputs
remain in the primary analysis. Reported tokens include visible reasoning usage
where the harness supplies it; hidden provider settings remain unavailable.

| Family | Requests | Complete | Truncated | Refused | Error | Missing | Off length | Reported tokens |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| luna | 152 | 152 | 0 | 0 | 0 | 0 | 24 | 410,828 |
| haiku | 152 | 152 | 0 | 0 | 0 | 0 | 34 | 1,077,160 |
| qwen | 152 | 151 | 1 | 0 | 0 | 0 | 41 | 494,397 |

The delivery record includes a declared resource amendment. After 148 Haiku
responses and 1,962,742 reported tokens across the three models, the runner’s
65,536-token per-call reservation prevented the final four requests. The
dispatch ceiling increased from 2,000,000 to 2,150,000 tokens for those fixed
requests. This happened after Qwen and Luna construction results were inspected;
the joint primary estimates had not been computed. The original freeze remains
unchanged in the input archive, and `generation-completion.json` records the
time, request IDs, wrapper hash, and prior exposure. Sample, prompts, models,
retry eligibility, deadline, hypotheses, and decision thresholds did not change.
The completed run reports 1,982,385 tokens in total. This is an explicit
execution amendment, not an unamended execution of the original resource plan.

## Frozen primary comparisons

Each row compares generate/neutral with its own originals. Differences and
95% percentile intervals use the Go joint component bootstrap (10,000 draws,
PCG seed 17). Tail probabilities receive Holm correction across all 12 tests.
A family result also needs an increase of at least three percentage points,
20 measured groups, and five groups with a response finding.

| Family | Rule | Originals | Responses | Difference and interval (pp) | Holm p | Groups / support | Criteria |
| --- | --- | --- | --- | --- | --- | --- | --- |
| luna | `syntax.paired-contrast-density` | 1/38 | 3/38 | +5.3 [-5.1, +15.8] | 1 | 20 / 3 | not met |
| luna | `syntax.passive-candidate-density` | 27/38 | 30/38 | +7.9 [-10.8, +26.3] | 1 | 20 / 18 | not met |
| luna | `syntax.noun-stack` | 3/38 | 2/38 | -2.6 [-14.6, +8.6] | 1 | 20 / 2 | not met |
| luna | `readability.grade-metric` | 10/38 | 17/38 | +18.4 [-0.1, +38.7] | 0.7524 | 20 / 13 | not met |
| haiku | `syntax.paired-contrast-density` | 1/38 | 2/38 | +2.6 [-5.7, +11.8] | 1 | 20 / 2 | not met |
| haiku | `syntax.passive-candidate-density` | 27/38 | 28/38 | +2.6 [-16.2, +21.2] | 1 | 20 / 16 | not met |
| haiku | `syntax.noun-stack` | 3/38 | 8/38 | +13.2 [+0.0, +27.0] | 0.7524 | 20 / 6 | not met |
| haiku | `readability.grade-metric` | 10/38 | 32/38 | +57.9 [+37.8, +78.0] | 0.0012 | 20 / 19 | met |
| qwen | `syntax.paired-contrast-density` | 1/38 | 1/38 | +0.0 [-7.7, +7.7] | 1 | 20 / 1 | not met |
| qwen | `syntax.passive-candidate-density` | 27/38 | 35/38 | +21.1 [+6.5, +36.1] | 0.07 | 20 / 18 | not met |
| qwen | `syntax.noun-stack` | 3/38 | 4/38 | +2.6 [-6.1, +10.8] | 1 | 20 / 4 | not met |
| qwen | `readability.grade-metric` | 10/38 | 34/38 | +63.2 [+45.9, +80.5] | 0.0012 | 20 / 20 | met |

A cross-family claim needs two qualifying families and a positive point estimate
in the third. A missing or weak result is not proof of equivalence. The frozen
advisory-load budget is at most one finding per 1,000 historical words and
findings on at most 25% of historical documents. This is review workload, not
a false-positive estimate. A load restriction limits an AI-associated advisory
claim; it does not disable an existing general readability rule.

| Rule | Qualifying families | Cross-family criteria | Historical /1k | Historical prevalence | Disposition |
| --- | --- | --- | --- | --- | --- |
| `syntax.paired-contrast-density` | 0 | false | 0.017 | 2.6% | association_not_established |
| `syntax.passive-candidate-density` | 0 | false | 1.789 | 71.1% | restricted_by_historical_review_load |
| `syntax.noun-stack` | 0 | false | 0.052 | 7.9% | association_not_established |
| `readability.grade-metric` | 2 | true | 0.516 | 26.3% | restricted_by_historical_review_load |

## Whole-profile load and length

All 40 rules use one explicit measurement policy. These totals are not the
technical or strict product profiles. Document prevalence grows with exposure;
per-word load, repository-macro prevalence, and realized-length slices are
reported alongside it. Historical provenance does not certify clean prose.

| Family | Operation | Prompt | Docs | Words | Findings | /1k words | Doc prevalence | Repo-macro prevalence |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| haiku | generate | neutral | 38 | 40,918 | 362 | 8.847 | 92.1% | 93.3% |
| haiku | generate | plain | 38 | 37,333 | 292 | 7.821 | 94.7% | 95.8% |
| haiku | polish | neutral | 38 | 57,426 | 237 | 4.127 | 84.2% | 83.3% |
| haiku | polish | plain | 38 | 52,746 | 185 | 3.507 | 76.3% | 73.3% |
| historical | original | original | 38 | 58,137 | 244 | 4.197 | 86.8% | 85.8% |
| luna | generate | neutral | 38 | 35,830 | 159 | 4.438 | 81.6% | 81.7% |
| luna | generate | plain | 38 | 36,235 | 155 | 4.278 | 73.7% | 78.3% |
| luna | polish | neutral | 38 | 48,869 | 152 | 3.110 | 76.3% | 80.0% |
| luna | polish | plain | 38 | 43,506 | 114 | 2.620 | 60.5% | 63.3% |
| qwen | generate | neutral | 38 | 48,228 | 432 | 8.957 | 97.4% | 97.5% |
| qwen | generate | plain | 38 | 43,125 | 353 | 8.186 | 97.4% | 97.5% |
| qwen | polish | neutral | 37 | 54,630 | 250 | 4.576 | 89.2% | 90.0% |
| qwen | polish | plain | 38 | 57,192 | 221 | 3.864 | 81.6% | 78.3% |


| Family | Operation | Prompt | Median extracted-length ratio | Same band | Within 30% |
| --- | --- | --- | --- | --- | --- |
| haiku | generate | neutral | 0.800 | 23/38 | 26/38 |
| haiku | generate | plain | 0.748 | 21/38 | 22/38 |
| haiku | polish | neutral | 0.995 | 37/38 | 38/38 |
| haiku | polish | plain | 0.910 | 30/38 | 33/38 |
| luna | generate | neutral | 0.602 | 11/38 | 10/38 |
| luna | generate | plain | 0.601 | 16/38 | 10/38 |
| luna | polish | neutral | 0.955 | 29/38 | 34/38 |
| luna | polish | plain | 0.813 | 25/38 | 23/38 |
| qwen | generate | neutral | 0.909 | 27/38 | 29/38 |
| qwen | generate | plain | 0.856 | 24/38 | 31/38 |
| qwen | polish | neutral | 1.000 | 34/37 | 37/37 |
| qwen | polish | plain | 1.000 | 34/38 | 38/38 |

The generation ledger measures whitespace fields; this sensitivity uses
engine-extracted prose words on both sides. Same-band and within-30% comparisons
are descriptive restrictions after observing length. They do not replace the
primary sample or provide a second independent confirmation.

## Context opportunities

Whitespace-adjacent paragraph runs describe available context. Possible sentence
pairs within eight positions are structural opportunities, not the matcher’s
rule-specific event clusters. Headings, code, and source gaps affect continuity.
Cross-block findings directly count diagnostics with evidence in multiple blocks.

| Family | Operation | Prompt | Paragraphs | Headings | Heading/paragraph pairs | Run sentences | Runs >=8 | Possible pairs | Cross-block findings |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| haiku | generate | neutral | 1184 | 606 | 548 | 2946 | 87 | 7856 | 91 |
| haiku | generate | plain | 1122 | 618 | 542 | 3095 | 94 | 8755 | 68 |
| haiku | polish | neutral | 1435 | 987 | 354 | 2365 | 42 | 5288 | 106 |
| haiku | polish | plain | 1378 | 1017 | 390 | 2239 | 42 | 4211 | 81 |
| historical | original | original | 1463 | 987 | 350 | 2395 | 33 | 5842 | 99 |
| luna | generate | neutral | 1478 | 504 | 466 | 2979 | 95 | 7683 | 85 |
| luna | generate | plain | 1532 | 573 | 525 | 3217 | 90 | 8387 | 84 |
| luna | polish | neutral | 1267 | 917 | 339 | 2045 | 39 | 4439 | 81 |
| luna | polish | plain | 1431 | 839 | 343 | 2348 | 45 | 5475 | 79 |
| qwen | generate | neutral | 1316 | 779 | 751 | 3311 | 95 | 8454 | 141 |
| qwen | generate | plain | 1054 | 501 | 465 | 3004 | 91 | 10297 | 94 |
| qwen | polish | neutral | 1338 | 881 | 349 | 2240 | 40 | 5045 | 105 |
| qwen | polish | plain | 1440 | 994 | 351 | 2416 | 44 | 5663 | 103 |


| Family | Rule | Applicable paragraphs | Positive activations | Activation share | Sentences in applicable paragraphs | Cross-block diagnostics |
| --- | --- | --- | --- | --- | --- | --- |
| haiku | `readability.grade-metric` | 189 | 143 | 75.7% | 858 | 0 |
| haiku | `syntax.noun-stack` | 1160 | 11 | 0.9% | 2922 | 0 |
| haiku | `syntax.paired-contrast-density` | 730 | 3 | 0.4% | 2329 | 1 |
| haiku | `syntax.passive-candidate-density` | 1052 | 225 | 21.4% | 2773 | 68 |
| historical | `readability.grade-metric` | 122 | 30 | 24.6% | 453 | 0 |
| historical | `syntax.noun-stack` | 1366 | 3 | 0.2% | 2293 | 0 |
| historical | `syntax.paired-contrast-density` | 502 | 2 | 0.4% | 1326 | 1 |
| historical | `syntax.passive-candidate-density` | 1093 | 179 | 16.4% | 1963 | 64 |
| luna | `readability.grade-metric` | 59 | 26 | 44.1% | 256 | 0 |
| luna | `syntax.noun-stack` | 1446 | 3 | 0.2% | 2946 | 0 |
| luna | `syntax.paired-contrast-density` | 704 | 7 | 1.0% | 1923 | 3 |
| luna | `syntax.passive-candidate-density` | 1219 | 184 | 15.1% | 2655 | 68 |
| qwen | `readability.grade-metric` | 259 | 161 | 62.2% | 1181 | 0 |
| qwen | `syntax.noun-stack` | 1295 | 4 | 0.3% | 3289 | 0 |
| qwen | `syntax.paired-contrast-density` | 840 | 1 | 0.1% | 2657 | 0 |
| qwen | `syntax.passive-candidate-density` | 1189 | 265 | 22.3% | 3148 | 82 |

`descriptive.json` also retains engine-owned applicability, reasons for missing
block activations, and sentence/paragraph bindings for every rule, role, arm,
and realized-length band. A positive block activation means contributing evidence;
it is not an additional diagnostic. Sentence opportunities inside applicable
paragraphs are not independently established per-sentence applicability.

The pre-repair comparison on exposed Ptah data is recorded separately in
`docs/research/contextual-prose.md`. Its counts are development evidence and are
not pooled with this new sample.

## Descriptive ablations

These tables remove rule families from saved counts; they do not change the
engine or repeat model generation. The complete rule-to-family mapping is in
`descriptive.json` and was assigned for descriptive reporting after the freeze.

| Family | All | Without phrase | Without structural | Without contextual |
| --- | --- | --- | --- | --- |
| haiku | 362 | 362 | 181 | 181 |
| historical | 244 | 227 | 158 | 103 |
| luna | 159 | 159 | 119 | 40 |
| qwen | 432 | 415 | 255 | 194 |

## Every rule

Counts below are descriptive generate/neutral counts, with historical load
alongside them. A zero is a result within this finite sample, not universal
absence. General editing rules need not separate model cohorts to remain useful.
Unlisted hypotheses cannot become confirmatory discoveries after inspection.
`analysis/evidence-cards.json` retains each rule definition, capability requirements,
classification, paired estimates for every arm, cases, and disposition.

| Rule | Historical | Luna | Haiku | Qwen | Disposition |
| --- | --- | --- | --- | --- | --- |
| `density.connective-overuse` | 0 | 0 | 0 | 1 | retain as general editing measurement |
| `filler.announced-importance` | 0 | 0 | 0 | 17 | observed candidate; no primary confirmation |
| `filler.empty-transition` | 0 | 0 | 0 | 0 | no observed support in this study; retain narrow scope |
| `filler.modern-world-opening` | 0 | 0 | 0 | 0 | no observed support in this study; retain narrow scope |
| `filler.section-announcement` | 0 | 0 | 0 | 0 | no observed support in this study; retain narrow scope |
| `filler.stacked-hedging` | 0 | 0 | 0 | 0 | no observed support in this study; retain narrow scope |
| `filler.weak-intensifiers` | 2 | 0 | 0 | 0 | observed candidate; no primary confirmation |
| `filler.wordy-phrase` | 15 | 0 | 0 | 0 | retain as general editing measurement |
| `format.em-dash-density` | 0 | 0 | 1 | 0 | retain as general editing measurement |
| `format.list-fragmentation` | 0 | 0 | 0 | 0 | no observed support in this study; retain narrow scope |
| `hype.absolute-claim` | 0 | 0 | 0 | 0 | no observed support in this study; retain narrow scope |
| `hype.metaphor-cluster` | 0 | 0 | 0 | 0 | no observed support in this study; retain narrow scope |
| `hype.modifier-cluster` | 0 | 0 | 0 | 0 | no observed support in this study; retain narrow scope |
| `hype.vague-praise` | 0 | 0 | 0 | 0 | no observed support in this study; retain narrow scope |
| `policy.banned-phrases` | 0 | 0 | 0 | 0 | configured policy; no phrase list tested |
| `readability.grade-metric` | 30 | 26 | 143 | 161 | restricted_by_historical_review_load |
| `readability.long-paragraph` | 3 | 0 | 1 | 1 | retain as general editing measurement |
| `repetition.exact-sentence` | 1 | 0 | 0 | 1 | retain as general editing measurement |
| `repetition.heading-echo` | 0 | 0 | 0 | 0 | no observed support in this study; retain narrow scope |
| `repetition.near-sentence` | 2 | 1 | 1 | 0 | retain as general editing measurement |
| `repetition.ngram-density` | 3 | 0 | 0 | 2 | retain as general editing measurement |
| `repetition.paragraph-openers` | 12 | 6 | 6 | 16 | retain as general editing measurement |
| `repetition.paragraph-overlap` | 1 | 0 | 0 | 0 | retain as general editing measurement |
| `repetition.sentence-openers` | 17 | 7 | 15 | 40 | retain as general editing measurement |
| `repetition.summary-echo` | 0 | 0 | 0 | 0 | no observed support in this study; retain narrow scope |
| `repetition.syntax-template` | 0 | 0 | 0 | 0 | no observed support in this study; retain narrow scope |
| `scaffold.ai-self-reference` | 0 | 0 | 0 | 0 | no observed support in this study; retain narrow scope |
| `scaffold.chat-preamble` | 0 | 0 | 0 | 0 | no observed support in this study; retain narrow scope |
| `scaffold.dive-in` | 0 | 0 | 0 | 0 | no observed support in this study; retain narrow scope |
| `scaffold.follow-up-offer` | 0 | 0 | 0 | 0 | no observed support in this study; retain narrow scope |
| `syntax.long-sentence` | 35 | 11 | 22 | 10 | retain as general editing measurement |
| `syntax.nominalization-chain` | 0 | 0 | 2 | 0 | retain as general editing measurement |
| `syntax.not-only-density` | 0 | 0 | 0 | 0 | no observed support in this study; retain narrow scope |
| `syntax.noun-stack` | 3 | 3 | 11 | 4 | association_not_established |
| `syntax.paired-contrast-density` | 1 | 4 | 2 | 1 | association_not_established |
| `syntax.parenthetical-load` | 15 | 0 | 1 | 1 | retain as general editing measurement |
| `syntax.passive-candidate-density` | 104 | 101 | 157 | 177 | restricted_by_historical_review_load |
| `syntax.rhetorical-question-density` | 0 | 0 | 0 | 0 | no observed support in this study; retain narrow scope |
| `syntax.triad-density` | 0 | 0 | 0 | 0 | no observed support in this study; retain narrow scope |
| `syntax.whether-preface-density` | 0 | 0 | 0 | 0 | no observed support in this study; retain narrow scope |

## Inspectable cases and limits

The source-bound Qwen preface examples and their agent-written edits are in
`e2e/testdata/long_prose_research`: the original prefaces are detected and the
edited wording retains benchmark numbers, unaffected behavior, file names,
volunteer delivery, and the absence of release-date promises. The Moment
counterexample retains two necessary behavioral contrasts. A repeated form
does not make either distinction expendable.

Ptah’s three contiguous contrast examples and contract-preserving alternatives
remain in `e2e/testdata/contextual_ptah`. These fixtures run offline through the
built CLI, JSON, and SARIF. They establish construction behavior and coordinates,
not human precision or reader preference. The Haiku nominalization fixture
replaces only `conducts an investigation of` with `investigates`, retaining
acknowledgment, rejection reasons, policy conditions, and fix-work notification.

Historical passive candidates include ordinary installation and distribution
descriptions. Repeated installation instructions and shim-command obligations
are counterexamples to treating parallel openings as automatically defective.
Grade estimates and noun sequences can reflect required terminology and layout.
Use the case record with full source spans when assessing a particular finding.

This study covers three fixed low-effort model/harness choices and a small
English technical source frame. It does not establish behavior for every model,
genre, language background, or future prompt. Source curation was by an agent.
Human quality qualification remains deferred to #22/#26, and none of these
findings is an authorship probability or a reason to block CI by provenance.

See the data card, frozen protocol, archive checksums, and reproduction commands
beside this report. The earlier study’s placebo and earlier-date sensitivities
remain under their original policies; their counts are not silently pooled here.
