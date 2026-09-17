# Before and after

Counts describe agreement with the same assistant's frozen judgments on these
pages. They are not population recall or authorship accuracy.

| Set | Profile | Defects | Before | After | Findings before / after | Added / removed |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| development | technical | 57 | 12 | 12 | 137 / 126 | 1 / 12 |
| development | strict | 57 | 12 | 12 | 156 / 145 | 1 / 12 |
| confirmation | technical | 5 | 0 | 0 | 38 / 37 | 0 / 1 |
| confirmation | strict | 5 | 0 | 0 | 44 / 43 | 0 / 1 |

## Repetition category

The development denominator above spans all editorial categories. Restricting
it to the frozen needless-repetition labels gives the following result.

| Set | Profile | Repetition defects | Before | After |
| --- | --- | ---: | ---: | ---: |
| development | technical | 6 | 1 | 1 |
| development | strict | 6 | 1 | 1 |
| confirmation | technical | 5 | 0 | 0 |
| confirmation | strict | 5 | 0 | 0 |

The gain is p014-d01; the loss is p004-d01. Development recall is unchanged.
The glossary paraphrase p012-d01 remains missed. The six confirmation pages
contain five annotated defects, five uncertain judgments and fourteen controls.
None of their five defects is detected before or after.

All scans complete with passing gates and no abstentions. The index and gate
thresholds are unchanged. Declining finding counts do not imply improved recall.

| Page | Cohort / selection | Defects | Before | After |
| --- | --- | ---: | ---: | ---: |
| p001 | ptah / sample | 4 | 0 | 0 |
| p002 | ptah / sample | 5 | 3 | 3 |
| p003 | ptah / sample | 2 | 1 | 1 |
| p004 | ptah / sample | 4 | 1 | 0 |
| p005 | ptah / sample | 0 | 0 | 0 |
| p006 | ptah / sample | 1 | 1 | 1 |
| p007 | ptah / sample | 4 | 0 | 0 |
| p008 | ptah / sample | 4 | 1 | 1 |
| p009 | ptah / sample | 1 | 0 | 0 |
| p010 | ptah / sample | 1 | 1 | 1 |
| p011 | ptah / sample | 0 | 0 | 0 |
| p012 | ptah / sample | 1 | 0 | 0 |
| p013 | historical / sample | 1 | 0 | 0 |
| p014 | historical / sample | 1 | 0 | 1 |
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
| c01 | ptah / confirmation | 0 | 0 | 0 |
| c02 | ptah / confirmation | 0 | 0 | 0 |
| c03 | ptah / confirmation | 0 | 0 | 0 |
| c04 | historical / confirmation | 2 | 0 | 0 |
| c05 | historical / confirmation | 0 | 0 | 0 |
| c06 | historical / confirmation | 3 | 0 | 0 |

Per-page event credit is identical between profiles.

## Resource observations

One fresh process per scan, including model startup and reports. These are local
observations, not an isolated performance comparison. Other development checks
ran during the initial before scans; no other task tests ran during the bounded
after scans. Both before sets preceded both bounded after sets; development
preceded confirmation and technical preceded strict within each phase.

| Set / profile | Wall seconds before / after | CPU seconds before / after | Peak RSS MiB before / after |
| --- | ---: | ---: | ---: |
| development / technical | 5.331 / 2.330 | 2.497 / 1.764 | 207.8 / 199.1 |
| development / strict | 4.527 / 1.449 | 2.538 / 1.730 | 200.3 / 196.9 |
| confirmation / technical | 0.717 / 0.504 | 0.868 / 0.734 | 128.8 / 114.0 |
| confirmation / strict | 0.545 / 0.504 | 0.780 / 0.735 | 126.0 / 116.8 |
