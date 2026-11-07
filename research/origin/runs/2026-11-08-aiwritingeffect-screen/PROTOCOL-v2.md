# Post-result document-level length sensitivity

This second screen was designed **after** the paragraph result was visible.
The first screen scored only 2,131 of 7,015 test AI paragraphs with a matched
human under the fixed 100-token minimum. This screen asks whether the same
unchanged lexical features retain their positive direction when an entire
paper variant is the scoring unit. It is exposed-data sensitivity, not an
independent confirmation or a new product qualification gate.

Use Jerry999/AIWritingEffect at commit
6a45eddc021257b7c49de398c2fffa370c1d1bfa, file
dataset/t100/docs.jsonl, exactly 75,432,568 bytes and SHA-256
428de968a71d64ddaf9f0d5afe7e4e295b2adeace77f57fd29c6252ae082d29c.
The file was downloaded and verified before this freeze. The first three rows
were inspected for schema and all 2,598 rows counted by metadata only; no
document feature score was computed. The test split has 89 human originals
and 143 AI variants from the same 89 papers. The val split has 83 human
originals; human_only has 400 additional human papers.

Apply the same prose(), ASCII word matcher, 50-token windows, and stride 25
from the pinned feature source in the first protocol. Score each complete
document with features(text) as the primary sensitivity measure. Also score
the first 100 cleaned, case-folded prose tokens as a secondary measure. The
registered direction is higher for AI variants. Under 50 prose tokens for
the complete document, or under 100 for the prefix measure, abstain for that
measure. Do not impute. AI and human variants match by paper_id; source
documents may be repeated once per AI variant in pooled AUROC.

Report eligible variants and papers, pooled directional AUROC, matched-pair
win fraction with ties worth one half, and results by generator and prompt
for each measure. Report median AI-to-human prose-token length ratio for
eligible pairs, because length imbalance may affect the full-document feature.
Groups with fewer than 20 eligible pairs have no AUROC. Do not fit a
classifier, choose a threshold, or claim a 1% false-positive operating point
from 83 validation papers. Do not select between the two measures after seeing
their results.

Fail on changed hashes, malformed rows, duplicated variants, missing matched
human originals, or overlapping paper splits. Keep raw texts outside the
repository. This dataset's sources and rewrites are the same as in the first
screen; positive values here cannot erase the paragraph coverage failure or
establish performance on software documentation. The frozen #349 editorial
events and gates remain unchanged.
