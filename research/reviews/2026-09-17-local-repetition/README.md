# Local repetition: measured improvement and remaining misses

This follow-up implements #304 and a limited construction related to #305.
It adds adjacent-word checks, short repeated assertions and explanatory
restarts. The full-page objective remains unmet. The new confirmation pages
gain no findings or event detections. Do not describe these changes as good
general recall or as qualified defaults.

The reviewer is the same Codex assistant that implemented the rules, accepted
under [ADR 0041](../../../docs/adr/0041-assistant-review-acceptance.md).
These are subjective development judgments with source-bound proposed edits.
They are not independent human ratings, model calibration or population
accuracy. Neither native-speaker review nor another rater blocks this work.

## Complete-page and targeted results

Both profiles have the same event counts. Keep the four denominators
separate: the original and new reviews cover all seven rubric categories;
the two intervening confirmation sets have narrower annotation scopes.

| Set | Pages | Frozen defects | Before | After |
| --- | ---: | ---: | ---: | ---: |
| Original development | 26 | 57 | 12 (21.1%) | 13 (22.8%) |
| Exposed repetition targets | 6 | 5 | 0 (0.0%) | 3 (60.0%) |
| Exposed framing targets | 9 | 11 | 1 (9.1%) | 1 (9.1%) |
| New complete-page confirmation | 12 | 78 | 1 (1.3%) | 1 (1.3%) |

The original exit-1 explanation is recovered. Three accidental word
duplications are detected in the exposed NATS and ripgrep pages. The
ripgrep complexity explanation receives a useful restart warning, but its
circular final rationale remains unresolved. That partial match does not
receive full event credit. The glossary paraphrase and permissive-license
restatement also remain missed; #305 stays open.

The new sample has 12 complete pages: six Ptah and six historical sources,
two per cohort/length cell. All source prose was reviewed before diagnostic
output: 78 defects, 10 uncertain judgments and 20 acceptable controls.
Selection excludes the preceding 41 pages. Source hashes and retained
notices are in [confirmation/manifest.json](confirmation/manifest.json).
The frozen labels are in [confirmation/annotations.json](confirmation/annotations.json).

The single detected confirmation event is a repeated document-scope
announcement in BuildKit. The existing metadiscourse warning locates both
announcements. A single scope announcement does not get repetition credit.
A length warning overlapping a different wording defect gets no credit.

## Findings and review burden

Every confirmation finding was reviewed, including unchanged findings.
An optional rewrite is uncertain, not automatically a true positive.

| Set | Profile | Findings before → after | Actionable after | Uncertain | Nonactionable |
| --- | --- | ---: | ---: | ---: | ---: |
| Original development | technical | 126 → 127 | 24 | 38 | 65 |
| Original development | strict | 145 → 146 | 24 | 42 | 80 |
| New complete-page confirmation | technical | 66 → 66 | 1 | 27 | 38 |
| New complete-page confirmation | strict | 72 → 72 | 1 | 27 | 44 |

The new rules add five diagnostics across the exposed sets and remove none.
Four identify complete frozen defects; one identifies part of the circular
explanation. The fresh sample supplies no new-rule positive cases, so it
cannot establish their recall or positive predictive value. Its unchanged
warnings still impose substantial review burden.

All scans pass the existing gate. Thresholds and weights of earlier rules
are unchanged. Gate PASS is compatible with missed wording defects.

The [complete diagnostic ledger](dispositions.json) preserves judgments and
report hashes. [CHANGES.md](CHANGES.md) lists every added warning.
[CONFIRMATION-MISSES.md](CONFIRMATION-MISSES.md) retains every missed event.

## Uncertainty and limits

The proposed 80% recall / 85% soft-diagnostic precision target is not met.
A paired page bootstrap resamples within cohort/length cells with seed
917304 and 2,000 draws. It is sensitivity to this small selected page set,
not an interval for all technical writing. Related Ptah pages share one
repository; the single reviewer, fixed source frame, and subjective labels
limit transfer. Changelog entries contribute many vague-claim labels, so
the page and category breakdowns must accompany the aggregate.

