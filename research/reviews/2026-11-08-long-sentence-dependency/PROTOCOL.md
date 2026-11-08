# Source-only long-sentence dependency audit

Date: November 8, 2026. This bounded exposed-development audit continues the
six-pair feasibility question selected in the [#421 decision](../2026-11-03-issue421/README.md).
It tests one source-visible distinction: whether a long sentence combines
independently actionable cases or roles that readers must retrieve separately,
versus one coherent rule whose condition, operation, and consequence belong
together. It does not fit a threshold, change the runtime, or claim that sentence
length is an editorial defect.

Use the November 7 complete assistant review and the 36 original source pages.
Its 97 `syntax.long-sentence` findings contain 52 actionable, 44 uncertain, and
one nonactionable judgment. These source-context judgments are exposed
development labels, not independent human annotation. The older #421 decision
used v30's 52 accepted and 45 uncertain findings; this audit updates the input
to the current complete review, while keeping its six-pair scope and sole
FULL-event witness requirement. No reserved prospective page is opened.

## Frozen selection and blinded review

Select six same-page actionable/uncertain pairs from pages `whole/p026.md`,
`whole/p017.md`, `whole/p001.md`, `whole/p002.md`, `context/c04.md`, and
`confirmation/c01.mdx`. The accepted member on `whole/p026.md` is finding
`5f3c42940d99ad29`, the sole current FULL-event witness for
`whole/p026-d02`. For each other page, select the actionable finding with the
smallest SHA-256 of `long-sentence-dependency-v1:` plus its finding ID. Pair it
with the uncertain finding on that page closest in target word count using the
preparer's ASCII word regex `[A-Za-z]+(?:'[A-Za-z]+)?`;
resolve ties with the same hash order. Order A/B by the low bit of that hash.
This deterministic selection is frozen before viewing the resulting packet.

The preparation script verifies the frozen input hashes, all 36 source hashes,
and each exact original UTF-8 byte span. It creates a local packet with the
six full original pages and exact target quotes, plus a separate key containing
old grades and event identity. Do not read the key before sealing the review.
The reviewer may have seen other frozen labels before this audit, so this is
label-hidden within the attempt, not an independent or truly blind review.

For every A/B passage, read the complete original page and record:

1. Whether it contains at least two independent cases, actions, or outcomes
   whose separate presentation is necessary for lookup or scope comprehension.
2. The exact source evidence for those independent parts, or for the single
   dependency that should remain together.
3. `split_needed`, `retain`, or `uncertain`. A possible stylistic split alone
   is `retain`; technical conditions, negation, identifiers, quantities, and
   protected content must remain intact. Do not propose an applied edit.

Seal all twelve decisions and rationales before opening the key. A pair
separates only when its actionable member is `split_needed` and its uncertain
member is `retain`. The preregistered feasibility screen requires at least
five of six separated pairs **and** `split_needed` on the sole FULL witness.
Any `uncertain` decision fails its pair. If the screen fails, stop this family:
do not tune a threshold, change the case selection, or implement a rule from
these examples. If it passes, freeze exactly one source-grounded restriction
before replaying the full unchanged 36-page stream and all 123 events.

This audit does not itself improve #349 recall or accepted delivery. The
unchanged qualification gates remain at least 80% FULL-event recall and 85%
accepted actual delivered findings in each complete cohort, followed by
separate confirmation and advice-safety review. Its only possible immediate
outcomes are permission to test one restriction or a documented stop decision.
