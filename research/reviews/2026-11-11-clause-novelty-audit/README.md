# Clause lexical novelty does not triage long-sentence warnings

**Status: exposed exploratory negative screen.** This audit tests one fixed,
source-only relation proxy on the 84 whole-page technical
`syntax.long-sentence` findings in the [November 7 complete editorial
review](../2026-11-07-current-main-editorial-review/README.md). The review
labels were already exposed before this screen. It is not a registered
prospective experiment, a new editorial review, or a product change.

The proxy splits each reported sentence at commas, semicolons, colons, and
`but`, `while`, `whereas`, `because`, `and`, `or`, and `so`. It lowercases word
tokens, drops a fixed stopword set and terms shorter than three characters,
and retains segments with at least three distinct content terms. Its score is
one minus mean Jaccard overlap of adjacent retained segments. Higher scores
hypothesize less related clauses. There is no fitted vocabulary, threshold,
classifier, or post-result policy. The comparison uses the reported prose-word
count on exactly the same rows.

| Exposed whole-page measure | Result |
| --- | ---: |
| Eligible long-sentence findings | 84/84 |
| Accepted / nonaccepted review decisions | 46 / 38 |
| Clause lexical novelty AUROC | **0.4954** |
| Prose-word-count AUROC on the same rows | 0.7068 |

Nine pages have both labels. Their within-page novelty AUROCs range from 0 to
1, often on only two to five findings; the full matrix is in `result.json`.
The pooled result does not separate the assistant's accepted and nonaccepted
warnings. It cannot support a suppression rule or a threshold search. The
[earlier length audit](../2026-11-08-length-triage-audit/README.md) already
shows that even the better-ranking word count loses many accepted warnings
when used as a filter. In particular, a useful next candidate must represent
whether clauses express independent technical cases, conditions, or one
coherent causal explanation; low lexical overlap alone does not establish that
relation. The original #349 event labels, delivered findings, and all 80%/85%
gates remain unchanged.

`analyze.py` checks SHA-256 identities for the complete archived technical
report and editorial review, joins findings by ID, requires all 84 rows, and
recomputes `result.json`. The report was generated at `5dd7619`; the relevant
`builtin/context_local.go` and `builtin/context.go` files are unchanged at
main `05a263a`. The script and result were committed after this exposed
calculation, so this record does not claim a pre-inference freeze. No hosted
model, new source, new label, or prospective page was used.

Reproduce from the repository root:

```sh
python3 -B research/reviews/2026-11-11-clause-novelty-audit/analyze.py \
  > /tmp/unswell-clause-novelty-result.json
diff -u research/reviews/2026-11-11-clause-novelty-audit/result.json \
  /tmp/unswell-clause-novelty-result.json
```
