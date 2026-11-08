# #349 counterfactual admission controls: offline baseline

This is a **synthetic exposed-development control**, not a scored run on the
original 36 pages. It tests whether a diagnosis follows a specific source
relation when surrounding evidence changes. The unchanged Go engine was run on
16 frozen Markdown variants; no experimental model was called, no original
event or grade changed, and no runtime rule was added.

The evaluation design adapts the evidence-intervention idea in
[EviScope](https://aclanthology.org/2026.groundlm-1.20/), not its model results
or its explicit gate. [FOXGLOVE](https://aclanthology.org/2026.emnlp-main.781/)
motivates treating feedback goal and target location as separate measurements.
Neither paper validates these synthetic labels or Unswell's product behavior.

## Frozen packet and measured baseline

Four quartets hold their target sentence and narrowly worded diagnosis fixed.
The source changes a nearby heading, comparison, reader observation, or release
fact. There are eight variants where that **specific** diagnosis applies and
eight where it does not. Some negative variants have a different genuine
problem; an overlapping warning cannot be credited for the frozen diagnosis.

[The protocol](PROTOCOL.md), [all source variants and rationales](cases.json),
the [input verifier](verify.py), and the [runner](run.py) were committed in
`fe727da86cad38b60015fbf1f80dfbbad9884b01` before the baseline was run.
The freeze SHA-256 is
`14f7e939bc969eaca950595d8a06172effe96e17c42286167a4898b649d31aa0`.
The CLI was built from that commit, whose parent is the accepted main
`00a3b024488d3ca2e38f5fa71ac38588652de165`; the research-only commit
changed no runtime code.

| Measurement | Result |
| --- | ---: |
| Frozen source variants | 16 |
| Fixed-diagnosis positive controls | 8 |
| Complete CLI reports | 16/16 |
| All delivered findings | 5 |
| Findings overlapping the target bytes | 5 |
| Findings that express the fixed diagnosis | **0** |
| Operational errors | 0 |

Four `filler.evaluative-closure` findings address the carrier
“I want to stress that” on **all four** release-status variants, including
those without a repeated release fact. The fifth finding is a passive-voice
candidate on the no-table variant. The [source review](review.json) assigns
none of these to the fixed contextual diagnoses. This is review by a Codex
assistant under ADR 0041, not an independent human judgment. No applied edit
or meaning-preservation claim is made.

The complete [baseline summary](baseline.json), all
[16 raw CLI reports](baseline-reports.tar.gz), [decision seal](decision.json), and
[independent artifact verifier](verify_result.py) are committed. The verifier
checks exact tool and source hashes, every frozen target offset, report
completeness, compact/raw finding equality, overlap, and a review decision for
each overlapping finding. Fifteen positive and negative contract tests pass.

## Decision and next gate

The current engine does not express any of these eight fixed diagnoses on
these synthetic controls. This is a **baseline gap**, not evidence that a new
candidate can close it. It contributes zero new FULL-event credit to the
unchanged #349 denominators (57 whole, 46 context, 20 historical confirmation).

A future source-sensitive admission candidate needs its own frozen request
contract, failure policy, model/effort, budget, and execution authorization.
The registered preflight requires at least 14/16 correct fixed-diagnosis
decisions, both sides of every evidence-flip pair, exact coordinates, and zero
unsafe or unresolved advice. Passing that small preflight would only permit
the complete exposed 36-page/123-event screen at the existing 80% FULL-event
recall and 85% accepted-delivery gates, followed by separate prospective
confirmation. No candidate or model execution is included in this record.
