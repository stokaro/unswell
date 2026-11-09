# Source-only consistency correction before candidate execution

The first source-only seal (`labels.json`, commit `1b2d299`) had 19 positives
and 28 controls. A consistency pass over the three complete pages containing
`It is worth noting that` found that the label depended on the source, not a
defensible difference in the phrase's editorial function. The DexHunter
warning was a control while the Xid and ChatDoctor warnings were positives.
All three introduce substantive safety or repair information, and shortening
the lead-in is optional rather than a clear required diagnosis.

`labels-v2.json` changes only these judgments: DexHunter's warning is
`optional` instead of `retain`; Xid's and ChatDoctor's warnings are `optional`
instead of `positive`. The clinical-use warning remains a separate `retain`
control. The resulting source-only seal has 17 positives and 30 controls on
the same nine pages. No candidate detector or candidate output was run or
viewed before this correction. Preserve v1 for audit; use v2 for all subsequent
screens without further label changes based on candidate results.
