# Origin pilot, 2026-09-11

The first fit of the origin task on the pilot generation run of
[#154](https://github.com/stokaro/unswell/issues/154), produced by
`scripts/origin-experiment.sh` with its defaults. `digests.json` names the
tool commit and the digest and size of every artifact, including the
candidate artifact that stays outside the repository; `union.json` and
`union-plan.json` reproduce it from the acquisition work directory.

## Corpus

`union.json` joins every controlled source with the comment-role sources of
the nineteen repositories the sixty tasks came from. It keeps at most fifteen
sources per historical checkout in ID order, drops files above 100 KiB, and
extracts paragraphs only. Without the byte limit, one bundle of comments was a
quarter of the corpus. It also exhausted the measurement budget of its engine
run.

| Count | Value |
| --- | --- |
| Sources | 525: 240 controlled, 285 historical |
| Paragraph units | 4,342: 145 controlled, 4,197 historical |
| Groups | 19 connected components, one per repository, pinned to the dataset's partitions |
| Training | 2,441 fitted rows: 26 `endpoint_generated`, 2,415 `human_snapshot`; 44 polished responses excluded |
| Calibration | 169 rows: 2 endpoints, 167 snapshots; 4 polished excluded |
| Development | 76 candidates; 4 polished excluded |
| Confirmation | 1,608 candidates; 27 polished excluded |

Only 145 of the 240 responses carry a paragraph unit, because most `generate`
responses came back as structured Markdown; the pilot record of #154 states
that limit. A group is a repository, so a repository's controlled responses
sit in the same partition as its historical comments.

## Fit

`model.json` is a logistic fit with isotonic calibration. It uses fourteen
prepared paragraph features from four families: structure, lexical, part of
speech, and readability. The repetition family stayed out. It abstains on a
paragraph of one sentence, and most comments are one sentence. Both plans
froze the threshold at 0.5 before any partition was scored.

## Result

| Partition | Eligible | Covered | Endpoints | Recall | False positives | Brier | Constant Brier | ECE |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Development | 72 | 20 (27.8%) | 2 | 0 | 0 | 0.1002 | 0.0980 | 0.093 |
| Confirmation | 1,581 | 1,572 (99.4%) | 36 | 0 | 0 | 0.0226 | 0.0225 | 0.017 |

The model separates nothing. Every isotonic response is at or below 0.027.
No unit reaches the threshold, and all 38 endpoints of the two partitions are
missed. The Brier score equals the constant baseline of the training
prevalence. Before calibration, the logistic response averages 0.011 on the
controlled units and 0.006 on the historical units of the confirmation
partition. The calibration partition holds two endpoints. That is why the
isotonic map saturates near zero and abstains on most of the development
partition. The confirmation partition has ten provenance components.
`evaluation-final_test.json` carries the per-group and per-stratum metrics.

## What this run does not show

The fit saw 26 endpoints. That is not a test of the origin channel. It is the
first run of the pipeline on real provenance labels, and it sizes the next
one. Every partition needs endpoints in the hundreds before a recall or a
false-positive rate says anything. That takes more generation runs, not a
different fit.

Every endpoint is a Markdown document and every negative a code comment, so a
model that separated them could be reading the source format. The strata by
prompt condition and generator family that #179 asks for need a join with the
generation records; the evaluation stratifies by role, language, words, and
declared origin only. One family (`claude-opus-5` through session agents) ran;
every other family is untested. The generator may have seen the historical
text, and the generation records flag verbatim overlap. Nothing here is an
authorship verdict or a share of text written by a tool.
