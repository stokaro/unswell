# Origin-channel experiments

This directory records experiments on the origin channel of
[ADR 0035](../../docs/adr/0035-origin-channel.md) under the pattern protocol
of [ADR 0036](../../docs/adr/0036-llm-pattern-evidence.md). The task is
binary: is a unit the endpoint of a generation record? The labels are
provenance the corpus declares, never a judgment of the text, and never a
human label. [#179](https://github.com/stokaro/unswell/issues/179) is the
mirrored issue; its original, #58, stays on hold with the human-labeled corpus.

## How a run works

`scripts/origin-experiment.sh` orders the research commands and records the
digests. `corpus dataset union` joins the controlled cohort with the
historical shards of the repositories the generation tasks came from. `plan`
and `extract` build that corpus from the acquisition work directory.
`corpus train --labels provenance` fits the origin task. `predict` freezes
the selection and confirmation partitions, and `corpus evaluate --labels
provenance` scores them against the same labels.

```sh
bash scripts/origin-experiment.sh --record research/origin/runs/<run>
```

`--negative-role` picks the roles of the historical negatives; the default
is `comment`, the role the tasks came from, and the documentation roles give
negatives in the format of the endpoints.

A positive is a complete `generate` response. Its source sits in the
controlled cohort with origin `generated`, document scope, and a generation
record. A negative is a unit of a historical source with origin `human` or
`unknown`, a dated snapshot from before the boundary. Polished responses,
contemporary text, and mixed origins stay unresolved with a reason. They count
as exclusions, and the detector never labels them. The rule is the frozen text
of `annotation.OriginProfile`, and its digest is the profile of every artifact
a run produces.

## What a run does not establish

A run fits one family of generator on one prompt set, so every other family
is untested. The generator may have seen the historical text, and the
generation records flag verbatim overlap. Positives are Markdown documents
while the historical negatives of the same role are code comments, so a
separation could rest on the source format rather than on the writing. An
estimate stays experimental and outside every gate: the artifact records
`human_corpus: not_qualified`, and a pack built from it declares
`experimental`. No number here is an authorship verdict or a share of text
written by a tool.

## Runs

| Run | What it covers | Record |
| --- | --- | --- |
| 2026-09-11 pilot | First fit on the pilot generation run of #154: 145 controlled paragraphs against the comment paragraphs of the nineteen task repositories, structural features, threshold frozen at 0.5 | [runs/2026-09-11-pilot](runs/2026-09-11-pilot/README.md) |
| 2026-09-11 documentation | The same endpoints against the documentation, readme, and release-note paragraphs of the same repositories, so the negatives share the endpoints' format | [runs/2026-09-11-documentation](runs/2026-09-11-documentation/README.md) |
