# Source context as lexical admission features: negative

The protocol and implementation were committed as `400b05a65c2551004ea48a3ff66435ad0b55393f`
before the one offline fit. The frozen script added original source lines,
nearby heading text, and line-shape tags to the earlier structured-candidate
screen. All 2,847 aliases matched their exact original UTF-8 byte spans and
all 36 source hashes matched. Twelve leave-one-repository-out folds scored
the 2,101 whole-page aliases; the other cohorts were unfiltered. No model
service was called and no product rule changed.

| Whole-page measure | Candidate-only control | Source-context treatment |
| --- | ---: | ---: |
| Held-out AUC | 0.77747 | **0.77320** |
| Fixed 0.5 FULL | 37/57 | **44/57** |
| Fixed 0.5 accepted delivery | 1,183/1,358 (87.1%) | **1,425/1,709 (83.4%)** |
| Fixed 0.5 unresolved supplied edits | 4 | **5** |
| Best optimistic FULL at at least 85% accepted | 45/57 | **43/57** |

The fixed point gained recall by admitting many more candidates, including
nonaccepted ones. It failed the preregistered joint-improvement condition and
both whole-page numeric gates. Its optimistic threshold sweep is exposed,
post-hoc feasibility arithmetic, not a policy or prospective result. The
combined lexical representation is retired without reweighting or further
threshold tuning on these grades. Contextual and historical aliases remained
unfiltered at 39/46 and 17/20 FULL, respectively; those are previously
exposed candidate-bank results, not current-main runtime results.

This result narrows the next hypothesis: lexical source proximity alone does
not identify whether a phrase performs a useful source-specific function or
whether a proposed edit preserves meaning. Any replacement needs a genuinely
new mechanism, new page-disjoint transfer, complete-stream accounting, and
separate advice-safety evidence before the #349 development and prospective
gates. Issue #349 remains open.

The source-level archive remains outside this repository. To replay with the
five SHA-256-pinned inputs, use Python 3.12 with `numpy==2.5.4`,
`scipy==1.18.1`, and `scikit-learn==1.7.2`:

```sh
python scan.py --development-root /path/to/archived/development \
  --output /tmp/source-context-replay.json
```

`result.json` SHA-256: `f8965a0a8739d3cc4900f9525a80b97fd2b2b51c0e6d78b907f1178de12eb37a`.
The public verifier checks the frozen script and protocol hashes, source input
identities, denominators, fold accounting, measured failure, and false-pass
negative controls without access to the private source-level archive.
