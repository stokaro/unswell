# Before and after

These counts use the same assistant's frozen editorial labels. They are not
population recall or authorship measurements.

| Set | Profile | Defects | Before | After | Findings before / after | Added / removed |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| development | technical | 57 | 5 | 12 | 130 / 137 | 7 / 0 |
| development | strict | 57 | 5 | 12 | 149 / 156 | 7 / 0 |
| confirmation | technical | 11 | 1 | 1 | 103 / 103 | 0 / 0 |
| confirmation | strict | 11 | 1 | 1 | 129 / 129 | 0 / 0 |

Both profiles retain passing gates for all pages. No rule abstentions or
operational errors occurred. The earlier dispositions of retained development
findings remain in force. Confirmation credit evaluates only the three target
families; other existing diagnostics were not assigned new quality labels.

| Page | Cohort / selection | Defects | Before | After |
| --- | --- | ---: | ---: | ---: |
| p001 | ptah / sample | 4 | 0 | 0 |
| p002 | ptah / sample | 5 | 0 | 3 |
| p003 | ptah / sample | 2 | 0 | 1 |
| p004 | ptah / sample | 4 | 1 | 1 |
| p005 | ptah / sample | 0 | 0 | 0 |
| p006 | ptah / sample | 1 | 0 | 1 |
| p007 | ptah / sample | 4 | 0 | 0 |
| p008 | ptah / sample | 4 | 0 | 1 |
| p009 | ptah / sample | 1 | 0 | 0 |
| p010 | ptah / sample | 1 | 0 | 1 |
| p011 | ptah / sample | 0 | 0 | 0 |
| p012 | ptah / sample | 1 | 0 | 0 |
| p013 | historical / sample | 1 | 0 | 0 |
| p014 | historical / sample | 1 | 0 | 0 |
| p015 | historical / sample | 3 | 0 | 0 |
| p016 | historical / sample | 3 | 0 | 0 |
| p017 | historical / sample | 3 | 0 | 0 |
| p018 | historical / sample | 2 | 0 | 0 |
| p019 | historical / sample | 2 | 0 | 0 |
| p020 | historical / sample | 1 | 0 | 0 |
| p021 | historical / sample | 1 | 0 | 0 |
| p022 | historical / sample | 1 | 0 | 0 |
| p023 | historical / sample | 1 | 0 | 0 |
| p024 | historical / sample | 2 | 0 | 0 |
| p025 | ptah / exposed_anchor | 7 | 2 | 2 |
| p026 | ptah / exposed_anchor | 2 | 2 | 2 |
| c01 | ptah / confirmation | 1 | 0 | 0 |
| c02 | ptah / confirmation | 2 | 1 | 1 |
| c03 | ptah / confirmation | 1 | 0 | 0 |
| c04 | ptah / confirmation | 0 | 0 | 0 |
| c05 | ptah / confirmation | 2 | 0 | 0 |
| c06 | ptah / confirmation | 3 | 0 | 0 |
| c07 | historical / confirmation | 0 | 0 | 0 |
| c08 | historical / confirmation | 2 | 0 | 0 |
| c09 | historical / confirmation | 0 | 0 | 0 |

Per-page credit is identical between profiles.

## Resource observations

One fresh CLI process per cell, on the host recorded in `runs.json`.
These are observations from this replay, not a stable comparative benchmark.
Order was development-after, confirmation-before, confirmation-after,
development-before. Cold start and unrelated host activity are not controlled.

| Set / profile | Wall seconds before / after | CPU seconds before / after | Peak RSS MiB before / after |
| --- | ---: | ---: | ---: |
| development / technical | 1.420 / 2.458 | 1.689 / 1.727 | 203.2 / 190.9 |
| development / strict | 1.431 / 1.443 | 1.685 / 1.749 | 206.8 / 185.7 |
| confirmation / technical | 0.726 / 0.677 | 0.896 / 0.906 | 149.0 / 134.6 |
| confirmation / strict | 0.672 / 0.680 | 0.890 / 0.913 | 140.6 / 139.2 |
