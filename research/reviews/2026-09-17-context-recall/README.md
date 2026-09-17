# Contextual wording: known-case gains without confirmation gain

This iteration continues #299 with bounded extensions to three existing
rules. Five added warnings identify six more frozen defects on previously
reviewed pages. The eleven new complete pages gain no detections: only
one of 42 defects is found. The broad-recall objective remains unmet.

The implementing Codex assistant reviewed the prose under the maintainer
acceptance in [ADR 0041](../../../docs/adr/0041-assistant-review-acceptance.md).
The labels are subjective, source-bound editorial judgments. They are not
human annotations, independent agreement or population accuracy. Another
reviewer is not a prerequisite for continuing this work.

## Separate evaluation sets

Both profiles have the same event counts. Keep these denominators separate:
the framing and repetition sets have narrower annotation scopes than the
three complete-page reviews. An overlap with a length warning does not
count when the warning misses the actual wording construction.

| Set | Pages | Frozen defects | Before | After |
| --- | ---: | ---: | ---: | ---: |
| Original development | 26 | 57 | 13 (22.8%) | 13 (22.8%) |
| Exposed repetition targets | 6 | 5 | 3 (60.0%) | 3 (60.0%) |
| Exposed framing targets | 9 | 11 | 1 (9.1%) | 1 (9.1%) |
| Exposed complete pages | 12 | 78 | 1 (1.3%) | 7 (9.0%) |
| New complete-page confirmation | 11 | 42 | 1 (2.4%) | 1 (2.4%) |

The added warnings cover a page explaining its own explanation, a
cognitive-worth preface, two unrestricted reader claims, and two exaggerated
quality claims grouped into one diagnostic. [CHANGES.md](CHANGES.md) retains
their original source spans, credit and review rationale. No warning is removed.

Matcher versions change to `filler.document-justification` v3,
`filler.evaluative-closure` v3 and `filler.unscoped-assurance` v2.
The catalog remains at 51 rules. Weights, gate thresholds, profiles and
rule classes are unchanged. Clause analysis preserves explicit scope,
measurements, quoted claims, negation and protected text.

## New confirmation and review burden

The assistant froze the source identities and selection method, then read
all source prose. Both steps preceded implementation and any inspection of
diagnostic output. The review recorded 42 defects, 13 uncertain events and
25 acceptable controls.
Selection excludes the 53 preceding sources. The requested historical-long
cell had one unused page left, so the sample contains eleven pages rather
than twelve. This deficit was recorded before reading the selected prose.

The sole detected defect is an accidental adjacent-word duplication in the
Mermaid page. The new constructions produce no confirmation positives.
Every confirmation warning, including unchanged ones, has a disposition.

| Set | Profile | Findings before → after | Actionable after | Uncertain | Nonactionable |
| --- | --- | ---: | ---: | ---: | ---: |
| Original development | technical | 127 → 127 | 24 | 38 | 65 |
| Original development | strict | 146 → 146 | 24 | 42 | 80 |
| Exposed complete pages | technical | 66 → 71 | 6 | 27 | 38 |
| Exposed complete pages | strict | 72 → 77 | 6 | 27 | 44 |
| New complete-page confirmation | technical | 28 → 28 | 1 | 6 | 21 |
| New complete-page confirmation | strict | 37 → 37 | 1 | 6 | 30 |

All scans pass the existing gate. Those passes coexist with substantial
missed wording problems and review burden. The fresh sample cannot qualify
the new matchers: it contains no new-rule positive diagnostics. Do not turn
the known-case improvement into a general precision or recall claim.

The proposed 80% event recall and 85% soft-diagnostic precision target is
not met. [CONFIRMATION-MISSES.md](CONFIRMATION-MISSES.md) lists the remaining
frozen defects, including introductory scaffolding, vague claims, repetition
and wordiness. The [ledger](dispositions.json) binds judgments to report
hashes; [summary.json](summary.json) contains every page/category breakdown.

