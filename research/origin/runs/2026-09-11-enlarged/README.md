# Origin fit on the enlarged cohort, 2026-09-11

The origin task fitted again after the
[second generation run](../../../generation/runs/2026-09-11-run2/README.md)
joined the controlled cohort. It is the run the
[pilot](../2026-09-11-pilot/README.md) named as the next one: the same
driver, `scripts/origin-experiment.sh`, with the tasks and generation
records of both runs joined into one file each, the id
`origin-enlarged-v1`, and the same defaults. `digests.json` names the
tool commit, `e1b2c99`, and the digest and size of every artifact,
including the candidate artifact that stays outside the repository.
`union.json` and `union-plan.json` reproduce it from the acquisition work
directory. Nothing here qualifies an origin model.

## Corpus

`union.json` joins every controlled source with the comment-role sources
of the 25 repositories the 200 tasks came from. It keeps at most fifteen
sources per historical checkout in ID order, drops files above 100 KiB,
and extracts paragraphs only.

| Count | Value |
| --- | --- |
| Sources | 1,162: 800 controlled, 362 historical |
| Paragraph units | 5,608: 529 controlled, 5,079 historical |
| Groups | 25 connected components, one per repository, pinned to the dataset's partitions |
| Training | 1,414 fitted rows: 132 `endpoint_generated`, 1,282 `human_snapshot`; 162 polished responses excluded |
| Calibration | 698 rows: 24 endpoints, 674 snapshots; 28 polished excluded |
| Development | 518 candidates in 2 components; 34 polished excluded |
| Confirmation | 2,788 candidates in 12 components; 67 polished excluded |

511 of the 800 responses carry a paragraph unit, against 145 of 240 in
the pilot. The second run's responses came back as plain paragraphs more
often than the pilot's structured Markdown.

## Fit

`model.json` is a logistic fit with isotonic calibration on the same
fourteen prepared paragraph features as the pilot: structure, lexical,
part of speech, and readability. Both plans froze the threshold at 0.5
before any partition was scored.

## Result

| Partition | Eligible | Covered | Endpoints | Recall | False positives | Brier | Constant Brier | ECE |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Development | 484 | 484 (100%) | 24 | 0 | 0 | 0.0451 | 0.0490 | 0.017 |
| Confirmation | 2,721 | 2,721 (100%) | 58 | 0 | 0 | 0.0210 | 0.0261 | 0.017 |

The model separates nothing. The highest calibrated response on the
confirmation partition is 0.110. No unit reaches the threshold, and all 82
endpoints of the two partitions are missed. The Brier score sits below
the constant baseline of the training prevalence, inside its interval.
The intervals resample the twelve confirmation components 10,000 times
with the fixed seed.

| Confirmation metric | Estimate | Interval |
| --- | --- | --- |
| Coverage | 100% | 100% to 100% |
| Recall | 0 | 0 to 0 |
| False-positive rate | 0 | none: no component flags anything |
| Brier | 0.0210 | 0.0090 to 0.0523, against a constant of 0.0261 |

The strata come from the generation records of both runs. All 58
endpoints of the confirmation partition are `generate` responses of one
family, 28 under the `neutral` prompt and 30 under `plain`; 44 hold
fewer than 20 words and 14 hold 20 to 49. Every one is missed. The
negatives span ten source languages. The `polish` arm has no stratum,
because its responses are unresolved under the origin profile and
counted as exclusions.

## What this run does not show

The fit saw 132 endpoints, five times the pilot, and the confirmation
partition holds 58 against 36. The result is the same nothing: with
structural features on paragraphs of one to fifty words, the origin task
does not separate generated comments from historical ones. That says
nothing about constructions; the
[pattern tables](../../../acquisition/runs/2026-09-11-roles/README.md)
and the [reviews](../../../reviews/README.md) carry that question. One
family ran through session agents; every other family is untested. The
generator may have seen the historical text, and the generation records
flag verbatim overlap. Nothing here is an authorship verdict or a share
of text written by a tool.
