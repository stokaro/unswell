# Rhetorical rhythm in technical explanations

The PostgreSQL page exposed three missing constructions: an embedded page-as-map
introduction, an abstract boundary-versus-feature judgment, and an offline
declaration-as-evidence identity. Existing rules now report each with an
explanation specific to the construction. Repeated negative-action/positive-action
pairs also receive complete clause evidence through the existing reframing rule.

This is a bounded contribution to [#349](https://github.com/stokaro/unswell/issues/349).
It does not qualify broad contextual recall or determine authorship. The supplied
page, all original reference events, and the additional public snapshot are
exposed development material.

## Implementation and controls

The implementation uses the existing extraction, NLP, observation, and reporting
pipeline. It adds no model or external request. Versions advance to
`syntax.repeated-reframing` 2, `filler.document-metadiscourse` 3, and
`filler.evaluative-closure` 11. Severity, weights, gates, and thresholds are unchanged.

A single negative/positive explanation remains below the reframing allowance.
Semicolons provide clause boundaries; punctuation alone is not an editing
diagnosis. Action pairs require linked subjects and exclude quotations,
attribution, protected construction words, questions, and structural crossings.
They request review of a repeated shape, not deletion of necessary distinctions.

The page-as-map observer retains the topic and reference links. Literal geographic
and address-space maps remain controls. Both observational rules remain unscored.
The closure warning asks for concrete behavior while preserving its conditions
and limits. Qualified evidence, measurement methods, concrete actors, and
conditions before or after an abstract contrast remain controls.

The concurrent evidence-commentary change recognizes an explicit discourse subject
contrasting verification with assertion. Its suggestion preserves the actual
checks and results, including distinctions between measured and merely described
coverage. A bare `it` requires an information antecedent in the same sentence.

## Complete-source comparison

The evaluation keeps all 36 original sources, 4,549 original units, 804 frozen
criticism judgments, and 123 source-bound reference events. Each profile completes
with 10,553 available engine assessments, no operational error, and no abstention.
The larger assessment count reflects multiple engine scopes; it does not replace
the original unit denominator. Calibration remains unavailable.

| Profile | Before | After | Added | Changed explanations | Removed |
| --- | ---: | ---: | ---: | ---: | ---: |
| technical | 169 | 173 | 4 | 1 | 0 |
| strict | 190 | 194 | 4 | 1 | 0 |

All findings from unaffected rules are identical. Version-dependent finding
identities are excluded only when comparing the three changed rules. The four
additions occur in two sources; three are in the original PostgreSQL page.

| Added clause | Rule | Assistant disposition |
| --- | --- | --- |
| “what it covers is measured rather than asserted” | evaluative closure | State the checks and coverage directly; retain executed-versus-described scope |
| “this page is the map of what Ptah manages…” | document metadiscourse | Optional direct introduction; retain the subject, lifecycle, and links |
| “That is a boundary rather than a missing feature” | evaluative closure | Retain the refresh limitation and its concrete schema/data-operation rationale |
| “Offline, the declaration is the evidence” | evaluative closure | State how offline rendering derives capability; retain declaration-versus-observation limits |

Reviewer: Codex assistant under ADR 0041. The [review record](review.json) binds
each decision to the full source section and states the technical meaning that
an edit must preserve. Four accepted additions in two exposed sources are a small,
correlated sample, not an estimate establishing 85% population precision.
The changed explanation retains its existing accepted label.

Two additions address original events `whole/p025-d01` and
`confirmation/c02-r03`: their semantic judgments concern the complete rhetorical
clauses, not mere span overlap. The record does not recompute full recall on all
123 events. Original uncertain and rejected labels remain unchanged.
The 80% full-event-recall and 85% accepted-delivered-criticism requirements on the
required complete cohorts, followed by untouched confirmation, remain open.

## Additional public snapshot

The [current PostgreSQL source](https://github.com/stokaro/ptah/blob/e9d996d560bf90cfcd035d808e3515e798509a8a/docs/site/src/content/docs/databases/postgresql.md)
is pinned separately: 94,644 bytes, SHA-256
`dd35927e010c4db4ea1b554a6dc2c25fda8ab49eb850b2a0374af869dcb3732d`.
It is not substituted into the original evaluation.

| Profile | Before | After | Added |
| --- | ---: | ---: | ---: |
| technical | 54 | 57 | 3 |
| strict | 61 | 64 | 3 |

The three additions are the page-as-map clause, boundary judgment, and offline
identity. The complete page has 1,046 available assessments in each scan and no
error or abstention. No action-reframing group is added on this page: the example
negative/positive pair is ordinary in isolation, and tables and headings reset
the observer's prose window.

This change does not diagnose every suspected symptom in the supplied page.
The descriptive “What Ptah manages…” summary, a necessary negative limitation,
and semicolon usage do not individually establish an editing defect. General
argument rhythm across sections remains outside these bounded token constructions.

## Evidence and reproduction

[Measurement identities](measurement.json), the [full comparison](comparison.json),
and compressed complete before/after reports preserve source bytes, all findings,
related locations, assessments, and unchanged denominators. The validated baseline
build uses the recorded working tree; its 469 production Go files were checked
against base commit `ee1f88695fa4679b4727a180380cd2ba589e5c8e` with no differences.
The candidate identifies a modified working tree and binds its runtime source
hashes and binary hash. These local records do not certify a merged build or a
playground deployment.

The [protocol](protocol.json) and [evidence hypothesis](evidence-protocol.json)
were recorded before their rule edits. A [dated amendment](final-amendment.json)
records lint-driven helper extraction and the condition-prefix regression guard.
Earlier measurements and failed checks remain in the local research record.
The public-source identity corrects the protocol's blob-versus-commit URL; its
content hash did not change.

Build baseline and candidate CLI binaries in separate worktrees, then run:

```sh
python3 -B research/reviews/2026-10-04-rhetorical-rhythm/reproduce.py \
  --before /absolute/baseline/unswell \
  --after /absolute/candidate/unswell \
  --output /new/measurement
```

The script checks bundled report hashes, restores all original source bytes,
refuses an existing output directory, scans both profiles, and verifies the
complete diagnostic comparison. It performs no network or model call. Public
blackbox tests and two CLI fixtures cover the new constructions, technical
controls, complete related clauses, and Unicode/BOM/CRLF coordinates in JSON
and SARIF.
