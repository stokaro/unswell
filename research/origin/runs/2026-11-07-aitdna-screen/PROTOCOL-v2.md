# Post-result AITDNA sensitivity analysis

Date: 2026-11-07. This was specified after viewing the v1 aggregate scores
(richness AUROC 0.686, entropy AUROC 0.665) and the document representation.
It is an exposed-data sensitivity analysis, not independent confirmation.
The v1 protocol, freeze, raw pages, and result remain unchanged.

Use the exact AITDNA document parquet at commit
21d77f7d44833507750c0da309295c4147c1bda5, SHA-256
ff98ea823d2760b4f46f3ad189414a51cfc763591eccd3eb29be088a60d9d300.
Its 362 rows matched all four frozen datasets-server pages by ordered data and
metadata values. Use the same 100-token prefix and unchanged Unswell feature
source as v1. Require one document snippet per row; this was learned during
v1 and must be disclosed. Do not call a model or change a runtime rule.

Measure two questions, without fitting or threshold selection:

1. For each author and task with both human-only and AI-assisted documents,
   compare the mean assisted feature with the mean human-only feature. Report
   pair counts, positive/negative/tied directions, and median paired delta.
   This narrows task and participant confounding but does not prove causality.
2. Split AI-assisted documents by the document notion's single snippet author
   value, Bot versus User. For each group, compute directional rank AUROC
   against all human-only documents and report group count. The User-labelled
   group had AI access in the study but its final document is not labelled as
   bot-written by this notion. Do not interpret this as a raw edit-history
   token fraction. Also report the human-only snippet author counts.

Record the result and SHA-256. No new pass/fail gate is defined; these are
diagnostics of the already exposed v1 signal. The technical-document domain
and issue #349 editorial event gates remain untested by this sensitivity run.
