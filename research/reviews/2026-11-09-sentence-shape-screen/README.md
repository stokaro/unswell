# #349 technical sentence-shape screen: useful locations, excessive candidate load

**Status: exposed exploratory screen; no product change or editorial qualification.**
This audit transfers five measured sentence-shape expressions from
[Kensio's pinned technical-prose study](https://github.com/KensioSoftware/kensio.ai/blob/e86fd6a4ff9ad8fc6576348f00c2140c3536744c/plugins/technical-prose-style/skills/technical-prose-style/references/measurements.md)
to the 36 source pages and 123 frozen #349 events. It uses the expressions as
candidate locations only. It does not transfer Kensio's document thresholds,
claim AI authorship, or infer that a matched phrase needs an edit. The upstream
comparison used one generating model and a small set of human technical
documents; its corpus separation is not an Unswell editorial result.

The first frozen runner stopped before writing a result because it omitted two
intermediate scoped-degree credits from the current FULL baseline. The v2
freeze adds that public result and preserves the original archive, decisions,
five expressions, and metrics. `freeze-v1.json` and `freeze.json` retain both
identities. No model call or prospective page was used.

| Sentence shape | Eligible locations | Locations touching a frozen target | Previously non-FULL events touched |
| --- | ---: | ---: | ---: |
| Significance tail | 135 | 10 | 7 |
| Contrastive definition | 148 | 19 | 11 |
| Contrastive coda | 21 | 3 | 3 |
| Negation frame | 331 | 11 | 9 |
| Appositive tail | 45 | 10 | 6 |

Across the five shapes there are **680 eligible matches** after excluding 17
matches that intersect protected source ranges. They touch 40 distinct frozen
events, including **28 previously non-FULL events**: 11 whole-page, 7
contextual, and 10 historical confirmation. Every target span has some overlap
for those 28, but the expressions cover only small parts of most targets. This
is an optimistic location count, **not** 28 recovered FULL diagnoses. Another
627 eligible matches are away from frozen target spans. Those locations are
unannotated, not automatically false positives; they show the review and
warning burden of a direct transfer.

The close examples explain why the shape is insufficient. `Adoption is a path,
not a switch` is an annotated framing target, while `Moving the rows is a
choice, not a requirement` states an operational distinction. Similarly, an
appositive clause can repeat a point or specify a real command behavior. The
source relation and necessity of the second clause must be judged before
issuing advice. A document-level rate threshold cannot make that decision for
a particular byte span. The unchanged current reviewed baseline remains
28/57, 16/46, and 4/20 FULL in the three cohorts; #349 remains open.

The next admissible study is a separately frozen contextual selector over these
source-bound locations, with technical contrast and consequence controls,
repository-held-out evaluation, complete finding review, and suggestion-safety
checks. The present data cannot set its threshold or qualify it.

Reproduce with the pinned private archive:

```sh
python3 -B research/reviews/2026-11-09-sentence-shape-screen/screen.py \
  --archive /path/to/sha256-pinned/originals.json \
  --out /tmp/sentence-shape-result.json
```

`PROTOCOL.md`, both freeze records, `screen.py`, and `result.json` preserve the
inputs, byte-span checks, accounting, and the bounded conclusion. The result
contains aggregate counts and public event IDs, never private source text.
Its SHA-256 is
`e8fafc2468037626c68774b2ab5e94f33f9bec5f5acd34593511dd6aa9248e8d`.
Run `python3 -B research/reviews/2026-11-09-sentence-shape-screen/verify_result.py`
and `python3 -B -m unittest discover -s
research/reviews/2026-11-09-sentence-shape-screen -p 'test_*.py'` to verify the
public result identity and accounting. Negative tests reject inflated counts
and an unsupported qualification claim. The public verifier cannot
independently recheck private source text without the pinned archive. This
check stays outside `Makefile` because an earlier frozen research packet binds
that file's SHA-256.
