# Frozen feature screen: lexical richness and entropy

Date: November 7, 2026. This is a retrospective, offline feature-feasibility
screen of the exposed [long-prose-v3](../../../generation/studies/long-prose-v3/README.md)
archive. It tests a method idea from [Shan et al., EMNLP 2026](https://aclanthology.org/2026.emnlp-main.1725/)
and [El Attar et al., Findings of AACL 2026](https://aclanthology.org/2026.findings-aacl.10/).
It is not a reproduction of either paper, a product detector, or a new holdout.

## Fixed input and unit

Use all 38 source documents in `inputs.tar.gz` and all 456 successful saved
responses in `outputs.tar.gz`: three model families, two operations (`generate`
and `polish`), and two prompt conditions per source. Verify both archive hashes
against their adjacent JSON wrappers, all source hashes in `selection.json`,
and every request's source association. A source document is the paired
observation. Aggregate by the 20 source repositories; requests from one
repository are not independent samples. Do not call a model, download a
dataset, or alter any frozen labels or outcomes.

## Fixed computation

Remove fenced code blocks, inline code, URLs, and HTML tags from each complete
Markdown text. Count case-folded ASCII word tokens matching
`[A-Za-z]+(?:'[A-Za-z]+)?`. For each token sequence, compute 50-token windows
with stride 25. Lexical richness is mean distinct tokens divided by 50 across
windows. Entropy is mean Shannon word entropy divided by `log2(50)` across
windows. A text with fewer than 50 prose tokens abstains for both features;
it still counts in request coverage. Also record the ratio of surviving prose
tokens to all raw word tokens as a formatting diagnostic.

Compare each saved response only to its own source. Report operation-level
counts and median paired deltas, then one mean delta per repository and the
number of repositories with positive, negative, or zero delta. Report all
three model families separately. Give the same summary for the 0.75–1.33
response/source prose-token length-ratio subset, with its excluded count.
This subset is a sensitivity check; it does not replace the all-response
result. A whitespace reflow of the prose after code removal must leave both
features unchanged for all texts.

## Interpretation fixed before measurement

The literature-derived generated-text direction is **positive** for both
features. A positive feasibility screen requires positive repository-mean
deltas for both features in at least 15 of 20 repositories, plus positive
median deltas for both features in each of the three model families. If either
condition fails, do not build an origin classifier from just these features
on this corpus. The `polish` arm is descriptive: AI editing need not have the
same signature. No threshold is selected on these outputs; no AUROC,
false-positive rate, editorial finding recall, or authorship probability is
inferred. Historical source provenance remains `unknown`, not proven human.

Only the outputs of `run.py` after this protocol and the code are frozen count
as this screen. If a verification or computation error occurs, stop and record
it rather than editing the protocol in place.
