# Source-function v1: exposed diagnoses, unsafe edit

The exact offline candidate and protocol were committed in `44732a7` before
the first run. The candidate used all 920 prepared sentence assessments on
the nine complete pinned pages. It emitted 18 source-bound diagnoses; every
quote resolves to original UTF-8 bytes, and no target overlaps an excluded
region. The raw output is `candidate-output.json` (SHA-256
`32c66952272d7bf0a5bbeaab9e1fc5a2af17a9d53b6adb0e3e0206ceeef18a78`).
The complete same-assistant disposition is `candidate-review.json` (SHA-256
`00cbc5c3fa79388cf7d6e1f3fd252597238460d140fccc592164677050671a72`).

| Measure on these exposed development pages | Result |
| --- | ---: |
| Valid complete pages | 9/9 |
| Frozen clear positives located | 17/17 |
| FULL positive diagnoses | 15/17 |
| Partial positive diagnoses | 2/17 |
| Accepted delivered diagnoses | 17/18 |
| Optional delivered diagnoses | 1/18 |
| `retain` control hits | 0 |
| Proposed deletions safe as supplied | 1/5 |
| Proposed deletions needing format cleanup | 3/5 |
| Unsafe proposed deletions | 1/5 |
| Diagnoses without a supplied edit | 13/18 |

The two partial event diagnoses are the TNN conversion-tools sentence: the
candidate marks `a wealth of` but leaves the unsupported `easily` claim, and
the TNN business-and-praise sentence: it marks `widely praised` but leaves
the unspecified `various major businesses` claim. The extra `rich` diagnosis
in TNN's demo paragraph is optional: the adjacent three platform examples
give the description some support. All other source-only positives are
credited FULL by the same assistant who prepared the labels. These judgments
are development evidence, not independent confirmation.

The unsafe edit is decisive. The candidate proposes deleting DexHunter's
redundant item `4)` as a sentence but does not renumber the following item
`5)`. The page would retain a broken ordered list. Three other proposed
deletions leave leading whitespace before the following sentence, so they
also need cleanup before application. The registered no-unsafe-edit condition
fails even though the exposed recall and accepted-delivery counts exceed
their directional thresholds. **Retire this exact v1 candidate.** Do not
repair it and rescore these nine pages as if they were transfer evidence.

The unchanged CLI still delivered ten nonaccepted or optional findings on
these pages. Candidate-only selectivity does not establish the accepted
fraction of a combined product stream. No runtime rule, frozen 123-event
cohort, prospective page, or product artifact changed. A replacement must
model whole-document edit structure, including list numbering and adjacent
whitespace, and then face newly selected page-disjoint source-only controls
before a complete #349 cohort replay. This packet cannot close #349.
