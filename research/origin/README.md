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

The [RACE source-matched transfer screen](runs/2026-11-09-race-transfer/README.md)
ran the published, unchanged RACE encoder checkpoint on 60 ARB human sources
and 240 paired rewrites. All 300 documents scored, but the registered advance
rule failed: pooled AUROC was 0.692, paired-win fraction 0.683, and the `owt`
domain reversed direction (paired-win fraction 0.150). The checkpoint is
retired as an Unswell origin candidate on this evidence. This does not change
the editorial #349 result.

The [TextSight fixed-threshold transfer screen](runs/2026-11-08-textsight-transfer/README.md)
tested the existing interpretable ensemble and richness score on a pinned
September 2026 generation corpus. The ensemble flagged 30/56 AI passages but
also 41/338 human passages; richness flagged only 2/56 AI passages at 2/338
human false flags. Both registered advance rules fail. This is origin research,
not an editorial #349 result.

The [November 2026 lexical richness and entropy screen](runs/2026-11-07-richness-screen/README.md)
uses saved complete-document generation and editing responses. Its two
directional feature checks pass on exposed data; it does not qualify an origin
detector or an editorial finding.

The [AITDNA user-study screen](runs/2026-11-07-aitdna-screen/README.md)
checks the same features against 362 documents labeled by human-only versus
AI-assisted writing setting. Its registered direction screen passes, but the
post-result split is weak for AI-assisted documents labeled `User` under the
dataset's document notion. It remains origin research, not a detector claim.

The [AIWritingEffect paired academic-prose screen](runs/2026-11-08-aiwritingeffect-screen/README.md)
tested those unchanged features on a pinned external rewrite corpus. Its
paragraph-level feasibility rule failed because most passages were shorter
than the fixed 100-token input and low-FPR recall was 6.0%. A post-result
document-level sensitivity showed a strong direction on the same papers, but
has no stable low-FPR operating point or independent confirmation. Neither
result changes the editorial #349 gate.

The [ARB source-matched transfer screen](runs/2026-11-08-arb-transfer/README.md)
tested the unchanged features on 1,800 source groups across three new
domains and four open-weight generators. Its preregistered H2L richness
feasibility rule failed: test AUROC was 0.632 and TPR 4.5% at a threshold
calibrated for 1% human false positives. Direct generation reversed the
expected richness direction. This limits the earlier positive academic-paper
contrast and does not qualify an origin detector or editorial #349 finding.

The [Heterogeneous AI Spans transfer screen](runs/2026-11-08-heterogeneous-spans-transfer/README.md)
tests unchanged lexical features on 4,400 source-matched mixed documents and
human controls. Oracle-located replacement spans show richness AUROC 0.652,
but only 5.65% TPR at a train-human 1% FPR threshold; complete mixed documents
show AUROC 0.608. Both registered feasibility rules fail. Neither experiment
locates an unknown AI span or supplies an editorial finding.

| Run | What it covers | Record |
| --- | --- | --- |
| 2026-09-11 pilot | First fit on the pilot generation run of #154: 145 controlled paragraphs against the comment paragraphs of the nineteen task repositories, structural features, threshold frozen at 0.5 | [runs/2026-09-11-pilot](runs/2026-09-11-pilot/README.md) |
| 2026-09-11 documentation | The same endpoints against the documentation, readme, and release-note paragraphs of the same repositories, so the negatives share the endpoints' format | [runs/2026-09-11-documentation](runs/2026-09-11-documentation/README.md) |
| 2026-09-11 enlarged | The pilot's fit repeated after the second generation run: 132 training endpoints and 58 confirmation endpoints across 25 repositories, threshold frozen at 0.5 | [runs/2026-09-11-enlarged](runs/2026-09-11-enlarged/README.md) |