Confirmation recall sensitivity: 0.0%–3.3%; paired gain: 0.0%–0.0%.

| Confirmation page | Cohort | Defects | Detected |
| --- | --- | ---: | ---: |
| c01 `docs/site/src/content/docs/databases/support-matrix.md` | ptah | 1 | 0 |
| c02 `docs/site/src/content/docs/inference/guides/configure-a-provider.md` | ptah | 3 | 0 |
| c03 `docs/site/src/content/docs/testing/ci.md` | ptah | 2 | 0 |
| c04 `docs/site/src/content/docs/operate/deliver.mdx` | ptah | 4 | 0 |
| c05 `docs/site/src/content/docs/inference/reference/commands.md` | ptah | 14 | 0 |
| c06 `docs/site/src/content/docs/databases/clickhouse.md` | ptah | 5 | 0 |
| c07 `docs/stargz-estargz.md` | historical | 3 | 1 |
| c08 `README.md` | historical | 10 | 0 |
| c09 `RELEASE.md` | historical | 4 | 0 |
| c10 `runtime/v2/README.md` | historical | 3 | 0 |
| c11 `whatsnew-2.1.txt` | historical | 12 | 0 |
| c12 `CHANGELOG.md` | historical | 17 | 0 |

| Category | Confirmation defects | Detected |
| --- | ---: | ---: |
| empty_framing | 20 | 0 |
| formulaic_transitions | 0 | 0 |
| needless_complexity | 0 | 0 |
| needless_repetition | 7 | 1 |
| unjustified_intensifiers | 3 | 0 |
| vague_claims | 34 | 0 |
| wordiness | 14 | 0 |

## Runtime and replay identity

Before: `b3ffaae00dbc62dfea1aad828e2cc8ea42f20614`. Final measured code:
`0973ff633aee53bc9fc3592f80803806bf2d62e6`. Builds use `CGO_ENABLED=0`.
The reports include source for authorized offline verification; ordinary
saved reports still omit it by default. The runtime makes no model calls.

The [resource amendment](resource-amendment.md) records a correction made
after opening confirmation output. Initial reports remain under
`reports/*/initial`. All findings, assessments, source documents and gates
are identical to the corrected replay. The one initial budget abstention
on the exposed schema-commands page is gone. No final scan abstains.

| Set | Profile | Wall seconds before → after | Peak MiB before → after |
| --- | --- | ---: | ---: |
| Original development | technical | 2.535 → 2.126 | 186.2 → 187.2 |
| Original development | strict | 1.620 → 1.505 | 202.7 → 200.3 |
| Exposed repetition targets | technical | 0.589 → 0.541 | 133.9 → 116.5 |
| Exposed repetition targets | strict | 0.582 → 0.534 | 121.0 → 136.9 |
| Exposed framing targets | technical | 0.791 → 0.771 | 143.5 → 130.2 |
| Exposed framing targets | strict | 0.869 → 0.766 | 136.9 → 148.4 |
| New complete-page confirmation | technical | 0.706 → 0.608 | 115.9 → 119.8 |
| New complete-page confirmation | strict | 0.671 → 0.610 | 119.5 → 118.2 |

These are single macOS arm64 runs, with CPU time and host recorded in
`runs.json`. They are operational measurements, not evidence of a speedup
or the separate two-vCPU Linux performance target.

## Reproduce

From the repository root:

```sh
python3 research/reviews/2026-09-17-local-repetition/tools/render.py
python3 research/reviews/2026-09-17-local-repetition/tools/test_evidence.py
```

The renderer validates source hashes, complete reviews, configuration and
code identity, all diagnostic deltas, source-bound credits and the resource
correction before producing this page. To repeat a scan, build the named
revision and use `tools/measure.py --binary PATH --set SET --output NEW_DIR`.
The original [protocol](protocol.md) and both freeze manifests are unchanged.

Next work must address the remaining semantic restatements and contextual
framing/assurance gaps. These confirmation pages are now exposed development
material; further tuning requires another unexposed confirmation sample.
