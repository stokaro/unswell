# Origin pilot with documentation negatives, 2026-09-11

The second fit of the origin task on the pilot generation run of
[#154](https://github.com/stokaro/unswell/issues/154). It repeats the
[first pilot](../2026-09-11-pilot/README.md) with one change: the negatives
are the documentation, readme, and release-note paragraphs of the task
repositories instead of their comment paragraphs. Every endpoint is a
Markdown document, so these negatives share the format of the positives.
`scripts/origin-experiment.sh` produced the run with the three
documentation roles as `--negative-role` values. Every other setting is the
pilot's. `digests.json` names the tool commit and the digest and size of
every artifact, including the candidate artifact that stays outside the
repository. `union.json` and `union-plan.json` reproduce it from the
acquisition work directory.

## Corpus

`union.json` joins every controlled source with the documentation-role
sources of the nineteen repositories the sixty tasks came from. It keeps
at most fifteen sources per historical checkout in ID order. Files above
100 KiB stay out, and so do the sources the pattern measurement could not
analyze. Only paragraphs are extracted.

| Count | Value |
| --- | --- |
| Sources | 339: 240 controlled, 99 historical (59 documentation, 33 readme, 7 release notes) |
| Paragraph units | 971: 145 controlled, 826 historical; the median historical paragraph has 11 words |
| Groups | 19 connected components, one per repository, pinned to the dataset's partitions |
| Training | 232 fitted rows: 26 `endpoint_generated`, 206 `human_snapshot`; 44 polished responses excluded |
| Calibration | 66 rows: 2 endpoints, 64 snapshots; 4 polished excluded |
| Development | 29 candidates; 4 polished excluded |
| Confirmation | 596 candidates; 27 polished excluded |

The endpoints are the same 145 paragraphs as in the first pilot. The
negatives are a fifth as many as the comment negatives, because the task
repositories hold fewer documentation files than comments.

## Fit

`model.json` is a logistic fit with isotonic calibration on the same
fourteen prepared paragraph features as the first pilot. Both plans froze
the threshold at 0.5 before any partition was scored.

## Result

| Partition | Eligible | Covered | Endpoints | Recall | False positives | Brier | Constant Brier | ECE |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Development | 25 | 21 (84.0%) | 2 | 0 | 0 | 0.0955 | 0.0865 | 0.089 |
| Confirmation | 569 | 409 (71.9%) | 36 | 0 | 0 | 0.0765 | 0.0733 | 0.067 |

The model separates nothing here either. Every isotonic response is at or
below 0.046. No unit reaches the threshold, and all 38 endpoints of the two
partitions are missed. Before calibration, the logistic response averages
0.113 on the controlled units and 0.118 on the historical units of the
confirmation partition. The raw score leans the wrong way. The calibration
partition holds two endpoints. The isotonic map abstains on 169
confirmation units whose score falls outside its range: 156 negatives and
4 endpoints. The confirmation partition has ten provenance components.
`evaluation-final_test.json` carries the per-group and per-stratum metrics.

| Confirmation metric | Estimate | Interval |
| --- | --- | --- |
| Coverage | 71.9% | 64.0% to 80.6% |
| Recall | 0 | 0 to 0 |
| False-positive rate | 0 | none: no component flags anything |
| Brier | 0.0765 | 0.0179 to 0.2228, against a constant of 0.0733 |

The strata by role show where the negatives came from: 236 documentation,
168 readme, and 129 release-note paragraphs are eligible, against the 36
comment-role endpoints. All 36 endpoints are `generate` responses of one
family, 16 under the `neutral` prompt and 20 under `plain`; every one is
missed or abstained.

## What this run does not show

The format confound of the first pilot is gone: the negatives are Markdown
paragraphs like the endpoints, and the model still finds nothing. That
closes one explanation of the first result without adding a positive one.
The fit still saw 26 endpoints, so every partition needs endpoints in the
hundreds before a recall or a false-positive rate says anything, and that
takes more generation runs. The endpoints keep the `comment` role of their
tasks while the negatives carry documentation roles, so the role strata
compare nothing across the classes. One family ran; every other family is
untested. The generator may have seen the historical text, and the
generation records flag verbatim overlap. Nothing here is an authorship
verdict or a share of text written by a tool.
