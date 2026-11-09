# #349 technical sentence-shape reachability screen

Status: exposed, exploratory development audit. This protocol was written after an informal count of the same five source patterns. Its result is not an independently frozen model-selection test, a qualified editorial finding, or prospective confirmation. The result may only guide a separately frozen source-necessity experiment.

## Question

Do five measured technical-prose sentence shapes proposed by Kensio overlap source-eligible target bytes of previously non-FULL #349 events often enough to merit studying contextual admission? They are candidate locations, not diagnoses. A shape match alone never warrants an edit: technical contrasts, negations, and consequences can carry essential information.

## Inputs

- Kensio `prose-check.mjs` and `measurements.md` at repository commit `e86fd6a4ff9ad8fc6576348f00c2140c3536744c`. The source study compared 65,767 words of human technical documentation with 192,430 words of one-model generated technical documentation. Its document thresholds are not imported here. The five copied regular expressions are `significance-tail`, `contrastive-def`, `contrastive-coda`, `negation-frame`, and `appositive-tail`. The method's punctuation bans and lexical-spread advisory are excluded because neither gives a justified local editorial edit.
- Existing private `originals.json` archive (SHA-256 `4f0051badc4b16dfe088d88cae69bd0dda89af2e9f62b4bed317309c4db65082`) with the 36 source pages, 123 frozen events, raw source units, byte spans, and protected ranges.
- Public `events.json` from the complete #434 editorial review, `result.json` from the intermediate scoped-degree screen, and `result.json` from the later quality-claim screen. Their hashes are in `freeze.json`. The two scoped-degree and five quality-claim additions are reconciled to the current reviewed FULL counts of 28/57 whole, 16/46 context, and 4/20 historical confirmation.

The first frozen runner stopped before producing a result because it omitted the two intermediate scoped-degree credits and failed the baseline-count assertion. Version 2 adds only that public result as an input and reconstructs the same current FULL set. The shape expressions, source pages, labels, and metrics are unchanged. Both freeze records are retained.

## Procedure

1. Verify every input hash, each source SHA-256, each raw-unit UTF-8 byte slice, every event quote and denominator. Stop on an identity, span, or accounting error.
2. Match each source shape inside the raw source units. Map each match to original UTF-8 byte offsets. Exclude a match intersecting a protected range; count these exclusions separately.
3. Count eligible matches, events with any overlap, and previously non-FULL events with any overlap, by shape and cohort. Multiple matches on one event contribute one event to the union. Count all target spans touched separately. A match that touches one word is not a complete diagnosis or safe edit.
4. Report candidate burden, including matches away from target spans. These are unannotated locations, not false positives. Do not infer accepted-delivery precision from this screen.

No rule, model, score, threshold, baseline denominator, or protected-boundary contract changes. This audit does not read untouched prospective pages, run a model, or publish private source text. If the result suggests a useful shape, its next evaluation must predeclare source-context admission, editorial criteria, close legitimate controls, held-out pages, and the unchanged #349 dual gates before any product change.