| Confirmation page | Cohort | Defects | Detected |
| --- | --- | ---: | ---: |
| c01 `docs/site/src/content/docs/direct/overview.md` | ptah | 1 | 0 |
| c02 `docs/site/src/content/docs/inference/concepts/embeddings-and-inference-state.md` | ptah | 3 | 0 |
| c03 `docs/site/src/content/docs/direct/apply.md` | ptah | 2 | 0 |
| c04 `docs/site/src/content/docs/schema/go-annotations.md` | ptah | 4 | 0 |
| c05 `docs/site/src/content/docs/reference/exit-codes.md` | ptah | 4 | 0 |
| c06 `docs/site/src/content/docs/reference/go-annotations.md` | ptah | 2 | 0 |
| c07 `docs/sequenceDiagram.md` | historical | 9 | 1 |
| c08 `tokio/README.md` | historical | 3 | 0 |
| c09 `docs/ops.md` | historical | 8 | 0 |
| c10 `CONTRIBUTING.md` | historical | 3 | 0 |
| c11 `CHANGELOG.md` | historical | 3 | 0 |

| Category | Confirmation defects | Detected |
| --- | ---: | ---: |
| empty_framing | 8 | 0 |
| formulaic_transitions | 1 | 0 |
| needless_complexity | 2 | 0 |
| needless_repetition | 4 | 1 |
| unjustified_intensifiers | 3 | 0 |
| vague_claims | 9 | 0 |
| wordiness | 15 | 0 |

Paired page bootstrap within cohort/length cells uses seed 917304 and
2,000 draws. It describes sensitivity to this small sample, not a population
confidence interval. Shared repositories, the single reviewer, genre mix
and the exhausted historical-long cell limit transfer.

Confirmation recall sensitivity: 0.0%–4.9%; paired gain: 0.0%–0.0%.

## All-rule inspection

A separate exploratory scan enabled all 51 existing rules on the exposed
sets. On the twelve recently reviewed pages it increased findings from
66 to 166. Passive-voice candidates and readability grades accounted for
87 of the 100 additions. This is not a completed all-rule recall study:
the extra warnings were inspected by rule and context, not exhaustively
credited against every defect. Reports, policy and hashes remain under
`all-rules-probe/`. The experiment does not change shipped defaults.

## Replay and resources

Baseline runtime: `0973ff633aee53bc9fc3592f80803806bf2d62e6`.
Measured implementation: `c1032187fcc1708617ac9aff53ce12c84c42f70c`.
The four exposed baseline report sets come from the preceding iteration.
Their original resource records remain intact. This iteration ran the
confirmation baseline and all after scans. No scan abstained or reported
an operational error.

| Set | Profile | Wall seconds before → after | Peak MiB before → after |
| --- | --- | ---: | ---: |
| Original development | technical | 2.126 → 2.133 | 187.2 → 181.9 |
| Original development | strict | 1.505 → 1.560 | 200.3 → 193.4 |
| Exposed repetition targets | technical | 0.541 → 0.527 | 116.5 → 113.6 |
| Exposed repetition targets | strict | 0.534 → 0.534 | 136.9 → 124.4 |
| Exposed framing targets | technical | 0.771 → 0.727 | 130.2 → 137.6 |
| Exposed framing targets | strict | 0.766 → 0.724 | 148.4 → 140.8 |
| Exposed complete pages | technical | 0.608 → 0.667 | 119.8 → 120.0 |
| Exposed complete pages | strict | 0.610 → 0.623 | 118.2 → 116.8 |
| New complete-page confirmation | technical | 0.728 → 0.690 | 119.2 → 119.7 |
| New complete-page confirmation | strict | 0.651 → 0.659 | 116.9 → 123.3 |

These are single macOS arm64 runs, not a speedup claim or the separate
two-vCPU Linux target. Records retain CPU time, host and binary hashes.
Builds use `CGO_ENABLED=0`. Source is included in these authorized offline
research artifacts; ordinary saved reports still omit it by default.

From the repository root:

```sh
python3 research/reviews/2026-09-17-context-recall/tools/render.py
python3 research/reviews/2026-09-17-context-recall/tools/test_evidence.py
```

The renderer validates source hashes, freeze manifests, complete diagnostic
reviews, code and policy identity, every delta and source-bound event credit.
To repeat a scan, build the named revision and run
`tools/measure.py --binary PATH --set SET --output NEW_DIR` from this directory.
It requires a new output directory and performs no downloads.

The [protocol](protocol.md) and both freeze manifests remain unchanged.
This confirmation set is now exposed. Further tuning needs fresh pages
and a larger historical source frame. The negative result supports moving
beyond expansions of narrow surface constructions: semantic restatements
in #305 and broader contextual wordiness still require implementation.
