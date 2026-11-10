# #349 current-main replay and second editorial review

## Correction after comparison with #441 and #442

The seven findings below were new only relative to the older November 7
review at runtime 5dd7619. PRs #441 and #442 had already introduced and
reviewed all seven before this replay. The archived #442 and #504 reports
contain identical technical and strict finding streams on the same 36 pages.
This replay adds no detector finding. See reconciliation.json and
reconcile.py for the checked comparison.

The #442 assistant review counted all five quality-claim cases as FULL:
whole 28/57, context 16/46, historical confirmation 4/20. This independent
second assistant review counts three of those whole-page cases as PARTIAL:
whole 25/57, context 16/46, confirmation 4/20. The difference is editorial
judgment on unchanged diagnostics, not a runtime regression. Neither
review meets the #349 targets.

The exact-main replay at 3acba5e produced 162 technical findings and 183
strict findings on the same 36 full pages. Every one of the 176 strict
findings from the November 7 review remains at the same original byte
location, with unchanged normalized diagnostic content. Twenty-three finding
IDs changed only because a rule version or fingerprint changed. Seven
findings are new relative to that older review and each hits a distinct event
it had marked missed.

The assistant's complete-source review accepts all seven findings. Four
fully diagnose their frozen events; three only partially diagnose them:

| Cohort | Full events / total | Required at 80% | Accepted technical findings / total | Required at 85% |
| --- | ---: | ---: | ---: | ---: |
| Whole pages | 25 / 57 | 46 | 82 / 126 | 108 |
| Context pages | 16 / 46 | 37 | 21 / 27 | 23 |
| Historical confirmation | 4 / 20 | 16 | 6 / 9 | 8 |

The strict profile has the same full-event counts. Its accepted findings are
82 / 145 whole, 21 / 28 context, and 6 / 10 confirmation. Both profiles
remain below the qualification targets. These numbers describe the second
review, not a product gain since #442.

Run python3 research/reviews/2026-11-10-current-main-editorial-delta/verify.py
to validate summary.json, or use --write to regenerate it. The source
reviews for the seven additions are in new-review.json; prior judgments
remain in the November 7 review. PROTOCOL.md describes the transfer rule,
frozen inputs, and limitations.
