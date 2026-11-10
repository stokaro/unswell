# #349 current-main replay: seven new diagnoses

The exact-main replay at 3acba5e produced 162 technical findings and 183
strict findings on the same 36 full pages. Every one of the 176 strict
findings from the November 7 review remains at the same original byte
location, with unchanged normalized diagnostic content. Twenty-three finding
IDs changed only because a rule version or fingerprint changed. Seven
findings are genuinely new and each hits a distinct previously missed event.

The assistant's complete-source review accepts all seven findings. Four
fully diagnose their frozen events; three only partially diagnose them:

| Cohort | Full events / total | Required at 80% | Accepted technical findings / total | Required at 85% |
| --- | ---: | ---: | ---: | ---: |
| Whole pages | 25 / 57 | 46 | 82 / 126 | 108 |
| Context pages | 16 / 46 | 37 | 21 / 27 | 23 |
| Historical confirmation | 4 / 20 | 16 | 6 / 9 | 8 |

The strict profile has the same full-event counts. Its accepted findings are
82 / 145 whole, 21 / 28 context, and 6 / 10 confirmation. Both profiles
remain below the qualification targets. The new findings demonstrate a
measurable source-bound gain, not a completed solution to #349.

Run python3 research/reviews/2026-11-10-current-main-editorial-delta/verify.py
to validate summary.json, or use --write to regenerate it. The source
reviews for the seven additions are in new-review.json; prior judgments
remain in the November 7 review. PROTOCOL.md describes the transfer rule,
frozen inputs, and limitations.
